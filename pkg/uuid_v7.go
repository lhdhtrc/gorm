package gorm

import (
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
