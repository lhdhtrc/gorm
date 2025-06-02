package gorm

import (
	"database/sql/driver"
	"fmt"
	"github.com/google/uuid"
)

type UUID uuid.UUID

func (u *UUID) Scan(value any) error {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case []byte:
		id, err := uuid.FromBytes(v)
		if err != nil {
			return fmt.Errorf("invalid UUID bytes: %w", err)
		}
		*u = UUID(id)
	case string:
		id, err := uuid.Parse(v)
		if err != nil {
			return fmt.Errorf("invalid UUID string: %w", err)
		}
		*u = UUID(id)
	default:
		return fmt.Errorf("unsupported UUID format: %T", value)
	}

	return nil
}

func (u UUID) Value() (bytes driver.Value, err error) {
	id := uuid.UUID(u)
	if id != uuid.Nil {
		return id.MarshalBinary(), nil
	}
	return nil, nil
}

func (u UUID) String() string {
	id := uuid.UUID(u)
	if id != uuid.Nil {
		return id.String()
	}
	return ""
}
