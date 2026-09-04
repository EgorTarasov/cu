package command

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	tcli "github.com/EgorTarasov/cu/internal/gateway/timeclient"
	"github.com/EgorTarasov/cu/internal/usecase/recordings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	recordingsAll   bool
	recordingsLimit int
)

const defaultRecordingsLimit = 30

var mmRecordingsCmd = &cobra.Command{
	Use:   "recordings [query]",
	Short: "Search lesson recordings collected from the notification bot",
	Long: `Scans local time.cu.ru storage for "🎓 Записи занятий" posts, extracts
recordings and lets you pick one interactively.

Examples:
  cuni time recordings              # interactive over everything
  cuni time recordings go           # filter by subject substring
  cuni time recordings --all go     # print every match, no picker`,
	Args: cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
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

		uc := recordings.New(tcli.NewStoreSource(store)).WithPermalinker(cfg)
		query := strings.Join(args, " ")
		matches, err := uc.Search(ctx, query)
		if err != nil {
			exitErrf("Search failed: %v", err)
		}
		if len(matches) == 0 {
			fmt.Println("No recordings matched. Try 'cuni time sync' first.")
			return
		}
		if recordingsLimit > 0 && len(matches) > recordingsLimit {
			matches = matches[:recordingsLimit]
		}

		// Non-interactive paths: --all, single match, or stdin not a TTY.
		if recordingsAll || len(matches) == 1 || !term.IsTerminal(int(os.Stdin.Fd())) {
			for i, r := range matches {
				if i > 0 {
					fmt.Println()
				}
				fmt.Print(recordings.Format(r))
			}
			return
		}

		fmt.Printf("Found %d recording(s) matching %q:\n\n", len(matches), query)
		for i, r := range matches {
			fmt.Printf("  [%d] %s\n", i+1, recordings.FormatLine(r))
		}
		fmt.Print("\nSelect [number, or empty to cancel]: ")

		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line == "" {
			return
		}
		n, err := strconv.Atoi(line)
		if err != nil || n < 1 || n > len(matches) {
			exitErrf("Invalid selection: %s", line)
		}
		fmt.Println()
		fmt.Print(recordings.Format(matches[n-1]))
	},
}

func init() {
	mmRecordingsCmd.Flags().BoolVar(&recordingsAll, "all", false, "Print every match without picker")
	mmRecordingsCmd.Flags().IntVar(&recordingsLimit, "limit", defaultRecordingsLimit, "Cap on results (0 = no cap)")
	mmCmd.AddCommand(mmRecordingsCmd)
}
