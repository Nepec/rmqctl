package cli

import (
	"github.com/nepec/rmqctl/internal/api"
	"github.com/nepec/rmqctl/internal/cli/policy"
	"github.com/nepec/rmqctl/internal/cli/vhost"
	"github.com/spf13/cobra"
)

func NewMergeCommand(getClient api.ClientFactory) *cobra.Command {
	opts := &policy.MergeOptions{}

	mergeCmd := &cobra.Command{
		Use:          "merge",
		Short:        "Integrate existing resource definitions in a vhost",
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		PreRunE:      vhost.BindVhosts,
	}

	mergeCmd.PersistentFlags().SortFlags = false
	mergeCmd.Flags().SortFlags = false

	mergeCmd.PersistentFlags().StringVar(&opts.DefinitionsFile, "file", "", "File containing the definitions to integrate")
	mergeCmd.PersistentFlags().BoolVarP(&opts.Force, "force", "f", false, "Force definition override from provied definitions file")
	mergeCmd.PersistentFlags().BoolVar(&opts.DryRun, "dry-run", false, "Do not apply changes, print them only")
	mergeCmd.PersistentFlags().StringSlice("vhosts", []string{"/"}, "Virtual hosts")

	// Leaf commands
	mergeCmd.AddCommand(policy.NewMergeQueuePolicy(getClient, opts))

	return mergeCmd
}
