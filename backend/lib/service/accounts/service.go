package accounts

import (
	"coderchirag/folio-tracker/lib/apperr"
	"coderchirag/folio-tracker/lib/domain"
	"context"
	"fmt"
)

type Service struct {
	users UserStore
}

func New(users UserStore) *Service {
	return &Service{
		users: users,
	}
}

func (s *Service) SignUp(ctx context.Context, in *SignUpInput) (*domain.User, error) {
	u, err := domain.NewUser(domain.UserParams{
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

func (s *Service) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	userID, err := domain.ParseID(id)
	if err != nil {
		var errs apperr.ValidationError
		errs.Add("id", ErrInvalidUserID)
		return nil, errs
	}

	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("AccountsService.getUserById: %w", err)
	}

	return u, nil
}
