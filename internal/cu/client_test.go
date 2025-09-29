package cu

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type ClientTestSuite struct {
	suite.Suite
	client     *Client
	testServer *httptest.Server
	bffCookie  string
}

func (s *ClientTestSuite) SetupSuite() {
	s.bffCookie = "test-cookie-value-123"
}

func (s *ClientTestSuite) SetupTest() {

	s.testServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cookie, err := r.Cookie("bff.cookie")
		if err != nil || cookie.Value != s.bffCookie {
			http.Error(w, `{"message":"Unauthorized","code":"AUTH_ERROR"}`, http.StatusUnauthorized)
			return
		}

		s.Equal("application/json, text/plain, */*", r.Header.Get("Accept"))
		s.Equal("same-origin", r.Header.Get("Sec-Fetch-Site"))

		switch r.URL.Path {
		case "/api/micro-lms/courses/519/overview":
			s.handleCourseOverview(w, r)
		case "/api/micro-lms/courses/student":
			s.handleStudentCourses(w, r)
		case "/api/account/me":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id":"test-user","email":"test@example.com"}`))
		case "/api/micro-lms/courses":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"courses":[]}`))
		default:
			http.NotFound(w, r)
		}
	}))

	s.client = NewClient(s.bffCookie)
	s.client.SetBaseURL(s.testServer.URL)
}

func (s *ClientTestSuite) TearDownTest() {
	if s.testServer != nil {
		s.testServer.Close()
	}
}

func (s *ClientTestSuite) handleCourseOverview(w http.ResponseWriter, r *http.Request) {
	courseOverview := CourseOverview{
		ID:          519,
		Name:        "Case Evenings (Кейс-вечера)",
		IsArchived:  false,
		State:       "published",
		PublishDate: &time.Time{},
		PublishedAt: &time.Time{},
		Settings: CourseSettings{
			SkillLevel:          "none",
			IsSkillLevelEnabled: false,
		},
		Themes: []Theme{
			{
				ID:    4399,
				Name:  "Силлабус",
				Order: 1,
				State: "published",
				Longreads: []Longread{
					{
						ID:        7739,
						Type:      "common",
						Name:      "Ссылка на силлабус",
						State:     "published",
						Exercises: []Exercise{},
					},
				},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(courseOverview)
}

func (s *ClientTestSuite) handleStudentCourses(w http.ResponseWriter, r *http.Request) {

	limit := r.URL.Query().Get("limit")
	state := r.URL.Query().Get("state")

	items := []StudentCourse{
		{
			ID:          519,
			Name:        "Case Evenings (Кейс-вечера)",
			State:       "published",
			IsArchived:  false,
			PublishDate: &time.Time{},
			PublishedAt: &time.Time{},
			Settings: CourseSettings{
				SkillLevel:          "none",
				IsSkillLevelEnabled: false,
			},
		},
		{
			ID:          520,
			Name:        "Advanced Analytics",
			State:       "published",
			IsArchived:  false,
			PublishDate: &time.Time{},
			PublishedAt: &time.Time{},
			Settings: CourseSettings{
				SkillLevel:          "none",
				IsSkillLevelEnabled: false,
			},
		},
	}

	if state != "" && state != "published" {
		items = []StudentCourse{}
	}

	if limit == "1" && len(items) > 1 {
		items = items[:1]
	}

	response := StudentCoursesResponse{
		Items: items,
		Paging: Paging{
			Limit:      10000,
			Offset:     0,
			TotalCount: len(items),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *ClientTestSuite) TestNewClient() {
	client := NewClient("test-cookie")
	s.NotNil(client)
	s.Equal("test-cookie", client.GetBffCookie())
	s.Equal(BaseURL, client.baseURL)
}

func (s *ClientTestSuite) TestNewClientWithOptions() {
	customTimeout := 60 * time.Second
	customUserAgent := "CustomAgent/1.0"

	client := NewClientWithOptions("test-cookie", customTimeout, customUserAgent)
	s.NotNil(client)
	s.Equal("test-cookie", client.GetBffCookie())
	s.Equal(customTimeout, client.httpClient.Timeout)
	s.Equal(customUserAgent, client.userAgent)
}

func (s *ClientTestSuite) TestSetBffCookie() {
	newCookie := "new-cookie-value"
	s.client.SetBffCookie(newCookie)
	s.Equal(newCookie, s.client.GetBffCookie())
}

func (s *ClientTestSuite) TestGetCourseOverview_Success() {
	courseOverview, err := s.client.GetCourseOverview(519)
	s.NoError(err)
	s.NotNil(courseOverview)
	s.Equal(519, courseOverview.ID)
	s.Equal("Case Evenings (Кейс-вечера)", courseOverview.Name)
	s.False(courseOverview.IsArchived)
	s.Equal("published", courseOverview.State)
	s.Len(courseOverview.Themes, 1)
	s.Equal("Силлабус", courseOverview.Themes[0].Name)
}

func (s *ClientTestSuite) TestGetCourseOverview_NoCookie() {
	client := NewClient("")
	client.SetBaseURL(s.testServer.URL)

	_, err := client.GetCourseOverview(519)
	s.Error(err)
	s.Contains(err.Error(), "bff.cookie is required")
}

func (s *ClientTestSuite) TestGetCourseOverview_InvalidCookie() {
	client := NewClient("invalid-cookie")
	client.SetBaseURL(s.testServer.URL)

	_, err := client.GetCourseOverview(519)
	s.Error(err)
	s.Contains(err.Error(), "HTTP 401")
}

func (s *ClientTestSuite) TestValidateCookie_Valid() {
	err := s.client.ValidateCookie()
	s.NoError(err)
}

func (s *ClientTestSuite) TestValidateCookie_Invalid() {
	client := NewClient("invalid-cookie")
	client.SetBaseURL(s.testServer.URL)

	err := client.ValidateCookie()
	s.Error(err)
	s.Contains(err.Error(), "invalid or expired")
}

func (s *ClientTestSuite) TestValidateCookie_NoCookie() {
	client := NewClient("")

	err := client.ValidateCookie()
	s.Error(err)
	s.Contains(err.Error(), "no bff.cookie set")
}

func (s *ClientTestSuite) TestGetStudentCourses_Success() {
	courses, err := s.client.GetStudentCourses(10000, "published")
	s.NoError(err)
	s.NotNil(courses)
	s.Len(courses.Items, 2)
	s.Equal(519, courses.Items[0].ID)
	s.Equal("Case Evenings (Кейс-вечера)", courses.Items[0].Name)
	s.Equal("published", courses.Items[0].State)
	s.False(courses.Items[0].IsArchived)
	s.Equal(2, courses.Paging.TotalCount)
}

func (s *ClientTestSuite) TestGetStudentCourses_WithLimit() {
	courses, err := s.client.GetStudentCourses(1, "published")
	s.NoError(err)
	s.NotNil(courses)
	s.Len(courses.Items, 1)
	s.Equal(519, courses.Items[0].ID)
	s.Equal(1, courses.Paging.TotalCount)
}

func (s *ClientTestSuite) TestGetStudentCourses_NoParameters() {
	courses, err := s.client.GetStudentCourses(0, "")
	s.NoError(err)
	s.NotNil(courses)
	s.Len(courses.Items, 2)
	s.Equal(2, courses.Paging.TotalCount)
}

func (s *ClientTestSuite) TestGetStudentCourses_NoCookie() {
	client := NewClient("")
	client.SetBaseURL(s.testServer.URL)

	_, err := client.GetStudentCourses(10000, "published")
	s.Error(err)
	s.Contains(err.Error(), "bff.cookie is required")
}

func (s *ClientTestSuite) TestGetStudentCourses_InvalidCookie() {
	client := NewClient("invalid-cookie")
	client.SetBaseURL(s.testServer.URL)

	_, err := client.GetStudentCourses(10000, "published")
	s.Error(err)
	s.Contains(err.Error(), "HTTP 401")
}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
