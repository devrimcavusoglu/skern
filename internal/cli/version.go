package cli

import (
	"fmt"

	"github.com/devrimcavusoglu/scribe/internal/output"
	"github.com/spf13/cobra"
)

// Version info, set via ldflags at build time.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			result := output.VersionResult{
				Version: Version,
				Commit:  Commit,
				Date:    Date,
			}
			text := fmt.Sprintf("scribe %s (commit: %s, built: %s)\n", Version, Commit, Date)
			printer.PrintResult(result, text)
		},
	}
}
