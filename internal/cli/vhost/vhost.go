package vhost

import (
	"fmt"
	"io"
	"log/slog"
	"text/tabwriter"

	"github.com/nepec/rmqctl/internal/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type vhostCommand struct {
	getClient api.ClientFactory
}

func NewListVhostCommand(getClient api.ClientFactory) *cobra.Command {
	c := vhostCommand{getClient: getClient}

	vhostsCmd := &cobra.Command{
		Use:     "vhosts",
		Short:   "List virtual hosts in a node",
		Aliases: []string{"vhost", "vh"},
		Args:    cobra.NoArgs,
		RunE:    c.execute,
	}

	vhostsCmd.PersistentFlags().SortFlags = false
	vhostsCmd.Flags().SortFlags = false

	return vhostsCmd
}

func (v *vhostCommand) execute(cmd *cobra.Command, _ []string) error {
	client, err := v.getClient()
	if err != nil {
		return fmt.Errorf("could not create rabbitmq client: %w", err)
	}
	vhosts, err := client.ListVhosts()
	if err != nil {
		return fmt.Errorf("rmqctl %q: %w", client.Host(), err)
	}

	defer func() {
		formatter := vhostTableFormatter{}
		formatter.Print(cmd.OutOrStdout(), vhosts)
	}()

	return nil
}

type vhostTableFormatter struct{}

func (v vhostTableFormatter) Print(out io.Writer, results []api.Vhost) {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	defer func() {
		_ = w.Flush()
	}()

	count := 0
	fmt.Fprintln(w, "NAME")
	for _, r := range results {
		fmt.Fprintf(w, "%s\n", r.Name)
		count++
	}
	fmt.Fprintf(out, "\nFound %d vhosts\n", count)
	fmt.Fprintln(out, "")
}

func Resolve(c api.RabbitClient, requested []string) ([]string, error) {
	if len(requested) == 1 && requested[0] == "*" {
		slog.Debug("wildcard '*' detected, fetching all vhosts from broker...")
		vs, err := c.ListVhosts()
		if err != nil {
			return nil, err
		}

		vhosts := make([]string, 0, len(vs))
		for _, v := range vs {
			vhosts = append(vhosts, v.Name)
		}

		return vhosts, nil
	}

	return requested, nil
}

// BindVhosts wires the command's --vhosts flag into viper at runtime,
// avoiding cross-command flag binding collisions.
func BindVhosts(cmd *cobra.Command, _ []string) error {
	return viper.BindPFlag("vhosts", cmd.Flags().Lookup("vhosts"))
}
