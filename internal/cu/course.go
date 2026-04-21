package cu

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

func (c *Client) GetStudentCourses(ctx context.Context, limit int, state string) (*StudentCoursesResponse, error) {
	if c.bffCookie == "" {
		return nil, errors.New("bff.cookie is required for authentication")
	}

	endpoint := "/api/micro-lms/courses/student"
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

	req, err := c.prepareRequest(ctx, http.MethodGet, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request: %w", err)
	}

	resp, err := c.executeRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		if err = json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
			return nil, fmt.Errorf("HTTP %d: failed to decode error response", resp.StatusCode)
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, apiErr.Error())
	}

	var courses StudentCoursesResponse
	if err = json.NewDecoder(resp.Body).Decode(&courses); err != nil {
		return nil, fmt.Errorf("failed to decode student courses response: %w", err)
	}

	return &courses, nil
}

func (c *Client) GetCourseOverview(ctx context.Context, courseID int) (*CourseOverview, error) {
	if c.bffCookie == "" {
		return nil, errors.New("bff.cookie is required for authentication")
	}

	endpoint := fmt.Sprintf(CourseOverviewEndpoint, courseID)

	req, err := c.prepareRequest(ctx, http.MethodGet, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request: %w", err)
	}

	resp, err := c.executeRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		if err = json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
			return nil, fmt.Errorf("HTTP %d: failed to decode error response", resp.StatusCode)
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, apiErr.Error())
	}

	var courseOverview CourseOverview
	if err = json.NewDecoder(resp.Body).Decode(&courseOverview); err != nil {
		return nil, fmt.Errorf("failed to decode course overview response: %w", err)
	}

	return &courseOverview, nil
}

// ResolveCourse finds a course by ID (numeric string) or by substring match on course name.
// Returns the matched course ID. If multiple courses match, returns all matches and an error.
func (c *Client) ResolveCourse(ctx context.Context, query string) (int, string, error) {
	// Try numeric ID first.
	if id, err := strconv.Atoi(query); err == nil {
		courses, err := c.GetStudentCourses(ctx, 10000, "published")
		if err != nil {
			return id, "", nil // can't verify, but try with the ID
		}
		for _, course := range courses.Items {
			if course.ID == id {
				return id, course.Name, nil
			}
		}
		return 0, "", fmt.Errorf("course with ID %d not found", id)
	}

	// Substring match on name (case-insensitive).
	courses, err := c.GetStudentCourses(ctx, 10000, "published")
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
		return 0, "", fmt.Errorf("multiple courses match %q:\n%s\nspecify more precisely or use ID", query, strings.Join(lines, "\n"))
	}
}

// GetDeadlines fetches student deadlines, optionally filtered by courseId.
func (c *Client) GetDeadlines(ctx context.Context, limit int, courseID *int) ([]Deadline, error) {
	params := url.Values{}
	params.Set("limit", strconv.Itoa(limit))
	if courseID != nil {
		params.Set("courseId", strconv.Itoa(*courseID))
	}
	endpoint := "/api/micro-lms/deadlines?" + params.Encode()

	req, err := c.prepareRequest(ctx, http.MethodGet, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request: %w", err)
	}

	resp, err := c.executeRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var deadlines []Deadline
	if err = json.NewDecoder(resp.Body).Decode(&deadlines); err != nil {
		return nil, fmt.Errorf("failed to decode deadlines: %w", err)
	}
	return deadlines, nil
}

// GetCourseProgress fetches the student's overall score in a course.
func (c *Client) GetCourseProgress(ctx context.Context, courseID int) (*CourseProgress, error) {
	endpoint := fmt.Sprintf("/api/micro-lms/courses/%d/student/progress", courseID)
	req, err := c.prepareRequest(ctx, http.MethodGet, endpoint)
	if err != nil {
		return nil, err
	}
	resp, err := c.executeRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	var p CourseProgress
	if err = json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return nil, err
	}
	return &p, nil
}

// GetStudentPerformance fetches per-exercise scores for a course.
func (c *Client) GetStudentPerformance(ctx context.Context, courseID int) (*StudentPerformance, error) {
	endpoint := fmt.Sprintf("/api/micro-lms/courses/%d/student-performance", courseID)
	req, err := c.prepareRequest(ctx, http.MethodGet, endpoint)
	if err != nil {
		return nil, err
	}
	resp, err := c.executeRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	var sp StudentPerformance
	if err = json.NewDecoder(resp.Body).Decode(&sp); err != nil {
		return nil, err
	}
	return &sp, nil
}

// GetActivitiesPerformance fetches performance grouped by activity type.
func (c *Client) GetActivitiesPerformance(ctx context.Context, courseID int) (*ActivitiesPerformance, error) {
	endpoint := fmt.Sprintf("/api/micro-lms/courses/%d/activities-performance", courseID)
	req, err := c.prepareRequest(ctx, http.MethodGet, endpoint)
	if err != nil {
		return nil, err
	}
	resp, err := c.executeRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	var ap ActivitiesPerformance
	if err = json.NewDecoder(resp.Body).Decode(&ap); err != nil {
		return nil, err
	}
	return &ap, nil
}

// GetCourseExercises fetches all exercises for a course.
func (c *Client) GetCourseExercises(ctx context.Context, courseID int) (*CourseExercises, error) {
	endpoint := fmt.Sprintf("/api/micro-lms/courses/%d/exercises", courseID)
	req, err := c.prepareRequest(ctx, http.MethodGet, endpoint)
	if err != nil {
		return nil, err
	}
	resp, err := c.executeRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	var ce CourseExercises
	if err = json.NewDecoder(resp.Body).Decode(&ce); err != nil {
		return nil, err
	}
	return &ce, nil
}

// GetTask fetches a single task by ID.
func (c *Client) GetTask(ctx context.Context, taskID int) (*Task, error) {
	endpoint := fmt.Sprintf("/api/micro-lms/tasks/%d", taskID)
	req, err := c.prepareRequest(ctx, http.MethodGet, endpoint)
	if err != nil {
		return nil, err
	}
	resp, err := c.executeRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	var t Task
	if err = json.NewDecoder(resp.Body).Decode(&t); err != nil {
		return nil, err
	}
	return &t, nil
}
