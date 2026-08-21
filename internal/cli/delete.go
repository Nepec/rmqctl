package cli

import (
	"github.com/nepec/rmqctl/internal/api"
	"github.com/nepec/rmqctl/internal/cli/queue"
	"github.com/nepec/rmqctl/internal/cli/sharedopts"
	"github.com/nepec/rmqctl/internal/cli/vhost"
	"github.com/spf13/cobra"
)

// NewDeleteCommand builds the "delete" command, the parent for the
// imperative subcommands that delete existing RabbitMQ resource
// definitions into a vhost.
func NewDeleteCommand(getClient api.ClientFactory) *cobra.Command {
	opts := &sharedopts.DeleteOptions{}

	delCmd := &cobra.Command{
		Use:          "delete",
		Short:        "Delete resources in a vhost",
		Aliases:      []string{"del", "d"},
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		PreRunE:      vhost.BindVhosts,
	}

	delCmd.PersistentFlags().SortFlags = false
	delCmd.Flags().SortFlags = false

	sharedopts.AddDeleteFlags(delCmd.PersistentFlags(), opts)

	delCmd.PersistentFlags().StringSlice("vhosts", []string{"/"}, "Virtual hosts")

	// Leaf commands
	delCmd.AddCommand(queue.NewDeleteQueueCommand(getClient, opts))

	return delCmd
}
