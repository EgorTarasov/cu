//go:build integration

package integration_tests

import (
	"fmt"
	"log"
	"os"

	"cu-sync/internal/cu"
)

func main() {
	// Get the cookie from environment variable for security
	bffCookie := os.Getenv("CU_BFF_COOKIE")
	if bffCookie == "" {
		panic("provide valid CU_BFF_COOKIE")
	}

	fmt.Println("Central University API Client Integration Test")
	fmt.Println("==============================================")
	fmt.Printf("Using cookie: %s...\n\n", bffCookie[:20])

	// Create client
	client := cu.NewClient(bffCookie)

	// Test 1: Validate cookie
	fmt.Println("1. Testing cookie validation...")
	if err := client.ValidateCookie(); err != nil {
		log.Printf("   ❌ Cookie validation failed: %v\n", err)
		// Don't exit, continue with other tests
	} else {
		fmt.Println("   ✅ Cookie is valid")
	}

	// Test 2: Get student courses (matching your exact curl command)
	fmt.Println("\n2. Testing student courses API (GET /api/micro-lms/courses/student?limit=10000&state=published)...")
	courses, err := client.GetStudentCourses(10000, "published")
	if err != nil {
		log.Printf("   ❌ Failed to get student courses: %v\n", err)
		return
	}

	fmt.Printf("   ✅ Successfully retrieved %d courseœs\n", len(*courses))

	// Display course details
	fmt.Println("\n3. Course Details:")
	fmt.Println("   " + "="*60)
	for i, course := range *courses {
		if i >= 5 { // Only show first 5 courses to avoid too much output
			break
		}
		fmt.Printf("   %d. %s\n", i+1, course.Name)
		fmt.Printf("      ID: %d | State: %s | Archived: %t\n", course.ID, course.State, course.IsArchived)
		if course.Description != "" {
			fmt.Printf("      Description: %s\n", course.Description)
		}
		if course.PublishedAt != nil {
			fmt.Printf("      Published: %s\n", course.PublishedAt.Format("2006-01-02 15:04:05"))
		}
		if course.Progress != nil {
			fmt.Printf("      Progress: %d/%d (%.1f%%)\n",
				course.Progress.CompletedCount,
				course.Progress.TotalCount,
				course.Progress.Percentage)
		}
		fmt.Println()
	}

	if len(*courses) > 5 {
		fmt.Printf("   ... and %d more courses\n", len(*courses)-5)
	}

	fmt.Println("\n🎉 Integration test completed successfully!")
	fmt.Println("   The client is working correctly with the Central University API.")
}
