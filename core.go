package gormx

// WithLoggerConsole 设置是否将 SQL 日志输出到控制台。
func (config *Config) WithLoggerConsole(state bool) {
	config.loggerConsole = state
}

// WithLoggerHandle 设置日志回调，用于将 SQL 日志写入你的日志系统。
func (config *Config) WithLoggerHandle(handle func(b []byte)) {
	config.loggerHandle = handle
}

// WithAutoMigrate 设置是否在初始化连接后自动迁移表结构。
func (config *Config) WithAutoMigrate(state bool) {
	config.autoMigrate = state
}
