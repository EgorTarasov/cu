package listcourses

import (
	"context"
	"fmt"

	cuGw "github.com/EgorTarasov/cu/internal/gateway/cu"
	mcpfmt "github.com/EgorTarasov/cu/internal/mcp/format"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// LMSClient defines dependencies for this tool.
type LMSClient interface {
	GetAllCourses(ctx context.Context) (active, archived []cuGw.StudentCourse, err error)
}

// Input for the tool (empty for list_courses).
type Input struct{}

// Definition is the MCP tool definition.
var Definition = &mcp.Tool{
	Name:        "list_courses",
	Description: "List student courses split into Active and Archived sections, with IDs and categories",
}

// NewHandler creates the tool handler.
func NewHandler(lms LMSClient) func(context.Context, *mcp.CallToolRequest, Input) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, _ Input) (*mcp.CallToolResult, any, error) {
		active, archived, err := lms.GetAllCourses(ctx)
		if err != nil {
			return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
		}

		return textResult(mcpfmt.CoursesList(active, archived)), nil, nil
	}
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}
