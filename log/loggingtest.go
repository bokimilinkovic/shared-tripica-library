package log

import "github.com/sirupsen/logrus"

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

// NewTestLogger creates instance of Logger which could be used in tests.
func NewTestLogger() *StdLogger {
	logger := New()
	logger.Out = &loggerOut{
		source: make(map[int]string),
	}

	return logger
}

func (l *StdLogger) WithFields(fields map[string]interface{}) Logger {
	log := logrus.New()
	log.WithFields(fields)

	return &StdLogger{log}
}

// GetOutput returns all logged content, which could be easily used in assertions.
func (l *StdLogger) GetOutput() map[int]string {
	return l.Out.(*loggerOut).source
}

type loggerOut struct {
	source map[int]string
}

// Write writes given data to the source field.
func (l *loggerOut) Write(data []byte) (int, error) {
	l.source[len(l.source)] = string(data)

	return len(data), nil
}
