package logger

import (
	"land/settings"
	"net"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
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
			// 一个写入文件,一个额外写入控制台
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

// 日志文件写入器
func getLogWriter(filename string, maxSize, maxBackup, maxAge int) zapcore.WriteSyncer {
	return zapcore.AddSync(&lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSize,
		MaxBackups: maxBackup,
		MaxAge:     maxAge,
	})
}

// 自定义接收gin的日志
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()

		cost := time.Since(startTime)
		lg.Info(path,
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
			zap.Duration("cost", cost),
		)
	}
}

// GinRecovery recover掉项目可能出现的panic，并使用zap记录相关日志
func GinRecovery(stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe { // 客户端断开连接
					lg.Error(c.Request.URL.Path,
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
					c.Error(err.(error))
					c.Abort()
					return
				}

				// 根据输入参数决定是否输出错误堆栈信息
				if stack {
					lg.Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
						zap.String("stack", string(debug.Stack())),
					)
				} else {
					lg.Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
				}

			}
		}()
		c.Next()
	}
}
