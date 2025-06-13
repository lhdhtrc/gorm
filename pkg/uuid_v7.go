package gorm

import (
	"bytes"
	"database/sql/driver"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type BinUUID uuid.UUID

// NewBinUUIDv7 generates a uuid version 7, panics on generation failure.
func NewBinUUIDv7() BinUUID {
	return BinUUID(uuid.Must(uuid.NewRandom()))
}

// GormDataType gorm common data type.
func (BinUUID) GormDataType() string {
	return "BINARY(16)"
}

// GormDBDataType gorm db data type.
func (BinUUID) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "mysql":
		return "BINARY(16)"
	case "postgres":
		return "UUID"
	case "sqlserver":
		return "BINARY(16)"
	case "sqlite":
		return "BLOB"
	default:
		return ""
	}
}

// Scan is the scanner function for this datatype.
func (u *BinUUID) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case []byte:
		id, err := uuid.FromBytes(v)
		if err != nil {
			return fmt.Errorf("invalid UUID bytes: %w", err)
		}
		*u = BinUUID(id)
	case string:
		id, err := uuid.Parse(v)
		if err != nil {
			return fmt.Errorf("invalid UUID string: %w", err)
		}
		*u = BinUUID(id)
	default:
		return fmt.Errorf("unsupported UUID format: %T", value)
	}
	return nil
}

// Value is the valuer function for this datatype.
func (u BinUUID) Value() (driver.Value, error) {
	id := uuid.UUID(u)
	if id != uuid.Nil {
		return id.MarshalBinary()
	}
	return nil, nil
}

// Bytes returns the string form of the UUID.
func (u BinUUID) Bytes() []byte {
	b, err := uuid.UUID(u).MarshalBinary()
	if err != nil {
		return nil
	}
	return b
}

// String returns the string form of the UUID.
func (u BinUUID) String() string {
	id := uuid.UUID(u)
	if id != uuid.Nil {
		return id.String()
	}
	return ""
}

// Equals returns true if bytes form of BinUUID matches other, false otherwise.
func (u BinUUID) Equals(other BinUUID) bool {
	return bytes.Equal(u.Bytes(), other.Bytes())
}

// LengthBytes returns the number of characters in string form of UUID.
func (u BinUUID) LengthBytes() int {
	return len(u.Bytes())
}

// Length returns the number of characters in string form of UUID.
func (u BinUUID) Length() int {
	return len(u.String())
}

// IsNil returns true if the BinUUID is nil uuid (all zeroes), false otherwise.
func (u BinUUID) IsNil() bool {
	return uuid.UUID(u) == uuid.Nil
}

// IsEmpty returns true if BinUUID is nil uuid or of zero length, false otherwise.
func (u BinUUID) IsEmpty() bool {
	return u.IsNil() || u.Length() == 0
}

// IsNilPtr returns true if caller BinUUID ptr is nil, false otherwise.
func (u *BinUUID) IsNilPtr() bool {
	return u == nil
}

// IsEmptyPtr returns true if caller BinUUID ptr is nil or it's value is empty.
func (u *BinUUID) IsEmptyPtr() bool {
	return u.IsNilPtr() || u.IsEmpty()
}
