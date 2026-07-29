package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type BuildInfo struct {
	Version string
}

type versionCommand struct {
	cmd *cobra.Command

	info BuildInfo
}

func NewVersionCommand(info BuildInfo) *cobra.Command {
	c := &versionCommand{info: info}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Display the rmqctl version.",
		Args:  cobra.NoArgs,
		RunE:  c.execute,
	}

	c.cmd = versionCmd

	return versionCmd
}

func (c *versionCommand) execute(_ *cobra.Command, _ []string) error {
	_, err := fmt.Fprintf(os.Stdout, "rmqctl - %s", c.info.Version)
	return err
}
