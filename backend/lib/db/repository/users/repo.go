package users

import (
	"coderchirag/folio-tracker/lib/apperr"
	"coderchirag/folio-tracker/lib/db"
	"coderchirag/folio-tracker/lib/db/repository"
	"coderchirag/folio-tracker/lib/db/repository/internal/dberr"
	"coderchirag/folio-tracker/lib/db/repository/internal/jetgen/model"
	"coderchirag/folio-tracker/lib/db/repository/internal/jetgen/table"
	"coderchirag/folio-tracker/lib/domain"
	"context"
	"fmt"

	mysqldialect "github.com/go-jet/jet/v2/mysql"
)

const (
	logPrefix             = "repo:User"
	createLogPrefix       = "Create"
	getByIDLogPrefix      = "GetByID"
	getByEmailLogPrefix   = "GetByEmail"
	updateLogPrefix       = "Update"
	execContextLogPrefix  = "ExecContext"
	queryContextLogPrefix = "QueryContext"
)

var userUniqueKeys = map[string]dberr.UniqueKeyField{
	"uq_users_email": {Field: "email", Reason: domain.ErrUserEmailTaken},
	"uq_users_pan":   {Field: "pan", Reason: domain.ErrUserPANTaken},
}

type User struct {
	db  db.DBTX
	now db.Clock
}

func New(db db.DBTX, now db.Clock) *User {
	return &User{
		db:  db,
		now: now,
	}
}

// WithTx returns a copy of the repo that runs its queries in tx.
func (r *User) WithTx(tx db.DBTX) *User {
	return &User{
		db:  tx,
		now: r.now,
	}
}

func (r *User) Create(ctx context.Context, u *domain.User) error {
	if err := u.Validate(); err != nil {
		return err
	}

	now := r.now()
	row := ToUserModel(u)
	row.Version, row.CreatedAt, row.UpdatedAt = 1, now, now

	q := table.Users.INSERT(table.Users.AllColumns).MODEL(row)
	repository.LogQuery(ctx, fmt.Sprintf("%s.%s", logPrefix, createLogPrefix), q)

	if _, err := q.ExecContext(ctx, r.db); err != nil {
		return fmt.Errorf("%s.%s: %s: %w", logPrefix, createLogPrefix, execContextLogPrefix, dberr.MapWriteError(err, userUniqueKeys))
	}

	u.Version, u.CreatedAt, u.UpdatedAt = 1, now, now
	return nil
}

func (r *User) GetByID(ctx context.Context, userID domain.ID) (*domain.User, error) {
	q := table.Users.
		SELECT(table.Users.AllColumns).
		FROM(table.Users).
		WHERE(table.Users.ID.EQ(mysqldialect.Blob(userID.Binary())))
	repository.LogQuery(ctx, fmt.Sprintf("%s.%s", logPrefix, getByIDLogPrefix), q)

	var out model.Users
	err := q.QueryContext(ctx, r.db, &out)
	if err != nil {
		return nil, fmt.Errorf("%s.%s: %s: %w", logPrefix, getByIDLogPrefix, queryContextLogPrefix, dberr.MapError(err))
	}

	return ToUserDomain(&out), nil
}

func (r *User) GetByEmail(ctx context.Context, userEmail string) (*domain.User, error) {
	q := table.Users.
		SELECT(table.Users.AllColumns).
		FROM(table.Users).
		WHERE(table.Users.Email.EQ(mysqldialect.String(domain.NormalizeEmail(userEmail))))
	repository.LogQuery(ctx, fmt.Sprintf("%s.%s", logPrefix, getByEmailLogPrefix), q)

	var out model.Users
	err := q.QueryContext(ctx, r.db, &out)
	if err != nil {
		return nil, fmt.Errorf("%s.%s: %s: %w", logPrefix, getByEmailLogPrefix, queryContextLogPrefix, dberr.MapError(err))
	}

	return ToUserDomain(&out), nil
}

// Update saves name, email and PAN if nobody changed the row since u was read.
// It returns ErrConflict if the version moved on, ErrNotFound if the row is gone.
func (r *User) Update(ctx context.Context, u *domain.User) error {
	if err := u.Validate(); err != nil {
		return err
	}

	row := ToUserModel(u)
	now := r.now()
	row.UpdatedAt = now
	row.Version = u.Version + 1
	q := table.Users.
		UPDATE(
			table.Users.Name,
			table.Users.Email,
			table.Users.Pan,
			table.Users.Version,
			table.Users.UpdatedAt,
		).
		MODEL(row).
		WHERE(
			table.Users.ID.EQ(mysqldialect.Blob(row.ID.Binary())).
				AND(table.Users.Version.EQ(mysqldialect.Uint32(u.Version))),
		)
	repository.LogQuery(ctx, fmt.Sprintf("%s.%s", logPrefix, updateLogPrefix), q)

	res, err := q.ExecContext(ctx, r.db)
	if err != nil {
		return fmt.Errorf("%s.%s: %s: %w", logPrefix, updateLogPrefix, execContextLogPrefix, dberr.MapWriteError(err, userUniqueKeys))
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s.%s: RowsAffected: %w", logPrefix, updateLogPrefix, dberr.MapError(err))
	}
	if n == 0 {
		// Either the row is gone or its version moved on; tell the two apart.
		var cur model.Users
		q := table.Users.
			SELECT(table.Users.Version).
			FROM(table.Users).
			WHERE(table.Users.ID.EQ(mysqldialect.Blob(row.ID.Binary())))
		repository.LogQuery(ctx, fmt.Sprintf("%s.%s", logPrefix, updateLogPrefix), q)

		if err := q.QueryContext(ctx, r.db, &cur); err != nil {
			return fmt.Errorf("%s.%s: %s: %w", logPrefix, updateLogPrefix, queryContextLogPrefix, dberr.MapError(err)) // ErrNotFound if gone
		}
		return fmt.Errorf("%s.%s: %w", logPrefix, updateLogPrefix, apperr.ErrConflict)
	}

	u.Version++
	u.UpdatedAt = now
	return nil
}
