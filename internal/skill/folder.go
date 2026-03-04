package skill

import (
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"
)

// ListFiles walks the skill directory and returns relative paths of all files
// except SKILL.md. Returns an empty slice for directories containing only SKILL.md.
func ListFiles(skillDir string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(skillDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(skillDir, path)
		if err != nil {
			return err
		}

		if rel != "SKILL.md" {
			files = append(files, rel)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

var (
	// Matches backtick-enclosed paths that contain a slash (to avoid false positives like `v1.0.0`).
	backtickPathRe = regexp.MustCompile("`([^`]+/[^`]+)`")
	// Matches markdown link targets, excluding URLs (http) and anchors (#).
	mdLinkRe = regexp.MustCompile(`\]\(([^)]+)\)`)
)

// ExtractFileReferences extracts path-like references from a markdown body.
// It looks for backtick-enclosed paths containing '/' and markdown link targets
// that are not URLs or anchors.
func ExtractFileReferences(body string) []string {
	seen := make(map[string]bool)
	var refs []string

	for _, m := range backtickPathRe.FindAllStringSubmatch(body, -1) {
		p := m[1]
		if !seen[p] {
			seen[p] = true
			refs = append(refs, p)
		}
	}

	for _, m := range mdLinkRe.FindAllStringSubmatch(body, -1) {
		p := m[1]
		if strings.HasPrefix(p, "http") || strings.HasPrefix(p, "#") {
			continue
		}
		if !seen[p] {
			seen[p] = true
			refs = append(refs, p)
		}
	}

	return refs
}
