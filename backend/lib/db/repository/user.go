package repository

import (
	"coderchirag/folio-tracker/lib/db"
	// m "coderchirag/folio-tracker/lib/db/folio/model"
	// t "coderchirag/folio-tracker/lib/db/folio/table"
	"coderchirag/folio-tracker/lib/db/models"
	"coderchirag/folio-tracker/lib/db/models/types"
	"context"
	"fmt"
)

const userColumns = `id, name, email, pan, version, created_at, updated_at`

var userUniqueKeys = map[string]uniqueKeyField{
	"uq_users_email": {Field: "email", Reason: models.ErrUserEmailTaken},
	"uq_users_pan":   {Field: "pan", Reason: models.ErrUserPANTaken},
}

type UserRepo struct {
	db  db.DBTX
	now db.Clock
}

func NewUserRepo(db db.DBTX, now db.Clock) *UserRepo {
	return &UserRepo{
		db:  db,
		now: now,
	}
}

// WithTx returns a copy of the repo that runs its queries in tx.
func (r *UserRepo) WithTx(tx db.DBTX) *UserRepo {
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
		return fmt.Errorf("UserRepo.Create: %w", mapWriteError(err, userUniqueKeys))
	}
	u.Version, u.CreatedAt, u.UpdatedAt = 1, now, now
	return nil
}

func (r *UserRepo) GetByID(ctx context.Context, userID types.ID) (*models.User, error) {
	// qb, err := query_builder.New[models.User](r.db, models.UserTable, query_builder.SelectQuery)
	// if err != nil {
	// 	return nil, fmt.Errorf("UserRepo.GetByID: %w", mapError(err))
	// }

	// var u models.User
	// _, err = qb.SelectAll().Where(qb.Cond(
	// 	qb.Col(&qb.Model.ID),
	// 	query_builder.EqualOp,
	// 	userID,
	// )).ExecContext(ctx, &u)
	// if err != nil {
	// 	return nil, fmt.Errorf("UserRepo.GetByID: %w", mapError(err))
	// }

	// q := my.SELECT(
	// 	t.Users.AllColumns,
	// ).FROM(
	// 	t.Users,
	// ).WHERE(
	// 	t.Users.ID.EQ(my.Blob(userID[:])),
	// )
	// if slog.Default().Enabled(ctx, slog.LevelDebug) {
	// 	slog.DebugContext(ctx, "sql query", "query", q.DebugSql())
	// }

	// var dest struct {
	// 	m.Users
	// }
	// err := q.Query(r.db, &dest)
	// if err != nil {
	// 	return nil, fmt.Errorf("Users.Repo.GetByID: sql error: %w", err)
	// }
	// var id types.ID
	// if err := id.Scan(dest.ID); err != nil {
	// 	return nil, fmt.Errorf("Users.Repo.GetByID: parse id error: %w", err)
	// }
	return &models.User{
		// ID:        id,
		// Name:      dest.Name,
		// Email:     dest.Email,
		// PAN:       dest.Pan,
		// CreatedAt: dest.CreatedAt,
		// UpdatedAt: dest.UpdatedAt,
		// Version:   dest.Version,
	}, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, userEmail string) (*models.User, error) {
	var u models.User
	err := r.db.GetContext(ctx, &u,
		`SELECT `+userColumns+` FROM `+models.UserTable+`
			WHERE email = ?`,
		models.NormalizeEmail(userEmail),
	)
	if err != nil {
		return nil, fmt.Errorf("UserRepo.GetByEmail: %w", mapError(err))
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
		return fmt.Errorf("UserRepo.Update: %w", mapWriteError(err, userUniqueKeys))
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("UserRepo.Update: %w", mapError(err))
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
			return fmt.Errorf("UserRepo.Update: %w", mapError(err))
		}
		if !exists {
			return fmt.Errorf("UserRepo.Update: %w", ErrNotFound)
		}
		return fmt.Errorf("UserRepo.Update: %w", ErrConflict)
	}
	u.Version++
	u.UpdatedAt = now
	return nil
}
