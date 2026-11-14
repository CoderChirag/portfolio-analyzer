package domain

import (
	"coderchirag/folio-tracker/lib/apperr"
	"errors"
	"regexp"
	"strings"
	"time"
)

// Limits mirror the instrument_listings table columns.
const listingSymbolMaxLen = 32 // VARCHAR(32) ascii

type Exchange string

const (
	ExchangeNSE Exchange = "NSE"
	ExchangeBSE Exchange = "BSE"
)

func (e Exchange) IsValid() bool {
	switch e {
	case ExchangeBSE, ExchangeNSE:
		return true
	}
	return false
}

var (
	ErrListingInstrumentRequired = errors.New("instrument is required")
	ErrListingAssetTypeInvalid   = errors.New("only stocks and ETFs can have exchange listings")
	ErrListingExchangeInvalid    = errors.New("exchange must be NSE or BSE")
	ErrListingSymbolRequired     = errors.New("symbol is required")
	ErrListingSymbolTooLong      = errors.New("symbol must be at most 32 characters")
	ErrListingSymbolInvalid      = errors.New("symbol may contain only A-Z, 0-9, '-', '&' and '_'")
)

// InstrumentListing is one exchange listing of an exchange-traded instrument.
// BSE listings carry the numeric scrip code as the symbol.
type InstrumentListing struct {
	ID           ID
	InstrumentID ID
	Exchange     Exchange
	Symbol       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type InstrumentListingParams struct {
	Instrument *Instrument
	AssetType  AssetType // the instrument's asset type, from the lookup registry
	Exchange   Exchange
	Symbol     string
}

var symbolPattern = regexp.MustCompile(`^[A-Z0-9&_-]+$`)

// NewInstrumentListing normalizes and validates input and assigns a fresh ID.
// On failure it returns apperr.ValidationError listing every invalid field.
func NewInstrumentListing(p InstrumentListingParams) (*InstrumentListing, error) {
	if p.Instrument == nil {
		var errs apperr.ValidationError
		errs.Add("instrument", ErrListingInstrumentRequired)
		return nil, errs.Err()
	}

	l := &InstrumentListing{
		ID:           NewID(),
		InstrumentID: p.Instrument.ID,
		Exchange:     Exchange(strings.ToUpper(strings.TrimSpace(string(p.Exchange)))),
		Symbol:       strings.ToUpper(strings.TrimSpace(p.Symbol)),
	}
	if err := l.Validate(p.Instrument, p.AssetType); err != nil {
		return nil, err
	}

	return l, nil
}

// Validate mirrors the table's CHECK constraints. The table doesn't store the
// asset type, so the stock/ETF-only rule is enforced here against the instrument.
func (l *InstrumentListing) Validate(inst *Instrument, at AssetType) error {
	var errs apperr.ValidationError

	if l.InstrumentID.IsZero() {
		errs.Add("instrument", ErrListingInstrumentRequired)
	}

	if inst == nil || inst.ID != l.InstrumentID || at.ID != inst.AssetTypeID ||
		(at.Code != AssetStock && at.Code != AssetETF) {
		errs.Add("assetType", ErrListingAssetTypeInvalid)
	}

	if !l.Exchange.IsValid() {
		errs.Add("exchange", ErrListingExchangeInvalid)
	}

	switch {
	case l.Symbol == "":
		errs.Add("symbol", ErrListingSymbolRequired)
	case len(l.Symbol) > listingSymbolMaxLen:
		errs.Add("symbol", ErrListingSymbolTooLong)
	case !symbolPattern.MatchString(l.Symbol):
		errs.Add("symbol", ErrListingSymbolInvalid)
	}

	return errs.Err()
}
