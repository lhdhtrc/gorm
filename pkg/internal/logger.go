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

// New initialize CustomLogger
func New(prefix string, writer loger.Writer, config loger.Config, handle func(b []byte)) loger.Interface {
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

	return &CustomLogger{
		Writer:       writer,
		Config:       config,
		infoStr:      infoStr,
		warnStr:      warnStr,
		errStr:       errStr,
		traceStr:     traceStr,
		traceWarnStr: traceWarnStr,
		traceErrStr:  traceErrStr,
		handle:       handle,
		prefix:       prefix,
	}
}

type CustomLogger struct {
	loger.Writer
	loger.Config
	infoStr, warnStr, errStr            string
	traceStr, traceErrStr, traceWarnStr string
	handle                              func(b []byte)
	prefix                              string
}

// LogMode log mode
func (l *CustomLogger) LogMode(level loger.LogLevel) loger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

// Info print info
func (l *CustomLogger) Info(_ context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= loger.Info {
		l.Printf(l.infoStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

// Warn print warn messages
func (l *CustomLogger) Warn(_ context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= loger.Warn {
		l.Printf(l.warnStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

// Error print error messages
func (l *CustomLogger) Error(_ context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= loger.Error {
		l.Printf(l.errStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

// Trace print sql message
func (l *CustomLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= loger.Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= loger.Error && (!errors.Is(err, loger.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		sql, rows := fc()
		timer := float64(elapsed.Nanoseconds()) / 1e6
		file := utils.FileWithLineNum()
		if rows == -1 {
			l.Printf(l.traceErrStr, file, err, timer, "-", sql)
		} else {
			l.Printf(l.traceErrStr, file, err, timer, rows, sql)
		}
		if l.handle != nil {
			logMap := make(map[string]interface{})
			logMap["statement"] = sql
			logMap["result"] = err.Error()
			logMap["level"] = "error"
			logMap["timer"] = fmt.Sprintf("%.3fms", timer)
			logMap["type"] = l.prefix
			logMap["path"] = file

			md, err := metadata.FromIncomingContext(ctx)
			fmt.Println(md, err)
			if gd := md.Get("trace-id"); len(gd) != 0 {
				logMap["trace_id"] = gd[0]
			}
			if gd := md.Get("account-id"); len(gd) != 0 {
				logMap["account_id"] = gd[0]
			}
			if gd := md.Get("app-id"); len(gd) != 0 {
				logMap["invoke_app_id"] = gd[0]
			}

			b, _ := json.Marshal(logMap)
			l.handle(b)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= loger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		timer := float64(elapsed.Nanoseconds()) / 1e6
		file := utils.FileWithLineNum()
		if rows == -1 {
			l.Printf(l.traceWarnStr, file, slowLog, timer, "-", sql)
		} else {
			l.Printf(l.traceWarnStr, file, slowLog, timer, rows, sql)
		}
		if l.handle != nil {
			logMap := make(map[string]interface{})
			logMap["Statement"] = sql
			logMap["Result"] = slowLog
			logMap["Level"] = "warning"
			logMap["Timer"] = fmt.Sprintf("%.3fms", timer)
			logMap["Type"] = l.prefix
			logMap["Path"] = file
			b, _ := json.Marshal(logMap)
			l.handle(b)
		}
	case l.LogLevel == loger.Info:
		sql, rows := fc()
		timer := float64(elapsed.Nanoseconds()) / 1e6
		file := utils.FileWithLineNum()
		if rows == -1 {
			l.Printf(l.traceStr, file, timer, "-", sql)
		} else {
			l.Printf(l.traceStr, file, timer, rows, sql)
		}
		if l.handle != nil {
			logMap := make(map[string]interface{})
			logMap["Statement"] = sql
			logMap["Result"] = "success"
			logMap["Level"] = "info"
			logMap["Timer"] = fmt.Sprintf("%.3fms", timer)
			logMap["Type"] = l.prefix
			logMap["Path"] = file
			b, _ := json.Marshal(logMap)
			l.handle(b)
		}
	}
}

type CustomWriter struct{}

func (cw *CustomWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}
