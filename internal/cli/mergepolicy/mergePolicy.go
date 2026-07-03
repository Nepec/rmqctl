// Package mergepolicy provides a command for merging RabbitMQ policy definitions.
//
// The Merge function combines active policy definitions from a RabbitMQ
// Management API response with user-supplied policy overrides, with control
// over conflict resolution (soft merge vs. force overwrite).
package mergepolicy

import (
	"fmt"

	"github.com/nepec/rmqctl/internal/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	definitionFile string
	dryRun         bool
	force          bool
)

var mergePolicyCmd = &cobra.Command{
	Use:   "merge-policy",
	Short: "Merges policy definitions in one or more vhosts",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := cmd.Root().PersistentPreRunE(cmd, args); err != nil {
			return err
		}
		if definitionFile == "" {
			return fmt.Errorf("the flag 'file' must be set")
		}
		return nil
	},
}

func init() {
	cli.RootCmd.AddCommand(mergePolicyCmd)

	mergePolicyCmd.PersistentFlags().SortFlags = false
	mergePolicyCmd.Flags().SortFlags = false

	mergePolicyCmd.PersistentFlags().StringVar(&definitionFile, "file", "", "File containing the policy definitions to integrate")
	mergePolicyCmd.PersistentFlags().StringSlice("vhosts", []string{"/"}, "Virtual hosts")
	mergePolicyCmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "Force policy override from definitions file")
	mergePolicyCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Do not apply changes, print them only")

	_ = viper.BindPFlag("vhosts", mergePolicyCmd.PersistentFlags().Lookup("vhosts"))
}
