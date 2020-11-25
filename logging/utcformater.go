package logging

import (
	"github.com/sirupsen/logrus"
)

type formatter interface {
	Format(entry *logrus.Entry) ([]byte, error)
}

// UTCFormatter formats log in UTC.
type UTCFormatter struct {
	formatter formatter
}

// NewUTCJSONFormatter formats log as JSON in UTC.
func NewUTCJSONFormatter() *UTCFormatter {
	return &UTCFormatter{
		formatter: &logrus.JSONFormatter{},
	}
}

// NewUTCTextFormatter formats log as text in UTC.
func NewUTCTextFormatter() *UTCFormatter {
	return &UTCFormatter{
		formatter: &logrus.TextFormatter{
			DisableLevelTruncation: true,
			FullTimestamp:          true,
		},
	}
}

// Format renders a single log entry.
func (f *UTCFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	entry.Time = entry.Time.UTC()
	return f.formatter.Format(entry)
}
