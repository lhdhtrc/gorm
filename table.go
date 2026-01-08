package gormx

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

// Table 为常用的 uint64 主键基类（带软删除）。
type Table struct {
	ID        uint64                `json:"id" gorm:"primarykey"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
	DeletedAt soft_delete.DeletedAt `json:"deleted_at" gorm:"default:0;index"`
}

// TableUnique 为带 uniqueIndex 的基类（以 DeletedAt 实现“软删除下的唯一性”）。
type TableUnique struct {
	ID        uint64                `json:"id" gorm:"primaryKey;"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
	DeletedAt soft_delete.DeletedAt `json:"deleted_at" gorm:"default:0;uniqueIndex:idx_unique"`
}

// TableUUID 为 string 主键基类（使用 UUIDv7）。
type TableUUID struct {
	ID        string                `json:"id" gorm:"type:uuid;primaryKey;"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
	DeletedAt soft_delete.DeletedAt `json:"deleted_at" gorm:"default:0;index"`
}

// BeforeCreate 在插入前生成 UUIDv7 作为主键。
func (t *TableUUID) BeforeCreate(_ *gorm.DB) error {
	t.ID = NewUUIDv7()
	return nil
}

// TableUUIDUnique 为带 uniqueIndex 的 UUID 主键基类。
type TableUUIDUnique struct {
	ID        string                `json:"id" gorm:"type:uuid;primaryKey;"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
	DeletedAt soft_delete.DeletedAt `json:"deleted_at" gorm:"default:0;uniqueIndex:idx_unique"`
}

// BeforeCreate 在插入前生成 UUIDv7 作为主键。
func (t *TableUUIDUnique) BeforeCreate(_ *gorm.DB) error {
	t.ID = NewUUIDv7()
	return nil
}
