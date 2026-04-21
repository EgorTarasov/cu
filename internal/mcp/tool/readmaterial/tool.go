package readmaterial

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	cuGw "cu-sync/internal/gateway/cu"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// LMSClient defines LMS dependencies for this tool.
type LMSClient interface {
	ResolveCourse(ctx context.Context, query string) (int, string, error)
	GetCourseOverview(ctx context.Context, courseID int) (*cuGw.CourseOverview, error)
	GetLongReadContent(ctx context.Context, longReadID int) (*cuGw.MaterialsResponse, error)
}

// GitLabClient defines GitLab dependencies for this tool.
type GitLabClient interface {
	GetRawFile(ctx context.Context, project, ref, filePath string) ([]byte, error)
}

// Input for the tool.
type Input struct {
	Course string `json:"course" jsonschema:"Course name or ID"`
	Week   int    `json:"week" jsonschema:"Week number"`
	Type   string `json:"type,omitempty" jsonschema:"Material type: longread or seminar (optional — returns all markdown links)"`
}

// Definition is the MCP tool definition.
var Definition = &mcp.Tool{
	Name:        "read_material",
	Description: "Read a course material (longread/seminar) from GitLab and return its markdown content",
}

// NewHandler creates the tool handler.
func NewHandler(lms LMSClient, gitlab GitLabClient) func(context.Context, *mcp.CallToolRequest, Input) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, in Input) (*mcp.CallToolResult, any, error) {
		ctx := context.Background()

		if gitlab == nil {
			return textResult("Error: GitLab not configured. Run `cu login --gitlab` first."), nil, nil
		}

		courseID, _, err := lms.ResolveCourse(ctx, in.Course)
		if err != nil {
			return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
		}

		overview, err := lms.GetCourseOverview(ctx, courseID)
		if err != nil {
			return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
		}

		gitLinks := collectGitLinks(ctx, lms, overview, in.Week, in.Type)
		if len(gitLinks) == 0 {
			return textResult(fmt.Sprintf("No GitLab materials found for week %d.", in.Week)), nil, nil
		}

		content := fetchGitContent(ctx, gitlab, gitLinks)
		return textResult(content), nil, nil
	}
}

func collectGitLinks(
	ctx context.Context, lms LMSClient, overview *cuGw.CourseOverview, week int, typeFilter string,
) []string {
	var gitLinks []string

	for _, theme := range overview.Themes {
		if !matchesWeek(theme.Name, week) {
			continue
		}
		for _, lr := range theme.Longreads {
			mats, err := lms.GetLongReadContent(ctx, lr.ID)
			if err != nil {
				continue
			}
			for _, mat := range mats.Items {
				if mat.Type != "markdown" || mat.ViewContent == "" {
					continue
				}
				for _, link := range extractLinks(mat.ViewContent) {
					if !cuGw.IsGitLabLink(link) {
						continue
					}
					if typeFilter != "" && !strings.Contains(strings.ToLower(link), strings.ToLower(typeFilter)) {
						continue
					}
					gitLinks = append(gitLinks, link)
				}
			}
		}
	}

	return gitLinks
}

func fetchGitContent(ctx context.Context, gitlab GitLabClient, links []string) string {
	var b strings.Builder

	for _, link := range links {
		_, ref, path, _, ok := cuGw.ParseGitLabLink(link)
		if !ok {
			continue
		}
		if idx := strings.IndexByte(path, '?'); idx >= 0 {
			path = path[:idx]
		}
		project, _, _, _, _ := cuGw.ParseGitLabLink(link)

		data, err := gitlab.GetRawFile(ctx, project, ref, path)
		if err != nil {
			b.WriteString(fmt.Sprintf("<!-- Error fetching %s: %v -->\n\n", link, err))
			continue
		}

		b.WriteString(fmt.Sprintf("---\n## Source: %s\n\n", link))
		b.Write(data)
		b.WriteString("\n\n")
	}

	return b.String()
}

var weekRe = regexp.MustCompile(`(?i)(?:неделя|week)\s*(\d+)`)

func matchesWeek(themeName string, week int) bool {
	m := weekRe.FindStringSubmatch(themeName)
	if len(m) < 2 { //nolint:mnd // regex match minimum parts
		return false
	}
	n, err := strconv.Atoi(m[1])
	return err == nil && n == week
}

var linkPattern = regexp.MustCompile(`href=\\"([^"\\]+)\\"`)

func extractLinks(viewContent string) []string {
	matches := linkPattern.FindAllStringSubmatch(viewContent, -1)
	var links []string
	seen := make(map[string]bool)

	for _, m := range matches {
		link := m[1]
		if strings.HasPrefix(link, "#") || strings.Contains(link, "my.centraluniversity.ru") {
			continue
		}
		if !seen[link] {
			seen[link] = true
			links = append(links, link)
		}
	}

	return links
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}
