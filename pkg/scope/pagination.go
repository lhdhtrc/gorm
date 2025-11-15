package scope

import "gorm.io/gorm"

// WithPagination 使用分页
func WithPagination(page, size uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page == 0 {
			page = 1
		}

		if size < 5 {
			size = 5
		}

		if size > 100 {
			size = 100
		}

		return db.Offset(int((page - 1) * size)).Limit(int(size))
	}
}
