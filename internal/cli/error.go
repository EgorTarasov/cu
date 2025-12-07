package cli

import (
	"fmt"
	"log"
)

func cookieRequiredError(err error) {
	fmt.Println("⚠️  No CU_BFF_COOKIE environment variable found.")
	fmt.Println("Please set the CU_BFF_COOKIE environment variable with your bff.cookie value:")
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println("  export CU_BFF_COOKIE='your-cookie-value-here'")
	fmt.Println("  cu fetch course 519")
	fmt.Printf("Error details: %v\n", err)
	fmt.Println()
	log.Fatal("CU_BFF_COOKIE environment variable is required")
}
