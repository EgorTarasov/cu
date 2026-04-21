package deadlines

import (
	"context"
	"fmt"

	cuGw "cu-sync/internal/gateway/cu"
	mcpfmt "cu-sync/internal/mcp/format"
	"cu-sync/internal/model"
	ucDeadlines "cu-sync/internal/usecase/deadlines"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// LMSClient defines dependencies for this tool.
type LMSClient interface {
	ResolveCourse(ctx context.Context, query string) (int, string, error)
	GetDeadlines(ctx context.Context, limit int, courseID *int) ([]cuGw.Deadline, error)
}

// Input for the tool.
type Input struct {
	Course string `json:"course,omitempty" jsonschema:"description=Course name or ID (optional — omit for all courses)"`
}

// Definition is the MCP tool definition.
var Definition = &mcp.Tool{
	Name:        "list_deadlines",
	Description: "List upcoming deadlines sorted by date, with urgency markers. Optionally filter by course.",
}

// NewHandler creates the tool handler.
func NewHandler(lms LMSClient) func(context.Context, *mcp.CallToolRequest, Input) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, in Input) (*mcp.CallToolResult, any, error) {
		ctx := context.Background()
		uc := ucDeadlines.New(lms)

		result, err := uc.List(ctx, model.DeadlinesListInput{CourseQuery: in.Course})
		if err != nil {
			return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
		}

		return textResult(mcpfmt.Deadlines(result)), nil, nil
	}
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}
