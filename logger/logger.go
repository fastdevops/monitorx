package logger

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.Logger

// InitLogger 初始化 Zap 日志记录器
func InitLogger(logLevel, logPath, logName string, maxSize, maxBackups, maxAge int) {
	// 设置日志级别
	level := zapcore.InfoLevel
	if err := level.UnmarshalText([]byte(logLevel)); err != nil {
		panic(err)
	}

	// 创建日志文件路径
	if err := os.MkdirAll(logPath, os.ModePerm); err != nil {
		panic(err)
	}
	logFilePath := filepath.Join(logPath, logName)

	// 定制日志编码器配置，支持多行输出
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,   // 日志级别大写
		EncodeTime:     zapcore.ISO8601TimeEncoder,    // 时间格式为标准时间格式
		EncodeCaller:   zapcore.ShortCallerEncoder,    // 短路径调用者信息
		EncodeDuration: zapcore.StringDurationEncoder, // 使用字符串的时间间隔
	}

	// 创建一个 console 编码器，支持多行输出
	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	// 创建核心，使用 lumberjack 实现日志自动切割
	core := zapcore.NewCore(encoder, zapcore.AddSync(&lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    maxSize,    // MB
		MaxBackups: maxBackups, // 保留的备份数量
		MaxAge:     maxAge,     // 天
	}), level)

	// 创建日志记录器
	Logger = zap.New(core)
}

// Sync 确保日志缓冲被写入
func Sync() {
	_ = Logger.Sync()
}

// 日志配置初始化
func InitLoggerConfig(Loglevel, Logpath, LogName string, LogMaxSize, LogMaxBackups, LogMaxAge int) {
	// init log
	InitLogger(Loglevel, Logpath, LogName, LogMaxSize, LogMaxBackups, LogMaxAge)
	defer Sync()
}
