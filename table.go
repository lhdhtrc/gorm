package gormx

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

// Table 为常用的 uint64 主键基类（带软删除）。
type Table struct {
	Id        uint64                `json:"id" gorm:"primarykey"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
	DeletedAt soft_delete.DeletedAt `json:"deleted_at" gorm:"default:0;index"`
}

// TableUUID 为 string 主键基类（使用 UUIDv7）。
type TableUUID struct {
	Id        string                `json:"id" gorm:"type:uuid;primaryKey;"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
	DeletedAt soft_delete.DeletedAt `json:"deleted_at" gorm:"default:0;index"`
}

// BeforeCreate 在插入前生成 UUIDv7 作为主键。
func (t *TableUUID) BeforeCreate(_ *gorm.DB) error {
	t.Id = NewUUIDv7()
	return nil
}
