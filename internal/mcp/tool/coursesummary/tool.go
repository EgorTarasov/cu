package coursesummary

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
	GetCourse(ctx context.Context, courseID int) (*cuGw.Course, error)
}

// Input for the tool.
type Input struct {
	Course string `json:"course" jsonschema:"Course name or ID"`
}

// Definition is the MCP tool definition.
var Definition = &mcp.Tool{
	Name:        "get_course_summary",
	Description: "Get course summary info: state, category, publish dates, archived flag",
}

// NewHandler creates the tool handler.
func NewHandler(lms LMSClient) func(context.Context, *mcp.CallToolRequest, Input) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in Input) (*mcp.CallToolResult, any, error) {
		courseID, _, err := lms.ResolveCourse(ctx, in.Course)
		if err != nil {
			return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
		}

		course, err := lms.GetCourse(ctx, courseID)
		if err != nil {
			return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
		}

		return textResult(mcpfmt.CourseSummary(course)), nil, nil
	}
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}
