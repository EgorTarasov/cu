package command

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	cugw "github.com/EgorTarasov/cu/internal/gateway/cu"
	tcli "github.com/EgorTarasov/cu/internal/gateway/timeclient"
	"github.com/EgorTarasov/cu/internal/usecase/notifications"
	"github.com/EgorTarasov/cu/internal/usecase/recordings"

	"github.com/spf13/cobra"
)

const showCoursesLimit = 10000

var showCmd = &cobra.Command{
	Use:   "show [query]",
	Short: "Pick a course interactively and print its full structure",
	Long: `Lists your courses, lets you narrow them down by typing, picks one and prints
the full theme/longread tree.

If time.cu.ru storage is configured, attaches recording links and chat
permalinks per longread/course.`,
	Args: cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		client := mustClient()

		all, err := client.GetStudentCourses(ctx, showCoursesLimit, "published")
		if err != nil {
			exitErrf("Fetch courses: %v", err)
		}
		if len(all.Items) == 0 {
			fmt.Println("No courses found.")
			return
		}

		query := strings.TrimSpace(strings.Join(args, " "))
		if query == "" {
			fmt.Print("Search course: ")
			line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
			query = strings.TrimSpace(line)
		}

		matches := fuzzyCourses(all.Items, query)
		if len(matches) == 0 {
			fmt.Printf("No course matched %q. Available:\n\n", query)
			for i, c := range all.Items {
				fmt.Printf("  %d. [%d] %s\n", i+1, c.ID, c.Name)
			}
			return
		}

		var picked cugw.StudentCourse
		switch len(matches) {
		case 1:
			picked = matches[0]
		default:
			fmt.Printf("Matched %d courses:\n", len(matches))
			for i, c := range matches {
				fmt.Printf("  [%d] %s (id %d)\n", i+1, c.Name, c.ID)
			}
			fmt.Print("\nPick [number]: ")
			line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
			n, convErr := strconv.Atoi(strings.TrimSpace(line))
			if convErr != nil || n < 1 || n > len(matches) {
				exitErrf("Invalid selection")
			}
			picked = matches[n-1]
		}

		overview, err := client.GetCourseOverview(ctx, picked.ID)
		if err != nil {
			exitErrf("Fetch course %d: %v", picked.ID, err)
		}

		notifIdx, recsForCourse := loadMMEnrichments(ctx, overview.Name)
		renderCourseTree(overview, notifIdx, recsForCourse)
	},
}

func fuzzyCourses(items []cugw.StudentCourse, query string) []cugw.StudentCourse {
	if query == "" {
		return items
	}
	tokens := strings.Fields(strings.ToLower(query))
	var out []cugw.StudentCourse
	for _, c := range items {
		name := strings.ToLower(c.Name)
		ok := true
		for _, tok := range tokens {
			if !strings.Contains(name, tok) {
				ok = false
				break
			}
		}
		if ok {
			out = append(out, c)
		}
	}
	return out
}

// loadMMEnrichments returns notifications indexed by longread ID and recordings
// matching the course name. Both are empty if mm isn't configured/available.
func loadMMEnrichments(
	ctx context.Context,
	courseName string,
) (map[int][]notifications.Notification, []recordings.Recording) {
	cfg := tcli.LoadConfig()
	client, err := tcli.NewClientFromEnv()
	if err != nil {
		return nil, nil
	}
	_, _, ch, err := client.ResolveBotChannel(ctx, cfg.TeamName, cfg.BotUsername)
	if err != nil {
		return nil, nil
	}
	store, err := tcli.NewStore(ch.ID)
	if err != nil {
		return nil, nil
	}
	notifUC := notifications.New(tcli.NewNotificationsSource(store)).WithPermalinker(cfg)
	idx, _ := notifUC.ByLongread(ctx)

	recUC := recordings.New(tcli.NewStoreSource(store)).WithPermalinker(cfg)
	recs, _ := recUC.ForCourse(ctx, courseShortName(courseName))
	return idx, recs
}

// courseShortName picks the longest token from the course name to use as a
// recordings search key (bot subjects rarely match the full LMS title).
func courseShortName(name string) string {
	best := ""
	for _, w := range strings.Fields(name) {
		w = strings.Trim(w, ".,()«»\"'")
		if len(w) > len(best) {
			best = w
		}
	}
	return best
}

func renderCourseTree(
	c *cugw.CourseOverview,
	notifs map[int][]notifications.Notification,
	recs []recordings.Recording,
) {
	fmt.Printf("Course: %s (id %d)\n", c.Name, c.ID)
	fmt.Printf("State: %s | Archived: %v | Themes: %d\n", c.State, c.IsArchived, len(c.Themes))

	if len(recs) > 0 {
		fmt.Printf("\nЗаписи занятий (%d):\n", len(recs))
		shown := recs
		const maxRecs = 10
		if len(shown) > maxRecs {
			shown = shown[:maxRecs]
		}
		for _, r := range shown {
			line := r.Subject
			if !r.Date.IsZero() {
				line = r.Date.Format("02.01.2006") + " · " + line
			}
			fmt.Printf("  • %s\n", line)
			for _, l := range r.Links {
				fmt.Printf("      %s\n", l.URL)
			}
			if r.PostURL != "" {
				fmt.Printf("      чат: %s\n", r.PostURL)
			}
		}
		if len(recs) > maxRecs {
			fmt.Printf("  … +%d more (use 'cu time recordings %s')\n", len(recs)-maxRecs, courseShortName(c.Name))
		}
	}

	fmt.Println()
	themes := append([]cugw.Theme(nil), c.Themes...)
	sort.Slice(themes, func(i, j int) bool { return themes[i].Order < themes[j].Order })
	for _, t := range themes {
		fmt.Printf("  %d. %s\n", t.Order, t.Name)
		longs := append([]cugw.Longread(nil), t.Longreads...)
		sort.Slice(longs, func(i, j int) bool { return longs[i].Order < longs[j].Order })
		for _, l := range longs {
			fmt.Printf("     - [%d] %s (%s)", l.ID, l.Name, l.Type)
			if len(l.Exercises) > 0 {
				fmt.Printf(" · exercises: %d", len(l.Exercises))
			}
			fmt.Println()
			for _, n := range notifs[l.ID] {
				fmt.Printf("       %s: %s\n", labelFor(n.Kind), n.PostURL)
			}
		}
	}
}

func init() {
	RootCmd.AddCommand(showCmd)
}
