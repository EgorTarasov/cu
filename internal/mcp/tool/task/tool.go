package task

import (
	"context"
	"fmt"

	cuGw "cu-sync/internal/gateway/cu"
	mcpfmt "cu-sync/internal/mcp/format"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// LMSClient defines dependencies for this tool.
type LMSClient interface {
	GetTask(ctx context.Context, taskID int) (*cuGw.Task, error)
}

// Input for the tool.
type Input struct {
	TaskID int `json:"task_id" jsonschema:"Task ID (found in deadlines or grades output)"`
}

// Definition is the MCP tool definition.
var Definition = &mcp.Tool{
	Name:        "get_task",
	Description: "Get detailed task info: score, reviewer, solution URL, deadline, state",
}

// NewHandler creates the tool handler.
func NewHandler(lms LMSClient) func(context.Context, *mcp.CallToolRequest, Input) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, in Input) (*mcp.CallToolResult, any, error) {
		ctx := context.Background()

		t, err := lms.GetTask(ctx, in.TaskID)
		if err != nil {
			return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
		}

		return textResult(mcpfmt.Task(t)), nil, nil
	}
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}
