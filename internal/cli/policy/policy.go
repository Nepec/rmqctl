// Package policy provides the "policy" command, which merges a policy
// definition from a file into the effective policy of every queue
// matching the selected vhosts and filters.
package policy

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"text/tabwriter"

	"github.com/nepec/rmqctl/internal/api"
	"github.com/nepec/rmqctl/internal/cli/queue"
	"github.com/nepec/rmqctl/internal/cli/vhost"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// MergeOptions configures the policy merge command's behavior: which file
// to read the policy definition from, whether to overwrite conflicting
// keys (Force), and whether to preview changes without applying them
// (DryRun).
type MergeOptions struct {
	DefinitionsFile string
	Force           bool
	DryRun          bool
}

// PolicyFilterOptions narrows which existing policies are considered when
// merging.
type PolicyFilterOptions struct {
	ApplyTo string
}

type policyCommand struct {
	getClient api.ClientFactory

	// Flags
	opts         *MergeOptions
	policyFilter PolicyFilterOptions
	queueFilter  queue.FilterOptions
}

// NewMergeQueuePolicy builds the "policy" command. For each selected
// vhost, it fetches every matching queue's effective policy and merges in
// the definition from opts.DefinitionsFile — soft-merging by default, or
// overwriting existing keys if opts.Force is set. Queues with no
// effective policy get a new one containing only the file's definition.
func NewMergeQueuePolicy(getClient api.ClientFactory, opts *MergeOptions) *cobra.Command {
	c := &policyCommand{
		getClient:    getClient,
		opts:         opts,
		policyFilter: PolicyFilterOptions{},
		queueFilter:  queue.FilterOptions{},
	}

	queueCmd := newDefaultMergePolicyCommand(&c.policyFilter, &c.queueFilter, c.validate, c.mergeQueuePolicy)
	return queueCmd
}

func newDefaultMergePolicyCommand(policyFilter *PolicyFilterOptions, queueFilter *queue.FilterOptions, validate func(*cobra.Command, []string) error, execute func(*cobra.Command, []string) error) *cobra.Command {
	queueCmd := &cobra.Command{
		Use:   "policy",
		Short: "Merge policy definitions applied to queues",
		Long: `Merge a policy definition from a file into the effective policy of each matching queue.

For every selected vhost, this command fetches the current effective policy for each queue and 
merges in the definition from the input file. 
Queues may be filtered applying the corresponding filter flags.
If a queue has no effective policy, a new one is created using only
the definitions from the file.

The merge is a soft-merge by default: keys are only added, never overwritten. Use --force
to overwrite keys that already exist in the current policy.

Queue filtering is performed with the same options as the list queues command.

Use --dry-run to preview which queues would be updated without applying any changes.`,
		SilenceUsage: true,
		Aliases:      []string{"pols", "pol", "p"},
		Args:         cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := validate(cmd, args); err != nil {
				return err
			}
			if err := vhost.BindVhosts(cmd, args); err != nil {
				return err
			}
			return nil
		},
		RunE: execute,
	}

	queueCmd.PersistentFlags().SortFlags = false
	queueCmd.Flags().SortFlags = false

	// Policy filters
	queueCmd.Flags().StringVar(&policyFilter.ApplyTo, "apply-to", "queues", "Only inlcude policies which apply to this resource (queues or exchanges)")

	// Resource filters
	queue.AddFilterFlags(queueCmd.Flags(), queueFilter)

	queueCmd.PersistentFlags().StringSlice("vhosts", []string{"/"}, "Virtual hosts")

	return queueCmd
}

func (p *policyCommand) validate(_ *cobra.Command, _ []string) error {
	if err := validateApplyTo(p.policyFilter.ApplyTo); err != nil {
		return err
	}
	if err := queue.ValidateQueueType(p.queueFilter.QueueType); err != nil {
		return err
	}
	return nil
}

func validateApplyTo(t string) error {
	switch t {
	case "", "queues", "exchanges":
	// ok
	default:
		return fmt.Errorf("apply to may either by 'queues' or 'exchanges', got %q", t)
	}

	return nil
}

func (p *policyCommand) mergeQueuePolicy(cmd *cobra.Command, _ []string) error {
	client, err := p.getClient()
	if err != nil {
		return fmt.Errorf("could not create rabbitmq client: %w", err)
	}
	vhosts, err := vhost.Resolve(client, viper.GetStringSlice("vhosts"))
	if err != nil {
		return fmt.Errorf("parsing input vhost list '%v' for %q: %w", viper.GetStringSlice("vhosts"), client.Host(), err)
	}
	return executeMerge(cmd.OutOrStdout(), client, vhosts, p.opts.DefinitionsFile, p.queueFilter.ToQueueFilterOpts(cmd), p.opts.DryRun, p.opts.Force)
}

func executeMerge(out io.Writer, c api.RabbitClient, vhosts []string, definitionFile string, qFilterOpts *api.QueueFilterOpts, dryRun, force bool) error {
	// #nosec G304 -- definitionFile is a user-supplied CLI flag, not untrusted network input
	df, err := os.Open(definitionFile)
	if err != nil {
		return fmt.Errorf("reading definitions file %q: %w", definitionFile, err)
	}
	defer func() {
		_ = df.Close()
	}()

	var definition map[string]any
	if err := json.NewDecoder(df).Decode(&definition); err != nil {
		return fmt.Errorf("could not read custom policy definition in %q: %w", definitionFile, err)
	}
	if len(definition) == 0 {
		return fmt.Errorf("policy definition in %q is empty", definitionFile)
	}
	slog.Debug("policy definition loaded", "file", definitionFile, "contents", definition)

	var results []policyUpdateResult
	defer func() {
		formatter := PolicyTableFormatter{}
		formatter.Print(out, results, dryRun)
	}()

	for _, vhost := range vhosts {
		slog.Debug("processing vhost", "vhost", vhost)

		queues, err := c.ListQueuesIn(vhost)
		if err != nil {
			return fmt.Errorf("rmqctl %q: could not initialize queues list for: %w", c.Host(), err)
		}
		queues = api.FilterQueues(queues, qFilterOpts)

		policies, err := c.ListPoliciesIn(vhost)
		if err != nil {
			return fmt.Errorf("rmqctl %q: could not initialize policy list: %w", c.Host(), err)
		}

		slog.Debug("policies found", "vhost", vhost, "count", len(policies))

		for _, q := range queues {
			result := policyUpdateResult{
				Vhost:        vhost,
				Queue:        q.Name,
				Policy:       q.PolicyName, // default value for minimal information in case of errors
				UpdateStatus: statusDryRun,
			}

			policy, err := createOrMergePolicy(q, vhost, policies, definition, force)
			if err != nil {
				slog.Warn("could not create or merge policy", "vhost", vhost, "queue", q.Name, "error", err)
				result.UpdateStatus = statusFailed
				results = append(results, result)
				continue
			}

			slog.Debug("applying policy", "vhost", vhost, "queue", q.Name, "policy", policy.Name, "definition", policy.Definition)

			result.Policy = policy.Name

			if !dryRun {
				result.UpdateStatus = statusUpdated
				if err := c.PutPolicy(vhost, policy.Name, policy); err != nil {
					// Non-blocking error, log and continue
					slog.Warn("could not apply policy", "vhost", vhost, "queue", q.Name, "policy", policy.Name, "definition", policy.Definition, "error", err)
					result.UpdateStatus = statusFailed
				}
			}
			results = append(results, result)
		}
	}

	return nil
}

func createOrMergePolicy(q api.Queue, vhost string, policies map[string]api.Policy, definition map[string]any, force bool) (api.Policy, error) {
	if q.PolicyName == "" {
		return api.NewStandardQueuePolicy(vhost, q.Name, definition), nil
	}

	current, ok := policies[q.PolicyName]
	if !ok {
		return api.Policy{}, fmt.Errorf("policy %q not found", q.PolicyName)
	}
	slog.Debug("merging with existing policy", "queue", q.Name, "policy", q.PolicyName)
	current.Definition = api.MergePolicyDefinitions(current.Definition, definition, force)

	return current, nil
}

// PolicyTableFormatter renders policy merge results as a tab-aligned
// table.
type PolicyTableFormatter struct{}

// Print writes results to out as a tab-aligned table, followed by a
// totals line. dryRun controls whether rows are reported as
// "would-update" or "updated".
func (p PolicyTableFormatter) Print(out io.Writer, results []policyUpdateResult, dryRun bool) {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	defer func() {
		_ = w.Flush()
	}()

	fmt.Fprintln(w, "VHOST\tQUEUE\tPOLICY\tSTATUS")
	updated, wouldUpdate, failed := 0, 0, 0
	for _, r := range results {
		status := "updated"
		switch r.UpdateStatus {
		case statusDryRun:
			status = "would-update"
			wouldUpdate++
		case statusFailed:
			status = "skipped"
			failed++
		default:
			updated++
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", r.Vhost, r.Queue, r.Policy, status)
	}

	if dryRun {
		fmt.Fprintf(w, "\n(dry run) Total: %d would be updated, %d failed\n", wouldUpdate, failed)
	} else {
		fmt.Fprintf(w, "\nTotal: %d updated, %d failed\n", updated, failed)
	}
}

type updateStatus int

const (
	statusUpdated updateStatus = iota
	statusFailed
	statusDryRun
)

type policyUpdateResult struct {
	Vhost        string
	Queue        string
	Policy       string
	UpdateStatus updateStatus
}
