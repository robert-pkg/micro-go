package zap

import (
	"context"
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/robert-pkg/micro-go/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

type zaplog struct {
	cfg  zap.Config
	zap  *zap.Logger
	opts log.Options
	sync.RWMutex
	fields map[string]interface{}
}

func (l *zaplog) Init(opts ...log.Option) error {

	for _, o := range opts {
		o(&l.opts)
	}

	zapConfig := zap.NewProductionConfig()
	if zconfig, ok := l.opts.Context.Value(configKey{}).(zap.Config); ok {
		zapConfig = zconfig
	}

	if l.opts.Level != log.InfoLevel {
		zapConfig.Level.SetLevel(loggerToZapLevel(l.opts.Level))
	}

	if zcconfig, ok := l.opts.Context.Value(encoderConfigKey{}).(zapcore.EncoderConfig); ok {
		zapConfig.EncoderConfig = zcconfig
	}

	skip, ok := l.opts.Context.Value(callerSkipKey{}).(int)
	if !ok || skip < 1 {
		skip = 1
	}

	var enc zapcore.Encoder
	if zapConfig.Encoding == "json" {
		enc = zapcore.NewJSONEncoder(zapConfig.EncoderConfig)
	} else {
		enc = zapcore.NewConsoleEncoder(zapConfig.EncoderConfig)
	}

	isOutput2Consone := l.opts.IsOutputToConsole
	coreList := make([]zapcore.Core, 0, 2)
	if len(l.opts.LogFileName) > 0 {
		hook := &lumberjack.Logger{
			Filename:   l.opts.LogFileName,
			MaxSize:    100, // megabytes
			MaxBackups: 10,
			MaxAge:     10, // days
			LocalTime:  true,
			Compress:   true,
		}

		fileWriter := zapcore.AddSync(hook)

		core := zapcore.NewCore(enc, fileWriter, zapConfig.Level)
		coreList = append(coreList, core)
	} else {
		// 如果不输出到日志文件，那强制输出到console
		isOutput2Consone = true
	}

	if isOutput2Consone {
		consoleDebugging := zapcore.Lock(os.Stdout)
		consoleEncoder := zapcore.NewConsoleEncoder(zapConfig.EncoderConfig)

		core := zapcore.NewCore(consoleEncoder, consoleDebugging, zapConfig.Level)
		coreList = append(coreList, core)
	}

	log := zap.New(zapcore.NewTee(coreList...),
		zap.AddCaller(),
		zap.AddCallerSkip(skip),
		zap.AddStacktrace(zapcore.ErrorLevel))

	// Adding seed fields if exist
	if l.opts.Fields != nil {
		data := []zap.Field{}
		for k, v := range l.opts.Fields {
			data = append(data, zap.Any(k, v))
		}
		log = log.With(data...)
	}

	// Adding namespace
	if namespace, ok := l.opts.Context.Value(namespaceKey{}).(string); ok {
		log = log.With(zap.Namespace(namespace))
	}

	l.cfg = zapConfig
	l.zap = log
	l.fields = make(map[string]interface{})

	return nil
}

func (l *zaplog) Fields(fields map[string]interface{}) *log.Helper {
	l.Lock()
	nfields := make(map[string]interface{}, len(l.fields))
	for k, v := range l.fields {
		nfields[k] = v
	}
	l.Unlock()
	for k, v := range fields {
		nfields[k] = v
	}

	data := make([]zap.Field, 0, len(nfields))
	for k, v := range fields {
		data = append(data, zap.Any(k, v))
	}

	zl := &zaplog{
		cfg:    l.cfg,
		zap:    l.zap.With(data...),
		opts:   l.opts,
		fields: make(map[string]interface{}),
	}

	return log.NewHelper(zl)
}

func (l *zaplog) Error(err error) log.Logger {
	return l.Fields(map[string]interface{}{"error": err})
}

func (l *zaplog) Log(level log.Level, msg string, args ...interface{}) {
	l.RLock()
	data := make([]zap.Field, 0, len(l.fields))
	for k, v := range l.fields {
		data = append(data, zap.Any(k, v))
	}
	l.RUnlock()

	if len(args) > 0 {
		var sz = len(args)
		for i := 0; i < sz; i += 2 {
			k, ok := args[i].(string)
			var v interface{}
			if !ok {
				k, v = "ErrorKey", k
			}

			if (i + 1) < sz {
				v = args[i+1]
			}

			data = append(data, zap.Any(k, v))
		}
	}

	lvl := loggerToZapLevel(level)
	switch lvl {
	case zap.DebugLevel:
		l.zap.Debug(msg, data...)
	case zap.InfoLevel:
		l.zap.Info(msg, data...)
	case zap.WarnLevel:
		l.zap.Warn(msg, data...)
	case zap.ErrorLevel:
		l.zap.Error(msg, data...)
	case zap.FatalLevel:
		l.zap.Fatal(msg, data...)
	}
}

func (l *zaplog) Logf(level log.Level, format string, args ...interface{}) {
	l.RLock()
	data := make([]zap.Field, 0, len(l.fields))
	for k, v := range l.fields {
		data = append(data, zap.Any(k, v))
	}
	l.RUnlock()

	lvl := loggerToZapLevel(level)
	msg := fmt.Sprintf(format, args...)
	switch lvl {
	case zap.DebugLevel:
		l.zap.Debug(msg, data...)
	case zap.InfoLevel:
		l.zap.Info(msg, data...)
	case zap.WarnLevel:
		l.zap.Warn(msg, data...)
	case zap.ErrorLevel:
		l.zap.Error(msg, data...)
	case zap.FatalLevel:
		l.zap.Fatal(msg, data...)
	}
}

func (l *zaplog) Close() {
	l.zap.Sync()
}

func (l *zaplog) String() string {
	return "zap"
}

func (l *zaplog) Options() log.Options {
	return l.opts
}

// NewLogger New builds a new logger based on options
func NewLogger(asDefault bool, opts ...log.Option) (log.Logger, error) {
	// Default options
	options := log.Options{
		Level:   log.InfoLevel,
		Fields:  make(map[string]interface{}),
		Context: context.Background(),
	}

	l := &zaplog{opts: options}
	if err := l.Init(opts...); err != nil {
		return nil, err
	}

	if asDefault {
		log.DefaultLogger = l
	}

	return l, nil
}

func loggerToZapLevel(level log.Level) zapcore.Level {
	switch level {
	case log.TraceLevel, log.DebugLevel:
		return zap.DebugLevel
	case log.InfoLevel:
		return zap.InfoLevel
	case log.WarnLevel:
		return zap.WarnLevel
	case log.ErrorLevel:
		return zap.ErrorLevel
	case log.FatalLevel:
		return zap.FatalLevel
	default:
		return zap.InfoLevel
	}
}

func zapToLoggerLevel(level zapcore.Level) log.Level {
	switch level {
	case zap.DebugLevel:
		return log.DebugLevel
	case zap.InfoLevel:
		return log.InfoLevel
	case zap.WarnLevel:
		return log.WarnLevel
	case zap.ErrorLevel:
		return log.ErrorLevel
	case zap.FatalLevel:
		return log.FatalLevel
	default:
		return log.InfoLevel
	}
}
