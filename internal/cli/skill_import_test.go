package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/devrimcavusoglu/skern/internal/output"
	"github.com/devrimcavusoglu/skern/internal/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validSkillManifest returns a valid SKILL.md content for testing.
func validSkillManifest() string {
	return `---
name: test-skill
description: A test skill for import testing
metadata:
  author:
    name: alice
    type: human
  version: "1.0.0"
---

## Instructions

Do the thing.
`
}

// newImportTestServer creates a mock GitHub API server that serves a skill directory.
func newImportTestServer(t *testing.T, manifest string, companions map[string]string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	// GitHub repo contents endpoint
	mux.HandleFunc("/repos/owner/repo/contents/skills/test-skill", func(w http.ResponseWriter, r *http.Request) {
		entries := []map[string]interface{}{
			{"name": "SKILL.md", "type": "file", "download_url": fmt.Sprintf("http://%s/raw/SKILL.md", r.Host)},
		}
		for name := range companions {
			entries = append(entries, map[string]interface{}{
				"name":         name,
				"type":         "file",
				"download_url": fmt.Sprintf("http://%s/raw/%s", r.Host, name),
			})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(entries)
	})

	mux.HandleFunc("/raw/SKILL.md", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(manifest))
	})
	for name, content := range companions {
		name, content := name, content
		mux.HandleFunc("/raw/"+name, func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(content))
		})
	}

	// Gist endpoint
	mux.HandleFunc("/gists/test-gist", func(w http.ResponseWriter, _ *http.Request) {
		gistFiles := map[string]interface{}{
			"SKILL.md": map[string]interface{}{
				"filename": "SKILL.md",
				"content":  manifest,
			},
		}
		for name, content := range companions {
			gistFiles[name] = map[string]interface{}{
				"filename": name,
				"content":  content,
			}
		}
		gist := map[string]interface{}{"files": gistFiles}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(gist)
	})

	return httptest.NewServer(mux)
}

func testImportContext(t *testing.T, srv *httptest.Server) *CommandContext {
	t.Helper()
	userDir := filepath.Join(t.TempDir(), "user-skills")
	projectDir := filepath.Join(t.TempDir(), "project-skills")

	return &CommandContext{
		NewRegistry: func() (*registry.Registry, error) {
			return registry.New(userDir, projectDir), nil
		},
		NewDetector:   defaultNewDetector,
		HTTPClient:    srv.Client(),
		GitHubBaseURL: srv.URL,
	}
}

func TestSkillImport_GitHubRepo(t *testing.T) {
	srv := newImportTestServer(t, validSkillManifest(), nil)
	defer srv.Close()

	cc := testImportContext(t, srv)
	out, err := runCmd(t, cc, "skill", "import",
		"https://github.com/owner/repo/tree/main/skills/test-skill")
	require.NoError(t, err)
	assert.Contains(t, out, "Imported skill")
	assert.Contains(t, out, "test-skill")
}

func TestSkillImport_GitHubRepo_JSON(t *testing.T) {
	srv := newImportTestServer(t, validSkillManifest(), nil)
	defer srv.Close()

	cc := testImportContext(t, srv)
	out, err := runCmd(t, cc, "skill", "import",
		"https://github.com/owner/repo/tree/main/skills/test-skill", "--json")
	require.NoError(t, err)

	var result output.SkillImportResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "test-skill", result.Name)
	assert.Equal(t, "user", result.Scope)
	assert.NotEmpty(t, result.Path)
	assert.Contains(t, result.Source, "github.com")
}

func TestSkillImport_GitHubGist(t *testing.T) {
	srv := newImportTestServer(t, validSkillManifest(), nil)
	defer srv.Close()

	cc := testImportContext(t, srv)
	out, err := runCmd(t, cc, "skill", "import",
		"https://gist.github.com/alice/test-gist")
	require.NoError(t, err)
	assert.Contains(t, out, "Imported skill")
	assert.Contains(t, out, "test-skill")
}

func TestSkillImport_WithCompanionFiles(t *testing.T) {
	companions := map[string]string{
		"helper.md": "# Helper\nSome content",
	}
	srv := newImportTestServer(t, validSkillManifest(), companions)
	defer srv.Close()

	cc := testImportContext(t, srv)
	out, err := runCmd(t, cc, "skill", "import",
		"https://github.com/owner/repo/tree/main/skills/test-skill", "--json")
	require.NoError(t, err)

	var result output.SkillImportResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))

	// Verify companion file was written
	assert.FileExists(t, filepath.Join(result.Path, "helper.md"))
}

func TestSkillImport_NameOverride(t *testing.T) {
	srv := newImportTestServer(t, validSkillManifest(), nil)
	defer srv.Close()

	cc := testImportContext(t, srv)
	out, err := runCmd(t, cc, "skill", "import",
		"https://github.com/owner/repo/tree/main/skills/test-skill",
		"--name", "renamed-skill", "--json")
	require.NoError(t, err)

	var result output.SkillImportResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "renamed-skill", result.Name)
}

func TestSkillImport_InvalidNameOverride(t *testing.T) {
	srv := newImportTestServer(t, validSkillManifest(), nil)
	defer srv.Close()

	cc := testImportContext(t, srv)
	_, err := runCmd(t, cc, "skill", "import",
		"https://github.com/owner/repo/tree/main/skills/test-skill",
		"--name", "INVALID_NAME")
	assert.Error(t, err)
}

func TestSkillImport_ProjectScope(t *testing.T) {
	srv := newImportTestServer(t, validSkillManifest(), nil)
	defer srv.Close()

	cc := testImportContext(t, srv)
	out, err := runCmd(t, cc, "skill", "import",
		"https://github.com/owner/repo/tree/main/skills/test-skill",
		"--scope", "project", "--json")
	require.NoError(t, err)

	var result output.SkillImportResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "project", result.Scope)
}

func TestSkillImport_AlreadyExists(t *testing.T) {
	srv := newImportTestServer(t, validSkillManifest(), nil)
	defer srv.Close()

	cc := testImportContext(t, srv)

	// First import
	_, err := runCmd(t, cc, "skill", "import",
		"https://github.com/owner/repo/tree/main/skills/test-skill")
	require.NoError(t, err)

	// Second import without --force should fail
	_, err = runCmd(t, cc, "skill", "import",
		"https://github.com/owner/repo/tree/main/skills/test-skill")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestSkillImport_Force(t *testing.T) {
	srv := newImportTestServer(t, validSkillManifest(), nil)
	defer srv.Close()

	cc := testImportContext(t, srv)

	// First import
	_, err := runCmd(t, cc, "skill", "import",
		"https://github.com/owner/repo/tree/main/skills/test-skill")
	require.NoError(t, err)

	// Second import with --force should succeed
	_, err = runCmd(t, cc, "skill", "import",
		"https://github.com/owner/repo/tree/main/skills/test-skill", "--force")
	require.NoError(t, err)
}

func TestSkillImport_InvalidURL(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "import", "not-a-url")
	assert.Error(t, err)
}

func TestSkillImport_UnsupportedHost(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "import", "https://gitlab.com/owner/repo/tree/main/skill")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported host")
}

func TestSkillImport_NoArgs(t *testing.T) {
	cc := testRegistry(t)

	_, err := runCmd(t, cc, "skill", "import")
	assert.Error(t, err)
}
