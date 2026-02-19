package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/devrimcavusoglu/scribe/internal/output"
	"github.com/devrimcavusoglu/scribe/internal/overlap"
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
		force          bool
		fromTemplate   string
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

			// Overlap detection: check existing skills for similarity
			discovered, err := reg.ListAll()
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
						printer.PrintResult(overlapResult, text)
						return &ValidationError{Message: fmt.Sprintf("skill %q blocked due to near-duplicate (score %.2f); use --force to override", name, maxScore)}
					}

					// Warn but proceed
					text := formatOverlapWarn(name, matches)
					printer.Print("%s", text)
				}
			}

			// Skill count threshold warnings
			checkSkillCountWarnings(reg, scopeVal)

			// Read template body if --from-template is specified
			var body string
			if fromTemplate != "" {
				data, err := os.ReadFile(fromTemplate)
				if err != nil {
					return fmt.Errorf("reading template %q: %w", fromTemplate, err)
				}
				body = string(data)
			}

			s := skill.NewSkillWithBody(name, description, author, authorType, authorPlatform, body)

			// Validate on create (warnings only, don't block)
			issues := skill.Validate(s)
			if len(issues) > 0 {
				warnText := formatCreateValidationWarnings(issues)
				printer.Print("%s", warnText)
			}

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
	cmd.Flags().BoolVar(&force, "force", false, "bypass overlap detection block")
	cmd.Flags().StringVar(&fromTemplate, "from-template", "", "path to a template file for the skill body")

	return cmd
}

// Skill count thresholds
const (
	projectSkillCountWarn = 20
	userSkillCountWarn    = 50
)

func checkSkillCountWarnings(reg interface {
	List(skill.Scope) ([]skill.Skill, error)
}, scope skill.Scope) {
	skills, err := reg.List(scope)
	if err != nil {
		return
	}
	count := len(skills)

	threshold := userSkillCountWarn
	if scope == skill.ScopeProject {
		threshold = projectSkillCountWarn
	}

	if count >= threshold {
		printer.Print("Warning: %s scope has %d skills (threshold: %d). Consider reviewing for duplicates.\n", scope, count, threshold)
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
