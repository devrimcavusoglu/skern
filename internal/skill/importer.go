package skill

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// maxDownloadSize caps the size of any single fetched payload (API JSON or
// companion file) to protect against malicious or misconfigured remotes that
// could otherwise return an unbounded body.
const maxDownloadSize = 10 * 1024 * 1024 // 10 MiB

// SourceType identifies the kind of remote source for skill import.
type SourceType string

const (
	// SourceGitHubRepo is a GitHub repository directory URL.
	SourceGitHubRepo SourceType = "github-repo"
	// SourceGitHubGist is a GitHub gist URL.
	SourceGitHubGist SourceType = "github-gist"
)

// ImportSource holds parsed information about a remote skill source.
type ImportSource struct {
	Type   SourceType
	RawURL string
	// GitHub repo fields
	Owner string
	Repo  string
	Ref   string
	Path  string
	// GitHub gist fields
	GistID string
}

// FetchedSkill holds the downloaded files from a remote skill source.
type FetchedSkill struct {
	Files map[string][]byte // filename -> content
}

// HTTPClient is the interface for making HTTP requests, allowing test injection.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// ParseImportURL parses a URL and returns an ImportSource describing the remote skill location.
// Supported formats:
//   - GitHub repo directory: https://github.com/<owner>/<repo>/tree/<ref>/<path>
//   - GitHub gist: https://gist.github.com/<user>/<gist-id>
func ParseImportURL(rawURL string) (*ImportSource, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("unsupported URL scheme %q; use https", u.Scheme)
	}

	host := strings.ToLower(u.Host)

	switch host {
	case "gist.github.com":
		return parseGistURL(u, rawURL)
	case "github.com":
		return parseGitHubRepoURL(u, rawURL)
	default:
		return nil, fmt.Errorf("unsupported host %q; supported sources: github.com (repo directory), gist.github.com", host)
	}
}

func parseGitHubRepoURL(u *url.URL, rawURL string) (*ImportSource, error) {
	// Expected format: /<owner>/<repo>/tree/<ref>/<path...>
	//
	// NOTE: refs containing slashes (e.g., "feature/v2") cannot be unambiguously
	// parsed from this URL form — only the first segment after `tree/` is taken
	// as the ref. Use a ref without slashes or a commit SHA for reliable imports.
	parts := strings.SplitN(strings.TrimPrefix(u.Path, "/"), "/", 5)
	if len(parts) < 4 || parts[2] != "tree" {
		return nil, fmt.Errorf("expected GitHub directory URL format: https://github.com/<owner>/<repo>/tree/<ref>/<path>")
	}

	owner := parts[0]
	repo := parts[1]
	ref := parts[3]
	path := ""
	if len(parts) == 5 {
		path = parts[4]
	}

	if owner == "" || repo == "" || ref == "" {
		return nil, fmt.Errorf("invalid GitHub URL: owner, repo, and ref are required")
	}

	return &ImportSource{
		Type:   SourceGitHubRepo,
		RawURL: rawURL,
		Owner:  owner,
		Repo:   repo,
		Ref:    ref,
		Path:   path,
	}, nil
}

func parseGistURL(u *url.URL, rawURL string) (*ImportSource, error) {
	// Expected format: /<user>/<gist-id> or /<gist-id>
	parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	if len(parts) == 0 || (len(parts) == 1 && parts[0] == "") {
		return nil, fmt.Errorf("expected GitHub gist URL format: https://gist.github.com/<user>/<gist-id>")
	}

	var gistID string
	if len(parts) >= 2 {
		gistID = parts[1]
	} else {
		gistID = parts[0]
	}

	if gistID == "" {
		return nil, fmt.Errorf("expected GitHub gist URL format: https://gist.github.com/<user>/<gist-id>")
	}

	return &ImportSource{
		Type:   SourceGitHubGist,
		RawURL: rawURL,
		GistID: gistID,
	}, nil
}

// FetchSkill downloads all files from the remote skill source.
func FetchSkill(client HTTPClient, src *ImportSource) (*FetchedSkill, error) {
	switch src.Type {
	case SourceGitHubRepo:
		return fetchFromGitHubRepo(client, src)
	case SourceGitHubGist:
		return fetchFromGitHubGist(client, src)
	default:
		return nil, fmt.Errorf("unsupported source type: %s", src.Type)
	}
}

// FetchSkillWithBaseURL is like FetchSkill but allows overriding the API base URL (for testing).
func FetchSkillWithBaseURL(client HTTPClient, src *ImportSource, baseURL string) (*FetchedSkill, error) {
	switch src.Type {
	case SourceGitHubRepo:
		return fetchFromGitHubRepoWithBase(client, src, baseURL)
	case SourceGitHubGist:
		return fetchFromGitHubGistWithBase(client, src, baseURL)
	default:
		return nil, fmt.Errorf("unsupported source type: %s", src.Type)
	}
}

func fetchFromGitHubRepo(client HTTPClient, src *ImportSource) (*FetchedSkill, error) {
	return fetchFromGitHubRepoWithBase(client, src, "https://api.github.com")
}

func fetchFromGitHubRepoWithBase(client HTTPClient, src *ImportSource, baseURL string) (*FetchedSkill, error) {
	// GET /repos/{owner}/{repo}/contents/{path}?ref={ref}
	apiURL := fmt.Sprintf("%s/repos/%s/%s/contents/%s?ref=%s",
		baseURL, src.Owner, src.Repo, src.Path, url.QueryEscape(src.Ref))

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching directory listing: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %d for %s", resp.StatusCode, apiURL)
	}

	var entries []struct {
		Name        string `json:"name"`
		Type        string `json:"type"`
		DownloadURL string `json:"download_url"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxDownloadSize)).Decode(&entries); err != nil {
		return nil, fmt.Errorf("parsing directory listing: %w", err)
	}

	files := make(map[string][]byte)
	hasManifest := false

	for _, entry := range entries {
		if entry.Type != "file" {
			continue // skip subdirectories
		}
		if entry.DownloadURL == "" {
			continue
		}
		if entry.Name == ManifestFile {
			hasManifest = true
		}

		content, err := downloadFile(client, entry.DownloadURL)
		if err != nil {
			return nil, fmt.Errorf("downloading %s: %w", entry.Name, err)
		}
		files[entry.Name] = content
	}

	if !hasManifest {
		return nil, fmt.Errorf("no SKILL.md found in remote directory")
	}

	return &FetchedSkill{Files: files}, nil
}

func fetchFromGitHubGist(client HTTPClient, src *ImportSource) (*FetchedSkill, error) {
	return fetchFromGitHubGistWithBase(client, src, "https://api.github.com")
}

func fetchFromGitHubGistWithBase(client HTTPClient, src *ImportSource, baseURL string) (*FetchedSkill, error) {
	apiURL := fmt.Sprintf("%s/gists/%s", baseURL, src.GistID)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching gist: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %d for gist %s", resp.StatusCode, src.GistID)
	}

	var gist struct {
		Files map[string]struct {
			Filename string `json:"filename"`
			Content  string `json:"content"`
		} `json:"files"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxDownloadSize)).Decode(&gist); err != nil {
		return nil, fmt.Errorf("parsing gist response: %w", err)
	}

	files := make(map[string][]byte)
	hasManifest := false

	for _, f := range gist.Files {
		if f.Filename == ManifestFile {
			hasManifest = true
		}
		files[f.Filename] = []byte(f.Content)
	}

	if !hasManifest {
		return nil, fmt.Errorf("no SKILL.md found in gist")
	}

	return &FetchedSkill{Files: files}, nil
}

func downloadFile(client HTTPClient, fileURL string) ([]byte, error) {
	req, err := http.NewRequest("GET", fileURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d for %s", resp.StatusCode, fileURL)
	}

	// Read at most maxDownloadSize+1 to detect oversize without buffering
	// arbitrarily large bodies.
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxDownloadSize+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxDownloadSize {
		return nil, fmt.Errorf("file %s exceeds maximum download size of %d bytes", fileURL, maxDownloadSize)
	}
	return data, nil
}
