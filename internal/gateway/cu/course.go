package cu

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

const maxCoursesLimit = 10000

func (c *Client) GetStudentCourses(ctx context.Context, limit int, state string) (*StudentCoursesResponse, error) {
	endpoint := StudentCoursesEndpoint
	params := url.Values{}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	if state != "" {
		params.Set("state", state)
	}
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	return doJSON[StudentCoursesResponse](ctx, c, endpoint)
}

func (c *Client) GetCourseOverview(ctx context.Context, courseID int) (*CourseOverview, error) {
	return doJSON[CourseOverview](ctx, c, fmt.Sprintf(CourseOverviewEndpoint, courseID))
}

func (c *Client) GetCourse(ctx context.Context, courseID int) (*Course, error) {
	return doJSON[Course](ctx, c, fmt.Sprintf(CourseEndpoint, courseID))
}

func (c *Client) GetTheme(ctx context.Context, themeID int) (*Theme, error) {
	return doJSON[Theme](ctx, c, fmt.Sprintf(ThemeEndpoint, themeID))
}

// ResolveCourse finds a course by ID (numeric string) or by substring match on course name.
// Returns the matched course ID. If multiple courses match, returns all matches and an error.
func (c *Client) ResolveCourse(ctx context.Context, query string) (int, string, error) {
	// Try numeric ID first.
	if id, err := strconv.Atoi(query); err == nil {
		courses, err := c.GetStudentCourses(ctx, maxCoursesLimit, "published")
		if err != nil {
			return id, "", err
		}
		for _, course := range courses.Items {
			if course.ID == id {
				return id, course.Name, nil
			}
		}
		return 0, "", fmt.Errorf("course with ID %d not found", id)
	}

	// Substring match on name (case-insensitive).
	courses, err := c.GetStudentCourses(ctx, maxCoursesLimit, "published")
	if err != nil {
		return 0, "", fmt.Errorf("failed to fetch courses: %w", err)
	}

	queryLower := strings.ToLower(query)
	var matches []StudentCourse
	for _, course := range courses.Items {
		if strings.Contains(strings.ToLower(course.Name), queryLower) {
			matches = append(matches, course)
		}
	}

	switch len(matches) {
	case 0:
		return 0, "", fmt.Errorf("no course matching %q found", query)
	case 1:
		return matches[0].ID, matches[0].Name, nil
	default:
		// If multiple matches, try to find one where query matches as a whole word.
		wordBoundary := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(query) + `\b`)
		var wordMatches []StudentCourse
		for _, m := range matches {
			if wordBoundary.MatchString(m.Name) {
				wordMatches = append(wordMatches, m)
			}
		}
		if len(wordMatches) == 1 {
			return wordMatches[0].ID, wordMatches[0].Name, nil
		}

		var lines []string
		for _, m := range matches {
			lines = append(lines, fmt.Sprintf("  %d  %s", m.ID, m.Name))
		}
		return 0, "", fmt.Errorf(
			"multiple courses match %q:\n%s\nspecify more precisely or use ID",
			query, strings.Join(lines, "\n"),
		)
	}
}

// GetDeadlines fetches student deadlines, optionally filtered by courseId.
func (c *Client) GetDeadlines(ctx context.Context, limit int, courseID *int) ([]Deadline, error) {
	params := url.Values{}
	params.Set("limit", strconv.Itoa(limit))
	if courseID != nil {
		params.Set("courseId", strconv.Itoa(*courseID))
	}
	endpoint := DeadlinesEndpoint + "?" + params.Encode()

	out, err := doJSON[[]Deadline](ctx, c, endpoint)
	if err != nil {
		return nil, err
	}
	return *out, nil
}

// GetCourseProgress fetches the student's overall score in a course.
func (c *Client) GetCourseProgress(ctx context.Context, courseID int) (*CourseProgress, error) {
	return doJSON[CourseProgress](ctx, c, fmt.Sprintf(CourseProgressEndpoint, courseID))
}

// GetStudentPerformance fetches per-exercise scores for a course.
func (c *Client) GetStudentPerformance(ctx context.Context, courseID int) (*StudentPerformance, error) {
	return doJSON[StudentPerformance](ctx, c, fmt.Sprintf(StudentPerformanceEndpoint, courseID))
}

// GetActivitiesPerformance fetches performance grouped by activity type.
func (c *Client) GetActivitiesPerformance(ctx context.Context, courseID int) (*ActivitiesPerformance, error) {
	return doJSON[ActivitiesPerformance](ctx, c, fmt.Sprintf(ActivitiesPerformanceEndpoint, courseID))
}

// GetCourseExercises fetches all exercises for a course.
func (c *Client) GetCourseExercises(ctx context.Context, courseID int) (*CourseExercises, error) {
	return doJSON[CourseExercises](ctx, c, fmt.Sprintf(CourseExercisesEndpoint, courseID))
}

// GetTask fetches a single task by ID.
func (c *Client) GetTask(ctx context.Context, taskID int) (*Task, error) {
	return doJSON[Task](ctx, c, fmt.Sprintf(TaskEndpoint, taskID))
}
