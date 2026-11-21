package accounts

import (
	"coderchirag/folio-tracker/lib/apperr"
	"coderchirag/folio-tracker/lib/domain"
	"context"
	"fmt"
)

const (
	logPrefix               = "service:AccountsService"
	signUpLogPrefix         = "Signup"
	getUserByIDLogPrefix    = "GetUserByID"
	getUserByEmailLogPrefix = "GetUserByEmail"
)

type Service struct {
	users UserStore
}

func New(users UserStore) *Service {
	return &Service{
		users: users,
	}
}

func (s *Service) SignUp(ctx context.Context, in *domain.UserParams) (*domain.User, error) {
	u, err := domain.NewUser(domain.UserParams{
		Name:  in.Name,
		Email: in.Email,
		PAN:   in.PAN,
	})
	if err != nil {
		return nil, err
	}

	if err := s.users.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("%s.%s: %w", logPrefix, signUpLogPrefix, err)
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
		return nil, fmt.Errorf("%s.%s: %w", logPrefix, getUserByIDLogPrefix, err)
	}

	return u, nil
}

func (s *Service) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	if err := domain.ValidateEmail(email); err != nil {
		var errs apperr.ValidationError
		errs.Add("email", err)
		return nil, errs
	}

	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("%s.%s: %w", logPrefix, getUserByEmailLogPrefix, err)
	}

	return u, nil
}
