package core

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/lhdhtrc/gorm/core/internal"
	"github.com/lhdhtrc/gorm/model"
	"go.uber.org/zap"
	mysql2 "gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

func SetupMysql(logger *zap.Logger, config *model.ConfigEntity, tables *[]interface{}) *gorm.DB {
	logPrefix := "setup mysql"
	logger.Info(fmt.Sprintf("%s start ->", logPrefix))

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
			logger.Error(fmt.Sprintf("%s read %s error: %s", logPrefix, config.Tls.CaCert, CAErr.Error()))
			return nil
		}
		certPool.AppendCertsFromPEM(CAFile)

		clientCert, clientCertErr := tls.LoadX509KeyPair(config.Tls.ClientCert, config.Tls.ClientCertKey)
		if clientCertErr != nil {
			logger.Error(fmt.Sprintf("%s tls.LoadX509KeyPair err: %v", logPrefix, clientCertErr))
			return nil
		}

		tlsConfig := tls.Config{
			Certificates: []tls.Certificate{clientCert},
			RootCAs:      certPool,
		}

		if err := mysql.RegisterTLSConfig("custom", &tlsConfig); err != nil {
			logger.Error(fmt.Sprintf("%s tls.LoadX509KeyPair err: %v", logPrefix, err.Error()))
			return nil
		}

		clientOptions.TLSConfig = "custom"
	}

	var _default gormLogger.Interface
	if config.LoggerEnable {
		_default = gormLogger.New(internal.NewWriter(log.New(os.Stdout, "\r\n", log.LstdFlags), logger), gormLogger.Config{
			SlowThreshold: 200 * time.Millisecond,
			LogLevel:      gormLogger.Info,
			Colorful:      true,
		})
	}
	db, _oe := gorm.Open(mysql2.Open(clientOptions.FormatDSN()), &gorm.Config{
		SkipDefaultTransaction: config.SkipDefaultTransaction,
		PrepareStmt:            config.PrepareStmt,
		Logger:                 _default,
	})
	if _oe != nil {
		panic(fmt.Errorf("gorm open mysql error: %s", _oe))
	}

	if len(*tables) != 0 {
		// 初始化表结构
		if _te := db.AutoMigrate(*tables...); _te != nil {
			panic(fmt.Errorf("gorm db batch create table error: %s", _te))
		}
	}

	d, _de := db.DB()
	if _de != nil {
		panic(fmt.Errorf("gorm open db error: %s", _de))
	}
	d.SetMaxOpenConns(config.MaxOpenConnects)
	d.SetMaxIdleConns(config.MaxIdleConnects)
	d.SetConnMaxLifetime(time.Minute * time.Duration(config.ConnMaxLifeTime))

	logger.Info(fmt.Sprintf("%s success ->", logPrefix))

	return db
}
