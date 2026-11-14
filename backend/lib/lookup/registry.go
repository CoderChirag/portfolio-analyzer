package lookup

import (
	"fmt"
	"slices"
)

// Entry is a row of a small, seeded reference table (id + stable code).
type Entry[ID comparable, Code ~string] interface {
	LookupID() ID
	LookupCode() Code
}

// Registry is an immutable in-memory copy of a reference table,
// loaded once at startup and safe for concurrent reads.
type Registry[ID comparable, Code ~string, T Entry[ID, Code]] struct {
	name   string
	byID   map[ID]T
	byCode map[Code]T
	all    []T
}

func New[ID comparable, Code ~string, T Entry[ID, Code]](
	name string, rows []T, required ...Code,
) (*Registry[ID, Code, T], error) {
	r := &Registry[ID, Code, T]{
		name:   name,
		byID:   make(map[ID]T, len(rows)),
		byCode: make(map[Code]T, len(rows)),
		all:    slices.Clone(rows),
	}

	for _, row := range rows {
		r.byID[row.LookupID()] = row
		r.byCode[row.LookupCode()] = row
	}

	for _, code := range required {
		if _, ok := r.byCode[code]; !ok {
			return nil, fmt.Errorf("lookup.New: lookup code %q missing from registry %q", code, name)
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
	return slices.Clone(r.all)
}

// MustID returns the id for a code that Load verified as required.
func (r *Registry[ID, Code, T]) MustID(code Code) ID {
	e, ok := r.byCode[code]
	if !ok {
		panic(fmt.Sprintf("lookup code %q from registry %q not loaded", code, r.name))
	}
	return e.LookupID()
}

// IDs resolves API filter codes; an unknown code is an error, never "no filter".
func (r *Registry[ID, Code, T]) IDs(codes []Code) ([]ID, error) {
	ids := make([]ID, 0, len(codes))
	for _, code := range codes {
		e, ok := r.byCode[code]
		if !ok {
			return nil, fmt.Errorf("unknown code %q for registry %q", code, r.name)
		}
		ids = append(ids, e.LookupID())
	}
	return ids, nil
}
