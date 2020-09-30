package grpc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/robert-pkg/micro-go/log"
	"github.com/robert-pkg/micro-go/registry"
	"github.com/robert-pkg/micro-go/rpc/codec"
	_ "github.com/robert-pkg/micro-go/rpc/codec" // To register JSON codec
	grpc_go "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	ErrNoAvailableConn = errors.New("no available connection.")
)

// 服务实例
type serverInstance struct {
	service *registry.Service
}

// Client .
type Client struct {
	serviceName      string
	shortServiceName string
	watcher          registry.Watcher
	getOnce          sync.Once

	pos          int
	instanceList []*serverInstance
	instanceMap  map[string]*serverInstance
	connMap      map[string]*grpc_go.ClientConn // grpc 连接

	applyChan  chan struct{}            // 申请channal
	grantChan  chan *grpc_go.ClientConn // 发放channal
	updateChan chan *registry.Result
}

// NewClient create Client
func NewClient(serviceName string) (*Client, error) {
	c := &Client{
		serviceName:  serviceName,
		instanceList: make([]*serverInstance, 0, 5),
		instanceMap:  make(map[string]*serverInstance),
		connMap:      make(map[string]*grpc_go.ClientConn),

		applyChan:  make(chan struct{}),
		grantChan:  make(chan *grpc_go.ClientConn),
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
			c.grantChan <- c.getBestConn()

		}

	}

}

func (c *Client) getBestConn() *grpc_go.ClientConn {

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

	if existConn, ok := c.connMap[bestKey]; ok {
		return existConn
	}

	// Set up a connection to the server.
	var backoffConfig grpc_go.BackoffConfig
	backoffConfig.MaxDelay = time.Second * 10

	conn, err := grpc_go.Dial(bestKey, grpc_go.WithInsecure(),
		grpc_go.WithBackoffConfig(backoffConfig),
		grpc_go.WithDefaultCallOptions(grpc_go.CallContentSubtype(codec.JSON{}.Name())))
	if err != nil {
		log.Error("err", "err", err)
		return nil
	}

	c.connMap[bestKey] = conn
	return conn
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

		if conn, ok := c.connMap[key]; ok {
			conn.Close()
			delete(c.connMap, key)
		}
	}
}

// Call 调用
func (c *Client) Call(method string, header map[string]string, body []byte) ([]byte, error) {

	c.applyChan <- struct{}{}
	conn := <-c.grantChan
	if conn == nil {
		return nil, ErrNoAvailableConn
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if header != nil {
		md := metadata.New(header)
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	realMethodName := fmt.Sprintf("/%s.%s/%s", c.serviceName, c.shortServiceName, method)

	var out []byte
	err := conn.Invoke(ctx, realMethodName, body, &out)
	if err != nil {
		log.Error("err", "err", err)
		return nil, err
	}

	return out, nil
}
