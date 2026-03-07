package cli

import (
	"fmt"
	"strings"

	"github.com/devrimcavusoglu/skern/internal/output"
	"github.com/devrimcavusoglu/skern/internal/skill"
	"github.com/spf13/cobra"
)

func newSkillValidateCmd() *cobra.Command {
	var scope string

	cmd := &cobra.Command{
		Use:   "validate <name>",
		Short: "Validate a skill against the Agent Skills spec",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := getContext(cmd)
			name := args[0]

			reg, err := ctx.NewRegistry()
			if err != nil {
				return err
			}

			s, skillDir, _, err := resolveSkill(reg, name, scope)
			if err != nil {
				return err
			}

			issues := skill.Validate(s)
			issues = append(issues, skill.ValidateFolder(s, skillDir)...)
			result := toValidateResult(name, issues)
			text := formatValidateResult(result)
			ctx.Printer.PrintResult(result, text)

			if skill.HasErrors(issues) {
				return &ValidationError{Message: fmt.Sprintf("skill %q has validation errors", name)}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&scope, "scope", "", "skill scope (user or project)")

	return cmd
}

func toValidateResult(name string, issues []skill.ValidationIssue) output.SkillValidateResult {
	var issueResults []output.ValidationIssueResult
	errors := 0
	warns := 0
	hints := 0

	for _, issue := range issues {
		issueResults = append(issueResults, output.ValidationIssueResult{
			Field:    issue.Field,
			Severity: string(issue.Severity),
			Message:  issue.Message,
		})
		switch issue.Severity {
		case skill.SeverityError:
			errors++
		case skill.SeverityWarning:
			warns++
		case skill.SeverityHint:
			hints++
		}
	}

	return output.SkillValidateResult{
		Name:   name,
		Valid:  errors == 0,
		Issues: issueResults,
		Errors: errors,
		Warns:  warns,
		Hints:  hints,
	}
}

func formatValidateResult(r output.SkillValidateResult) string {
	var b strings.Builder

	if r.Valid && len(r.Issues) == 0 {
		fmt.Fprintf(&b, "Skill %q is valid.\n", r.Name)
		return b.String()
	}

	if r.Valid && r.Warns == 0 && r.Hints > 0 {
		fmt.Fprintf(&b, "Skill %q is valid with %d hint(s):\n", r.Name, r.Hints)
	} else if r.Valid {
		fmt.Fprintf(&b, "Skill %q is valid with %d warning(s), %d hint(s):\n", r.Name, r.Warns, r.Hints)
	} else {
		fmt.Fprintf(&b, "Skill %q has %d error(s), %d warning(s), %d hint(s):\n", r.Name, r.Errors, r.Warns, r.Hints)
	}

	for _, issue := range r.Issues {
		prefix := "  ✗"
		switch issue.Severity {
		case "warning":
			prefix = "  !"
		case "hint":
			prefix = "  ~"
		}
		fmt.Fprintf(&b, "%s %s: %s\n", prefix, issue.Field, issue.Message)
	}

	return b.String()
}
