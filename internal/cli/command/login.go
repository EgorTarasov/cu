package command

import (
	"context"
	"fmt"
	"time"

	cu2 "github.com/EgorTarasov/cu/internal/gateway/cu"
	"github.com/EgorTarasov/cu/internal/gateway/ktalk"
	tcli "github.com/EgorTarasov/cu/internal/gateway/timeclient"

	"github.com/EgorTarasov/cu/internal/model"
	"github.com/EgorTarasov/cu/internal/usecase/login"

	"github.com/spf13/cobra"
)

const defaultLoginTimeout = 5 * time.Minute

var loginTimeout time.Duration

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Central University via browser",
	Long: `Opens Chrome browser for Keycloak login, captures auth cookie automatically.

Use --gitlab to authenticate with git.culab.ru (GitLab).
Use --time to authenticate with time.cu.ru (chat / notifications).
Use --ktalk to authenticate with centraluniversity.ktalk.ru.

Examples:
  cu login              # LMS authentication
  cu login --gitlab     # GitLab authentication
  cu login --time       # time.cu.ru authentication
  cu login --ktalk      # Ktalk authentication`,
	Run: func(cmd *cobra.Command, _ []string) {
		gitlabMode, _ := cmd.Flags().GetBool("gitlab")
		timeMode, _ := cmd.Flags().GetBool("time")
		ktalkMode, _ := cmd.Flags().GetBool("ktalk")
		ctx := context.Background()
		in := model.LoginInput{Timeout: loginTimeout}

		switch {
		case ktalkMode:
			fmt.Println("Opening browser for Ktalk authentication...")
			fmt.Println("Please log in via the browser window.")
			fmt.Println("Snapshot fires once you land back on centraluniversity.ktalk.ru.")

			tokens, err := ktalk.LoginWithBrowser(ctx, in.Timeout)
			if err != nil {
				exitErrf("Ktalk login failed: %v", err)
			}
			if !tokens.Complete() {
				exitErrf("Ktalk login: required cookie %q not captured", ktalk.CookieNgToken)
			}
			if err := ktalk.SaveTokens(tokens); err != nil {
				exitErrf("Failed to save Ktalk tokens: %v", err)
			}
			path, _ := ktalk.TokensFilePath()
			fmt.Printf("Ktalk tokens saved to %s\n", path)
			fmt.Printf("  cookies: %d, localStorage: %d, sessionStorage: %d\n",
				len(tokens.Cookies), len(tokens.LocalStorage), len(tokens.SessionStorage))
			return
		case timeMode:
			fmt.Println("Opening browser for time.cu.ru authentication...")
			fmt.Println("Please log in via the browser window.")

			uc := login.New(tcli.LoginWithBrowser, tcli.SaveCookie, nil)
			if _, err := uc.Execute(ctx, in); err != nil {
				exitErrf("time.cu.ru login failed: %v", err)
			}

			path, _ := tcli.CookieFilePath()
			fmt.Printf("time.cu.ru token saved to %s\n", path)
			fmt.Println("You can now use 'cu time sync' to fetch posts from the notification bot.")
			return
		}

		if gitlabMode {
			fmt.Println("Opening browser for GitLab authentication...")
			fmt.Println("Click \"Центральный Университет\" to sign in via SSO.")

			uc := login.New(cu2.LoginGitLabWithBrowser, cu2.SaveGitLabCookie, nil)
			if _, err := uc.Execute(ctx, in); err != nil {
				fmt.Printf("GitLab login failed: %v\n", err)
				return
			}

			path, _ := cu2.GitLabCookieFilePath()
			fmt.Printf("GitLab cookie saved to %s\n", path)
			fmt.Println("You can now use 'cu materials' to download longreads from git.culab.ru.")
		} else {
			fmt.Println("Opening browser for LMS authentication...")
			fmt.Println("Please log in via the browser window.")

			uc := login.New(cu2.LoginWithBrowser, cu2.SaveCookie, nil)
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

			path, _ := cu2.CookieFilePath()
			fmt.Printf("Cookie saved to %s\n", path)
		}
	},
}

func init() {
	loginCmd.Flags().DurationVar(&loginTimeout, "timeout", defaultLoginTimeout, "Login timeout")
	loginCmd.Flags().Bool("gitlab", false, "Authenticate with git.culab.ru (GitLab)")
	loginCmd.Flags().Bool("time", false, "Authenticate with time.cu.ru")
	loginCmd.Flags().Bool("ktalk", false, "Authenticate with centraluniversity.ktalk.ru (Ktalk)")
}
