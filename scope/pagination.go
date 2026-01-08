package scope

import "gorm.io/gorm"

// WithPagination 使用分页
func WithPagination(page, size uint64) func(db *gorm.DB) *gorm.DB {
	// 返回一个可被 db.Scopes(...) 组合调用的 scope 函数。
	return func(db *gorm.DB) *gorm.DB {
		// 拷贝一份入参，避免闭包内修改影响外部变量的可读性。
		p := page
		// 拷贝一份入参，便于在闭包内做边界修正。
		s := size

		// page 从 1 开始计数。
		if p == 0 {
			p = 1
		}
		// size 太小则兜底到 5。
		if s < 5 {
			s = 5
		}
		// size 太大则限制到 100，避免一次拉取过多。
		if s > 100 {
			s = 100
		}

		// 将页码与页大小转换为 offset/limit。
		return db.Offset(int((p - 1) * s)).Limit(int(s))
	}
}
