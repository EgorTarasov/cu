package searchcourses

import (
	"context"
	"fmt"
	"strings"

	cuGw "github.com/EgorTarasov/cu/internal/gateway/cu"
	mcpfmt "github.com/EgorTarasov/cu/internal/mcp/format"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// LMSClient defines dependencies for this tool.
type LMSClient interface {
	GetAllCourses(ctx context.Context) (active, archived []cuGw.StudentCourse, err error)
}

// Input for the tool.
type Input struct {
	Query string `json:"query" jsonschema:"Search query (case-insensitive substring match)"`
}

// Definition is the MCP tool definition.
var Definition = &mcp.Tool{
	Name:        "search_courses",
	Description: "Search courses by name (case-insensitive substring) across both active and archived courses; results are tagged with status",
}

// NewHandler creates the tool handler.
func NewHandler(lms LMSClient) func(context.Context, *mcp.CallToolRequest, Input) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in Input) (*mcp.CallToolResult, any, error) {
		active, archived, err := lms.GetAllCourses(ctx)
		if err != nil {
			return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
		}

		query := strings.ToLower(in.Query)
		return textResult(mcpfmt.SearchResults(
			filterByName(active, query),
			filterByName(archived, query),
			in.Query,
		)), nil, nil
	}
}

func filterByName(items []cuGw.StudentCourse, lowerQuery string) []cuGw.StudentCourse {
	var out []cuGw.StudentCourse
	for _, c := range items {
		if strings.Contains(strings.ToLower(c.Name), lowerQuery) {
			out = append(out, c)
		}
	}
	return out
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}
