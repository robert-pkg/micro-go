package trace

import (
	"time"
)

// Option .
type Option func(*Options)

// NewOptions .
func NewOptions() *Options {
	o := &Options{
		Type:                "rateLimiting",
		TypeParam:           100, // 每秒最多次采样次数
		LogSpans:            true,
		BufferFlushInterval: 1 * time.Second,
		QueueSize:           2000,
	}

	return o
}

// Options .
type Options struct {
	Type                string // const, rateLimiting, probabilistic
	TypeParam           float64
	LogSpans            bool
	BufferFlushInterval time.Duration // second
	QueueSize           int           // span queue size in memory
	AgentAddr           string
}

// WithSampleConst 1收集所有的trace，0代表什么也不收集
func WithSampleConst() Option {
	return func(o *Options) {
		o.Type = "const"
		o.TypeParam = 1
	}
}

// WithSampleRate 限流采集
// rate:每秒采样次数.  比如rate为2代表每秒有2个trace被收集，超出的将被抛弃
func WithSampleRate(rate int) Option {
	return func(o *Options) {
		o.Type = "rateLimiting"
		o.TypeParam = float64(rate)
	}
}

// WithSamplePro 基于概率进行采样。
// value 范围: 0-1.0 间的浮点数, 比如0.1为10%的trace被收集。   1.0为全采样上报
func WithSamplePro(value float64) Option {
	return func(o *Options) {
		o.Type = "probabilistic"
		o.TypeParam = value
	}
}

// WithBufferFlushInterval .
func WithBufferFlushInterval(value time.Duration) Option {
	return func(o *Options) {
		o.BufferFlushInterval = value
	}
}

// WithSetQueueSize .
func WithSetQueueSize(size int) Option {
	return func(o *Options) {
		o.QueueSize = size
	}
}

// WithAgentAddr .
func WithAgentAddr(addr string) Option {
	return func(o *Options) {
		o.AgentAddr = addr
	}
}
