package logging_test

import (
	"testing"
	"time"

	"shared-tripica-library/logging"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestUTCFormatter_Format(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name           string
		formatter      *logging.UTCFormatter
		formattedEntry string
	}{
		{
			name:           "JSON formatter entry time is in utc",
			formatter:      logging.NewUTCJSONFormatter(),
			formattedEntry: `{"level":"panic","msg":"","time":"2018-11-18T01:00:58Z"}`,
		},
		{
			name:           "Text formatter entry time is in utc",
			formatter:      logging.NewUTCTextFormatter(),
			formattedEntry: `time="2018-11-18T01:00:58Z" level=panic`,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			entry := logrus.NewEntry(logrus.New())
			location, err := time.LoadLocation("EST")
			if err != nil {
				t.Logf("unexpected error %v", err.Error())
				t.Fail()
			}
			entry.Time = time.Date(2018, 11, 17, 20, 00, 58, 651387237, location)

			data, err := test.formatter.Format(entry)
			if err != nil {
				t.Logf("unexpected error %v", err.Error())
				t.Fail()
			}

			assert.Contains(string(data), test.formattedEntry)
		})
	}
}
