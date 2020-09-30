package http

import (
	"github.com/gin-gonic/gin"
	"github.com/robert-pkg/micro-go/log"
	"github.com/robert-pkg/micro-go/rpc"
)

func logger() gin.HandlerFunc {

	return func(c *gin.Context) {

		var reqID string
		if c.Request.Header != nil {
			if list, ok := c.Request.Header[rpc.RequestID]; ok {
				if len(list) > 0 {
					reqID = list[0]
				}
			}
		}

		if len(reqID) > 0 {
			// 将 requestID 注入到log中
			log.SetReqMetaForGoroutine(reqID)
			defer log.DeleteMetaForGoroutine()
		}

		c.Next()
	}

}
