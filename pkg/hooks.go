package gorm

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *TableUUID) BeforeCreate(_ *gorm.DB) (err error) {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return
}
