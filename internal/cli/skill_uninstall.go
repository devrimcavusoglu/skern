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
		filter       skillFilter
	)

	cmd := &cobra.Command{
		Use:   "uninstall [<name>...] [--tag <tag>] [--category <cat:value>]",
		Short: "Uninstall one or more skills from a platform (registry untouched)",
		Long: `Uninstall one or more skills from a single platform.

Removes the skill from the platform's skill directory only — the skill remains
in skern's registry, and other platforms keep their copies. To delete a skill
from skern entirely, use 'skern skill remove'.

Each invocation targets exactly one platform. Multiple skill names can be
passed in one call to uninstall them as a batch — useful for evicting a set
of stale skills at once.

Instead of names, a group can be selected with --tag <tag> and/or
--category <category:value> (repeatable), resolved against the registry at
--scope and then narrowed to the skills actually installed on the platform.
Names and filters are mutually exclusive; a filter that matches nothing
installed is an error rather than a silent no-op.`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := getContext(cmd)

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
				return &ValidationError{Message: fmt.Sprintf("platform %q not recognized; valid platforms: %s", platformFlag, platformNamesList())}
			}

			names := args
			if filter.active() {
				// Registry membership defines the group; the platform's
				// installed set decides which of them there is anything to
				// remove. Tagged-but-not-installed skills are skipped, not
				// reported as failures.
				reg, regErr := ctx.NewRegistry()
				if regErr != nil {
					return regErr
				}
				names, err = resolveActionTargets(reg, scopeVal, args, &filter)
				if err != nil {
					return err
				}
				installed, listErr := p.InstalledSkills(scopeVal)
				if listErr != nil {
					return fmt.Errorf("listing installed skills on %s: %w", p.Name(), listErr)
				}
				names = intersectNames(names, installed)
				if len(names) == 0 {
					return fmt.Errorf("no installed skills match %s on %s (%s scope)", filter.describe(), p.Name(), scopeVal)
				}
			} else if len(args) == 0 {
				return &ValidationError{Message: "requires at least one skill name, or a --tag/--category filter"}
			} else {
				for _, name := range args {
					if err := skill.ValidateName(name); err != nil {
						return &ValidationError{Message: err.Error()}
					}
				}
			}

			var entries []output.SkillActionEntry
			var successCount int
			for _, name := range names {
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

	cmd.Flags().StringVar(&platformFlag, "platform", "", "target platform (one of: "+platformNamesList()+")")
	cmd.Flags().StringVar(&scope, "scope", "user", "skill scope (user or project)")
	filter.register(cmd)
	_ = cmd.MarkFlagRequired("platform")

	return cmd
}

// intersectNames keeps the entries of names that also appear in installed,
// preserving the order of names.
func intersectNames(names, installed []string) []string {
	set := make(map[string]bool, len(installed))
	for _, n := range installed {
		set[n] = true
	}
	var out []string
	for _, n := range names {
		if set[n] {
			out = append(out, n)
		}
	}
	return out
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
