package cu

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func (c *Client) GetStudentCourses(ctx context.Context, limit int, state string) (*StudentCoursesResponse, error) {
	if c.bffCookie == "" {
		return nil, fmt.Errorf("bff.cookie is required for authentication")
	}

	endpoint := "/api/micro-lms/courses/student"
	params := url.Values{}

	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
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
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
			return nil, fmt.Errorf("HTTP %d: failed to decode error response", resp.StatusCode)
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, apiErr.Error())
	}

	var courses StudentCoursesResponse
	if err := json.NewDecoder(resp.Body).Decode(&courses); err != nil {
		return nil, fmt.Errorf("failed to decode student courses response: %w", err)
	}

	return &courses, nil
}

func (c *Client) GetCourseOverview(ctx context.Context, courseID int) (*CourseOverview, error) {
	if c.bffCookie == "" {
		return nil, fmt.Errorf("bff.cookie is required for authentication")
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
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
			return nil, fmt.Errorf("HTTP %d: failed to decode error response", resp.StatusCode)
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, apiErr.Error())
	}

	var courseOverview CourseOverview
	if err := json.NewDecoder(resp.Body).Decode(&courseOverview); err != nil {
		return nil, fmt.Errorf("failed to decode course overview response: %w", err)
	}

	return &courseOverview, nil
}
