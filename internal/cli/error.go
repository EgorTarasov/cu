package cli

import (
	"fmt"
)

func cookieRequiredError(err error) {
	fmt.Println("No authentication found.")
	fmt.Println()
	fmt.Println("Option 1 — login via browser:")
	fmt.Println("  cu login")
	fmt.Println()
	fmt.Println("Option 2 — set cookie manually:")
	fmt.Println("  export CU_BFF_COOKIE='your-cookie-value-here'")
	fmt.Println("  cu fetch course 519")
	fmt.Println()
	fmt.Printf("Error details: %v\n", err)
	panic("authentication required")
}
