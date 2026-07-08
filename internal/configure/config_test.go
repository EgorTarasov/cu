package configure

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseEnv(t *testing.T) {
	raw := `
# comment
OPDASHBOARD_CLIENT_ID="abc-123"
OPDASHBOARD_SECRET=sec_plain
OP_API_URL = "https://example.com"

BROKEN LINE
`
	values := parseEnv(raw)

	assert.Equal(t, "abc-123", values["OPDASHBOARD_CLIENT_ID"])
	assert.Equal(t, "sec_plain", values["OPDASHBOARD_SECRET"])
	assert.Equal(t, "https://example.com", values["OP_API_URL"])
	assert.NotContains(t, values, "BROKEN LINE")
}

func TestLoad_EnvOverride(t *testing.T) {
	t.Setenv("OP_API_URL", "https://override.example.com")

	cfg := Load()

	assert.Equal(t, "https://override.example.com", cfg.ClickstreamAPIURL)
}
