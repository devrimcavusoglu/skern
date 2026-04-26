package cli

import (
	"fmt"
	"net/http"
	"time"

	"github.com/devrimcavusoglu/skern/internal/output"
	"github.com/devrimcavusoglu/skern/internal/overlap"
	"github.com/devrimcavusoglu/skern/internal/skill"
	"github.com/spf13/cobra"
)

// importHTTPTimeout caps individual HTTP requests during skill import so a
// slow or malicious endpoint cannot hang the CLI indefinitely.
const importHTTPTimeout = 30 * time.Second

func newSkillImportCmd() *cobra.Command {
	var (
		scope string
		name  string
		force bool
	)

	cmd := &cobra.Command{
		Use:   "import <url>",
		Short: "Import a skill from a remote URL",
		Long:  "Import a skill from a GitHub repository directory or gist into the local registry.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := getContext(cmd)
			rawURL := args[0]

			src, err := skill.ParseImportURL(rawURL)
			if err != nil {
				return &ValidationError{Message: err.Error()}
			}

			scopeVal, err := parseScope(scope)
			if err != nil {
				return err
			}

			// Fetch skill files from remote
			ctx.Printer.Print("Fetching skill from %s...\n", rawURL)
			httpClient := ctx.HTTPClient
			if httpClient == nil {
				httpClient = &http.Client{Timeout: importHTTPTimeout}
			}
			var fetched *skill.FetchedSkill
			if ctx.GitHubBaseURL != "" {
				fetched, err = skill.FetchSkillWithBaseURL(httpClient, src, ctx.GitHubBaseURL)
			} else {
				fetched, err = skill.FetchSkill(httpClient, src)
			}
			if err != nil {
				return fmt.Errorf("fetching skill: %w", err)
			}

			// Parse the downloaded SKILL.md
			manifestData, ok := fetched.Files["SKILL.md"]
			if !ok {
				return fmt.Errorf("no SKILL.md found in fetched files")
			}

			s, err := skill.ParseManifestFromBytes(manifestData)
			if err != nil {
				return fmt.Errorf("parsing imported manifest: %w", err)
			}

			// Apply --name override
			if name != "" {
				if err := skill.ValidateName(name); err != nil {
					return &ValidationError{Message: err.Error()}
				}
				s.Name = name
			}

			if s.Name == "" {
				return &ValidationError{Message: "imported skill has no name; use --name to specify one"}
			}

			if err := skill.ValidateName(s.Name); err != nil {
				return &ValidationError{Message: fmt.Sprintf("imported skill name is invalid: %s", err)}
			}

			reg, err := ctx.NewRegistry()
			if err != nil {
				return err
			}

			// Overlap detection
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

				matches := overlap.Check(s.Name, s.Description, existing, scopes)
				if len(matches) > 0 {
					blocked := overlap.ShouldBlock(matches) && !force

					if blocked {
						maxScore := overlap.MaxScore(matches)
						text := formatOverlapBlock(s.Name, matches)
						overlapResult := output.OverlapCheckResult{
							Blocked:  true,
							MaxScore: maxScore,
						}
						for _, m := range matches {
							overlapResult.Matches = append(overlapResult.Matches, output.OverlapResult{
								Name:  m.Name,
								Score: m.Score,
								Scope: string(m.Scope),
							})
						}
						ctx.Printer.PrintResult(overlapResult, text)
						return &ValidationError{Message: fmt.Sprintf("skill %q blocked due to near-duplicate (score %.2f); use --force to override", s.Name, maxScore)}
					}

					// Warn but proceed
					text := formatOverlapWarn(s.Name, matches)
					ctx.Printer.Print("%s", text)
				}
			}

			// Import into registry
			path, err := reg.Import(s, fetched.Files, scopeVal, force)
			if err != nil {
				return err
			}

			result := output.SkillImportResult{
				Name:   s.Name,
				Scope:  scope,
				Path:   path,
				Source: rawURL,
			}
			text := fmt.Sprintf("Imported skill %q from %s into %s scope at %s\n", s.Name, formatSource(src), scope, path)
			ctx.Printer.PrintResult(result, text)
			return nil
		},
	}

	cmd.Flags().StringVar(&scope, "scope", "user", "skill scope (user or project)")
	cmd.Flags().StringVar(&name, "name", "", "override the skill name from the manifest")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite if skill already exists and bypass overlap block")

	return cmd
}

func formatSource(src *skill.ImportSource) string {
	switch src.Type {
	case skill.SourceGitHubRepo:
		return fmt.Sprintf("github.com/%s/%s", src.Owner, src.Repo)
	case skill.SourceGitHubGist:
		return fmt.Sprintf("gist %s", src.GistID)
	default:
		return src.RawURL
	}
}
