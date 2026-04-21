package cli

import (
	"context"
	"fmt"
	"time"

	"cu-sync/internal/cu"

	"github.com/spf13/cobra"
)

const defaultLoginTimeout = 5 * time.Minute

var loginTimeout time.Duration

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Central University via browser",
	Long: `Opens Chrome browser for Keycloak login, captures auth cookie automatically.

Use --gitlab to authenticate with git.culab.ru (GitLab) instead.

Examples:
  cu login            # LMS authentication
  cu login --gitlab   # GitLab authentication`,
	Run: func(cmd *cobra.Command, _ []string) {
		gitlabMode, _ := cmd.Flags().GetBool("gitlab")

		if gitlabMode {
			loginGitLab(cmd)
		} else {
			loginLMS(cmd)
		}
	},
}

func init() {
	loginCmd.Flags().DurationVar(&loginTimeout, "timeout", defaultLoginTimeout, "Login timeout")
	loginCmd.Flags().Bool("gitlab", false, "Authenticate with git.culab.ru (GitLab)")
}

func loginLMS(_ *cobra.Command) {
	fmt.Println("Opening browser for LMS authentication...")
	fmt.Println("Please log in via the browser window.")

	ctx := context.Background()
	cookie, err := cu.LoginWithBrowser(ctx, loginTimeout)
	if err != nil {
		fmt.Printf("Login failed: %v\n", err)
		return
	}

	client := cu.NewClient(cookie)
	if err := client.ValidateCookie(); err != nil {
		fmt.Printf("Warning: cookie validation failed: %v\n", err)
		fmt.Println("Saving cookie anyway.")
	} else {
		fmt.Println("Cookie validated successfully.")
	}

	if err := cu.SaveCookie(cookie); err != nil {
		fmt.Printf("Failed to save cookie: %v\n", err)
		fmt.Printf("Set manually: export CU_BFF_COOKIE='%s'\n", cookie)
		return
	}

	path, _ := cu.CookieFilePath()
	fmt.Printf("Cookie saved to %s\n", path)
}

func loginGitLab(_ *cobra.Command) {
	fmt.Println("Opening browser for GitLab authentication...")
	fmt.Println("Click \"Центральный Университет\" to sign in via SSO.")

	ctx := context.Background()
	cookie, err := cu.LoginGitLabWithBrowser(ctx, loginTimeout)
	if err != nil {
		fmt.Printf("GitLab login failed: %v\n", err)
		return
	}

	if err := cu.SaveGitLabCookie(cookie); err != nil {
		fmt.Printf("Failed to save GitLab cookie: %v\n", err)
		return
	}

	path, _ := cu.GitLabCookieFilePath()
	fmt.Printf("GitLab cookie saved to %s\n", path)
	fmt.Println("You can now use 'cu materials' to download longreads from git.culab.ru.")
}
