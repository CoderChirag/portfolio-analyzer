package models

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

const AssetTypeTable = "asset_types"

var _ lookup.Entry[AssetTypeID, AssetTypeCode] = AssetType{}

type AssetType struct {
	ID        AssetTypeID   `db:"id"          json:"-"`
	Code      AssetTypeCode `db:"code"        json:"code"`
	Name      string        `db:"name"        json:"name"`
	CreatedAt time.Time     `db:"created_at"  json:"createdAt"`
	UpdatedAt time.Time     `db:"updated_at"  json:"updatedAt"`
}

func (a AssetType) LookupID() AssetTypeID {
	return a.ID
}

func (a AssetType) LookupCode() AssetTypeCode {
	return a.Code
}
