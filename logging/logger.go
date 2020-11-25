package logging

import (
	"io"
	"os"

	"github.com/labstack/gommon/log"
	"github.com/sirupsen/logrus"
)

// Level is level type of the logger itself.
type Level uint8

const (
	// DEBUG is the most expressive log level.
	DEBUG Level = iota + 1
	// INFO comes after DEBUG.
	INFO
	// WARNING comes after INFO.
	WARNING
	// ERROR comes after WARNING.
	ERROR
)

// Logger represents the main logging interface.
type Logger interface {
	logrus.FieldLogger

	// Configure configures Logger.
	Configure(formatter logrus.Formatter, out io.Writer, level Level)
}

// StdLogger unifies two loggers implementations by implementing types logrus.FieldLogger and echo.Logger.
type StdLogger struct {
	*logrus.Logger
}

// New creates a new Logger instance.
func New() *StdLogger {
	return &StdLogger{
		Logger: logrus.New(),
	}
}

// Configure configures Logger.
func (l *StdLogger) Configure(formatter logrus.Formatter, out io.Writer, level Level) {
	l.Logger.SetFormatter(formatter)
	l.SetOutput(out)
	l.SetLevel(loggingToEchoLevel(level))
}

// Output returns place where the log will be written (part of echo.Logger interface).
func (l *StdLogger) Output() io.Writer {
	return l.Logger.Out
}

// SetOutput configures place where the log will be written (part of echo.Logger interface).
func (l *StdLogger) SetOutput(w io.Writer) {
	l.Logger.Out = w
}

// Level returns current log level (part of echo.Logger interface).
func (l *StdLogger) Level() log.Lvl {
	return logrusToEchoLevel(l.Logger.Level)
}

// SetLevel configures log level (part of echo.Logger interface).
func (l *StdLogger) SetLevel(level log.Lvl) {
	l.Logger.SetLevel(echoToLogrusLevel(level))
}

// Prefix is part of echo.Logger interface. It should not be called from a client code.
func (l *StdLogger) Prefix() string {
	l.Warn("Prefix() should not be called")
	return ""
}

// SetPrefix is part of echo.Logger interface. It should not be called from a client code.
func (l *StdLogger) SetPrefix(p string) {
	l.Warn("SetPrefix() should not be called")
}

// SetHeader is part of echo.Logger interface. It should not be called from a client code.
func (l *StdLogger) SetHeader(h string) {
	l.Warn("SetHeader() should not be called")
}

// Printj is part of echo.Logger interface. It should not be called from a client code.
func (l *StdLogger) Printj(j log.JSON) {
	l.Warn("Printj() should not be called")
}

// Debugj is part of echo.Logger interface. It should not be called from a client code.
func (l *StdLogger) Debugj(j log.JSON) {
	l.Warn("Debugj() should not be called")
}

// Infoj is part of echo.Logger interface. It should not be called from a client code.
func (l *StdLogger) Infoj(j log.JSON) {
	l.Warn("Infoj() should not be called")
}

// Warnj is part of echo.Logger interface. It should not be called from a client code.
func (l *StdLogger) Warnj(j log.JSON) {
	l.Warn("Warnj() should not be called")
}

// Errorj is part of echo.Logger interface. It should not be called from a client code.
func (l *StdLogger) Errorj(j log.JSON) {
	l.Warn("Errorj() should not be called")
}

// Fatalj is part of echo.Logger interface. It should not be called from a client code.
func (l *StdLogger) Fatalj(j log.JSON) {
	l.Warn("Fatalj() should not be called")
}

// Panicj is part of echo.Logger interface. It should not be called from a client code.
func (l *StdLogger) Panicj(j log.JSON) {
	l.Warn("Panicj() should not be called")
}

// CreateDefaultApplicationLogger creates default logger for an Application instance.
func CreateDefaultApplicationLogger(isDebugMode bool) *StdLogger {
	logger := New()

	// Ensure that logs are always in UTC.
	var formatter *UTCFormatter
	var level Level

	if isDebugMode {
		level = DEBUG
		formatter = NewUTCTextFormatter()
	} else {
		level = INFO
		formatter = NewUTCJSONFormatter()
	}

	logger.Configure(formatter, os.Stdout, level)
	return logger
}

// EchoLoggerFormat specifies the logger format to be used by Echo's logger middleware.
func EchoLoggerFormat() string {
	return `{"time":"${time_rfc3339_nano}","id":"${id}","remote_ip":"${remote_ip}","host":"${host}",` +
		`"method":"${method}","uri":"${uri}","path":"${path}","status":${status},"referer":"${referer}",` +
		`"user_agent":"${user_agent}","latency":${latency},"latency_human":"${latency_human}"` +
		`,"bytes_in":${bytes_in},"bytes_out":${bytes_out}}` + "\n\n"
}

// nolint: exhaustive
func echoToLogrusLevel(level log.Lvl) logrus.Level {
	switch level {
	case log.DEBUG:
		return logrus.DebugLevel
	case log.INFO:
		return logrus.InfoLevel
	case log.WARN:
		return logrus.WarnLevel
	case log.ERROR:
		return logrus.ErrorLevel
	default:
		return logrus.InfoLevel
	}
}

// nolint: exhaustive
func logrusToEchoLevel(level logrus.Level) log.Lvl {
	switch level {
	case logrus.DebugLevel:
		return log.DEBUG
	case logrus.InfoLevel:
		return log.INFO
	case logrus.WarnLevel:
		return log.WARN
	case logrus.ErrorLevel:
		return log.ERROR
	default:
		return log.INFO
	}
}

func loggingToEchoLevel(level Level) log.Lvl {
	switch level {
	case DEBUG:
		return log.DEBUG
	case INFO:
		return log.INFO
	case WARNING:
		return log.WARN
	case ERROR:
		return log.ERROR
	default:
		return log.INFO
	}
}
