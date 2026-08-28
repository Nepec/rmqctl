package queue

import (
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/nepec/rmqctl/internal/api"
	"github.com/nepec/rmqctl/internal/cli/sharedopts"
	"github.com/nepec/rmqctl/internal/cli/vhost"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewDeleteQueueCommand builds the "queues" command, which deletes queues
// across one or more vhosts reachable through getClient, narrowed by the
// filter flags added via AddFilterFlags.
func NewDeleteQueueCommand(getClient api.ClientFactory, opts *sharedopts.DeleteOptions) *cobra.Command {
	c := &queueCommand{
		getClient: getClient,
		filter:    FilterOptions{},
		delOpts:   opts,
	}

	short := "Delete queues"
	long := `Delete queues across one or more vhosts.

Queues may be selected either by name (as a positional argument) or by
filter flags, the two are mutually exclusive.
For details on how the filters behave, see 'rmqctl list queues --help'.

A queue that still has messages or active consumers fails to delete unless:
	--force/-f   bypasses that safety check for queues that already
	             matched the filter (or the positional name), deleting
	             them even if they still have messages or active
	             consumers.

Use --dry-run to list, across all selected vhosts, which queues would be
deleted, without deleting anything.
`

	queueCmd := newDefaultQueueCommand(short, long, &c.filter, c.validate, c.delete)
	queueCmd.Args = cobra.MaximumNArgs(1)
	queueCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if err := c.validate(cmd, args); err != nil {
			return err
		}
		if err := vhost.BindVhosts(cmd, args); err != nil {
			return err
		}

		hasArg := len(args) == 1
		hasFilterFlags := cmd.Flags().Changed("contains") ||
			cmd.Flags().Changed("type") ||
			cmd.Flags().Changed("empty") ||
			cmd.Flags().Changed("active") ||
			cmd.Flags().Changed("active")

		if hasArg && hasFilterFlags {
			return fmt.Errorf("cannot use filter flags and a queue name argument at the same time")
		}
		if !hasArg && !hasFilterFlags {
			return fmt.Errorf("must provide either filter flags or a queue name argument")
		}
		return nil
	}

	return queueCmd
}

func (q *queueCommand) delete(cmd *cobra.Command, args []string) error {
	client, err := q.getClient()
	if err != nil {
		return fmt.Errorf("could not create rabbitmq client: %w", err)
	}
	vhosts, err := vhost.Resolve(client, viper.GetStringSlice("vhosts"))
	if err != nil {
		return fmt.Errorf("parsing input vhost list '%v' for %q: %w", viper.GetStringSlice("vhosts"), client.Host(), err)
	}

	if len(args) == 1 {
		return deleteByName(cmd.OutOrStdout(), client, vhosts, args[0], q.delOpts)
	}

	return deleteByFilter(cmd.OutOrStdout(), client, vhosts, q.filter.ToQueueFilterOpts(cmd), q.delOpts)
}

func deleteByFilter(out io.Writer, c api.RabbitClient, vhosts []string, filter *api.QueueFilterOpts, deleteOpts *sharedopts.DeleteOptions) error {
	slog.Debug("applying delete filter", "filter", filter)

	var deleted, failed, wouldDelete int
	defer func() {
		if deleteOpts.DryRun {
			fmt.Fprintf(out, "\n(dry run) Total: %d would be deleted, %d failed\n", wouldDelete, failed)
		} else {
			fmt.Fprintf(out, "\nTotal: %d deleted, %d failed\n", deleted, failed)
		}
	}()

	for _, vhost := range vhosts {
		slog.Debug("processing vhost", "vhost", vhost)

		queues, err := c.ListQueuesIn(vhost)
		if err != nil {
			return fmt.Errorf("rmqctl %q: %w", c.Host(), err)
		}
		matched := api.FilterQueues(queues, filter)
		slog.Debug("matched queues", "count", len(matched))

		if deleteOpts.DryRun {
			for _, q := range matched {
				fmt.Fprintf(out, "%s: would delete queue %q\n", vhost, q.Name)
				wouldDelete++
			}
			break
		}

		d, f := deleteQueues(c, out, vhost, matched, deleteOpts.Force)
		deleted += d
		failed += f
	}

	return nil
}

func deleteQueues(c api.RabbitClient, out io.Writer, vhost string, queues []api.Queue, force bool) (deleted, failed int) {
	for _, q := range queues {
		if err := c.DeleteQueue(vhost, q.Name, force); err != nil {
			if errors.Is(err, api.ErrQueueNotSafeToDelete) {
				fmt.Fprintf(out, "%s: could not delete queue %q: %s (use --force to delete anyway)\n", vhost, q.Name, err)
			} else {
				fmt.Fprintf(out, "%s: could not delete queue %q: %s\n", vhost, q.Name, err)
			}
			failed++
			continue
		}
		fmt.Fprintf(out, "%s: deleted queue %q\n", vhost, q.Name)
		deleted++
	}
	return deleted, failed
}

func deleteByName(out io.Writer, c api.RabbitClient, vhosts []string, queueName string, deleteOpts *sharedopts.DeleteOptions) error {
	for _, vhost := range vhosts {
		slog.Debug("processing vhost", "vhost", vhost)

		if deleteOpts.DryRun {
			fmt.Fprintf(out, "%s: would delete queue %q\n", vhost, queueName)
			break
		}

		_, _ = deleteQueues(c, out, vhost, []api.Queue{{Name: queueName}}, deleteOpts.Force)
	}

	return nil
}
