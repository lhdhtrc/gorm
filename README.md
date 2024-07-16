## Gorm Extension
Extension library based on gorm

### How to use it?
`go get github.com/lhdhtrc/gorm`

```go
package main

import (
	gorm "github.com/lhdhtrc/gorm/pkg"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	gorm.InstallMysql(logger, &gorm.ConfigEntity{}, []interface{})
}
```

### Finally
- If you feel good, click on star.
- If you have a good suggestion, please ask the issue.