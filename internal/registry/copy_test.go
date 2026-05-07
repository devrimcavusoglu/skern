package registry

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCopySiblings_NestedFilesAndSkipsManifest(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	// Source layout:
	//   SKILL.md             (must NOT be copied — caller writes manifest separately)
	//   VENDORED.md
	//   references/architecture.md
	//   templates/example.txt
	require.NoError(t, os.WriteFile(filepath.Join(src, "SKILL.md"), []byte("manifest"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(src, "VENDORED.md"), []byte("vendored"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(src, "references"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(src, "references", "architecture.md"), []byte("arch"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(src, "templates"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(src, "templates", "example.txt"), []byte("example"), 0o644))

	require.NoError(t, CopySiblings(src, dst))

	// Manifest must not be copied.
	_, err := os.Stat(filepath.Join(dst, "SKILL.md"))
	assert.True(t, os.IsNotExist(err), "SKILL.md should not be copied by CopySiblings")

	// Everything else should be copied verbatim.
	for rel, want := range map[string]string{
		"VENDORED.md":                "vendored",
		"references/architecture.md": "arch",
		"templates/example.txt":      "example",
	} {
		got, err := os.ReadFile(filepath.Join(dst, rel))
		require.NoError(t, err, "reading %s", rel)
		assert.Equal(t, want, string(got), "content of %s", rel)
	}
}

func TestCopySiblings_EmptyDirectory(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	require.NoError(t, CopySiblings(src, dst))

	entries, err := os.ReadDir(dst)
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestCopySiblings_NestedSkillMdIsCopied(t *testing.T) {
	// Only the *top-level* SKILL.md is excluded. A nested SKILL.md (e.g.,
	// inside templates/) should be copied like any other file.
	src := t.TempDir()
	dst := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(src, "SKILL.md"), []byte("top"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(src, "templates"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(src, "templates", "SKILL.md"), []byte("nested"), 0o644))

	require.NoError(t, CopySiblings(src, dst))

	_, err := os.Stat(filepath.Join(dst, "SKILL.md"))
	assert.True(t, os.IsNotExist(err))

	got, err := os.ReadFile(filepath.Join(dst, "templates", "SKILL.md"))
	require.NoError(t, err)
	assert.Equal(t, "nested", string(got))
}
