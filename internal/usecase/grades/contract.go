package grades

import (
	"context"
	"cu-sync/internal/gateway/cu"
)

type LMSClient interface {
	ResolveCourse(ctx context.Context, query string) (int, string, error)
	GetStudentCourses(ctx context.Context, limit int, state string) (*cu.StudentCoursesResponse, error)
	GetCourseProgress(ctx context.Context, courseID int) (*cu.CourseProgress, error)
	GetStudentPerformance(ctx context.Context, courseID int) (*cu.StudentPerformance, error)
	GetActivitiesPerformance(ctx context.Context, courseID int) (*cu.ActivitiesPerformance, error)
	GetCourseExercises(ctx context.Context, courseID int) (*cu.CourseExercises, error)
}
