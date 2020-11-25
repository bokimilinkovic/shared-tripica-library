package logging_test

import (
	"os"
	"testing"

	"github.com/labstack/gommon/log"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"shared-tripica-library/logging"
)

func TestStdLogger_Configure(t *testing.T) {
	assert := assert.New(t)
	logger := createLogger()

	formatter := logging.NewUTCJSONFormatter()
	testCases := []struct {
		loggingLevel logging.Level
		echoLevel    log.Lvl
	}{
		{logging.DEBUG, log.DEBUG},
		{logging.INFO, log.INFO},
		{logging.WARNING, log.WARN},
		{logging.ERROR, log.ERROR},
		{logging.Level(0), log.INFO}, // invalid level
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(string(testCase.loggingLevel), func(t *testing.T) {
			logger.Configure(formatter, logger.Out, testCase.loggingLevel)

			assert.Equal(formatter, logger.Logger.Formatter)
			assert.Equal(testCase.echoLevel, logger.Level())
			assert.Equal(logger.Out, logger.Output())
		})
	}
}

func TestStdLogger_Output(t *testing.T) {
	assert := assert.New(t)
	logger := createLogger()

	out := logger.Output()

	assert.Equal(logger.Out, out)
}

func TestStdLogger_SetOutput(t *testing.T) {
	assert := assert.New(t)
	logger := createLogger()

	logger.SetOutput(os.Stdout)

	assert.Equal(os.Stdout, logger.Output())
}

func TestStdLogger_Level(t *testing.T) {
	assert := assert.New(t)
	logger := createLogger()

	t.Run("default level", func(t *testing.T) {
		level := logger.Level()

		assert.Equal(log.INFO, level)
	})

	t.Run("invalid level", func(t *testing.T) {
		logger.Logger.Level = logrus.FatalLevel
		level := logger.Level()

		assert.Equal(log.INFO, level)
	})
}

func TestStdLogger_SetLevel(t *testing.T) {
	assert := assert.New(t)
	logger := createLogger()

	testCases := []struct {
		param  log.Lvl
		result log.Lvl
	}{
		{log.DEBUG, log.DEBUG},
		{log.INFO, log.INFO},
		{log.WARN, log.WARN},
		{log.ERROR, log.ERROR},
		{log.Lvl(0), log.INFO}, // invalid level
	}

	for _, testCase := range testCases {
		testCase := testCase
		logger := logger
		t.Run(string(testCase.param), func(t *testing.T) {
			logger.SetLevel(testCase.param)

			assert.Equal(testCase.result, logger.Level())
		})
	}
}

func TestStdLogger_Prefix(t *testing.T) {
	assert := assert.New(t)
	logger := createLogger()

	prefix := logger.Prefix()

	assert.Empty(prefix)
	assert.Contains(logger.Out.(*loggerOut).source, "level=warning msg=\"Prefix() should not be called")
}

func TestStdLogger_SetPrefix(t *testing.T) {
	assert := assert.New(t)
	logger := createLogger()

	logger.SetPrefix("")

	assert.Contains(logger.Out.(*loggerOut).source, "level=warning msg=\"SetPrefix() should not be called")
}

func TestStdLogger_SetHeader(t *testing.T) {
	assert := assert.New(t)
	logger := createLogger()

	logger.SetHeader("test")

	assert.Contains(logger.Out.(*loggerOut).source, "level=warning msg=\"SetHeader() should not be called")
}

func TestStdLogger_Printj(t *testing.T) {
	assert := assert.New(t)
	logger := createLogger()

	logger.Printj(log.JSON{})

	assert.Contains(logger.Out.(*loggerOut).source, "level=warning msg=\"Printj() should not be called")
}

func TestStdLogger_Debugj(t *testing.T) {
	assert := assert.New(t)
	l := createLogger()

	l.Debugj(log.JSON{})

	assert.Contains(l.Out.(*loggerOut).source, "level=warning msg=\"Debugj() should not be called")
}

func TestStdLogger_Infoj(t *testing.T) {
	assert := assert.New(t)
	logger := createLogger()

	logger.Infoj(log.JSON{})

	assert.Contains(logger.Out.(*loggerOut).source, "level=warning msg=\"Infoj() should not be called")
}

func TestStdLogger_Warnj(t *testing.T) {
	assert := assert.New(t)
	logger := createLogger()

	logger.Warnj(log.JSON{})

	assert.Contains(logger.Out.(*loggerOut).source, "level=warning msg=\"Warnj() should not be called")
}

func TestStdLogger_Errorj(t *testing.T) {
	assert := assert.New(t)
	logger := createLogger()

	logger.Errorj(log.JSON{})

	assert.Contains(logger.Out.(*loggerOut).source, "level=warning msg=\"Errorj() should not be called")
}

func TestStdLogger_Fatalj(t *testing.T) {
	assert := assert.New(t)
	logger := createLogger()

	logger.Fatalj(log.JSON{})

	assert.Contains(logger.Out.(*loggerOut).source, "level=warning msg=\"Fatalj() should not be called")
}

func TestStdLogger_Panicj(t *testing.T) {
	assert := assert.New(t)
	logger := createLogger()

	logger.Panicj(log.JSON{})

	assert.Contains(logger.Out.(*loggerOut).source, "level=warning msg=\"Panicj() should not be called")
}

func TestCreateDefaultApplicationLogger(t *testing.T) {
	assert := assert.New(t)
	testCases := []struct {
		level   logrus.Level
		isDebug bool
	}{
		{logrus.DebugLevel, true},
		{logrus.InfoLevel, false},
	}

	for _, testCase := range testCases {
		logger := logging.CreateDefaultApplicationLogger(testCase.isDebug)
		assert.Equal(testCase.level, logger.GetLevel())
	}
}

func createLogger() *logging.StdLogger {
	logger := logging.New()
	logger.Out = &loggerOut{}

	return logger
}

type loggerOut struct {
	source string
}

func (l *loggerOut) Write(data []byte) (int, error) {
	l.source = string(data)
	return len(data), nil
}
