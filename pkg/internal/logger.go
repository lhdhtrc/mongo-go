package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Colors
const (
	Reset       = "\033[0m"
	Green       = "\033[32m"
	Yellow      = "\033[33m"
	Magenta     = "\033[35m"
	BlueBold    = "\033[34;1m"
	MagentaBold = "\033[35;1m"
	RedBold     = "\033[31;1m"
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
	Trace(ctx context.Context, elapsed time.Duration, smt string, err string)
}

// New initialize logger
func New(writer Writer, config Config, handle func(b []byte)) Interface {
	var (
		traceStr     = "%s\n[%.3fms] [rows:%v] %s"
		traceWarnStr = "%s %s\n[%.3fms] [rows:%v] %s"
		traceErrStr  = "%s %s\n[%.3fms] [rows:%v] %s"
	)

	if config.Colorful {
		traceStr = Green + "%s\n" + Reset + Yellow + "[%.3fms] " + BlueBold + "[rows:%v]" + Reset + " %s"
		traceWarnStr = Green + "%s " + Yellow + "%s\n" + Reset + RedBold + "[%.3fms] " + Yellow + "[rows:%v]" + Magenta + " %s" + Reset
		traceErrStr = RedBold + "%s " + MagentaBold + "%s\n" + Reset + Yellow + "[%.3fms] " + BlueBold + "[rows:%v]" + Reset + " %s"
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
func (l *logger) Trace(_ context.Context, elapsed time.Duration, smt string, err string) {
	if l.LogLevel <= Silent {
		return
	}

	switch {
	case len(err) != 0 && l.LogLevel >= Error:
		file := FileWithLineNum()
		l.Printf(l.traceErrStr, file, err, float64(elapsed.Nanoseconds())/1e6, "-", smt)
		if l.handle != nil {
			logMap := make(map[string]interface{})
			logMap["Statement"] = smt
			logMap["Result"] = err
			logMap["Level"] = "error"
			logMap["Timer"] = elapsed.String()
			logMap["Type"] = Prefix
			logMap["Path"] = file
			b, _ := json.Marshal(logMap)
			l.handle(b)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= Warn:
		file := FileWithLineNum()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		l.Printf(l.traceWarnStr, file, slowLog, float64(elapsed.Nanoseconds())/1e6, "-", smt)
		if l.handle != nil {
			logMap := make(map[string]interface{})
			logMap["Statement"] = smt
			logMap["Result"] = slowLog
			logMap["Level"] = "warning"
			logMap["Timer"] = elapsed.String()
			logMap["Type"] = Prefix
			logMap["Path"] = file
			b, _ := json.Marshal(logMap)
			l.handle(b)
		}
	case l.LogLevel == Info:
		file := FileWithLineNum()
		l.Printf(l.traceStr, file, float64(elapsed.Nanoseconds())/1e6, "-", smt)
		if l.handle != nil {
			logMap := make(map[string]interface{})
			logMap["Statement"] = smt
			logMap["Result"] = "success"
			logMap["Level"] = "info"
			logMap["Timer"] = elapsed.String()
			logMap["Type"] = Prefix
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
