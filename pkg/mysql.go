package gorm

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/go-sql-driver/mysql"
	"github.com/lhdhtrc/gorm/pkg/internal"
	mysql2 "gorm.io/driver/mysql"
	"gorm.io/gorm"
	loger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"os"
	"time"
)

type MysqlConf struct {
	Conf *Config
}

type MysqlDB struct {
	DB *gorm.DB
}

func NewMysql(mc *MysqlConf, tables []interface{}) (*MysqlDB, error) {
	clientOptions := mysql.Config{
		Net:       "tcp",
		Addr:      mc.Conf.Address,
		DBName:    mc.Conf.Database,
		Loc:       time.Local,
		ParseTime: true,
	}

	if mc.Conf.Username != "" && mc.Conf.Password != "" {
		clientOptions.User = mc.Conf.Username
		clientOptions.Passwd = mc.Conf.Password
	}
	if mc.Conf.Tls != nil && mc.Conf.Tls.CaCert != "" && mc.Conf.Tls.ClientCert != "" && mc.Conf.Tls.ClientCertKey != "" {
		certPool := x509.NewCertPool()
		CAFile, CAErr := os.ReadFile(mc.Conf.Tls.CaCert)
		if CAErr != nil {
			return nil, CAErr
		}
		certPool.AppendCertsFromPEM(CAFile)

		clientCert, clientCertErr := tls.LoadX509KeyPair(mc.Conf.Tls.ClientCert, mc.Conf.Tls.ClientCertKey)
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
	if mc.Conf.Logger {
		_default = internal.New(internal.Config{
			Config: loger.Config{
				SlowThreshold: 200 * time.Millisecond,
				LogLevel:      loger.Info,
				Colorful:      true,
			},
			Console:      mc.Conf.loggerConsole,
			Database:     mc.Conf.Database,
			DatabaseType: mc.Conf.Type,
		}, mc.Conf.loggerHandle)
	}
	db, err := gorm.Open(mysql2.Open(clientOptions.FormatDSN()), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   mc.Conf.TablePrefix,
			SingularTable: mc.Conf.SingularTable,
		},

		DisableForeignKeyConstraintWhenMigrating: mc.Conf.DisableForeignKeyConstraintWhenMigrating,

		SkipDefaultTransaction: mc.Conf.SkipDefaultTransaction,
		PrepareStmt:            mc.Conf.PrepareStmt,
		Logger:                 _default,
	})
	if err != nil {
		return nil, err
	}

	if len(tables) != 0 && mc.Conf.autoMigrate {
		// 初始化表结构
		if err = db.AutoMigrate(tables...); err != nil {
			return nil, err
		}
	}

	d, _de := db.DB()
	if _de != nil {
		return nil, _de
	}
	d.SetMaxOpenConns(mc.Conf.MaxOpenConnects)
	d.SetMaxIdleConns(mc.Conf.MaxIdleConnects)
	d.SetConnMaxLifetime(time.Minute * time.Duration(mc.Conf.ConnMaxLifeTime))

	return &MysqlDB{DB: db}, nil
}
