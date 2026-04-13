package cli

import (
	"fmt"
	"runtime/debug"

	"github.com/devrimcavusoglu/skern/internal/output"
	"github.com/spf13/cobra"
)

// Version info, set via ldflags at build time.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

func init() {
	initVersionFromBuildInfo(debug.ReadBuildInfo)
}

// initVersionFromBuildInfo populates Version, Commit, and Date from Go build
// info when ldflags have not been set. The reader parameter allows tests to
// inject a fake ReadBuildInfo.
func initVersionFromBuildInfo(reader func() (*debug.BuildInfo, bool)) {
	if Version != "dev" {
		return // ldflags already set
	}
	info, ok := reader()
	if !ok {
		return
	}
	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		Version = info.Main.Version
	}
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			if s.Value != "" {
				Commit = s.Value
			}
		case "vcs.time":
			if s.Value != "" {
				Date = s.Value
			}
		}
	}
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := getContext(cmd)
			result := output.VersionResult{
				Version: Version,
				Commit:  Commit,
				Date:    Date,
			}
			text := fmt.Sprintf("skern %s (commit: %s, built: %s)\n", Version, Commit, Date)
			ctx.Printer.PrintResult(result, text)
		},
	}
}
