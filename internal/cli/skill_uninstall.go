package cli

import (
	"fmt"
	"strings"

	"github.com/devrimcavusoglu/scribe/internal/output"
	"github.com/devrimcavusoglu/scribe/internal/platform"
	"github.com/devrimcavusoglu/scribe/internal/skill"
	"github.com/spf13/cobra"
)

func newSkillUninstallCmd() *cobra.Command {
	var (
		platformFlag string
		scope        string
	)

	cmd := &cobra.Command{
		Use:   "uninstall <name>",
		Short: "Uninstall a skill from a platform",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if err := skill.ValidateName(name); err != nil {
				return &ValidationError{Message: err.Error()}
			}

			platformType, err := platform.ParsePlatformType(platformFlag)
			if err != nil {
				return &ValidationError{Message: err.Error()}
			}

			scopeVal, err := parseScope(scope)
			if err != nil {
				return err
			}

			det, err := newDetectorFunc()
			if err != nil {
				return err
			}

			// Determine target platforms
			var targets []platform.Platform
			if platformType == platform.Type("all") {
				targets = det.DetectAll()
				if len(targets) == 0 {
					return fmt.Errorf("no platforms detected; install a supported platform first (run 'scribe platform list' to see options)")
				}
			} else {
				p := det.Get(platformType)
				if p == nil {
					return &ValidationError{Message: fmt.Sprintf("platform %q not recognized; valid platforms: claude-code, codex-cli, opencode", platformFlag)}
				}
				targets = []platform.Platform{p}
			}

			// Uninstall from each target platform
			var entries []output.PlatformUninstallEntry
			var successCount int
			for _, p := range targets {
				entry := output.PlatformUninstallEntry{
					Platform: string(p.Name()),
				}
				if uninstallErr := p.Uninstall(name, scopeVal); uninstallErr != nil {
					entry.Error = uninstallErr.Error()
				} else {
					entry.Success = true
					successCount++
				}
				entries = append(entries, entry)
			}

			result := output.SkillUninstallResult{
				Skill:     name,
				Scope:     string(scopeVal),
				Platforms: entries,
			}

			text := formatUninstallResult(name, entries)
			printer.PrintResult(result, text)

			if successCount == 0 {
				return fmt.Errorf("failed to uninstall %q from any platform", name)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&platformFlag, "platform", "", "target platform (claude-code, codex-cli, opencode, or all)")
	cmd.Flags().StringVar(&scope, "scope", "user", "skill scope (user or project)")
	_ = cmd.MarkFlagRequired("platform")

	return cmd
}

func formatUninstallResult(name string, entries []output.PlatformUninstallEntry) string {
	var b strings.Builder
	for _, e := range entries {
		if e.Success {
			b.WriteString(fmt.Sprintf("Uninstalled %q from %s\n", name, e.Platform))
		} else {
			b.WriteString(fmt.Sprintf("Failed to uninstall %q from %s: %s\n", name, e.Platform, e.Error))
		}
	}
	return b.String()
}
