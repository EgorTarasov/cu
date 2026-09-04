package format

import (
	"fmt"
	"strings"

	"github.com/EgorTarasov/cu/internal/gateway/cu"
	"github.com/EgorTarasov/cu/internal/model"
	"github.com/EgorTarasov/cu/internal/render"
	materialsUC "github.com/EgorTarasov/cu/internal/usecase/materials"
)

func CoursesList(active, archived []cu.StudentCourse) string {
	var b strings.Builder
	b.WriteString("## Your Courses\n\n")

	writeCoursesSection(&b, "Active", active)
	if len(archived) > 0 {
		b.WriteString("\n")
		writeCoursesSection(&b, "Archived", archived)
	}

	fmt.Fprintf(&b, "\n%d active, %d archived.\n", len(active), len(archived))
	return b.String()
}

func writeCoursesSection(b *strings.Builder, title string, courses []cu.StudentCourse) {
	fmt.Fprintf(b, "### %s (%d)\n\n", title, len(courses))
	if len(courses) == 0 {
		b.WriteString("_none_\n")
		return
	}
	b.WriteString("| ID | Name | Category |\n")
	b.WriteString("|-----|------|----------|\n")
	for _, c := range courses {
		cat := c.Category
		if cat == "" {
			cat = "-"
		}
		fmt.Fprintf(b, "| %d | %s | %s |\n", c.ID, c.Name, cat)
	}
}

func SearchResults(active, archived []cu.StudentCourse, query string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## Search: %q\n\n", query)

	if len(active) == 0 && len(archived) == 0 {
		b.WriteString("No courses found.\n")
		return b.String()
	}

	b.WriteString("| ID | Name | Status |\n")
	b.WriteString("|-----|------|--------|\n")
	for _, c := range active {
		fmt.Fprintf(&b, "| %d | %s | active |\n", c.ID, c.Name)
	}
	for _, c := range archived {
		fmt.Fprintf(&b, "| %d | %s | archived |\n", c.ID, c.Name)
	}

	return b.String()
}

func CourseStructure(overview *cu.CourseOverview) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## %s\n\n", overview.Name)
	fmt.Fprintf(&b, "**ID:** %d | **State:** %s\n\n", overview.ID, overview.State)

	for _, theme := range overview.Themes {
		exerciseCount := 0
		for _, lr := range theme.Longreads {
			exerciseCount += len(lr.Exercises)
		}

		fmt.Fprintf(&b, "### %d. %s\n", theme.Order, theme.Name)

		for _, lr := range theme.Longreads {
			fmt.Fprintf(&b, "- **%s** (longread#%d)\n", lr.Name, lr.ID)
			for _, ex := range lr.Exercises {
				dl := ""
				if ex.Deadline != nil {
					dl = fmt.Sprintf(", deadline %s", ex.Deadline.Format(model.DateTimeShortFormat))
				}
				fmt.Fprintf(&b, "  - %s — max %d%s\n", ex.Name, ex.MaxScore, dl)
			}
		}

		b.WriteByte('\n')
	}

	return b.String()
}

func Deadlines(result *model.DeadlinesListOutput) string {
	var b strings.Builder

	if result.CourseName != "" {
		fmt.Fprintf(&b, "## Deadlines: %s\n\n", result.CourseName)
	} else {
		b.WriteString("## All Deadlines\n\n")
	}

	if len(result.Items) == 0 {
		b.WriteString("No upcoming deadlines.\n")
		return b.String()
	}

	b.WriteString("| Task ID | Urgency | Status | Time Left | Deadline | Exercise | Course |\n")
	b.WriteString("|---------|---------|--------|-----------|----------|----------|--------|\n")

	urgent, soon := 0, 0

	for _, dl := range result.Items {
		icon := "⚪"

		switch dl.Deadline.Urgency() {
		case model.UrgencyUrgent:
			icon = "🔴"
			urgent++
		case model.UrgencySoon:
			icon = "🟡"
			soon++
		case model.UrgencyNormal:
			// default
		}

		fmt.Fprintf(&b, "| %d | %s | %s | %s | %s | %s | %s |\n",
			dl.TaskID, icon, dl.StateLabel, dl.Deadline.TimeLeft(),
			dl.Deadline.Format(model.DateTimeShortFormat),
			dl.ExerciseName, dl.CourseName,
		)
	}

	fmt.Fprintf(&b, "\n%d deadlines total.", len(result.Items))
	if urgent > 0 || soon > 0 {
		fmt.Fprintf(&b, " %d urgent, %d soon.", urgent, soon)
	}

	b.WriteByte('\n')
	return b.String()
}

func GradesSummary(items []model.GradesSummaryItem) string {
	var b strings.Builder
	b.WriteString("## Grades Summary\n\n")
	b.WriteString("| Course | Score | Max |\n")
	b.WriteString("|--------|-------|-----|\n")

	for _, item := range items {
		if item.Error != nil {
			fmt.Fprintf(&b, "| %s | error | - |\n", item.CourseName)
			continue
		}
		fmt.Fprintf(&b, "| %s | %.1f | %.0f |\n",
			item.CourseName, item.EarnedScore, item.MaxScore)
	}

	return b.String()
}

func GradesDetailed(result *model.GradesDetailedOutput) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## Grades: %s\n\n", result.CourseName)

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
		fmt.Fprintf(&b, "| %s | %s | %.1f | %.1f | %s |\n",
			a.Name, weight, a.Average, a.Total, blocker)
	}

	fmt.Fprintf(&b, "\n**Total score: %.1f**\n\n", result.TotalScore)

	b.WriteString("### Tasks\n\n")
	b.WriteString("| Task ID | Status | Score | Exercise |\n")
	b.WriteString("|---------|--------|-------|----------|\n")

	for _, t := range result.Tasks {
		score := "-"
		if t.Score != nil {
			score = fmt.Sprintf("%.0f", *t.Score)
		}
		fmt.Fprintf(&b, "| %d | %s | %s/%d | %s |\n",
			t.TaskID, t.State.Label(), score, t.MaxScore, t.Name)
	}

	if len(result.Blockers) > 0 {
		b.WriteString("\n### Blockers\n\n")
		for _, bl := range result.Blockers {
			fmt.Fprintf(&b, "- **%s** — need avg >= %.0f\n", bl.ActivityName, bl.Threshold)
		}
	}

	return b.String()
}

func Task(t *model.TaskOutput) string {
	var b strings.Builder

	fmt.Fprintf(&b, "## Task: %s\n\n", t.ExerciseName)
	fmt.Fprintf(&b, "**Course:** %s\n", t.CourseName)
	fmt.Fprintf(&b, "**Theme:** %s\n", t.ThemeName)
	fmt.Fprintf(&b, "**Activity:** %s (%.0f%%)\n\n",
		t.ActivityName, t.ActivityWeight)

	fmt.Fprintf(&b, "**State:** %s\n", t.StateLabel)
	fmt.Fprintf(&b, "**Score:** %s\n", t.ScoreFormatted)
	fmt.Fprintf(&b, "**Deadline:** %s (%s)\n",
		t.Deadline.Format(model.DateTimeFormat), t.Deadline.TimeLeft())

	render.MDTimeLine(&b, "Started", t.StartedAt)
	render.MDTimeLine(&b, "Submitted", t.SubmitAt)
	render.MDTimeLine(&b, "Rejected", t.RejectAt)
	render.MDTimeLine(&b, "Evaluated", t.EvaluateAt)

	if t.Reviewer != nil {
		fmt.Fprintf(&b, "\n**Reviewer:** %s (%s)\n",
			t.Reviewer.FullName(), t.Reviewer.Email)
	}
	if t.SolutionURL != "" {
		fmt.Fprintf(&b, "**Solution:** %s\n", t.SolutionURL)
	}

	fmt.Fprintf(&b, "\n**Late days balance:** %d\n", t.LateDaysBalance)

	return b.String()
}

func MaterialsList(overview *cu.CourseOverview, materials map[int]*cu.MaterialsResponse) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## Materials: %s\n\n", overview.Name)

	for _, theme := range overview.Themes {
		fmt.Fprintf(&b, "### %s\n\n", theme.Name)

		for _, lr := range theme.Longreads {
			mats, ok := materials[lr.ID]
			if !ok {
				continue
			}

			for _, mat := range mats.Items {
				switch {
				case mat.Discriminator == "file" && mat.Content != nil:
					fmt.Fprintf(&b, "- 📄 **%s** (%.1f KB)\n", mat.Content.Name, float64(mat.Length)/1024) //nolint:mnd // bytes to KB
				case mat.Type == "markdown" && mat.ViewContent != "":
					links := materialsUC.ExtractLinks(mat.ViewContent)
					for _, link := range links {
						fmt.Fprintf(&b, "- 🔗 %s\n", link)
					}
				}
			}
		}

		b.WriteByte('\n')
	}

	return b.String()
}

func StudentProfile(student *cu.Student) string {
	var b strings.Builder

	fullName := strings.TrimSpace(strings.Join([]string{
		student.LastName,
		student.FirstName,
		student.MiddleName,
	}, " "))
	if fullName == "" {
		fullName = "Unknown"
	}

	fmt.Fprintf(&b, "## Student: %s\n\n", fullName)
	fmt.Fprintf(&b, "**ID:** %s\n", student.ID)
	if student.UniversityEmail != "" {
		fmt.Fprintf(&b, "**Email:** %s\n", student.UniversityEmail)
	}
	if student.StudyLevel != "" {
		fmt.Fprintf(&b, "**Study level:** %s\n", student.StudyLevel)
	}
	if student.StudyStartYear > 0 {
		fmt.Fprintf(&b, "**Study start year:** %d\n", student.StudyStartYear)
	}
	if student.TimeAccount != "" {
		fmt.Fprintf(&b, "**Time account:** %s\n", student.TimeAccount)
	}
	fmt.Fprintf(&b, "**Late days balance:** %d\n", student.LateDaysBalance)

	return b.String()
}

func CourseSummary(course *cu.Course) string {
	var b strings.Builder

	fmt.Fprintf(&b, "## Course: %s\n\n", course.Name)
	fmt.Fprintf(&b, "**ID:** %d | **State:** %s | **Archived:** %t\n",
		course.ID, course.State, course.IsArchived)

	if course.Category != "" {
		fmt.Fprintf(&b, "**Category:** %s\n", course.Category)
	}
	if course.CategoryCover != "" {
		fmt.Fprintf(&b, "**Category cover:** %s\n", course.CategoryCover)
	}
	if course.SubjectID != nil {
		fmt.Fprintf(&b, "**Subject ID:** %d\n", *course.SubjectID)
	}
	fmt.Fprintf(&b, "**Skill level:** %s (enabled: %t)\n",
		course.Settings.SkillLevel, course.Settings.IsSkillLevelEnabled)
	render.MDTimeLine(&b, "Publish date", course.PublishDate)
	render.MDTimeLine(&b, "Published at", course.PublishedAt)

	return b.String()
}

func ThemeSummary(theme *cu.Theme) string {
	var b strings.Builder

	fmt.Fprintf(&b, "## Theme: %s\n\n", theme.Name)
	fmt.Fprintf(&b, "**ID:** %d | **State:** %s | **Order:** %d\n",
		theme.ID, theme.State, theme.Order)
	render.MDTimeLine(&b, "Publish date", theme.PublishDate)
	render.MDTimeLine(&b, "Published at", theme.PublishedAt)

	return b.String()
}

func LongreadSummary(longread *cu.Longread) string {
	var b strings.Builder

	fmt.Fprintf(&b, "## Longread: %s\n\n", longread.Name)
	fmt.Fprintf(&b, "**ID:** %d | **Type:** %s | **State:** %s | **Order:** %d\n",
		longread.ID, longread.Type, longread.State, longread.Order)
	render.MDTimeLine(&b, "Publish date", longread.PublishDate)
	render.MDTimeLine(&b, "Published at", longread.PublishedAt)

	return b.String()
}
