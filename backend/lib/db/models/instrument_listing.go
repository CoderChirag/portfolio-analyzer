package models

import (
	"coderchirag/folio-tracker/lib/id"
	"coderchirag/folio-tracker/lib/utils"
	"coderchirag/folio-tracker/lib/validation"
	"errors"
	"regexp"
	"time"
)

const InstrumentListingTable = "instrument_listings"

// Limits mirror the instrument_listings table columns.
const listingSymbolMaxLen = 32 // VARCHAR(32) ascii

type Exchange string

const (
	ExchangeNSE Exchange = "NSE"
	ExchangeBSE Exchange = "BSE"
)

func (e Exchange) IsValid() bool {
	return e == ExchangeNSE || e == ExchangeBSE
}

var (
	ErrListingInstrumentRequired = errors.New("instrument is required")
	ErrListingAssetTypeInvalid   = errors.New("only stocks and ETFs can have exchange listings")
	ErrListingExchangeInvalid    = errors.New("exchange must be NSE or BSE")
	ErrListingSymbolTooLong      = errors.New("symbol must be at most 32 characters")
	ErrListingSymbolInvalid      = errors.New("symbol may contain only A-Z, 0-9, '-', '&' and '_'")
)

// InstrumentListing is one exchange listing of an exchange-traded instrument.
// BSE listings carry the numeric scrip code as the symbol.
type InstrumentListing struct {
	ID           id.ID     `db:"id"             json:"id"`
	InstrumentID id.ID     `db:"instrument_id"  json:"instrumentId"`
	Exchange     *Exchange `db:"exchange"       json:"exchange,omitempty"`
	Symbol       *string   `db:"symbol"         json:"symbol,omitempty"`
	CreatedAt    time.Time `db:"created_at"     json:"createdAt"`
	UpdatedAt    time.Time `db:"updated_at"     json:"updatedAt"`
}

type InstrumentListingParams struct {
	Instrument *Instrument
	AssetType  AssetType // the instrument's asset type, from the lookup registry
	Exchange   Exchange
	Symbol     string
}

var symbolPattern = regexp.MustCompile(`^[A-Z0-9&_-]+$`)

// NewInstrumentListing normalizes and validates input and assigns a fresh ID.
// On failure it returns validation.Errors listing every invalid field.
func NewInstrumentListing(p InstrumentListingParams) (*InstrumentListing, error) {
	if p.Instrument == nil {
		var errs validation.Errors
		errs.Add("instrument", ErrListingInstrumentRequired)
		return nil, errs.Err()
	}

	l := &InstrumentListing{
		ID:           id.New(),
		InstrumentID: p.Instrument.ID,
		Exchange:     (*Exchange)(utils.OptionalUpper(string(p.Exchange))),
		Symbol:       utils.OptionalUpper(p.Symbol),
	}
	if err := l.Validate(p.Instrument, p.AssetType); err != nil {
		return nil, err
	}

	return l, nil
}

// Validate mirrors the table's CHECK constraints. The table doesn't store the
// asset type, so the stock/ETF-only rule is enforced here against the instrument.
func (l *InstrumentListing) Validate(inst *Instrument, at AssetType) error {
	var errs validation.Errors

	if l.InstrumentID.IsZero() {
		errs.Add("instrument", ErrListingInstrumentRequired)
	}

	if inst == nil || inst.ID != l.InstrumentID || at.ID != inst.AssetTypeID ||
		(at.Code != AssetStock && at.Code != AssetETF) {
		errs.Add("assetType", ErrListingAssetTypeInvalid)
	}

	if l.Exchange != nil && !l.Exchange.IsValid() {
		errs.Add("exchange", ErrListingExchangeInvalid)
	}

	if l.Symbol != nil {
		switch {
		case len(*l.Symbol) > listingSymbolMaxLen:
			errs.Add("symbol", ErrListingSymbolTooLong)
		case !symbolPattern.MatchString(*l.Symbol):
			errs.Add("symbol", ErrListingSymbolInvalid)
		}
	}

	return errs.Err()
}
