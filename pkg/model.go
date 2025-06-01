package gorm

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

const (
	Postgres int32 = iota + 1
	Oracle
	Sqlite
	Mysql
	Mssql
)

type TLS struct {
	CaCert        string `json:"ca_cert" bson:"ca_cert" yaml:"ca_cert" mapstructure:"ca_cert"`
	ClientCert    string `json:"client_cert" bson:"client_cert" yaml:"client_cert" mapstructure:"client_cert"`
	ClientCertKey string `json:"client_cert_key" bson:"client_cert_key" yaml:"client_cert_key" mapstructure:"client_cert_key"`
}

type Config struct {
	// 1-postgres 2-oracle 3-sqlite 4-mysql 5-mssql
	Type int32 `json:"type"`
	// TLS加密配置（生产环境建议启用），如果不为null则启用tls加密
	Tls *TLS `json:"tls" bson:"tls" yaml:"tls" mapstructure:"tls"`

	Address  string `json:"address" yaml:"address" mapstructure:"address"`
	Database string `json:"database" yaml:"database" mapstructure:"database"`
	Username string `json:"username" yaml:"username" mapstructure:"username"`
	Password string `json:"password" bson:"password" yaml:"password" mapstructure:"password"`

	// 最大打开连接数（建议：根据负载设置，默认100），0表示无限制（不推荐生产环境使用）
	MaxOpenConnects int `json:"max_open_connects" bson:"max_open_connects" yaml:"max_open_connects" mapstructure:"max_open_connects"`
	// 最大空闲连接数（建议：保持适当空闲连接减少握手开销），0表示无限制（需配合max_open_connects使用）
	MaxIdleConnects int `json:"max_idle_connects" bson:"max_idle_connects" yaml:"max_idle_connects" mapstructure:"max_idle_connects"`
	// 连接最大生命周期（单位：秒，建议：300-600秒），超时后连接会被强制回收重建
	ConnMaxLifeTime int `json:"conn_max_life_time" bson:"conn_max_life_time" yaml:"conn_max_life_time" mapstructure:"conn_max_life_time"`

	// 是否跳过默认事务（特殊场景使用，如批量导入）
	SkipDefaultTransaction bool `json:"skip_default_transaction" bson:"skip_default_transaction" yaml:"skip_default_transaction" mapstructure:"skip_default_transaction"`
	// 是否启用预处理语句（安全建议：始终开启防止SQL注入）
	PrepareStmt bool `json:"prepare_stmt" bson:"prepare_stmt" yaml:"prepare_stmt" mapstructure:"prepare_stmt"`

	// 是否启用SQL日志（调试建议开启，生产环境建议关闭）
	Logger bool `json:"logger" bson:"logger" yaml:"logger" mapstructure:"logger"`

	autoMigrate   bool
	loggerHandle  func(b []byte)
	loggerConsole bool
}

type Table struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

type TableUUID struct {
	ID        uuid.UUID      `json:"id" gorm:"type:binary(16);primaryKey;"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

type PostgresTableUUID struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
