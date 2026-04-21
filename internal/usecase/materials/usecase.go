package materials

import (
	"context"
	cu2 "cu-sync/internal/gateway/cu"
	"fmt"
	"path/filepath"
	"sync/atomic"

	"cu-sync/internal/model"

	"golang.org/x/sync/errgroup"
)

const (
	defaultMaxDownloads = 10
	bytesPerKB          = 1024
)

// EventCallback is called for each material event during download.
type EventCallback func(event model.MaterialEvent)

// UseCase implements the materials business logic.
type UseCase struct {
	lms    LMSClient
	gitlab GitLabDownloader // may be nil
}

// New creates a new materials usecase.
func New(lms LMSClient, gitlab GitLabDownloader) *UseCase {
	return &UseCase{lms: lms, gitlab: gitlab}
}

// Download fetches course materials and emits events via the callback.
func (uc *UseCase) Download(
	ctx context.Context,
	in model.MaterialsDownloadInput,
	onEvent EventCallback,
) (*model.MaterialsDownloadOutput, error) {
	courseID, courseName, err := uc.lms.ResolveCourse(ctx, in.CourseQuery)
	if err != nil {
		return nil, fmt.Errorf("resolving course: %w", err)
	}

	overview, err := uc.lms.GetCourseOverview(ctx, courseID)
	if err != nil {
		return nil, fmt.Errorf("fetching course overview: %w", err)
	}

	var totalFiles atomic.Int32
	var downloaded atomic.Int32

	var g *errgroup.Group
	if !in.LinksOnly {
		eg, _ := errgroup.WithContext(ctx)
		eg.SetLimit(defaultMaxDownloads)
		g = eg
	}

	for _, theme := range overview.Themes {
		if in.WeekFilter > 0 && !matchesWeek(theme.Name, in.WeekFilter) {
			continue
		}

		onEvent(model.MaterialEvent{Type: model.MaterialEventTheme, Message: theme.Name})

		themeDir := filepath.Join(in.BasePath, sanitizeFilename(courseName),
			fmt.Sprintf("%02d-%s", theme.Order, sanitizeFilename(theme.Name)))

		for _, lr := range theme.Longreads {
			uc.processLongread(ctx, lr.ID, lr.Name, themeDir, in.LinksOnly, g, &totalFiles, &downloaded, onEvent)
		}
	}

	if g != nil {
		if err := g.Wait(); err != nil {
			return nil, fmt.Errorf("download error: %w", err)
		}
	}

	return &model.MaterialsDownloadOutput{
		TotalFiles:      totalFiles.Load(),
		DownloadedFiles: downloaded.Load(),
	}, nil
}

func (uc *UseCase) processLongread(
	ctx context.Context,
	longreadID int,
	longreadName, themeDir string,
	linksOnly bool,
	g *errgroup.Group,
	totalFiles, downloaded *atomic.Int32,
	onEvent EventCallback,
) {
	materials, err := uc.lms.GetLongReadContent(ctx, longreadID)
	if err != nil {
		onEvent(model.MaterialEvent{
			Type:    model.MaterialEventError,
			Message: fmt.Sprintf("failed to fetch %s: %v", longreadName, err),
		})
		return
	}

	for _, mat := range materials.Items {
		switch {
		case mat.Discriminator == "file" && mat.Content != nil:
			uc.processFile(ctx, mat, themeDir, linksOnly, g, totalFiles, downloaded, onEvent)
		case mat.Type == "markdown" && mat.ViewContent != "":
			uc.processMarkdown(ctx, mat.ViewContent, themeDir, linksOnly, g, totalFiles, downloaded, onEvent)
		}
	}
}

func (uc *UseCase) processFile(
	ctx context.Context,
	mat cu2.Material,
	themeDir string,
	linksOnly bool,
	g *errgroup.Group,
	totalFiles, downloaded *atomic.Int32,
	onEvent EventCallback,
) {
	if linksOnly {
		onEvent(model.MaterialEvent{
			Type:    model.MaterialEventPDF,
			Message: fmt.Sprintf("%s (%.1f KB)", mat.Content.Name, float64(mat.Length)/bytesPerKB),
		})
		return
	}

	totalFiles.Add(1)

	g.Go(func() error {
		fp, err := uc.lms.DownloadFile(ctx, mat, themeDir)
		if err != nil {
			onEvent(model.MaterialEvent{
				Type:    model.MaterialEventError,
				Message: fmt.Sprintf("failed: %s: %v", mat.Content.Name, err),
			})
			return nil
		}
		downloaded.Add(1)
		onEvent(model.MaterialEvent{
			Type:     model.MaterialEventSaved,
			Message:  filepath.Base(fp),
			FilePath: fp,
		})
		return nil
	})
}

func (uc *UseCase) processMarkdown(
	ctx context.Context,
	viewContent, themeDir string,
	linksOnly bool,
	g *errgroup.Group,
	totalFiles, downloaded *atomic.Int32,
	onEvent EventCallback,
) {
	links := extractLinks(viewContent)
	for _, link := range links {
		if !linksOnly && uc.gitlab != nil && cu2.IsGitLabLink(link) {
			totalFiles.Add(1)
			link := link
			g.Go(func() error {
				saved, err := uc.gitlab.DownloadGitLabLink(ctx, link, themeDir)
				if err != nil {
					onEvent(model.MaterialEvent{
						Type:    model.MaterialEventError,
						Message: fmt.Sprintf("failed: %s: %v", link, err),
					})
					return nil
				}
				for _, s := range saved {
					downloaded.Add(1)
					onEvent(model.MaterialEvent{
						Type:     model.MaterialEventSaved,
						Message:  filepath.Base(s),
						FilePath: s,
					})
				}
				return nil
			})
		} else {
			onEvent(model.MaterialEvent{
				Type:    model.MaterialEventLink,
				Message: link,
			})
		}
	}
}
