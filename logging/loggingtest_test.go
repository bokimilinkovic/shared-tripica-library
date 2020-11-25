package logging_test

import (
	"testing"

	"shared-tripica-library/logging"
	"github.com/stretchr/testify/assert"
)

func TestLoggerOut_Write(t *testing.T) {
	assert := assert.New(t)
	logger := logging.NewTestLogger()

	logger.Info("Hello")
	logger.Info("World")

	output := logger.GetOutput()

	assert.Contains(output[0], "level=info msg=Hello")
	assert.Contains(output[1], "level=info msg=World")
}
