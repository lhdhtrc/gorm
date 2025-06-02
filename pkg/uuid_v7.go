package gorm

import (
	"database/sql/driver"
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type UUID uuid.UUID

func (u *UUID) GormDataType() string {
	return "binary(16)"
}

func (u *UUID) GormDBDataType(_ *gorm.DB, _ *schema.Field) string {
	return "binary"
}

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
