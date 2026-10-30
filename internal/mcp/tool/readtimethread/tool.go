package readtimethread

import (
	"context"
	"fmt"

	mcpfmt "github.com/EgorTarasov/cu/internal/mcp/format"
	"github.com/EgorTarasov/cu/internal/model"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type TimeChat interface {
	Thread(ctx context.Context, postID string) (*model.TimeChannelPosts, error)
}

type Input struct {
	PostID string `json:"post_id" jsonschema:"post_id of the root message or any reply, as shown by read_time_channel, get_time_unread or search_time_posts"`
}

var Definition = &mcp.Tool{
	Name: "read_time_thread",
	Description: "Read a whole time.cu.ru thread: the root message and all replies in order. Read-only. " +
		"Requires a post_id from read_time_channel, get_time_unread or search_time_posts.",
}

func NewHandler(chat TimeChat) func(context.Context, *mcp.CallToolRequest, Input) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in Input) (*mcp.CallToolResult, any, error) {
		if in.PostID == "" {
			return textResult("Error: post_id is required"), nil, nil
		}
		out, err := chat.Thread(ctx, in.PostID)
		if err != nil {
			return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
		}
		return textResult(mcpfmt.TimeThread(out)), nil, nil
	}
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: text}}}
}
