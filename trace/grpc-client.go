package trace

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	opentracelog "github.com/opentracing/opentracing-go/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// ClientInterceptor .
func ClientInterceptor(tracer opentracing.Tracer) grpc.UnaryClientInterceptor {

	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

		if tracer == nil {
			// no tracer, so invoke directly
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		var parentCtx opentracing.SpanContext
		if true {
			if pc := ctx.Value("ParentSpanContext"); pc != nil {

				if realPC, ok := pc.(opentracing.SpanContext); ok {
					parentCtx = realPC
				}

			}
		}

		span := tracer.StartSpan(
			method,
			opentracing.ChildOf(parentCtx), // can be nil
			opentracing.Tag{Key: string(ext.Component), Value: "gRPC"},
			ext.SpanKindRPCClient,
		)
		defer span.Finish()

		//将之前放入context中的metadata数据取出，如果没有则新建一个metadata
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		} else {
			md = md.Copy()
		}

		//将追踪数据注入到metadata中
		mdWriter := MDReaderWriter{md}
		err := tracer.Inject(span.Context(), opentracing.TextMap, mdWriter)
		if err != nil {
			span.LogFields(opentracelog.String("inject-error", err.Error()))
		}

		//将metadata数据装入context中
		newCtx := metadata.NewOutgoingContext(ctx, md)

		//使用带有追踪数据的context进行grpc调用
		err = invoker(newCtx, method, req, reply, cc, opts...)
		if err != nil {
			span.LogFields(opentracelog.String("call-error", err.Error()))
		}

		return err
	}
}
