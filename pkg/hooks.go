package gorm

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *TableUUIDEntity) BeforeCreate(_ *gorm.DB) (err error) {
	s.ID = uuid.New().String()
	return
}
