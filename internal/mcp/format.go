package mcp

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"cu-sync/internal/gateway/cu"
	"cu-sync/internal/model"
)

const (
	hoursPerDay    = 24
	minutesPerHour = 60
)

func formatCoursesList(courses []cu.StudentCourse) string {
	var b strings.Builder
	b.WriteString("## Your Courses\n\n")
	b.WriteString("| ID | Name | Category |\n")
	b.WriteString("|-----|------|----------|\n")

	for _, c := range courses {
		cat := c.Category
		if cat == "" {
			cat = "-"
		}

		b.WriteString(fmt.Sprintf("| %d | %s | %s |\n", c.ID, c.Name, cat))
	}

	b.WriteString(fmt.Sprintf("\n%d courses total.\n", len(courses)))
	return b.String()
}

func formatSearchResults(courses []cu.StudentCourse, query string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("## Search: %q\n\n", query))

	if len(courses) == 0 {
		b.WriteString("No courses found.\n")
		return b.String()
	}

	b.WriteString("| ID | Name |\n")
	b.WriteString("|-----|------|\n")

	for _, c := range courses {
		b.WriteString(fmt.Sprintf("| %d | %s |\n", c.ID, c.Name))
	}

	return b.String()
}

func formatCourseStructure(overview *cu.CourseOverview) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("## %s\n\n", overview.Name))
	b.WriteString(fmt.Sprintf("**ID:** %d | **State:** %s\n\n", overview.ID, overview.State))

	for _, theme := range overview.Themes {
		exerciseCount := 0
		for _, lr := range theme.Longreads {
			exerciseCount += len(lr.Exercises)
		}

		b.WriteString(fmt.Sprintf("### %d. %s\n", theme.Order, theme.Name))

		for _, lr := range theme.Longreads {
			b.WriteString(fmt.Sprintf("- **%s** (longread#%d)\n", lr.Name, lr.ID))
			for _, ex := range lr.Exercises {
				dl := ""
				if ex.Deadline != nil {
					dl = fmt.Sprintf(", deadline %s", ex.Deadline.Format("02 Jan 15:04"))
				}
				b.WriteString(fmt.Sprintf("  - %s — max %d%s\n", ex.Name, ex.MaxScore, dl))
			}
		}

		b.WriteByte('\n')
	}

	return b.String()
}

func formatDeadlines(result *model.DeadlinesListOutput) string {
	var b strings.Builder

	if result.CourseName != "" {
		b.WriteString(fmt.Sprintf("## Deadlines: %s\n\n", result.CourseName))
	} else {
		b.WriteString("## All Deadlines\n\n")
	}

	if len(result.Items) == 0 {
		b.WriteString("No upcoming deadlines.\n")
		return b.String()
	}

	b.WriteString("| Urgency | Status | Time Left | Deadline | Exercise | Course |\n")
	b.WriteString("|---------|--------|-----------|----------|----------|--------|\n")

	urgent, soon := 0, 0

	for _, dl := range result.Items {
		icon := "⚪"

		switch dl.Urgency {
		case model.UrgencyUrgent:
			icon = "🔴"
			urgent++
		case model.UrgencySoon:
			icon = "🟡"
			soon++
		case model.UrgencyNormal:
			// default
		}

		b.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %s |\n",
			icon, dl.StateLabel, dl.TimeLeft,
			dl.Deadline.Format("02 Jan 15:04"),
			dl.ExerciseName, dl.CourseName,
		))
	}

	b.WriteString(fmt.Sprintf("\n%d deadlines total.", len(result.Items)))
	if urgent > 0 || soon > 0 {
		b.WriteString(fmt.Sprintf(" %d urgent, %d soon.", urgent, soon))
	}

	b.WriteByte('\n')
	return b.String()
}

func formatGradesSummary(items []model.GradesSummaryItem) string {
	var b strings.Builder
	b.WriteString("## Grades Summary\n\n")
	b.WriteString("| Course | Score | Max |\n")
	b.WriteString("|--------|-------|-----|\n")

	for _, item := range items {
		if item.Error != nil {
			b.WriteString(fmt.Sprintf("| %s | error | - |\n", item.CourseName))
			continue
		}
		b.WriteString(fmt.Sprintf("| %s | %.1f | %.0f |\n",
			item.CourseName, item.EarnedScore, item.MaxScore))
	}

	return b.String()
}

func formatGradesDetailed(result *model.GradesDetailedOutput) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("## Grades: %s\n\n", result.CourseName))

	b.WriteString("### Activity Breakdown\n\n")
	b.WriteString("| Activity | Weight | Average | Total | Blocker |\n")
	b.WriteString("|----------|--------|---------|-------|---------|\n")

	for _, a := range result.Activities {
		weight := "-"
		if a.Weight > 0 {
			weight = fmt.Sprintf("%.0f%%", a.Weight*100) //nolint:mnd // percentage
		}
		blocker := ""
		if a.IsBlocker {
			blocker = "yes"
		}
		b.WriteString(fmt.Sprintf("| %s | %s | %.1f | %.1f | %s |\n",
			a.Name, weight, a.Average, a.Total, blocker))
	}

	b.WriteString(fmt.Sprintf("\n**Total score: %.1f**\n\n", result.TotalScore))

	b.WriteString("### Tasks\n\n")
	b.WriteString("| Status | Score | Exercise |\n")
	b.WriteString("|--------|-------|----------|\n")

	for _, t := range result.Tasks {
		score := "-"
		if t.Score != nil {
			score = fmt.Sprintf("%.0f", *t.Score)
		}
		b.WriteString(fmt.Sprintf("| %s | %s/%d | %s |\n",
			t.StateLabel, score, t.MaxScore, t.Name))
	}

	if len(result.Blockers) > 0 {
		b.WriteString("\n### Blockers\n\n")
		for _, bl := range result.Blockers {
			b.WriteString(fmt.Sprintf("- **%s** — need avg >= %.0f\n", bl.ActivityName, bl.Threshold))
		}
	}

	return b.String()
}

func formatTask(t *cu.Task) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("## Task: %s\n\n", t.Exercise.Name))
	b.WriteString(fmt.Sprintf("**Course:** %s\n", t.Course.Name))
	b.WriteString(fmt.Sprintf("**Theme:** %s\n", t.Theme.Name))
	weight := t.Exercise.Activity.Weight * 100 //nolint:mnd // percentage
	b.WriteString(fmt.Sprintf("**Activity:** %s (%.0f%%)\n\n",
		t.Exercise.Activity.Name, weight))

	b.WriteString(fmt.Sprintf("**State:** %s\n", stateLabel(t.State)))

	if t.Score != nil {
		b.WriteString(fmt.Sprintf("**Score:** %.0f/%d\n", *t.Score, t.Exercise.MaxScore))
	} else {
		b.WriteString(fmt.Sprintf("**Score:** -/%d\n", t.Exercise.MaxScore))
	}

	b.WriteString(fmt.Sprintf("**Deadline:** %s (%s)\n", t.Deadline.Format("02 Jan 2006 15:04"), timeLeft(t.Deadline)))

	if t.StartedAt != nil {
		b.WriteString(fmt.Sprintf("**Started:** %s\n", t.StartedAt.Format("02 Jan 2006 15:04")))
	}
	if t.SubmitAt != nil {
		b.WriteString(fmt.Sprintf("**Submitted:** %s\n", t.SubmitAt.Format("02 Jan 2006 15:04")))
	}
	if t.RejectAt != nil {
		b.WriteString(fmt.Sprintf("**Rejected:** %s\n", t.RejectAt.Format("02 Jan 2006 15:04")))
	}
	if t.EvaluateAt != nil {
		b.WriteString(fmt.Sprintf("**Evaluated:** %s\n", t.EvaluateAt.Format("02 Jan 2006 15:04")))
	}

	if t.Reviewer != nil {
		b.WriteString(fmt.Sprintf("\n**Reviewer:** %s %s (%s)\n",
			t.Reviewer.FirstName, t.Reviewer.LastName, t.Reviewer.Email))
	}
	if t.Solution != nil && t.Solution.SolutionURL != "" {
		b.WriteString(fmt.Sprintf("**Solution:** %s\n", t.Solution.SolutionURL))
	}

	b.WriteString(fmt.Sprintf("\n**Late days balance:** %d\n", t.Student.LateDaysBalance))

	return b.String()
}

func formatMaterialsList(overview *cu.CourseOverview, materials map[int]*cu.MaterialsResponse) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("## Materials: %s\n\n", overview.Name))

	for _, theme := range overview.Themes {
		b.WriteString(fmt.Sprintf("### %s\n\n", theme.Name))

		for _, lr := range theme.Longreads {
			mats, ok := materials[lr.ID]
			if !ok {
				continue
			}

			for _, mat := range mats.Items {
				switch {
				case mat.Discriminator == "file" && mat.Content != nil:
					b.WriteString(fmt.Sprintf("- 📄 **%s** (%.1f KB)\n", mat.Content.Name, float64(mat.Length)/1024)) //nolint:mnd // bytes to KB
				case mat.Type == "markdown" && mat.ViewContent != "":
					links := extractLinks(mat.ViewContent)
					for _, link := range links {
						b.WriteString(fmt.Sprintf("- 🔗 %s\n", link))
					}
				}
			}
		}

		b.WriteByte('\n')
	}

	return b.String()
}

var linkPattern = regexp.MustCompile(`href=\\"([^"\\]+)\\"`)

func extractLinks(viewContent string) []string {
	matches := linkPattern.FindAllStringSubmatch(viewContent, -1)
	var links []string
	seen := make(map[string]bool)

	for _, m := range matches {
		link := m[1]
		if strings.HasPrefix(link, "#") || strings.Contains(link, "my.centraluniversity.ru") {
			continue
		}
		if !seen[link] {
			seen[link] = true
			links = append(links, link)
		}
	}

	return links
}

func stateLabel(state string) string {
	switch state {
	case "backlog":
		return "TODO"
	case "inProgress":
		return "IN PROGRESS"
	case "submitted":
		return "SUBMITTED"
	case "evaluated":
		return "DONE"
	case "failed":
		return "FAILED"
	default:
		return strings.ToUpper(state)
	}
}

func timeLeft(t time.Time) string {
	d := time.Until(t)
	if d < 0 {
		return "OVERDUE"
	}

	days := int(d.Hours() / hoursPerDay)
	hours := int(math.Mod(d.Hours(), hoursPerDay))

	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, int(math.Mod(d.Minutes(), minutesPerHour)))
	}

	return fmt.Sprintf("%dm", int(d.Minutes()))
}
