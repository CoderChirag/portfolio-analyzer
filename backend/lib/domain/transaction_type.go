package domain

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

func (c TransactionTypeCode) IsValid() bool {
	switch c {
	case TxnBuy, TxnSell:
		return true
	}
	return false
}

var _ lookup.Entry[TransactionTypeID, TransactionTypeCode] = TransactionType{}

type TransactionType struct {
	ID        TransactionTypeID
	Code      TransactionTypeCode
	Name      string
	Direction int8
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (t TransactionType) LookupID() TransactionTypeID {
	return t.ID
}

func (t TransactionType) LookupCode() TransactionTypeCode {
	return t.Code
}
