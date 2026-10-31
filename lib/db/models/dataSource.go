package models

import "time"

type DataSource struct {
	Code      string    `db:"code" json:"code"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

const (
	SourceGroww = "groww"
)
