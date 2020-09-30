package main

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/robert-pkg/micro-go/log"
	zap_log "github.com/robert-pkg/micro-go/log/zap-log"
)

func initlog() {

	level, err := log.GetLevel("info")
	if err != nil {
		panic(err)
	}

	zCfg := zap.NewProductionConfig()
	zCfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	zCfg.Encoding = "json"
	zCfg.EncoderConfig.TimeKey = "t"
	zCfg.EncoderConfig.LevelKey = "l"
	zCfg.EncoderConfig.NameKey = "logger"
	zCfg.EncoderConfig.CallerKey = "c"
	zCfg.EncoderConfig.MessageKey = "msg"
	zCfg.EncoderConfig.StacktraceKey = "st"
	zCfg.EncoderConfig.LineEnding = zapcore.DefaultLineEnding
	zCfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	zCfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
	}
	zCfg.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	zCfg.EncoderConfig.EncodeCaller = zapcore.FullCallerEncoder //zapcore.ShortCallerEncoder

	_, err = zap_log.NewLogger(true,
		log.WithFields(map[string]interface{}{"header-x": "value-x"}),
		zap_log.WithConfig(zCfg),
		zap_log.WithCallerSkip(2),
		log.WithLevel(level),
		log.WithOutputToConsole(true),
		log.WithLogFileName("xx.log"),
	)
	if err != nil {
		panic(err)
	}

}
func main() {
	initlog()
	defer log.Close()

	// 推荐使用方式
	log.Info("hello")
	log.Info("hello", "name", "robert", "number", 10)
	log.Info("hello", "name", "robert", "number")
	log.Warn("hello")

	// 这种方式会多创建对象，不推荐使用
	log.Fields(map[string]interface{}{"key": "value"}).Info("xxx")

	log.Debug("xxxxxx")
}
