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

// Config 包含日志记录器的配置参数
type Config struct {
	SlowThreshold             time.Duration // 慢查询阈值
	Colorful                  bool          // 是否启用彩色输出
	IgnoreRecordNotFoundError bool          // 是否忽略记录未找到错误
	ParameterizedQueries      bool          // 是否记录参数化查询
	LogLevel                  LogLevel      // 日志记录级别
}

// Interface 日志记录器接口
type Interface interface {
	Trace(ctx context.Context, id int64, elapsed time.Duration, smt string, err string)
}

// logger 日志记录器实现
type logger struct {
	Writer
	Config
	traceStr     string
	traceWarnStr string
	traceErrStr  string
	database     string
	handle       func([]byte)
}

// New 创建并初始化一个新的日志记录器实例
func New(database string, writer Writer, config Config, handle func([]byte)) Interface {
	baseFormat := "[RequestId:%d] [Timer:%.3fms]%s\n%s"
	traceStr := baseFormat
	traceWarnStr := baseFormat
	traceErrStr := "[RequestId:%d] [Timer:%.3fms] %s\n%s"

	if config.Colorful {
		colorPrefix := ColorBlueBold + "[RequestId:%d]" + ColorYellow
		traceStr = colorPrefix + " [%.3fms]\n" + ColorReset + "%s\n"
		traceWarnStr = colorPrefix + " [%.3fms] " + ColorYellow + "%s\n" + ColorReset + "%s\n"
		traceErrStr = colorPrefix + "[%.3fms] " + ColorRedBold + "%s\n" + ColorReset + " %s\n"
	}

	return &logger{
		Writer:       writer,
		Config:       config,
		traceStr:     traceStr,
		traceWarnStr: traceWarnStr,
		traceErrStr:  traceErrStr,
		database:     database,
		handle:       handle,
	}
}

// Trace 记录跟踪日志
func (l *logger) Trace(ctx context.Context, id int64, elapsed time.Duration, smt string, err string) {
	if l.LogLevel <= Silent {
		return
	}

	path := FileWithLineNum()

	switch {
	case len(err) > 0 && l.LogLevel >= Error:
		l.Printf(l.traceErrStr, id, float64(elapsed.Nanoseconds())/1e6, err, smt)
		l.handleLog(ctx, 4, path, smt, err, elapsed)

	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= Warn:
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		l.Printf(l.traceWarnStr, id, float64(elapsed.Nanoseconds())/1e6, slowLog, smt)
		l.handleLog(ctx, 3, path, smt, slowLog, elapsed)

	case l.LogLevel >= Info:
		l.Printf(l.traceStr, id, float64(elapsed.Nanoseconds())/1e6, smt)
		l.handleLog(ctx, 1, path, smt, ResultSuccess, elapsed)
	}
}

// handleLog 统一处理日志记录
func (l *logger) handleLog(ctx context.Context, level LogLevel, path, smt, result string, elapsed time.Duration) {
	if l.handle != nil {
		logMap := map[string]interface{}{
			"Database":  l.database,
			"Statement": smt,
			"Result":    result,
			"Duration":  elapsed.Milliseconds(),
			"Level":     level,
			"Path":      path,
			"Type":      LogTypeMongo,
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

// CustomWriter 自定义日志写入器实现
type CustomWriter struct{}

// Write 实现io.Writer接口
func (cw *CustomWriter) Write(p []byte) (n int, err error) {
	// 这里可以添加自定义写入逻辑
	return len(p), nil
}
