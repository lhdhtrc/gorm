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

type MysqlConf Config

type MysqlDB struct {
	DB *gorm.DB
}

func NewMysql(mc *MysqlConf, tables []interface{}) (*MysqlDB, error) {
	clientOptions := mysql.Config{
		Net:       "tcp",
		Addr:      mc.Address,
		DBName:    mc.Database,
		Loc:       time.UTC,
		ParseTime: true,
	}

	if mc.Username != "" && mc.Password != "" {
		clientOptions.User = mc.Username
		clientOptions.Passwd = mc.Password
	}
	if mc.Tls != nil && mc.Tls.CaCert != "" && mc.Tls.ClientCert != "" && mc.Tls.ClientCertKey != "" {
		certPool := x509.NewCertPool()
		CAFile, CAErr := os.ReadFile(mc.Tls.CaCert)
		if CAErr != nil {
			return nil, CAErr
		}
		certPool.AppendCertsFromPEM(CAFile)

		clientCert, clientCertErr := tls.LoadX509KeyPair(mc.Tls.ClientCert, mc.Tls.ClientCertKey)
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
	if mc.Logger {
		_default = internal.New(internal.Config{
			Config: loger.Config{
				SlowThreshold: 200 * time.Millisecond,
				LogLevel:      loger.Info,
				Colorful:      true,
			},
			Console:      mc.loggerConsole,
			Database:     mc.Database,
			DatabaseType: mc.Type,
		}, mc.loggerHandle)
	}
	db, err := gorm.Open(mysql2.Open(clientOptions.FormatDSN()), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   mc.TablePrefix,
			SingularTable: mc.SingularTable,
		},
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},

		DisableForeignKeyConstraintWhenMigrating: mc.DisableForeignKeyConstraintWhenMigrating,

		SkipDefaultTransaction: mc.SkipDefaultTransaction,
		PrepareStmt:            mc.PrepareStmt,
		Logger:                 _default,
	})
	if err != nil {
		return nil, err
	}

	if len(tables) != 0 && mc.autoMigrate {
		// 初始化表结构
		if err = db.AutoMigrate(tables...); err != nil {
			return nil, err
		}
	}

	d, _de := db.DB()
	if _de != nil {
		return nil, _de
	}
	d.SetMaxOpenConns(mc.MaxOpenConnects)
	d.SetMaxIdleConns(mc.MaxIdleConnects)
	d.SetConnMaxLifetime(time.Minute * time.Duration(mc.ConnMaxLifeTime))

	return &MysqlDB{DB: db}, nil
}
