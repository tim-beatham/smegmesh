// Provides a generic interface for logging
package logging

import (
	"os"

	"github.com/sirupsen/logrus"
)

var (
	Log Logger
)

type Logger interface {
	WriteInfof(msg string, args ...interface{})
	WriteErrorf(msg string, args ...interface{})
	WriteWarnf(msg string, args ...interface{})
}

type LogrusLogger struct {
	logger *logrus.Logger
}

func (l *LogrusLogger) WriteInfof(msg string, args ...interface{}) {
	l.logger.Infof(msg, args...)
}

func (l *LogrusLogger) WriteErrorf(msg string, args ...interface{}) {
	l.logger.Errorf(msg, args...)
}

func (l *LogrusLogger) WriteWarnf(msg string, args ...interface{}) {
	l.logger.Warnf(msg, args...)
}

func NewLogrusLogger() *LogrusLogger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)

	return &LogrusLogger{logger: logger}
}

func init() {
	SetLogger(NewLogrusLogger())
}

func SetLogger(l Logger) {
	Log = l
}
