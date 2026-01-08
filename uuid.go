package gormx

import "github.com/google/uuid"

// NewUUIDv7 generates a uuid version 7, panics on generation failure.
func NewUUIDv7() string {
	return uuid.Must(uuid.NewV7()).String()
}
