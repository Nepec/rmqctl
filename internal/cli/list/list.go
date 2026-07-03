// Package list lists RabbitMQ resources
package list

import (
	"github.com/nepec/rmqctl/internal/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var listCmd = &cobra.Command{
	Use:          "list",
	Short:        "List resources in a vhost",
	Aliases:      []string{"ls", "l"},
	SilenceUsage: true,
}

func init() {
	cli.RootCmd.AddCommand(listCmd)

	listCmd.PersistentFlags().StringSlice("vhosts", []string{"/"}, "Virtual Host where to list resources from")

	_ = viper.BindPFlag("vhosts", listCmd.PersistentFlags().Lookup("vhosts"))
}
