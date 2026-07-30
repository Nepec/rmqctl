package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// BuildInfo carries build-time metadata into the CLI, such as the
// version string set via ldflags at release time.
type BuildInfo struct {
	Version string
}

type versionCommand struct {
	cmd *cobra.Command

	info BuildInfo
}

// NewVersionCommand builds the "version" command, which prints info.Version.
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
