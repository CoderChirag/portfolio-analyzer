package accounts

import (
	"coderchirag/folio-tracker/lib/domain"
	"context"
)

type UserStore interface {
	Create(ctx context.Context, u *domain.User) error
	GetByID(ctx context.Context, userID domain.ID) (*domain.User, error)
	GetByEmail(ctx context.Context, userEmail string) (*domain.User, error)
	Update(ctx context.Context, u *domain.User) error
}
