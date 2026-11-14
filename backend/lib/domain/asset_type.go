package domain

import (
	"coderchirag/folio-tracker/lib/lookup"
	"time"
)

type AssetTypeID uint16

// IsZero reports whether id is the zero value (never assigned).
func (d AssetTypeID) IsZero() bool {
	return d == 0
}

type AssetTypeCode string

func (d AssetTypeCode) IsEmpty() bool {
	return d == ""
}

const (
	AssetStock AssetTypeCode = "STOCK"
	AssetETF   AssetTypeCode = "ETF"
	AssetMF    AssetTypeCode = "MF"
)

func (d AssetTypeCode) IsValid() bool {
	switch d {
	case AssetStock, AssetETF, AssetMF:
		return true
	}
	return false
}

var _ lookup.Entry[AssetTypeID, AssetTypeCode] = AssetType{}

type AssetType struct {
	ID        AssetTypeID
	Code      AssetTypeCode
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (a AssetType) LookupID() AssetTypeID {
	return a.ID
}

func (a AssetType) LookupCode() AssetTypeCode {
	return a.Code
}
