package logger

import (
	"strings"

	"github.com/sirupsen/logrus"
)

type Logger interface {
	logrus.FieldLogger
}

type wrap struct {
	*logrus.Logger
}

const (
	LogFormatJSON = "json"
	LogFormatText = "text"
)

func NewLogrusStandardLogger() Logger {
	return &wrap{Logger: logrus.StandardLogger()}
}

// SetupLogger to setup log config
func SetupLogger(level string, out, format string, debug bool) (Logger, error) {

	l := logrus.StandardLogger()
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return nil, err
	}
	l.SetLevel(logLevel)

	var formatter logrus.Formatter
	switch strings.ToLower(format) {
	case LogFormatJSON:
		formatter = &logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000",
		}
		if debug {
			formatter.(*logrus.JSONFormatter).PrettyPrint = true
		}
	case LogFormatText:
		formatter = &logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02T15:04:05.000Z",
		}
	}
	if debug {
		l.SetReportCaller(true)
	}

	l.SetFormatter(formatter)
	return &wrap{Logger: l}, nil
}
