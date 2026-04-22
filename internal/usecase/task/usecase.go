package task

import (
	"context"
	"fmt"

	"cu-sync/internal/model"
)

const percentMul = 100

type UseCase struct {
	lms LMSClient
}

func New(lms LMSClient) *UseCase {
	return &UseCase{lms: lms}
}

func (uc *UseCase) Get(ctx context.Context, in model.TaskGetInput) (*model.TaskOutput, error) {
	t, err := uc.lms.GetTask(ctx, in.TaskID)
	if err != nil {
		return nil, fmt.Errorf("fetching task %d: %w", in.TaskID, err)
	}

	out := &model.TaskOutput{
		CourseName:      t.Course.Name,
		ThemeName:       t.Theme.Name,
		ExerciseName:    t.Exercise.Name,
		ActivityName:    t.Exercise.Activity.Name,
		ActivityWeight:  t.Exercise.Activity.Weight * percentMul,
		Deadline:        model.DeadLine(t.Deadline),
		StartedAt:       t.StartedAt,
		SubmitAt:        t.SubmitAt,
		RejectAt:        t.RejectAt,
		EvaluateAt:      t.EvaluateAt,
		MaxScore:        t.Exercise.MaxScore,
		LateDaysBalance: t.Student.LateDaysBalance,
		StateLabel:      model.TaskState(t.State).Label(),
	}

	if t.Score != nil {
		out.ScoreFormatted = fmt.Sprintf("%.0f/%d", *t.Score, t.Exercise.MaxScore)
	} else {
		out.ScoreFormatted = fmt.Sprintf("-/%d", t.Exercise.MaxScore)
	}

	if t.Reviewer != nil {
		out.Reviewer = &model.Reviewer{
			FirstName: t.Reviewer.FirstName,
			LastName:  t.Reviewer.LastName,
			Email:     t.Reviewer.Email,
		}
	}

	if t.Solution != nil {
		out.SolutionURL = t.Solution.SolutionURL
	}

	return out, nil
}
