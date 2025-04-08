package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// 颜色常量定义
const (
	ColorReset    = "\033[0m"
	ColorYellow   = "\033[33m"
	ColorBlueBold = "\033[34;1m"
	ColorRedBold  = "\033[31;1m"
	LogTypeMongo  = "mongo"
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
	handle       func([]byte)
}

// New 创建并初始化一个新的日志记录器实例
func New(writer Writer, config Config, handle func([]byte)) Interface {
	baseFormat := "[RequestId:%d] [Timer:%.3fms]%s\n%s"
	traceStr := baseFormat
	traceWarnStr := baseFormat
	traceErrStr := "[RequestId:%d] [Timer:%.3fms] %s\n%s"

	if config.Colorful {
		colorPrefix := ColorBlueBold + "[RequestId:%d]" + ColorYellow
		traceStr = colorPrefix + " [%.3fms]\n" + ColorReset + "%s"
		traceWarnStr = colorPrefix + " [%.3fms] " + ColorYellow + "%s\n" + ColorReset + "%s"
		traceErrStr = colorPrefix + "[%.3fms] " + ColorRedBold + "%s\n" + ColorReset + " %s"
	}

	return &logger{
		Writer:       writer,
		Config:       config,
		traceStr:     traceStr,
		traceWarnStr: traceWarnStr,
		traceErrStr:  traceErrStr,
		handle:       handle,
	}
}

// Trace 记录跟踪日志
func (l *logger) Trace(_ context.Context, id int64, elapsed time.Duration, smt string, err string) {
	if l.LogLevel <= Silent {
		return
	}

	path := l.getCallerLocation(4)

	switch {
	case len(err) > 0 && l.LogLevel >= Error:
		l.Printf(l.traceErrStr, id, float64(elapsed.Nanoseconds())/1e6, err, smt)
		l.handleLog(Error, path, err, smt, elapsed)

	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= Warn:
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		l.Printf(l.traceWarnStr, id, float64(elapsed.Nanoseconds())/1e6, slowLog, smt)
		l.handleLog(Warn, path, slowLog, smt, elapsed)

	case l.LogLevel >= Info:
		l.Printf(l.traceStr, id, float64(elapsed.Nanoseconds())/1e6, smt)
		l.handleLog(Info, path, ResultSuccess, smt, elapsed)
	}
}

// 获取调用位置
func (l *logger) getCallerLocation(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown:0"
	}

	// 优化路径显示：去掉项目绝对路径前缀
	if idx := strings.Index(file, "/src/"); idx != -1 {
		file = file[idx+5:]
	} else if idx = strings.Index(file, "/pkg/mod/"); idx != -1 {
		file = file[idx+9:]
	}

	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}

// handleLog 统一处理日志记录
func (l *logger) handleLog(level LogLevel, path, smt, result string, elapsed time.Duration) {
	if l.handle != nil {
		logEntry := map[string]interface{}{
			"Statement": smt,
			"Result":    result,
			"Level":     level,
			"Duration":  elapsed.Milliseconds(),
			"Type":      LogTypeMongo,
			"Path":      path,
		}
		if b, err := json.Marshal(logEntry); err == nil {
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
