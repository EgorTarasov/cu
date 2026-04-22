package materials

import (
	"context"
	"cu-sync/internal/gateway/cu"
)

type LMSClient interface {
	ResolveCourse(ctx context.Context, query string) (int, string, error)
	GetCourseOverview(ctx context.Context, courseID int) (*cu.CourseOverview, error)
	GetLongReadContent(ctx context.Context, longReadID int) (*cu.MaterialsResponse, error)
	DownloadFile(ctx context.Context, material cu.Material, destDir string) (string, error)
}

type GitLabDownloader interface {
	DownloadGitLabLink(ctx context.Context, link, destDir string) ([]string, error)
}
