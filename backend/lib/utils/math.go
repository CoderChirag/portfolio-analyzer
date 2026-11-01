package utils

import "github.com/shopspring/decimal"

// fitsDecimal reports whether d is positive and fits DECIMAL(precision, scale)
// exactly. MySQL would otherwise round extra decimal places silently.
func FitsDecimal(d decimal.Decimal, precision, scale int32) bool {
	if !d.IsPositive() {
		return false
	}
	if !d.Equal(d.Truncate(scale)) {
		return false
	}
	limit := decimal.New(1, precision-scale) // 10^(integer digits)
	return d.LessThan(limit)
}
