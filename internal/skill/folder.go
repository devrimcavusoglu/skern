package skill

import (
	"fmt"
	"io/fs"
	"path"
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

// looksLikeURL reports whether s looks like a URL rather than a file path.
// It checks for protocol prefixes (http://, https://, ftp://) and domain-like
// patterns (e.g., example.com/path).
func looksLikeURL(s string) bool {
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") || strings.HasPrefix(s, "ftp://") {
		return true
	}
	// Check for domain-like pattern: word with dot before the first slash
	// e.g., "example.com/path", "docs.example.io/guide"
	slashIdx := strings.Index(s, "/")
	if slashIdx > 0 {
		prefix := s[:slashIdx]
		if strings.Contains(prefix, ".") && !strings.HasPrefix(prefix, ".") && !strings.HasPrefix(prefix, "..") {
			return true
		}
	}
	return false
}

// ExtractFileReferences extracts path-like references from a markdown body.
// It looks for backtick-enclosed paths containing '/' and markdown link targets
// that are not URLs or anchors.
func ExtractFileReferences(body string) []string {
	seen := make(map[string]bool)
	var refs []string

	for _, m := range backtickPathRe.FindAllStringSubmatch(body, -1) {
		p := m[1]
		if looksLikeURL(p) {
			continue
		}
		if !seen[p] {
			seen[p] = true
			refs = append(refs, p)
		}
	}

	for _, m := range mdLinkRe.FindAllStringSubmatch(body, -1) {
		p := m[1]
		if looksLikeURL(p) || strings.HasPrefix(p, "#") {
			continue
		}
		if !seen[p] {
			seen[p] = true
			refs = append(refs, p)
		}
	}

	return refs
}

// ValidateExcludePattern checks one `install.exclude` entry. Patterns are
// slash-separated paths relative to the skill directory using path.Match
// syntax (`*`, `?`, `[...]`, `\` escapes a metacharacter; no `**`).
// Rejected: empty, absolute, any `..` segment, a `**` segment, a malformed
// glob, and the literal SKILL.md. (A wildcard that merely also matches
// SKILL.md — `*`, `*.md` — is allowed: MatchExclude never excludes the
// manifest, so such a pattern means "everything except SKILL.md".)
func ValidateExcludePattern(pattern string) error {
	p := normalizeExcludePattern(pattern)
	if p == "" {
		return fmt.Errorf("exclude pattern must not be empty")
	}
	if strings.HasPrefix(p, "/") || filepath.IsAbs(pattern) {
		return fmt.Errorf("exclude pattern %q must be relative to the skill directory", pattern)
	}
	for _, seg := range strings.Split(p, "/") {
		if seg == ".." {
			return fmt.Errorf("exclude pattern %q must not contain \"..\"", pattern)
		}
		if strings.Contains(seg, "**") {
			return fmt.Errorf("exclude pattern %q: \"**\" is not supported (a bare directory name already excludes its whole subtree; `*` does not cross \"/\")", pattern)
		}
	}
	if _, err := path.Match(p, "x"); err != nil {
		return fmt.Errorf("exclude pattern %q is not a valid glob: %v", pattern, err)
	}
	if p == ManifestFile {
		return fmt.Errorf("exclude pattern %q names %s, which is always installed", pattern, ManifestFile)
	}
	return nil
}

// ValidateExcludePatterns validates every pattern and returns the first
// error, prefixed with the pattern's index for multi-entry lists.
func ValidateExcludePatterns(patterns []string) error {
	for i, p := range patterns {
		if err := ValidateExcludePattern(p); err != nil {
			if len(patterns) > 1 {
				return fmt.Errorf("install.exclude[%d]: %w", i, err)
			}
			return fmt.Errorf("install.exclude: %w", err)
		}
	}
	return nil
}

// normalizeExcludePattern trims whitespace and strips a leading "./" and any
// trailing "/" (a directory pattern written as "eval/" means the same as
// "eval"). Backslashes are left alone: patterns are slash-separated and `\`
// is path.Match's escape character.
func normalizeExcludePattern(pattern string) string {
	p := strings.TrimSpace(pattern)
	p = strings.TrimPrefix(p, "./")
	p = strings.TrimRight(p, "/")
	return p
}

// MatchExclude reports whether rel — a slash-separated path relative to the
// skill directory — is excluded by any of patterns. A pattern matches a path
// when path.Match matches the full relative path or any of its leading
// directory prefixes, so `eval` excludes `eval/` and everything beneath it,
// `fixtures/*` excludes the direct children of fixtures/, and `*.test.md`
// excludes top-level files with that suffix. SKILL.md is never excluded.
func MatchExclude(patterns []string, rel string) bool {
	rel = filepath.ToSlash(rel)
	if rel == "" || rel == "." || rel == ManifestFile {
		return false
	}
	for _, raw := range patterns {
		p := normalizeExcludePattern(raw)
		if p == "" {
			continue
		}
		// Walk the path prefixes: "a/b/c" -> "a", "a/b", "a/b/c".
		for i := 0; i <= len(rel); i++ {
			if i == len(rel) || rel[i] == '/' {
				if ok, _ := path.Match(p, rel[:i]); ok {
					return true
				}
			}
		}
	}
	return false
}

// ExcludedFiles returns the relative paths under skillDir (excluding
// SKILL.md) that MatchExclude would drop for patterns, in walk order. Used by
// the validator to warn about patterns that match nothing.
func ExcludedFiles(skillDir string, patterns []string) ([]string, error) {
	files, err := ListFiles(skillDir)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, f := range files {
		if MatchExclude(patterns, f) {
			out = append(out, f)
		}
	}
	return out, nil
}
