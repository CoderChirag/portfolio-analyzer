package models

import (
	"coderchirag/folio-tracker/lib/db/models/types"
	"coderchirag/folio-tracker/lib/utils"
	"coderchirag/folio-tracker/lib/validation"
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const InstrumentTable = "instruments"

// Limits mirror the instruments table columns.
const instrumentNameMaxLen = 255 // VARCHAR(255), counted in characters

var (
	ErrInstrumentAssetTypeRequired = errors.New("asset type is required")
	ErrInstrumentNameRequired      = errors.New("name is required")
	ErrInstrumentNameTooLong       = errors.New("name must be at most 255 characters")
	ErrInstrumentNameInvalid       = errors.New("name contains invalid characters")
	ErrInstrumentISINRequired      = errors.New("ISIN is required for stocks and ETFs")
	ErrInstrumentISINInvalid       = errors.New("ISIN must be 2 letters, 9 letters/digits and a check digit")
)

type Instrument struct {
	ID          types.ID    `db:"id"             json:"id"`
	AssetTypeID AssetTypeID `db:"asset_type_id"  json:"-"`
	Name        string      `db:"name"           json:"name"`
	ISIN        *string     `db:"isin"           json:"isin,omitempty"`
	CreatedAt   time.Time   `db:"created_at"     json:"createdAt"`
	UpdatedAt   time.Time   `db:"updated_at"     json:"updatedAt"`
}

// InstrumentParams uses named fields so string arguments can't be swapped by position.
type InstrumentParams struct {
	AssetType AssetType // from the lookup registry; carries both ID and Code
	Name      string
	ISIN      string
}

var isinPattern = regexp.MustCompile(`^[A-Z]{2}[A-Z0-9]{9}[0-9]$`)

// NewInstrument normalizes and validates input and assigns a fresh ID.
// On failure it returns validation.Errors listing every invalid field.
func NewInstrument(p InstrumentParams) (*Instrument, error) {
	i := &Instrument{
		ID:          types.New(),
		AssetTypeID: p.AssetType.ID,
		Name:        strings.TrimSpace(p.Name),
		ISIN:        utils.OptionalUpper(p.ISIN),
	}
	if err := i.Validate(p.AssetType); err != nil {
		return nil, err
	}

	return i, nil
}

// Validate checks every field and reports all problems, not just the first.
// It mirrors the table's CHECK constraints so bad data fails here, not at INSERT.
func (i *Instrument) Validate(at AssetType) error {
	var errs validation.Errors

	if i.AssetTypeID.IsZero() || at.Code.IsEmpty() || at.ID != i.AssetTypeID {
		errs.Add("assetType", ErrInstrumentAssetTypeRequired)
	}

	switch n := utf8.RuneCountInString(i.Name); {
	case n == 0:
		errs.Add("name", ErrInstrumentNameRequired)
	case n > instrumentNameMaxLen:
		errs.Add("name", ErrInstrumentNameTooLong)
	case strings.IndexFunc(i.Name, unicode.IsControl) >= 0:
		errs.Add("name", ErrInstrumentNameInvalid)
	}

	switch {
	case i.ISIN != nil && !isinPattern.MatchString(*i.ISIN):
		errs.Add("isin", ErrInstrumentISINInvalid)
	case i.ISIN == nil && at.Code != "" && at.Code != AssetMF:
		errs.Add("isin", ErrInstrumentISINRequired)
	}

	return errs.Err()
}
