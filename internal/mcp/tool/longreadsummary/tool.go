package longreadsummary

import (
	"context"
	"fmt"

	cuGw "github.com/EgorTarasov/cu/internal/gateway/cu"
	mcpfmt "github.com/EgorTarasov/cu/internal/mcp/format"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// LMSClient defines dependencies for this tool.
type LMSClient interface {
	GetLongread(ctx context.Context, longreadID int) (*cuGw.Longread, error)
}

// Input for the tool.
type Input struct {
	LongreadID int `json:"longread_id" jsonschema:"Numeric longread ID (obtain from get_course_structure)"`
}

// Definition is the MCP tool definition.
var Definition = &mcp.Tool{
	Name: "get_longread_summary",
	Description: "Get longread summary info: name, type, state, order, publish dates. " +
		"Requires a numeric longread_id — discover it via get_course_structure.",
}

// NewHandler creates the tool handler.
func NewHandler(lms LMSClient) func(context.Context, *mcp.CallToolRequest, Input) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in Input) (*mcp.CallToolResult, any, error) {
		longread, err := lms.GetLongread(ctx, in.LongreadID)
		if err != nil {
			return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
		}

		return textResult(mcpfmt.LongreadSummary(longread)), nil, nil
	}
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}
