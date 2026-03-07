package cli

import (
	"github.com/devrimcavusoglu/skern/internal/output"
	"github.com/spf13/cobra"
)

func newSkillSearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search for skills by name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := getContext(cmd)
			query := args[0]

			reg, err := ctx.NewRegistry()
			if err != nil {
				return err
			}

			discovered, err := reg.Search(query)
			if err != nil {
				return err
			}

			var skillResults []output.SkillResult
			for _, d := range discovered {
				skillResults = append(skillResults, toDiscoveredSkillResult(d))
			}

			result := output.SkillSearchResult{
				Query:   query,
				Results: skillResults,
				Count:   len(skillResults),
			}
			text := formatSearchResults(query, skillResults)
			ctx.Printer.PrintResult(result, text)
			return nil
		},
	}

	return cmd
}
