package datasourcesrepo

import (
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
	logPrefix             = "repo:DataSource"
	getAllPrefix          = "GetAll"
	queryContextLogPrefix = "QueryContext"
)

type DataSource struct {
	db db.DBTX
}

func New(db db.DBTX) *DataSource {
	return &DataSource{
		db: db,
	}
}

func (r *DataSource) GetAll(ctx context.Context) ([]domain.DataSource, error) {
	q := mysqldialect.
		SELECT(table.DataSources.AllColumns).
		FROM(table.DataSources).
		ORDER_BY(table.DataSources.ID)
	repository.LogQuery(ctx, fmt.Sprintf("%s.%s", logPrefix, getAllPrefix), q)

	var rows []model.DataSources
	err := q.QueryContext(ctx, r.db, &rows)
	if err != nil {
		return nil, fmt.Errorf("%s.%s: %s: %w", logPrefix, getAllPrefix, queryContextLogPrefix, dberr.MapError(err))
	}

	out := make([]domain.DataSource, len(rows))
	for i, r := range rows {
		out[i] = *ToDataSourceDomain(&r)
	}

	return out, nil
}
