## gormx

基于 gorm 的工程化封装，提供：
- MySQL/Postgres 初始化（TLS、连接池、命名策略、可选 AutoMigrate）
- SQL 日志输出（控制台/回调）
- 常用模型基类（自增主键、UUIDv7 主键、软删除）
- 常用 scope（分页）

## 安装

```bash
go get github.com/fireflycore/gormx
```

## 快速开始

### MySQL

```go
package main

import (
	"github.com/fireflycore/gormx"
)

func main() {
	conf := &gormx.MysqlConf{
		Conf: gormx.Conf{
			Type:            gormx.Mysql,
			Address:         "127.0.0.1:3306",
			Database:        "demo",
			Username:        "root",
			Password:        "root",
			MaxOpenConnects: 100,
			MaxIdleConnects: 10,
			ConnMaxLifeTime: 600,
			SingularTable:   true,
			PrepareStmt:     true,
			Logger:          true,
		},
	}

	conf.WithLoggerConsole(true)
	conf.WithAutoMigrate(true)

	db, err := gormx.NewMysql(conf, []interface{}{})
	if err != nil {
		panic(err)
	}

	_ = db.DB
}
```

### Postgres

```go
package main

import (
	"github.com/fireflycore/gormx"
)

func main() {
	conf := &gormx.PostgresConf{
		Conf: gormx.Conf{
			Type:            gormx.Postgres,
			Address:         "127.0.0.1:5432",
			Database:        "demo",
			Username:        "postgres",
			Password:        "postgres",
			MaxOpenConnects: 100,
			MaxIdleConnects: 10,
			ConnMaxLifeTime: 600,
			SingularTable:   true,
			PrepareStmt:     true,
			Logger:          true,
		},
	}

	conf.WithLoggerConsole(true)
	conf.WithAutoMigrate(true)

	db, err := gormx.NewPostgres(conf, []interface{}{})
	if err != nil {
		panic(err)
	}

	_ = db.DB
}
```

## 配置说明

初始化配置为 gormx.Conf，MySQL/Postgres 的配置结构分别为 gormx.MysqlConf / gormx.PostgresConf（匿名嵌入 Conf）。

常用字段：
- Address：MySQL 为 host:port；Postgres 可为 host 或 host:port（未带端口默认 5432）
- Database/Username/Password：连接信息
- MaxOpenConnects/MaxIdleConnects/ConnMaxLifeTime：连接池（ConnMaxLifeTime 单位为秒，<=0 表示不限制）
- TablePrefix/SingularTable：命名策略
- DisableForeignKeyConstraintWhenMigrating：AutoMigrate 时不创建物理外键
- SkipDefaultTransaction：跳过 gorm 默认事务
- PrepareStmt：启用预处理语句缓存
- Logger：启用 SQL 日志（配合 WithLoggerConsole / WithLoggerHandle）

### TLS

当 Conf.Tls 同时配置了 CaCert / ClientCert / ClientCertKey 三个文件路径时启用 TLS，否则视为不启用：

```go
conf := &gormx.PostgresConf{
	Conf: gormx.Conf{
		Type:     gormx.Postgres,
		Address:  "127.0.0.1",
		Database: "demo",
		Username: "postgres",
		Password: "postgres",
		Tls: &gormx.TLS{
			CaCert:        "/path/to/ca.pem",
			ClientCert:    "/path/to/client.pem",
			ClientCertKey: "/path/to/client.key",
		},
	},
}
```

### SQL 日志回调

开启 Logger 后，你可以把 SQL 日志输出到控制台，或写入自定义回调：

```go
conf := &gormx.MysqlConf{
	Conf: gormx.Conf{
		Type:     gormx.Mysql,
		Address:  "127.0.0.1:3306",
		Database: "demo",
		Username: "root",
		Password: "root",
		Logger:   true,
	},
}

conf.WithLoggerConsole(true)
conf.WithLoggerHandle(func(b []byte) {
	_ = b
})
```

回调收到的是一段 JSON bytes。若 gorm 操作使用了 db.WithContext(ctx)，并且 ctx 中包含 gRPC incoming metadata，则会尝试从 metadata 里读取链路字段：
- gormx.TraceId（trace-id）
- gormx.UserId（user-id）
- gormx.AppId（app-id）

## 模型基类

gormx 提供了一组可直接嵌入的模型基类：
- gormx.Table：uint64 主键 + 软删除
- gormx.TableUnique：uint64 主键 + 软删除（DeletedAt 上 uniqueIndex:idx_unique）
- gormx.TableUUID：string 主键（UUIDv7）+ 软删除
- gormx.TableUUIDUnique：string 主键（UUIDv7）+ 软删除（DeletedAt 上 uniqueIndex:idx_unique）

### 分页 Scope

```go
import "github.com/fireflycore/gormx/scope"

db = db.Scopes(scope.WithPagination(1, 20))
```

分页规则：
- page 从 1 开始（0 会被修正为 1）
- size 范围为 [5, 100]（小于 5 修正为 5，大于 100 修正为 100）
