package accounts

import (
	"coderchirag/folio-tracker/lib/db/models"
	"coderchirag/folio-tracker/lib/db/models/types"
	"coderchirag/folio-tracker/lib/validation"
	"context"
	"fmt"
)

type Service struct {
	users UserStore
}

func NewService(users UserStore) *Service {
	return &Service{
		users: users,
	}
}

func (s *Service) SignUp(ctx context.Context, in SignUpInput) (*models.User, error) {
	u, err := models.NewUser(models.UserParams{
		Name:  in.Name,
		Email: in.Email,
		PAN:   in.PAN,
	})
	if err != nil {
		return nil, err
	}

	if err := s.users.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("AccountsService.signup: %w", err)
	}

	return u, nil
}

func (s *Service) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	userID, err := types.ParseID(id)
	if err != nil {
		var errs validation.Errors
		errs.Add("id", ErrInvalidUserID)
		return nil, errs
	}

	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("AccountsService.getUserById: %w", err)
	}

	return u, nil
}
