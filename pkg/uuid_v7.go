package gorm

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// NewUUIDv7 generates a uuid version 7, panics on generation failure.
func NewUUIDv7() datatypes.UUID {
	return datatypes.UUID(uuid.Must(uuid.NewV7()))
}

func ParseUUID(s string) datatypes.UUID {
	return datatypes.UUID(uuid.MustParse(s))
}

func ParseUUIDPtr(s string) *datatypes.UUID {
	u := ParseUUID(s)
	return &u
}
