package materials

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	cuGw "cu-sync/internal/gateway/cu"
	mcpfmt "cu-sync/internal/mcp/format"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// LMSClient defines dependencies for this tool.
type LMSClient interface {
	ResolveCourse(ctx context.Context, query string) (int, string, error)
	GetCourseOverview(ctx context.Context, courseID int) (*cuGw.CourseOverview, error)
	GetLongReadContent(ctx context.Context, longReadID int) (*cuGw.MaterialsResponse, error)
}

// Input for the tool.
type Input struct {
	Course string `json:"course" jsonschema:"Course name or ID"`
	Week   int    `json:"week,omitempty" jsonschema:"Week number (optional — omit for all weeks)"`
}

// Definition is the MCP tool definition.
var Definition = &mcp.Tool{
	Name:        "get_materials",
	Description: "List course materials: PDF files and external links (without downloading)",
}

// NewHandler creates the tool handler.
func NewHandler(lms LMSClient) func(context.Context, *mcp.CallToolRequest, Input) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, in Input) (*mcp.CallToolResult, any, error) {
		ctx := context.Background()

		courseID, _, err := lms.ResolveCourse(ctx, in.Course)
		if err != nil {
			return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
		}

		overview, err := lms.GetCourseOverview(ctx, courseID)
		if err != nil {
			return textResult(fmt.Sprintf("Error: %v", err)), nil, nil
		}

		allMaterials := make(map[int]*cuGw.MaterialsResponse)

		for _, theme := range overview.Themes {
			if in.Week > 0 && !matchesWeek(theme.Name, in.Week) {
				continue
			}
			for _, lr := range theme.Longreads {
				mats, err := lms.GetLongReadContent(ctx, lr.ID)
				if err != nil {
					continue
				}
				allMaterials[lr.ID] = mats
			}
		}

		return textResult(mcpfmt.MaterialsList(overview, allMaterials)), nil, nil
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

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}
