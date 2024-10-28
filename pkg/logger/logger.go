package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
	Level        string
	Format       string
	Filename     string
	MaxSize      int
	MaxBackups   int
	MaxAge       int
	Compress     bool
	ReportCaller bool
}

type StandardFields struct {
	TraceID    string
	SpanID     string
	ServiceID  string
	InstanceID string
}

var (
	logger = logrus.New()
	fields StandardFields
)

func Init(cfg Config, serviceFields StandardFields) error {
	if cfg.Format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
		})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: time.RFC3339Nano,
			FullTimestamp:   true,
		})
	}

	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		return fmt.Errorf("parse log level error: %v", err)
	}
	logger.SetLevel(level)
	logger.SetReportCaller(cfg.ReportCaller)

	writers := []io.Writer{os.Stdout}
	if cfg.Filename != "" {
		fileWriter := &lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}
		writers = append(writers, fileWriter)
	}
	logger.SetOutput(io.MultiWriter(writers...))
	fields = serviceFields
	return nil
}

type Entry struct {
	*logrus.Entry
}

// getEntry 根据是否有ctx返回对应的Entry
func getEntry(ctx context.Context) *Entry {
	entry := logger.WithFields(logrus.Fields{
		"service_id":  fields.ServiceID,
		"instance_id": fields.InstanceID,
	})

	if ctx != nil {
		if span := opentracing.SpanFromContext(ctx); span != nil {
			spanCtx := span.Context()
			if spanCtx != nil {
				entry = entry.WithFields(logrus.Fields{
					"trace_id": fields.TraceID,
					"span_id":  fields.SpanID,
				})
			}
		}
	}

	if entry.Logger.ReportCaller {
		if pc, file, line, ok := runtime.Caller(2); ok {
			entry = entry.WithFields(logrus.Fields{
				"caller": fmt.Sprintf("%s:%d:%s",
					path.Base(file),
					line,
					runtime.FuncForPC(pc).Name(),
				),
			})
		}
	}

	return &Entry{entry}
}

// 不带context的日志方法
func Debug(args ...interface{}) {
	getEntry(nil).Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	getEntry(nil).Debugf(format, args...)
}

func Info(args ...interface{}) {
	getEntry(nil).Info(args...)
}

func Infof(format string, args ...interface{}) {
	getEntry(nil).Infof(format, args...)
}

func Warn(args ...interface{}) {
	getEntry(nil).Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	getEntry(nil).Warnf(format, args...)
}

func Error(args ...interface{}) {
	getEntry(nil).Error(args...)
}

func Errorf(format string, args ...interface{}) {
	getEntry(nil).Errorf(format, args...)
}

func Fatal(args ...interface{}) {
	getEntry(nil).Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	getEntry(nil).Fatalf(format, args...)
}

// 带context的日志方法
func DebugContext(ctx context.Context, args ...interface{}) {
	getEntry(ctx).Debug(args...)
}

func DebugContextf(ctx context.Context, format string, args ...interface{}) {
	getEntry(ctx).Debugf(format, args...)
}

func InfoContext(ctx context.Context, args ...interface{}) {
	getEntry(ctx).Info(args...)
}

func InfoContextf(ctx context.Context, format string, args ...interface{}) {
	getEntry(ctx).Infof(format, args...)
}

func WarnContext(ctx context.Context, args ...interface{}) {
	getEntry(ctx).Warn(args...)
}

func WarnContextf(ctx context.Context, format string, args ...interface{}) {
	getEntry(ctx).Warnf(format, args...)
}

func ErrorContext(ctx context.Context, args ...interface{}) {
	getEntry(ctx).Error(args...)
}

func ErrorContextf(ctx context.Context, format string, args ...interface{}) {
	getEntry(ctx).Errorf(format, args...)
}

func FatalContext(ctx context.Context, args ...interface{}) {
	getEntry(ctx).Fatal(args...)
}

func FatalContextf(ctx context.Context, format string, args ...interface{}) {
	getEntry(ctx).Fatalf(format, args...)
}

// Entry methods
func (e *Entry) WithField(key string, value interface{}) *Entry {
	return &Entry{e.Entry.WithField(key, value)}
}

func (e *Entry) WithFields(fields logrus.Fields) *Entry {
	return &Entry{e.Entry.WithFields(fields)}
}

func (e *Entry) WithError(err error) *Entry {
	return &Entry{e.Entry.WithError(err)}
}
