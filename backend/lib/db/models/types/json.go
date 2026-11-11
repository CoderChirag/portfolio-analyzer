package types

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JSON is a document in a MySQL JSON column. Nil or empty means SQL NULL.
type JSON []byte

// Value sends text, not bytes: a []byte parameter reaches MySQL as a binary
// string, which a JSON column can reject ("Cannot create a JSON value from a
// string with CHARACTER SET 'binary'").
func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return string(j), nil
}

func (j *JSON) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		*j = nil
	case []byte:
		*j = bytes.Clone(v) // the driver reuses its buffer after Scan returns
	case string:
		*j = JSON(v)
	default:
		return fmt.Errorf("models.types.JSON: cannot scan %T", src)
	}
	return nil
}

// MarshalJSON embeds the document as-is, so API responses don't double-encode it.
func (j JSON) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return j, nil
}

// IsObject mirrors CHECK (JSON_TYPE(col) = 'OBJECT').
func (j JSON) IsObject() bool {
	b := bytes.TrimSpace(j)
	return len(b) > 0 && b[0] == '{' && json.Valid(b)
}
