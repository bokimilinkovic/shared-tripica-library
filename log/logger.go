package log

type Logger interface {
	WithFields(map[string]interface{}) Logger
	Debug(v ...interface{})
}
