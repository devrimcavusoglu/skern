package cli

import (
	"fmt"

	"github.com/devrimcavusoglu/scribe/internal/output"
	"github.com/devrimcavusoglu/scribe/internal/skill"
	"github.com/spf13/cobra"
)

func newSkillCreateCmd() *cobra.Command {
	var (
		author         string
		authorType     string
		authorPlatform string
		description    string
		scope          string
	)

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new skill",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if err := skill.ValidateName(name); err != nil {
				return &ValidationError{Message: err.Error()}
			}

			scopeVal, err := parseScope(scope)
			if err != nil {
				return err
			}

			reg, err := newRegistryFunc()
			if err != nil {
				return err
			}

			s := skill.NewSkill(name, description, author, authorType, authorPlatform)
			path, err := reg.Create(s, scopeVal)
			if err != nil {
				return err
			}

			result := output.SkillCreateResult{
				Name:  name,
				Scope: scope,
				Path:  path,
			}
			text := fmt.Sprintf("Created skill %q in %s scope at %s\n", name, scope, path)
			printer.PrintResult(result, text)
			return nil
		},
	}

	cmd.Flags().StringVar(&author, "author", "", "author name")
	cmd.Flags().StringVar(&authorType, "author-type", "human", "author type (human or agent)")
	cmd.Flags().StringVar(&authorPlatform, "author-platform", "", "author platform (e.g., claude-code)")
	cmd.Flags().StringVar(&description, "description", "", "skill description")
	cmd.Flags().StringVar(&scope, "scope", "user", "skill scope (user or project)")

	return cmd
}
