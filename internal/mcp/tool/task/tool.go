package task

import (
	"context"
	"fmt"

	mcpfmt "cu-sync/internal/mcp/format"
	"cu-sync/internal/model"
	taskUC "cu-sync/internal/usecase/task"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Input struct {
	TaskID int `json:"task_id" jsonschema:"Task ID (found in deadlines or grades output)"`
}

var Definition = &mcp.Tool{
	Name:        "get_task",
	Description: "Get detailed task info: score, reviewer, solution URL, deadline, state",
}

func NewHandler(lms taskUC.LMSClient) func(context.Context, *mcp.CallToolRequest, Input) (*mcp.CallToolResult, any, error) {
	uc := taskUC.New(lms)
	return func(_ context.Context, _ *mcp.CallToolRequest, in Input) (*mcp.CallToolResult, any, error) {
		ctx := context.Background()

		out, err := uc.Get(ctx, model.TaskGetInput{TaskID: in.TaskID})
		if err != nil {
			return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
		}

		return textResult(mcpfmt.Task(out)), nil, nil
	}
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}
