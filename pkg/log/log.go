// Provides a generic interface for logging
package logging

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/tim-beatham/smegmesh/pkg/conf"
)

var (
	Log Logger
)

type Logger interface {
	WriteInfof(msg string, args ...interface{})
	WriteErrorf(msg string, args ...interface{})
	WriteWarnf(msg string, args ...interface{})
	Writer() io.Writer
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

func (l *LogrusLogger) Writer() io.Writer {
	return l.logger.Writer()
}

func NewLogrusLogger(confLevel conf.LogLevel) *LogrusLogger {

	var level logrus.Level

	switch confLevel {
	case conf.ERROR:
		level = logrus.ErrorLevel
	case conf.WARNING:
		level = logrus.WarnLevel
	case conf.INFO:
		level = logrus.InfoLevel
	}

	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(level)

	return &LogrusLogger{logger: logger}
}

func init() {
	SetLogger(NewLogrusLogger(conf.INFO))
}

func SetLogger(l Logger) {
	Log = l
}
