package cli

import (
	"github.com/devrimcavusoglu/skern/internal/output"
	"github.com/devrimcavusoglu/skern/internal/overlap"
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
			reg, err := newRegistryFunc()
			if err != nil {
				return err
			}

			var skillResults []output.SkillResult

			if scope == "all" {
				discovered, err := reg.ListAll()
				if err != nil {
					return err
				}
				for _, d := range discovered {
					skillResults = append(skillResults, toDiscoveredSkillResult(d))
				}
			} else {
				scopeVal, err := parseScope(scope)
				if err != nil {
					return err
				}
				skills, err := reg.List(scopeVal)
				if err != nil {
					return err
				}
				dir := ""
				if scopeVal == skill.ScopeUser {
					dir = "user"
				} else {
					dir = "project"
				}
				for _, s := range skills {
					skillResults = append(skillResults, toSkillResult(&s, dir, ""))
				}
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

			result := output.SkillListResult{
				Skills:     skillResults,
				Count:      len(skillResults),
				Duplicates: dupHints,
			}
			text := formatSkillTable(skillResults)
			if len(dupHints) > 0 {
				text += formatDedupHints(dupHints)
			}
			printer.PrintResult(result, text)
			return nil
		},
	}

	cmd.Flags().StringVar(&scope, "scope", "all", "skill scope (user, project, or all)")

	return cmd
}
