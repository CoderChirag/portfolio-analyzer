package datasources

import (
	"coderchirag/folio-tracker/lib/db/models"
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Registry struct {
	byID   map[models.DataSourceID]models.DataSource
	byCode map[models.DataSourceCode]models.DataSource
	all    []models.DataSource
}

func Load(ctx context.Context, db *sqlx.DB) (*Registry, error) {
	var rows []models.DataSource
	if err := db.SelectContext(ctx, &rows,
		`SELECT id, code, name, created_at FROM data_sources ORDER BY id`); err != nil {
		return nil, fmt.Errorf("load data sources: %w", err)
	}

	r := &Registry{
		byID:   make(map[models.DataSourceID]models.DataSource, len(rows)),
		byCode: make(map[models.DataSourceCode]models.DataSource, len(rows)),
		all:    rows,
	}

	for _, row := range rows {
		r.byID[row.ID] = row
		r.byCode[row.Code] = row
	}

	// Fail fast if code expects a source the database doesn't have.
	for _, code := range []models.DataSourceCode{models.SourceGroww} {
		if _, ok := r.byCode[code]; !ok {
			return nil, fmt.Errorf("data source %q missing from data_sources table", code)
		}
	}
	return r, nil
}

func (r *Registry) ByID(id models.DataSourceID) (models.DataSource, bool) {
	ds, ok := r.byID[id]
	return ds, ok
}

func (r *Registry) ByCode(code models.DataSourceCode) (models.DataSource, bool) {
	ds, ok := r.byCode[code]
	return ds, ok
}

// IDs resolves codes from an API filter like ?source=groww,manual.
// Unknown codes are an error, so a typo never silently means "no filter".
func (r *Registry) IDs(codes []models.DataSourceCode) ([]models.DataSourceID, error) {
	ids := make([]models.DataSourceID, 0)
	for _, code := range codes {
		ds, ok := r.byCode[code]
		if !ok {
			return nil, fmt.Errorf("unknown data source %q", code)
		}
		ids = append(ids, ds.ID)
	}
	return ids, nil
}

func (r *Registry) All() []models.DataSource {
	return r.all
}
