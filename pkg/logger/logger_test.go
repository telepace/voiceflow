// // logger/logger_test.go
package logger

//
//import (
//	"bytes"
//	"context"
//	"encoding/json"
//	"errors"
//	"github.com/opentracing/opentracing-go"
//	"github.com/stretchr/testify/assert"
//	"github.com/stretchr/testify/mock"
//	"io"
//	"os"
//	"testing"
//)
//
//// mockWriter 用于捕获日志输出
//type mockWriter struct {
//	bytes.Buffer
//}
//
//func (w *mockWriter) Close() error {
//	return nil
//}
//
//// mockSpan 模拟 opentracing.Span
//type mockSpan struct {
//	mock.Mock
//}
//
//func (m *mockSpan) Context() *mockSpanContext {
//	return &mockSpanContext{
//		traceID: "test-trace-id",
//		spanID:  "test-span-id",
//	}
//}
//x3
//// 实现其他必要的 opentracing.Span 接口方法...
//
//type mockSpanContext struct {
//	traceID string
//	spanID  string
//}
//
//func (m *mockSpanContext) TraceID() string { return m.traceID }
//func (m *mockSpanContext) SpanID() string  { return m.spanID }
//
//// TestLoggerCreation 测试日志实例创建
//func TestLoggerCreation(t *testing.T) {
//	tests := []struct {
//		name    string
//		cfg     Config
//		wantErr bool
//	}{
//		{
//			name: "valid config",
//			cfg: Config{
//				Level:  "info",
//				Format: "json",
//				ServiceInfo: ServiceInfo{
//					ServiceID:  "test-service",
//					InstanceID: "test-instance",
//				},
//			},
//			wantErr: false,
//		},
//		{
//			name: "invalid level",
//			cfg: Config{
//				Level:  "invalid",
//				Format: "json",
//			},
//			wantErr: true,
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			logger, err := NewLogger(tt.cfg)
//			if tt.wantErr {
//				assert.Error(t, err)
//				assert.Nil(t, logger)
//			} else {
//				assert.NoError(t, err)
//				assert.NotNil(t, logger)
//			}
//		})
//	}
//}
//
//// TestBasicLogging 测试基本日志功能
//func TestBasicLogging(t *testing.T) {
//	buf := &mockWriter{}
//	logger, err := NewLogger(Config{
//		Level:   "debug",
//		Format:  "json",
//		Outputs: []io.Writer{buf},
//		ServiceInfo: ServiceInfo{
//			ServiceID:  "test-service",
//			InstanceID: "test-instance",
//		},
//	})
//	assert.NoError(t, err)
//
//	tests := []struct {
//		name     string
//		logFunc  func()
//		checkLog func(t *testing.T, log map[string]interface{})
//	}{
//		{
//			name: "basic info log",
//			logFunc: func() {
//				logger.Info("test message")
//			},
//			checkLog: func(t *testing.T, log map[string]interface{}) {
//				assert.Equal(t, "info", log["level"])
//				assert.Equal(t, "test message", log["msg"])
//				assert.Equal(t, "test-service", log["service_id"])
//			},
//		},
//		{
//			name: "log with fields",
//			logFunc: func() {
//				logger.WithField("key", "value").Info("test with field")
//			},
//			checkLog: func(t *testing.T, log map[string]interface{}) {
//				assert.Equal(t, "info", log["level"])
//				assert.Equal(t, "test with field", log["msg"])
//				assert.Equal(t, "value", log["key"])
//			},
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			buf.Reset()
//			tt.logFunc()
//
//			var logMap map[string]interface{}
//			err := json.Unmarshal(buf.Bytes(), &logMap)
//			assert.NoError(t, err)
//			tt.checkLog(t, logMap)
//		})
//	}
//}
//
//// Finish() FinishWithOptions(opts FinishOptions) SetOperationName(operationName string) Span
//func (m *mockSpan) Finish() {
//	m.Called()
//}
//
//func (m *mockSpan) FinishWithOptions(opts opentracing.FinishOptions) {
//	m.Called(opts)
//}
//
//func (m *mockSpan) SetOperationName(operationName string) opentracing.Span {
//	args := m.Called(operationName)
//	return args.Get(0).(opentracing.Span)
//}
////
////// TestContextLogging 测试带 Context 的日志功能
////func TestContextLogging(t *testing.T) {
////	buf := &mockWriter{}
////	logger, err := NewLogger(Config{
////		Level:   "debug",
////		Format:  "json",
////		Outputs: []io.Writer{buf},
////		ServiceInfo: ServiceInfo{
////			ServiceID: "test-service",
////		},
////	})
////	assert.NoError(t, err)
////
////	// 创建带追踪信息的 Context
////	span := &mockSpan{}
////	ctx := opentracing.ContextWithSpan(context.Background(), span)
////
////	tests := []struct {
////		name     string
////		logFunc  func()
////		checkLog func(t *testing.T, log map[string]interface{})
////	}{
////		{
////			name: "context with trace info",
////			logFunc: func() {
////				logger.InfoContext(ctx, "test with context")
////			},
////			checkLog: func(t *testing.T, log map[string]interface{}) {
////				assert.Equal(t, "info", log["level"])
////				assert.Equal(t, "test with context", log["msg"])
////				assert.Equal(t, "test-trace-id", log["trace_id"])
////				assert.Equal(t, "test-span-id", log["span_id"])
////			},
////		},
////	}
////
////	for _, tt := range tests {
////		t.Run(tt.name, func(t *testing.T) {
////			buf.Reset()
////			tt.logFunc()
////
////			var logMap map[string]interface{}
////			err := json.Unmarshal(buf.Bytes(), &logMap)
////			assert.NoError(t, err)
////			tt.checkLog(t, logMap)
////		})
////	}
////}
//
//// TestFormattedLogging 测试格式化日志功能
//func TestFormattedLogging(t *testing.T) {
//	buf := &mockWriter{}
//	logger, err := NewLogger(Config{
//		Level:   "debug",
//		Format:  "json",
//		Outputs: []io.Writer{buf},
//	})
//	assert.NoError(t, err)
//
//	tests := []struct {
//		name     string
//		logFunc  func()
//		checkLog func(t *testing.T, log map[string]interface{})
//	}{
//		{
//			name: "formatted log",
//			logFunc: func() {
//				logger.Infof("test %s %d", "message", 123)
//			},
//			checkLog: func(t *testing.T, log map[string]interface{}) {
//				assert.Equal(t, "info", log["level"])
//				assert.Equal(t, "test message 123", log["msg"])
//			},
//		},
//		{
//			name: "formatted context log",
//			logFunc: func() {
//				logger.InfoContextf(context.Background(), "test %s %d", "context", 456)
//			},
//			checkLog: func(t *testing.T, log map[string]interface{}) {
//				assert.Equal(t, "info", log["level"])
//				assert.Equal(t, "test context 456", log["msg"])
//			},
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			buf.Reset()
//			tt.logFunc()
//
//			var logMap map[string]interface{}
//			err := json.Unmarshal(buf.Bytes(), &logMap)
//			assert.NoError(t, err)
//			tt.checkLog(t, logMap)
//		})
//	}
//}
//
//// TestErrorLogging 测试错误日志功能
//func TestErrorLogging(t *testing.T) {
//	buf := &mockWriter{}
//	logger, err := NewLogger(Config{
//		Level:   "debug",
//		Format:  "json",
//		Outputs: []io.Writer{buf},
//	})
//	assert.NoError(t, err)
//
//	testError := errors.New("test error")
//
//	tests := []struct {
//		name     string
//		logFunc  func()
//		checkLog func(t *testing.T, log map[string]interface{})
//	}{
//		{
//			name: "error logging",
//			logFunc: func() {
//				logger.WithError(testError).Error("operation failed")
//			},
//			checkLog: func(t *testing.T, log map[string]interface{}) {
//				assert.Equal(t, "error", log["level"])
//				assert.Equal(t, "operation failed", log["msg"])
//				assert.Equal(t, "test error", log["error"])
//			},
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			buf.Reset()
//			tt.logFunc()
//
//			var logMap map[string]interface{}
//			err := json.Unmarshal(buf.Bytes(), &logMap)
//			assert.NoError(t, err)
//			tt.checkLog(t, logMap)
//		})
//	}
//}
//
//// TestCallerReporting 测试调用者信息记录功能
//func TestCallerReporting(t *testing.T) {
//	buf := &mockWriter{}
//	logger, err := NewLogger(Config{
//		Level:        "debug",
//		Format:       "json",
//		Outputs:      []io.Writer{buf},
//		ReportCaller: true,
//	})
//	assert.NoError(t, err)
//
//	logger.Info("test caller")
//
//	var logMap map[string]interface{}
//	err = json.Unmarshal(buf.Bytes(), &logMap)
//	assert.NoError(t, err)
//
//	caller, ok := logMap["caller"].(string)
//	assert.True(t, ok)
//	assert.Contains(t, caller, "logger_test.go")
//}
//
//// TestLogLevels 测试日志级别功能
//func TestLogLevels(t *testing.T) {
//	tests := []struct {
//		name      string
//		level     string
//		logFunc   func(Logger)
//		shouldLog bool
//	}{
//		{
//			name:  "info level",
//			level: "info",
//			logFunc: func(l Logger) {
//				l.Debug("should not log")
//			},
//			shouldLog: false,
//		},
//		{
//			name:  "debug level",
//			level: "debug",
//			logFunc: func(l Logger) {
//				l.Debug("should log")
//			},
//			shouldLog: true,
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			buf := &mockWriter{}
//			logger, err := NewLogger(Config{
//				Level:   tt.level,
//				Format:  "json",
//				Outputs: []io.Writer{buf},
//			})
//			assert.NoError(t, err)
//
//			tt.logFunc(logger)
//
//			logged := buf.Len() > 0
//			assert.Equal(t, tt.shouldLog, logged)
//		})
//	}
//}
//
//// TestMultipleOutputs 测试多输出功能
//func TestMultipleOutputs(t *testing.T) {
//	buf1 := &mockWriter{}
//	buf2 := &mockWriter{}
//
//	logger, err := NewLogger(Config{
//		Level:   "info",
//		Format:  "json",
//		Outputs: []io.Writer{buf1, buf2},
//	})
//	assert.NoError(t, err)
//
//	testMessage := "test multiple outputs"
//	logger.Info(testMessage)
//
//	// 检查两个输出是否都收到了日志
//	var log1, log2 map[string]interface{}
//	err = json.Unmarshal(buf1.Bytes(), &log1)
//	assert.NoError(t, err)
//	err = json.Unmarshal(buf2.Bytes(), &log2)
//	assert.NoError(t, err)
//
//	assert.Equal(t, testMessage, log1["msg"])
//	assert.Equal(t, testMessage, log2["msg"])
//}
//
//// TestFileOutput 测试文件输出功能
//func TestFileOutput(t *testing.T) {
//	// 使用临时文件作为测试
//	tmpFile := t.TempDir() + "/test.log"
//
//	fileOutput := NewFileOutput(FileOutputConfig{
//		Filename:   tmpFile,
//		MaxSize:    1,
//		MaxBackups: 3,
//		MaxAge:     1,
//		Compress:   false,
//	})
//
//	logger, err := NewLogger(Config{
//		Level:   "info",
//		Format:  "json",
//		Outputs: []io.Writer{fileOutput},
//	})
//	assert.NoError(t, err)
//
//	logger.Info("test file output")
//
//	// 验证文件是否创建并写入
//	content, err := os.ReadFile(tmpFile)
//	assert.NoError(t, err)
//	assert.Contains(t, string(content), "test file output")
//}
