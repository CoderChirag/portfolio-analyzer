package handlers

import (
	"coderchirag/folio-tracker/lib/db/models"
	"coderchirag/folio-tracker/lib/service/accounts"
	"context"
)

type AccountService interface {
	SignUp(ctx context.Context, in accounts.SignUpInput) (*models.User, error)
	GetUserByID(ctx context.Context, id string) (*models.User, error)
}
