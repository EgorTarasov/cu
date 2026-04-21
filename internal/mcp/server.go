package mcp

import (
	"context"

	"cu-sync/internal/gateway/cu"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// LMSClient combines all LMS API methods needed by MCP tools.
type LMSClient interface {
	GetStudentCourses(ctx context.Context, limit int, state string) (*cu.StudentCoursesResponse, error)
	ResolveCourse(ctx context.Context, query string) (int, string, error)
	GetCourseOverview(ctx context.Context, courseID int) (*cu.CourseOverview, error)
	GetDeadlines(ctx context.Context, limit int, courseID *int) ([]cu.Deadline, error)
	GetCourseProgress(ctx context.Context, courseID int) (*cu.CourseProgress, error)
	GetStudentPerformance(ctx context.Context, courseID int) (*cu.StudentPerformance, error)
	GetActivitiesPerformance(ctx context.Context, courseID int) (*cu.ActivitiesPerformance, error)
	GetCourseExercises(ctx context.Context, courseID int) (*cu.CourseExercises, error)
	GetTask(ctx context.Context, taskID int) (*cu.Task, error)
	GetLongReadContent(ctx context.Context, longReadID int) (*cu.MaterialsResponse, error)
	DownloadFile(ctx context.Context, material cu.Material, destDir string) (string, error)
}

// GitLabClient provides GitLab file access.
type GitLabClient interface {
	GetRawFile(ctx context.Context, project, ref, filePath string) ([]byte, error)
	DownloadGitLabLink(ctx context.Context, link, destDir string) ([]string, error)
}

// Server wraps an MCP server with LMS and GitLab clients.
type Server struct {
	lms    LMSClient
	gitlab GitLabClient
	srv    *mcp.Server
}

// NewServer creates an MCP server with all CU tools registered.
func NewServer(lms LMSClient, gitlab GitLabClient) *Server {
	s := &Server{lms: lms, gitlab: gitlab}

	s.srv = mcp.NewServer(&mcp.Implementation{
		Name:    "cu-university",
		Version: "1.0.0",
	}, nil)

	s.registerTools()

	return s
}

// Run starts the MCP server on stdio transport.
func (s *Server) Run(ctx context.Context) error {
	return s.srv.Run(ctx, &mcp.StdioTransport{})
}
