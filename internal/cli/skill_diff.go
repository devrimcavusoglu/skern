package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/devrimcavusoglu/skern/internal/output"
	"github.com/devrimcavusoglu/skern/internal/platform"
	"github.com/devrimcavusoglu/skern/internal/skill"
	"github.com/spf13/cobra"
)

func newSkillDiffCmd() *cobra.Command {
	var (
		scope        string
		platformFlag string
	)

	cmd := &cobra.Command{
		Use:   "diff <name> [name-b]",
		Short: "Compare two skills or a registry skill against its installed copy",
		Long: `Compare two skills side by side.

With one argument, compares a registry skill against its installed copy on a platform
(requires --platform and --scope flags).

With two arguments, compares two registry skills by name.`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := getContext(cmd)

			if len(args) == 2 {
				return diffTwoSkills(ctx, args[0], args[1], scope)
			}

			return diffRegistryVsPlatform(ctx, args[0], scope, platformFlag)
		},
	}

	cmd.Flags().StringVar(&scope, "scope", "", "skill scope (user or project)")
	cmd.Flags().StringVar(&platformFlag, "platform", "", "platform to compare against (claude-code, codex-cli, opencode)")

	return cmd
}

// diffTwoSkills compares two registry skills by name.
func diffTwoSkills(ctx *CommandContext, nameA, nameB, scopeStr string) error {
	reg, err := ctx.NewRegistry()
	if err != nil {
		return err
	}

	skillA, _, scopeA, err := resolveSkill(reg, nameA, scopeStr)
	if err != nil {
		return fmt.Errorf("resolving skill %q: %w", nameA, err)
	}

	skillB, _, scopeB, err := resolveSkill(reg, nameB, scopeStr)
	if err != nil {
		return fmt.Errorf("resolving skill %q: %w", nameB, err)
	}

	sourceA := fmt.Sprintf("registry (%s)", scopeA)
	sourceB := fmt.Sprintf("registry (%s)", scopeB)

	result := compareSkills(skillA, nameA, sourceA, skillB, nameB, sourceB)
	text := formatDiffResult(result)
	ctx.Printer.PrintResult(result, text)
	return nil
}

// diffRegistryVsPlatform compares a registry skill against its installed copy on a platform.
func diffRegistryVsPlatform(ctx *CommandContext, name, scopeStr, platformFlag string) error {
	if platformFlag == "" {
		return &ValidationError{Message: "comparing a registry skill against a platform requires --platform flag"}
	}

	if scopeStr == "" {
		scopeStr = "user"
	}

	scopeVal, err := parseScope(scopeStr)
	if err != nil {
		return err
	}

	platformType, err := platform.ParsePlatformType(platformFlag)
	if err != nil {
		return &ValidationError{Message: err.Error()}
	}

	if platformType == platform.TypeAll {
		return &ValidationError{Message: "diff requires a specific platform, not \"all\""}
	}

	reg, err := ctx.NewRegistry()
	if err != nil {
		return err
	}

	registrySkill, _, err := reg.Get(name, scopeVal)
	if err != nil {
		return fmt.Errorf("skill %q not found in %s scope: %w", name, scopeStr, err)
	}

	det, err := ctx.NewDetector()
	if err != nil {
		return err
	}

	p := det.Get(platformType)
	if p == nil {
		return &ValidationError{Message: fmt.Sprintf("platform %q not recognized", platformFlag)}
	}

	var platformDir string
	if scopeVal == skill.ScopeProject {
		platformDir = p.ProjectSkillsDir()
	} else {
		platformDir = p.UserSkillsDir()
	}

	manifestPath := filepath.Join(platformDir, name, "SKILL.md")
	platformSkill, err := skill.ParseManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("skill %q not installed on %s (%s scope): %w", name, platformFlag, scopeStr, err)
	}

	sourceA := fmt.Sprintf("registry (%s)", scopeStr)
	sourceB := fmt.Sprintf("platform (%s)", platformFlag)

	result := compareSkills(registrySkill, name, sourceA, platformSkill, name, sourceB)
	text := formatDiffResult(result)
	ctx.Printer.PrintResult(result, text)
	return nil
}

// compareSkills compares two skills and produces a SkillDiffResult.
func compareSkills(a *skill.Skill, nameA, sourceA string, b *skill.Skill, nameB, sourceB string) output.SkillDiffResult {
	var fields []output.FieldDiff

	if a.Name != b.Name {
		fields = append(fields, output.FieldDiff{Field: "name", Left: a.Name, Right: b.Name})
	}

	descA := strings.TrimSpace(a.Description)
	descB := strings.TrimSpace(b.Description)
	if descA != descB {
		fields = append(fields, output.FieldDiff{Field: "description", Left: descA, Right: descB})
	}

	if a.Metadata.Version != b.Metadata.Version {
		fields = append(fields, output.FieldDiff{Field: "version", Left: a.Metadata.Version, Right: b.Metadata.Version})
	}

	if a.Metadata.Author.Name != b.Metadata.Author.Name {
		fields = append(fields, output.FieldDiff{Field: "author.name", Left: a.Metadata.Author.Name, Right: b.Metadata.Author.Name})
	}
	if a.Metadata.Author.Type != b.Metadata.Author.Type {
		fields = append(fields, output.FieldDiff{Field: "author.type", Left: a.Metadata.Author.Type, Right: b.Metadata.Author.Type})
	}
	if a.Metadata.Author.Platform != b.Metadata.Author.Platform {
		fields = append(fields, output.FieldDiff{Field: "author.platform", Left: a.Metadata.Author.Platform, Right: b.Metadata.Author.Platform})
	}

	tagsA := strings.Join(a.Tags, ", ")
	tagsB := strings.Join(b.Tags, ", ")
	if tagsA != tagsB {
		fields = append(fields, output.FieldDiff{Field: "tags", Left: tagsA, Right: tagsB})
	}

	toolsA := strings.Join(a.AllowedTools, ", ")
	toolsB := strings.Join(b.AllowedTools, ", ")
	if toolsA != toolsB {
		fields = append(fields, output.FieldDiff{Field: "allowed-tools", Left: toolsA, Right: toolsB})
	}

	bodyDiff := a.Body != b.Body

	result := output.SkillDiffResult{
		LeftName:    nameA,
		LeftSource:  sourceA,
		RightName:   nameB,
		RightSource: sourceB,
		Identical:   len(fields) == 0 && !bodyDiff,
		Fields:      fields,
		BodyDiff:    bodyDiff,
	}

	if bodyDiff {
		result.LeftBody = a.Body
		result.RightBody = b.Body
	}

	return result
}

// formatDiffResult formats a diff result for text output.
func formatDiffResult(r output.SkillDiffResult) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Comparing %s (%s) vs %s (%s)\n\n", r.LeftName, r.LeftSource, r.RightName, r.RightSource)

	if r.Identical {
		b.WriteString("Skills are identical.\n")
		return b.String()
	}

	if len(r.Fields) > 0 {
		b.WriteString("Metadata differences:\n")
		for _, f := range r.Fields {
			fmt.Fprintf(&b, "  %s:\n", f.Field)
			fmt.Fprintf(&b, "    - %s\n", displayValue(f.Left))
			fmt.Fprintf(&b, "    + %s\n", displayValue(f.Right))
		}
	}

	if r.BodyDiff {
		if len(r.Fields) > 0 {
			b.WriteString("\n")
		}
		b.WriteString("Body content differs.\n")
	}

	return b.String()
}

// displayValue returns the value or "(empty)" if blank.
func displayValue(v string) string {
	if v == "" {
		return "(empty)"
	}
	return v
}
