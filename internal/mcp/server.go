package mcp

import (
	"context"
	"cu-sync/internal/mcp/tool/coursestructure"
	"cu-sync/internal/mcp/tool/downloadmaterials"
	"cu-sync/internal/mcp/tool/readmaterial"
	searchCourses "cu-sync/internal/mcp/tool/searchcourses"
	"cu-sync/internal/version"

	"cu-sync/internal/gateway/cu"
	"cu-sync/internal/mcp/tool/deadlines"
	"cu-sync/internal/mcp/tool/grades"
	"cu-sync/internal/mcp/tool/listcourses"
	"cu-sync/internal/mcp/tool/materials"
	"cu-sync/internal/mcp/tool/task"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

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

type GitLabClient interface {
	GetRawFile(ctx context.Context, project, ref, filePath string) ([]byte, error)
	DownloadGitLabLink(ctx context.Context, link, destDir string) ([]string, error)
}

type Server struct {
	lms    LMSClient
	gitlab GitLabClient
	srv    *mcp.Server
}

func NewServer(lms LMSClient, gitlab GitLabClient) *Server {
	s := &Server{lms: lms, gitlab: gitlab}

	s.srv = mcp.NewServer(&mcp.Implementation{
		Name:    "cu-university",
		Version: version.Version,
	}, nil)

	s.registerTools()

	return s
}

func (s *Server) Run(ctx context.Context) error {
	return s.srv.Run(ctx, &mcp.StdioTransport{})
}

func (s *Server) registerTools() {
	mcp.AddTool(s.srv, listcourses.Definition, listcourses.NewHandler(s.lms))
	mcp.AddTool(s.srv, searchCourses.Definition, searchCourses.NewHandler(s.lms))
	mcp.AddTool(s.srv, coursestructure.Definition, coursestructure.NewHandler(s.lms))
	mcp.AddTool(s.srv, deadlines.Definition, deadlines.NewHandler(s.lms))
	mcp.AddTool(s.srv, task.Definition, task.NewHandler(s.lms))
	mcp.AddTool(s.srv, grades.Definition, grades.NewHandler(s.lms))
	mcp.AddTool(s.srv, materials.Definition, materials.NewHandler(s.lms))
	mcp.AddTool(s.srv, readmaterial.Definition, readmaterial.NewHandler(s.lms, s.gitlab))
	mcp.AddTool(s.srv, downloadmaterials.Definition, downloadmaterials.NewHandler(s.lms, s.gitlab))
}
