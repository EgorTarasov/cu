package searchcourses

import (
	"context"
	"fmt"
	"strings"

	cuGw "cu-sync/internal/gateway/cu"
	mcpfmt "cu-sync/internal/mcp/format"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const maxCoursesLimit = 10000

// LMSClient defines dependencies for this tool.
type LMSClient interface {
	GetStudentCourses(ctx context.Context, limit int, state string) (*cuGw.StudentCoursesResponse, error)
}

// Input for the tool.
type Input struct {
	Query string `json:"query" jsonschema:"description=Search query (case-insensitive substring match)"`
}

// Definition is the MCP tool definition.
var Definition = &mcp.Tool{
	Name:        "search_courses",
	Description: "Search courses by name (case-insensitive substring match)",
}

// NewHandler creates the tool handler.
func NewHandler(lms LMSClient) func(context.Context, *mcp.CallToolRequest, Input) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, in Input) (*mcp.CallToolResult, any, error) {
		ctx := context.Background()

		resp, err := lms.GetStudentCourses(ctx, maxCoursesLimit, "published")
		if err != nil {
			return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
		}

		query := strings.ToLower(in.Query)
		var matches []cuGw.StudentCourse

		for _, c := range resp.Items {
			if strings.Contains(strings.ToLower(c.Name), query) {
				matches = append(matches, c)
			}
		}

		return textResult(mcpfmt.SearchResults(matches, in.Query)), nil, nil
	}
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}
