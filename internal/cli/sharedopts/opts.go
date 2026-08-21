// Package sharedopts holds option types shared across sibling command
// packages beneath a single top-level verb command (e.g. every resource
// type under "delete").
package sharedopts

import "github.com/spf13/pflag"

// DeleteOptions configures the behavior of every "delete <resource>"
// subcommand: whether to override safety checks (Force), or only preview
// without applying (DryRun).
type DeleteOptions struct {
	Force  bool
	DryRun bool
}

// AddDeleteFlags registers the values describe in DeleteOptions
// binding their values
func AddDeleteFlags(fs *pflag.FlagSet, opts *DeleteOptions) {
	fs.BoolVarP(&opts.Force, "force", "f", false, "Force deletion of a resource bypassing safety checks.")
	fs.BoolVar(&opts.DryRun, "dry-run", false, "Only print what would be deleted")
}
