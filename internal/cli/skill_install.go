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
		platformFlag  string
		scope         string
		force         bool
		enforceBudget bool
	)

	cmd := &cobra.Command{
		Use:   "install <name>...",
		Short: "Install one or more registered skills onto a platform",
		Long: `Install one or more skills from skern's registry onto a single platform.

Each skill must already exist in skern's registry — create it with 'skern skill
create' or pull it in with 'skern skill import' before installing. Install copies
the registered skill into the platform's skill directory; the registry copy is
left untouched.

Each invocation targets exactly one platform — agents are expected to specify
the platform they are running on. Multiple skill names can be passed in one
call to install them as a batch.

When --enforce-budget is set, install refuses to proceed if the resulting
installed-skill count would meet or exceed the per-platform threshold (see
'skern platform status' for current capacity).`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := getContext(cmd)

			// Validate every skill name up-front so we fail fast.
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

			reg, err := ctx.NewRegistry()
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

			// Capacity pre-check: if --enforce-budget is set, refuse the entire
			// batch when the resulting count would exceed the threshold. This
			// is intentionally strict — agents that hit this should evict
			// stale skills first or invoke without --enforce-budget.
			if enforceBudget {
				pre := buildCapacityReport(p, scopeVal)
				if pre != nil && pre.Installed+len(args) > pre.Threshold {
					return fmt.Errorf("capacity: %s (%s) has %d/%d skills installed; installing %d more would exceed the threshold (uninstall stale skills or drop --enforce-budget to proceed)",
						pre.Platform, pre.Scope, pre.Installed, pre.Threshold, len(args))
				}
			}

			// Resolve skills and perform installs in input order. Failures
			// for one skill do not abort the batch — each gets its own entry
			// so the agent can react per-skill.
			var entries []output.SkillActionEntry
			var successCount int
			for _, name := range args {
				entry := output.SkillActionEntry{Skill: name}

				_, skillDir, getErr := reg.Get(name, scopeVal)
				if getErr != nil {
					entry.Error = fmt.Sprintf("not registered in skern (%s scope) — run 'skern skill create' or 'skern skill import' first, or 'skern skill list' to see available skills", scope)
					entries = append(entries, entry)
					continue
				}

				if force {
					// Best-effort cleanup: skill may not be installed yet.
					_ = p.Uninstall(name, scopeVal)
				}

				if installErr := p.Install(skillDir, name, scopeVal); installErr != nil {
					entry.Error = installErr.Error()
				} else {
					entry.Success = true
					successCount++
				}
				entries = append(entries, entry)
			}

			capacity := buildCapacityReport(p, scopeVal)

			result := output.SkillInstallResult{
				Platform: string(p.Name()),
				Scope:    scope,
				Skills:   entries,
				Capacity: capacity,
			}

			text := formatInstallResult(string(p.Name()), entries) + formatCapacityWarning(capacity)
			ctx.Printer.PrintResult(result, text)

			if successCount == 0 {
				return fmt.Errorf("failed to install any skill to %s", p.Name())
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&platformFlag, "platform", "", "target platform (one of: "+platformNamesList()+")")
	cmd.Flags().StringVar(&scope, "scope", "user", "skill scope (user or project)")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing installation")
	cmd.Flags().BoolVar(&enforceBudget, "enforce-budget", false, "refuse to install when at or over capacity")
	_ = cmd.MarkFlagRequired("platform")

	return cmd
}

func formatInstallResult(platformName string, entries []output.SkillActionEntry) string {
	var b strings.Builder
	for _, e := range entries {
		if e.Success {
			fmt.Fprintf(&b, "Installed %q to %s\n", e.Skill, platformName)
		} else {
			fmt.Fprintf(&b, "Failed to install %q to %s: %s\n", e.Skill, platformName, e.Error)
		}
	}
	return b.String()
}
