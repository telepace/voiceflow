// // example/main.go
package main

//
//import (
//	"context"
//	"errors"
//	"io"
//	"os"
//
//	"github.com/telepace/voiceflow/pkg/logger"
//)
//
//func main() {
//	// 创建文件输出
//	fileOutput := logger.NewFileOutput(logger.FileOutputConfig{
//		Filename:   "_output/log/app.log",
//		MaxSize:    100,
//		MaxBackups: 3,
//		MaxAge:     28,
//		Compress:   true,
//	})
//
//	// 创建日志实例
//	log, err := logger.NewLogger(logger.Config{
//		Level:  "info",
//		Format: "json",
//		Outputs: []io.Writer{
//			os.Stdout,
//			fileOutput,
//		},
//		ServiceInfo: logger.ServiceInfo{
//			ServiceID:  "order-service",
//			InstanceID: "instance-1",
//			Extra: logger.Fields{
//				"region": "us-west",
//				"env":    "production",
//			},
//		},
//		ReportCaller: true,
//		CallerSkip:   0,
//	})
//	if err != nil {
//		panic(err)
//	}
//
//	// 基础用法
//	log.Info("service starting...")
//	log.Infof("listening on port %d", 8080)
//
//	// 带ctx的用法
//	ctx := context.Background()
//	log.InfoContext(ctx, "handling request")
//	log.InfoContextf(ctx, "processing order %s", "ORD-123")
//
//	// 链式调用
//	log.WithField("user_id", "U123").Info("user logged in")
//
//	err = errors.New("database connection failed")
//	log.WithError(err).Error("failed to connect to database")
//
//	log.WithFields(logger.Fields{
//		"order_id": "ORD-123",
//		"user_id":  "U123",
//		"amount":   99.99,
//	}).Info("order created")
//
//	// 组合使用
//	log.WithContext(ctx).
//		WithField("user_id", "U123").
//		WithError(err).
//		Error("operation failed")
//}
