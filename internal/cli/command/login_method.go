package command

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/EgorTarasov/cu/internal/browser"
	cu2 "github.com/EgorTarasov/cu/internal/gateway/cu"
	"github.com/EgorTarasov/cu/internal/settings"
)

const cookieName = "bff.cookie"

// stdinFd returns os.Stdin's descriptor as an int. The bounds check makes the
// uintptr -> int conversion provably safe for gosec G115; a descriptor that
// cannot fit is reported as -1, which every caller treats as "not a terminal".
func stdinFd() int {
	fd := os.Stdin.Fd()
	if fd > uintptr(math.MaxInt) {
		return -1
	}
	return int(fd)
}

func isInteractive() bool {
	return term.IsTerminal(stdinFd())
}

// resolveLoginMethod applies explicit flags first, then the saved preference,
// and only prompts on a real terminal when nothing has been chosen yet.
// One-off flags are deliberately not persisted.
func resolveLoginMethod(forceManual, forceBrowser, rerunSetup bool) settings.LoginMethod {
	switch {
	case forceManual:
		return settings.LoginMethodManual
	case forceBrowser:
		return settings.LoginMethodBrowser
	}

	saved, err := settings.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Не удалось прочитать настройки: %v\n", err)
	}

	if saved.LoginMethod != settings.LoginMethodUnset && !rerunSetup {
		return saved.LoginMethod
	}

	if !isInteractive() {
		return settings.LoginMethodManual
	}

	chosen := promptLoginMethod()
	if err := settings.Save(settings.Settings{LoginMethod: chosen}); err != nil {
		fmt.Fprintf(os.Stderr, "Не удалось сохранить выбор: %v\n", err)
	} else {
		path, _ := settings.FilePath()
		fmt.Printf("Выбор сохранён в %s (изменить: cu login --setup)\n\n", path)
	}
	return chosen
}

func promptLoginMethod() settings.LoginMethod {
	fmt.Println("Как вы хотите получать cookie доступа к LMS?")
	fmt.Println()
	fmt.Println("  1) Вручную (по умолчанию)")
	fmt.Println("     Скопировать bff.cookie из браузера, где вы уже вошли.")
	fmt.Println("     Ничего дополнительно ставить не нужно.")
	fmt.Println()
	fmt.Println("  2) Автоматически через браузер")
	fmt.Println("     cu откроет окно Chrome и заберёт cookie сам.")
	fmt.Println("     Требуется Google Chrome. Если его нет — cu предложит")
	fmt.Println("     скачать Chrome for Testing (~170 МБ) в ~/.cu-cli/chrome.")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Выбор [1]: ")
		line, err := reader.ReadString('\n')
		if err != nil && strings.TrimSpace(line) == "" {
			fmt.Println()
			return settings.LoginMethodManual
		}

		switch strings.TrimSpace(line) {
		case "", "1":
			fmt.Println()
			return settings.LoginMethodManual
		case "2":
			fmt.Println()
			return settings.LoginMethodBrowser
		default:
			fmt.Println("Введите 1 или 2.")
		}
	}
}

// printManualInstructions explains where bff.cookie lives in a browser.
func printManualInstructions() {
	fmt.Println("Как получить bff.cookie вручную:")
	fmt.Println()
	fmt.Println("  1. Откройте https://my.centraluniversity.ru и войдите в аккаунт.")
	fmt.Println("  2. Откройте DevTools:")
	fmt.Println("       macOS          — Cmd+Option+I")
	fmt.Println("       Windows/Linux  — F12 или Ctrl+Shift+I")
	fmt.Println("  3. Перейдите на вкладку Application (Chrome/Edge)")
	fmt.Println("     или Storage (Firefox).")
	fmt.Println("  4. Слева выберите Cookies → https://my.centraluniversity.ru")
	fmt.Println("  5. Найдите строку bff.cookie и скопируйте её значение (Value).")
	fmt.Println()
	fmt.Println("Значение длинное — копируйте целиком, без кавычек.")
	fmt.Println()
}

// readCookieFromTerminal prompts for the cookie with echo disabled so the
// session value does not linger in scrollback.
func readCookieFromTerminal() (string, error) {
	fmt.Print("Вставьте значение bff.cookie (ввод скрыт): ")
	raw, err := term.ReadPassword(stdinFd())
	fmt.Println()
	if err != nil {
		return "", fmt.Errorf("не удалось прочитать ввод: %w", err)
	}

	cookie := strings.TrimSpace(string(raw))
	cookie = strings.Trim(cookie, `"'`)
	if cookie == "" {
		return "", errors.New("пустое значение")
	}
	if strings.HasPrefix(cookie, cookieName+"=") {
		cookie = strings.TrimPrefix(cookie, cookieName+"=")
	}
	return cookie, nil
}

// ensureChrome resolves a Chrome binary, offering the opt-in download when the
// host has none.
func ensureChrome(ctx context.Context) error {
	if path := browser.Resolve(); path != "" {
		return nil
	}

	if !isInteractive() {
		return errors.New("не найден Google Chrome. Установите Chrome, задайте CHROME_PATH " +
			"или используйте ручной способ: cu login --manual")
	}

	fmt.Println("Google Chrome не найден.")
	fmt.Println()
	fmt.Println("cu может скачать Chrome for Testing (~170 МБ) в ~/.cu-cli/chrome.")
	fmt.Println("Он используется только для входа и не трогает ваш обычный браузер.")
	fmt.Println()
	fmt.Print("Скачать? [y/N]: ")

	line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes", "д", "да":
	default:
		return errors.New("нужен Chrome. Установите его, задайте CHROME_PATH " +
			"или используйте: cu login --manual")
	}

	path, err := browser.Download(ctx, os.Stdout)
	if err != nil {
		return fmt.Errorf("не удалось скачать Chrome: %w", err)
	}

	fmt.Printf("Chrome готов: %s\n\n", path)
	return nil
}

// validateSavedCookie checks the cookie that was just persisted, so the
// success message reflects a real API response.
func validateSavedCookie(ctx context.Context) error {
	cookie, err := cu2.LoadCookie()
	if err != nil {
		return err
	}
	if cookie == "" {
		return errors.New("cookie не сохранился")
	}
	return cu2.NewClient(cookie).ValidateCookieWithContext(ctx)
}
