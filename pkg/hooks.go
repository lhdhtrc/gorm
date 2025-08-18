package gorm

import "gorm.io/gorm"

// WithPaging 使用分页
func WithPaging(sql *gorm.DB, page, size uint64) {
	if size < 5 {
		size = 5
	}
	if size > 100 {
		size = 100
	}
	sql.Offset(int((page - 1) * size)).Limit(int(size))
}
