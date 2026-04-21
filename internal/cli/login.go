package cli

import (
	"context"
	"fmt"
	"time"

	"cu-sync/internal/cu"

	"github.com/spf13/cobra"
)

var loginTimeout time.Duration

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Central University via browser",
	Long:  "Opens Chrome browser for Keycloak login, captures auth cookie automatically.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Opening browser for authentication...")
		fmt.Println("Please log in via the browser window that opens.")
		fmt.Println("The browser will close automatically after successful login.")
		fmt.Println()

		ctx := context.Background()
		cookie, err := cu.LoginWithBrowser(ctx, loginTimeout)
		if err != nil {
			fmt.Printf("Login failed: %v\n", err)
			return
		}

		client := cu.NewClient(cookie)
		if err := client.ValidateCookie(); err != nil {
			fmt.Printf("Warning: cookie validation failed: %v\n", err)
			fmt.Println("Saving cookie anyway — it may work for some endpoints.")
		} else {
			fmt.Println("Cookie validated successfully.")
		}

		if err := cu.SaveCookie(cookie); err != nil {
			fmt.Printf("Failed to save cookie: %v\n", err)
			fmt.Println("You can set it manually:")
			fmt.Printf("  export CU_BFF_COOKIE='%s'\n", cookie)
			return
		}

		path, _ := cu.CookieFilePath()
		fmt.Printf("Cookie saved to %s\n", path)
		fmt.Println("You can now use 'cu fetch courses' and other commands.")
	},
}

func init() {
	loginCmd.Flags().DurationVar(&loginTimeout, "timeout", 5*time.Minute, "Login timeout")
}
