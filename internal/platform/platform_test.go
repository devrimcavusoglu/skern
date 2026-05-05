package platform

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/devrimcavusoglu/skern/internal/skill"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createSkillDir creates a minimal skill directory with a SKILL.md file for testing.
func createSkillDir(t *testing.T, baseDir, name string) string {
	t.Helper()
	dir := filepath.Join(baseDir, name)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	content := "---\nname: " + name + "\ndescription: test skill\n---\n\nInstructions here.\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644))
	return dir
}

// --- Spec registry ---

func TestSpecsAreUnique(t *testing.T) {
	seen := map[Type]bool{}
	for _, s := range Specs {
		assert.False(t, seen[s.Name], "duplicate spec for %q", s.Name)
		seen[s.Name] = true
		assert.NotEmpty(t, s.UserDir, "%q must have UserDir", s.Name)
		assert.NotEmpty(t, s.ProjectDir, "%q must have ProjectDir", s.Name)
		assert.NotEmpty(t, s.DetectHome, "%q must have at least one DetectHome path", s.Name)
	}
}

func TestSpecsCoverExpectedPlatforms(t *testing.T) {
	// Acceptance criterion (#80): at least 5 new platforms beyond the
	// original three. Lock the minimum here so future edits can't silently
	// drop one.
	expected := []Type{
		TypeClaudeCode, TypeCodexCLI, TypeOpenCode,
		TypeCursor, TypeGeminiCLI, TypeGitHubCopilot, TypeWindsurf, TypeContinue,
	}
	for _, e := range expected {
		assert.NotNil(t, SpecFor(e), "expected spec for %q to be registered", e)
	}
	assert.GreaterOrEqual(t, len(Specs), 8)
}

func TestSpecFor_Unknown(t *testing.T) {
	assert.Nil(t, SpecFor(Type("does-not-exist")))
}

func TestSupportedNames(t *testing.T) {
	names := SupportedNames()
	assert.Equal(t, len(Specs), len(names))
	for i, s := range Specs {
		assert.Equal(t, s.Name, names[i], "SupportedNames must preserve declaration order")
	}
}

// --- Generic adapter (one row per spec) ---

func TestAdapter_NamesAndPaths(t *testing.T) {
	for _, s := range Specs {
		t.Run(string(s.Name), func(t *testing.T) {
			a := New(s.Name, "/home/test", "/project")
			require.NotNil(t, a)
			assert.Equal(t, s.Name, a.Name())
			assert.Equal(t, filepath.Join("/home/test", s.UserDir), a.UserSkillsDir())
			assert.Equal(t, filepath.Join("/project", s.ProjectDir), a.ProjectSkillsDir())
		})
	}
}

func TestAdapter_New_UnknownReturnsNil(t *testing.T) {
	assert.Nil(t, New(Type("nope"), "/h", "/p"))
}

func TestAdapter_New_DefaultsResolveHomeAndCwd(t *testing.T) {
	a := New(TypeClaudeCode, "", "")
	require.NotNil(t, a)
	// Must produce a non-empty path; exact value depends on the test runner's
	// home directory.
	assert.NotEmpty(t, a.UserSkillsDir())
	assert.NotEmpty(t, a.ProjectSkillsDir())
}

func TestAdapter_Detect(t *testing.T) {
	for _, s := range Specs {
		t.Run(string(s.Name)+"_positive", func(t *testing.T) {
			home := t.TempDir()
			require.NoError(t, os.MkdirAll(filepath.Join(home, s.DetectHome[0]), 0o755))
			a := New(s.Name, home, t.TempDir())
			assert.True(t, a.Detect())
		})
		t.Run(string(s.Name)+"_negative", func(t *testing.T) {
			home := t.TempDir()
			a := New(s.Name, home, t.TempDir())
			assert.False(t, a.Detect())
		})
	}
}

// TestAdapter_Detect_FallbackPaths ensures every entry in DetectHome
// independently triggers detection — important for codex (~/.codex or
// ~/.agents) and windsurf (~/.codeium/windsurf or ~/.windsurf).
func TestAdapter_Detect_FallbackPaths(t *testing.T) {
	for _, s := range Specs {
		if len(s.DetectHome) <= 1 {
			continue
		}
		for i, p := range s.DetectHome {
			t.Run(string(s.Name)+"_path"+string(rune('0'+i)), func(t *testing.T) {
				home := t.TempDir()
				require.NoError(t, os.MkdirAll(filepath.Join(home, p), 0o755))
				a := New(s.Name, home, t.TempDir())
				assert.True(t, a.Detect(), "path %q should trigger detection", p)
			})
		}
	}
}

// TestAdapter_RoundTrip verifies install/list/uninstall lifecycle for every
// platform via the spec table — replaces the per-platform install/uninstall
// tests that used to live in claude.go/codex.go/opencode.go.
func TestAdapter_RoundTrip(t *testing.T) {
	for _, s := range Specs {
		t.Run(string(s.Name), func(t *testing.T) {
			home := t.TempDir()
			project := t.TempDir()
			registry := t.TempDir()

			skillDir := createSkillDir(t, registry, "round-trip")

			a := New(s.Name, home, project)

			// User scope
			require.NoError(t, a.Install(skillDir, "round-trip", skill.ScopeUser))
			userInstalled := filepath.Join(home, s.UserDir, "round-trip", "SKILL.md")
			_, err := os.Stat(userInstalled)
			require.NoError(t, err, "expected SKILL.md at %s", userInstalled)

			names, err := a.InstalledSkills(skill.ScopeUser)
			require.NoError(t, err)
			assert.Contains(t, names, "round-trip")

			require.NoError(t, a.Uninstall("round-trip", skill.ScopeUser))
			_, err = os.Stat(filepath.Join(home, s.UserDir, "round-trip"))
			assert.True(t, os.IsNotExist(err))

			// Project scope
			require.NoError(t, a.Install(skillDir, "round-trip", skill.ScopeProject))
			projectInstalled := filepath.Join(project, s.ProjectDir, "round-trip", "SKILL.md")
			_, err = os.Stat(projectInstalled)
			require.NoError(t, err, "expected SKILL.md at %s", projectInstalled)

			require.NoError(t, a.Uninstall("round-trip", skill.ScopeProject))
		})
	}
}

func TestAdapter_Install_Duplicate(t *testing.T) {
	home := t.TempDir()
	registry := t.TempDir()
	skillDir := createSkillDir(t, registry, "dup")

	a := New(TypeClaudeCode, home, t.TempDir())
	require.NoError(t, a.Install(skillDir, "dup", skill.ScopeUser))

	err := a.Install(skillDir, "dup", skill.ScopeUser)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already installed")
}

func TestAdapter_Uninstall_NotFound(t *testing.T) {
	a := New(TypeClaudeCode, t.TempDir(), t.TempDir())
	err := a.Uninstall("ghost", skill.ScopeUser)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not installed")
}

func TestAdapter_InstalledSkills_Empty(t *testing.T) {
	a := New(TypeClaudeCode, t.TempDir(), t.TempDir())
	names, err := a.InstalledSkills(skill.ScopeUser)
	require.NoError(t, err)
	assert.Empty(t, names)
}

// TestAdapter_SharedProjectDir captures the design decision from #80: when
// several platforms share .agents/skills/ as their project dir, a skill
// installed via one of them is visible to the others. Capacity reporting is
// per-directory by design.
func TestAdapter_SharedProjectDir(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	registry := t.TempDir()
	skillDir := createSkillDir(t, registry, "shared")

	cursor := New(TypeCursor, home, project)
	gemini := New(TypeGeminiCLI, home, project)
	require.Equal(t, cursor.ProjectSkillsDir(), gemini.ProjectSkillsDir(),
		"sanity: cursor and gemini-cli should share .agents/skills/")

	require.NoError(t, cursor.Install(skillDir, "shared", skill.ScopeProject))

	// Both adapters report the skill because they read the same directory.
	cursorList, err := cursor.InstalledSkills(skill.ScopeProject)
	require.NoError(t, err)
	assert.Contains(t, cursorList, "shared")

	geminiList, err := gemini.InstalledSkills(skill.ScopeProject)
	require.NoError(t, err)
	assert.Contains(t, geminiList, "shared")
}

// --- Detector ---

func TestDetector_DetectAll(t *testing.T) {
	home := t.TempDir()

	// Only create .claude directory
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".claude"), 0o755))

	det := NewDetectorWithPlatforms([]Platform{
		New(TypeClaudeCode, home, t.TempDir()),
		New(TypeCodexCLI, home, t.TempDir()),
		New(TypeOpenCode, home, t.TempDir()),
	})

	detected := det.DetectAll()
	require.Len(t, detected, 1)
	assert.Equal(t, TypeClaudeCode, detected[0].Name())
}

func TestDetector_Get(t *testing.T) {
	det := NewDetectorWithPlatforms([]Platform{
		New(TypeClaudeCode, t.TempDir(), t.TempDir()),
		New(TypeCodexCLI, t.TempDir(), t.TempDir()),
	})

	p := det.Get(TypeCodexCLI)
	require.NotNil(t, p)
	assert.Equal(t, TypeCodexCLI, p.Name())
}

func TestDetector_Get_NotFound(t *testing.T) {
	det := NewDetectorWithPlatforms([]Platform{
		New(TypeClaudeCode, t.TempDir(), t.TempDir()),
	})

	p := det.Get(TypeOpenCode)
	assert.Nil(t, p)
}

func TestDetector_All_DefaultIncludesEverySpec(t *testing.T) {
	det, err := NewDetector()
	require.NoError(t, err)
	assert.Len(t, det.All(), len(Specs))
}

// --- ParsePlatformType ---

func TestParsePlatformType(t *testing.T) {
	tests := []struct {
		input   string
		want    Type
		wantErr bool
	}{
		{"claude-code", TypeClaudeCode, false},
		{"codex-cli", TypeCodexCLI, false},
		{"opencode", TypeOpenCode, false},
		{"Claude-Code", TypeClaudeCode, false},
		// New platforms — must round-trip through the parser.
		{"cursor", TypeCursor, false},
		{"gemini-cli", TypeGeminiCLI, false},
		{"github-copilot", TypeGitHubCopilot, false},
		{"windsurf", TypeWindsurf, false},
		{"continue", TypeContinue, false},
		// "all" is rejected per #52 D6 — agents must specify their own platform.
		{"all", "", true},
		{"ALL", "", true},
		{"", "", true},
		{"unknown", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParsePlatformType(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestParsePlatformType_ErrorListsRegisteredPlatforms(t *testing.T) {
	_, err := ParsePlatformType("nope")
	require.Error(t, err)
	// Error message must enumerate every supported platform so users have a
	// recovery path without needing to consult the docs.
	for _, n := range SupportedNames() {
		assert.Contains(t, err.Error(), string(n))
	}
}

// --- Integration: full lifecycle ---

func TestFullLifecycle(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	registry := t.TempDir()

	// Simulate creating a skill in registry
	skillDir := createSkillDir(t, registry, "lifecycle-skill")

	// Create adapters
	claude := New(TypeClaudeCode, home, project)
	codex := New(TypeCodexCLI, home, project)

	// Install to Claude Code
	require.NoError(t, claude.Install(skillDir, "lifecycle-skill", skill.ScopeUser))

	// Install to Codex CLI
	require.NoError(t, codex.Install(skillDir, "lifecycle-skill", skill.ScopeUser))

	// List installed on both
	claudeSkills, err := claude.InstalledSkills(skill.ScopeUser)
	require.NoError(t, err)
	assert.Contains(t, claudeSkills, "lifecycle-skill")

	codexSkills, err := codex.InstalledSkills(skill.ScopeUser)
	require.NoError(t, err)
	assert.Contains(t, codexSkills, "lifecycle-skill")

	// Uninstall from Claude Code
	require.NoError(t, claude.Uninstall("lifecycle-skill", skill.ScopeUser))

	claudeSkills, err = claude.InstalledSkills(skill.ScopeUser)
	require.NoError(t, err)
	assert.NotContains(t, claudeSkills, "lifecycle-skill")

	// Codex should still have it
	codexSkills, err = codex.InstalledSkills(skill.ScopeUser)
	require.NoError(t, err)
	assert.Contains(t, codexSkills, "lifecycle-skill")

	// Uninstall from Codex
	require.NoError(t, codex.Uninstall("lifecycle-skill", skill.ScopeUser))

	codexSkills, err = codex.InstalledSkills(skill.ScopeUser)
	require.NoError(t, err)
	assert.NotContains(t, codexSkills, "lifecycle-skill")
}

// --- Helpers ---

func TestCopyDir(t *testing.T) {
	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "copy")

	// Create a nested structure
	require.NoError(t, os.MkdirAll(filepath.Join(src, "sub"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(src, "file1.txt"), []byte("hello"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(src, "sub", "file2.txt"), []byte("world"), 0o644))

	require.NoError(t, copyDir(src, dst))

	// Verify structure
	data1, err := os.ReadFile(filepath.Join(dst, "file1.txt"))
	require.NoError(t, err)
	assert.Equal(t, "hello", string(data1))

	data2, err := os.ReadFile(filepath.Join(dst, "sub", "file2.txt"))
	require.NoError(t, err)
	assert.Equal(t, "world", string(data2))
}

func TestListInstalledSkills_SkipsNonSkillDirs(t *testing.T) {
	base := t.TempDir()

	// Valid skill dir
	skillDir := filepath.Join(base, "valid-skill")
	require.NoError(t, os.MkdirAll(skillDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("test"), 0o644))

	// Invalid dir (no SKILL.md)
	require.NoError(t, os.MkdirAll(filepath.Join(base, "not-a-skill"), 0o755))

	// Regular file (should be skipped)
	require.NoError(t, os.WriteFile(filepath.Join(base, "random.txt"), []byte("test"), 0o644))

	names, err := listInstalledSkills(base)
	require.NoError(t, err)
	assert.Equal(t, []string{"valid-skill"}, names)
}
