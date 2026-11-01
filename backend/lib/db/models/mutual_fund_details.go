package models

import (
	"coderchirag/folio-tracker/lib/id"
	"coderchirag/folio-tracker/lib/utils"
	"coderchirag/folio-tracker/lib/validation"
	"errors"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const MutualFundDetailsTable = "mutual_fund_details"

// Limits mirror the mutual_fund_details table columns.
const mfAMCMaxLen = 128 // VARCHAR(128), counted in characters

var (
	ErrMFDetailsInstrumentRequired = errors.New("instrument is required")
	ErrMFDetailsAssetTypeInvalid   = errors.New("only mutual funds can have mutual fund details")
	ErrMFDetailsAMCTooLong         = errors.New("AMC must be at most 128 characters")
	ErrMFDetailsAMCInvalid         = errors.New("AMC contains invalid characters")
)

// MutualFundDetails is the 1:1 mutual-fund extension of an Instrument.
// AMC is optional because sources reveal it unevenly
// (Groww orders CSV has none; the holdings CSV has it).
type MutualFundDetails struct {
	InstrumentID id.ID     `db:"instrument_id"  json:"instrumentId"`
	AMC          *string   `db:"amc"            json:"amc,omitempty"`
	CreatedAt    time.Time `db:"created_at"     json:"createdAt"`
	UpdatedAt    time.Time `db:"updated_at"     json:"updatedAt"`
}

type MutualFundDetailsParams struct {
	Instrument *Instrument
	AssetType  AssetType // the instrument's asset type, from the lookup registry
	AMC        string
}

// NewMutualFundDetails normalizes and validates input.
// On failure it returns validation.Errors listing every invalid field.
func NewMutualFundDetails(p MutualFundDetailsParams) (*MutualFundDetails, error) {
	if p.Instrument == nil {
		var errs validation.Errors
		errs.Add("instrument", ErrMFDetailsInstrumentRequired)
		return nil, errs.Err()
	}

	d := &MutualFundDetails{
		InstrumentID: p.Instrument.ID,
		AMC:          utils.OptionalTrim(p.AMC),
	}
	if err := d.Validate(p.Instrument, p.AssetType); err != nil {
		return nil, err
	}

	return d, nil
}

// Validate mirrors the table's constraints. The table doesn't store the
// asset type, so the MF-only rule is enforced here against the instrument.
func (d *MutualFundDetails) Validate(inst *Instrument, at AssetType) error {
	var errs validation.Errors

	if d.InstrumentID.IsZero() {
		errs.Add("instrument", ErrMFDetailsInstrumentRequired)
	}

	if inst == nil || inst.ID != d.InstrumentID || at.ID != inst.AssetTypeID || at.Code != AssetMF {
		errs.Add("assetType", ErrMFDetailsAssetTypeInvalid)
	}

	if d.AMC != nil {
		switch {
		case utf8.RuneCountInString(*d.AMC) > mfAMCMaxLen:
			errs.Add("amc", ErrMFDetailsAMCTooLong)
		case strings.IndexFunc(*d.AMC, unicode.IsControl) >= 0:
			errs.Add("amc", ErrMFDetailsAMCInvalid)
		}
	}

	return errs.Err()
}
