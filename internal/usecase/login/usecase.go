package login

import (
	"context"
	"fmt"

	"cu-sync/internal/model"
)

// UseCase implements the login business logic.
type UseCase struct {
	login    AuthFunc
	save     SaveFunc
	validate ValidateFunc // may be nil
}

// New creates a new login usecase.
func New(login AuthFunc, save SaveFunc, validate ValidateFunc) *UseCase {
	return &UseCase{
		login:    login,
		save:     save,
		validate: validate,
	}
}

// Execute performs the authentication flow: login via browser, optionally validate, and save cookie.
func (uc *UseCase) Execute(ctx context.Context, in model.LoginInput) (*model.LoginOutput, error) {
	cookie, err := uc.login(ctx, in.Timeout)
	if err != nil {
		return nil, fmt.Errorf("login: %w", err)
	}

	var validationErr error
	if uc.validate != nil {
		validationErr = uc.validate()
	}

	if err := uc.save(cookie); err != nil {
		return nil, fmt.Errorf("saving cookie: %w", err)
	}

	return &model.LoginOutput{
		ValidationError: validationErr,
	}, nil
}
