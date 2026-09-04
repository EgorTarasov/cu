package telemetry

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommandEventName_IgnoresBinaryName(t *testing.T) {
	// The binary was renamed cu -> cuni; event names must not move with it.
	for _, bin := range []string{"cu", "cuni", "anything"} {
		require.Equal(t, EventName("command_fetch_theme"),
			CommandEventName(bin+" fetch theme"), "binary %q", bin)
		require.Equal(t, EventName("command_root"), CommandEventName(bin))
	}
}
