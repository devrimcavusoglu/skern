package cli

import (
	"github.com/devrimcavusoglu/skern/internal/skill"
	"github.com/spf13/cobra"
)

func newSkillShowCmd() *cobra.Command {
	var scope string

	cmd := &cobra.Command{
		Use:   "show <name>",
		Short: "Show skill details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := getContext(cmd)
			name := args[0]

			reg, err := ctx.NewRegistry()
			if err != nil {
				return err
			}

			s, path, foundScope, err := resolveSkill(reg, name, scope)
			if err != nil {
				return err
			}

			result := toSkillResult(s, string(foundScope), path)
			if files, err := skill.ListFiles(path); err == nil && len(files) > 0 {
				result.Files = files
			}
			text := formatSkillShow(result)
			ctx.Printer.PrintResult(result, text)
			return nil
		},
	}

	cmd.Flags().StringVar(&scope, "scope", "", "skill scope (user or project)")

	return cmd
}
