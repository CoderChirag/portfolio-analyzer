package models

import (
	"coderchirag/folio-tracker/lib/db/models/types"
	"coderchirag/folio-tracker/lib/utils"
	"coderchirag/folio-tracker/lib/validation"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

const TransactionTable = "transactions"

// Limits mirror the transactions table's DECIMAL(precision, scale) columns.
const (
	txnQuantityPrecision, txnQuantityScale = 20, 6
	txnAmountPrecision, txnAmountScale     = 20, 4
	txnPricePrecision, txnPriceScale       = 20, 6
)

const DefaultCurrency = "INR"

// marketLocation is the exchange calendar's time zone; trade_date is a date in it.
var marketLocation = time.FixedZone("IST", 5*60*60+30*60)

var (
	ErrTxnUserRequired       = errors.New("user is required")
	ErrTxnSourceRequired     = errors.New("data source is required")
	ErrTxnInstrumentRequired = errors.New("instrument is required")
	ErrTxnTypeRequired       = errors.New("transaction type is required")
	ErrTxnTradeDateRequired  = errors.New("trade date is required")
	ErrTxnTradeDateMismatch  = errors.New("trade date must be the Indian calendar date of the execution time")
	ErrTxnQuantityInvalid    = errors.New("quantity must be positive with at most 14 integer and 6 decimal digits")
	ErrTxnAmountInvalid      = errors.New("amount must be positive with at most 16 integer and 4 decimal digits")
	ErrTxnPriceInvalid       = errors.New("price must be positive with at most 14 integer and 6 decimal digits")
	ErrTxnCurrencyInvalid    = errors.New("currency must be a 3-letter ISO 4217 code")
)

// Transaction is one row of the core ledger, common to every asset class.
type Transaction struct {
	ID                types.ID          `db:"id"                   json:"id"`
	UserID            types.ID          `db:"user_id"              json:"userId"`
	DataSourceID      DataSourceID      `db:"data_source_id"       json:"-"`
	InstrumentID      types.ID          `db:"instrument_id"        json:"instrumentId"`
	TransactionTypeID TransactionTypeID `db:"transaction_type_id"  json:"-"`
	TradeDate         time.Time         `db:"trade_date"           json:"tradeDate"` // midnight UTC of the Indian calendar date
	ExecutedAt        *time.Time        `db:"executed_at"          json:"executedAt,omitempty"`
	Quantity          decimal.Decimal   `db:"quantity"             json:"quantity"`
	Amount            decimal.Decimal   `db:"amount"               json:"amount"`
	Price             *decimal.Decimal  `db:"price"                json:"price,omitempty"`
	Currency          string            `db:"currency"             json:"currency"`
	CreatedAt         time.Time         `db:"created_at"           json:"createdAt"`
	UpdatedAt         time.Time         `db:"updated_at"           json:"updatedAt"`
}

// TransactionParams uses named fields so same-typed values can't be swapped by position.
type TransactionParams struct {
	UserID          types.ID
	DataSourceID    DataSourceID
	Instrument      *Instrument
	TransactionType TransactionType // from the lookup registry
	TradeDate       time.Time       // only the year, month and day are used
	ExecutedAt      *time.Time      // any time zone; stored as UTC
	Quantity        decimal.Decimal
	Amount          decimal.Decimal
	Price           *decimal.Decimal
	Currency        string // empty means DefaultCurrency
}

var currencyPattern = regexp.MustCompile(`^[A-Z]{3}$`)

// NewTransaction normalizes and validates input and assigns a fresh ID.
// On failure it returns validation.Errors listing every invalid field.
func NewTransaction(p TransactionParams) (*Transaction, error) {
	if p.Instrument == nil {
		var errs validation.Errors
		errs.Add("instrument", ErrTxnInstrumentRequired)
		return nil, errs.Err()
	}

	t := &Transaction{
		ID:                types.NewID(),
		UserID:            p.UserID,
		DataSourceID:      p.DataSourceID,
		InstrumentID:      p.Instrument.ID,
		TransactionTypeID: p.TransactionType.ID,
		TradeDate:         utils.DateOnly(p.TradeDate),
		Quantity:          p.Quantity,
		Amount:            p.Amount,
		Price:             p.Price,
		Currency:          strings.ToUpper(strings.TrimSpace(p.Currency)),
	}
	if t.Currency == "" {
		t.Currency = DefaultCurrency
	}
	if p.ExecutedAt != nil {
		// DATETIME(3) keeps milliseconds; truncate so the struct matches what is stored.
		at := p.ExecutedAt.UTC().Truncate(time.Millisecond)
		t.ExecutedAt = &at
	}
	if err := t.Validate(p.TransactionType); err != nil {
		return nil, err
	}

	return t, nil
}

// Validate mirrors the table's CHECK constraints and column limits.
func (t *Transaction) Validate(tt TransactionType) error {
	var errs validation.Errors

	if t.UserID.IsZero() {
		errs.Add("user", ErrTxnUserRequired)
	}

	if t.DataSourceID.IsZero() {
		errs.Add("dataSource", ErrTxnSourceRequired)
	}

	if t.InstrumentID.IsZero() {
		errs.Add("instrument", ErrTxnInstrumentRequired)
	}

	if t.TransactionTypeID.IsZero() || tt.Code.IsEmpty() || tt.ID != t.TransactionTypeID {
		errs.Add("transactionType", ErrTxnTypeRequired)
	}

	switch {
	case t.TradeDate.IsZero():
		errs.Add("tradeDate", ErrTxnTradeDateRequired)
	case t.ExecutedAt != nil && !utils.DateOnly(t.ExecutedAt.In(marketLocation)).Equal(t.TradeDate):
		errs.Add("tradeDate", ErrTxnTradeDateMismatch)
	}

	if !utils.FitsDecimal(t.Quantity, txnQuantityPrecision, txnQuantityScale) {
		errs.Add("quantity", ErrTxnQuantityInvalid)
	}

	if !utils.FitsDecimal(t.Amount, txnAmountPrecision, txnAmountScale) {
		errs.Add("amount", ErrTxnAmountInvalid)
	}

	if t.Price != nil && !utils.FitsDecimal(*t.Price, txnPricePrecision, txnPriceScale) {
		errs.Add("price", ErrTxnPriceInvalid)
	}

	if !currencyPattern.MatchString(t.Currency) {
		errs.Add("currency", ErrTxnCurrencyInvalid)
	}

	return errs.Err()
}
