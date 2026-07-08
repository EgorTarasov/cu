// Package configure holds build-time configuration embedded into the binary.
// `make embed-env` (or the release CI) drops an .env file into embedded/
// before the build; without it the binary compiles fine and Load returns an
// empty Config, which switches telemetry into no-op mode. At runtime a
// same-named OS environment variable overrides the embedded value.
package configure

import (
	"embed"
	"os"
	"strings"
)

// The all: prefix keeps the pattern valid even when embedded/ holds only
// .gitkeep — embedded/.env itself is gitignored and generated at build time.
//
//go:embed all:embedded
var embeddedFS embed.FS

// Config carries clickstream (OpenPanel) credentials and endpoint. Zero
// values mean "telemetry disabled".
type Config struct {
	ClickstreamClientID string
	ClickstreamSecret   string
	ClickstreamAPIURL   string
}

// Load parses the embedded .env (if it was present at build time) and
// applies OS environment overrides.
func Load() Config {
	values := map[string]string{}
	if data, err := embeddedFS.ReadFile("embedded/.env"); err == nil {
		values = parseEnv(string(data))
	}
	return Config{
		ClickstreamClientID: getValue(values, "OPDASHBOARD_CLIENT_ID"),
		ClickstreamSecret:   getValue(values, "OPDASHBOARD_SECRET"),
		ClickstreamAPIURL:   getValue(values, "OP_API_URL"),
	}
}

func getValue(values map[string]string, key string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return values[key]
}

func parseEnv(raw string) map[string]string {
	values := make(map[string]string)
	for line := range strings.Lines(raw) {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		values[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(value), `"`)
	}
	return values
}
