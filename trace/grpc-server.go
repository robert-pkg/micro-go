package trace

import (
	"context"

	"github.com/robert-pkg/micro-go/rpc"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	opentracelog "github.com/opentracing/opentracing-go/log"
	"github.com/robert-pkg/micro-go/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// ServerInterceptor .
func ServerInterceptor(tracer opentracing.Tracer) grpc.UnaryServerInterceptor {

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {

		if tracer == nil {
			return handler(ctx, req)
		}

		//从context中取出metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}

		//从metadata中取出最终数据，并创建出span对象
		spanContext, err := tracer.Extract(opentracing.TextMap, MDReaderWriter{md})
		if err != nil && err != opentracing.ErrSpanContextNotFound {
			log.Error("extract from metadata fail", "err", err)
		} else {
			// 生成 server 端的span
			span := tracer.StartSpan(
				info.FullMethod,
				ext.RPCServerOption(spanContext),
				opentracing.Tag{Key: string(ext.Component), Value: "gRPC"},
				ext.SpanKindRPCServer,
			)
			defer span.Finish()

			ctx = context.WithValue(ctx, "ParentSpanContext", span.Context())

			// 将requestID注入到日志中
			requestIDs := md.Get(rpc.RequestID)
			if len(requestIDs) >= 1 {

				// 将 requestID 注入到log中
				log.SetReqMetaForGoroutine(requestIDs[0])
				span.LogFields(opentracelog.String(rpc.RequestID, requestIDs[0]))

				defer log.DeleteMetaForGoroutine()
			}
		}

		return handler(ctx, req)
	}
}
