package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc/metadata"
	loger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

const (
	ResultSuccess = "success"

	TraceId = "trace-id"
	UserId  = "user-id"
	AppId   = "app-id"
)

type Config struct {
	loger.Config

	// 控制台是否输出日志
	Console bool
	// 数据库
	Database string
	// 数据库类型
	DatabaseType int32
}

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

func (l *logger) LogMode(level loger.LogLevel) loger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

func (l *logger) Info(_ context.Context, msg string, data ...interface{}) {
	if l.console && l.LogLevel >= loger.Info {
		args := append([]interface{}{utils.FileWithLineNum()}, data...)
		fmt.Printf(l.infoStr+msg+"\n", args...)
	}
}

func (l *logger) Warn(_ context.Context, msg string, data ...interface{}) {
	if l.console && l.LogLevel >= loger.Warn {
		args := append([]interface{}{utils.FileWithLineNum()}, data...)
		fmt.Printf(l.warnStr+msg+"\n", args...)
	}
}

func (l *logger) Error(_ context.Context, msg string, data ...interface{}) {
	if l.console && l.LogLevel >= loger.Error {
		args := append([]interface{}{utils.FileWithLineNum()}, data...)
		fmt.Printf(l.errStr+msg+"\n", args...)
	}
}

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
				fmt.Printf(l.traceErrStr+"\n", file, err, timer, "-", sql)
			} else {
				fmt.Printf(l.traceErrStr+"\n", file, err, timer, rows, sql)
			}
		}
		l.handleLog(ctx, loger.Error, file, sql, err.Error(), elapsed)

	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= loger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		timer := float64(elapsed.Nanoseconds()) / 1e6
		file := utils.FileWithLineNum()
		if l.console {
			if rows == -1 {
				fmt.Printf(l.traceWarnStr+"\n", file, slowLog, timer, "-", sql)
			} else {
				fmt.Printf(l.traceWarnStr+"\n", file, slowLog, timer, rows, sql)
			}
		}
		l.handleLog(ctx, loger.Warn, file, sql, slowLog, elapsed)

	case l.LogLevel == loger.Info:
		sql, rows := fc()
		timer := float64(elapsed.Nanoseconds()) / 1e6
		file := utils.FileWithLineNum()
		if l.console {
			if rows == -1 {
				fmt.Printf(l.traceStr+"\n", file, timer, "-", sql)
			} else {
				fmt.Printf(l.traceStr+"\n", file, timer, rows, sql)
			}
		}
		l.handleLog(ctx, loger.Info, file, sql, ResultSuccess, elapsed)
	}
}

func (l *logger) handleLog(ctx context.Context, level loger.LogLevel, path, smt, result string, elapsed time.Duration) {
	if l.handle == nil {
		return
	}

	logMap := map[string]interface{}{
		"Database":  l.database,
		"Statement": smt,
		"Result":    result,
		"Duration":  elapsed.Microseconds(),
		"Level":     level,
		"Path":      path,
		"Type":      l.databaseType,
	}

	md, _ := metadata.FromIncomingContext(ctx)
	if gd := md.Get(TraceId); len(gd) != 0 {
		logMap["trace_id"] = gd[0]
	}
	if gd := md.Get(UserId); len(gd) != 0 {
		logMap["user_id"] = gd[0]
	}
	if gd := md.Get(AppId); len(gd) != 0 {
		logMap["invoke_app_id"] = gd[0]
	}

	if b, err := json.Marshal(logMap); err == nil {
		l.handle(b)
	}
}
