package domain

import (
	"coderchirag/folio-tracker/lib/lookup"
	"time"
)

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

func (d DataSourceCode) IsValid() bool {
	switch d {
	case SourceGroww:
		return true
	}
	return false
}

var _ lookup.Entry[DataSourceID, DataSourceCode] = DataSource{}

type DataSource struct {
	ID        DataSourceID
	Code      DataSourceCode
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (d DataSource) LookupID() DataSourceID {
	return d.ID
}

func (d DataSource) LookupCode() DataSourceCode {
	return d.Code
}
