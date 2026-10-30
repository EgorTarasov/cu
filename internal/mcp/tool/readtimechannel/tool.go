package readtimechannel

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
	ChannelPosts(ctx context.Context, query string, since *time.Time, limit int) (*model.TimeChannelPosts, error)
}

type Input struct {
	Channel string `json:"channel"         jsonschema:"Channel title (e.g. MSc-DBMS-autumn-2026), DM partner as @username or name, or channel_id from list_time_channels"`
	Since   string `json:"since,omitempty" jsonschema:"Only messages after this moment: 24h, 3d, 2026-10-29, 2026-10-29 18:00 or RFC 3339. Default: latest messages"`
	Limit   int    `json:"limit,omitempty" jsonschema:"Max messages to return, newest kept (default 30, max 200)"`
}

var Definition = &mcp.Tool{
	Name: "read_time_channel",
	Description: "Read messages of one time.cu.ru channel or direct message in chronological order, " +
		"with author, time, post_id and thread info. Read-only: does not mark anything as read. " +
		"Find channels via list_time_channels; open a thread via read_time_thread.",
}

func NewHandler(chat TimeChat) func(context.Context, *mcp.CallToolRequest, Input) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in Input) (*mcp.CallToolResult, any, error) {
		since, err := timechat.ParseSince(in.Since, time.Now())
		if err != nil {
			return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
		}
		out, err := chat.ChannelPosts(ctx, in.Channel, since, in.Limit)
		if err != nil {
			return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
		}
		return textResult(mcpfmt.TimeChannelPosts(out)), nil, nil
	}
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: text}}}
}
