package list

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/nepec/rmqctl/internal/api"
	"github.com/nepec/rmqctl/internal/cli"
	"github.com/spf13/cobra"
)

var vhostsCmd = &cobra.Command{
	Use:     "vhosts",
	Short:   "List virtual hosts in a node",
	Aliases: []string{"vhost"},
	RunE: func(cmd *cobra.Command, args []string) error {
		adapter, err := cli.ClientFromConfig()
		if err != nil {
			return err
		}

		return listVhostsAction(os.Stdout, adapter)
	},
}

func init() {
	listCmd.AddCommand(vhostsCmd)
}

func listVhostsAction(out io.Writer, c api.RabbitClient) error {
	vhosts, err := c.ListVhosts()
	if err != nil {
		return fmt.Errorf("rmqctl %q: %w", c.Host(), err)
	}

	defer func() {
		formatter := VhostTableFormatter{}
		formatter.Print(out, vhosts)
	}()

	return nil
}

type VhostTableFormatter struct{}

func (v VhostTableFormatter) Print(out io.Writer, results []api.Vhost) {
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
