package timeunread

import (
	"context"
	"fmt"
	"time"

	mcpfmt "github.com/EgorTarasov/cu/internal/mcp/format"
	"github.com/EgorTarasov/cu/internal/model"
	"github.com/EgorTarasov/cu/internal/usecase/timechat"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type TimeChat interface {
	Updates(ctx context.Context, in model.TimeUpdatesInput) (*model.TimeUpdatesOutput, error)
}

type Input struct {
	Since         string `json:"since,omitempty"           jsonschema:"Instead of unread, return every message after this moment, read or not: 24h, 3d, 2026-10-29, 2026-10-29 18:00 or RFC 3339"`
	Kind          string `json:"kind,omitempty"            jsonschema:"Filter: dm (direct and group messages), group, channels (public and private), public, private, or all (default)"`
	IncludeMuted  bool   `json:"include_muted,omitempty"   jsonschema:"Also read muted channels (by default they are only listed unless they mention the user)"`
	MaxPerChannel int    `json:"max_per_channel,omitempty" jsonschema:"Max messages per channel, newest kept (default 20, max 200)"`
}

var Definition = &mcp.Tool{
	Name: "get_time_unread",
	Description: "Get new messages across time.cu.ru (CU Mattermost) in one call, grouped by channel: " +
		"unread messages and mentions by default, or everything after `since`. Covers direct messages, " +
		"group chats and course channels. Read-only: does not mark anything as read.",
}

func NewHandler(chat TimeChat) func(context.Context, *mcp.CallToolRequest, Input) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in Input) (*mcp.CallToolResult, any, error) {
		since, err := timechat.ParseSince(in.Since, time.Now())
		if err != nil {
			return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
		}
		kinds, err := timechat.ParseKind(in.Kind)
		if err != nil {
			return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
		}
		out, err := chat.Updates(ctx, model.TimeUpdatesInput{
			Since: since, Kinds: kinds, IncludeMuted: in.IncludeMuted, MaxPerChannel: in.MaxPerChannel,
		})
		if err != nil {
			return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
		}
		return textResult(mcpfmt.TimeUpdates(out)), nil, nil
	}
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: text}}}
}
