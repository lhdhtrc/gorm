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
		Config: gormx.Config{
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
		Config: gormx.Config{
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

### 分页 Scope

```go
import "github.com/fireflycore/gormx/scope"

db = db.Scopes(scope.WithPagination(1, 20))
```
