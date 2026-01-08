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

// NewPostgres 使用配置初始化 Postgres 的 gorm.DB，并可选执行 AutoMigrate。
func NewPostgres(mc *PostgresConf, tables []interface{}) (*PostgresDB, error) {
	// 将 address 拆为 host/port，未带端口时使用默认值 5432。
	host, port, err := splitHostPort(mc.Address, "5432")
	// 解析失败直接返回错误。
	if err != nil {
		return nil, err
	}

	// dsnParts 为 key=value 形式的 DSN 片段，最终将用空格拼接。
	dsnParts := []string{
		"host=" + host,
		"port=" + port,
		"dbname=" + mc.Database,
		"TimeZone=UTC",
	}
	// 若提供用户名，则写入 DSN。
	if mc.Username != "" {
		dsnParts = append(dsnParts, "user="+mc.Username)
		// 若提供密码，则写入 DSN。
		if mc.Password != "" {
			dsnParts = append(dsnParts, "password="+mc.Password)
		}
	}

	// tlsConfig 非空表示启用 TLS。
	var tlsConfig *tls.Config
	// 当 TLS 配置完整时，加载证书并启用 sslmode=require。
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

		// 构造 TLS 配置供 pgx 使用。
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{clientCert},
			RootCAs:      certPool,
		}

		// 指定强制使用 TLS。
		dsnParts = append(dsnParts, "sslmode=require")
	} else {
		// 未提供 TLS 配置时禁用 TLS（与旧实现兼容）。
		dsnParts = append(dsnParts, "sslmode=disable")
	}

	// 将 DSN 解析为 pgx 连接配置。
	connConfig, err := pgx.ParseConfig(strings.Join(dsnParts, " "))
	// 解析失败直接返回错误。
	if err != nil {
		return nil, err
	}
	// 若启用 TLS，则将 TLSConfig 注入到 pgx ConnConfig。
	if tlsConfig != nil {
		connConfig.TLSConfig = tlsConfig
	}

	// 使用 pgx stdlib 将 ConnConfig 转成 *sql.DB，交给 gorm driver 复用连接池。
	sqlDB := stdlib.OpenDB(*connConfig)

	// gormLogger 默认为丢弃输出，避免 nil 语义不清。
	gormLogger := loger.Discard
	// 若开启 Logger，则注入自定义 logger。
	if mc.Logger {
		// internal.New 构造一个实现 loger.Interface 的自定义 logger。
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

	// 打开 gorm DB，并配置命名策略、NowFunc、事务与 logger 等选项。
	db, err := gorm.Open(postgres.New(postgres.Config{
		// Conn 复用上面创建的 *sql.DB。
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
	// 打开失败时关闭 sqlDB，避免连接泄漏。
	if err != nil {
		_ = sqlDB.Close()
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

	// 返回封装后的 PostgresDB。
	return &PostgresDB{DB: db}, nil
}

// splitHostPort 将 addr 解析为 host/port，addr 不带端口时返回 defaultPort。
func splitHostPort(addr string, defaultPort string) (string, string, error) {
	// 空地址直接返回错误。
	if addr == "" {
		return "", "", errors.New("empty address")
	}

	// 优先尝试按 host:port 解析。
	host, port, err := net.SplitHostPort(addr)
	// 解析成功则做基本校验后返回。
	if err == nil {
		// host 不能为空。
		if host == "" {
			return "", "", errors.New("empty host")
		}
		// 端口为空时使用默认端口。
		if port == "" {
			return host, defaultPort, nil
		}
		// 校验端口为数字。
		if _, err := strconv.Atoi(port); err != nil {
			return "", "", err
		}
		// 返回解析出的 host 与 port。
		return host, port, nil
	}

	// 未带端口时，把整个 addr 当作 host。
	host = strings.TrimSpace(addr)
	// host 不能为空。
	if host == "" {
		return "", "", errors.New("empty host")
	}
	// 返回 host 与默认端口。
	return host, defaultPort, nil
}
