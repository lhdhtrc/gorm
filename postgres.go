package gormx

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fireflycore/gormx/internal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	loger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type PostgresConf struct {
	Config
}

type PostgresDB struct {
	DB *gorm.DB
}

func NewPostgres(mc *PostgresConf, tables []interface{}) (*PostgresDB, error) {
	host, port, err := splitHostPort(mc.Address, "5432")
	if err != nil {
		return nil, err
	}

	dsnParts := []string{
		"host=" + host,
		"port=" + port,
		"dbname=" + mc.Database,
		"TimeZone=UTC",
	}
	if mc.Username != "" {
		dsnParts = append(dsnParts, "user="+mc.Username)
		if mc.Password != "" {
			dsnParts = append(dsnParts, "password="+mc.Password)
		}
	}

	var tlsConfig *tls.Config
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

		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{clientCert},
			RootCAs:      certPool,
		}

		dsnParts = append(dsnParts, "sslmode=require")
	} else {
		dsnParts = append(dsnParts, "sslmode=disable")
	}

	connConfig, err := pgx.ParseConfig(strings.Join(dsnParts, " "))
	if err != nil {
		return nil, err
	}
	if tlsConfig != nil {
		connConfig.TLSConfig = tlsConfig
	}

	sqlDB := stdlib.OpenDB(*connConfig)

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

	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
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
		_ = sqlDB.Close()
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

	return &PostgresDB{DB: db}, nil
}

func splitHostPort(addr string, defaultPort string) (string, string, error) {
	if addr == "" {
		return "", "", errors.New("empty address")
	}

	host, port, err := net.SplitHostPort(addr)
	if err == nil {
		if host == "" {
			return "", "", errors.New("empty host")
		}
		if port == "" {
			return host, defaultPort, nil
		}
		if _, err := strconv.Atoi(port); err != nil {
			return "", "", err
		}
		return host, port, nil
	}

	host = strings.TrimSpace(addr)
	if host == "" {
		return "", "", errors.New("empty host")
	}
	return host, defaultPort, nil
}
