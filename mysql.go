package gormx

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/fireflycore/gormx/internal"
	"github.com/go-sql-driver/mysql"
	mysql2 "gorm.io/driver/mysql"
	"gorm.io/gorm"
	loger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type MysqlConf struct {
	Config
}

type MysqlDB struct {
	DB *gorm.DB
}

var mysqlTLSConfigSeq uint64

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
		caFile, err := os.ReadFile(mc.Tls.CaCert)
		if err != nil {
			return nil, err
		}
		if ok := certPool.AppendCertsFromPEM(caFile); !ok {
			return nil, errors.New("failed to append ca cert")
		}

		clientCert, err := tls.LoadX509KeyPair(mc.Tls.ClientCert, mc.Tls.ClientCertKey)
		if err != nil {
			return nil, err
		}

		tlsConfig := tls.Config{
			Certificates: []tls.Certificate{clientCert},
			RootCAs:      certPool,
		}

		tlsConfigName := "gormx_" + strconv.FormatUint(atomic.AddUint64(&mysqlTLSConfigSeq, 1), 10)
		if err := mysql.RegisterTLSConfig(tlsConfigName, &tlsConfig); err != nil {
			return nil, err
		}
		clientOptions.TLSConfig = tlsConfigName
	}

	gormLogger := loger.Discard
	if mc.Logger {
		gormLogger = internal.New(internal.Config{
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
		SkipDefaultTransaction:                   mc.SkipDefaultTransaction,
		PrepareStmt:                              mc.PrepareStmt,
		Logger:                                   gormLogger,
	})
	if err != nil {
		return nil, err
	}

	if len(tables) != 0 && mc.autoMigrate {
		if err = db.AutoMigrate(tables...); err != nil {
			return nil, err
		}
	}

	d, err := db.DB()
	if err != nil {
		return nil, err
	}

	d.SetMaxOpenConns(mc.MaxOpenConnects)
	d.SetMaxIdleConns(mc.MaxIdleConnects)
	if mc.ConnMaxLifeTime > 0 {
		d.SetConnMaxLifetime(time.Second * time.Duration(mc.ConnMaxLifeTime))
	} else {
		d.SetConnMaxLifetime(0)
	}

	return &MysqlDB{DB: db}, nil
}
