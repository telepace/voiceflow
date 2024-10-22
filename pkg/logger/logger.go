package logger

import (
    "log"
    "os"

    "github.com/sirupsen/logrus"
)

var Logger = logrus.New()

func Init(level string) {
    // 设置日志输出格式
    Logger.SetFormatter(&logrus.TextFormatter{
        FullTimestamp: true,
    })

    // 设置日志级别
    lvl, err := logrus.ParseLevel(level)
    if err != nil {
        log.Printf("Invalid log level '%s', defaulting to 'info'", level)
        lvl = logrus.InfoLevel
    }
    Logger.SetLevel(lvl)

    // 设置日志输出位置
    Logger.SetOutput(os.Stdout)
}