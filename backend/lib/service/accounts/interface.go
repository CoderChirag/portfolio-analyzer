package accounts

import (
	"coderchirag/folio-tracker/lib/db/models"
	"context"
)

type UserStore interface {
	Create(ctx context.Context, u *models.User) error
}
