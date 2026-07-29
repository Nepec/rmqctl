package cli

import (
	"github.com/nepec/rmqctl/internal/api"
	"github.com/nepec/rmqctl/internal/cli/queue"
	"github.com/nepec/rmqctl/internal/cli/vhost"
	"github.com/spf13/cobra"
)

func NewListCommand(getClient api.ClientFactory) *cobra.Command {
	listCmd := &cobra.Command{
		Use:          "list",
		Short:        "List resources in a vhost",
		Aliases:      []string{"ls", "l"},
		SilenceUsage: true,
		Args:         cobra.NoArgs,
	}

	listCmd.AddCommand(queue.NewListQueueCommand(getClient))
	listCmd.AddCommand(vhost.NewListVhostCommand(getClient))

	return listCmd
}
