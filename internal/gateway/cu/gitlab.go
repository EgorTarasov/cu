package cu

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const gitlabLinkMinParts = 5

// GitLabClient provides access to git.culab.ru using session cookie.
type GitLabClient struct {
	httpClient    *http.Client
	sessionCookie string
}

// NewGitLabClient creates a client with a _gitlab_session cookie.
func NewGitLabClient(sessionCookie string) *GitLabClient {
	return &GitLabClient{
		httpClient:    &http.Client{Timeout: DefaultTimeout},
		sessionCookie: sessionCookie,
	}
}

// NewGitLabClientFromEnv loads the GitLab session cookie from env or file.
func NewGitLabClientFromEnv() (*GitLabClient, error) {
	cookie := os.Getenv("CU_GITLAB_COOKIE")
	if cookie == "" {
		saved, err := LoadGitLabCookie()
		if err != nil || saved == "" {
			return nil, errors.New("no GitLab cookie found. Run: cu login --gitlab")
		}
		cookie = saved
	}
	return NewGitLabClient(cookie), nil
}

// gitlabLinkPattern matches git.culab.ru blob/tree links.
var gitlabLinkPattern = regexp.MustCompile(
	`^https?://git\.culab\.ru/([^/]+/[^/]+)/-/(blob|tree)/([^/]+)/(.+)$`,
)

// ParseGitLabLink extracts project, ref, and path from a git.culab.ru link.
func ParseGitLabLink(link string) (string, string, string, bool, bool) {
	m := gitlabLinkPattern.FindStringSubmatch(link)
	if len(m) < gitlabLinkMinParts {
		return "", "", "", false, false
	}
	return m[1], m[3], m[4], m[2] == "tree", true
}

// IsGitLabLink checks if a URL is a git.culab.ru link.
func IsGitLabLink(link string) bool {
	return gitlabLinkPattern.MatchString(link)
}

func (g *GitLabClient) prepareRequest(ctx context.Context, rawURL string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.AddCookie(&http.Cookie{
		Name:  "_gitlab_session",
		Value: g.sessionCookie,
	})
	return req, nil
}

// GetRawFile downloads a single file from GitLab via the raw endpoint.
func (g *GitLabClient) GetRawFile(ctx context.Context, project, ref, filePath string) ([]byte, error) {
	// Strip query params from path.
	if idx := strings.IndexByte(filePath, '?'); idx >= 0 {
		filePath = filePath[:idx]
	}

	rawURL := fmt.Sprintf("%s/%s/-/raw/%s/%s",
		GitLabBaseURL,
		project,
		url.PathEscape(ref),
		filePath,
	)

	req, err := g.prepareRequest(ctx, rawURL)
	if err != nil {
		return nil, err
	}

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab %d for %s/%s", resp.StatusCode, project, filePath)
	}

	return io.ReadAll(resp.Body)
}

// DownloadGitLabLink downloads content from a git.culab.ru link and saves to destDir.
// For blob links — downloads the single file.
// For tree links — downloads all text files in the directory via GitLab API.
func (g *GitLabClient) DownloadGitLabLink(ctx context.Context, link, destDir string) ([]string, error) {
	project, ref, path, isTree, ok := ParseGitLabLink(link)
	if !ok {
		return nil, fmt.Errorf("not a valid git.culab.ru link: %s", link)
	}

	// Strip query params.
	if idx := strings.IndexByte(path, '?'); idx >= 0 {
		path = path[:idx]
	}

	if err := os.MkdirAll(destDir, 0o750); err != nil {
		return nil, err
	}

	if !isTree {
		data, err := g.GetRawFile(ctx, project, ref, path)
		if err != nil {
			return nil, err
		}
		filename := filepath.Base(path)
		destPath := filepath.Join(destDir, filename)
		if err := os.WriteFile(destPath, data, 0o600); err != nil {
			return nil, err
		}
		return []string{destPath}, nil
	}

	// Tree — list via API and download each file.
	entries, err := g.listTree(ctx, project, ref, path)
	if err != nil {
		return nil, err
	}

	var saved []string
	for _, entry := range entries {
		if entry.Type != "blob" {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name))
		switch ext {
		case ".md", ".go", ".java", ".txt", ".yaml", ".yml", ".json", ".xml", ".sql", ".sh", ".py":
			data, err := g.GetRawFile(ctx, project, ref, entry.Path)
			if err != nil {
				continue
			}
			destPath := filepath.Join(destDir, entry.Name)
			if err := os.WriteFile(destPath, data, 0o600); err != nil {
				continue
			}
			saved = append(saved, destPath)
		}
	}
	return saved, nil
}

type treeEntry struct {
	Name string `json:"name"`
	Type string `json:"type"` // "blob" or "tree"
	Path string `json:"path"`
}

func decodeJSON(r io.Reader, v any) error {
	return json.NewDecoder(r).Decode(v)
}

func (g *GitLabClient) listTree(ctx context.Context, project, ref, path string) ([]treeEntry, error) {
	encodedProject := url.PathEscape(project)
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/repository/tree?ref=%s&path=%s&per_page=100",
		GitLabBaseURL, encodedProject, url.QueryEscape(ref), url.QueryEscape(path))

	req, err := g.prepareRequest(ctx, apiURL)
	if err != nil {
		return nil, err
	}

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API %d for tree %s/%s", resp.StatusCode, project, path)
	}

	var entries []treeEntry
	if err := decodeJSON(resp.Body, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}
