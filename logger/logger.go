package logger

import (
	"land/settings"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var lg *zap.Logger

func Init(cfg *settings.LogConfig, mode string) (err error) {
	writer := getLogWriter(cfg.Filename, cfg.MaxSize, cfg.MaxBackups, cfg.MaxAge)
	encoder := getEncoder()

	l := new(zapcore.Level)
	if err := l.UnmarshalText([]byte(cfg.Level)); err != nil {
		return err
	}

	var core zapcore.Core

	if mode == "dev" {
		// dev模式下，日志输出到控制台
		consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		core = zapcore.NewTee(
			// 一个写入文件一个额外写入控制台
			zapcore.NewCore(encoder, writer, l),
			zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), zap.DebugLevel),
		)
	} else {
		core = zapcore.NewCore(encoder, writer, l)
	}

	lg = zap.New(core, zap.AddCaller())
	zap.ReplaceGlobals(lg)
	zap.L().Info("logger init success")
	return nil
}

// 编码器默认设置
func getEncoder() zapcore.Encoder {
	en := zap.NewProductionEncoderConfig()
	en.EncodeTime = zapcore.ISO8601TimeEncoder
	en.TimeKey = "time"
	en.EncodeLevel = zapcore.CapitalLevelEncoder
	en.EncodeDuration = zapcore.SecondsDurationEncoder
	en.EncodeCaller = zapcore.ShortCallerEncoder
	return zapcore.NewJSONEncoder(en)
}

func getLogWriter(filename string, maxSize, maxBackup, maxAge int) zapcore.WriteSyncer {
	return zapcore.AddSync(&lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSize,
		MaxBackups: maxBackup,
		MaxAge:     maxAge,
	})
}
