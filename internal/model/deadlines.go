package model

import (
	"fmt"
	"math"
	"time"
)

type DeadlinesListInput struct {
	CourseQuery string
}

type UrgencyLevel string

const (
	UrgencyUrgent UrgencyLevel = "urgent"
	UrgencySoon   UrgencyLevel = "soon"
	UrgencyNormal UrgencyLevel = "normal"

	hoursPerDay    = 24
	minutesPerHour = 60

	DateTimeFormat      = "02 Jan 2006 15:04"
	DateTimeShortFormat = "02 Jan 15:04"
)

type DeadLine time.Time

func (d DeadLine) Time() time.Time {
	return time.Time(d)
}

func (d DeadLine) Format(layout string) string {
	return time.Time(d).Format(layout)
}

func (d DeadLine) Before(other DeadLine) bool {
	return time.Time(d).Before(time.Time(other))
}

func (d DeadLine) TimeLeft() string {
	dur := time.Until(time.Time(d))
	if dur < 0 {
		return "OVERDUE"
	}
	days := int(dur.Hours() / hoursPerDay)
	hours := int(math.Mod(dur.Hours(), hoursPerDay))
	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, int(math.Mod(dur.Minutes(), minutesPerHour)))
	}
	return fmt.Sprintf("%dm", int(dur.Minutes()))
}

func (d DeadLine) Urgency() UrgencyLevel {
	return d.UrgencyAt(time.Now())
}

func (d DeadLine) UrgencyAt(now time.Time) UrgencyLevel {
	remaining := time.Time(d).Sub(now)
	switch {
	case remaining < 0 || remaining < 24*time.Hour:
		return UrgencyUrgent
	case remaining < 3*24*time.Hour:
		return UrgencySoon
	default:
		return UrgencyNormal
	}
}

type DeadlineItem struct {
	ExerciseName string
	CourseName   string
	State        TaskState
	StateLabel   string
	Deadline     DeadLine
	Reviewer     *Reviewer
}

type DeadlinesListOutput struct {
	Items      []DeadlineItem
	CourseName string
}
