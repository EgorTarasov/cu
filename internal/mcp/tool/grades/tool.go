package grades

import (
	"context"
	"fmt"

	cuGw "cu-sync/internal/gateway/cu"
	mcpfmt "cu-sync/internal/mcp/format"
	"cu-sync/internal/model"
	ucGrades "cu-sync/internal/usecase/grades"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// LMSClient defines dependencies for this tool.
type LMSClient interface {
	ResolveCourse(ctx context.Context, query string) (int, string, error)
	GetStudentCourses(ctx context.Context, limit int, state string) (*cuGw.StudentCoursesResponse, error)
	GetCourseProgress(ctx context.Context, courseID int) (*cuGw.CourseProgress, error)
	GetStudentPerformance(ctx context.Context, courseID int) (*cuGw.StudentPerformance, error)
	GetActivitiesPerformance(ctx context.Context, courseID int) (*cuGw.ActivitiesPerformance, error)
	GetCourseExercises(ctx context.Context, courseID int) (*cuGw.CourseExercises, error)
}

// Input for the tool.
type Input struct {
	Course string `json:"course,omitempty" jsonschema:"Course name or ID (optional — omit for all courses)"`
}

// Definition is the MCP tool definition.
var Definition = &mcp.Tool{
	Name:        "get_grades",
	Description: "Get grades. Without course: summary table. With course: detailed activity breakdown and per-task scores.",
}

// NewHandler creates the tool handler.
func NewHandler(lms LMSClient) func(context.Context, *mcp.CallToolRequest, Input) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, in Input) (*mcp.CallToolResult, any, error) {
		ctx := context.Background()
		uc := ucGrades.New(lms)

		if in.Course == "" {
			result, err := uc.Summary(ctx, model.GradesSummaryInput{})
			if err != nil {
				return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
			}
			return textResult(mcpfmt.GradesSummary(result.Items)), nil, nil
		}

		result, err := uc.Detailed(ctx, model.GradesDetailedInput{CourseQuery: in.Course})
		if err != nil {
			return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
		}

		return textResult(mcpfmt.GradesDetailed(result)), nil, nil
	}
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}
