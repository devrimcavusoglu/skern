package cli

import (
	"fmt"
	"strings"

	"github.com/devrimcavusoglu/scribe/internal/output"
	"github.com/devrimcavusoglu/scribe/internal/registry"
	"github.com/devrimcavusoglu/scribe/internal/skill"
	"github.com/spf13/cobra"
)

// Recommendation thresholds.
const (
	recommendReuseThreshold  = 0.8 // Score >= 0.8 → REUSE
	recommendExtendThreshold = 0.5 // Score >= 0.5 → EXTEND
	recommendDefaultMinScore = 0.3 // Default minimum relevance for results
)

func newSkillRecommendCmd() *cobra.Command {
	var (
		threshold float64
		scope     string
	)

	cmd := &cobra.Command{
		Use:   "recommend <query>",
		Short: "Get a recommended action for a skill need",
		Long: `Analyze existing skills and recommend whether to reuse, extend, or create a new skill.

The query should be a natural language description of what the agent needs,
e.g., "format Go source code", "run database migrations", "lint markdown files".`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]

			reg, err := newRegistryFunc()
			if err != nil {
				return err
			}

			var scored []registry.ScoredSkill

			if scope == "all" || scope == "" {
				scored, err = reg.FuzzySearch(query, threshold)
			} else {
				scopeVal, sErr := parseScope(scope)
				if sErr != nil {
					return sErr
				}
				scored, err = fuzzySearchScoped(reg, query, threshold, scopeVal)
			}
			if err != nil {
				return err
			}

			result := buildRecommendation(query, scored, threshold)
			text := formatRecommendation(result)
			printer.PrintResult(result, text)
			return nil
		},
	}

	cmd.Flags().Float64Var(&threshold, "threshold", recommendDefaultMinScore, "minimum relevance score")
	cmd.Flags().StringVar(&scope, "scope", "all", "scope to search: user, project, all")

	return cmd
}

// fuzzySearchScoped performs FuzzySearch filtered to a single scope.
func fuzzySearchScoped(reg *registry.Registry, query string, threshold float64, scope skill.Scope) ([]registry.ScoredSkill, error) {
	all, err := reg.FuzzySearch(query, threshold)
	if err != nil {
		return nil, err
	}

	var filtered []registry.ScoredSkill
	for _, s := range all {
		if s.Scope == scope {
			filtered = append(filtered, s)
		}
	}
	return filtered, nil
}

func buildRecommendation(query string, scored []registry.ScoredSkill, threshold float64) output.SkillRecommendResult {
	var matches []output.ScoredSkillResult
	for _, s := range scored {
		matches = append(matches, output.ScoredSkillResult{
			SkillResult: toDiscoveredSkillResult(s.DiscoveredSkill),
			Score:       s.Score,
		})
	}

	result := output.SkillRecommendResult{
		Query:   query,
		Matches: matches,
		Count:   len(matches),
	}

	if len(scored) == 0 {
		result.Action = output.RecommendCreate
		result.Reason = "No existing skills match your needs."
		result.SuggestedName = skill.SuggestName(query)
		return result
	}

	topScore := scored[0].Score

	switch {
	case topScore >= recommendReuseThreshold:
		result.Action = output.RecommendReuse
		result.Reason = "Existing skill closely matches your needs."
	case topScore >= recommendExtendThreshold:
		result.Action = output.RecommendExtend
		result.Reason = "Existing skill partially matches — consider extending it."
	default:
		result.Action = output.RecommendCreate
		result.Reason = "Found loosely related skills but none closely match."
		result.SuggestedName = skill.SuggestName(query)
	}

	return result
}

func formatRecommendation(r output.SkillRecommendResult) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Recommendation: %s\n\n", strings.ToUpper(string(r.Action))))
	b.WriteString(fmt.Sprintf("  Query:  %q\n", r.Query))
	b.WriteString(fmt.Sprintf("  Action: %s — %s\n", r.Action, r.Reason))

	if r.SuggestedName != "" {
		b.WriteString(fmt.Sprintf("  Suggested name: %s\n", r.SuggestedName))
	}

	if len(r.Matches) > 0 {
		b.WriteString("\n  Matching skills:\n")
		for _, m := range r.Matches {
			b.WriteString(fmt.Sprintf("    %s  (score: %.2f, scope: %s)\n", m.Name, m.Score, m.Scope))
			if m.Description != "" {
				b.WriteString(fmt.Sprintf("      %s\n", m.Description))
			}
		}
	}

	if r.Action == output.RecommendCreate && r.SuggestedName != "" {
		b.WriteString(fmt.Sprintf("\n  Run: scribe skill create %q --description %q\n", r.SuggestedName, r.Query))
	}

	return b.String()
}
