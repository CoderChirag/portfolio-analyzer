package accounts

import (
	"coderchirag/folio-tracker/lib/db/models"
	"coderchirag/folio-tracker/lib/db/models/types"
	"context"
)

type UserStore interface {
	Create(ctx context.Context, u *models.User) error
	GetByID(ctx context.Context, userID types.ID) (*models.User, error)
	GetByEmail(ctx context.Context, userEmail string) (*models.User, error)
	Update(ctx context.Context, u *models.User) error
}
