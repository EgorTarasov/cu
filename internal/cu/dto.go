package cu

import "time"

// CourseOverview represents the complete course overview response.
type CourseOverview struct {
	ID          int            `json:"id"`
	Name        string         `json:"name"`
	IsArchived  bool           `json:"isArchived"`
	State       string         `json:"state"`
	PublishDate *time.Time     `json:"publishDate"`
	PublishedAt *time.Time     `json:"publishedAt"`
	Settings    CourseSettings `json:"settings"`
	Themes      []Theme        `json:"themes"`
}

// CourseSettings represents course configuration settings.
type CourseSettings struct {
	SkillLevel          string `json:"skillLevel"`
	IsSkillLevelEnabled bool   `json:"isSkillLevelEnabled"`
}

// Theme represents a course theme/module.
type Theme struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	Order       int        `json:"order"`
	State       string     `json:"state"`
	PublishDate *time.Time `json:"publishDate"`
	PublishedAt *time.Time `json:"publishedAt"`
	Longreads   []Longread `json:"longreads"`
}

// Longread represents a learning material within a theme.
type Longread struct {
	ID          int        `json:"id"`
	Type        string     `json:"type"`
	Name        string     `json:"name"`
	State       string     `json:"state"`
	PublishDate *time.Time `json:"publishDate"`
	PublishedAt *time.Time `json:"publishedAt"`
	Exercises   []Exercise `json:"exercises"`
}

// Exercise represents an exercise within a longread.
type Exercise struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	Type        string     `json:"type"`
	State       string     `json:"state"`
	PublishDate *time.Time `json:"publishDate"`
	PublishedAt *time.Time `json:"publishedAt"`
}

// StudentCourse represents a course in the student courses list.
type StudentCourse struct {
	ID          int            `json:"id"`
	Name        string         `json:"name"`
	State       string         `json:"state"`
	IsArchived  bool           `json:"isArchived"`
	PublishDate *time.Time     `json:"publishDate"`
	PublishedAt *time.Time     `json:"publishedAt"`
	Settings    CourseSettings `json:"settings"`
	SubjectID   *int           `json:"subjectId"`
	Progress    *Progress      `json:"progress,omitempty"`
}

// Progress represents course progress information.
type Progress struct {
	CompletedCount int     `json:"completedCount"`
	TotalCount     int     `json:"totalCount"`
	Percentage     float64 `json:"percentage"`
}

// Paging represents pagination information.
type Paging struct {
	Limit      int `json:"limit"`
	Offset     int `json:"offset"`
	TotalCount int `json:"totalCount"`
}

// StudentCoursesResponse represents the response from /api/micro-lms/courses/student.
type StudentCoursesResponse struct {
	Items  []StudentCourse `json:"items"`
	Paging Paging          `json:"paging"`
}

// CoursesListResponse represents a list of courses (for future use).
type CoursesListResponse struct {
	Courses []CourseOverview `json:"courses"`
	Total   int              `json:"total"`
}

// Material represents a material item in a longread.
type Material struct {
	Discriminator string       `json:"discriminator"`
	ViewContent   string       `json:"viewContent,omitempty"`
	ViewType      string       `json:"viewType,omitempty"`
	MediaType     string       `json:"mediaType,omitempty"`
	Filename      string       `json:"filename,omitempty"`
	Version       string       `json:"version,omitempty"`
	Length        int          `json:"length,omitempty"`
	State         string       `json:"state"`
	PublishDate   *time.Time   `json:"publishDate"`
	PublishedAt   *time.Time   `json:"publishedAt"`
	Content       *FileContent `json:"content,omitempty"`
	Type          string       `json:"type"`
	ID            int          `json:"id"`
	Order         int          `json:"order"`
}

// FileContent represents file content information.
type FileContent struct {
	Name      string `json:"name"`
	Filename  string `json:"filename"`
	MediaType string `json:"mediaType"`
	Version   string `json:"version"`
	Length    int    `json:"length"`
}

// MaterialsResponse represents the response from /api/micro-lms/longreads/{id}/materials.
type MaterialsResponse struct {
	Items  []Material `json:"items"`
	Paging Paging     `json:"paging"`
}

// DownloadLinkResponse represents the response from /api/micro-lms/content/download-link.
type DownloadLinkResponse struct {
	URL string `json:"url"`
}

// APIError represents an API error response.
type APIError struct {
	Message string `json:"message"`
	Code    string `json:"code"`
	Details string `json:"details"`
}

func (e APIError) Error() string {
	return e.Message
}
