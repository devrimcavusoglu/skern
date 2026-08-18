package cli

import (
	"github.com/spf13/cobra"
)

const (
	skillGroupRegistry = "registry"
	skillGroupPlatform = "platform"
)

func newSkillCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skill",
		Short: "Manage Agent Skills",
		Long: `Manage Agent Skills in skern's registry, then install registered skills onto specific platforms (` + platformNamesList() + `).

Skills live platform-independent in skern's registry — under ~/.skern/skills/ (user scope) or .skern/skills/ (project scope). Use ` + "`skern skill install`" + ` to copy a registered skill into a platform's skill directory; ` + "`skern skill uninstall`" + ` removes it from a platform without affecting the registry. To delete a skill from skern entirely, use ` + "`skern skill remove`" + `.`,
	}

	cmd.AddGroup(
		&cobra.Group{ID: skillGroupRegistry, Title: "Registry commands (manage skills in skern):"},
		&cobra.Group{ID: skillGroupPlatform, Title: "Platform commands (deploy skills to platforms):"},
	)

	registry := []*cobra.Command{
		newSkillCreateCmd(),
		newSkillImportCmd(),
		newSkillEditCmd(),
		newSkillRemoveCmd(),
		newSkillListCmd(),
		newSkillShowCmd(),
		newSkillSearchCmd(),
		newSkillValidateCmd(),
		newSkillVersionCmd(),
		newSkillDiffCmd(),
		newSkillRecommendCmd(),
	}
	for _, c := range registry {
		c.GroupID = skillGroupRegistry
		cmd.AddCommand(c)
	}

	platform := []*cobra.Command{
		newSkillInstallCmd(),
		newSkillUninstallCmd(),
	}
	for _, c := range platform {
		c.GroupID = skillGroupPlatform
		cmd.AddCommand(c)
	}

	return cmd
}
