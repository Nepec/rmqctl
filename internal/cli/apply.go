package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/nepec/rmqctl/internal/api"
	"github.com/nepec/rmqctl/internal/cli/vhost"
	"github.com/nepec/rmqctl/internal/manifest"
	"github.com/nepec/rmqctl/internal/reconcile"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type applyCommand struct {
	getClient api.ClientFactory

	// Flags
	manifestFile string
	dryRun       bool
}

// NewApplyCommand builds the "apply" command, which parses a manifest
// file and declaratively provisions its resources, via getClient, on
// every selected vhost. With --dry-run, nothing is provisioned and the
// resulting definitions are printed instead.
func NewApplyCommand(getClient api.ClientFactory) *cobra.Command {
	c := applyCommand{getClient: getClient}

	applyCmd := &cobra.Command{
		Use:          "apply",
		Short:        "Apply resources from a manifest file",
		Aliases:      []string{"a"},
		SilenceUsage: true,
		PreRunE:      vhost.BindVhosts,
		RunE:         c.apply,
	}

	applyCmd.PersistentFlags().SortFlags = false
	applyCmd.Flags().SortFlags = false

	applyCmd.Flags().StringVar(&c.manifestFile, "file", "", "Manifest file containing resource specs")
	applyCmd.Flags().BoolVar(&c.dryRun, "dry-run", false, "Do not apply changes, print them only")
	applyCmd.Flags().StringSlice("vhosts", []string{"/"}, "Virtual Host where to apply resources to")

	return applyCmd
}

func (a *applyCommand) apply(cmd *cobra.Command, _ []string) error {
	client, err := a.getClient()
	if err != nil {
		return fmt.Errorf("could not create rabbitmq client: %w", err)
	}
	vhosts, err := vhost.Resolve(client, viper.GetStringSlice("vhosts"))
	if err != nil {
		return fmt.Errorf("parsing input vhost list '%v': %w", viper.GetStringSlice("vhosts"), err)
	}

	return execute(cmd.OutOrStdout(), client, vhosts, a.manifestFile, a.dryRun)
}

func execute(out io.Writer, client api.RabbitClient, vhosts []string, file string, dryRun bool) error {
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
		err = reconcile.Apply(client, out, v, man.Spec.MarshalDefs(v), dryRun)
		if err != nil {
			return err
		}

	}

	return nil
}
