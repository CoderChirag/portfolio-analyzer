package accounts

import (
	"coderchirag/folio-tracker/lib/db/models"
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
	u, err := models.NewUser(in.Name, in.Email, in.PAN)
	if err != nil {
		return nil, err
	}

	if err := s.users.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("signup: %w", err)
	}

	return u, nil
}
