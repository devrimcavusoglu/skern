package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/devrimcavusoglu/skern/internal/output"
	"github.com/devrimcavusoglu/skern/internal/skill"
	"github.com/spf13/cobra"
)

func newSkillEditCmd() *cobra.Command {
	var (
		scope              string
		description        string
		author             string
		authorType         string
		authorPlatform     string
		version            string
		modifiedByName     string
		modifiedByType     string
		modifiedByPlatform string
	)

	cmd := &cobra.Command{
		Use:   "edit <name>",
		Short: "Edit a skill's metadata or body",
		Long:  "Update skill metadata via flags, or open $EDITOR to edit the SKILL.md file when no flags are provided.",
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

			manifestPath := filepath.Join(skillDir, "SKILL.md")

			hasFieldFlags := cmd.Flags().Changed("description") ||
				cmd.Flags().Changed("author") ||
				cmd.Flags().Changed("author-type") ||
				cmd.Flags().Changed("author-platform") ||
				cmd.Flags().Changed("version")

			if !hasFieldFlags {
				// Open $EDITOR
				editor := os.Getenv("EDITOR")
				if editor == "" {
					editor = "vi"
				}

				editorCmd := exec.Command(editor, manifestPath)
				editorCmd.Stdin = os.Stdin
				editorCmd.Stdout = os.Stdout
				editorCmd.Stderr = os.Stderr
				if err := editorCmd.Run(); err != nil {
					return fmt.Errorf("editor exited with error: %w", err)
				}

				// Re-parse the manifest after editing
				edited, err := skill.ParseManifest(manifestPath)
				if err != nil {
					return fmt.Errorf("parsing edited manifest: %w", err)
				}

				// Append modified-by if requested
				if modifiedByName != "" {
					appendModifiedBy(edited, modifiedByName, modifiedByType, modifiedByPlatform)
					if err := skill.WriteManifest(edited, manifestPath); err != nil {
						return fmt.Errorf("writing modified-by: %w", err)
					}
				}

				result := output.SkillEditResult{
					Name:    name,
					Scope:   string(foundScope),
					Updated: []string{"body"},
				}
				text := fmt.Sprintf("Edited skill %q via $EDITOR.\n", name)
				ctx.Printer.PrintResult(result, text)
				return nil
			}

			// Apply field updates
			var updated []string

			if cmd.Flags().Changed("description") {
				s.Description = description
				updated = append(updated, "description")
			}
			if cmd.Flags().Changed("author") {
				s.Metadata.Author.Name = author
				updated = append(updated, "author.name")
			}
			if cmd.Flags().Changed("author-type") {
				s.Metadata.Author.Type = authorType
				updated = append(updated, "author.type")
			}
			if cmd.Flags().Changed("author-platform") {
				s.Metadata.Author.Platform = authorPlatform
				updated = append(updated, "author.platform")
			}
			if cmd.Flags().Changed("version") {
				s.Metadata.Version = version
				updated = append(updated, "version")
			}

			// Append modified-by if requested
			if modifiedByName != "" {
				appendModifiedBy(s, modifiedByName, modifiedByType, modifiedByPlatform)
				updated = append(updated, "modified-by")
			}

			if err := skill.WriteManifest(s, manifestPath); err != nil {
				return fmt.Errorf("writing manifest: %w", err)
			}

			result := output.SkillEditResult{
				Name:    name,
				Scope:   string(foundScope),
				Updated: updated,
			}
			text := fmt.Sprintf("Updated skill %q: %s\n", name, strings.Join(updated, ", "))
			ctx.Printer.PrintResult(result, text)
			return nil
		},
	}

	cmd.Flags().StringVar(&scope, "scope", "", "skill scope (user or project)")
	cmd.Flags().StringVar(&description, "description", "", "new description")
	cmd.Flags().StringVar(&author, "author", "", "new author name")
	cmd.Flags().StringVar(&authorType, "author-type", "", "new author type (human or agent)")
	cmd.Flags().StringVar(&authorPlatform, "author-platform", "", "new author platform")
	cmd.Flags().StringVar(&version, "version", "", "new version")
	cmd.Flags().StringVar(&modifiedByName, "modified-by", "", "name of the modifier")
	cmd.Flags().StringVar(&modifiedByType, "modified-by-type", "human", "type of the modifier (human or agent)")
	cmd.Flags().StringVar(&modifiedByPlatform, "modified-by-platform", "", "platform of the modifier")

	return cmd
}

func appendModifiedBy(s *skill.Skill, name, typ, platform string) {
	s.Metadata.ModifiedBy = append(s.Metadata.ModifiedBy, skill.ModifiedByEntry{
		Name:     name,
		Type:     typ,
		Platform: platform,
		Date:     time.Now().UTC().Format(time.RFC3339),
	})
}
