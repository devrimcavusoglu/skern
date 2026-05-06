package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/devrimcavusoglu/skern/internal/output"
	"github.com/devrimcavusoglu/skern/internal/overlap"
	"github.com/devrimcavusoglu/skern/internal/registry"
	"github.com/devrimcavusoglu/skern/internal/skill"
	"github.com/spf13/cobra"
)

func newSkillCreateCmd() *cobra.Command {
	var (
		author         string
		authorType     string
		authorPlatform string
		description    string
		scope          string
		force          bool
		fromTemplate   string
		tags           []string
		version        string
	)

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new skill in skern's registry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := getContext(cmd)
			name := args[0]

			if err := skill.ValidateName(name); err != nil {
				return &ValidationError{Message: err.Error()}
			}

			scopeVal, err := parseScope(scope)
			if err != nil {
				return err
			}

			reg, err := ctx.NewRegistry()
			if err != nil {
				return err
			}

			// Overlap detection: check existing skills for similarity
			discovered, _, err := reg.ListAll()
			if err != nil {
				return fmt.Errorf("checking for overlapping skills: %w", err)
			}

			if len(discovered) > 0 {
				var existing []skill.Skill
				var scopes []skill.Scope
				for _, d := range discovered {
					existing = append(existing, d.Skill)
					scopes = append(scopes, d.Scope)
				}

				matches := overlap.Check(name, description, existing, scopes)
				if len(matches) > 0 {
					maxScore := overlap.MaxScore(matches)
					blocked := overlap.ShouldBlock(matches) && !force

					overlapResult := output.OverlapCheckResult{
						Blocked:  blocked,
						MaxScore: maxScore,
					}
					for _, m := range matches {
						overlapResult.Matches = append(overlapResult.Matches, output.OverlapResult{
							Name:  m.Name,
							Score: m.Score,
							Scope: string(m.Scope),
						})
					}

					if blocked {
						text := formatOverlapBlock(name, matches)
						ctx.Printer.PrintResult(overlapResult, text)
						return &ValidationError{Message: fmt.Sprintf("skill %q blocked due to near-duplicate (score %.2f); use --force to override", name, maxScore)}
					}

					// Warn but proceed
					text := formatOverlapWarn(name, matches)
					ctx.Printer.Print("%s", text)
				}
			}

			// Skill count threshold warnings
			checkSkillCountWarnings(ctx.Printer, reg, scopeVal)

			// Resolve --from-template (a SKILL.md file, a raw-body markdown
			// file, or a skill directory containing SKILL.md and siblings).
			tmpl, err := loadTemplate(fromTemplate)
			if err != nil {
				return err
			}

			s := buildSkillFromTemplate(tmpl, name, description, author, authorType, authorPlatform, tags, cmd)

			if version != "" {
				if _, err := skill.ParseVersion(version); err != nil {
					return &ValidationError{Message: err.Error()}
				}
				s.Metadata.Version = version
			}

			// Validate on create (warnings only, don't block)
			issues := skill.Validate(s)
			if len(issues) > 0 {
				warnText := formatCreateValidationWarnings(issues)
				ctx.Printer.Print("%s", warnText)
			}

			path, err := reg.Create(s, scopeVal)
			if err != nil {
				return err
			}

			// If the template was a skill directory, copy sibling assets
			// (references/, templates/, VENDORED.md, ...) alongside the
			// freshly written SKILL.md. On failure, roll back the partial
			// skill so the registry doesn't keep a half-populated entry.
			if tmpl != nil && tmpl.sourceDir != "" {
				if err := registry.CopySiblings(tmpl.sourceDir, path); err != nil {
					_ = os.RemoveAll(path)
					return fmt.Errorf("copying template assets from %q: %w", tmpl.sourceDir, err)
				}
			}

			result := output.SkillCreateResult{
				Name:  name,
				Scope: scope,
				Path:  path,
			}
			text := fmt.Sprintf("Created skill %q in %s scope at %s\n", name, scope, path)
			ctx.Printer.PrintResult(result, text)
			return nil
		},
	}

	cmd.Flags().StringVar(&author, "author", "", "author name")
	cmd.Flags().StringVar(&authorType, "author-type", "human", "author type (human or agent)")
	cmd.Flags().StringVar(&authorPlatform, "author-platform", "", "author platform (e.g., claude-code)")
	cmd.Flags().StringVar(&description, "description", "", "skill description")
	cmd.Flags().StringVar(&scope, "scope", "user", "skill scope (user or project)")
	cmd.Flags().BoolVar(&force, "force", false, "bypass overlap detection block")
	cmd.Flags().StringVar(&fromTemplate, "from-template", "", "path to a skill directory (containing SKILL.md and optional companion files) to seed the new skill from")
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "comma-separated tags for the skill")
	cmd.Flags().StringVar(&version, "version", "", "initial version (default: 0.0.1)")

	return cmd
}

// checkSkillCountWarnings emits a warning when a registry scope is at or above
// its capacity threshold (see internal/skill/capacity.go for values).
func checkSkillCountWarnings(p *output.Printer, reg interface {
	List(skill.Scope) ([]skill.Skill, []registry.ParseWarning, error)
}, scope skill.Scope) {
	skills, _, err := reg.List(scope)
	if err != nil {
		return
	}
	count := len(skills)
	threshold := skill.ScopeThreshold(scope)

	if count >= threshold {
		p.Print("Warning: %s scope has %d skills (threshold: %d). Consider reviewing for duplicates.\n", scope, count, threshold)
	}
}

func formatOverlapBlock(name string, matches []overlap.Match) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Skill %q blocked — near-duplicate detected:\n", name)
	for _, m := range matches {
		fmt.Fprintf(&b, "  - %s (score: %.2f, scope: %s)\n", m.Name, m.Score, m.Scope)
	}
	b.WriteString("Use --force to override.\n")
	return b.String()
}

func formatOverlapWarn(name string, matches []overlap.Match) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Warning: skill %q has similar existing skills:\n", name)
	for _, m := range matches {
		fmt.Fprintf(&b, "  - %s (score: %.2f, scope: %s)\n", m.Name, m.Score, m.Scope)
	}
	b.WriteString("Proceeding with creation...\n")
	return b.String()
}

func formatCreateValidationWarnings(issues []skill.ValidationIssue) string {
	var b strings.Builder
	for _, issue := range issues {
		prefix := "  !"
		if issue.Severity == skill.SeverityError {
			prefix = "  ✗"
		}
		fmt.Fprintf(&b, "%s %s: %s\n", prefix, issue.Field, issue.Message)
	}
	return b.String()
}

// templateInput captures the resolved `--from-template` source. The flag
// only accepts a skill *directory* containing a SKILL.md; sourceDir is
// always set for use by registry.CopySiblings, and parsed holds the
// frontmatter parsed from <sourceDir>/SKILL.md.
type templateInput struct {
	sourceDir string
	parsed    *skill.Skill
}

func loadTemplate(path string) (*templateInput, error) {
	if path == "" {
		return nil, nil
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("reading template %q: %w", path, err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf(
			"--from-template must point to a skill directory containing a SKILL.md file, "+
				"but %q is a file; pass the parent directory instead",
			path,
		)
	}

	manifestPath := filepath.Join(path, skill.ManifestFile)
	manifestInfo, err := os.Stat(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf(
				"--from-template directory %q has no SKILL.md; a skill template must be a directory containing a SKILL.md file",
				path,
			)
		}
		return nil, fmt.Errorf("reading template SKILL.md from directory %q: %w", path, err)
	}
	if manifestInfo.IsDir() {
		return nil, fmt.Errorf("--from-template directory %q contains a SKILL.md entry that is itself a directory; expected a regular file", path)
	}

	parsed, err := skill.ParseManifest(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("parsing template SKILL.md in directory %q: %w", path, err)
	}
	return &templateInput{sourceDir: path, parsed: parsed}, nil
}

// buildSkillFromTemplate constructs the new Skill, applying CLI flag values on
// top of any template-derived defaults. The CLI <name> argument always wins
// over the template's name. Other flags only override when the user
// explicitly set them, so an unspecified --description on a SKILL.md template
// preserves the template's description rather than stamping the placeholder.
func buildSkillFromTemplate(
	tmpl *templateInput,
	name, description, author, authorType, authorPlatform string,
	tags []string,
	cmd *cobra.Command,
) *skill.Skill {
	if tmpl == nil {
		s := skill.NewSkillWithBody(name, description, author, authorType, authorPlatform, "")
		s.Tags = tags
		return s
	}

	s := tmpl.parsed
	s.Name = name

	flags := cmd.Flags()
	if flags.Changed("description") {
		s.Description = description
	}
	if flags.Changed("author") {
		s.Metadata.Author.Name = author
	}
	if flags.Changed("author-type") {
		s.Metadata.Author.Type = authorType
	}
	if flags.Changed("author-platform") {
		s.Metadata.Author.Platform = authorPlatform
	}
	if flags.Changed("tags") {
		s.Tags = tags
	}

	// Fill in fields the template didn't provide so a freshly written
	// manifest still has sane defaults (matches NewSkillWithBody's behavior).
	if strings.TrimSpace(s.Description) == "" {
		s.Description = "Use when TODO: describe the triggering conditions for this skill.\n"
	}
	if s.Metadata.Author.Type == "" {
		s.Metadata.Author.Type = "human"
	}
	if s.Metadata.Version == "" {
		s.Metadata.Version = "0.0.1"
	}

	return s
}
