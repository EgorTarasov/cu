package command

import (
	"fmt"
	"os"
)

// exitErrf prints a formatted error to stderr and exits with code 1.
func exitErrf(format string, args ...any) {
	reportCommandFailure("runtime")
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func cookieRequiredError(err error) {
	reportCommandFailure("auth")
	fmt.Fprintln(os.Stderr, "No authentication found.")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Option 1 — login via browser:")
	fmt.Fprintln(os.Stderr, "  cu login")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Option 2 — set cookie manually:")
	fmt.Fprintln(os.Stderr, "  export CU_BFF_COOKIE='your-cookie-value-here'")
	fmt.Fprintln(os.Stderr, "  cu fetch course 519")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "Error details: %v\n", err)
	os.Exit(1)
}
