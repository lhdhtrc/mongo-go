package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc/metadata"
	"time"
)

// 颜色常量定义
const (
	ColorReset    = "\033[0m"
	ColorYellow   = "\033[33m"
	ColorBlueBold = "\033[34;1m"
	ColorRedBold  = "\033[31;1m"

	LogTypeMongo  = 6 // 定义于proto中
	ResultSuccess = "success"

	TraceId = "trace-id"
	UserId  = "user-id"
	AppId   = "app-id"
)

// LogLevel 定义日志级别类型
type LogLevel int

const (
	Silent LogLevel = iota + 1 // 静默模式，不记录任何日志
	Error                      // 仅记录错误日志
	Warn                       // 记录警告和错误日志
	Info                       // 记录所有日志
)

// Writer 日志写入器接口
type Writer interface {
	Printf(string, ...interface{})
}

// Conf 包含日志记录器的配置参数
type Conf struct {
	Console                   bool          // 控制台是否输出
	SlowThreshold             time.Duration // 慢查询阈值
	Colorful                  bool          // 是否启用彩色输出
	IgnoreRecordNotFoundError bool          // 是否忽略记录未找到错误
	ParameterizedQueries      bool          // 是否记录参数化查询

	Database string   // 数据库
	LogLevel LogLevel // 日志记录级别
}

// Interface 日志记录器接口
type Interface interface {
	Trace(ctx context.Context, id int64, elapsed time.Duration, smt string, err string)
}

// logger 日志记录器实现
type logger struct {
	Writer
	Conf
	traceStr     string
	traceWarnStr string
	traceErrStr  string
	handle       func([]byte)
}

// New 创建并初始化一个新的日志记录器实例
func New(conf Conf, handle func([]byte)) Interface {
	baseFormat := "[%s] [%s] [Database:%s] [RequestId:%d] [Duration:%.3fms]%s\n%s"
	traceStr := baseFormat
	traceWarnStr := baseFormat
	traceErrStr := "[%s] [%s] [Database:%s] [RequestId:%d] [Duration:%.3fms] %s\n%s"

	if conf.Colorful {
		colorPrefix := "[%s] [%s] " + ColorBlueBold + "[Database:%s] " + ColorBlueBold + "[RequestId:%d] " + ColorYellow
		traceStr = colorPrefix + "[Duration:%.3fms]\n" + ColorReset + "%s"
		traceWarnStr = colorPrefix + "[Duration:%.3fms] " + ColorYellow + "%s\n" + ColorReset + "%s"
		traceErrStr = colorPrefix + "[Duration:%.3fms] " + ColorRedBold + "%s\n" + ColorReset + " %s"
	}

	return &logger{
		Conf:         conf,
		traceStr:     traceStr,
		traceWarnStr: traceWarnStr,
		traceErrStr:  traceErrStr,
		handle:       handle,
	}
}

// Trace 记录跟踪日志
func (l *logger) Trace(ctx context.Context, id int64, elapsed time.Duration, smt string, err string) {
	if l.LogLevel <= Silent {
		return
	}

	date := time.Now().Format(time.DateTime)

	switch {
	case len(err) > 0 && l.LogLevel >= Error:
		if l.Console {
			fmt.Println(fmt.Sprintf(l.traceErrStr, date, "error", l.Database, id, float64(elapsed.Nanoseconds())/1e6, err, smt))
		}
		l.handleLog(ctx, 4, smt, err, elapsed)

	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= Warn:
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if l.Console {
			fmt.Println(fmt.Sprintf(l.traceWarnStr, date, "warn", l.Database, id, float64(elapsed.Nanoseconds())/1e6, slowLog, smt))
		}
		l.handleLog(ctx, 3, smt, slowLog, elapsed)

	case l.LogLevel >= Info:
		if l.Console {
			fmt.Println(fmt.Sprintf(l.traceStr, date, "info", l.Database, id, float64(elapsed.Nanoseconds())/1e6, smt))
		}
		l.handleLog(ctx, 1, smt, ResultSuccess, elapsed)
	}
}

// handleLog 统一处理日志记录
func (l *logger) handleLog(ctx context.Context, level LogLevel, smt, result string, elapsed time.Duration) {
	if l.handle != nil {
		logMap := map[string]interface{}{
			"Database":  l.Database,
			"Statement": smt,
			"Result":    result,
			"Duration":  elapsed.Microseconds(),
			"Level":     level,
			"Type":      LogTypeMongo,
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
}
