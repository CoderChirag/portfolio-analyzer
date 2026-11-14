package domain

import (
	"coderchirag/folio-tracker/lib/db/models/types"
	"coderchirag/folio-tracker/lib/utils"
	"coderchirag/folio-tracker/lib/validation"
	"errors"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// Limits mirror the instrument_aliases table columns.
const (
	aliasMaxLen           = 255 // alias VARCHAR(255), counted in characters
	maxNormalizedAliasLen = 255 // normalized_alias VARCHAR(255) ascii
)

var (
	ErrAliasSourceRequired     = errors.New("data source is required")
	ErrAliasRequired           = errors.New("alias is required")
	ErrAliasTooLong            = errors.New("alias must be at most 255 characters")
	ErrAliasInvalid            = errors.New("alias contains invalid characters")
	ErrAliasNotIdentifiable    = errors.New("alias has no identifying words")
	ErrAliasInstrumentRequired = errors.New("instrument is required")
)

type InstrumentAlias struct {
	ID              types.ID
	DataSourceID    DataSourceID
	Alias           string
	NormalizedAlias string
	InstrumentID    types.ID
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type InstrumentAliasParams struct {
	Source       DataSource
	Alias        string
	InstrumentID types.ID
}

func NewInstrumentAlias(p InstrumentAliasParams) (*InstrumentAlias, error) {
	canonical := strings.TrimSpace(p.Alias)
	a := &InstrumentAlias{
		ID:              types.NewID(),
		DataSourceID:    p.Source.ID,
		Alias:           canonical,
		NormalizedAlias: utils.NormalizeFundName(canonical, maxNormalizedAliasLen),
		InstrumentID:    p.InstrumentID,
	}
	if err := a.Validate(); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *InstrumentAlias) Validate() error {
	var errs validation.Errors

	if a.DataSourceID.IsZero() {
		errs.Add("dataSource", ErrAliasSourceRequired)
	}

	switch n := utf8.RuneCountInString(a.Alias); {
	case n == 0:
		errs.Add("alias", ErrAliasRequired)
	case n > aliasMaxLen:
		errs.Add("alias", ErrAliasTooLong)
	case strings.IndexFunc(a.Alias, unicode.IsControl) >= 0:
		errs.Add("alias", ErrAliasInvalid)
	case a.NormalizedAlias == "":
		errs.Add("alias", ErrAliasNotIdentifiable)
	}

	if a.InstrumentID.IsZero() {
		errs.Add("instrumentId", ErrAliasInstrumentRequired)
	}

	return errs.Err()
}
