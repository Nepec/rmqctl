package list

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/nepec/rmqctl/internal/api"
	"github.com/nepec/rmqctl/internal/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var queuesCmd = &cobra.Command{
	Use:   "queues",
	Short: "List queues",
	Long: `List queues across one or more vhosts, with selected details.

Without failures, all queues in the targeted vhost are listed. Filter
flags can be combined; a queue must match them all to be shown.
Boolean filters also accept explicit negation, e.g. --empty=false lists
only queues that have messages.`,
	SilenceUsage: true,
	Aliases:      []string{"queue", "qs", "q"},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		qType, _ := cmd.Flags().GetString("type")
		return cli.ValidateQueueType(qType)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		adapter, err := cli.ClientFromConfig()
		if err != nil {
			return fmt.Errorf("init command %q: %w", cmd.Name(), err)
		}

		vhosts, err := cli.ResolveVhosts(adapter, viper.GetStringSlice("vhosts"))
		if err != nil {
			return fmt.Errorf("parsing input vhost list '%v' for %q: %w", viper.GetStringSlice("vhosts"), adapter.Host(), err)
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

		return listAction(os.Stdout, adapter, vhosts, filterOpts)
	},
}

func init() {
	listCmd.AddCommand(queuesCmd)

	queuesCmd.PersistentFlags().SortFlags = false
	queuesCmd.Flags().SortFlags = false

	queuesCmd.Flags().StringP("contains", "c", "", "Only include queues whose name contains this substring")
	queuesCmd.Flags().StringP("type", "t", "", "Only include queues of this type (classic or quorum)")
	queuesCmd.Flags().BoolP("empty", "e", false, "Only include empty queues")
	queuesCmd.Flags().BoolP("active", "a", false, "Only include queues with active consumers")
	queuesCmd.Flags().Bool("with-policy", false, "Only include queues with an effective policy")
}

func listAction(out io.Writer, c api.RabbitClient, vhosts []string, opts *api.QueueFilterOpts) error {
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
