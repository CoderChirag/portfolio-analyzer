package stocksgroww

import "fmt"

var (
	ErrFileParse       = fmt.Errorf("file parse error")
	ErrRowError        = fmt.Errorf("row parse error")
	ErrMissingHeaders  = fmt.Errorf("missing headers")
	ErrHeadersNotFound = fmt.Errorf("required headers not found in the sheet")
	ErrMissingColumns  = fmt.Errorf("missing columns")
	ErrInvalidValue    = fmt.Errorf("invalid value")
	ErrDecimalParse    = fmt.Errorf("number parse error")
	ErrTimeParse       = fmt.Errorf("time parse error")

	ErrInvalidOrderStatus = fmt.Errorf("invalid order status")
)
