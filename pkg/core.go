package gorm

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/lhdhtrc/gorm/pkg/internal"
	mysql2 "gorm.io/driver/mysql"
	"gorm.io/gorm"
	loger "gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

func InstallMysql(config *ConfigEntity, tables []interface{}, loggerHandle func(b []byte)) (*gorm.DB, error) {
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
		_default = internal.New(log.New(os.Stdout, "\r\n", log.LstdFlags), loger.Config{
			SlowThreshold: 200 * time.Millisecond,
			LogLevel:      loger.Info,
			Colorful:      true,
		}, loggerHandle)
	}
	db, _oe := gorm.Open(mysql2.Open(clientOptions.FormatDSN()), &gorm.Config{
		SkipDefaultTransaction: config.SkipDefaultTransaction,
		PrepareStmt:            config.PrepareStmt,
		Logger:                 _default,
	})
	if _oe != nil {
		panic(fmt.Errorf("gorm open mysql error: %s", _oe))
	}

	if len(tables) != 0 {
		// 初始化表结构
		if _te := db.AutoMigrate(tables...); _te != nil {
			panic(fmt.Errorf("gorm db batch create table error: %s", _te))
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
