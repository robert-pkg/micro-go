package zap

import (
	"sync"

	"github.com/petermattis/goid"
)

type storeMeta struct {
	reqID string
}

var (
	requestIDs sync.Map
)

func getGoID() int64 {
	return goid.Get()
}

// SetReqMetaForGoroutine .
func (l *zaplog) SetReqMetaForGoroutine(requestID string) {
	requestIDs.Store(getGoID(), storeMeta{
		reqID: requestID,
	})
}

func (l *zaplog) DeleteMetaForGoroutine() {
	requestIDs.Delete(getGoID())
}

// Get 返回设置的 ReqMeta
func getReqMetaForGoroutine() (interface{}, bool) {
	return requestIDs.Load(getGoID())
}

// GetReqIDForGoroutine 返回设置的 RequestID
func GetReqIDForGoroutine() (interface{}, bool) {
	meta, ok := getReqMetaForGoroutine()
	if ok {
		return meta.(storeMeta).reqID, ok
	}
	return nil, ok
}
