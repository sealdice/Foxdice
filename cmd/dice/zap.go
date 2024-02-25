package main

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	Log *Logger
)

type Logger = zap.SugaredLogger

type LoggerOption struct {
	NoFile      bool   // 只在控制台输出 不写入文件
	NoJson      bool   // 写入文件时不使用 json 编码器
	Name        string // 日志文件名
	MaxSize     int    // 日志文件大小（M）
	MaxBackups  int    // 最多存在多少个切片文件
	MaxAge      int    // 保存的最大天数
	Debug       bool
	Dir         string
	LoggerLevel int8 // 从第几级别的日志开始输入（日志级别 -1 ~ 5 debug ~ Fatal）
}

func (opt *LoggerOption) getWSFile(ext string) zapcore.WriteSyncer {
	return zapcore.AddSync(&lumberjack.Logger{
		Filename:   fmt.Sprintf("%s.%s.txt", path.Join(opt.Dir, opt.Name), ext),
		MaxSize:    opt.MaxSize,
		MaxBackups: opt.MaxBackups,
		MaxAge:     opt.MaxAge,
		Compress:   false,
		LocalTime:  true,
	})
}

func NewLogger(opt *LoggerOption) *Logger {
	// 在终端使用编码器
	dc := zap.NewDevelopmentEncoderConfig()
	// 时间格式
	dc.EncodeTime = func(t time.Time, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString("[" + t.Format(time.TimeOnly) + "]")
	}
	dc.EncodeCaller = func(c zapcore.EntryCaller, e zapcore.PrimitiveArrayEncoder) {
		if opt.Debug {
			e.AppendString(c.FullPath())
		} else {
			e.AppendString("[" + c.TrimmedPath() + "]")
		}

	}
	dc.EncodeLevel = zapcore.CapitalColorLevelEncoder
	dc.ConsoleSeparator = strings.Repeat(" ", 2)
	con := zapcore.NewConsoleEncoder(dc)
	// JSON 方便序列化 做一些其他小玩意
	pc := zap.NewProductionEncoderConfig()
	json := zapcore.NewJSONEncoder(pc)

	var cores []zapcore.Core
	cores = append(cores,
		zapcore.NewCore(
			con,
			zapcore.AddSync(zapcore.Lock(os.Stdout)),
			zap.LevelEnablerFunc(func(level zapcore.Level) bool {
				return zapcore.Level(opt.LoggerLevel) >= level
			})))

	if !opt.NoFile {
		en := con
		if !opt.NoJson {
			en = json
		}
		cores = append(cores, zapcore.NewCore(
			en,
			opt.getWSFile("main"),
			zap.LevelEnablerFunc(func(level zapcore.Level) bool {
				// 只记录 info warn 且 debug 模式下全记录
				return (level < zap.ErrorLevel && level > zap.DebugLevel) || opt.Debug
			})))
	}

	if !opt.Debug {
		// 方便别人快速截图错误信息
		cores = append(cores, zapcore.NewCore(
			con,
			opt.getWSFile("error"),
			zap.LevelEnablerFunc(func(level zapcore.Level) bool {
				// 只记录 info warn 且 debug 模式下全记录
				return (level < zap.ErrorLevel && level > zap.DebugLevel) || opt.Debug
			})))
	}

	slog := zap.New(zapcore.NewTee(cores...), zap.AddCaller()).Sugar()
	slog.Infof("<%s> Logger Go!", opt.Name)
	return slog
}
