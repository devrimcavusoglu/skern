package skill

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseImportURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    *ImportSource
		wantErr string
	}{
		{
			name: "github repo with path",
			url:  "https://github.com/owner/repo/tree/main/skills/my-skill",
			want: &ImportSource{
				Type:   SourceGitHubRepo,
				RawURL: "https://github.com/owner/repo/tree/main/skills/my-skill",
				Owner:  "owner",
				Repo:   "repo",
				Ref:    "main",
				Path:   "skills/my-skill",
			},
		},
		{
			name: "github repo with branch ref",
			url:  "https://github.com/org/project/tree/feature/v2/path/to/skill",
			want: &ImportSource{
				Type:   SourceGitHubRepo,
				RawURL: "https://github.com/org/project/tree/feature/v2/path/to/skill",
				Owner:  "org",
				Repo:   "project",
				Ref:    "feature",
				Path:   "v2/path/to/skill",
			},
		},
		{
			name: "github repo root tree",
			url:  "https://github.com/owner/repo/tree/main",
			want: &ImportSource{
				Type:   SourceGitHubRepo,
				RawURL: "https://github.com/owner/repo/tree/main",
				Owner:  "owner",
				Repo:   "repo",
				Ref:    "main",
				Path:   "",
			},
		},
		{
			name: "github gist with user",
			url:  "https://gist.github.com/alice/abc123def456",
			want: &ImportSource{
				Type:   SourceGitHubGist,
				RawURL: "https://gist.github.com/alice/abc123def456",
				GistID: "abc123def456",
			},
		},
		{
			name: "github gist without user",
			url:  "https://gist.github.com/abc123def456",
			want: &ImportSource{
				Type:   SourceGitHubGist,
				RawURL: "https://gist.github.com/abc123def456",
				GistID: "abc123def456",
			},
		},
		{
			name:    "unsupported host",
			url:     "https://gitlab.com/owner/repo/tree/main/skill",
			wantErr: "unsupported host",
		},
		{
			name:    "github repo without tree",
			url:     "https://github.com/owner/repo",
			wantErr: "expected GitHub directory URL format",
		},
		{
			name:    "github repo blob instead of tree",
			url:     "https://github.com/owner/repo/blob/main/SKILL.md",
			wantErr: "expected GitHub directory URL format",
		},
		{
			name:    "no scheme",
			url:     "github.com/owner/repo/tree/main/skill",
			wantErr: "unsupported URL scheme",
		},
		{
			name:    "ftp scheme",
			url:     "ftp://github.com/owner/repo/tree/main/skill",
			wantErr: "unsupported URL scheme",
		},
		{
			name:    "empty gist path",
			url:     "https://gist.github.com/",
			wantErr: "expected GitHub gist URL format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseImportURL(tt.url)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want.Type, got.Type)
			assert.Equal(t, tt.want.RawURL, got.RawURL)
			assert.Equal(t, tt.want.Owner, got.Owner)
			assert.Equal(t, tt.want.Repo, got.Repo)
			assert.Equal(t, tt.want.Ref, got.Ref)
			assert.Equal(t, tt.want.Path, got.Path)
			assert.Equal(t, tt.want.GistID, got.GistID)
		})
	}
}

// validManifestContent returns a valid SKILL.md content for testing.
func validManifestContent() string {
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

func TestFetchSkill_GitHubRepo(t *testing.T) {
	manifest := validManifestContent()
	companion := "# Helper\nSome helper content"

	// Mock GitHub API: contents endpoint returns a directory listing
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/contents/skills/my-skill", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "main", r.URL.Query().Get("ref"))
		entries := []map[string]interface{}{
			{"name": "SKILL.md", "type": "file", "download_url": fmt.Sprintf("http://%s/raw/SKILL.md", r.Host)},
			{"name": "helper.md", "type": "file", "download_url": fmt.Sprintf("http://%s/raw/helper.md", r.Host)},
			{"name": "subdir", "type": "dir", "download_url": ""},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entries)
	})
	mux.HandleFunc("/raw/SKILL.md", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(manifest))
	})
	mux.HandleFunc("/raw/helper.md", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(companion))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	src := &ImportSource{
		Type:   SourceGitHubRepo,
		RawURL: "https://github.com/owner/repo/tree/main/skills/my-skill",
		Owner:  "owner",
		Repo:   "repo",
		Ref:    "main",
		Path:   "skills/my-skill",
	}

	fetched, err := FetchSkillWithBaseURL(srv.Client(), src, srv.URL)
	require.NoError(t, err)
	assert.Len(t, fetched.Files, 2) // SKILL.md + helper.md (subdir skipped)
	assert.Equal(t, []byte(manifest), fetched.Files["SKILL.md"])
	assert.Equal(t, []byte(companion), fetched.Files["helper.md"])
}

func TestFetchSkill_GitHubRepo_NoManifest(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/contents/skill-dir", func(w http.ResponseWriter, r *http.Request) {
		entries := []map[string]interface{}{
			{"name": "README.md", "type": "file", "download_url": fmt.Sprintf("http://%s/raw/README.md", r.Host)},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entries)
	})
	mux.HandleFunc("/raw/README.md", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("just a readme"))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	src := &ImportSource{
		Type:  SourceGitHubRepo,
		Owner: "owner",
		Repo:  "repo",
		Ref:   "main",
		Path:  "skill-dir",
	}

	_, err := FetchSkillWithBaseURL(srv.Client(), src, srv.URL)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no SKILL.md found")
}

func TestFetchSkill_GitHubRepo_APIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/contents/bad-path", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	src := &ImportSource{
		Type:  SourceGitHubRepo,
		Owner: "owner",
		Repo:  "repo",
		Ref:   "main",
		Path:  "bad-path",
	}

	_, err := FetchSkillWithBaseURL(srv.Client(), src, srv.URL)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestFetchSkill_GitHubGist(t *testing.T) {
	manifest := validManifestContent()

	mux := http.NewServeMux()
	mux.HandleFunc("/gists/abc123", func(w http.ResponseWriter, _ *http.Request) {
		gist := map[string]interface{}{
			"files": map[string]interface{}{
				"SKILL.md": map[string]interface{}{
					"filename": "SKILL.md",
					"content":  manifest,
				},
				"config.yaml": map[string]interface{}{
					"filename": "config.yaml",
					"content":  "key: value",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(gist)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	src := &ImportSource{
		Type:   SourceGitHubGist,
		RawURL: "https://gist.github.com/alice/abc123",
		GistID: "abc123",
	}

	fetched, err := FetchSkillWithBaseURL(srv.Client(), src, srv.URL)
	require.NoError(t, err)
	assert.Len(t, fetched.Files, 2)
	assert.Equal(t, []byte(manifest), fetched.Files["SKILL.md"])
	assert.Equal(t, []byte("key: value"), fetched.Files["config.yaml"])
}

func TestFetchSkill_GitHubGist_NoManifest(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/gists/no-manifest", func(w http.ResponseWriter, _ *http.Request) {
		gist := map[string]interface{}{
			"files": map[string]interface{}{
				"README.md": map[string]interface{}{
					"filename": "README.md",
					"content":  "just a readme",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(gist)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	src := &ImportSource{
		Type:   SourceGitHubGist,
		GistID: "no-manifest",
	}

	_, err := FetchSkillWithBaseURL(srv.Client(), src, srv.URL)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no SKILL.md found")
}

func TestParseManifestFromBytes(t *testing.T) {
	data := []byte(validManifestContent())

	s, err := ParseManifestFromBytes(data)
	require.NoError(t, err)
	assert.Equal(t, "test-skill", s.Name)
	assert.Equal(t, "A test skill for import testing", s.Description)
	assert.Equal(t, "alice", s.Metadata.Author.Name)
	assert.Equal(t, "1.0.0", s.Metadata.Version)
	assert.Contains(t, s.Body, "Do the thing.")
}

func TestParseManifestFromBytes_Empty(t *testing.T) {
	_, err := ParseManifestFromBytes([]byte(""))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestParseManifestFromBytes_InvalidYAML(t *testing.T) {
	data := []byte("---\n: invalid yaml [\n---\n\nbody\n")
	_, err := ParseManifestFromBytes(data)
	require.Error(t, err)
}
