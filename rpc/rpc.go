package rpc

import (
	"context"
	"strconv"

	"github.com/robert-pkg/micro-go/rpc/metadata"
	uuid "github.com/satori/go.uuid"
)

const (
	// RequestID 请求ID的标识
	RequestID string = "Request-Id" // title格式
	// UserID 用户ID的标识
	UserID     string = "User-Id"     // title格式
	DeviceType string = "Device-Type" // // title格式
	SkipTrace  string = "Skip-Trace"
)

// GetOrCreateReqIDFromCtx .
func GetOrCreateReqIDFromCtx(ctx context.Context) (context.Context, string) {

	if reqID, ok := metadata.Get(ctx, RequestID); ok {
		return ctx, reqID
	}

	newReqID := uuid.NewV4().String()
	ctx = metadata.Set(ctx, RequestID, newReqID)

	return ctx, newReqID
}

// GetUserIDFromCtx 返回 用户ID
func GetUserIDFromCtx(ctx context.Context) int64 {

	if strUserID, ok := metadata.Get(ctx, UserID); ok {
		if userID, err := strconv.ParseInt(strUserID, 10, 64); err != nil {
			return 0
		} else {
			return userID
		}
	}

	return 0
}
