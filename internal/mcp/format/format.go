package format

import (
	"fmt"
	"strings"

	"cu-sync/internal/gateway/cu"
	"cu-sync/internal/model"
	materialsUC "cu-sync/internal/usecase/materials"
)

func CoursesList(courses []cu.StudentCourse) string {
	var b strings.Builder
	b.WriteString("## Your Courses\n\n")
	b.WriteString("| ID | Name | Category |\n")
	b.WriteString("|-----|------|----------|\n")

	for _, c := range courses {
		cat := c.Category
		if cat == "" {
			cat = "-"
		}

		fmt.Fprintf(&b, "| %d | %s | %s |\n", c.ID, c.Name, cat)
	}

	fmt.Fprintf(&b, "\n%d courses total.\n", len(courses))
	return b.String()
}

func SearchResults(courses []cu.StudentCourse, query string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## Search: %q\n\n", query)

	if len(courses) == 0 {
		b.WriteString("No courses found.\n")
		return b.String()
	}

	b.WriteString("| ID | Name |\n")
	b.WriteString("|-----|------|\n")

	for _, c := range courses {
		fmt.Fprintf(&b, "| %d | %s |\n", c.ID, c.Name)
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

	b.WriteString("| Urgency | Status | Time Left | Deadline | Exercise | Course |\n")
	b.WriteString("|---------|--------|-----------|----------|----------|--------|\n")

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

		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s |\n",
			icon, dl.StateLabel, dl.Deadline.TimeLeft(),
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
	b.WriteString("| Status | Score | Exercise |\n")
	b.WriteString("|--------|-------|----------|\n")

	for _, t := range result.Tasks {
		score := "-"
		if t.Score != nil {
			score = fmt.Sprintf("%.0f", *t.Score)
		}
		fmt.Fprintf(&b, "| %s | %s/%d | %s |\n",
			t.State.Label(), score, t.MaxScore, t.Name)
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

	if t.StartedAt != nil {
		fmt.Fprintf(&b, "**Started:** %s\n", t.StartedAt.Format(model.DateTimeFormat))
	}
	if t.SubmitAt != nil {
		fmt.Fprintf(&b, "**Submitted:** %s\n", t.SubmitAt.Format(model.DateTimeFormat))
	}
	if t.RejectAt != nil {
		fmt.Fprintf(&b, "**Rejected:** %s\n", t.RejectAt.Format(model.DateTimeFormat))
	}
	if t.EvaluateAt != nil {
		fmt.Fprintf(&b, "**Evaluated:** %s\n", t.EvaluateAt.Format(model.DateTimeFormat))
	}

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
