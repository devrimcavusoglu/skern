package cli

import (
	"github.com/devrimcavusoglu/skern/internal/output"
	"github.com/devrimcavusoglu/skern/internal/overlap"
	"github.com/devrimcavusoglu/skern/internal/registry"
	"github.com/devrimcavusoglu/skern/internal/skill"
	"github.com/spf13/cobra"
)

func newSkillListCmd() *cobra.Command {
	var scope string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List skills",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := getContext(cmd)
			reg, err := ctx.NewRegistry()
			if err != nil {
				return err
			}

			var skillResults []output.SkillResult

			var discovered []registry.DiscoveredSkill
			var parseWarnings []registry.ParseWarning

			if scope == "all" {
				discovered, parseWarnings, err = reg.ListAll()
				if err != nil {
					return err
				}
			} else {
				scopeVal, err := parseScope(scope)
				if err != nil {
					return err
				}
				all, pw, err := reg.ListAll()
				if err != nil {
					return err
				}
				parseWarnings = pw
				for _, d := range all {
					if d.Scope == scopeVal {
						discovered = append(discovered, d)
					}
				}
			}

			for _, d := range discovered {
				r := toDiscoveredSkillResult(d)
				if files, err := skill.ListFiles(d.Path); err == nil && len(files) > 0 {
					r.Files = files
				}
				skillResults = append(skillResults, r)
			}

			// Pairwise dedup detection
			var dupHints []output.DuplicateHint
			for i := 0; i < len(skillResults); i++ {
				for j := i + 1; j < len(skillResults); j++ {
					a := skillResults[i]
					b := skillResults[j]
					score := overlap.ScoreWithTools(
						a.Name, a.Description, a.AllowedTools,
						b.Name, b.Description, b.AllowedTools,
					)
					if score >= overlap.WarnThreshold {
						dupHints = append(dupHints, output.DuplicateHint{
							SkillA: a.Name,
							SkillB: b.Name,
							Score:  score,
						})
					}
				}
			}

			var pwResults []output.ParseWarningResult
			for _, w := range parseWarnings {
				pwResults = append(pwResults, output.ParseWarningResult{
					Name:  w.Name,
					Error: w.Error,
				})
			}

			result := output.SkillListResult{
				Skills:        skillResults,
				Count:         len(skillResults),
				Duplicates:    dupHints,
				ParseWarnings: pwResults,
			}
			text := formatSkillTable(skillResults)
			if len(dupHints) > 0 {
				text += formatDedupHints(dupHints)
			}
			if len(parseWarnings) > 0 {
				text += formatParseWarnings(parseWarnings)
			}
			ctx.Printer.PrintResult(result, text)
			return nil
		},
	}

	cmd.Flags().StringVar(&scope, "scope", "all", "skill scope (user, project, or all)")

	return cmd
}
