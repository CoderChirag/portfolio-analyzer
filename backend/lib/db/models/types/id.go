package types

import (
	"database/sql/driver"
	"fmt"

	"github.com/google/uuid"
)

// ID is a UUIDv7 stored in MySQL as BINARY(16)
type ID uuid.UUID

func NewID() ID {
	return ID(uuid.Must(uuid.NewV7()))
}

func ParseID(s string) (ID, error) {
	id, err := uuid.Parse(s)
	return ID(id), err
}

func (id ID) String() string {
	return uuid.UUID(id).String()
}

// Value writes the 16 raw bytes, not the 36-char text form
func (id ID) Value() (driver.Value, error) {
	return id[:], nil
}

func (id *ID) Scan(src any) error {
	b, ok := src.([]byte)
	if !ok || len(b) != 16 {
		return fmt.Errorf("models.ID: cannot scan %T", src)
	}
	copy(id[:], b)
	return nil
}

func (id ID) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

func (id *ID) UnmarshalText(b []byte) error {
	parsed, err := ParseID(string(b))
	if err != nil {
		return err
	}
	*id = ID(parsed)
	return nil
}

// IsZero reports whether id is the zero value (never assigned).
func (id ID) IsZero() bool { return id == ID{} }
