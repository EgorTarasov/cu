package mcp

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	cuGw "cu-sync/internal/gateway/cu"
	"cu-sync/internal/model"
	"cu-sync/internal/usecase/deadlines"
	"cu-sync/internal/usecase/grades"
	"cu-sync/internal/usecase/materials"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const maxCoursesLimit = 10000

// Tool input types.

type emptyInput struct{}

type courseQueryInput struct {
	Course string `json:"course,omitempty" jsonschema:"description=Course name or ID (optional — omit for all courses)"`
}

type courseStructureInput struct {
	Course string `json:"course"`
}

type searchInput struct {
	Query string `json:"query" jsonschema:"description=Search query (case-insensitive substring match)"`
}

type taskInput struct {
	TaskID int `json:"task_id" jsonschema:"description=Task ID (found in deadlines or grades output)"`
}

type materialsInput struct {
	Course string `json:"course" jsonschema:"description=Course name or ID"`
	Week   int    `json:"week,omitempty" jsonschema:"description=Week number (optional — omit for all weeks)"`
}

type readMaterialInput struct {
	Course string `json:"course" jsonschema:"description=Course name or ID"`
	Week   int    `json:"week" jsonschema:"description=Week number"`
	Type   string `json:"type,omitempty" jsonschema:"description=Material type: longread or seminar (optional — returns all markdown links)"`
}

type downloadMaterialsInput struct {
	Course string `json:"course" jsonschema:"description=Course name or ID"`
	Week   int    `json:"week,omitempty" jsonschema:"description=Week number (optional)"`
	Path   string `json:"path" jsonschema:"description=Directory to save files to"`
}

func (s *Server) registerTools() {
	mcp.AddTool(s.srv, &mcp.Tool{
		Name:        "list_courses",
		Description: "List all student courses with IDs and categories",
	}, s.handleListCourses)

	mcp.AddTool(s.srv, &mcp.Tool{
		Name:        "search_courses",
		Description: "Search courses by name (case-insensitive substring match)",
	}, s.handleSearchCourses)

	mcp.AddTool(s.srv, &mcp.Tool{
		Name:        "get_course_structure",
		Description: "Get full course tree: themes, longreads, exercises with deadlines",
	}, s.handleCourseStructure)

	mcp.AddTool(s.srv, &mcp.Tool{
		Name:        "list_deadlines",
		Description: "List upcoming deadlines sorted by date, with urgency markers. Optionally filter by course.",
	}, s.handleListDeadlines)

	mcp.AddTool(s.srv, &mcp.Tool{
		Name:        "get_task",
		Description: "Get detailed task info: score, reviewer, solution URL, deadline, state",
	}, s.handleGetTask)

	mcp.AddTool(s.srv, &mcp.Tool{
		Name:        "get_grades",
		Description: "Get grades. Without course: summary table. With course: detailed activity breakdown and per-task scores.",
	}, s.handleGetGrades)

	mcp.AddTool(s.srv, &mcp.Tool{
		Name:        "get_materials",
		Description: "List course materials: PDF files and external links (without downloading)",
	}, s.handleGetMaterials)

	mcp.AddTool(s.srv, &mcp.Tool{
		Name:        "read_material",
		Description: "Read a course material (longread/seminar) from GitLab and return its markdown content",
	}, s.handleReadMaterial)

	mcp.AddTool(s.srv, &mcp.Tool{
		Name:        "download_materials",
		Description: "Download course PDFs and markdown files to a local directory",
	}, s.handleDownloadMaterials)
}

func (s *Server) handleListCourses(_ context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
	ctx := context.Background()

	resp, err := s.lms.GetStudentCourses(ctx, maxCoursesLimit, "published")
	if err != nil {
		return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
	}

	return textResult(formatCoursesList(resp.Items)), nil, nil
}

func (s *Server) handleSearchCourses(_ context.Context, _ *mcp.CallToolRequest, in searchInput) (*mcp.CallToolResult, any, error) {
	ctx := context.Background()

	resp, err := s.lms.GetStudentCourses(ctx, maxCoursesLimit, "published")
	if err != nil {
		return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
	}

	query := strings.ToLower(in.Query)
	var matches []cuGw.StudentCourse

	for _, c := range resp.Items {
		if strings.Contains(strings.ToLower(c.Name), query) {
			matches = append(matches, c)
		}
	}

	return textResult(formatSearchResults(matches, in.Query)), nil, nil
}

func (s *Server) handleCourseStructure(_ context.Context, _ *mcp.CallToolRequest, in courseStructureInput) (*mcp.CallToolResult, any, error) {
	ctx := context.Background()

	courseID, _, err := s.lms.ResolveCourse(ctx, in.Course)
	if err != nil {
		return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
	}

	overview, err := s.lms.GetCourseOverview(ctx, courseID)
	if err != nil {
		return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
	}

	return textResult(formatCourseStructure(overview)), nil, nil
}

func (s *Server) handleListDeadlines(_ context.Context, _ *mcp.CallToolRequest, in courseQueryInput) (*mcp.CallToolResult, any, error) {
	ctx := context.Background()
	uc := deadlines.New(s.lms)

	result, err := uc.List(ctx, model.DeadlinesListInput{CourseQuery: in.Course})
	if err != nil {
		return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
	}

	return textResult(formatDeadlines(result)), nil, nil
}

func (s *Server) handleGetTask(_ context.Context, _ *mcp.CallToolRequest, in taskInput) (*mcp.CallToolResult, any, error) {
	ctx := context.Background()

	task, err := s.lms.GetTask(ctx, in.TaskID)
	if err != nil {
		return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
	}

	return textResult(formatTask(task)), nil, nil
}

func (s *Server) handleGetGrades(_ context.Context, _ *mcp.CallToolRequest, in courseQueryInput) (*mcp.CallToolResult, any, error) {
	ctx := context.Background()
	uc := grades.New(s.lms)

	if in.Course == "" {
		result, err := uc.Summary(ctx, model.GradesSummaryInput{})
		if err != nil {
			return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
		}
		return textResult(formatGradesSummary(result.Items)), nil, nil
	}

	result, err := uc.Detailed(ctx, model.GradesDetailedInput{CourseQuery: in.Course})
	if err != nil {
		return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
	}

	return textResult(formatGradesDetailed(result)), nil, nil
}

func (s *Server) handleGetMaterials(_ context.Context, _ *mcp.CallToolRequest, in materialsInput) (*mcp.CallToolResult, any, error) {
	ctx := context.Background()

	courseID, _, err := s.lms.ResolveCourse(ctx, in.Course)
	if err != nil {
		return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
	}

	overview, err := s.lms.GetCourseOverview(ctx, courseID)
	if err != nil {
		return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
	}

	allMaterials := make(map[int]*cuGw.MaterialsResponse)

	for _, theme := range overview.Themes {
		if in.Week > 0 && !matchesWeek(theme.Name, in.Week) {
			continue
		}
		for _, lr := range theme.Longreads {
			mats, err := s.lms.GetLongReadContent(ctx, lr.ID)
			if err != nil {
				continue
			}
			allMaterials[lr.ID] = mats
		}
	}

	return textResult(formatMaterialsList(overview, allMaterials)), nil, nil
}

func (s *Server) handleReadMaterial(
	_ context.Context, _ *mcp.CallToolRequest, in readMaterialInput,
) (*mcp.CallToolResult, any, error) {
	ctx := context.Background()

	if s.gitlab == nil {
		return textResult("Error: GitLab not configured. Run `cu login --gitlab` first."), nil, nil
	}

	courseID, _, err := s.lms.ResolveCourse(ctx, in.Course)
	if err != nil {
		return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
	}

	overview, err := s.lms.GetCourseOverview(ctx, courseID)
	if err != nil {
		return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
	}

	gitLinks := s.collectGitLinks(ctx, overview, in.Week, in.Type)
	if len(gitLinks) == 0 {
		return textResult(fmt.Sprintf("No GitLab materials found for week %d.", in.Week)), nil, nil
	}

	content := s.fetchGitContent(ctx, gitLinks)
	return textResult(content), nil, nil
}

func (s *Server) collectGitLinks(
	ctx context.Context, overview *cuGw.CourseOverview, week int, typeFilter string,
) []string {
	var gitLinks []string

	for _, theme := range overview.Themes {
		if !matchesWeek(theme.Name, week) {
			continue
		}
		for _, lr := range theme.Longreads {
			mats, err := s.lms.GetLongReadContent(ctx, lr.ID)
			if err != nil {
				continue
			}
			for _, mat := range mats.Items {
				if mat.Type != "markdown" || mat.ViewContent == "" {
					continue
				}
				for _, link := range extractLinks(mat.ViewContent) {
					if !cuGw.IsGitLabLink(link) {
						continue
					}
					if typeFilter != "" && !strings.Contains(strings.ToLower(link), strings.ToLower(typeFilter)) {
						continue
					}
					gitLinks = append(gitLinks, link)
				}
			}
		}
	}

	return gitLinks
}

func (s *Server) fetchGitContent(ctx context.Context, links []string) string {
	var b strings.Builder

	for _, link := range links {
		_, ref, path, _, ok := cuGw.ParseGitLabLink(link)
		if !ok {
			continue
		}
		if idx := strings.IndexByte(path, '?'); idx >= 0 {
			path = path[:idx]
		}
		project, _, _, _, _ := cuGw.ParseGitLabLink(link)

		data, err := s.gitlab.GetRawFile(ctx, project, ref, path)
		if err != nil {
			b.WriteString(fmt.Sprintf("<!-- Error fetching %s: %v -->\n\n", link, err))
			continue
		}

		b.WriteString(fmt.Sprintf("---\n## Source: %s\n\n", link))
		b.Write(data)
		b.WriteString("\n\n")
	}

	return b.String()
}

func (s *Server) handleDownloadMaterials(_ context.Context, _ *mcp.CallToolRequest, in downloadMaterialsInput) (*mcp.CallToolResult, any, error) {
	ctx := context.Background()
	uc := materials.New(s.lms, s.gitlab)

	var events []string
	onEvent := func(event model.MaterialEvent) {
		switch event.Type {
		case model.MaterialEventSaved:
			events = append(events, fmt.Sprintf("saved: %s", event.Message))
		case model.MaterialEventError:
			events = append(events, fmt.Sprintf("error: %s", event.Message))
		case model.MaterialEventTheme, model.MaterialEventPDF, model.MaterialEventLink:
			// skip verbose events
		}
	}

	result, err := uc.Download(ctx, model.MaterialsDownloadInput{
		CourseQuery: in.Course,
		WeekFilter:  in.Week,
		BasePath:    in.Path,
	}, onEvent)
	if err != nil {
		return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
	}

	var b strings.Builder
	b.WriteString("## Download Complete\n\n")
	b.WriteString(fmt.Sprintf("Downloaded %d/%d files to `%s`\n\n", result.DownloadedFiles, result.TotalFiles, in.Path))

	for _, e := range events {
		b.WriteString(fmt.Sprintf("- %s\n", e))
	}

	return textResult(b.String()), nil, nil
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}

var weekRe = regexp.MustCompile(`(?i)(?:неделя|week)\s*(\d+)`)

func matchesWeek(themeName string, week int) bool {
	m := weekRe.FindStringSubmatch(themeName)
	if len(m) < 2 { //nolint:mnd // regex match minimum parts
		return false
	}
	n, err := strconv.Atoi(m[1])
	return err == nil && n == week
}
