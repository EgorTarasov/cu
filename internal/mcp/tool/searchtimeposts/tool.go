package searchtimeposts

import (
	"context"
	"fmt"

	mcpfmt "github.com/EgorTarasov/cu/internal/mcp/format"
	"github.com/EgorTarasov/cu/internal/model"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type TimeChat interface {
	Search(ctx context.Context, terms string, limit int) (*model.TimeSearchOutput, error)
}

type Input struct {
	Terms string `json:"terms"           jsonschema:"Mattermost search query. Words match whole words; supports \"exact phrase\", from:username, in:channel-name, after:YYYY-MM-DD, before:YYYY-MM-DD, on:YYYY-MM-DD, and @username for mentions"`
	Limit int    `json:"limit,omitempty" jsonschema:"Max results (default 20, max 100)"`
}

var Definition = &mcp.Tool{
	Name: "search_time_posts",
	Description: "Search messages in time.cu.ru (CU Mattermost) across the team: course channels, " +
		"direct and group messages. Returns matches with author, time, channel and post_id. Read-only. " +
		"Search @<your username> to find mentions; open a match's thread via read_time_thread.",
}

func NewHandler(chat TimeChat) func(context.Context, *mcp.CallToolRequest, Input) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in Input) (*mcp.CallToolResult, any, error) {
		out, err := chat.Search(ctx, in.Terms, in.Limit)
		if err != nil {
			return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
		}
		return textResult(mcpfmt.TimeSearch(out)), nil, nil
	}
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: text}}}
}
