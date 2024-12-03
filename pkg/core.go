package gorm

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/go-sql-driver/mysql"
	"github.com/lhdhtrc/gorm/pkg/internal"
	mysql2 "gorm.io/driver/mysql"
	"gorm.io/gorm"
	loger "gorm.io/gorm/logger"
	"io"
	"log"
	"os"
	"time"
)

func NewMysql(config *Config, tables []interface{}) (*gorm.DB, error) {
	clientOptions := mysql.Config{
		Net:       "tcp",
		Addr:      config.Address,
		DBName:    config.Database,
		Loc:       time.Local,
		ParseTime: true,
	}

	if config.Account != "" && config.Password != "" {
		clientOptions.User = config.Account
		clientOptions.Passwd = config.Password
	}
	if config.Tls.CaCert != "" && config.Tls.ClientCert != "" && config.Tls.ClientCertKey != "" {
		certPool := x509.NewCertPool()
		CAFile, CAErr := os.ReadFile(config.Tls.CaCert)
		if CAErr != nil {
			return nil, CAErr
		}
		certPool.AppendCertsFromPEM(CAFile)

		clientCert, clientCertErr := tls.LoadX509KeyPair(config.Tls.ClientCert, config.Tls.ClientCertKey)
		if clientCertErr != nil {
			return nil, clientCertErr
		}

		tlsConfig := tls.Config{
			Certificates: []tls.Certificate{clientCert},
			RootCAs:      certPool,
		}

		if err := mysql.RegisterTLSConfig("custom", &tlsConfig); err != nil {
			return nil, err
		}

		clientOptions.TLSConfig = "custom"
	}

	var _default loger.Interface
	if config.LoggerEnable {
		var writer io.Writer
		if config.loggerConsole {
			writer = os.Stdout
		} else {
			writer = &internal.CustomWriter{}
		}

		_default = internal.New(config.Prefix, log.New(writer, "\r", log.LstdFlags), loger.Config{
			SlowThreshold: 200 * time.Millisecond,
			LogLevel:      loger.Info,
			Colorful:      true,
		}, config.loggerHandle)
	}
	db, err := gorm.Open(mysql2.Open(clientOptions.FormatDSN()), &gorm.Config{
		SkipDefaultTransaction: config.SkipDefaultTransaction,
		PrepareStmt:            config.PrepareStmt,
		Logger:                 _default,
	})
	if err != nil {
		return nil, err
	}

	if len(tables) != 0 && config.autoMigrate {
		// 初始化表结构
		if err = db.AutoMigrate(tables...); err != nil {
			return nil, err
		}
	}

	d, _de := db.DB()
	if _de != nil {
		return nil, _de
	}
	d.SetMaxOpenConns(config.MaxOpenConnects)
	d.SetMaxIdleConns(config.MaxIdleConnects)
	d.SetConnMaxLifetime(time.Minute * time.Duration(config.ConnMaxLifeTime))

	return db, nil
}

func (config *Config) WithLoggerConsole(state bool) {
	config.loggerConsole = state
}

func (config *Config) WithLoggerHandle(handle func(b []byte)) {
	config.loggerHandle = handle
}

func (config *Config) WithAutoMigrate(state bool) {
	config.autoMigrate = state
}
