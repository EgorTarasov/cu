package listcourses

import (
	"context"
	"fmt"

	cuGw "cu-sync/internal/gateway/cu"
	mcpfmt "cu-sync/internal/mcp/format"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const maxCoursesLimit = 10000

// LMSClient defines dependencies for this tool.
type LMSClient interface {
	GetStudentCourses(ctx context.Context, limit int, state string) (*cuGw.StudentCoursesResponse, error)
}

// Input for the tool (empty for list_courses).
type Input struct{}

// Definition is the MCP tool definition.
var Definition = &mcp.Tool{
	Name:        "list_courses",
	Description: "List all student courses with IDs and categories",
}

// NewHandler creates the tool handler.
func NewHandler(lms LMSClient) func(context.Context, *mcp.CallToolRequest, Input) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, _ Input) (*mcp.CallToolResult, any, error) {
		ctx := context.Background()

		resp, err := lms.GetStudentCourses(ctx, maxCoursesLimit, "published")
		if err != nil {
			return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
		}

		return textResult(mcpfmt.CoursesList(resp.Items)), nil, nil
	}
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}
