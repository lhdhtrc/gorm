package gormx

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

const (
	// ResultSuccess 表示成功执行 SQL 的结果标记。
	ResultSuccess = "success"

	// TraceId 为从 metadata 读取 trace id 的 key。
	TraceId = "trace-id"
	// UserId 为从 metadata 读取 user id 的 key。
	UserId = "user-id"
	// AppId 为从 metadata 读取调用方 app id 的 key。
	AppId = "app-id"
)
