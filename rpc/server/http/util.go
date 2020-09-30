package http

import (
	"context"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robert-pkg/micro-go/log"

	"github.com/robert-pkg/micro-go/rpc/metadata"
	//"google.golang.org/grpc/metadata"
)

// GetContext .
func GetContext(c *gin.Context) (context.Context, context.CancelFunc) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	md := make(metadata.Metadata)
	for k, v := range c.Request.Header {
		if len(v) >= 1 {
			key := strings.ToLower(k)
			md[key] = v[0]
			log.Info("FillContext", "k", k, "key", key, "v", v)
		}
	}

	if parentSpanContext, ok := c.Get("ParentSpanContext"); ok {
		ctx = context.WithValue(ctx, "ParentSpanContext", parentSpanContext)
	}

	// c.Set("ParentSpanContext", parentSpan.Context())

	return metadata.NewContext(ctx, md), cancel
}
