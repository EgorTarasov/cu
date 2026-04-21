package coursestructure

import (
	"context"
	"fmt"

	cuGw "cu-sync/internal/gateway/cu"
	mcpfmt "cu-sync/internal/mcp/format"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// LMSClient defines dependencies for this tool.
type LMSClient interface {
	ResolveCourse(ctx context.Context, query string) (int, string, error)
	GetCourseOverview(ctx context.Context, courseID int) (*cuGw.CourseOverview, error)
}

// Input for the tool.
type Input struct {
	Course string `json:"course"`
}

// Definition is the MCP tool definition.
var Definition = &mcp.Tool{
	Name:        "get_course_structure",
	Description: "Get full course tree: themes, longreads, exercises with deadlines",
}

// NewHandler creates the tool handler.
func NewHandler(lms LMSClient) func(context.Context, *mcp.CallToolRequest, Input) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, in Input) (*mcp.CallToolResult, any, error) {
		ctx := context.Background()

		courseID, _, err := lms.ResolveCourse(ctx, in.Course)
		if err != nil {
			return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
		}

		overview, err := lms.GetCourseOverview(ctx, courseID)
		if err != nil {
			return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
		}

		return textResult(mcpfmt.CourseStructure(overview)), nil, nil
	}
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}
