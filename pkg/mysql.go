package gorm

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/go-sql-driver/mysql"
	"github.com/lhdhtrc/gorm/pkg/internal"
	mysql2 "gorm.io/driver/mysql"
	"gorm.io/gorm"
	loger "gorm.io/gorm/logger"
	"os"
	"time"
)

type MysqlConf struct {
	Data *Config
}

func (ist *MysqlConf) GetConfig() *Config {
	return ist.Data
}

type MysqlDB struct {
	DB *gorm.DB
}

func (ist *MysqlDB) GetDB() *gorm.DB {
	return ist.DB
}

func NewMysql(config *MysqlConf, tables []interface{}) (*MysqlDB, error) {
	confData := config.GetConfig()

	clientOptions := mysql.Config{
		Net:       "tcp",
		Addr:      confData.Address,
		DBName:    confData.Database,
		Loc:       time.Local,
		ParseTime: true,
	}

	if confData.Username != "" && confData.Password != "" {
		clientOptions.User = confData.Username
		clientOptions.Passwd = confData.Password
	}
	if confData.Tls.CaCert != "" && confData.Tls.ClientCert != "" && confData.Tls.ClientCertKey != "" {
		certPool := x509.NewCertPool()
		CAFile, CAErr := os.ReadFile(confData.Tls.CaCert)
		if CAErr != nil {
			return nil, CAErr
		}
		certPool.AppendCertsFromPEM(CAFile)

		clientCert, clientCertErr := tls.LoadX509KeyPair(confData.Tls.ClientCert, confData.Tls.ClientCertKey)
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
	if confData.Logger {
		_default = internal.New(internal.Config{
			Config: loger.Config{
				SlowThreshold: 200 * time.Millisecond,
				LogLevel:      loger.Info,
				Colorful:      true,
			},
			Console:      confData.loggerConsole,
			Database:     confData.Database,
			DatabaseType: confData.Type,
		}, confData.loggerHandle)
	}
	db, err := gorm.Open(mysql2.Open(clientOptions.FormatDSN()), &gorm.Config{
		SkipDefaultTransaction: confData.SkipDefaultTransaction,
		PrepareStmt:            confData.PrepareStmt,
		Logger:                 _default,
	})
	if err != nil {
		return nil, err
	}

	if len(tables) != 0 && confData.autoMigrate {
		// 初始化表结构
		if err = db.AutoMigrate(tables...); err != nil {
			return nil, err
		}
	}

	d, _de := db.DB()
	if _de != nil {
		return nil, _de
	}
	d.SetMaxOpenConns(confData.MaxOpenConnects)
	d.SetMaxIdleConns(confData.MaxIdleConnects)
	d.SetConnMaxLifetime(time.Minute * time.Duration(confData.ConnMaxLifeTime))

	return &MysqlDB{DB: db}, nil
}
