// logger/interface.go
package logger

//
//import (
//	"context"
//	"io"
//)
//
//// Logger 定义统一的日志接口
//type Logger interface {
//	// 基础日志方法(不带ctx)
//	Debug(args ...interface{})
//	Info(args ...interface{})
//	Warn(args ...interface{})
//	Error(args ...interface{})
//	Fatal(args ...interface{})
//
//	// 格式化日志方法(不带ctx)
//	Debugf(format string, args ...interface{})
//	Infof(format string, args ...interface{})
//	Warnf(format string, args ...interface{})
//	Errorf(format string, args ...interface{})
//	Fatalf(format string, args ...interface{})
//
//	// 带ctx的日志方法
//	DebugContext(ctx context.Context, args ...interface{})
//	InfoContext(ctx context.Context, args ...interface{})
//	WarnContext(ctx context.Context, args ...interface{})
//	ErrorContext(ctx context.Context, args ...interface{})
//	FatalContext(ctx context.Context, args ...interface{})
//
//	// 带ctx的格式化日志方法
//	DebugContextf(ctx context.Context, format string, args ...interface{})
//	InfoContextf(ctx context.Context, format string, args ...interface{})
//	WarnContextf(ctx context.Context, format string, args ...interface{})
//	ErrorContextf(ctx context.Context, format string, args ...interface{})
//	FatalContextf(ctx context.Context, format string, args ...interface{})
//
//	// 链式调用方法
//	WithField(key string, value interface{}) Logger
//	WithFields(fields Fields) Logger
//	WithError(err error) Logger
//	WithContext(ctx context.Context) Logger
//}
//
//// Fields 定义字段类型
//type Fields map[string]interface{}
//
//// Config 日志配置结构
//type Config struct {
//	Level        string      // 日志级别
//	Format       string      // 日志格式 (json/text)
//	Outputs      []io.Writer // 输出目标
//	ServiceInfo  ServiceInfo // 服务信息
//	ReportCaller bool        // 是否记录调用信息
//	CallerSkip   int         // 调用栈跳过层数
//}
//
//// ServiceInfo 服务信息结构
//type ServiceInfo struct {
//	ServiceID  string
//	InstanceID string
//	Extra      Fields // 额外的服务级别字段
//}
//
//// FileOutputConfig 文件输出配置
//type FileOutputConfig struct {
//	Filename   string // 日志文件路径
//	MaxSize    int    // 单个日志文件最大尺寸(MB)
//	MaxBackups int    // 最大保留文件数
//	MaxAge     int    // 最大保留天数
//	Compress   bool   // 是否压缩
//}
