package logger

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ILogger interface {
	Info(msg string, fields ...map[string]interface{})
	Error(msg string, err error, fields ...map[string]interface{})
	Debug(msg string, fields ...map[string]interface{})
	Warn(msg string, fields ...map[string]interface{})
	Fatal(msg string, fields ...map[string]interface{})
	With(fields ...map[string]interface{}) ILogger
	LogWritter() io.Writer
	//zap.Logger
}

type Logger struct {
	logger    *zap.Logger
	beginTime time.Time
}

const (
	LogFormatJSON = "json"
	LogFormatText = "text"
)

func NewStandardLogger() (ILogger, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	return &Logger{logger: logger}, nil
}

// convert maps to zapFields
func toZapFields(fields []map[string]interface{}) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))
	for _, field := range fields {
		for k, v := range field {
			zapFields = append(zapFields, zap.Any(k, v))
		}
	}
	return zapFields
}

func (l *Logger) Info(msg string, fields ...map[string]interface{}) {
	l.logger.Info(msg, toZapFields(fields)...)
}

func (l *Logger) Error(msg string, err error, fields ...map[string]interface{}) {
	l.logger.With(zap.Error(err)).Error(msg, toZapFields(fields)...)
}

func (l *Logger) Debug(msg string, fields ...map[string]interface{}) {
	l.logger.Debug(msg, toZapFields(fields)...)
}

func (l *Logger) Warn(msg string, fields ...map[string]interface{}) {
	l.logger.Warn(msg, toZapFields(fields)...)
}

func (l *Logger) Fatal(msg string, fields ...map[string]interface{}) {
	l.logger.Fatal(msg, toZapFields(fields)...)
}

func (l *Logger) With(fields ...map[string]interface{}) ILogger {
	return &Logger{logger: l.logger.With(toZapFields(fields)...)}
}

func (l *Logger) LogWritter() io.Writer {
	zapLog := zap.NewStdLog(l.logger)
	lw := zapLog.Writer()
	return lw
}

func parseLogLevel(level string) (zapcore.Level, error) {
	switch level {

	case "info":
		return zapcore.InfoLevel, nil
	case "warn":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	case "debug":
		return zapcore.DebugLevel, nil
	case "panic":
		return zapcore.PanicLevel, nil
	case "fatal":
		return zapcore.FatalLevel, nil
	default:
		return zapcore.InfoLevel, errors.New("Invalid log level" + level)
	}
}

func NewLogger(level string, out, format string, debug bool) (ILogger, error) {
	var logger *zap.Logger

	zapLevel, err := parseLogLevel(level)
	if err != nil {
		return nil, err
	}
	cfg := zap.Config{
		Level:             zap.NewAtomicLevelAt(zapLevel),
		DisableStacktrace: true,
		Encoding:          format,
		OutputPaths:       []string{"stdout"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "file",
			FunctionKey:    "method",
			MessageKey:     "msg",
			StacktraceKey:  "stack trace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.RFC3339TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
	}
	logger, err = cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err

	}
	//logger = logger.With(zap.String("pkg", "server/main"))
	beginTime := time.Now().UTC()
	// adding host name field
	hostname, err := os.Hostname()
	if err != nil {
		logger.Error(fmt.Errorf("not able to get the hostname %s", err.Error()).Error())
	}
	return &Logger{logger: logger.WithOptions(
		zap.AddCallerSkip(0),
		zap.Fields(zap.String("hostname", hostname)),
		zap.Fields(zap.String("pkg", "server/main"))),

		beginTime: beginTime}, nil
}

func FormatErrorLog(requestID interface{}, error interface{}) string {
	return fmt.Sprintf("Request id : %v have error : %v", requestID, error)
}
