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

type PostgresConf Config

type PostgresDB struct {
	DB *gorm.DB
}

func NewPostgres(mc *PostgresConf, tables []interface{}) (*PostgresDB, error) {
	addr := strings.Split(mc.Address, ":")

	var dsn []string
	dsn = append(dsn, fmt.Sprintf("host=%s port=%v dbname=%s TimeZone=UTC", addr[0], addr[1], mc.Database))
	if mc.Username != "" && mc.Password != "" {
		dsn = append(dsn, fmt.Sprintf("user=%s password=%s", mc.Username, mc.Password))
	}
	if mc.Tls != nil && mc.Tls.CaCert != "" && mc.Tls.ClientCert != "" && mc.Tls.ClientCertKey != "" {
		dsn = append(dsn, "sslmode=require")
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
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: strings.Join(dsn, " "),
	}), &gorm.Config{
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

	return &PostgresDB{DB: db}, nil
}
