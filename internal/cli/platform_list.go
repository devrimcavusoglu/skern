package cli

import (
	"fmt"
	"strings"

	"github.com/devrimcavusoglu/scribe/internal/output"
	"github.com/spf13/cobra"
)

func newPlatformListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List known platforms and their detection status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			det, err := newDetectorFunc()
			if err != nil {
				return err
			}

			all := det.All()
			var platforms []output.PlatformResult
			for _, p := range all {
				platforms = append(platforms, output.PlatformResult{
					Name:        string(p.Name()),
					Detected:    p.Detect(),
					UserPath:    p.UserSkillsDir(),
					ProjectPath: p.ProjectSkillsDir(),
				})
			}

			result := output.PlatformListResult{
				Platforms: platforms,
				Count:     len(platforms),
			}

			text := formatPlatformList(platforms)
			printer.PrintResult(result, text)
			return nil
		},
	}

	return cmd
}

func formatPlatformList(platforms []output.PlatformResult) string {
	if len(platforms) == 0 {
		return "No platforms registered.\n"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%-15s %-10s %-40s %s\n", "PLATFORM", "DETECTED", "USER PATH", "PROJECT PATH")
	for _, p := range platforms {
		detected := "no"
		if p.Detected {
			detected = "yes"
		}
		fmt.Fprintf(&b, "%-15s %-10s %-40s %s\n", p.Name, detected, p.UserPath, p.ProjectPath)
	}
	return b.String()
}
