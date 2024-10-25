package gorm

import "gorm.io/gorm"

func WithPagingFilter(sql *gorm.DB, page, size uint64) {
	if page == 0 {
		page = 1
		if size < 5 {
			size = 5
		}
		if size > 100 {
			size = 100
		}
		sql.Offset(int((page - 1) * size)).Limit(int(size))
	}
}
