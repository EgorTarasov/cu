package downloadmaterials

import (
	"context"
	"fmt"
	"strings"

	cuGw "cu-sync/internal/gateway/cu"
	"cu-sync/internal/model"
	ucMaterials "cu-sync/internal/usecase/materials"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// LMSClient defines LMS dependencies for this tool.
type LMSClient interface {
	ResolveCourse(ctx context.Context, query string) (int, string, error)
	GetCourseOverview(ctx context.Context, courseID int) (*cuGw.CourseOverview, error)
	GetLongReadContent(ctx context.Context, longReadID int) (*cuGw.MaterialsResponse, error)
	DownloadFile(ctx context.Context, material cuGw.Material, destDir string) (string, error)
}

// GitLabClient defines GitLab dependencies for this tool.
type GitLabClient interface {
	DownloadGitLabLink(ctx context.Context, link, destDir string) ([]string, error)
}

// Input for the tool.
type Input struct {
	Course string `json:"course" jsonschema:"Course name or ID"`
	Week   int    `json:"week,omitempty" jsonschema:"Week number (optional)"`
	Path   string `json:"path" jsonschema:"Directory to save files to"`
}

// Definition is the MCP tool definition.
var Definition = &mcp.Tool{
	Name:        "download_materials",
	Description: "Download course PDFs and markdown files to a local directory",
}

// NewHandler creates the tool handler.
func NewHandler(lms LMSClient, gitlab GitLabClient) func(context.Context, *mcp.CallToolRequest, Input) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, in Input) (*mcp.CallToolResult, any, error) {
		ctx := context.Background()
		uc := ucMaterials.New(lms, gitlab)

		var events []string
		onEvent := func(event model.MaterialEvent) {
			switch event.Type {
			case model.MaterialEventSaved:
				events = append(events, fmt.Sprintf("saved: %s", event.Message))
			case model.MaterialEventError:
				events = append(events, fmt.Sprintf("error: %s", event.Message))
			case model.MaterialEventTheme, model.MaterialEventPDF, model.MaterialEventLink:
				// skip verbose events
			}
		}

		result, err := uc.Download(ctx, model.MaterialsDownloadInput{
			CourseQuery: in.Course,
			WeekFilter:  in.Week,
			BasePath:    in.Path,
		}, onEvent)
		if err != nil {
			return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
		}

		var b strings.Builder
		b.WriteString("## Download Complete\n\n")
		b.WriteString(fmt.Sprintf("Downloaded %d/%d files to `%s`\n\n", result.DownloadedFiles, result.TotalFiles, in.Path))

		for _, e := range events {
			b.WriteString(fmt.Sprintf("- %s\n", e))
		}

		return textResult(b.String()), nil, nil
	}
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}
