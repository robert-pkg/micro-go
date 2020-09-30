package trace

import (
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// SetUpTraceForGinServer .
func SetUpTraceForGinServer() gin.HandlerFunc {

	return func(c *gin.Context) {

		tracer := opentracing.GlobalTracer()
		if tracer != nil {

			spCtx, err := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(c.Request.Header))
			if err != nil {
				//parentSpan = tracer.StartSpan(c.Request.URL.Path)
				//defer parentSpan.Finish()
			} else {
				span := opentracing.StartSpan(
					c.Request.URL.Path,
					opentracing.ChildOf(spCtx),
					opentracing.Tag{Key: string(ext.Component), Value: "HTTP"},
					ext.SpanKindRPCServer,
				)
				defer span.Finish()

				c.Set("Tracer", tracer)
				c.Set("ParentSpanContext", span.Context())
			}

		}

		c.Next()
	}

}
