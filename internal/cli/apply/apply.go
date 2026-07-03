package apply

import (
	"fmt"
	"io"
	"os"

	"github.com/nepec/rmqctl/internal/api"
	"github.com/nepec/rmqctl/internal/cli"
	"github.com/nepec/rmqctl/internal/manifest"
	"github.com/nepec/rmqctl/internal/reconcile"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	dryRun       bool
	manifestFile string
)

var applyCmd = &cobra.Command{
	Use:     "apply",
	Short:   "Apply resources from a manifest file",
	Aliases: []string{"a"},
	RunE: func(cmd *cobra.Command, args []string) error {
		adapter, err := cli.ClientFromConfig()
		if err != nil {
			return fmt.Errorf("init cmd %q: %w", cmd.Name(), err)
		}

		vhosts, err := cli.ResolveVhosts(adapter, viper.GetStringSlice("vhosts"))
		if err != nil {
			return fmt.Errorf("parsing input vhost list '%v': %w", viper.GetStringSlice("vhosts"), err)
		}

		return applyAction(os.Stdout, adapter, vhosts, manifestFile, dryRun)
	},
	SilenceUsage: true,
}

func init() {
	cli.RootCmd.AddCommand(applyCmd)

	applyCmd.PersistentFlags().StringSlice("vhosts", []string{"/"}, "Virtual Host where to apply resources to")
	applyCmd.PersistentFlags().StringVar(&manifestFile, "file", "", "Manifest file containing resource specs")
	applyCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Do not apply changes, print them only")

	_ = viper.BindPFlag("vhosts", applyCmd.PersistentFlags().Lookup("vhosts"))
}

// applyAction reads a manifest from input (either file or stdin)
// it parses the manifest to produce a manifest obj of a certain kind and with a certain spec
// it then calls the apply method of the parsed spec
func applyAction(out io.Writer, c api.RabbitClient, vhosts []string, file string, dryRun bool) error {
	// #nosec G304 -- definitionFile is a user-supplied CLI flag, not untrusted network input
	rawManifest, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	man, err := manifest.Parse(rawManifest)
	if err != nil {
		return err
	}

	for _, v := range vhosts {
		err = reconcile.Apply(c, out, v, man.Spec.MarshalDefs(v), dryRun)
		if err != nil {
			return err
		}

	}

	return nil
}
