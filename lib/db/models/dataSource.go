package models

import "time"

type DataSourceID uint16

type DataSourceCode string

const (
	SourceGroww DataSourceCode = "groww"
)

type DataSource struct {
	ID        DataSourceID   `db:"id"          json:"-"`
	Code      DataSourceCode `db:"code"        json:"code"`
	Name      string         `db:"name"        json:"name"`
	CreatedAt time.Time      `db:"created_at"  json:"createdAt"`
}
