package gorm

import (
	"database/sql/driver"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type UUID string

// GormDataType gorm common data type.
func (UUID) GormDataType() string {
	return "string"
}

// GormDBDataType gorm db data type.
func (UUID) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "mysql":
		return "LONGTEXT"
	case "postgres":
		return "UUID"
	case "sqlserver":
		return "NVARCHAR(128)"
	case "sqlite":
		return "TEXT"
	default:
		return ""
	}
}

// Scan is the scanner function for this datatype.
func (u *UUID) Scan(value interface{}) error {
	if value == nil {
		*u = ""
		return nil
	}

	if v, ok := value.(string); ok {
		*u = UUID(v)
	}

	return nil
}

// Value is the valuer function for this datatype.
func (u UUID) Value() (driver.Value, error) {
	if u == "" {
		return nil, nil
	}
	return string(u), nil
}

// String returns the string form of the UUID.
func (u UUID) String() string {
	return string(u)
}

// Equals returns true if string form of UUID matches other, false otherwise.
func (u UUID) Equals(other UUID) bool {
	return u.String() == other.String()
}

// Length returns the number of characters in string form of UUID.
func (u UUID) Length() int {
	return len(u.String())
}

// IsNil returns true if the UUID is a nil UUID (all zeroes), false otherwise.
func (u UUID) IsNil() bool {
	return u.String() == ""
}

// IsEmpty returns true if UUID is nil UUID or of zero length, false otherwise.
func (u UUID) IsEmpty() bool {
	return u.IsNil() || u.Length() == 0
}

// IsNilPtr returns true if caller UUID ptr is nil, false otherwise.
func (u *UUID) IsNilPtr() bool {
	return u == nil
}

// IsEmptyPtr returns true if caller UUID ptr is nil or it's value is empty.
func (u *UUID) IsEmptyPtr() bool {
	return u.IsNilPtr() || u.IsEmpty()
}

// NewUUIDv7 generates a uuid version 7, panics on generation failure.
func NewUUIDv7() UUID {
	id := uuid.Must(uuid.NewV7())
	return UUID(id.String())
}
