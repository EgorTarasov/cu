package studentprofile

import (
	"context"
	"fmt"

	cuGw "cu-sync/internal/gateway/cu"
	mcpfmt "cu-sync/internal/mcp/format"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// LMSClient defines dependencies for this tool.
type LMSClient interface {
	GetCurrentStudent(ctx context.Context) (*cuGw.Student, error)
}

// Input for the tool (empty for current student).
type Input struct{}

// Definition is the MCP tool definition.
var Definition = &mcp.Tool{
	Name:        "get_student_profile",
	Description: "Get current student profile details: name, email, study level, late days balance",
}

// NewHandler creates the tool handler.
func NewHandler(lms LMSClient) func(context.Context, *mcp.CallToolRequest, Input) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, _ Input) (*mcp.CallToolResult, any, error) {
		student, err := lms.GetCurrentStudent(ctx)
		if err != nil {
			return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
		}

		return textResult(mcpfmt.StudentProfile(student)), nil, nil
	}
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}
