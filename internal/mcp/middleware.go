package mcp

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/EgorTarasov/cu/internal/telemetry"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// telemetryMiddleware instruments the JSON-RPC layer: session start on
// initialize, one event per tools/call with sequence number and timing.
// Tool handlers stay telemetry-free.
func (s *Server) telemetryMiddleware() mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			switch method {
			case "initialize":
				s.trackInitialize(req)
				return next(ctx, method, req)
			case "tools/call":
				return s.trackToolCall(ctx, method, req, next)
			default:
				return next(ctx, method, req)
			}
		}
	}
}

func (s *Server) trackInitialize(req mcp.Request) {
	var name, ver string
	if r, ok := req.(*mcp.InitializeRequest); ok && r.Params != nil &&
		r.Params.ClientInfo != nil {
		name = r.Params.ClientInfo.Name
		ver = r.Params.ClientInfo.Version
	}
	s.clientName.Store(name)
	telemetry.Default().MCPSessionStarted(s.sessionID, name, ver)
}

func (s *Server) trackToolCall(
	ctx context.Context,
	method string,
	req mcp.Request,
	next mcp.MethodHandler,
) (mcp.Result, error) {
	start := time.Now()
	seq := s.toolCalls.Add(1)
	prevMS := s.lastToolCallMS.Swap(start.UnixMilli())

	var sincePrev time.Duration
	if prevMS != 0 {
		sincePrev = time.Duration(start.UnixMilli()-prevMS) * time.Millisecond
	}

	res, err := next(ctx, method, req)

	tool, argKeys := callDetails(req)
	clientName, _ := s.clientName.Load().(string)
	telemetry.Default().MCPToolCalled(telemetry.ToolCallEvent{
		SessionID:  s.sessionID,
		ClientName: clientName,
		Tool:       tool,
		Seq:        int(seq),
		SincePrev:  sincePrev,
		Duration:   time.Since(start),
		IsError:    err != nil || isErrorResult(res),
		ArgKeys:    argKeys,
	})

	return res, err
}

func callDetails(req mcp.Request) (string, []string) {
	r, ok := req.(*mcp.CallToolRequest)
	if !ok || r.Params == nil {
		return "unknown", nil
	}
	var argKeys []string
	if len(r.Params.Arguments) > 0 {
		var args map[string]json.RawMessage
		if json.Unmarshal(r.Params.Arguments, &args) == nil {
			for k := range args {
				argKeys = append(argKeys, k)
			}
			sort.Strings(argKeys)
		}
	}
	return r.Params.Name, argKeys
}

// isErrorResult detects failures our tools report as plain text: they return
// nil error with a "Error: ..." message so the LLM sees the reason.
func isErrorResult(res mcp.Result) bool {
	r, ok := res.(*mcp.CallToolResult)
	if !ok {
		return false
	}
	if r.IsError {
		return true
	}
	for _, content := range r.Content {
		if text, ok := content.(*mcp.TextContent); ok {
			return strings.HasPrefix(text.Text, "Error:")
		}
	}
	return false
}
