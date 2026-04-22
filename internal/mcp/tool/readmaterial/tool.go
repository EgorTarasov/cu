package readmaterial

import (
	"context"
	"fmt"
	"strings"

	cuGw "cu-sync/internal/gateway/cu"
	"cu-sync/internal/usecase/materials"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type LMSClient interface {
	ResolveCourse(ctx context.Context, query string) (int, string, error)
	GetCourseOverview(ctx context.Context, courseID int) (*cuGw.CourseOverview, error)
	GetLongReadContent(ctx context.Context, longReadID int) (*cuGw.MaterialsResponse, error)
}

type GitLabClient interface {
	GetRawFile(ctx context.Context, project, ref, filePath string) ([]byte, error)
}

type Input struct {
	Course string `json:"course" jsonschema:"Course name or ID"`
	Week   int    `json:"week" jsonschema:"Week number"`
	Type   string `json:"type,omitempty" jsonschema:"Material type: longread or seminar (optional — returns all markdown links)"`
}

var Definition = &mcp.Tool{
	Name:        "read_material",
	Description: "Read a course material (longread/seminar) from GitLab and return its markdown content",
}

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
		if !materials.MatchesWeek(theme.Name, week) {
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
				for _, link := range materials.ExtractLinks(mat.ViewContent) {
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
			_, _ = fmt.Fprintf(&b, "Error: %v", err)
			fmt.Fprintf(&b, "<!-- Error fetching %s: %v -->\n\n", link, err)
			continue
		}

		fmt.Fprintf(&b, "---\n## Source: %s\n\n", link)
		b.Write(data)
		b.WriteString("\n\n")
	}

	return b.String()
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}
