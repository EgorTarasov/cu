package command

import (
	"context"
	"fmt"
	"time"

	cu2 "github.com/EgorTarasov/cu/internal/gateway/cu"
	"github.com/EgorTarasov/cu/internal/gateway/ktalk"
	tcli "github.com/EgorTarasov/cu/internal/gateway/timeclient"

	"github.com/EgorTarasov/cu/internal/model"
	"github.com/EgorTarasov/cu/internal/settings"
	"github.com/EgorTarasov/cu/internal/telemetry"
	"github.com/EgorTarasov/cu/internal/usecase/login"

	"github.com/spf13/cobra"
)

const defaultLoginTimeout = 5 * time.Minute

var (
	loginTimeout time.Duration
	loginManual  bool
	loginBrowser bool
	loginSetup   bool
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Central University",
	Long: `Saves the LMS auth cookie so other commands can use it.

By default cu asks once how you want to log in:

  1. Manually  — you copy bff.cookie from a browser where you are already
                 signed in. Nothing extra to install.
  2. Browser   — cu drives Chrome and captures the cookie itself. Needs
                 Google Chrome; cu can download Chrome for Testing on demand.

The choice is stored in ~/.cu-cli/config.json and can be changed with
--setup, or overridden per run with --manual / --browser.

Use --gitlab to authenticate with git.culab.ru (GitLab).
Use --time to authenticate with time.cu.ru (chat / notifications).
Use --ktalk to authenticate with centraluniversity.ktalk.ru.

Examples:
  cuni login              # LMS authentication
  cuni login --manual     # paste the cookie yourself
  cuni login --browser    # drive Chrome this once
  cuni login --setup      # choose the default method again
  cuni login --gitlab     # GitLab authentication`,
	Run: func(cmd *cobra.Command, _ []string) {
		gitlabMode, _ := cmd.Flags().GetBool("gitlab")
		timeMode, _ := cmd.Flags().GetBool("time")
		ktalkMode, _ := cmd.Flags().GetBool("ktalk")
		ctx := cmd.Context()
		if ctx == nil {
			ctx = context.Background()
		}
		in := model.LoginInput{Timeout: loginTimeout}

		switch {
		case ktalkMode:
			runKtalkLogin(ctx, in)
		case timeMode:
			runTimeLogin(ctx, in)
		case gitlabMode:
			runGitLabLogin(ctx, in)
		default:
			runLMSLogin(ctx, in)
		}
	},
}

// runLMSLogin picks the configured method and stores a validated LMS cookie.
func runLMSLogin(ctx context.Context, in model.LoginInput) {
	if loginManual && loginBrowser {
		exitErrf("Укажите только один из флагов --manual / --browser.")
	}

	switch resolveLoginMethod(loginManual, loginBrowser, loginSetup) {
	case settings.LoginMethodBrowser:
		loginLMSViaBrowser(ctx, in)
	case settings.LoginMethodManual, settings.LoginMethodUnset:
		loginLMSManually(ctx)
	default:
		loginLMSManually(ctx)
	}

	telemetry.Default().LoginCompleted("lms")
	path, _ := cu2.CookieFilePath()
	fmt.Printf("Cookie сохранён в %s\n", path)
}

func loginLMSManually(ctx context.Context) {
	printManualInstructions()

	if !isInteractive() {
		fmt.Println("Терминал неинтерактивный — вставить cookie некуда.")
		fmt.Println("Передайте значение через переменную окружения:")
		fmt.Println()
		fmt.Println(`  export CU_BFF_COOKIE="значение-cookie"`)
		fmt.Println()
		exitErrf("Cookie не сохранён.")
	}

	cookie, err := readCookieFromTerminal()
	if err != nil {
		exitErrf("Не удалось прочитать cookie: %v", err)
	}

	fmt.Println("Проверяем cookie...")
	if err := cu2.NewClient(cookie).ValidateCookieWithContext(ctx); err != nil {
		fmt.Printf("Предупреждение: проверка не прошла: %v\n", err)
		fmt.Println("Сохраняем всё равно — возможно, LMS временно недоступна.")
	} else {
		fmt.Println("Cookie принят LMS.")
	}

	if err := cu2.SaveCookie(cookie); err != nil {
		exitErrf("Не удалось сохранить cookie: %v", err)
	}
}

func loginLMSViaBrowser(ctx context.Context, in model.LoginInput) {
	if err := ensureChrome(ctx); err != nil {
		exitErrf("%v", err)
	}

	fmt.Println("Открываем браузер для входа в LMS...")
	fmt.Println("Войдите в аккаунт в открывшемся окне.")

	uc := login.New(cu2.LoginWithBrowser, cu2.SaveCookie, nil)
	if _, err := uc.Execute(ctx, in); err != nil {
		exitErrf("Вход не удался: %v", err)
	}

	fmt.Println("Проверяем cookie...")
	if err := validateSavedCookie(ctx); err != nil {
		fmt.Printf("Предупреждение: проверка не прошла: %v\n", err)
	} else {
		fmt.Println("Cookie принят LMS.")
	}
}

func runGitLabLogin(ctx context.Context, in model.LoginInput) {
	if err := ensureChrome(ctx); err != nil {
		exitErrf("%v", err)
	}

	fmt.Println("Открываем браузер для входа в GitLab...")
	fmt.Println(`Нажмите "Центральный Университет" для входа через SSO.`)

	uc := login.New(cu2.LoginGitLabWithBrowser, cu2.SaveGitLabCookie, nil)
	if _, err := uc.Execute(ctx, in); err != nil {
		exitErrf("Вход в GitLab не удался: %v", err)
	}

	telemetry.Default().LoginCompleted("gitlab")
	path, _ := cu2.GitLabCookieFilePath()
	fmt.Printf("GitLab cookie сохранён в %s\n", path)
	fmt.Println("Теперь доступна команда 'cuni materials' для лонгридов из git.culab.ru.")
}

func runTimeLogin(ctx context.Context, in model.LoginInput) {
	if err := ensureChrome(ctx); err != nil {
		exitErrf("%v", err)
	}

	fmt.Println("Открываем браузер для входа в time.cu.ru...")
	fmt.Println("Войдите в аккаунт в открывшемся окне.")

	uc := login.New(tcli.LoginWithBrowser, tcli.SaveCookie, nil)
	if _, err := uc.Execute(ctx, in); err != nil {
		exitErrf("Вход в time.cu.ru не удался: %v", err)
	}

	telemetry.Default().LoginCompleted("time")
	path, _ := tcli.CookieFilePath()
	fmt.Printf("Токен time.cu.ru сохранён в %s\n", path)
	fmt.Println("Теперь доступна команда 'cuni time sync'.")
}

func runKtalkLogin(ctx context.Context, in model.LoginInput) {
	if err := ensureChrome(ctx); err != nil {
		exitErrf("%v", err)
	}

	fmt.Println("Открываем браузер для входа в Ktalk...")
	fmt.Println("Войдите в аккаунт в открывшемся окне.")
	fmt.Println("Снимок будет сделан, когда вы вернётесь на centraluniversity.ktalk.ru.")

	tokens, err := ktalk.LoginWithBrowser(ctx, in.Timeout)
	if err != nil {
		exitErrf("Вход в Ktalk не удался: %v", err)
	}
	if !tokens.Complete() {
		exitErrf("Вход в Ktalk: не удалось получить обязательную cookie %q", ktalk.CookieNgToken)
	}
	if err := ktalk.SaveTokens(tokens); err != nil {
		exitErrf("Не удалось сохранить токены Ktalk: %v", err)
	}

	telemetry.Default().LoginCompleted("ktalk")
	path, _ := ktalk.TokensFilePath()
	fmt.Printf("Токены Ktalk сохранены в %s\n", path)
	fmt.Printf("  cookies: %d, localStorage: %d, sessionStorage: %d\n",
		len(tokens.Cookies), len(tokens.LocalStorage), len(tokens.SessionStorage))
}

func init() {
	loginCmd.Flags().DurationVar(&loginTimeout, "timeout", defaultLoginTimeout, "Login timeout")
	loginCmd.Flags().BoolVar(&loginManual, "manual", false, "Paste the cookie yourself (no Chrome needed)")
	loginCmd.Flags().BoolVar(&loginBrowser, "browser", false, "Drive Chrome to capture the cookie")
	loginCmd.Flags().BoolVar(&loginSetup, "setup", false, "Choose the default login method again")
	loginCmd.Flags().Bool("gitlab", false, "Authenticate with git.culab.ru (GitLab)")
	loginCmd.Flags().Bool("time", false, "Authenticate with time.cu.ru")
	loginCmd.Flags().Bool("ktalk", false, "Authenticate with centraluniversity.ktalk.ru (Ktalk)")
}
