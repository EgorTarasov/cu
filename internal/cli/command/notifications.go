package command

import (
	"fmt"

	tcli "github.com/EgorTarasov/cu/internal/gateway/timeclient"
	"github.com/EgorTarasov/cu/internal/usecase/notifications"

	"github.com/spf13/cobra"
)

var (
	notifsCourse   int
	notifsLongread int
	notifsLimit    int
)

const defaultNotifsLimit = 50

var mmNotificationsCmd = &cobra.Command{
	Use:     "notifications",
	Aliases: []string{"notifs"},
	Short:   "List LMS notifications collected from the bot channel",
	Long: `Scans local time.cu.ru storage for "Новые задачи" / "Задача оценена" / "Записи занятий"
posts and prints them with the LMS link and a permalink back into the chat.

Filter:
  --course <id>     keep only posts referencing this course
  --longread <id>   keep only posts referencing this longread`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		cfg := tcli.LoadConfig()

		client, err := tcli.NewClientFromEnv()
		if err != nil {
			exitErrf("%v", err)
		}
		_, _, ch, err := client.ResolveBotChannel(ctx, cfg.TeamName, cfg.BotUsername)
		if err != nil {
			exitErrf("Failed to resolve bot channel: %v", err)
		}
		store, err := tcli.NewStore(ch.ID)
		if err != nil {
			exitErrf("Failed to open local storage: %v", err)
		}

		uc := notifications.New(tcli.NewNotificationsSource(store)).WithPermalinker(cfg)

		var items []notifications.Notification
		switch {
		case notifsLongread > 0:
			items, err = uc.ForLongread(ctx, notifsLongread)
		case notifsCourse > 0:
			items, err = uc.ForCourse(ctx, notifsCourse)
		default:
			items, err = uc.All(ctx)
		}
		if err != nil {
			exitErrf("List notifications: %v", err)
		}
		if notifsLimit > 0 && len(items) > notifsLimit {
			items = items[:notifsLimit]
		}
		if len(items) == 0 {
			fmt.Println("No notifications matched.")
			return
		}
		for _, n := range items {
			fmt.Printf("[%s] %s\n", n.PostedAt.Format("2006-01-02 15:04"), labelFor(n.Kind))
			if n.Title != "" {
				fmt.Printf("  %s\n", n.Title)
			}
			if n.LMSURL != "" {
				fmt.Printf("  LMS:   %s\n", n.LMSURL)
			}
			if n.PostURL != "" {
				fmt.Printf("  Чат:   %s\n", n.PostURL)
			}
			fmt.Println()
		}
	},
}

func labelFor(kind string) string {
	switch kind {
	case notifications.KindNewTask:
		return "Новая задача"
	case notifications.KindGraded:
		return "Задача оценена"
	case notifications.KindRecording:
		return "Запись занятия"
	default:
		return "Сообщение"
	}
}

func init() {
	mmNotificationsCmd.Flags().IntVar(&notifsCourse, "course", 0, "Filter by LMS course ID")
	mmNotificationsCmd.Flags().IntVar(&notifsLongread, "longread", 0, "Filter by LMS longread ID")
	mmNotificationsCmd.Flags().IntVar(&notifsLimit, "limit", defaultNotifsLimit, "Cap on results (0 = no cap)")
	mmCmd.AddCommand(mmNotificationsCmd)
}
