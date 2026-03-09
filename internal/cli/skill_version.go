package cli

import (
	"fmt"
	"path/filepath"

	"github.com/devrimcavusoglu/skern/internal/output"
	"github.com/devrimcavusoglu/skern/internal/skill"
	"github.com/spf13/cobra"
)

func newSkillVersionCmd() *cobra.Command {
	var (
		scope string
		bump  string
	)

	cmd := &cobra.Command{
		Use:   "version <name>",
		Short: "Show or bump a skill's version",
		Long:  "Display the current version of a skill, or bump it with --bump patch|minor|major.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := getContext(cmd)
			name := args[0]

			reg, err := ctx.NewRegistry()
			if err != nil {
				return err
			}

			s, skillDir, foundScope, err := resolveSkill(reg, name, scope)
			if err != nil {
				return err
			}

			if bump == "" {
				// Show current version
				result := output.SkillVersionResult{
					Name:    name,
					Version: s.Metadata.Version,
					Scope:   string(foundScope),
					Bumped:  false,
				}
				text := fmt.Sprintf("%s\n", s.Metadata.Version)
				ctx.Printer.PrintResult(result, text)
				return nil
			}

			// Validate bump level
			if bump != "patch" && bump != "minor" && bump != "major" {
				return &ValidationError{Message: fmt.Sprintf("invalid bump level %q: must be patch, minor, or major", bump)}
			}

			previousVersion := s.Metadata.Version
			newVersion, err := skill.BumpVersion(previousVersion, bump)
			if err != nil {
				return fmt.Errorf("bumping version: %w", err)
			}

			s.Metadata.Version = newVersion
			manifestPath := filepath.Join(skillDir, "SKILL.md")
			if err := skill.WriteManifest(s, manifestPath); err != nil {
				return fmt.Errorf("writing manifest: %w", err)
			}

			result := output.SkillVersionResult{
				Name:            name,
				Version:         newVersion,
				Scope:           string(foundScope),
				PreviousVersion: previousVersion,
				Bumped:          true,
			}
			text := fmt.Sprintf("Bumped %q from %s to %s (%s)\n", name, previousVersion, newVersion, bump)
			ctx.Printer.PrintResult(result, text)
			return nil
		},
	}

	cmd.Flags().StringVar(&scope, "scope", "", "skill scope (user or project)")
	cmd.Flags().StringVar(&bump, "bump", "", "bump level: patch, minor, or major")

	return cmd
}
