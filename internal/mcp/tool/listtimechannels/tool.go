package listtimechannels

import (
	"context"
	"fmt"

	mcpfmt "github.com/EgorTarasov/cu/internal/mcp/format"
	"github.com/EgorTarasov/cu/internal/model"
	"github.com/EgorTarasov/cu/internal/usecase/timechat"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type TimeChat interface {
	ListChannels(ctx context.Context, f timechat.ChannelsFilter) ([]model.TimeChannel, error)
}

type Input struct {
	Kind       string `json:"kind,omitempty"        jsonschema:"Filter: dm (direct and group messages), group, channels (public and private), public, private, or all (default)"`
	UnreadOnly bool   `json:"unread_only,omitempty" jsonschema:"Only channels with unread messages or mentions"`
}

var Definition = &mcp.Tool{
	Name: "list_time_channels",
	Description: "List the user's time.cu.ru (CU Mattermost) course channels, direct messages and group chats " +
		"with unread and mention counters, muted flag, last message time and channel_id. " +
		"Read-only. Use the title, @username or channel_id with read_time_channel.",
}

func NewHandler(chat TimeChat) func(context.Context, *mcp.CallToolRequest, Input) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in Input) (*mcp.CallToolResult, any, error) {
		kinds, err := timechat.ParseKind(in.Kind)
		if err != nil {
			return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
		}
		chs, err := chat.ListChannels(ctx, timechat.ChannelsFilter{Kinds: kinds, UnreadOnly: in.UnreadOnly})
		if err != nil {
			return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
		}
		return textResult(mcpfmt.TimeChannels(chs)), nil, nil
	}
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: text}}}
}
