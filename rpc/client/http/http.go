package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"

	"github.com/robert-pkg/micro-go/log"
	"github.com/robert-pkg/micro-go/registry"
	"github.com/robert-pkg/micro-go/rpc"
	"github.com/robert-pkg/micro-go/trace"
)

var (
	// ErrNoAvailableConn .
	ErrNoAvailableConn = errors.New("no available connection.")
)

// Client .
type Client struct {
	serviceName      string
	shortServiceName string
	watcher          registry.Watcher
	getOnce          sync.Once

	pos          int
	instanceList []*serverInstance
	instanceMap  map[string]*serverInstance

	applyChan  chan struct{}        // 申请channal
	grantChan  chan *serverInstance // 发放channal
	updateChan chan *registry.Result
}

// NewClient create Client
func NewClient(serviceName string) (*Client, error) {
	c := &Client{
		serviceName:  serviceName,
		instanceList: make([]*serverInstance, 0, 5),
		instanceMap:  make(map[string]*serverInstance),

		applyChan:  make(chan struct{}),
		grantChan:  make(chan *serverInstance),
		updateChan: make(chan *registry.Result),
	}

	if len(serviceName) > 0 {
		nameList := strings.Split(serviceName, ".")
		if len(nameList) > 0 {
			c.shortServiceName = nameList[len(nameList)-1]
		}
	}

	watcher, err := registry.Watch(registry.WatchService(c.serviceName))
	if err != nil {
		return nil, err
	}

	c.watcher = watcher

	go c.watchRegistry()
	go c.run()
	return c, nil
}

// Stop .
func (c *Client) Stop() {

	// 关掉watcher， close c.updateChan
	c.watcher.Stop()
}

func (c *Client) watchRegistry() {
	for {
		// 如果没有数据，卡住
		res, err := c.watcher.Next()
		if err != nil {
			close(c.updateChan)
			log.Info("watch 退出了")
			break
		}

		c.updateChan <- res
	}
}

func (c *Client) run() {

	for {
		select {
		case res, ok := <-c.updateChan:
			if !ok {
				// 该chan已关闭，watch也关了,这里也退出吧
				log.Info("updateChan 被关了")
				return
			}

			c.updateInstance(res)
		case <-c.applyChan:
			c.grantChan <- c.getBestInstance()

		}

	}

}

func (c *Client) getBestInstance() *serverInstance {

	if len(c.instanceList) <= 0 {
		// 刚创建Client, watch的数据还没来得及过来,
		c.getOnce.Do(func() {
			resultMap, err := registry.GetService(c.serviceName)
			if err != nil {
				log.Error("err", "err", err)
				return
			}

			for k, v := range resultMap {
				s := &serverInstance{
					service: v,
				}

				c.instanceList = append(c.instanceList, s)
				c.instanceMap[k] = s
			}
		})
	}

	if len(c.instanceList) <= 0 {
		return nil
	}

	c.checkPos()

	//if r.service.Nodes[0].Address != key {
	bestKey := c.instanceList[c.pos].service.Nodes[0].Address
	if len(bestKey) <= 0 {
		return nil
	}

	c.change2NextPos()

	if existInstance, ok := c.instanceMap[bestKey]; ok {
		return existInstance
	}

	return nil
}

func (c *Client) checkPos() {
	if c.pos < 0 || c.pos > (len(c.instanceList)-1) {
		// 越界了，校正下
		c.pos = 0
	}
}

func (c *Client) change2NextPos() {
	c.pos++
	if c.pos >= len(c.instanceList) {
		c.pos = 0
	}
}

func (c *Client) updateInstance(res *registry.Result) {
	key := res.Service.Nodes[0].Address

	switch res.Action {
	case "create":

		if _, ok := c.instanceMap[key]; !ok {
			log.Info("实例注册", "服务名", res.Service.Name, "addr", key)

			s := &serverInstance{
				service: res.Service,
			}

			c.instanceList = append(c.instanceList, s)
			c.instanceMap[key] = s
		}

	case "delete":

		newList := make([]*serverInstance, 0, len(c.instanceList))
		for _, r := range c.instanceList {
			if r.service.Nodes[0].Address != key {
				newList = append(newList, r)
			}
		}
		c.instanceList = newList

		if _, ok := c.instanceMap[key]; ok {
			log.Info("实例注销", "服务名", res.Service.Name, "addr", key)
			delete(c.instanceMap, key)
		}

	}
}

func (c *Client) getExistRequestID(ctx context.Context) string {
	reqID := ctx.Value(rpc.RequestID)
	if reqID != nil {
		return reqID.(string)
	}

	return ""
}

// RawCall .
func (c *Client) RawCall(ctx context.Context, method string, reqData []byte) ([]byte, error) {

	var reqID string
	ctx, reqID = rpc.GetOrCreateReqIDFromCtx(ctx)

	c.applyChan <- struct{}{}
	serverInstance := <-c.grantChan
	if serverInstance == nil {
		return nil, ErrNoAvailableConn
	}

	newTraceID := ""
	if tracer := opentracing.GlobalTracer(); tracer != nil {

		if parentSpanContext := ctx.Value("ParentSpanContext"); parentSpanContext == nil {
			var rootSpan opentracing.Span
			rootSpan, newTraceID, ctx = trace.NewRootSpan(ctx, tracer, "http-call", reqID)
			defer rootSpan.Finish()
		}
	}

	if true {
		args := make([]interface{}, 0, 6)
		args = append(args, rpc.RequestID, reqID, "method", method, "body", string(reqData))

		if len(newTraceID) > 0 {
			args = append(args, "newTraceID", newTraceID)
		}

		log.Info("invoke http call", args...)
	}

	url := fmt.Sprintf("http://%s/api/%s/%s", serverInstance.GetAddr(), c.shortServiceName, method)

	out, err := serverInstance.Call(ctx, http.MethodPost, url, reqData)
	if err != nil {
		log.Error("invoke http call fail", rpc.RequestID, reqID, "method", method, "err", err)
		return nil, err
	}

	log.Info("invoke grpc call success", rpc.RequestID, reqID, "method", method, "reply", string(out))

	return out, nil
}

// Call 调用
func (c *Client) Call(ctx context.Context, method string, req, reply interface{}) (err error) {

	var reqData []byte
	if req == nil {
		reqData = []byte("{}")
	} else {
		reqData, err = json.Marshal(req)
		if err != nil {
			return err
		}
	}

	replyData, err := c.RawCall(ctx, method, reqData)
	if err != nil {
		return err
	}

	if reply != nil {
		if err = json.Unmarshal(replyData, reply); err != nil {
			return err
		}
	}

	return nil
}
