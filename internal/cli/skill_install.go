package cli

import (
	"fmt"
	"strings"

	"github.com/devrimcavusoglu/skern/internal/output"
	"github.com/devrimcavusoglu/skern/internal/platform"
	"github.com/devrimcavusoglu/skern/internal/skill"
	"github.com/spf13/cobra"
)

func newSkillInstallCmd() *cobra.Command {
	var (
		platformFlag string
		scope        string
	)

	cmd := &cobra.Command{
		Use:   "install <name>",
		Short: "Install a skill to a platform",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := getContext(cmd)
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

			reg, err := ctx.NewRegistry()
			if err != nil {
				return err
			}

			// Resolve skill from registry
			s, skillDir, err := reg.Get(name, scopeVal)
			if err != nil {
				return fmt.Errorf("skill %q not found in %s scope (run 'skern skill list' to see available skills)", name, scope)
			}
			_ = s // skill metadata not needed for install

			det, err := ctx.NewDetector()
			if err != nil {
				return err
			}

			// Determine target platforms
			var targets []platform.Platform
			if platformType == platform.TypeAll {
				targets = det.DetectAll()
				if len(targets) == 0 {
					return fmt.Errorf("no platforms detected; install a supported platform first (run 'skern platform list' to see options)")
				}
			} else {
				p := det.Get(platformType)
				if p == nil {
					return &ValidationError{Message: fmt.Sprintf("platform %q not recognized; valid platforms: claude-code, codex-cli, opencode", platformFlag)}
				}
				targets = []platform.Platform{p}
			}

			// Install to each target platform
			var entries []output.PlatformActionEntry
			var successCount int
			for _, p := range targets {
				entry := output.PlatformActionEntry{
					Platform: string(p.Name()),
				}
				if installErr := p.Install(skillDir, name, scopeVal); installErr != nil {
					entry.Error = installErr.Error()
				} else {
					entry.Success = true
					successCount++
				}
				entries = append(entries, entry)
			}

			result := output.SkillInstallResult{
				Skill:     name,
				Scope:     scope,
				Platforms: entries,
			}

			text := formatInstallResult(name, entries)
			ctx.Printer.PrintResult(result, text)

			if successCount == 0 {
				return fmt.Errorf("failed to install %q to any platform", name)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&platformFlag, "platform", "", "target platform (claude-code, codex-cli, opencode, or all)")
	cmd.Flags().StringVar(&scope, "scope", "user", "skill scope (user or project)")
	_ = cmd.MarkFlagRequired("platform")

	return cmd
}

func formatInstallResult(name string, entries []output.PlatformActionEntry) string {
	var b strings.Builder
	for _, e := range entries {
		if e.Success {
			fmt.Fprintf(&b, "Installed %q to %s\n", name, e.Platform)
		} else {
			fmt.Fprintf(&b, "Failed to install %q to %s: %s\n", name, e.Platform, e.Error)
		}
	}
	return b.String()
}
