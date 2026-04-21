package materials

import (
	"context"

	"cu-sync/internal/cu"
)

// LMSClient defines the subset of the LMS API needed by the materials usecase.
type LMSClient interface {
	ResolveCourse(ctx context.Context, query string) (int, string, error)
	GetCourseOverview(ctx context.Context, courseID int) (*cu.CourseOverview, error)
	GetLongReadContent(ctx context.Context, longReadID int) (*cu.MaterialsResponse, error)
	DownloadFile(ctx context.Context, material cu.Material, destDir string) (string, error)
}

// GitLabDownloader handles downloading files from git.culab.ru.
type GitLabDownloader interface {
	DownloadGitLabLink(ctx context.Context, link, destDir string) ([]string, error)
}
