package internal

import (
	"fmt"
	"go.uber.org/zap"
	"gorm.io/gorm/logger"
)

type GormWriter struct {
	logger.Writer
	Logger *zap.Logger
}

// NewWriter writer 构造函数
func NewWriter(w logger.Writer, s *zap.Logger) *GormWriter {
	return &GormWriter{
		Writer: w,
		Logger: s,
	}
}

// Printf 格式化打印日志
func (w *GormWriter) Printf(message string, data ...interface{}) {
	w.Logger.Info(fmt.Sprintf(message, data...))
	fmt.Println(data)
}
