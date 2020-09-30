package zap

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/robert-pkg/micro-go/log"
)

type Options struct {
	log.Options
}

type callerSkipKey struct{}

func WithCallerSkip(i int) log.Option {
	return log.SetOption(callerSkipKey{}, i)
}

type configKey struct{}

// WithConfig pass zap.Config to logger
func WithConfig(c zap.Config) log.Option {
	return log.SetOption(configKey{}, c)
}

type encoderConfigKey struct{}

// WithEncoderConfig pass zapcore.EncoderConfig to logger
func WithEncoderConfig(c zapcore.EncoderConfig) log.Option {
	return log.SetOption(encoderConfigKey{}, c)
}

type namespaceKey struct{}

func WithNamespace(namespace string) log.Option {
	return log.SetOption(namespaceKey{}, namespace)
}
