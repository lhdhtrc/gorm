package gormx

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"
	"time"

	"github.com/fireflycore/gormx/internal"
	loger "gorm.io/gorm/logger"
)

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

// newGormLogger 根据 Config 构造 gorm logger。
func newGormLogger(cfg *Config) loger.Interface {
	// cfg 为空或未开启 Logger 时，直接丢弃日志输出。
	if cfg == nil || !cfg.Logger {
		// Discard 为 gorm 提供的空实现。
		return loger.Discard
	}

	// internal.New 返回一个实现 loger.Interface 的自定义 logger。
	return internal.New(internal.Config{
		// Config 复用 gorm 自带的 logger.Config。
		Config: loger.Config{
			// SlowThreshold 为慢 SQL 阈值。
			SlowThreshold: 200 * time.Millisecond,
			// LogLevel 使用 Info，输出 SQL Trace。
			LogLevel: loger.Info,
			// Colorful 控制控制台彩色输出。
			Colorful: true,
		},
		// Console 控制是否输出到控制台。
		Console: cfg.loggerConsole,
		// Database 记录库名，便于日志聚合。
		Database: cfg.Database,
		// DatabaseType 记录库类型，便于日志聚合。
		DatabaseType: cfg.Type,
		// cfg.loggerHandle 为结构化日志回调句柄。
	}, cfg.loggerHandle)
}

// newClientTLSConfig 根据 TLS 配置生成 *tls.Config。
func newClientTLSConfig(tlsCfg *TLS) (*tls.Config, bool, error) {
	// TLS 配置为空或字段不完整时，认为不启用 TLS。
	if tlsCfg == nil || tlsCfg.CaCert == "" || tlsCfg.ClientCert == "" || tlsCfg.ClientCertKey == "" {
		// 返回 enabled=false，且不视为错误。
		return nil, false, nil
	}

	// certPool 保存信任的 CA 证书。
	certPool := x509.NewCertPool()
	// 读取 CA 证书文件。
	caFile, err := os.ReadFile(tlsCfg.CaCert)
	// 读取失败则返回错误。
	if err != nil {
		return nil, false, err
	}
	// 将 CA 证书追加到 certPool 中。
	if ok := certPool.AppendCertsFromPEM(caFile); !ok {
		return nil, false, errors.New("failed to append ca cert")
	}

	// 加载客户端证书与私钥。
	clientCert, err := tls.LoadX509KeyPair(tlsCfg.ClientCert, tlsCfg.ClientCertKey)
	// 加载失败则返回错误。
	if err != nil {
		return nil, false, err
	}

	// 返回构造完成的 TLS 配置，并标记 enabled=true。
	return &tls.Config{
		// Certificates 为客户端证书链。
		Certificates: []tls.Certificate{clientCert},
		// RootCAs 为服务端证书的信任根。
		RootCAs: certPool,
	}, true, nil
}
