package themesummary

import (
	"context"
	"fmt"

	cuGw "cu-sync/internal/gateway/cu"
	mcpfmt "cu-sync/internal/mcp/format"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// LMSClient defines dependencies for this tool.
type LMSClient interface {
	GetTheme(ctx context.Context, themeID int) (*cuGw.Theme, error)
}

// Input for the tool.
type Input struct {
	ThemeID int `json:"theme_id" jsonschema:"Numeric theme ID (obtain from get_course_structure)"`
}

// Definition is the MCP tool definition.
var Definition = &mcp.Tool{
	Name: "get_theme_summary",
	Description: "Get theme summary info: name, state, order, publish dates. " +
		"Requires a numeric theme_id — discover it via get_course_structure.",
}

// NewHandler creates the tool handler.
func NewHandler(lms LMSClient) func(context.Context, *mcp.CallToolRequest, Input) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in Input) (*mcp.CallToolResult, any, error) {
		theme, err := lms.GetTheme(ctx, in.ThemeID)
		if err != nil {
			return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
		}

		return textResult(mcpfmt.ThemeSummary(theme)), nil, nil
	}
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}
