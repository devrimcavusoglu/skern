package cli

import (
	"fmt"

	"github.com/devrimcavusoglu/scribe/internal/output"
	"github.com/devrimcavusoglu/scribe/internal/skill"
	"github.com/spf13/cobra"
)

func newSkillRemoveCmd() *cobra.Command {
	var scope string

	cmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a skill",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if err := skill.ValidateName(name); err != nil {
				return &ValidationError{Message: err.Error()}
			}

			reg, err := newRegistryFunc()
			if err != nil {
				return err
			}

			// Find which scope the skill is in if not specified
			resolvedScope := scope
			if scope == "" {
				// Check project first, then user
				for _, s := range []string{"project", "user"} {
					sc, _ := parseScope(s)
					if reg.Exists(name, sc) {
						resolvedScope = s
						break
					}
				}
				if resolvedScope == "" {
					return fmt.Errorf("skill %q not found", name)
				}
			}

			scopeVal, err := parseScope(resolvedScope)
			if err != nil {
				return err
			}

			if err := reg.Remove(name, scopeVal); err != nil {
				return err
			}

			result := output.SkillRemoveResult{
				Name:  name,
				Scope: resolvedScope,
			}
			text := fmt.Sprintf("Removed skill %q from %s scope\n", name, resolvedScope)
			printer.PrintResult(result, text)
			return nil
		},
	}

	cmd.Flags().StringVar(&scope, "scope", "", "skill scope (user or project)")

	return cmd
}
