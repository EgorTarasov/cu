package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"

	"cu-sync/internal/cu"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

const defaultMaxDownloads = 10

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

		courseID, courseName := mustResolveCourse(ctx, client, args[0])
		weekFilter, _ := cmd.Flags().GetInt("week")
		linksOnly, _ := cmd.Flags().GetBool("links")
		basePath, _ := cmd.Flags().GetString("path")

		fmt.Printf("Materials: %s\n\n", courseName)

		overview, err := client.GetCourseOverview(ctx, courseID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to fetch course: %v\n", err)
			return
		}

		var totalFiles atomic.Int32
		var downloaded atomic.Int32
		var g *errgroup.Group

		if !linksOnly {
			eg, egctx := errgroup.WithContext(ctx)
			eg.SetLimit(defaultMaxDownloads)
			g = eg
			_ = egctx
		}

		for _, theme := range overview.Themes {
			// Filter by week if specified.
			if weekFilter > 0 && !matchesWeek(theme.Name, weekFilter) {
				continue
			}

			fmt.Printf("[%s]\n", theme.Name)

			for _, lr := range theme.Longreads {
				materials, err := client.GetLongReadContent(ctx, lr.ID)
				if err != nil {
					fmt.Fprintf(os.Stderr, "  failed to fetch %s: %v\n", lr.Name, err)
					continue
				}

				for _, mat := range materials.Items {
					switch {
					case mat.Discriminator == "file" && mat.Content != nil:
						if linksOnly {
							fmt.Printf("  [PDF] %s (%.1f KB)\n", mat.Content.Name, float64(mat.Length)/1024)
						} else {
							totalFiles.Add(1)
							themeDir := filepath.Join(basePath, sanitizeFilename(courseName),
								fmt.Sprintf("%02d-%s", theme.Order, sanitizeFilename(theme.Name)))

							mat := mat
							g.Go(func() error {
								dlClient, err := cu.NewClientFromEnv()
								if err != nil {
									fmt.Fprintf(os.Stderr, "  failed: %s: %v\n", mat.Content.Name, err)
									return nil
								}
								fp, err := dlClient.DownloadFile(ctx, mat, themeDir)
								if err != nil {
									fmt.Fprintf(os.Stderr, "  failed: %s: %v\n", mat.Content.Name, err)
									return nil
								}
								downloaded.Add(1)
								fmt.Printf("  saved: %s\n", filepath.Base(fp))
								return nil
							})
						}

					case mat.Type == "markdown" && mat.ViewContent != "":
						links := extractLinks(mat.ViewContent)
						for _, link := range links {
							fmt.Printf("  [link] %s\n", link)
						}
					}
				}
			}
			fmt.Println()
		}

		if g != nil {
			if err := g.Wait(); err != nil {
				fmt.Fprintf(os.Stderr, "Download error: %v\n", err)
			}
			fmt.Printf("Downloaded %d/%d files\n", downloaded.Load(), totalFiles.Load())
		}
	},
}

var weekPattern = regexp.MustCompile(`(?i)(?:неделя|week)\s*(\d+)`)

func matchesWeek(themeName string, week int) bool {
	matches := weekPattern.FindStringSubmatch(themeName)
	if len(matches) < 2 {
		return false
	}
	n, err := strconv.Atoi(matches[1])
	return err == nil && n == week
}

var linkPattern = regexp.MustCompile(`href=\\"([^"\\]+)\\"`)

func extractLinks(viewContent string) []string {
	matches := linkPattern.FindAllStringSubmatch(viewContent, -1)
	var links []string
	seen := make(map[string]bool)
	for _, m := range matches {
		link := m[1]
		// Skip internal CU links and anchors.
		if strings.HasPrefix(link, "#") || strings.Contains(link, "my.centraluniversity.ru") {
			continue
		}
		if !seen[link] {
			seen[link] = true
			links = append(links, link)
		}
	}
	return links
}

// dumpCourse downloads all course materials (used by fetch course --dump).
func dumpCourse(
	ctx context.Context,
	client *cu.Client,
	course *cu.CourseOverview,
	courseDir string,
) error {
	fmt.Println("Downloading course materials...")

	totalFiles := atomic.Int32{}
	downloadedFiles := atomic.Int32{}

	g, _ := errgroup.WithContext(ctx)
	g.SetLimit(defaultMaxDownloads)

	for _, theme := range course.Themes {
		themeDir := filepath.Join(courseDir, fmt.Sprintf("%d-%s", theme.Order, sanitizeFilename(theme.Name)))
		processTheme(ctx, client, theme, themeDir, g, &totalFiles, &downloadedFiles)
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("error downloading course materials: %w", err)
	}

	fmt.Printf("Download complete: %d/%d files to %s\n",
		downloadedFiles.Load(), totalFiles.Load(), courseDir)
	return nil
}

func processTheme(
	ctx context.Context,
	client *cu.Client,
	theme cu.Theme,
	themeDir string,
	g *errgroup.Group,
	totalFiles *atomic.Int32,
	downloadedFiles *atomic.Int32,
) {
	for _, longread := range theme.Longreads {
		longreadDir := filepath.Join(themeDir, sanitizeFilename(longread.Name))
		processLongread(ctx, client, longread, longreadDir, g, totalFiles, downloadedFiles)
	}
}

func processLongread(
	ctx context.Context,
	client *cu.Client,
	longread cu.Longread,
	longreadDir string,
	g *errgroup.Group,
	totalFiles *atomic.Int32,
	downloadedFiles *atomic.Int32,
) {
	materials, err := client.GetLongReadContent(ctx, longread.ID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  failed to fetch %s: %v\n", longread.Name, err)
		return
	}

	for _, material := range materials.Items {
		if material.Discriminator != "file" {
			continue
		}
		totalFiles.Add(1)
		material := material
		g.Go(func() error {
			dlClient, err := cu.NewClientFromEnv()
			if err != nil {
				return nil
			}
			_, err = dlClient.DownloadFile(ctx, material, longreadDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  failed: %s: %v\n", material.Content.Name, err)
				return nil
			}
			downloadedFiles.Add(1)
			return nil
		})
	}
}
