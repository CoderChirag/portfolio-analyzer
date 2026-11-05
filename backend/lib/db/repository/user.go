package repository

import (
	"coderchirag/folio-tracker/lib/db/models"
	"coderchirag/folio-tracker/lib/id"
	"context"
	"fmt"
)

const userColumns = `id, name, email, pan, version, created_at, updated_at`

var userUniqueKeys = map[string]uniqueKeyField{
	"uq_users_email": {Field: "email", Reason: models.ErrUserEmailTaken},
	"uq_users_pan":   {Field: "pan", Reason: models.ErrUserPANTaken},
}

type UserRepo struct {
	db  DBTX
	now Clock
}

func NewUserRepo(db DBTX, now Clock) *UserRepo {
	return &UserRepo{
		db:  db,
		now: now,
	}
}

// WithTx returns a copy of the repo that runs its queries in tx.
func (r *UserRepo) WithTx(tx DBTX) *UserRepo {
	return &UserRepo{
		db:  tx,
		now: r.now,
	}
}

func (r *UserRepo) Create(ctx context.Context, u *models.User) error {
	if err := u.Validate(); err != nil {
		return err
	}
	now := r.now()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO `+models.UserTable+` ( `+userColumns+` )
			VALUES (?, ?, ?, ?, 1, ?, ?)`,
		u.ID, u.Name, u.Email, u.PAN, now, now,
	)
	if err != nil {
		return fmt.Errorf("create user: %w", mapWriteError(err, userUniqueKeys))
	}
	u.Version, u.CreatedAt, u.UpdatedAt = 1, now, now
	return nil
}

func (r *UserRepo) GetByID(ctx context.Context, userID id.ID) (*models.User, error) {
	var u models.User
	err := r.db.GetContext(ctx, &u,
		`SELECT `+userColumns+` FROM `+models.UserTable+`
		 WHERE id = ?`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", mapError(err))
	}
	return &u, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, userEmail string) (*models.User, error) {
	var u models.User
	err := r.db.GetContext(ctx, &u,
		`SELECT `+userColumns+` FROM `+models.UserTable+`
			WHERE email = ?`,
		models.NormalizeEmail(userEmail),
	)
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", mapError(err))
	}
	return &u, nil
}

// Update saves name, email and PAN if nobody changed the row since u was read.
// It returns ErrConflict if the version moved on, ErrNotFound if the row is gone.
func (r *UserRepo) Update(ctx context.Context, u *models.User) error {
	if err := u.Validate(); err != nil {
		return err
	}
	now := r.now()
	res, err := r.db.ExecContext(ctx,
		`UPDATE `+models.UserTable+`
		 SET name = ?, email = ?, pan = ?, version = version + 1, updated_at = ?
		 WHERE id = ? AND version = ?`,
		u.Name, u.Email, u.PAN, now, u.ID, u.Version,
	)
	if err != nil {
		return fmt.Errorf("update user: %w", mapWriteError(err, userUniqueKeys))
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update user: %w", mapError(err))
	}
	if n == 0 {
		// Either the row is gone or its version moved on; tell the two apart.
		var exists bool
		if err := r.db.GetContext(ctx, &exists,
			`SELECT EXISTS(
				SELECT 1 FROM `+models.UserTable+` WHERE id = ?
			)`,
			u.ID,
		); err != nil {
			return fmt.Errorf("update user: %w", mapError(err))
		}
		if !exists {
			return fmt.Errorf("update user: %w", ErrNotFound)
		}
		return fmt.Errorf("update user: %w", ErrConflict)
	}
	u.Version++
	u.UpdatedAt = now
	return nil
}
