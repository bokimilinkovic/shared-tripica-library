package log

type Logger interface {
	Warnf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	WithFields(map[string]interface{}) Logger
	Debug(v ...interface{})
}
