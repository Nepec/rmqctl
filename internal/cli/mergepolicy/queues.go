package mergepolicy

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"text/tabwriter"

	"github.com/nepec/rmqctl/internal/api"
	"github.com/nepec/rmqctl/internal/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var applyToQueuesCmd = &cobra.Command{
	Use:   "queues",
	Short: "Merge policies applied to queues",
	Long: `Merge a policy definition from a file into the effective policy of each matching queue.

For every selected vhost, this command fetches the current effective policy for each queue and 
merges in the definition from the input file. 
Queues may be filtered applying the corresponding filter flags.
If a queue has no effective policy, a new one is created using only
the definitions from the file.

The merge is a soft-merge by default: keys are only added, never overwritten. Use --force
to overwrite keys that already exist in the current policy.

Filtering:
  The --contains and --type flags narrow the queue set by name and type.

  The boolean filters (--empty, --active, --with-policy) are three-valued:
    omitted     no filtering on that property
    --flag      include only queues where the property is true
    --flag=false include only queues where the property is false

  For example, --empty selects only empty queues, while --empty=false selects only
  non-empty queues, and omitting it includes queues regardless of message count.

Use --dry-run to preview which queues would be updated without applying any changes.`,
	Aliases:      []string{"queue", "qs", "q"},
	SilenceUsage: true,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		qType, _ := cmd.Flags().GetString("type")
		return cli.ValidateQueueType(qType)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		adapter, err := cli.ClientFromConfig()
		if err != nil {
			return fmt.Errorf("failed to connect to rabbitmq: %w", err)
		}

		vhosts, err := cli.ResolveVhosts(adapter, viper.GetStringSlice("vhosts"))
		if err != nil {
			return fmt.Errorf("parsing input vhost list '%v': %w", viper.GetStringSlice("vhosts"), err)
		}

		filterOpts := &api.QueueFilterOpts{}

		filterOpts.Contains, _ = cmd.Flags().GetString("contains")
		filterOpts.Type, _ = cmd.Flags().GetString("type")
		empty, _ := cmd.Flags().GetBool("empty")
		if cmd.Flags().Changed("empty") {
			filterOpts.Empty = &empty
		}
		active, _ := cmd.Flags().GetBool("active")
		if cmd.Flags().Changed("active") {
			filterOpts.Active = &active
		}
		withPolicy, _ := cmd.Flags().GetBool("with-policy")
		if cmd.Flags().Changed("with-policy") {
			filterOpts.WithPolicy = &withPolicy
		}

		return MergeQueuePolicyAction(os.Stdout, adapter, vhosts, definitionFile, filterOpts, dryRun, force)
	},
}

func init() {
	mergePolicyCmd.AddCommand(applyToQueuesCmd)

	applyToQueuesCmd.Flags().StringP("contains", "c", "", "Only include queues whose name contains this substring")
	applyToQueuesCmd.Flags().StringP("type", "t", "", "Only include queues of this type (classic or quorum)")
	applyToQueuesCmd.Flags().BoolP("empty", "e", false, "Only include empty queues")
	applyToQueuesCmd.Flags().BoolP("active", "a", false, "Only include queues with active consumers")
	applyToQueuesCmd.Flags().Bool("with-policy", false, "Only include queues with an effective policy")
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

func MergeQueuePolicyAction(out io.Writer, c api.RabbitClient, vhosts []string, definitionFile string, qFilterOpts *api.QueueFilterOpts, dryRun, force bool) error {
	// #nosec G304 -- definitionFile is a user-supplied CLI flag, not untrusted network input
	df, err := os.Open(definitionFile)
	if err != nil {
		return err
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

type PolicyTableFormatter struct{}

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
