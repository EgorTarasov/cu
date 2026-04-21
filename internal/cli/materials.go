package cli

import (
	"fmt"
	"os"

	"cu-sync/internal/cu"
	"cu-sync/internal/usecase/materials"
	"cu-sync/internal/usecase/materials/model/input"
	"cu-sync/internal/usecase/materials/model/output"

	"github.com/spf13/cobra"
)

func init() {
	materialsCmd.Flags().Int("week", 0, "download only a specific week number")
	materialsCmd.Flags().Bool("links", false, "only show links, don't download files")
	materialsCmd.Flags().String("path", ".", "output directory for downloads")
}

var materialsCmd = &cobra.Command{
	Use:   "materials <course>",
	Short: "Download course materials (PDFs, links)",
	Long: `Download all PDFs and show git/external links for a course.
Course can be specified by ID or by name (substring match).

Examples:
  cu materials go                 # download all Go course materials
  cu materials алго --week 8      # only week 8
  cu materials go --links         # just show links, no download
  cu materials 901 --path ./docs  # save to ./docs`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		client := mustClient()

		weekFilter, _ := cmd.Flags().GetInt("week")
		linksOnly, _ := cmd.Flags().GetBool("links")
		basePath, _ := cmd.Flags().GetString("path")

		// Try to create GitLab client (optional — will skip git downloads if unavailable).
		var gitlab materials.GitLabDownloader
		gitlabClient, gitlabErr := cu.NewGitLabClientFromEnv()
		if gitlabErr != nil && !linksOnly {
			fmt.Println("GitLab not configured — git.culab.ru links will be shown but not downloaded.")
			fmt.Println("Run 'cu login --gitlab' to enable.")
			fmt.Println()
		}
		if gitlabClient != nil {
			gitlab = gitlabClient
		}

		uc := materials.New(client, gitlab)

		onEvent := func(event output.MaterialEvent) {
			switch event.Type {
			case output.EventTheme:
				fmt.Printf("[%s]\n", event.Message)
			case output.EventPDF:
				fmt.Printf("  [PDF] %s\n", event.Message)
			case output.EventLink:
				fmt.Printf("  [link] %s\n", event.Message)
			case output.EventSaved:
				fmt.Printf("  saved: %s\n", event.Message)
			case output.EventError:
				fmt.Fprintf(os.Stderr, "  %s\n", event.Message)
			}
		}

		result, err := uc.Download(ctx, input.DownloadInput{
			CourseQuery: args[0],
			WeekFilter:  weekFilter,
			LinksOnly:   linksOnly,
			BasePath:    basePath,
		}, onEvent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed: %v\n", err)
			return
		}

		if !linksOnly {
			fmt.Printf("Downloaded %d/%d files\n", result.DownloadedFiles, result.TotalFiles)
		}
	},
}
