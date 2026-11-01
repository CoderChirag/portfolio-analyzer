package lookup

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// Entry is a row of a small, seeded reference table (id + stable code).
type Entry[ID comparable, Code ~string] interface {
	LookupID() ID
	LookupCode() Code
}

// Registry is an immutable in-memory copy of a reference table,
// loaded once at startup and safe for concurrent reads.
type Registry[ID comparable, Code ~string, T Entry[ID, Code]] struct {
	byID      map[ID]T
	byCode    map[Code]T
	all       []T
	tableName string
}

// Load reads every row with query and fails if any required code is missing.
func Load[ID comparable, Code ~string, T Entry[ID, Code]](
	ctx context.Context, db *sqlx.DB, tableName string, required ...Code,
) (*Registry[ID, Code, T], error) {
	var rows []T
	// SELECT * so table-specific columns (e.g. transaction_types.direction) load too.
	// Unsafe ignores columns the struct doesn't have yet, so an expand migration that
	// adds a column doesn't break replicas still running the previous build.
	if err := db.Unsafe().SelectContext(ctx, &rows,
		fmt.Sprintf(`SELECT * FROM %s ORDER BY id`, tableName),
	); err != nil {
		return nil, fmt.Errorf("load lookup: %w", err)
	}

	r := &Registry[ID, Code, T]{
		byID:      make(map[ID]T, len(rows)),
		byCode:    make(map[Code]T, len(rows)),
		all:       rows,
		tableName: tableName,
	}

	for _, row := range rows {
		r.byID[row.LookupID()] = row
		r.byCode[row.LookupCode()] = row
	}

	for _, code := range required {
		if _, ok := r.byCode[code]; !ok {
			return nil, fmt.Errorf("lookup code %q missing from database(table: %q)", code, tableName)
		}
	}
	return r, nil
}

func (r *Registry[ID, Code, T]) ByID(id ID) (T, bool) {
	e, ok := r.byID[id]
	return e, ok
}

func (r *Registry[ID, Code, T]) ByCode(code Code) (T, bool) {
	e, ok := r.byCode[code]
	return e, ok
}

func (r *Registry[ID, Code, T]) All() []T {
	return r.all
}

// MustID returns the id for a code that Load verified as required.
func (r *Registry[ID, Code, T]) MustID(code Code) ID {
	e, ok := r.byCode[code]
	if !ok {
		panic(fmt.Sprintf("lookup code %q from table %q not loaded", code, r.tableName))
	}
	return e.LookupID()
}

// IDs resolves API filter codes; an unknown code is an error, never "no filter".
func (r *Registry[ID, Code, T]) IDs(codes []Code) ([]ID, error) {
	ids := make([]ID, 0, len(codes))
	for _, code := range codes {
		e, ok := r.byCode[code]
		if !ok {
			return nil, fmt.Errorf("unknown code %q for %q", code, r.tableName)
		}
		ids = append(ids, e.LookupID())
	}
	return ids, nil
}
