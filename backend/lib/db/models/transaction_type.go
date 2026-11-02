package models

import (
	"coderchirag/folio-tracker/lib/lookup"
	"time"
)

type TransactionTypeID uint16

// IsZero reports whether id is the zero value (never assigned).
func (t TransactionTypeID) IsZero() bool {
	return t == 0
}

type TransactionTypeCode string

func (c TransactionTypeCode) IsEmpty() bool {
	return c == ""
}

const (
	TxnBuy  TransactionTypeCode = "BUY"
	TxnSell TransactionTypeCode = "SELL"
)

const TransactionTypeTable = "transaction_types"

var _ lookup.Entry[TransactionTypeID, TransactionTypeCode] = TransactionType{}

type TransactionType struct {
	ID        TransactionTypeID   `db:"id"          json:"-"`
	Code      TransactionTypeCode `db:"code"        json:"code"`
	Name      string              `db:"name"        json:"name"`
	Direction int8                `db:"direction"   json:"direction"` // +1 adds units, -1 removes them
	CreatedAt time.Time           `db:"created_at"  json:"createdAt"`
	UpdatedAt time.Time           `db:"updated_at"  json:"updatedAt"`
}

func (t TransactionType) LookupID() TransactionTypeID {
	return t.ID
}

func (t TransactionType) LookupCode() TransactionTypeCode {
	return t.Code
}
