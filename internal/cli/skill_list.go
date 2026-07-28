package cli

import (
	"sort"

	"github.com/devrimcavusoglu/skern/internal/output"
	"github.com/devrimcavusoglu/skern/internal/overlap"
	"github.com/devrimcavusoglu/skern/internal/registry"
	"github.com/devrimcavusoglu/skern/internal/skill"
	"github.com/spf13/cobra"
)

func newSkillListCmd() *cobra.Command {
	var (
		scope           string
		tag             string
		categories      []string
		includeUntagged bool
		withPlatforms   bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List skills",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := getContext(cmd)
			reg, err := ctx.NewRegistry()
			if err != nil {
				return err
			}

			categoryFilters, err := parseCategoryFilters(categories)
			if err != nil {
				return err
			}

			var skillResults []output.SkillResult

			var discovered []registry.DiscoveredSkill
			var parseWarnings []registry.ParseWarning

			if scope == "all" {
				discovered, parseWarnings, err = reg.ListAll()
				if err != nil {
					return err
				}
			} else {
				scopeVal, err := parseScope(scope)
				if err != nil {
					return err
				}
				all, pw, err := reg.ListAll()
				if err != nil {
					return err
				}
				parseWarnings = pw
				for _, d := range all {
					if d.Scope == scopeVal {
						discovered = append(discovered, d)
					}
				}
			}

			// When --with-platforms is set, build a per-scope index of installed
			// skill names per platform once, then look each registry skill up
			// against the index that matches its scope. Skipping platforms
			// that aren't detected keeps the output honest.
			var installedIndex map[skill.Scope]map[string]map[string]bool // scope -> platform -> skill -> true
			if withPlatforms {
				det, err := ctx.NewDetector()
				if err != nil {
					return err
				}
				installedIndex = make(map[skill.Scope]map[string]map[string]bool)
				for _, p := range det.DetectAll() {
					for _, sc := range []skill.Scope{skill.ScopeUser, skill.ScopeProject} {
						names, listErr := p.InstalledSkills(sc)
						if listErr != nil {
							continue
						}
						if installedIndex[sc] == nil {
							installedIndex[sc] = make(map[string]map[string]bool)
						}
						set := make(map[string]bool, len(names))
						for _, n := range names {
							set[n] = true
						}
						installedIndex[sc][string(p.Name())] = set
					}
				}
			}

			for _, d := range discovered {
				if tag != "" && !hasTag(d.Skill.Tags, tag) {
					continue
				}
				if !matchesCategories(d.Skill.Tags, categoryFilters, includeUntagged) {
					continue
				}
				r := toDiscoveredSkillResult(d)
				if files, err := skill.ListFiles(d.Path); err == nil && len(files) > 0 {
					r.Files = files
				}
				if withPlatforms {
					r.InstalledOn = lookupInstalledOn(installedIndex[d.Scope], d.Skill.Name)
				}
				skillResults = append(skillResults, r)
			}

			// Pairwise dedup detection
			var dupHints []output.DuplicateHint
			for i := 0; i < len(skillResults); i++ {
				for j := i + 1; j < len(skillResults); j++ {
					a := skillResults[i]
					b := skillResults[j]
					score := overlap.ScoreWithTools(
						a.Name, a.Description, a.AllowedTools,
						b.Name, b.Description, b.AllowedTools,
					)
					if score >= overlap.WarnThreshold {
						dupHints = append(dupHints, output.DuplicateHint{
							SkillA: a.Name,
							SkillB: b.Name,
							Score:  score,
						})
					}
				}
			}

			var pwResults []output.ParseWarningResult
			for _, w := range parseWarnings {
				pwResults = append(pwResults, output.ParseWarningResult{
					Name:  w.Name,
					Error: w.Error,
				})
			}

			result := output.SkillListResult{
				Skills:        skillResults,
				Count:         len(skillResults),
				Duplicates:    dupHints,
				ParseWarnings: pwResults,
			}
			text := formatSkillTable(skillResults)
			if len(dupHints) > 0 {
				text += formatDedupHints(dupHints)
			}
			if len(parseWarnings) > 0 {
				text += formatParseWarnings(parseWarnings)
			}
			ctx.Printer.PrintResult(result, text)
			return nil
		},
	}

	cmd.Flags().StringVar(&scope, "scope", "all", "skill scope (user, project, or all)")
	cmd.Flags().StringVar(&tag, "tag", "", "filter skills by tag")
	cmd.Flags().StringArrayVar(&categories, "category", nil, "filter by namespaced tag \"category:value\" (repeatable; comma-lists values; OR within a category, AND across categories)")
	cmd.Flags().BoolVar(&includeUntagged, "include-untagged", false, "treat a skill with no tag in a requested category as matching that category")
	cmd.Flags().BoolVar(&withPlatforms, "with-platforms", false, "include the list of detected platforms each skill is installed on")

	return cmd
}

// lookupInstalledOn returns the sorted list of platform names where the named
// skill appears in the per-platform installed-skills sets. Returns an empty
// (non-nil) slice when the skill is in no platform — this lets JSON consumers
// distinguish "queried, no platforms" from "not queried at all".
func lookupInstalledOn(byPlatform map[string]map[string]bool, name string) []string {
	out := []string{}
	for plat, set := range byPlatform {
		if set[name] {
			out = append(out, plat)
		}
	}
	sort.Strings(out)
	return out
}
