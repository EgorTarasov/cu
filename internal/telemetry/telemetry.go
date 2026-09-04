// Package telemetry is the single entry point for product analytics.
// CLI and MCP code call its measurement methods (CommandExecuted,
// MCPToolCalled, ...) and never see the underlying clickstream client or
// event wire format. Events are sent in background goroutines; call Flush
// before the process exits.
//
// Telemetry is disabled (all methods become no-ops) when CU_NO_TELEMETRY is
// set or the embedded credentials are missing.
package telemetry

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/EgorTarasov/cu/internal/configure"
	"github.com/EgorTarasov/cu/internal/gateway/clickstream"
)

const (
	sendTimeout         = 5 * time.Second
	DefaultFlushTimeout = 2 * time.Second
)

// sender is the slice of the clickstream gateway the tracker needs.
type sender interface {
	Track(ctx context.Context, payload clickstream.TrackPayload) (*clickstream.TrackResponse, error)
	Identify(ctx context.Context, payload clickstream.IdentifyPayload) (*clickstream.TrackResponse, error)
}

// Tracker sends analytics events. The zero value is unusable; a nil *Tracker
// is a valid no-op, so callers never guard their calls.
type Tracker struct {
	sender   sender
	deviceID string
	version  string
	spool    string
	wg       sync.WaitGroup
}

var defaultTracker *Tracker

// Init builds the process-wide tracker from embedded configuration and fires
// the first-run events if the device ID was just created. Safe to skip:
// without Init every telemetry call is a no-op.
func Init(appVersion string) {
	if envDisabled() {
		return
	}
	cfg := configure.Load()
	if cfg.ClickstreamAPIURL == "" || cfg.ClickstreamClientID == "" ||
		cfg.ClickstreamSecret == "" {
		return
	}
	deviceID, created, err := loadOrCreateDeviceID()
	if err != nil {
		return
	}

	client := clickstream.NewClient(
		cfg.ClickstreamAPIURL,
		cfg.ClickstreamClientID,
		cfg.ClickstreamSecret,
		appVersion,
	)
	client.SetUserAgent(userAgent(appVersion))

	queue, err := spoolPath()
	if err != nil {
		queue = ""
	}

	defaultTracker = &Tracker{
		sender:   client,
		deviceID: deviceID,
		version:  appVersion,
		spool:    queue,
	}

	// Ship whatever the previous run queued. Deliberately a bare goroutine
	// and not async(): it must stay outside the WaitGroup so Flush returns
	// without waiting for it, letting the send overlap with this run's real
	// work instead of delaying the exit.
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), sendTimeout)
		defer cancel()
		defaultTracker.drainSpool(ctx)
	}()

	if created {
		defaultTracker.firstRun()
	}
}

func envDisabled() bool {
	return os.Getenv("CU_NO_TELEMETRY") != ""
}

// userAgent builds a UA that OpenPanel's parser recognizes as a real device.
// Plain "cu-cli/1.0" would be classified as a server SDK (isServer), and
// server events never open sessions — killing the Sessions page and Sankey
// (user paths) reports. The OS token makes the parser resolve an OS, which
// is enough to count as a client.
func userAgent(version string) string {
	// Named "platform" rather than "token": gosec G101 matches credential-ish
	// identifiers, and a //nolint directive does not reach the standalone
	// gosec run in CI (that one only honours #nosec).
	var osPlatform string
	switch runtime.GOOS {
	case "darwin":
		osPlatform = "Macintosh; Intel Mac OS X 10_15_7"
	case "windows":
		osPlatform = "Windows NT 10.0; Win64; x64"
	default:
		osPlatform = "X11; Linux x86_64"
	}
	return fmt.Sprintf("cu-cli/%s (%s)", version, osPlatform)
}

// Default returns the tracker built by Init, or nil (a valid no-op tracker).
func Default() *Tracker {
	return defaultTracker
}

type contextKey struct{}

// WithContext stores the tracker in ctx for handlers that prefer DI over the
// package default.
func WithContext(ctx context.Context, t *Tracker) context.Context {
	return context.WithValue(ctx, contextKey{}, t)
}

// FromContext returns the tracker stored by WithContext, falling back to
// Default().
func FromContext(ctx context.Context) *Tracker {
	if t, ok := ctx.Value(contextKey{}).(*Tracker); ok {
		return t
	}
	return Default()
}

// CommandEvent describes one finished CLI command invocation.
type CommandEvent struct {
	Command   string        // full cobra path, e.g. "cuni fetch theme"
	Flags     []string      // names of flags explicitly set, no values
	Duration  time.Duration //
	Success   bool          //
	ErrorKind string        // "auth" | "runtime" | "usage"; empty when Success
}

// CommandExecuted records a finished CLI command.
func (t *Tracker) CommandExecuted(e CommandEvent) {
	if t == nil {
		return
	}
	props := map[string]any{
		propCommand:    e.Command,
		propDurationMS: e.Duration.Milliseconds(),
		propSuccess:    e.Success,
	}
	if len(e.Flags) > 0 {
		props[propFlags] = e.Flags
	}
	if e.ErrorKind != "" {
		props[propErrorKind] = e.ErrorKind
	}
	t.trackSpooled(CommandEventName(e.Command), props)
}

// LoginCompleted records a successful login; target is the auth backend:
// "lms", "gitlab", "time" or "ktalk".
func (t *Tracker) LoginCompleted(target string) {
	if t == nil {
		return
	}
	t.trackSpooled(EventLoginCompleted, map[string]any{propTarget: target})
}

// NewSessionID returns a random ID that groups all events of one MCP session.
func NewSessionID() string {
	return newRandomID()
}

// MCPSessionStarted records an MCP initialize handshake; clientName and
// clientVersion identify the LLM host (e.g. "claude-code").
func (t *Tracker) MCPSessionStarted(sessionID, clientName, clientVersion string) {
	if t == nil {
		return
	}
	t.track(EventMCPSessionStarted, sessionID, map[string]any{
		propSessionID:     sessionID,
		propClientName:    clientName,
		propClientVersion: clientVersion,
	})
}

// ToolCallEvent describes one finished MCP tools/call.
type ToolCallEvent struct {
	SessionID  string        // ID from NewSessionID, shared by all calls of the session
	ClientName string        // LLM host from the initialize handshake
	Tool       string        // tool name, e.g. "list_courses"
	Seq        int           // 1-based index of the call within the session
	SincePrev  time.Duration // time since the previous tool call; 0 for the first
	Duration   time.Duration //
	IsError    bool          //
	ArgKeys    []string      // names of arguments the host passed, no values
}

// MCPToolCalled records a finished MCP tool call.
func (t *Tracker) MCPToolCalled(e ToolCallEvent) {
	if t == nil {
		return
	}
	props := map[string]any{
		propSessionID:   e.SessionID,
		propClientName:  e.ClientName,
		propTool:        e.Tool,
		propSeq:         e.Seq,
		propSincePrevMS: e.SincePrev.Milliseconds(),
		propDurationMS:  e.Duration.Milliseconds(),
		propIsError:     e.IsError,
	}
	if len(e.ArgKeys) > 0 {
		props[propArgKeys] = e.ArgKeys
	}
	t.track(ToolCallEventName(e.Tool), e.SessionID, props)
}

// MCPSessionEnded records the end of an MCP session.
func (t *Tracker) MCPSessionEnded(sessionID string, duration time.Duration, toolCalls int) {
	if t == nil {
		return
	}
	t.track(EventMCPSessionEnded, sessionID, map[string]any{
		propSessionID:  sessionID,
		propDurationMS: duration.Milliseconds(),
		propToolCalls:  toolCalls,
	})
}

// Flush waits until in-flight events are delivered, at most timeout.
func (t *Tracker) Flush(timeout time.Duration) {
	if t == nil {
		return
	}
	done := make(chan struct{})
	go func() {
		t.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(timeout):
	}
}

// firstRun identifies the fresh device and emits the install event.
func (t *Tracker) firstRun() {
	t.async(func(ctx context.Context) {
		_, _ = t.sender.Identify(ctx, clickstream.IdentifyPayload{
			ProfileID: t.deviceID,
			Properties: map[string]any{
				propAppVersion: t.version,
				propOS:         runtime.GOOS,
				propArch:       runtime.GOARCH,
			},
		})
	})
	t.track(EventFirstRun, "", map[string]any{})
}

// track enriches props with build/platform info and sends in background.
// deviceOverride (properties.__deviceId) tells OpenPanel which "device" — and
// therefore which session — the event belongs to; empty means the persistent
// device id.
func (t *Tracker) track(name EventName, deviceOverride string, props map[string]any) {
	payload := t.payload(name, deviceOverride, props)
	t.async(func(ctx context.Context) {
		_, _ = t.sender.Track(ctx, payload)
	})
}

// trackSpooled queues the event on disk instead of sending it now. CLI
// commands use it because the process exits immediately afterwards: a
// synchronous POST costs its own TLS handshake — measured at ~500ms, roughly
// a third of a typical invocation — for data nobody is waiting on. The next
// run ships it, concurrently with that run's work.
func (t *Tracker) trackSpooled(name EventName, props map[string]any) {
	payload := t.payload(name, "", props)
	if t.spool == "" {
		t.async(func(ctx context.Context) {
			_, _ = t.sender.Track(ctx, payload)
		})
		return
	}
	if err := appendSpool(t.spool, payload); err != nil {
		t.async(func(ctx context.Context) {
			_, _ = t.sender.Track(ctx, payload)
		})
	}
}

// payload stamps the device/app context every event carries.
func (t *Tracker) payload(
	name EventName,
	deviceOverride string,
	props map[string]any,
) clickstream.TrackPayload {
	if deviceOverride == "" {
		deviceOverride = t.deviceID
	}
	props[propDeviceOverride] = deviceOverride
	props[propPath] = name.path()
	props[propBrowserOverride] = "cu-cli"
	props[propBrowserVersionOverride] = t.version
	props[propAppVersion] = t.version
	props[propOS] = runtime.GOOS
	props[propArch] = runtime.GOARCH
	return clickstream.TrackPayload{
		Name:       string(name),
		ProfileID:  t.deviceID,
		Properties: props,
	}
}

func (t *Tracker) async(fn func(ctx context.Context)) {
	t.wg.Go(func() {
		ctx, cancel := context.WithTimeout(context.Background(), sendTimeout)
		defer cancel()
		fn(ctx)
	})
}
