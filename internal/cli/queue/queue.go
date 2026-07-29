package queue

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/nepec/rmqctl/internal/api"
	"github.com/nepec/rmqctl/internal/cli/vhost"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type queueCommand struct {
	getClient api.ClientFactory

	filter FilterOptions
}

func NewListQueueCommand(getClient api.ClientFactory) *cobra.Command {
	c := &queueCommand{getClient: getClient, filter: FilterOptions{}}

	short := "List queues"
	long := `List queues across one or more vhosts, with selected details.

Without failures, all queues in the targeted vhost are listed.

Filtering:
The --contains and --type flags narrow the queue set by name and type.

The boolean filters (--empty, --active, --with-policy) are three-valued:
	omitted     no filtering on that property
	--flag      include only queues where the property is true
	--flag=false include only queues where the property is false

For example, --empty selects only empty queues, while --empty=false selects only
non-empty queues, and omitting it includes queues regardless of message count.
`

	queueCmd := newDefaultQueueCommand(short, long, &c.filter, c.validate, c.list)

	return queueCmd
}

func newDefaultQueueCommand(short, long string, filter *FilterOptions, validate func(*cobra.Command, []string) error, execute func(*cobra.Command, []string) error) *cobra.Command {
	queueCmd := &cobra.Command{
		Use:          "queues",
		Short:        short,
		Long:         long,
		SilenceUsage: true,
		Aliases:      []string{"queue", "qs", "q"},
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

	queueCmd.Flags().StringSlice("vhosts", []string{"/"}, "Virtual Host where to list queues from")

	AddFilterFlags(queueCmd.Flags(), filter)

	return queueCmd
}

func (q *queueCommand) validate(_ *cobra.Command, _ []string) error {
	return ValidateQueueType(q.filter.QueueType)
}

func AddFilterFlags(fs *pflag.FlagSet, opts *FilterOptions) {
	fs.StringVarP(&opts.Contains, "contains", "c", "", "Only include queues whose name contains this substring")
	fs.StringVarP(&opts.QueueType, "type", "t", "", "Only include queues of this type (classic or quorum)")
	fs.BoolVarP(&opts.Empty, "empty", "e", false, "Only include empty queues")
	fs.BoolVarP(&opts.Active, "active", "a", false, "Only include queues with active consumers")
	fs.BoolVar(&opts.WithPolicy, "with-policy", false, "Only include queues with an effective policy")
}

func (q *queueCommand) list(cmd *cobra.Command, _ []string) error {
	client, err := q.getClient()
	if err != nil {
		return fmt.Errorf("could not create rabbitmq client: %w", err)
	}
	vhosts, err := vhost.Resolve(client, viper.GetStringSlice("vhosts"))
	if err != nil {
		return fmt.Errorf("parsing input vhost list '%v' for %q: %w", viper.GetStringSlice("vhosts"), client.Host(), err)
	}

	return executeList(cmd.OutOrStdout(), client, vhosts, q.filter.ToQueueFilterOpts(cmd))
}

func executeList(out io.Writer, c api.RabbitClient, vhosts []string, opts *api.QueueFilterOpts) error {
	var results []api.Queue

	defer func() {
		formatter := QueueTableFormatter{}
		formatter.Print(out, results)
	}()

	for _, vhost := range vhosts {
		queues, err := c.ListQueuesIn(vhost)
		if err != nil {
			return fmt.Errorf("rmqctl %q: %w", c.Host(), err)
		}
		results = append(results, api.FilterQueues(queues, opts)...)
	}

	return nil
}

type FilterOptions struct {
	Contains   string
	QueueType  string
	Empty      bool
	Active     bool
	WithPolicy bool
}

func (f *FilterOptions) ToQueueFilterOpts(cmd *cobra.Command) *api.QueueFilterOpts {
	opts := &api.QueueFilterOpts{
		Contains: f.Contains,
		Type:     f.QueueType,
	}

	if cmd.Flags().Changed("empty") {
		opts.Empty = &f.Empty
	}
	if cmd.Flags().Changed("active") {
		opts.Active = &f.Active
	}
	if cmd.Flags().Changed("with-policy") {
		opts.WithPolicy = &f.WithPolicy
	}

	return opts
}

func ValidateQueueType(t string) error {
	switch t {
	case "", "classic", "quorum":
	// ok
	default:
		return fmt.Errorf("queue type may wither by 'classic' or 'quorum', got %q", t)
	}

	return nil
}

type QueueTableFormatter struct{}

func (q QueueTableFormatter) Print(out io.Writer, results []api.Queue) {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	defer func() {
		_ = w.Flush()
	}()

	fmt.Fprintln(w, "VHOST\tNAME\tTYPE\tMESSAGES\tACTIVE\tPOLICY NAME")

	for _, r := range results {
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%t\t%s\n", r.Vhost, r.Name, r.Type, r.Messages, r.Active, r.PolicyName)
	}
	fmt.Fprintf(out, "\nFound %d queue(s)\n", len(results))
	fmt.Fprintln(out, "")
}
