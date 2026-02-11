package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc/metadata"
	loger "gorm.io/gorm/logger"
)

const (
	// ResultSuccess 表示成功执行 SQL 的结果标记
	ResultSuccess = "success"

	// Firefly系统自定义头部（统一前缀）
	HeaderPrefix = "x-firefly-"
	// TraceId 为从 metadata 读取 trace id 的 key
	TraceId = HeaderPrefix + "trace-id"
	// UserId 为从 metadata 读取 user id 的 key
	UserId = HeaderPrefix + "user-id"
	// AppId 为从 metadata 读取调用方 app id 的 key
	AppId = HeaderPrefix + "app-id"
)

// Config 为自定义 gorm logger 的配置
type Config struct {
	// Config 复用 gorm 内置 logger.Config（包含 SlowThreshold/LogLevel/Colorful 等）
	loger.Config

	// 控制台是否输出日志
	Console bool
	// 数据库
	Database string
	// 数据库类型
	DatabaseType int32
}

// NewLogger 构造一个 gorm logger，实现控制台输出与自定义回调输出
func NewLogger(config Config, handle func(b []byte)) loger.Interface {
	// 定义各级别日志输出模板（可选彩色）
	var (
		infoStr      = "[%s] [info] [Database:%s]\n%s\n%s"
		warnStr      = "[%s] [warn] [Database:%s]\n%s\n%s"
		errStr       = "[%s] [error] [Database:%s]\n%s\n%s"
		traceStr     = "[%s] [info] [Database:%s] [Rows:%v] [Duration:%.3fms] [Path:%s]\n%s"
		traceWarnStr = "[%s] [warn] [Database:%s] [Rows:%v] [Duration:%.3fms]	[Path:%s]\n%s\n%s"
		traceErrStr  = "[%s] [error] [Database:%s] [Rows:%v] [Duration:%.3fms] [Path:%s]\n%s\n%s"
	)

	// 若开启彩色输出，则替换模板为 gorm logger 预置的 ANSI 颜色字符串
	if config.Colorful {
		// date, level, db, file, msg
		infoStr = loger.BlueBold + "[%s] " + loger.BlueBold + "[info] " + loger.BlueBold + "[Database:%s]\n" + loger.Green + "%s\n" + loger.Reset + "%s"
		warnStr = loger.YellowBold + "[%s] " + loger.YellowBold + "[warn] " + loger.BlueBold + "[Database:%s]\n" + loger.Green + "%s\n" + loger.Yellow + "%s\n" + loger.Reset
		errStr = loger.RedBold + "[%s] " + loger.RedBold + "[error] " + loger.BlueBold + "[Database:%s]\n" + loger.Green + "%s\n" + loger.Red + "%s\n" + loger.Reset

		// date, level, db, rows, timer, file, sql
		traceStr = loger.BlueBold + "[%s] " + loger.BlueBold + "[info] " + loger.BlueBold + "[Database:%s] " + loger.YellowBold + "[Rows:%v]" + loger.Yellow + " [Duration:%.3fms]" + loger.Green + " [Path:%s]\n" + loger.Reset + "%s"
		// date, level, db, rows, timer, file, slowLog, sql
		traceWarnStr = loger.YellowBold + "[%s] " + loger.YellowBold + "[warn] " + loger.BlueBold + "[Database:%s] " + loger.YellowBold + "[Rows:%v]" + loger.Yellow + " [Duration:%.3fms]" + loger.Green + " [Path:%s]\n" + loger.Yellow + "%s\n" + loger.Reset + "%s"
		// date, level, db, rows, timer, file, err, sql
		traceErrStr = loger.RedBold + "[%s] " + loger.RedBold + "[error] " + loger.BlueBold + "[Database:%s] " + loger.YellowBold + "[Rows:%v]" + loger.Yellow + " [Duration:%.3fms]" + loger.Green + " [Path:%s]\n" + loger.Red + "%s\n" + loger.Reset + "%s"
	}

	// 返回实现 loger.Interface 的 logger 实例
	return &logger{
		// 复用 gorm logger 配置
		Config: config.Config,
		// 保存各级别输出模板
		infoStr:      infoStr,
		warnStr:      warnStr,
		errStr:       errStr,
		traceStr:     traceStr,
		traceWarnStr: traceWarnStr,
		traceErrStr:  traceErrStr,
		// handle 为自定义回调，可用于写入结构化日志
		handle: handle,
		// console 控制是否输出到控制台
		console: config.Console,
		// database 记录库名便于聚合检索
		database: config.Database,
		// databaseType 记录库类型便于聚合检索
		databaseType: config.DatabaseType,
	}
}

// logger 为内部实现，满足 gorm 的 logger.Interface
type logger struct {
	// Config 为 gorm logger 配置（嵌入以复用字段）
	loger.Config
	// infoStr/warnStr/errStr 为普通日志模板
	infoStr, warnStr, errStr string
	// traceStr/traceErrStr/traceWarnStr 为 SQL Trace 日志模板
	traceStr, traceErrStr, traceWarnStr string
	// database 为库名
	database string
	// databaseType 为库类型
	databaseType int32
	// console 控制台输出开关
	console bool
	// handle 为结构化日志回调
	handle func(b []byte)
}

// LogMode 设置日志级别，返回一个新的 logger（符合 gorm 约定）
func (l *logger) LogMode(level loger.LogLevel) loger.Interface {
	// 复制一份，避免修改原实例带来的并发问题
	newLogger := *l
	// 更新日志级别
	newLogger.LogLevel = level
	// 返回新实例
	return &newLogger
}

// Info 输出 info 日志（受 LogLevel 与 console 开关控制）
func (l *logger) Info(_ context.Context, msg string, data ...interface{}) {
	// 仅当启用控制台输出且 LogLevel 允许时才输出
	if l.console && l.LogLevel >= loger.Info {
		date := time.Now().Format(time.DateTime)
		file := fileWithLineNum()
		// msg 和 data 组合成完整消息
		fullMsg := fmt.Sprintf(msg, data...)
		// 输出到标准输出: date, db, file, msg
		fmt.Printf(l.infoStr+"\n", date, l.database, file, fullMsg)
	}
}

// Warn 输出 warn 日志（受 LogLevel 与 console 开关控制）
func (l *logger) Warn(_ context.Context, msg string, data ...interface{}) {
	// 仅当启用控制台输出且 LogLevel 允许时才输出
	if l.console && l.LogLevel >= loger.Warn {
		date := time.Now().Format(time.DateTime)
		file := fileWithLineNum()
		fullMsg := fmt.Sprintf(msg, data...)
		// 输出到标准输出: date, db, file, msg
		fmt.Printf(l.warnStr+"\n", date, l.database, file, fullMsg)
	}
}

// Error 输出 error 日志（受 LogLevel 与 console 开关控制）
func (l *logger) Error(_ context.Context, msg string, data ...interface{}) {
	// 仅当启用控制台输出且 LogLevel 允许时才输出
	if l.console && l.LogLevel >= loger.Error {
		date := time.Now().Format(time.DateTime)
		file := fileWithLineNum()
		fullMsg := fmt.Sprintf(msg, data...)
		// 输出到标准输出: date, db, file, msg
		fmt.Printf(l.errStr+"\n", date, l.database, file, fullMsg)
	}
}

// Trace 记录 SQL 执行信息（成功/慢 SQL/错误）
func (l *logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	// Silent 模式不输出任何日志
	if l.LogLevel <= loger.Silent {
		return
	}

	// elapsed 为 SQL 执行耗时
	elapsed := time.Since(begin)
	// 按错误/慢 SQL/普通 SQL 分支处理
	switch {
	case err != nil && l.LogLevel >= loger.Error && (!errors.Is(err, loger.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		// 从回调取出 SQL 文本与影响行数
		sql, rows := fc()
		// timer 为耗时的毫秒值（浮点便于输出 3 位小数）
		timer := float64(elapsed.Nanoseconds()) / 1e6
		// file 为调用位置
		file := fileWithLineNum()
		date := time.Now().Format(time.DateTime)

		// 控制台输出（若开启）
		if l.console {
			rowsStr := "-"
			if rows != -1 {
				rowsStr = fmt.Sprintf("%v", rows)
			}
			// traceErrStr expects: date, db, rows, timer, file, err, sql
			fmt.Printf(l.traceErrStr+"\n", date, l.database, rowsStr, timer, file, err, sql)
		}
		// 结构化回调输出
		l.handleLog(ctx, loger.Error, file, sql, err.Error(), elapsed)

	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= loger.Warn:
		// 从回调取出 SQL 文本与影响行数
		sql, rows := fc()
		// slowLog 为慢 SQL 标记文本
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		// timer 为耗时的毫秒值（浮点便于输出 3 位小数）
		timer := float64(elapsed.Nanoseconds()) / 1e6
		// file 为调用位置
		file := fileWithLineNum()
		date := time.Now().Format(time.DateTime)

		// 控制台输出（若开启）
		if l.console {
			rowsStr := "-"
			if rows != -1 {
				rowsStr = fmt.Sprintf("%v", rows)
			}
			// traceWarnStr expects: date, db, rows, timer, file, slowLog, sql
			fmt.Printf(l.traceWarnStr+"\n", date, l.database, rowsStr, timer, file, slowLog, sql)
		}
		// 结构化回调输出
		l.handleLog(ctx, loger.Warn, file, sql, slowLog, elapsed)

	case l.LogLevel == loger.Info:
		// 从回调取出 SQL 文本与影响行数
		sql, rows := fc()
		// timer 为耗时的毫秒值（浮点便于输出 3 位小数）
		timer := float64(elapsed.Nanoseconds()) / 1e6
		// file 为调用位置
		file := fileWithLineNum()
		date := time.Now().Format(time.DateTime)

		// 控制台输出（若开启）
		if l.console {
			rowsStr := "-"
			if rows != -1 {
				rowsStr = fmt.Sprintf("%v", rows)
			}
			// traceStr expects: date, db, rows, timer, file, sql
			fmt.Printf(l.traceStr+"\n", date, l.database, rowsStr, timer, file, sql)
		}
		// 结构化回调输出（成功分支）
		l.handleLog(ctx, loger.Info, file, sql, ResultSuccess, elapsed)
	}
}

// handleLog 将日志以 JSON 形式写入回调（若提供）
func (l *logger) handleLog(ctx context.Context, level loger.LogLevel, path, smt, result string, elapsed time.Duration) {
	// 未设置回调时直接返回
	if l.handle == nil {
		return
	}

	// logMap 为结构化日志内容，字段名保持相对稳定便于下游解析
	logMap := map[string]interface{}{
		"Database":  l.database,
		"Statement": smt,
		"Result":    result,
		"Duration":  elapsed.Microseconds(),
		"Level":     levelToInt(level),
		"Path":      path,
		"Type":      l.databaseType,
	}

	// 从 gRPC metadata 中提取链路字段（存在则写入结构化日志）
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

	// 将结构化日志序列化为 JSON，序列化失败则忽略
	if b, err := json.Marshal(logMap); err == nil {
		// 执行回调写入
		l.handle(b)
	}
}

func fileWithLineNum() string {
	for i := 2; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)
		if ok && (!strings.Contains(file, "gorm.io/gorm") || strings.HasSuffix(file, "_test.go")) && !strings.Contains(file, "gormx/internal/logger.go") {
			return file + ":" + strconv.FormatInt(int64(line), 10)
		}
	}
	return ""
}

func levelToInt(level loger.LogLevel) int {
	switch level {
	case loger.Info:
		return 1
	case loger.Warn:
		return 2
	case loger.Error:
		return 3
	default:
		return 0
	}
}
