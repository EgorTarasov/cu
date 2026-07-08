package telemetry

import "strings"

// EventName is a typed name of an analytics event. Values come only from the
// constants and constructor functions declared here.
type EventName string

const (
	// EventFirstRun fires once per device, when the device-id file is created.
	EventFirstRun EventName = "first_run"
	// EventLoginCompleted fires after a successful `cu login` of any kind.
	EventLoginCompleted EventName = "login_completed"
	// EventMCPSessionStarted fires when an MCP host initializes a session.
	EventMCPSessionStarted EventName = "mcp_session_started"
	// EventMCPSessionEnded fires when the MCP server stops.
	EventMCPSessionEnded EventName = "mcp_session_ended"

	// eventCommandPrefix + eventToolPrefix build per-command / per-tool event
	// names. Names (not properties) are what OpenPanel's Sankey chart uses as
	// path nodes, so "which tool" must live in the event name.
	eventCommandPrefix = "command_"
	eventToolPrefix    = "tool_"
)

// path renders the event as a pseudo URL path ("/tool/list_courses",
// "/command/fetch_theme"). It feeds OpenPanel's __path property, giving
// sessions meaningful entry/exit pages in the dashboard.
func (n EventName) path() string {
	s := string(n)
	switch {
	case strings.HasPrefix(s, eventToolPrefix):
		return "/tool/" + strings.TrimPrefix(s, eventToolPrefix)
	case strings.HasPrefix(s, eventCommandPrefix):
		return "/command/" + strings.TrimPrefix(s, eventCommandPrefix)
	default:
		return "/" + s
	}
}

// CommandEventName maps a cobra command path ("cu fetch theme") onto a typed
// event name ("command_fetch_theme").
func CommandEventName(commandPath string) EventName {
	slug := strings.TrimPrefix(commandPath, "cu")
	slug = strings.TrimSpace(slug)
	if slug == "" {
		slug = "root"
	}
	return EventName(eventCommandPrefix + strings.ReplaceAll(slug, " ", "_"))
}

// ToolCallEventName maps an MCP tool name onto a typed event name
// ("tool_list_courses").
func ToolCallEventName(tool string) EventName {
	return EventName(eventToolPrefix + tool)
}

// Property keys shared across events. Kept as constants so dashboards never
// chase renamed fields.
const (
	// propDeviceOverride is OpenPanel's caller-supplied device id
	// (properties.__deviceId). OpenPanel derives its sessions from the device,
	// so this controls how events group into paths: MCP events pass the MCP
	// session id (one OpenPanel session per agent session), CLI events pass
	// the persistent device id (30-min usage windows per machine).
	propDeviceOverride = "__deviceId"
	// propPath (properties.__path) becomes the event's "page"; sessions expose
	// it as entry/exit page and the Pages report groups by it.
	propPath = "__path"
	// propBrowserOverride / propBrowserVersionOverride replace the parsed
	// browser columns in the dashboard with the actual client.
	propBrowserOverride        = "__browser"
	propBrowserVersionOverride = "__browserVersion"

	propAppVersion = "app_version"
	propOS         = "os"
	propArch       = "arch"

	propCommand    = "command"
	propFlags      = "flags"
	propDurationMS = "duration_ms"
	propSuccess    = "success"
	propErrorKind  = "error_kind"

	propTarget = "target"

	propSessionID     = "session_id"
	propClientName    = "client_name"
	propClientVersion = "client_version"

	propTool        = "tool"
	propSeq         = "seq"
	propSincePrevMS = "since_prev_ms"
	propIsError     = "is_error"
	propArgKeys     = "arg_keys"
	propToolCalls   = "tool_calls"
)
