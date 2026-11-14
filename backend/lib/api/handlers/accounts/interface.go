package accountshandler

import (
	"coderchirag/folio-tracker/lib/domain"
	"coderchirag/folio-tracker/lib/service/accounts"
	"context"
)

type AccountService interface {
	SignUp(ctx context.Context, in *accounts.SignUpInput) (*domain.User, error)
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
}
