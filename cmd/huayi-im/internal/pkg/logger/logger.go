package logger

import (
	"os"
	"path/filepath"

	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

// Init 初始化日志
func Init(cfg config.LogConfig) {
	// 编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder, // 日志级别大写（INFO/WARN）
		EncodeTime:     zapcore.ISO8601TimeEncoder,  // 时间格式 ISO8601
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder, // 调用者信息（文件名:行号）
	}

	// 输出配置
	var writeSyncer zapcore.WriteSyncer
	if cfg.Output == "file" {
		// 创建日志目录
		if err := os.MkdirAll(filepath.Dir(cfg.FilePath), 0755); err != nil {
			panic("创建日志目录失败：" + err.Error())
		}

	} else {
		// 控制台输出
		writeSyncer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout))
	}

	// 日志级别
	level := zap.InfoLevel
	switch cfg.Level {
	case "debug":
		level = zap.DebugLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	case "dpanic":
		level = zap.DPanicLevel
	case "panic":
		level = zap.PanicLevel
	case "fatal":
		level = zap.FatalLevel
	}

	// 构建 logger
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		writeSyncer,
		level,
	)
	logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	defer logger.Sync() // 退出时刷新缓冲区
}

// Debug 调试日志
func Debug(msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}

// Info 信息日志
func Info(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

// Warn 警告日志
func Warn(msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}

// Error 错误日志
func Error(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}

// Fatal 致命日志（会终止程序）
func Fatal(msg string, fields ...zap.Field) {
	logger.Fatal(msg, fields...)
}

// Field 创建日志字段
func Field(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}
