package httpaccountshandler

import (
	"coderchirag/folio-tracker/lib/domain"
	"context"
)

type AccountService interface {
	SignUp(ctx context.Context, in *domain.UserParams) (*domain.User, error)
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
}
