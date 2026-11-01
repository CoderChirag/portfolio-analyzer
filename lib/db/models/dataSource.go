package models

import "time"

type DataSourceID uint16

// IsZero reports whether id is the zero value (never assigned).
func (d DataSourceID) IsZero() bool {
	return d == 0
}

type DataSourceCode string

func (d DataSourceCode) IsEmpty() bool {
	return d == ""
}

const (
	SourceGroww DataSourceCode = "groww"
)

const DataSourceTable = "data_sources"

type DataSource struct {
	ID        DataSourceID   `db:"id"          json:"-"`
	Code      DataSourceCode `db:"code"        json:"code"`
	Name      string         `db:"name"        json:"name"`
	CreatedAt time.Time      `db:"created_at"  json:"createdAt"`
	UpdatedAt time.Time      `db:"updated_at"  json:"updatedAt"`
}

func (d DataSource) LookupID() DataSourceID {
	return d.ID
}

func (d DataSource) LookupCode() DataSourceCode {
	return d.Code
}
