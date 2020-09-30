package jaeger

import (
	"errors"
	"io"
	"time"

	"github.com/opentracing/opentracing-go"

	jaeger_cfg "github.com/uber/jaeger-client-go/config"
)

// Config .
type Config struct {
	Type                string        `yaml:"type"` // const, rateLimiting, probabilistic
	TypeParam           float64       `yaml:"type_parm"`
	LogSpans            bool          `yaml:"log_spans"`
	BufferFlushInterval time.Duration `yaml:"buffer_flush_interval"` // second
	QueueSize           int           `yaml:"queue_size"`            // span queue size in memory
	AgentAddr           string        `yaml:"agent_addr"`
}

// NewTracer for current service
// "127.0.0.1:6831"
func NewTracer(serviceName string, c *Config) (tracer opentracing.Tracer, closer io.Closer, err error) {

	if c == nil {
		return nil, nil, errors.New("no trace config")
	}

	jcfg := jaeger_cfg.Configuration{
		Sampler: &jaeger_cfg.SamplerConfig{
			Type:  c.Type,
			Param: c.TypeParam,
		},
		Reporter: &jaeger_cfg.ReporterConfig{
			LogSpans:            c.LogSpans,
			BufferFlushInterval: c.BufferFlushInterval,
			QueueSize:           c.QueueSize,
			LocalAgentHostPort:  c.AgentAddr, // agent服务地址
		},
	}

	tracer, closer, err = jcfg.New(
		serviceName,
	)
	if err != nil {
		return
	}
	if tracer == nil {
		err = errors.New("NewJaegerTracer Error!")
		return
	}

	// 设置为全局的单例tracer
	opentracing.SetGlobalTracer(tracer)

	return
}
