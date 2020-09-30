package trace

import (
	"context"

	"github.com/opentracing/opentracing-go"
	opentracelog "github.com/opentracing/opentracing-go/log"
	"github.com/robert-pkg/micro-go/rpc"
	"github.com/uber/jaeger-client-go"
)

// NewRootSpan create new span, then set trace id to ctx
func NewRootSpan(ctx context.Context, tracer opentracing.Tracer, spanName string, requestID string) (newSpan opentracing.Span, newTraceID string, newCtx context.Context) {

	// 生成开始一个新的Span
	newSpan = tracer.StartSpan(spanName)

	// set TraceID to RequestID
	if sc, ok := newSpan.Context().(jaeger.SpanContext); ok {
		newTraceID = sc.TraceID().String()
		newSpan.LogFields(opentracelog.String(rpc.RequestID, requestID))
	} else {
		// 后续增加zipkin的支持
	}

	// 返回span的SpanContext
	newCtx = context.WithValue(ctx, "ParentSpanContext", newSpan.Context())
	//newCtx = opentracing.ContextWithSpan(ctx, newSpan)
	return
}
