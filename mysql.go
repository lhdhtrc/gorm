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

// mysqlTLSConfigSeq 用于生成唯一的 TLS 配置名，避免重复注册冲突。
var mysqlTLSConfigSeq uint64

// NewMysql 使用配置初始化 MySQL 的 gorm.DB，并可选执行 AutoMigrate。
func NewMysql(mc *MysqlConf, tables []interface{}) (*MysqlDB, error) {
	// clientOptions 为 go-sql-driver/mysql 的连接配置。
	clientOptions := mysql.Config{
		Net:    "tcp",
		Addr:   mc.Address,
		DBName: mc.Database,
		// Loc 统一使用 UTC。
		Loc: time.UTC,
		// ParseTime 让 time 类型字段可被正确扫描。
		ParseTime: true,
	}

	if mc.Username != "" {
		clientOptions.User = mc.Username
		if mc.Password != "" {
			clientOptions.Passwd = mc.Password
		}
	}

	// 若提供 TLS 证书配置，则注册 TLS 并写入 DSN 的 TLSConfig 名称。
	if mc.Tls != nil && mc.Tls.CaCert != "" && mc.Tls.ClientCert != "" && mc.Tls.ClientCertKey != "" {
		// certPool 用于保存 CA 证书链。
		certPool := x509.NewCertPool()
		// 读取 CA 证书文件内容。
		caFile, err := os.ReadFile(mc.Tls.CaCert)
		// 读取失败直接返回错误。
		if err != nil {
			return nil, err
		}
		// 将 CA 证书追加到证书池。
		if ok := certPool.AppendCertsFromPEM(caFile); !ok {
			return nil, errors.New("failed to append ca cert")
		}

		// 加载客户端证书与私钥，用于双向 TLS。
		clientCert, err := tls.LoadX509KeyPair(mc.Tls.ClientCert, mc.Tls.ClientCertKey)
		// 加载失败直接返回错误。
		if err != nil {
			return nil, err
		}

		// tlsConfig 为本次连接使用的 TLS 配置。
		tlsConfig := tls.Config{
			// Certificates 为客户端证书链。
			Certificates: []tls.Certificate{clientCert},
			// RootCAs 为服务端证书的信任根。
			RootCAs: certPool,
		}

		// tlsConfigName 为唯一 TLS 配置名，避免 mysql.RegisterTLSConfig 重复注册冲突。
		tlsConfigName := "gormx_" + strconv.FormatUint(atomic.AddUint64(&mysqlTLSConfigSeq, 1), 10)
		// 将 TLS 配置注册到 go-sql-driver/mysql 的全局注册表中。
		if err := mysql.RegisterTLSConfig(tlsConfigName, &tlsConfig); err != nil {
			return nil, err
		}
		// 在 DSN 中引用该 TLS 配置名。
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
		// NamingStrategy 控制表名前缀与单复数规则。
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   mc.TablePrefix,
			SingularTable: mc.SingularTable,
		},
		// NowFunc 统一生成 UTC 时间。
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		// DisableForeignKeyConstraintWhenMigrating 控制迁移时是否创建外键。
		DisableForeignKeyConstraintWhenMigrating: mc.DisableForeignKeyConstraintWhenMigrating,
		// SkipDefaultTransaction 控制 gorm 默认事务行为。
		SkipDefaultTransaction: mc.SkipDefaultTransaction,
		// PrepareStmt 控制是否启用预处理语句。
		PrepareStmt: mc.PrepareStmt,
		// Logger 为 gorm 的日志实现。
		Logger: gormLogger,
	})
	// 打开失败直接返回错误。
	if err != nil {
		return nil, err
	}

	// 当启用 autoMigrate 且传入表模型时，执行自动迁移。
	if len(tables) != 0 && mc.autoMigrate {
		// AutoMigrate 会创建/修改表结构以匹配模型。
		if err = db.AutoMigrate(tables...); err != nil {
			return nil, err
		}
	}

	// 获取底层 *sql.DB 以配置连接池参数。
	d, err := db.DB()
	// 获取失败直接返回错误。
	if err != nil {
		return nil, err
	}

	// 设置最大打开连接数。
	d.SetMaxOpenConns(mc.MaxOpenConnects)
	// 设置最大空闲连接数。
	d.SetMaxIdleConns(mc.MaxIdleConnects)
	// ConnMaxLifeTime 约定为秒，<=0 表示不限制。
	if mc.ConnMaxLifeTime > 0 {
		d.SetConnMaxLifetime(time.Second * time.Duration(mc.ConnMaxLifeTime))
	} else {
		d.SetConnMaxLifetime(0)
	}

	// 返回封装后的 MysqlDB。
	return &MysqlDB{DB: db}, nil
}
