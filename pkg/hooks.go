package gorm

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *TableUUID) BeforeCreate(_ *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID, err = uuid.NewV7()
	}
	return
}
