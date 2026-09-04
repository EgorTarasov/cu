package command

import (
	"context"
	"fmt"
	"time"

	tcli "github.com/EgorTarasov/cu/internal/gateway/timeclient"

	"github.com/spf13/cobra"
)

var mmCmd = &cobra.Command{
	Use:     "time",
	Aliases: []string{"mm"},
	Short:   "time.cu.ru commands",
	Long:    "Access the CU time.cu.ru chat: sync bot posts locally, list aggregated messages and notifications.",
}

var (
	mmSyncBatchSize int
	mmSyncMaxPages  int
	mmPostsLimit    int
)

const (
	defaultSyncPageSize = 200
	defaultSyncMaxPages = 100
	defaultPostsLimit   = 20
)

var mmSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Pull new posts from the notification bot channel into local storage",
	Long: `Resolves the direct-message channel with the configured bot username and
appends new posts to ~/.cu-cli/mm/<channel-id>/posts.jsonl.

Configuration (env vars):
  CU_MM_BASE_URL      — base URL (default https://time.cu.ru)
  CU_MM_TEAM          — team slug (default tsentralnyy-universitet)
  CU_MM_BOT_USERNAME  — bot username (default cu_notification_bot)`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		cfg := tcli.LoadConfig()

		client, err := tcli.NewClientFromEnv()
		if err != nil {
			exitErrf("%v", err)
		}

		me, bot, ch, err := client.ResolveBotChannel(ctx, cfg.TeamName, cfg.BotUsername)
		if err != nil {
			exitErrf("Failed to resolve bot channel: %v", err)
		}
		fmt.Printf("User: @%s (%s)\n", me.Username, me.ID)
		fmt.Printf("Bot:  @%s (%s)\n", bot.Username, bot.ID)
		fmt.Printf("DM channel: %s\n", ch.ID)

		store, err := tcli.NewStore(ch.ID)
		if err != nil {
			exitErrf("Failed to open local storage: %v", err)
		}
		state, err := store.LoadState()
		if err != nil {
			exitErrf("Failed to read state: %v", err)
		}
		state.ChannelID = ch.ID
		state.BotUserID = bot.ID
		state.BotUsername = bot.Username

		total, lastID, lastAt := syncChannel(ctx, client, store, ch.ID, state, mmSyncBatchSize, mmSyncMaxPages)
		if lastID != "" {
			state.LastPostID = lastID
		}
		if lastAt > state.LastCreateAt {
			state.LastCreateAt = lastAt
		}
		state.SyncedAt = time.Now().UnixMilli()
		if err := store.SaveState(state); err != nil {
			exitErrf("Failed to save state: %v", err)
		}

		fmt.Printf("Synced %d new post(s) into %s\n", total, store.Dir())
	},
}

// syncChannel pages through posts using either an incremental `since` cursor
// (first run keeps `since==0` which becomes a full backfill) or `after` if a
// last post ID is known. Returns the count appended and the latest seen id/time.
func syncChannel(
	ctx context.Context,
	c *tcli.Client,
	store *tcli.Store,
	channelID string,
	state tcli.State,
	perPage, maxPages int,
) (int, string, int64) {
	if perPage <= 0 {
		perPage = defaultSyncPageSize
	}
	if maxPages <= 0 {
		maxPages = defaultSyncMaxPages
	}

	q := tcli.PostsQuery{PerPage: perPage}
	if state.LastPostID != "" {
		q.After = state.LastPostID
	} else if state.LastCreateAt > 0 {
		q.Since = state.LastCreateAt
	}

	total := 0
	var lastID string
	var lastAt int64

	for page := range maxPages {
		list, err := c.GetChannelPosts(ctx, channelID, q)
		if err != nil {
			exitErrf("Failed to fetch posts (page %d): %v", page, err)
		}
		if list == nil || len(list.Posts) == 0 {
			break
		}

		batch := make([]*tcli.Post, 0, len(list.Posts))
		for _, p := range list.Posts {
			batch = append(batch, p)
		}
		added, lid, lat, err := store.AppendPosts(batch)
		if err != nil {
			exitErrf("Failed to write posts: %v", err)
		}
		total += added
		if lid != "" {
			lastID = lid
		}
		if lat > lastAt {
			lastAt = lat
		}

		// Page forward using NextPostID when available; otherwise stop.
		if list.NextPostID == "" {
			break
		}
		q = tcli.PostsQuery{After: list.NextPostID, PerPage: perPage}
	}

	return total, lastID, lastAt
}

var mmPostsCmd = &cobra.Command{
	Use:   "posts",
	Short: "Print recent posts from local storage",
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
		posts, err := store.ReadLastN(mmPostsLimit)
		if err != nil {
			exitErrf("Failed to read posts: %v", err)
		}
		if len(posts) == 0 {
			fmt.Println("No posts in local storage. Run 'cuni time sync' first.")
			return
		}
		for _, p := range posts {
			ts := time.UnixMilli(p.CreateAt).Format(time.RFC3339)
			fmt.Printf("─── %s  %s ───\n%s\n\n", ts, p.ID, p.Message)
		}
	},
}

func init() {
	mmSyncCmd.Flags().IntVar(&mmSyncBatchSize, "page-size", defaultSyncPageSize, "Posts per API page (max 200)")
	mmSyncCmd.Flags().IntVar(&mmSyncMaxPages, "max-pages", defaultSyncMaxPages, "Safety cap on pagination")
	mmPostsCmd.Flags().IntVar(&mmPostsLimit, "limit", defaultPostsLimit, "How many recent posts to print")

	mmCmd.AddCommand(mmSyncCmd)
	mmCmd.AddCommand(mmPostsCmd)
}
