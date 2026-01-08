package gormx

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

const (
	// Postgres 表示 Postgres 数据库类型。
	Postgres int32 = iota + 1
	// Oracle 表示 Oracle 数据库类型。
	Oracle
	// Sqlite 表示 Sqlite 数据库类型。
	Sqlite
	// Mysql 表示 MySQL 数据库类型。
	Mysql
	// Mssql 表示 Microsoft SQL Server 数据库类型。
	Mssql
)

// TLS 描述数据库连接的 TLS（双向）证书配置。
type TLS struct {
	// CaCert 为 CA 证书文件路径。
	CaCert string `json:"ca_cert"`
	// ClientCert 为客户端证书文件路径。
	ClientCert string `json:"client_cert"`
	// ClientCertKey 为客户端证书私钥文件路径。
	ClientCertKey string `json:"client_cert_key"`
}

// Config 为 gorm 初始化所需的配置项集合。
type Config struct {
	// 1-postgres 2-oracle 3-sqlite 4-mysql 5-mssql
	// Type 表示数据库类型枚举。
	Type int32 `json:"type"`
	// TLS加密配置（生产环境建议启用），如果不为null则启用tls加密
	// Tls 为 TLS 配置，非空时启用 TLS。
	Tls *TLS `json:"tls"`

	// Address 为数据库地址，一般为 host:port。
	Address string `json:"address"`
	// Database 为数据库名称。
	Database string `json:"database"`
	// Username 为数据库用户名。
	Username string `json:"username"`
	// Password 为数据库密码。
	Password string `json:"password"`

	// 表名前缀
	// TablePrefix 会拼接到 gorm 的表名之前。
	TablePrefix string `json:"table_prefix"`

	// 最大打开连接数（建议：根据负载设置，默认100），0表示无限制（不推荐生产环境使用）
	// MaxOpenConnects 会设置到 database/sql 连接池的 MaxOpenConns。
	MaxOpenConnects int `json:"max_open_connects"`
	// 最大空闲连接数（建议：保持适当空闲连接减少握手开销），0表示无限制（需配合max_open_connects使用）
	// MaxIdleConnects 会设置到 database/sql 连接池的 MaxIdleConns。
	MaxIdleConnects int `json:"max_idle_connects"`
	// 连接最大生命周期（单位：秒，建议：300-600秒），超时后连接会被强制回收重建
	// ConnMaxLifeTime 会设置到 database/sql 连接池的 ConnMaxLifetime（按秒）。
	ConnMaxLifeTime int `json:"conn_max_life_time"`

	// 是否为单数表名
	// SingularTable 为 true 时，表名不做复数化。
	SingularTable bool `json:"singular_table"`
	// 是否禁用物理外键
	// DisableForeignKeyConstraintWhenMigrating 为 true 时，自动迁移时不创建物理外键。
	DisableForeignKeyConstraintWhenMigrating bool `json:"disable_foreign_key_constraint_when_migrating"`
	// 是否跳过默认事务（特殊场景使用，如批量导入）
	// SkipDefaultTransaction 为 true 时，gorm 默认写操作不启事务。
	SkipDefaultTransaction bool `json:"skip_default_transaction"`
	// 是否启用预处理语句（安全建议：始终开启防止SQL注入）
	// PrepareStmt 为 true 时，gorm 将启用预处理语句缓存。
	PrepareStmt bool `json:"prepare_stmt"`

	// 是否启用SQL日志（调试建议开启，生产环境建议关闭）
	// Logger 为 true 时启用 gorm logger，并可通过 WithLoggerConsole/WithLoggerHandle 控制输出。
	Logger bool `json:"logger"`

	// autoMigrate 控制 NewMysql/NewPostgres 是否执行 AutoMigrate。
	autoMigrate bool
	// loggerHandle 为日志回调句柄。
	loggerHandle func(b []byte)
	// loggerConsole 控制是否输出到控制台。
	loggerConsole bool
}

// Table 为常用的 uint64 主键基类（带软删除）。
type Table struct {
	ID        uint64                `json:"id" gorm:"primarykey"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
	DeletedAt soft_delete.DeletedAt `json:"deleted_at" gorm:"default:0;index"`
}

// TableUnique 为带 uniqueIndex 的基类（以 DeletedAt 实现“软删除下的唯一性”）。
type TableUnique struct {
	ID        uint64                `json:"id" gorm:"primaryKey;"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
	DeletedAt soft_delete.DeletedAt `json:"deleted_at" gorm:"default:0;uniqueIndex:idx_unique"`
}

// TableUUID 为 string 主键基类（使用 UUIDv7）。
type TableUUID struct {
	ID        string                `json:"id" gorm:"type:uuid;primaryKey;"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
	DeletedAt soft_delete.DeletedAt `json:"deleted_at" gorm:"default:0;index"`
}

// BeforeCreate 在插入前生成 UUIDv7 作为主键。
func (t *TableUUID) BeforeCreate(_ *gorm.DB) error {
	t.ID = NewUUIDv7()
	return nil
}

// TableUUIDUnique 为带 uniqueIndex 的 UUID 主键基类。
type TableUUIDUnique struct {
	ID        string                `json:"id" gorm:"type:uuid;primaryKey;"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
	DeletedAt soft_delete.DeletedAt `json:"deleted_at" gorm:"default:0;uniqueIndex:idx_unique"`
}

// BeforeCreate 在插入前生成 UUIDv7 作为主键。
func (t *TableUUIDUnique) BeforeCreate(_ *gorm.DB) error {
	t.ID = NewUUIDv7()
	return nil
}
