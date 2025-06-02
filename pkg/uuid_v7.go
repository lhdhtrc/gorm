package gorm

import (
	"database/sql/driver"
	"errors"
	"github.com/google/uuid"
)

type UUID uuid.UUID

func (u *UUID) Scan(value any) (err error) {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("not scan uuid")
	}
	parseByte, err := uuid.FromBytes(bytes)
	*u = UUID(parseByte)
	return
}

func (u UUID) Value() (bytes driver.Value, err error) {
	return uuid.UUID(u).MarshalBinary()
}

func (u UUID) String() string {
	return uuid.UUID(u).String()
}
