package domain

import (
	"coderchirag/folio-tracker/lib/db/models/types"
	"coderchirag/folio-tracker/lib/validation"
	"errors"
	"regexp"
	"strings"
	"time"
)

// Limits mirror the stock_trade_details table columns.
const exchangeOrderIDMaxLen = 32 // VARCHAR(32) ascii

var (
	ErrStockTradeTransactionRequired = errors.New("transaction is required")
	ErrStockTradeAssetTypeInvalid    = errors.New("only stocks and ETFs can have stock trade details")
	ErrStockTradeExchangeInvalid     = errors.New("exchange must be NSE or BSE")
	ErrStockTradeOrderIDRequired     = errors.New("exchange order id is required")
	ErrStockTradeOrderIDTooLong      = errors.New("exchange order id must be at most 32 characters")
	ErrStockTradeOrderIDInvalid      = errors.New("exchange order id may contain only letters and digits")
)

// StockTradeDetails is the 1:1 exchange-trade extension of a Transaction.
// (Exchange, TradeDate, ExchangeOrderID) identifies the order: exchanges
// restart order numbering every trading day.
type StockTradeDetails struct {
	TransactionID   types.ID
	TradeDate       time.Time
	Exchange        Exchange
	ExchangeOrderID string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type StockTradeDetailsParams struct {
	Transaction     *Transaction
	Instrument      *Instrument // the transaction's instrument
	AssetType       AssetType   // the instrument's asset type, from the lookup registry
	Exchange        Exchange
	ExchangeOrderID string
}

var exchangeOrderIDPattern = regexp.MustCompile(`^[A-Za-z0-9]+$`)

// NewStockTradeDetails normalizes and validates input.
// On failure it returns validation.Errors listing every invalid field.
func NewStockTradeDetails(p StockTradeDetailsParams) (*StockTradeDetails, error) {
	if p.Transaction == nil {
		var errs validation.Errors
		errs.Add("transaction", ErrStockTradeTransactionRequired)
		return nil, errs.Err()
	}

	d := &StockTradeDetails{
		TransactionID:   p.Transaction.ID,
		TradeDate:       p.Transaction.TradeDate,
		Exchange:        Exchange(strings.ToUpper(strings.TrimSpace(string(p.Exchange)))),
		ExchangeOrderID: strings.TrimSpace(p.ExchangeOrderID),
	}
	if err := d.Validate(p.Transaction, p.Instrument, p.AssetType); err != nil {
		return nil, err
	}

	return d, nil
}

// Validate mirrors the table's CHECK constraints. The table doesn't store the
// asset type, so the stock/ETF-only rule is enforced here against the instrument.
func (d *StockTradeDetails) Validate(txn *Transaction, inst *Instrument, at AssetType) error {
	var errs validation.Errors

	if txn == nil || d.TransactionID.IsZero() || txn.ID != d.TransactionID || !txn.TradeDate.Equal(d.TradeDate) {
		errs.Add("transaction", ErrStockTradeTransactionRequired)
	}

	if txn == nil || inst == nil || inst.ID != txn.InstrumentID || at.ID != inst.AssetTypeID ||
		(at.Code != AssetStock && at.Code != AssetETF) {
		errs.Add("assetType", ErrStockTradeAssetTypeInvalid)
	}

	if !d.Exchange.IsValid() {
		errs.Add("exchange", ErrStockTradeExchangeInvalid)
	}

	switch {
	case d.ExchangeOrderID == "":
		errs.Add("exchangeOrderId", ErrStockTradeOrderIDRequired)
	case len(d.ExchangeOrderID) > exchangeOrderIDMaxLen:
		errs.Add("exchangeOrderId", ErrStockTradeOrderIDTooLong)
	case !exchangeOrderIDPattern.MatchString(d.ExchangeOrderID):
		errs.Add("exchangeOrderId", ErrStockTradeOrderIDInvalid)
	}

	return errs.Err()
}
