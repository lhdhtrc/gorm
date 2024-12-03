package gorm

import (
	"gorm.io/gorm"
	"time"
)

type TLS struct {
	CaCert        string `json:"ca_cert" bson:"ca_cert" yaml:"ca_cert" mapstructure:"ca_cert"`
	ClientCert    string `json:"client_cert" bson:"client_cert" yaml:"client_cert" mapstructure:"client_cert"`
	ClientCertKey string `json:"client_cert_key" bson:"client_cert_key" yaml:"client_cert_key" mapstructure:"client_cert_key"`
}

type Config struct {
	Tls TLS `json:"tls" bson:"tls" yaml:"tls" mapstructure:"tls"`

	Prefix   string `json:"prefix" bson:"prefix" yaml:"prefix" mapstructure:"prefix"`
	Account  string `json:"account" bson:"account" yaml:"account" mapstructure:"account"`
	Password string `json:"password" bson:"password" yaml:"password" mapstructure:"password"`

	Address  string `json:"address" yaml:"address" mapstructure:"address"`
	Database string `json:"database" yaml:"database" mapstructure:"database"`

	Mode bool `json:"mode" yaml:"mode" mapstructure:"mode"` // Mode is true cluster

	MaxOpenConnects        int  `json:"max_open_connects" bson:"max_open_connects" yaml:"max_open_connects" mapstructure:"max_open_connects"`
	MaxIdleConnects        int  `json:"max_idle_connects" bson:"max_idle_connects" yaml:"max_idle_connects" mapstructure:"max_idle_connects"`
	ConnMaxLifeTime        int  `json:"conn_max_life_time" bson:"conn_max_life_time" yaml:"conn_max_life_time" mapstructure:"conn_max_life_time"`
	SkipDefaultTransaction bool `json:"skip_default_transaction" bson:"skip_default_transaction" yaml:"skip_default_transaction" mapstructure:"skip_default_transaction"`
	PrepareStmt            bool `json:"prepare_stmt" bson:"prepare_stmt" yaml:"prepare_stmt" mapstructure:"prepare_stmt"`

	LoggerEnable bool `json:"logger_enable" bson:"logger_enable" yaml:"logger_enable" mapstructure:"logger_enable"`

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
	ID        string         `json:"id" gorm:"primarykey;size:36;"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
