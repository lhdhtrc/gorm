package gorm

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/lhdhtrc/gorm/pkg/internal"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	loger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"os"
	"strings"
	"time"
)

type PostgresConf struct {
	Conf *Config
}

type PostgresDB struct {
	DB *gorm.DB
}

func NewPostgres(mc *PostgresConf, tables []interface{}) (*PostgresDB, error) {
	addr := strings.Split(mc.Conf.Address, ":")

	var dsn []string
	dsn = append(dsn, fmt.Sprintf("host=%s port=%v dbname=%s TimeZone=Asia/Shanghai", addr[0], addr[1], mc.Conf.Database))
	if mc.Conf.Username != "" && mc.Conf.Password != "" {
		dsn = append(dsn, fmt.Sprintf("user=%s password=%s", mc.Conf.Username, mc.Conf.Password))
	}
	if mc.Conf.Tls != nil && mc.Conf.Tls.CaCert != "" && mc.Conf.Tls.ClientCert != "" && mc.Conf.Tls.ClientCertKey != "" {
		dsn = append(dsn, "sslmode=require")
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

		stdlib.RegisterConnConfig(&pgx.ConnConfig{
			Config: pgconn.Config{
				TLSConfig: &tlsConfig,
			},
			Tracer:                   nil,
			StatementCacheCapacity:   0,
			DescriptionCacheCapacity: 0,
			DefaultQueryExecMode:     0,
		})
	} else {
		dsn = append(dsn, "sslmode=disable")
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
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: strings.Join(dsn, " "),
	}), &gorm.Config{
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

	return &PostgresDB{DB: db}, nil
}
