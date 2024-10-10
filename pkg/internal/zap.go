package internal

import (
	"fmt"
	"go.uber.org/zap"
	"gorm.io/gorm/logger"
)

type Writer struct {
	writer logger.Writer
	logger *zap.Logger
}

// NewWriter writer 构造函数
func NewWriter(l *zap.Logger, w logger.Writer) *Writer {
	return &Writer{logger: l, writer: w}
}

// Printf 格式化打印日志
func (w *Writer) Printf(message string, data ...interface{}) {
	w.logger.Info(fmt.Sprintf(message, data...))
}
