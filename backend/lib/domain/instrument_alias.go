package domain

import (
	"coderchirag/folio-tracker/lib/apperr"
	"coderchirag/folio-tracker/lib/utils"
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
	ID              ID
	DataSourceID    DataSourceID
	Alias           string
	NormalizedAlias string
	InstrumentID    ID
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type InstrumentAliasParams struct {
	Source       DataSource
	Alias        string
	InstrumentID ID
}

func NewInstrumentAlias(p *InstrumentAliasParams) (*InstrumentAlias, error) {
	canonical := strings.TrimSpace(p.Alias)
	a := &InstrumentAlias{
		ID:              NewID(),
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
	var errs apperr.ValidationError

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
