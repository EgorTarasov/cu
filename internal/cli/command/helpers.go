package command

import (
	"cu-sync/internal/gateway/cu"
	"fmt"
	"os"
)

// mustClient creates an authenticated client or exits.
func mustClient() *cu.Client {
	client, err := cu.NewClientFromEnv()
	if err != nil {
		cookieRequiredError(err)
	}
	if err = client.ValidateCookie(); err != nil {
		fmt.Fprintf(os.Stderr, "Cookie expired: %v\nRun: cu login\n", err)
		os.Exit(1)
	}
	return client
}
