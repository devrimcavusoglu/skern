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

// --- ClaudeCode adapter ---

func TestClaudeCode_Name(t *testing.T) {
	c := NewClaudeCode("/home/test", "/project")
	assert.Equal(t, TypeClaudeCode, c.Name())
}

func TestClaudeCode_Detect_Positive(t *testing.T) {
	home := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".claude"), 0o755))

	c := NewClaudeCode(home, t.TempDir())
	assert.True(t, c.Detect())
}

func TestClaudeCode_Detect_Negative(t *testing.T) {
	home := t.TempDir()
	c := NewClaudeCode(home, t.TempDir())
	assert.False(t, c.Detect())
}

func TestClaudeCode_Paths(t *testing.T) {
	c := NewClaudeCode("/home/test", "/project")
	assert.Equal(t, filepath.Join("/home/test", ".claude", "skills"), c.UserSkillsDir())
	assert.Equal(t, filepath.Join("/project", ".claude", "skills"), c.ProjectSkillsDir())
}

func TestClaudeCode_Install(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	registry := t.TempDir()

	skillDir := createSkillDir(t, registry, "my-skill")

	c := NewClaudeCode(home, project)
	require.NoError(t, c.Install(skillDir, "my-skill", skill.ScopeUser))

	// Verify installed
	installed := filepath.Join(home, ".claude", "skills", "my-skill", "SKILL.md")
	_, err := os.Stat(installed)
	require.NoError(t, err)
}

func TestClaudeCode_Install_Duplicate(t *testing.T) {
	home := t.TempDir()
	registry := t.TempDir()

	skillDir := createSkillDir(t, registry, "my-skill")

	c := NewClaudeCode(home, t.TempDir())
	require.NoError(t, c.Install(skillDir, "my-skill", skill.ScopeUser))

	err := c.Install(skillDir, "my-skill", skill.ScopeUser)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already installed")
}

func TestClaudeCode_Uninstall(t *testing.T) {
	home := t.TempDir()
	registry := t.TempDir()

	skillDir := createSkillDir(t, registry, "my-skill")

	c := NewClaudeCode(home, t.TempDir())
	require.NoError(t, c.Install(skillDir, "my-skill", skill.ScopeUser))
	require.NoError(t, c.Uninstall("my-skill", skill.ScopeUser))

	// Verify removed
	installed := filepath.Join(home, ".claude", "skills", "my-skill")
	_, err := os.Stat(installed)
	assert.True(t, os.IsNotExist(err))
}

func TestClaudeCode_Uninstall_NotFound(t *testing.T) {
	home := t.TempDir()
	c := NewClaudeCode(home, t.TempDir())

	err := c.Uninstall("nonexistent", skill.ScopeUser)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not installed")
}

func TestClaudeCode_InstalledSkills(t *testing.T) {
	home := t.TempDir()
	registry := t.TempDir()

	skillDir1 := createSkillDir(t, registry, "skill-a")
	skillDir2 := createSkillDir(t, registry, "skill-b")

	c := NewClaudeCode(home, t.TempDir())
	require.NoError(t, c.Install(skillDir1, "skill-a", skill.ScopeUser))
	require.NoError(t, c.Install(skillDir2, "skill-b", skill.ScopeUser))

	installed, err := c.InstalledSkills(skill.ScopeUser)
	require.NoError(t, err)
	assert.Len(t, installed, 2)
	assert.Contains(t, installed, "skill-a")
	assert.Contains(t, installed, "skill-b")
}

func TestClaudeCode_InstalledSkills_Empty(t *testing.T) {
	home := t.TempDir()
	c := NewClaudeCode(home, t.TempDir())

	installed, err := c.InstalledSkills(skill.ScopeUser)
	require.NoError(t, err)
	assert.Empty(t, installed)
}

// --- CodexCLI adapter ---

func TestCodexCLI_Name(t *testing.T) {
	c := NewCodexCLI("/home/test", "/project")
	assert.Equal(t, TypeCodexCLI, c.Name())
}

func TestCodexCLI_Detect_Agents(t *testing.T) {
	home := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".agents"), 0o755))

	c := NewCodexCLI(home, t.TempDir())
	assert.True(t, c.Detect())
}

func TestCodexCLI_Detect_CodexFallback(t *testing.T) {
	home := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".codex"), 0o755))

	c := NewCodexCLI(home, t.TempDir())
	assert.True(t, c.Detect())
}

func TestCodexCLI_Detect_Negative(t *testing.T) {
	home := t.TempDir()
	c := NewCodexCLI(home, t.TempDir())
	assert.False(t, c.Detect())
}

func TestCodexCLI_Paths(t *testing.T) {
	c := NewCodexCLI("/home/test", "/project")
	assert.Equal(t, filepath.Join("/home/test", ".agents", "skills"), c.UserSkillsDir())
	assert.Equal(t, filepath.Join("/project", ".agents", "skills"), c.ProjectSkillsDir())
}

func TestCodexCLI_Install(t *testing.T) {
	home := t.TempDir()
	registry := t.TempDir()

	skillDir := createSkillDir(t, registry, "my-skill")

	c := NewCodexCLI(home, t.TempDir())
	require.NoError(t, c.Install(skillDir, "my-skill", skill.ScopeUser))

	installed := filepath.Join(home, ".agents", "skills", "my-skill", "SKILL.md")
	_, err := os.Stat(installed)
	require.NoError(t, err)
}

func TestCodexCLI_Uninstall(t *testing.T) {
	home := t.TempDir()
	registry := t.TempDir()

	skillDir := createSkillDir(t, registry, "my-skill")

	c := NewCodexCLI(home, t.TempDir())
	require.NoError(t, c.Install(skillDir, "my-skill", skill.ScopeUser))
	require.NoError(t, c.Uninstall("my-skill", skill.ScopeUser))

	_, err := os.Stat(filepath.Join(home, ".agents", "skills", "my-skill"))
	assert.True(t, os.IsNotExist(err))
}

// --- OpenCode adapter ---

func TestOpenCode_Name(t *testing.T) {
	o := NewOpenCode("/home/test", "/project")
	assert.Equal(t, TypeOpenCode, o.Name())
}

func TestOpenCode_Detect_Positive(t *testing.T) {
	home := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".config", "opencode"), 0o755))

	o := NewOpenCode(home, t.TempDir())
	assert.True(t, o.Detect())
}

func TestOpenCode_Detect_Negative(t *testing.T) {
	home := t.TempDir()
	o := NewOpenCode(home, t.TempDir())
	assert.False(t, o.Detect())
}

func TestOpenCode_Paths(t *testing.T) {
	o := NewOpenCode("/home/test", "/project")
	assert.Equal(t, filepath.Join("/home/test", ".config", "opencode", "skills"), o.UserSkillsDir())
	assert.Equal(t, filepath.Join("/project", ".opencode", "skills"), o.ProjectSkillsDir())
}

func TestOpenCode_Install(t *testing.T) {
	home := t.TempDir()
	registry := t.TempDir()

	skillDir := createSkillDir(t, registry, "my-skill")

	o := NewOpenCode(home, t.TempDir())
	require.NoError(t, o.Install(skillDir, "my-skill", skill.ScopeUser))

	installed := filepath.Join(home, ".config", "opencode", "skills", "my-skill", "SKILL.md")
	_, err := os.Stat(installed)
	require.NoError(t, err)
}

func TestOpenCode_Uninstall(t *testing.T) {
	home := t.TempDir()
	registry := t.TempDir()

	skillDir := createSkillDir(t, registry, "my-skill")

	o := NewOpenCode(home, t.TempDir())
	require.NoError(t, o.Install(skillDir, "my-skill", skill.ScopeUser))
	require.NoError(t, o.Uninstall("my-skill", skill.ScopeUser))

	_, err := os.Stat(filepath.Join(home, ".config", "opencode", "skills", "my-skill"))
	assert.True(t, os.IsNotExist(err))
}

// --- Project scope ---

func TestClaudeCode_ProjectScope(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	registry := t.TempDir()

	skillDir := createSkillDir(t, registry, "proj-skill")

	c := NewClaudeCode(home, project)
	require.NoError(t, c.Install(skillDir, "proj-skill", skill.ScopeProject))

	installed := filepath.Join(project, ".claude", "skills", "proj-skill", "SKILL.md")
	_, err := os.Stat(installed)
	require.NoError(t, err)

	// User scope should be empty
	userInstalled, err := c.InstalledSkills(skill.ScopeUser)
	require.NoError(t, err)
	assert.Empty(t, userInstalled)

	// Project scope should have the skill
	projectInstalled, err := c.InstalledSkills(skill.ScopeProject)
	require.NoError(t, err)
	assert.Len(t, projectInstalled, 1)
	assert.Equal(t, "proj-skill", projectInstalled[0])
}

// --- Detector ---

func TestDetector_DetectAll(t *testing.T) {
	home := t.TempDir()

	// Only create .claude directory
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".claude"), 0o755))

	det := NewDetectorWithPlatforms([]Platform{
		NewClaudeCode(home, t.TempDir()),
		NewCodexCLI(home, t.TempDir()),
		NewOpenCode(home, t.TempDir()),
	})

	detected := det.DetectAll()
	assert.Len(t, detected, 1)
	assert.Equal(t, TypeClaudeCode, detected[0].Name())
}

func TestDetector_Get(t *testing.T) {
	det := NewDetectorWithPlatforms([]Platform{
		NewClaudeCode(t.TempDir(), t.TempDir()),
		NewCodexCLI(t.TempDir(), t.TempDir()),
	})

	p := det.Get(TypeCodexCLI)
	require.NotNil(t, p)
	assert.Equal(t, TypeCodexCLI, p.Name())
}

func TestDetector_Get_NotFound(t *testing.T) {
	det := NewDetectorWithPlatforms([]Platform{
		NewClaudeCode(t.TempDir(), t.TempDir()),
	})

	p := det.Get(TypeOpenCode)
	assert.Nil(t, p)
}

func TestDetector_All(t *testing.T) {
	det := NewDetectorWithPlatforms([]Platform{
		NewClaudeCode(t.TempDir(), t.TempDir()),
		NewCodexCLI(t.TempDir(), t.TempDir()),
		NewOpenCode(t.TempDir(), t.TempDir()),
	})

	all := det.All()
	assert.Len(t, all, 3)
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
		{"all", TypeAll, false},
		{"Claude-Code", TypeClaudeCode, false},
		{"ALL", TypeAll, false},
		{"", "", true},
		{"unknown", "", true},
		{"github-copilot", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParsePlatformType(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
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
	claude := NewClaudeCode(home, project)
	codex := NewCodexCLI(home, project)

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
