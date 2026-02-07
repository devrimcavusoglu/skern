package cli

import (
	"github.com/spf13/cobra"
)

func newSkillShowCmd() *cobra.Command {
	var scope string

	cmd := &cobra.Command{
		Use:   "show <name>",
		Short: "Show skill details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			reg, err := newRegistryFunc()
			if err != nil {
				return err
			}

			s, path, foundScope, err := resolveSkill(reg, name, scope)
			if err != nil {
				return err
			}

			result := toSkillResult(s, string(foundScope), path)
			text := formatSkillShow(result)
			printer.PrintResult(result, text)
			return nil
		},
	}

	cmd.Flags().StringVar(&scope, "scope", "", "skill scope (user or project)")

	return cmd
}
