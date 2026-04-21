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
	ID       int          `json:"id"`
	Name     string       `json:"name"`
	MaxScore int          `json:"maxScore"`
	Activity *ActivityRef `json:"activity,omitempty"`
	Deadline *time.Time   `json:"deadline,omitempty"`
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
	Category    string         `json:"category"`
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
	Name          string       `json:"name,omitempty"`
	ID            int          `json:"id"`
	Order         int          `json:"order"`
	TaskID        *int         `json:"taskId,omitempty"`
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

// Deadline represents a student's deadline for an exercise.
type Deadline struct {
	ID       int            `json:"id"`
	Exercise DeadlineExercise `json:"exercise"`
	State    string         `json:"state"`
	Deadline time.Time      `json:"deadline"`
	CreatedAt time.Time     `json:"createdAt"`
	RejectAt *time.Time     `json:"rejectAt"`
	Reviewer *Reviewer      `json:"reviewer"`
	Course   CourseRef      `json:"course"`
	Theme    ThemeRef       `json:"theme"`
	Longread LongreadRef    `json:"longread"`
}

// DeadlineExercise is the exercise info embedded in a deadline.
type DeadlineExercise struct {
	ID        int          `json:"id"`
	Name      string       `json:"name"`
	Type      string       `json:"type"`
	MaxScore  int          `json:"maxScore"`
	StartDate *time.Time   `json:"startDate"`
	Deadline  time.Time    `json:"deadline"`
	Activity  DeadlineActivity `json:"activity"`
}

// DeadlineActivity is the activity info in a deadline exercise.
type DeadlineActivity struct {
	ID                int     `json:"id"`
	Name              string  `json:"name"`
	Weight            float64 `json:"weight"`
	IsLateDaysEnabled bool    `json:"isLateDaysEnabled"`
}

// Reviewer represents a task reviewer.
type Reviewer struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	LastName  string `json:"lastName"`
	FirstName string `json:"firstName"`
}

// CourseRef is a minimal course reference.
type CourseRef struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	IsArchived bool   `json:"isArchived"`
}

// ThemeRef is a minimal theme reference.
type ThemeRef struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// LongreadRef is a minimal longread reference.
type LongreadRef struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// CourseProgress represents the student's overall progress in a course.
type CourseProgress struct {
	EarnedScore    float64 `json:"earnedScore"`
	LeftToEarnScore float64 `json:"leftToEarnScore"`
	MaxScore       float64 `json:"maxScore"`
}

// StudentPerformance represents the student's performance in a course.
type StudentPerformance struct {
	Tasks    []TaskScore `json:"tasks"`
	Blockers []Blocker   `json:"blockers"`
}

// TaskScore represents a single task's score in the gradebook.
type TaskScore struct {
	ID         int          `json:"id"`
	State      string       `json:"state"`
	Score      *float64     `json:"score"`
	ExerciseID int          `json:"exerciseId"`
	MaxScore   int          `json:"maxScore"`
	Activity   ActivityFull `json:"activity"`
}

// ActivityFull represents an activity with all fields.
type ActivityFull struct {
	ID                    int      `json:"id"`
	Name                  string   `json:"name"`
	Weight                float64  `json:"weight"`
	MaxExercisesCount     int      `json:"maxExercisesCount"`
	AverageScoreThreshold *float64 `json:"averageScoreThreshold"`
}

// ActivityRef is a minimal activity reference.
type ActivityRef struct {
	ID     int     `json:"id"`
	Name   string  `json:"name"`
	Weight float64 `json:"weight"`
}

// Blocker represents a blocker activity in performance.
type Blocker struct {
	ActivityID            int     `json:"activityId"`
	ActivityName          string  `json:"activityName"`
	AverageScoreThreshold float64 `json:"averageScoreThreshold"`
}

// ActivitiesPerformance represents performance grouped by activity type.
type ActivitiesPerformance struct {
	Items       []ActivityPerformanceItem `json:"items"`
	TotalWeight float64                   `json:"totalWeight"`
	TotalScore  float64                   `json:"totalScore"`
}

// ActivityPerformanceItem represents one activity's performance.
type ActivityPerformanceItem struct {
	Activity  ActivityFull `json:"activity"`
	Total     float64      `json:"total"`
	Average   float64      `json:"average"`
	IsBlocker bool         `json:"isBlocker"`
}

// Task represents a full task detail (student's assignment instance).
type Task struct {
	ID          int        `json:"id"`
	Type        string     `json:"type"`
	State       string     `json:"state"`
	Score       *float64   `json:"score"`
	CreatedAt   time.Time  `json:"createdAt"`
	StartedAt   *time.Time `json:"startedAt"`
	SubmitAt    *time.Time `json:"submitAt"`
	RejectAt    *time.Time `json:"rejectAt"`
	EvaluateAt  *time.Time `json:"evaluateAt"`
	Deadline    time.Time  `json:"deadline"`
	Exercise    TaskExercise `json:"exercise"`
	Course      CourseRef  `json:"course"`
	Theme       ThemeRef   `json:"theme"`
	Longread    LongreadRef `json:"longread"`
	Student     TaskStudent `json:"student"`
	Reviewer    *Reviewer  `json:"reviewer"`
	Solution    *Solution  `json:"solution"`
}

// TaskExercise is the exercise info embedded in a task.
type TaskExercise struct {
	ID          int          `json:"id"`
	Name        string       `json:"name"`
	Type        string       `json:"type"`
	MaxScore    int          `json:"maxScore"`
	StartDate   *time.Time   `json:"startDate"`
	Deadline    time.Time    `json:"deadline"`
	ViewContent string       `json:"viewContent"`
	Activity    ActivityRef  `json:"activity"`
}

// TaskStudent is the student info embedded in a task.
type TaskStudent struct {
	ID              string `json:"id"`
	LastName        string `json:"lastName"`
	FirstName       string `json:"firstName"`
	LateDaysBalance int    `json:"lateDaysBalance"`
}

// Solution represents a task solution.
type Solution struct {
	Type        string `json:"type"`
	SolutionURL string `json:"solutionUrl"`
}

// CourseExercises represents exercises list for a course.
type CourseExercises struct {
	ID        int               `json:"id"`
	Name      string            `json:"name"`
	Exercises []CourseExercise   `json:"exercises"`
}

// CourseExercise is an exercise in the course exercises list.
type CourseExercise struct {
	ID       int         `json:"id"`
	Name     string      `json:"name"`
	Type     string      `json:"type"`
	Activity ActivityRef `json:"activity"`
	Longread LongreadRef `json:"longread"`
	Theme    ThemeRef    `json:"theme"`
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
