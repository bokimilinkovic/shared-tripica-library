package logging

// NewTestLogger creates instance of Logger which could be used in tests.
func NewTestLogger() *StdLogger {
	logger := New()
	logger.Out = &loggerOut{
		source: make(map[int]string),
	}
	return logger
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
