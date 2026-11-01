package models

import (
	"coderchirag/folio-tracker/lib/id"
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
const (
	instrumentNameMaxLen   = 255 // VARCHAR(255), counted in characters
	instrumentSymbolMaxLen = 32  // VARCHAR(32) ascii
)

type Exchange string

const (
	ExchangeNSE Exchange = "NSE"
	ExchangeBSE Exchange = "BSE"
)

var (
	ErrInstrumentAssetTypeRequired = errors.New("asset type is required")
	ErrInstrumentNameRequired      = errors.New("name is required")
	ErrInstrumentNameTooLong       = errors.New("name must be at most 255 characters")
	ErrInstrumentNameInvalid       = errors.New("name contains invalid characters")
	ErrInstrumentISINRequired      = errors.New("ISIN is required for stocks and ETFs")
	ErrInstrumentISINInvalid       = errors.New("ISIN must be 2 letters, 9 letters/digits and a check digit")
	ErrInstrumentSymbolTooLong     = errors.New("symbol must be at most 32 characters")
	ErrInstrumentSymbolInvalid     = errors.New("symbol may contain only A-Z, 0-9, '-', '&' and '_'")
	ErrInstrumentExchangeInvalid   = errors.New("exchange must be NSE or BSE")
)

type Instrument struct {
	ID          id.ID       `db:"id"             json:"id"`
	AssetTypeID AssetTypeID `db:"asset_type_id"  json:"-"`
	Name        string      `db:"name"           json:"name"`
	ISIN        *string     `db:"isin"           json:"isin,omitempty"`
	Symbol      *string     `db:"symbol"         json:"symbol,omitempty"`
	Exchange    *Exchange   `db:"exchange"       json:"exchange,omitempty"`
	CreatedAt   time.Time   `db:"created_at"     json:"createdAt"`
	UpdatedAt   time.Time   `db:"updated_at"     json:"updatedAt"`
}

// InstrumentParams uses named fields so ISIN and Symbol can't be swapped by position.
type InstrumentParams struct {
	AssetType AssetType // from the lookup registry; carries both ID and Code
	Name      string
	ISIN      string
	Symbol    string
	Exchange  Exchange
}

var (
	isinPattern   = regexp.MustCompile(`^[A-Z]{2}[A-Z0-9]{9}[0-9]$`)
	symbolPattern = regexp.MustCompile(`^[A-Z0-9&_-]+$`)
)

// NewInstrument normalizes and validates input and assigns a fresh ID.
// On failure it returns validation.Errors listing every invalid field.
func NewInstrument(p InstrumentParams) (*Instrument, error) {
	i := &Instrument{
		ID:          id.New(),
		AssetTypeID: p.AssetType.ID,
		Name:        strings.TrimSpace(p.Name),
		ISIN:        utils.OptionalUpper(p.ISIN),
		Symbol:      utils.OptionalUpper(p.Symbol),
		Exchange:    (*Exchange)(utils.OptionalUpper(string(p.Exchange))),
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

	if i.Symbol != nil {
		switch {
		case len(*i.Symbol) > instrumentSymbolMaxLen:
			errs.Add("symbol", ErrInstrumentSymbolTooLong)
		case !symbolPattern.MatchString(*i.Symbol):
			errs.Add("symbol", ErrInstrumentSymbolInvalid)
		}
	}

	if i.Exchange != nil && *i.Exchange != ExchangeNSE && *i.Exchange != ExchangeBSE {
		errs.Add("exchange", ErrInstrumentExchangeInvalid)
	}

	return errs.Err()
}
