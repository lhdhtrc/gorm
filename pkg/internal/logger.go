package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"google.golang.org/grpc/metadata"
	"time"

	loger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

const (
	ResultSuccess = "success"
)

// Config 包含日志记录器的配置参数
type Config struct {
	loger.Config

	Console      bool   // 控制台是否输出日志
	Database     string // 数据库
	DatabaseType int32  // 数据库类型
}

// New initialize CustomLogger
func New(config Config, handle func(b []byte)) loger.Interface {
	var (
		infoStr      = "%s\n[info] "
		warnStr      = "%s\n[warn] "
		errStr       = "%s\n[error] "
		traceStr     = "%s\n[%.3fms] [rows:%v] %s"
		traceWarnStr = "%s %s\n[%.3fms] [rows:%v] %s"
		traceErrStr  = "%s %s\n[%.3fms] [rows:%v] %s"
	)

	if config.Colorful {
		infoStr = loger.Green + "%s\n" + loger.Reset + loger.Green + "[info] " + loger.Reset
		warnStr = loger.BlueBold + "%s\n" + loger.Reset + loger.Magenta + "[warn] " + loger.Reset
		errStr = loger.Magenta + "%s\n" + loger.Reset + loger.Red + "[error] " + loger.Reset
		traceStr = loger.Green + "%s\n" + loger.Reset + loger.Yellow + "[%.3fms] " + loger.BlueBold + "[rows:%v]" + loger.Reset + " %s"
		traceWarnStr = loger.Green + "%s " + loger.Yellow + "%s\n" + loger.Reset + loger.RedBold + "[%.3fms] " + loger.Yellow + "[rows:%v]" + loger.Magenta + " %s" + loger.Reset
		traceErrStr = loger.RedBold + "%s " + loger.MagentaBold + "%s\n" + loger.Reset + loger.Yellow + "[%.3fms] " + loger.BlueBold + "[rows:%v]" + loger.Reset + " %s"
	}

	return &logger{
		Config:       config.Config,
		infoStr:      infoStr,
		warnStr:      warnStr,
		errStr:       errStr,
		traceStr:     traceStr,
		traceWarnStr: traceWarnStr,
		traceErrStr:  traceErrStr,
		handle:       handle,
		console:      config.Console,
		database:     config.Database,
		databaseType: config.DatabaseType,
	}
}

type logger struct {
	loger.Config
	infoStr, warnStr, errStr            string
	traceStr, traceErrStr, traceWarnStr string
	database                            string
	databaseType                        int32
	console                             bool
	handle                              func(b []byte)
}

// LogMode log mode
func (l *logger) LogMode(level loger.LogLevel) loger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

// Info print info
func (l *logger) Info(_ context.Context, msg string, data ...interface{}) {
	if l.console && l.LogLevel >= loger.Info {
		fmt.Println(fmt.Sprintf(l.infoStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...))
	}
}

// Warn print warn messages
func (l *logger) Warn(_ context.Context, msg string, data ...interface{}) {
	if l.console && l.LogLevel >= loger.Warn {
		fmt.Println(fmt.Sprintf(l.warnStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...))
	}
}

// Error print error messages
func (l *logger) Error(_ context.Context, msg string, data ...interface{}) {
	if l.console && l.LogLevel >= loger.Error {
		fmt.Println(fmt.Sprintf(l.errStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...))
	}
}

// Trace print sql message
func (l *logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= loger.Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= loger.Error && (!errors.Is(err, loger.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		sql, rows := fc()
		timer := float64(elapsed.Nanoseconds()) / 1e6
		file := utils.FileWithLineNum()
		if l.console {
			if rows == -1 {
				fmt.Println(fmt.Sprintf(l.traceErrStr, file, err, timer, "-", sql))
			} else {
				fmt.Println(fmt.Sprintf(l.traceErrStr, file, err, timer, rows, sql))
			}
		}
		l.handleLog(ctx, 4, file, sql, err.Error(), elapsed)

	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= loger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		timer := float64(elapsed.Nanoseconds()) / 1e6
		file := utils.FileWithLineNum()
		if l.console {
			if rows == -1 {
				fmt.Println(fmt.Sprintf(l.traceWarnStr, file, slowLog, timer, "-", sql))
			} else {
				fmt.Println(fmt.Sprintf(l.traceWarnStr, file, slowLog, timer, rows, sql))
			}
		}
		l.handleLog(ctx, 3, file, sql, slowLog, elapsed)

	case l.LogLevel == loger.Info:
		sql, rows := fc()
		timer := float64(elapsed.Nanoseconds()) / 1e6
		file := utils.FileWithLineNum()
		if l.console {
			if rows == -1 {
				fmt.Println(fmt.Sprintf(l.traceStr, file, timer, "-", sql))
			} else {
				fmt.Println(fmt.Sprintf(l.traceStr, file, timer, rows, sql))
			}
		}
		l.handleLog(ctx, 1, file, sql, ResultSuccess, elapsed)
	}
}

// handleLog 统一处理日志记录
func (l *logger) handleLog(ctx context.Context, level loger.LogLevel, path, smt, result string, elapsed time.Duration) {
	if l.handle != nil {
		logMap := map[string]interface{}{
			"Database":  l.database,
			"Statement": smt,
			"Result":    result,
			"Duration":  elapsed.Milliseconds(),
			"Level":     level,
			"Path":      path,
			"Type":      l.databaseType,
		}
		md, _ := metadata.FromIncomingContext(ctx)
		if gd := md.Get("trace-id"); len(gd) != 0 {
			logMap["trace_id"] = gd[0]
		}
		if gd := md.Get("account-id"); len(gd) != 0 {
			logMap["account_id"] = gd[0]
		}
		if gd := md.Get("app-id"); len(gd) != 0 {
			logMap["invoke_app_id"] = gd[0]
		}
		if b, err := json.Marshal(logMap); err == nil {
			l.handle(b)
		}
	}
}
