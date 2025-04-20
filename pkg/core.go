package gorm

func (config *Config) WithLoggerConsole(state bool) {
	config.loggerConsole = state
}

func (config *Config) WithLoggerHandle(handle func(b []byte)) {
	config.loggerHandle = handle
}

func (config *Config) WithAutoMigrate(state bool) {
	config.autoMigrate = state
}
