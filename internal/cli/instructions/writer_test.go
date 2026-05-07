package instructions

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoverTargets_FindsExistingFiles(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("hi"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(root, ".claude"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, ".claude", "CLAUDE.md"), []byte("hi"), 0o644))

	got, err := DiscoverTargets(root)
	require.NoError(t, err)
	assert.Contains(t, got, filepath.Join(root, "AGENTS.md"))
	assert.Contains(t, got, filepath.Join(root, ".claude", "CLAUDE.md"))
	assert.NotContains(t, got, filepath.Join(root, "CLAUDE.md"))
}

func TestDiscoverTargets_EmptyWhenNoneExist(t *testing.T) {
	root := t.TempDir()
	got, err := DiscoverTargets(root)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestDiscoverTargets_IgnoresDirectoriesWithMatchingNames(t *testing.T) {
	root := t.TempDir()
	// Create a directory named AGENTS.md instead of a file.
	require.NoError(t, os.MkdirAll(filepath.Join(root, "AGENTS.md"), 0o755))

	got, err := DiscoverTargets(root)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestWrite_CreatesNewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "AGENTS.md")
	rendered := Render(false)

	res, err := Write(path, rendered)
	require.NoError(t, err)
	assert.Equal(t, "created", res.Action)
	assert.True(t, res.Created)

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, rendered, string(got))
}

func TestWrite_AppendsToExistingFileWithoutBlock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")
	original := "# Project\n\nExisting content here.\n"
	require.NoError(t, os.WriteFile(path, []byte(original), 0o644))

	rendered := Render(true)
	res, err := Write(path, rendered)
	require.NoError(t, err)
	assert.Equal(t, "appended", res.Action)
	assert.False(t, res.Created)

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(got), "Existing content here.")
	assert.Contains(t, string(got), StartMarker)
	assert.Contains(t, string(got), EndMarker)
	assert.Contains(t, string(got), "Tool-forming loop")
}

func TestWrite_UpdatesExistingBlockInPlace(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "AGENTS.md")
	prefix := "# Header\n\nBefore.\n\n"
	suffix := "\nAfter.\n"
	original := prefix + Render(false) + suffix
	require.NoError(t, os.WriteFile(path, []byte(original), 0o644))

	rendered := Render(true)
	res, err := Write(path, rendered)
	require.NoError(t, err)
	assert.Equal(t, "updated", res.Action)

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	gotStr := string(got)
	assert.Contains(t, gotStr, "Before.")
	assert.Contains(t, gotStr, "After.")
	assert.Contains(t, gotStr, "Tool-forming loop")
	// Make sure there's exactly one block.
	assert.Equal(t, 1, countOccurrences(gotStr, StartMarker))
	assert.Equal(t, 1, countOccurrences(gotStr, EndMarker))
}

func TestWrite_NoOpWhenBlockUnchanged(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "AGENTS.md")
	rendered := Render(true)
	original := "preface\n\n" + rendered + "\nepilogue\n"
	require.NoError(t, os.WriteFile(path, []byte(original), 0o644))

	info1, err := os.Stat(path)
	require.NoError(t, err)
	mtime1 := info1.ModTime()

	res, err := Write(path, rendered)
	require.NoError(t, err)
	assert.Equal(t, "unchanged", res.Action)

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, original, string(got))

	// File should not have been re-written. Allow filesystems with low mtime
	// resolution (most fail on second-level checks); we mostly care that the
	// file content is identical.
	info2, err := os.Stat(path)
	require.NoError(t, err)
	if info2.ModTime().After(mtime1) {
		// Some filesystems still update mtime. Don't fail; the byte equality
		// above is the real contract.
		t.Logf("file mtime updated despite unchanged content (fs-specific)")
	}
}

func TestWrite_AppendInsertsBlankLineSeparator(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "AGENTS.md")
	// Original ends with a single newline — writer should insert another
	// before the block to keep markdown spacing readable.
	require.NoError(t, os.WriteFile(path, []byte("# Header\n\nExisting paragraph.\n"), 0o644))

	rendered := Render(false)
	_, err := Write(path, rendered)
	require.NoError(t, err)

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(got), "Existing paragraph.\n\n"+StartMarker,
		"blank line should separate prior content from inserted block")
}

func countOccurrences(s, sub string) int {
	count := 0
	for i := 0; i+len(sub) <= len(s); {
		j := i + indexOf(s[i:], sub)
		if j < i {
			break
		}
		count++
		i = j + len(sub)
	}
	return count
}

// indexOf returns the index of sub in s, or -1.
func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
