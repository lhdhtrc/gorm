package gormx

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

// Table 为常用的 uint64 主键基类（带软删除）。
type Table struct {
	Id        uint64                `json:"id" gorm:"primarykey"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
	DeletedAt soft_delete.DeletedAt `json:"deleted_at" gorm:"default:0;index"`
}

// TableUUID 为 string 主键基类（使用 UUIDv7，数据库自动生成）。
type TableUUID struct {
	Id        string                `json:"id" gorm:"type:uuid;primaryKey;default:uuidv7()"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
	DeletedAt soft_delete.DeletedAt `json:"deleted_at" gorm:"default:0;index"`
}
