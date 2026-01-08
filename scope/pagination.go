package scope

import "gorm.io/gorm"

// WithPagination 使用分页
func WithPagination(page, size uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		p := page
		s := size

		if p == 0 {
			p = 1
		}
		if s < 5 {
			s = 5
		}
		if s > 100 {
			s = 100
		}

		return db.Offset(int((p - 1) * s)).Limit(int(s))
	}
}
