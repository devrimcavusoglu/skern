package cli

import (
	"github.com/spf13/cobra"
)

func newSkillCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skill",
		Short: "Manage Agent Skills",
		Long:  "Create, list, show, search, and remove Agent Skills in user or project scope.",
	}

	cmd.AddCommand(newSkillCreateCmd())
	cmd.AddCommand(newSkillListCmd())
	cmd.AddCommand(newSkillShowCmd())
	cmd.AddCommand(newSkillSearchCmd())
	cmd.AddCommand(newSkillRemoveCmd())

	return cmd
}
