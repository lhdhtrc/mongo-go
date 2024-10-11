package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Colors
const (
	Reset    = "\033[0m"
	Yellow   = "\033[33m"
	BlueBold = "\033[34;1m"
	RedBold  = "\033[31;1m"
)

// LogLevel log level
type LogLevel int

const (
	// Silent silent log level
	Silent LogLevel = iota + 1
	// Error error log level
	Error
	// Warn warn log level
	Warn
	// Info info log level
	Info
)

const Prefix = "mongo"

// Writer log writer interface
type Writer interface {
	Printf(string, ...interface{})
}

// Config logger config
type Config struct {
	SlowThreshold             time.Duration
	Colorful                  bool
	IgnoreRecordNotFoundError bool
	ParameterizedQueries      bool
	LogLevel                  LogLevel
}

// Interface logger interface
type Interface interface {
	Trace(ctx context.Context, id int64, elapsed time.Duration, smt string, err string)
}

// New initialize logger
func New(writer Writer, config Config, handle func(b []byte)) Interface {
	var (
		traceStr     = "[RequestId:%d] [Timer:%.3fms]\n%s"
		traceWarnStr = "[RequestId:%d] [Timer:%.3fms] %s\n%s"
		traceErrStr  = "[RequestId:%d] [Timer:%.3fms] %s\n%s"
	)

	if config.Colorful {
		traceStr = BlueBold + "[RequestId:%d]" + Yellow + " [%.3fms]\n" + Reset + "%s"
		traceWarnStr = BlueBold + "[RequestId:%d]" + Yellow + " [%.3fms] " + Yellow + "%s\n" + Reset + "%s" + Reset
		traceErrStr = BlueBold + "[RequestId:%d]" + Yellow + "[%.3fms] " + RedBold + "%s\n" + Reset + " %s"
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

type logger struct {
	Writer
	Config
	traceStr, traceErrStr, traceWarnStr string

	handle func(b []byte)
}

// Trace print sql message
func (l *logger) Trace(_ context.Context, id int64, elapsed time.Duration, smt string, err string) {
	if l.LogLevel <= Silent {
		return
	}

	switch {
	case len(err) != 0 && l.LogLevel >= Error:
		l.Printf(l.traceErrStr, id, float64(elapsed.Nanoseconds())/1e6, err, smt)
		if l.handle != nil {
			logMap := make(map[string]interface{})
			logMap["Statement"] = smt
			logMap["Result"] = err
			logMap["Level"] = "error"
			logMap["Timer"] = elapsed.String()
			logMap["Type"] = Prefix
			logMap["Path"] = ""
			b, _ := json.Marshal(logMap)
			l.handle(b)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= Warn:
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		l.Printf(l.traceWarnStr, id, float64(elapsed.Nanoseconds())/1e6, slowLog, smt)
		if l.handle != nil {
			logMap := make(map[string]interface{})
			logMap["Statement"] = smt
			logMap["Result"] = slowLog
			logMap["Level"] = "warning"
			logMap["Timer"] = elapsed.String()
			logMap["Type"] = Prefix
			logMap["Path"] = ""
			b, _ := json.Marshal(logMap)
			l.handle(b)
		}
	case l.LogLevel == Info:
		l.Printf(l.traceStr, id, float64(elapsed.Nanoseconds())/1e6, smt)
		if l.handle != nil {
			logMap := make(map[string]interface{})
			logMap["Statement"] = smt
			logMap["Result"] = "success"
			logMap["Level"] = "info"
			logMap["Timer"] = elapsed.String()
			logMap["Type"] = Prefix
			logMap["Path"] = ""
			b, _ := json.Marshal(logMap)
			l.handle(b)
		}
	}
}

type CustomWriter struct{}

func (cw *CustomWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}
