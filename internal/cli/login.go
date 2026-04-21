package cli

import (
	"context"
	"fmt"
	"time"

	"cu-sync/internal/cu"
	"cu-sync/internal/usecase/login"
	"cu-sync/internal/usecase/login/model/input"

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
		ctx := context.Background()
		in := input.LoginInput{Timeout: loginTimeout}

		if gitlabMode {
			fmt.Println("Opening browser for GitLab authentication...")
			fmt.Println("Click \"Центральный Университет\" to sign in via SSO.")

			uc := login.New(cu.LoginGitLabWithBrowser, cu.SaveGitLabCookie, nil)
			if _, err := uc.Execute(ctx, in); err != nil {
				fmt.Printf("GitLab login failed: %v\n", err)
				return
			}

			path, _ := cu.GitLabCookieFilePath()
			fmt.Printf("GitLab cookie saved to %s\n", path)
			fmt.Println("You can now use 'cu materials' to download longreads from git.culab.ru.")
		} else {
			fmt.Println("Opening browser for LMS authentication...")
			fmt.Println("Please log in via the browser window.")

			uc := login.New(cu.LoginWithBrowser, cu.SaveCookie, nil)
			result, err := uc.Execute(ctx, in)
			if err != nil {
				fmt.Printf("Login failed: %v\n", err)
				return
			}

			if result.ValidationError != nil {
				fmt.Printf("Warning: cookie validation failed: %v\n", result.ValidationError)
				fmt.Println("Saving cookie anyway.")
			} else {
				fmt.Println("Cookie validated successfully.")
			}

			path, _ := cu.CookieFilePath()
			fmt.Printf("Cookie saved to %s\n", path)
		}
	},
}

func init() {
	loginCmd.Flags().DurationVar(&loginTimeout, "timeout", defaultLoginTimeout, "Login timeout")
	loginCmd.Flags().Bool("gitlab", false, "Authenticate with git.culab.ru (GitLab)")
}
