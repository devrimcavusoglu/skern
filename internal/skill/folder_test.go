package skill

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListFiles(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(dir string)
		expected []string
		wantErr  bool
	}{
		{
			name: "only SKILL.md",
			setup: func(dir string) {
				require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("test"), 0o644))
			},
			expected: nil,
		},
		{
			name: "with scripts",
			setup: func(dir string) {
				require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("test"), 0o644))
				require.NoError(t, os.MkdirAll(filepath.Join(dir, "scripts"), 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "scripts", "convert.py"), []byte("# python"), 0o644))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "scripts", "setup.sh"), []byte("#!/bin/bash"), 0o644))
			},
			expected: []string{
				filepath.Join("scripts", "convert.py"),
				filepath.Join("scripts", "setup.sh"),
			},
		},
		{
			name: "with nested dirs",
			setup: func(dir string) {
				require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("test"), 0o644))
				require.NoError(t, os.MkdirAll(filepath.Join(dir, "assets", "templates"), 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "assets", "templates", "main.json"), []byte("{}"), 0o644))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("readme"), 0o644))
			},
			expected: []string{
				"README.md",
				filepath.Join("assets", "templates", "main.json"),
			},
		},
		{
			name:    "nonexistent dir",
			setup:   func(dir string) {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			skillDir := filepath.Join(dir, "my-skill")

			if tt.name != "nonexistent dir" {
				require.NoError(t, os.MkdirAll(skillDir, 0o755))
				tt.setup(skillDir)
			}

			files, err := ListFiles(skillDir)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, files)
		})
	}
}

func TestExtractFileReferences(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expected []string
	}{
		{
			name:     "backtick paths",
			body:     "Use `scripts/convert.py` to transform data and `assets/template.json` for config.",
			expected: []string{"scripts/convert.py", "assets/template.json"},
		},
		{
			name:     "markdown links",
			body:     "See [converter](scripts/convert.py) and [template](assets/template.json).",
			expected: []string{"scripts/convert.py", "assets/template.json"},
		},
		{
			name:     "URLs skipped",
			body:     "See [docs](https://example.com) and [API](http://api.example.com).",
			expected: nil,
		},
		{
			name:     "anchors skipped",
			body:     "See [section](#overview) for details.",
			expected: nil,
		},
		{
			name:     "no refs",
			body:     "Simple instructions without any file references.",
			expected: nil,
		},
		{
			name:     "backtick without slash ignored",
			body:     "Use version `v1.0.0` for stability.",
			expected: nil,
		},
		{
			name:     "deduplication",
			body:     "Use `scripts/run.py` first, then run `scripts/run.py` again.",
			expected: []string{"scripts/run.py"},
		},
		{
			name:     "mixed backtick and link",
			body:     "Run `scripts/run.py` or see [script](scripts/run.py).",
			expected: []string{"scripts/run.py"},
		},
		{
			name:     "backtick URL with protocol skipped",
			body:     "See `https://example.com/docs/guide` for details.",
			expected: nil,
		},
		{
			name:     "backtick URL without protocol skipped",
			body:     "Use `example.com/api/v1` as the endpoint.",
			expected: nil,
		},
		{
			name:     "backtick domain-like paths skipped",
			body:     "Check `docs.example.io/guide` and `pkg.go.dev/encoding/json` for reference.",
			expected: nil,
		},
		{
			name:     "markdown link with domain-like URL skipped",
			body:     "See [docs](docs.example.com/guide) for more info.",
			expected: nil,
		},
		{
			name:     "backtick URL mixed with real path",
			body:     "Use `example.com/api` for the API and `scripts/run.py` for processing.",
			expected: []string{"scripts/run.py"},
		},
		{
			name:     "relative path with dots not skipped",
			body:     "Use `./scripts/run.py` and `../lib/utils.py` as helpers.",
			expected: []string{"./scripts/run.py", "../lib/utils.py"},
		},
		{
			name:     "ftp URL in backtick skipped",
			body:     "Download from `ftp://files.example.com/data`.",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			refs := ExtractFileReferences(tt.body)
			assert.Equal(t, tt.expected, refs)
		})
	}
}

func TestLooksLikeURL(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"https://example.com/path", true},
		{"http://example.com/path", true},
		{"ftp://files.example.com/data", true},
		{"example.com/path", true},
		{"docs.example.io/guide", true},
		{"pkg.go.dev/encoding/json", true},
		{"scripts/run.py", false},
		{"./scripts/run.py", false},
		{"../lib/utils.py", false},
		{"assets/template.json", false},
		{"C://Windows//Users//", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, looksLikeURL(tt.input))
		})
	}
}

func TestValidateFolder(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setup      func(dir string)
		wantIssues int
	}{
		{
			name: "referenced file exists",
			body: "Run `scripts/run.py` to process data.",
			setup: func(dir string) {
				require.NoError(t, os.MkdirAll(filepath.Join(dir, "scripts"), 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "scripts", "run.py"), []byte("# python"), 0o644))
			},
			wantIssues: 0,
		},
		{
			name:       "referenced file missing",
			body:       "Run `scripts/run.py` to process data.",
			setup:      func(dir string) {},
			wantIssues: 1,
		},
		{
			name:       "no references",
			body:       "Simple instructions without file references.",
			setup:      func(dir string) {},
			wantIssues: 0,
		},
		{
			name:       "URLs in backticks do not warn",
			body:       "See `https://example.com/docs` and `example.com/api/v1` for details.",
			setup:      func(dir string) {},
			wantIssues: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.setup(dir)

			s := &Skill{Body: tt.body}
			issues := ValidateFolder(s, dir)
			assert.Len(t, issues, tt.wantIssues)
			for _, issue := range issues {
				assert.Equal(t, "folder", issue.Field)
				assert.Equal(t, SeverityWarning, issue.Severity)
			}
		})
	}
}

func TestMatchExclude(t *testing.T) {
	cases := []struct {
		name     string
		patterns []string
		rel      string
		want     bool
	}{
		{"dir name excludes subtree", []string{"eval"}, "eval/s1.md", true},
		{"dir name excludes nested subtree", []string{"eval"}, "eval/deep/s1.md", true},
		{"dir name excludes the dir itself", []string{"eval"}, "eval", true},
		{"trailing slash is equivalent", []string{"eval/"}, "eval/s1.md", true},
		{"leading ./ is equivalent", []string{"./eval"}, "eval/s1.md", true},
		{"no partial name match", []string{"eval"}, "evaluation/notes.md", false},
		{"children glob", []string{"fixtures/*"}, "fixtures/a.json", true},
		{"children glob prunes nested too", []string{"fixtures/*"}, "fixtures/sub/a.json", true},
		{"children glob does not match the dir", []string{"fixtures/*"}, "fixtures", false},
		{"suffix glob at top level", []string{"*.test.md"}, "a.test.md", true},
		{"suffix glob does not cross slashes", []string{"*.test.md"}, "sub/a.test.md", false},
		{"nested suffix glob", []string{"sub/*.test.md"}, "sub/a.test.md", true},
		{"SKILL.md never excluded", []string{"*"}, "SKILL.md", false},
		{"star excludes everything else", []string{"*"}, "notes.md", true},
		{"multiple patterns, any matches", []string{"a", "b"}, "b/x", true},
		{"empty patterns match nothing", nil, "eval/s1.md", false},
		{"blank pattern ignored", []string{"  "}, "eval/s1.md", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, MatchExclude(tc.patterns, tc.rel))
		})
	}
}

func TestValidateExcludePattern(t *testing.T) {
	// `*` / `*.md` are legal: SKILL.md is protected at match time, so they
	// mean "everything except SKILL.md". `\*.md` is path.Match's escape for
	// a literal star, not a separator.
	for _, ok := range []string{"eval", "eval/", "fixtures/*", "*.test.md", "a/b/c", "./scratch", "[abc]*", "*", "*.md", "SKILL.[m]d", `\*.md`, "./SKILL.md/x"} {
		assert.NoError(t, ValidateExcludePattern(ok), ok)
	}
	bad := map[string]string{
		"":            "must not be empty",
		"   ":         "must not be empty",
		"/abs":        "must be relative",
		"../up":       `must not contain ".."`,
		"a/../b":      `must not contain ".."`,
		"[unclosed":   "not a valid glob",
		"SKILL.md":    "is always installed",
		"./SKILL.md/": "is always installed",
		"**/x":        `"**" is not supported`,
		"a/**":        `"**" is not supported`,
		"a**b":        `"**" is not supported`,
	}
	for pattern, wantMsg := range bad {
		err := ValidateExcludePattern(pattern)
		require.Error(t, err, "pattern %q", pattern)
		assert.Contains(t, err.Error(), wantMsg, "pattern %q", pattern)
	}

	// ValidateExcludePatterns prefixes the index only for multi-entry lists.
	assert.NoError(t, ValidateExcludePatterns(nil))
	err := ValidateExcludePatterns([]string{"ok", "[x"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "install.exclude[1]")
	err = ValidateExcludePatterns([]string{"[x"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "install.exclude: ")
	assert.NotContains(t, err.Error(), "[0]")
}

func TestExcludedFiles(t *testing.T) {
	dir := t.TempDir()
	mk := func(rel string) {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
		require.NoError(t, os.WriteFile(p, []byte("x"), 0o644))
	}
	mk("SKILL.md")
	mk("eval/s1.md")
	mk("eval/s2.md")
	mk("references/a.md")

	got, err := ExcludedFiles(dir, []string{"eval"})
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{filepath.FromSlash("eval/s1.md"), filepath.FromSlash("eval/s2.md")}, got)

	got, err = ExcludedFiles(dir, []string{"nope"})
	require.NoError(t, err)
	assert.Empty(t, got)
}
