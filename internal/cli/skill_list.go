package cli

import (
	"github.com/devrimcavusoglu/scribe/internal/output"
	"github.com/devrimcavusoglu/scribe/internal/skill"
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

			result := output.SkillListResult{
				Skills: skillResults,
				Count:  len(skillResults),
			}
			text := formatSkillTable(skillResults)
			printer.PrintResult(result, text)
			return nil
		},
	}

	cmd.Flags().StringVar(&scope, "scope", "all", "skill scope (user, project, or all)")

	return cmd
}
