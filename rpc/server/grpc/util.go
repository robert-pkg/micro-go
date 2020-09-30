package grpc

import (
	"context"
	"strconv"

	"google.golang.org/grpc/metadata"
)

// GetUserIDFromCtx .
func GetUserIDFromCtx(ctx context.Context) int64 {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0
	}

	userIDList, ok := md["user_id"]
	if !ok {
		return 0
	}

	if len(userIDList) <= 0 {
		return 0
	}

	userID, err := strconv.ParseInt(userIDList[0], 10, 64)
	if err != nil {
		return 0
	}

	return userID
}
