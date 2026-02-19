package cli

import (
	"fmt"
	"strings"

	"github.com/devrimcavusoglu/scribe/internal/output"
	"github.com/spf13/cobra"
)

func newPlatformStatusCmd() *cobra.Command {
	var scope string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show skill installation status across platforms",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			scopeVal, err := parseScope(scope)
			if err != nil {
				return err
			}

			reg, err := newRegistryFunc()
			if err != nil {
				return err
			}

			det, err := newDetectorFunc()
			if err != nil {
				return err
			}

			// Get skills from registry
			skills, err := reg.List(scopeVal)
			if err != nil {
				return fmt.Errorf("listing skills: %w", err)
			}

			detected := det.DetectAll()

			// Build installed skills index per platform
			type platformSkills struct {
				name      string
				installed map[string]bool
			}
			var platformIndex []platformSkills
			for _, p := range detected {
				names, listErr := p.InstalledSkills(scopeVal)
				if listErr != nil {
					continue
				}
				installed := make(map[string]bool)
				for _, n := range names {
					installed[n] = true
				}
				platformIndex = append(platformIndex, platformSkills{
					name:      string(p.Name()),
					installed: installed,
				})
			}

			// Build status entries
			var entries []output.PlatformStatusEntry
			for _, s := range skills {
				entry := output.PlatformStatusEntry{
					Skill: s.Name,
				}
				for _, pi := range platformIndex {
					entry.Platforms = append(entry.Platforms, output.PlatformInstallStatus{
						Platform:  pi.name,
						Installed: pi.installed[s.Name],
					})
				}
				entries = append(entries, entry)
			}

			result := output.PlatformStatusResult{
				Scope:  scope,
				Status: entries,
			}

			text := formatPlatformStatus(scope, entries)
			printer.PrintResult(result, text)
			return nil
		},
	}

	cmd.Flags().StringVar(&scope, "scope", "user", "skill scope (user or project)")

	return cmd
}

func formatPlatformStatus(scope string, entries []output.PlatformStatusEntry) string {
	if len(entries) == 0 {
		return fmt.Sprintf("No skills in %s scope.\n", scope)
	}

	// Collect platform names from the first entry
	var platformNames []string
	if len(entries[0].Platforms) > 0 {
		for _, p := range entries[0].Platforms {
			platformNames = append(platformNames, p.Platform)
		}
	}

	if len(platformNames) == 0 {
		return fmt.Sprintf("No platforms detected. Skills in %s scope: %d\n", scope, len(entries))
	}

	var b strings.Builder
	// Header
	fmt.Fprintf(&b, "%-30s", "SKILL")
	for _, name := range platformNames {
		fmt.Fprintf(&b, " %-15s", name)
	}
	b.WriteString("\n")

	// Rows
	for _, entry := range entries {
		fmt.Fprintf(&b, "%-30s", entry.Skill)
		for _, p := range entry.Platforms {
			status := "-"
			if p.Installed {
				status = "installed"
			}
			fmt.Fprintf(&b, " %-15s", status)
		}
		b.WriteString("\n")
	}

	return b.String()
}
