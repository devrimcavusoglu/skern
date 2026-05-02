package cli

import (
	"fmt"
	"strings"

	"github.com/devrimcavusoglu/skern/internal/output"
	"github.com/devrimcavusoglu/skern/internal/platform"
	"github.com/devrimcavusoglu/skern/internal/skill"
	"github.com/spf13/cobra"
)

func newSkillUninstallCmd() *cobra.Command {
	var (
		platformFlag string
		scope        string
	)

	cmd := &cobra.Command{
		Use:   "uninstall <name>...",
		Short: "Uninstall one or more skills from a platform",
		Long: `Uninstall one or more skills from a single platform.

Each invocation targets exactly one platform. Multiple skill names can be
passed in one call to uninstall them as a batch — useful for evicting a set
of stale skills at once.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := getContext(cmd)

			for _, name := range args {
				if err := skill.ValidateName(name); err != nil {
					return &ValidationError{Message: err.Error()}
				}
			}

			platformType, err := platform.ParsePlatformType(platformFlag)
			if err != nil {
				return &ValidationError{Message: err.Error()}
			}

			scopeVal, err := parseScope(scope)
			if err != nil {
				return err
			}

			det, err := ctx.NewDetector()
			if err != nil {
				return err
			}

			p := det.Get(platformType)
			if p == nil {
				return &ValidationError{Message: fmt.Sprintf("platform %q not recognized; valid platforms: claude-code, codex-cli, opencode", platformFlag)}
			}

			var entries []output.SkillActionEntry
			var successCount int
			for _, name := range args {
				entry := output.SkillActionEntry{Skill: name}
				if uninstallErr := p.Uninstall(name, scopeVal); uninstallErr != nil {
					entry.Error = uninstallErr.Error()
				} else {
					entry.Success = true
					successCount++
				}
				entries = append(entries, entry)
			}

			capacity := buildCapacityReport(p, scopeVal)

			result := output.SkillUninstallResult{
				Platform: string(p.Name()),
				Scope:    string(scopeVal),
				Skills:   entries,
				Capacity: capacity,
			}

			text := formatUninstallResult(string(p.Name()), entries) + formatCapacityWarning(capacity)
			ctx.Printer.PrintResult(result, text)

			if successCount == 0 {
				return fmt.Errorf("failed to uninstall any skill from %s", p.Name())
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&platformFlag, "platform", "", "target platform (claude-code, codex-cli, or opencode)")
	cmd.Flags().StringVar(&scope, "scope", "user", "skill scope (user or project)")
	_ = cmd.MarkFlagRequired("platform")

	return cmd
}

func formatUninstallResult(platformName string, entries []output.SkillActionEntry) string {
	var b strings.Builder
	for _, e := range entries {
		if e.Success {
			fmt.Fprintf(&b, "Uninstalled %q from %s\n", e.Skill, platformName)
		} else {
			fmt.Fprintf(&b, "Failed to uninstall %q from %s: %s\n", e.Skill, platformName, e.Error)
		}
	}
	return b.String()
}
