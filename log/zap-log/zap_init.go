package zap

import (
	"errors"
	"time"

	"github.com/robert-pkg/micro-go/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// InitByConfig quick init
func InitByConfig(c *log.LogConfig) (err error) {

	if c == nil {
		return errors.New("no log config")
	}

	level := log.InfoLevel
	if len(c.Level) > 0 {
		level, err = log.GetLevel("info")
		if err != nil {
			return err
		}
	}

	zCfg := zap.NewProductionConfig()
	zCfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	zCfg.Encoding = c.Encoding
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

	_, err = NewLogger(true,
		WithConfig(zCfg),
		WithCallerSkip(2),
		log.WithLevel(level),
		log.WithOutputToConsole(c.OutputConsole),
		log.WithLogFileName(c.LogPath),
	)
	if err != nil {
		return err
	}

	return nil
}
