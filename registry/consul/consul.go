package consul

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	consul "github.com/hashicorp/consul/api"
	"github.com/robert-pkg/micro-go/log"
	"github.com/robert-pkg/micro-go/registry"
)

var (
	defaultConsulAddr = "127.0.0.1:8500"
)

type consulRegistry struct {
	ctx    context.Context
	cancel context.CancelFunc

	Address []string
	opts    registry.Options
	client  *consul.Client
	config  *consul.Config

	rsvc *registry.Service
}

func (c *consulRegistry) Init(opts ...registry.Option) error {

	// set opts
	for _, o := range opts {
		o(&c.opts)
	}

	// check if there are any addrs
	var addrs []string

	// iterate the options addresses
	for _, address := range c.opts.Addrs {
		// check we have a port
		addr, port, err := net.SplitHostPort(address)
		if ae, ok := err.(*net.AddrError); ok && ae.Err == "missing port in address" {
			port = "8500"
			addr = address
			addrs = append(addrs, net.JoinHostPort(addr, port))
		} else if err == nil {
			addrs = append(addrs, net.JoinHostPort(addr, port))
		}
	}

	config := consul.DefaultConfig()

	// set the addrs
	if len(addrs) <= 0 {
		addrs = append(addrs, defaultConsulAddr)
	}
	c.Address = addrs

	// set timeout
	if c.opts.Timeout > 0 {
		config.HttpClient.Timeout = c.opts.Timeout
	}

	// set the config
	c.config = config

	// remove client
	c.client = nil

	// setup the client
	if c.Client() == nil {
		return errors.New("consul连接失败")
	}
	return nil
}

func (c *consulRegistry) Client() *consul.Client {
	if c.client != nil {
		return c.client
	}

	for _, addr := range c.Address {
		// set the address
		c.config.Address = addr

		// create a new client
		tmpClient, _ := consul.NewClient(c.config)

		// test the client
		_, err := tmpClient.Agent().Host()
		if err != nil {
			continue
		}

		// set the client
		c.client = tmpClient
		return c.client
	}

	// set the default
	c.client, _ = consul.NewClient(c.config)

	// return the client
	return c.client
}

func (c *consulRegistry) Options() registry.Options {
	return c.opts
}

func (c *consulRegistry) Register(s *registry.Service, opts ...registry.RegisterOption) error {
	if len(s.Nodes) == 0 {
		return errors.New("Require at least one node")
	}

	var options registry.RegisterOptions
	for _, o := range opts {
		o(&options)
	}

	// use first node
	node := s.Nodes[0]

	var err error
	var hostIP string
	var port int
	if true {
		ss := strings.Split(node.Address, ":")
		if len(ss) != 2 {
			err := fmt.Errorf("node.Address格式错误. [%s]", node.Address)
			log.Error("err", "err", err)
			return err
		}

		hostIP = ss[0]

		port, err = strconv.Atoi(ss[1])
		if err != nil {
			log.Error("err", "err", err)
			return err
		}
	}

	regFunc := func(hostIP string, port int) error {
		var check *consul.AgentServiceCheck
		if true {
			check = &consul.AgentServiceCheck{
				//HTTP:                           "http://" + d.hostIP + ":" + strconv.Itoa(d.healthCheckPort) + "/" + serviceName + "/status",
				//Interval:                       "10s", // 健康检查间隔
				TTL:                            "21s",
				Timeout:                        "5s",
				DeregisterCriticalServiceAfter: "5m", // 注销时间，相当于过期时间
			}
		}

		service := &consul.AgentServiceRegistration{
			ID:      node.Id,
			Name:    s.Name,
			Port:    port,
			Address: hostIP,
			Check:   check,
		}

		if err := c.client.Agent().ServiceRegister(service); err != nil {
			log.Error("register to consul fail", "error", err)
			return err
		}
		c.client.Agent().PassTTL("service:"+node.Id, "")

		return nil
	}

	if err := regFunc(hostIP, port); err != nil {
		return err
	}

	go func() {
		keepAliveTicker := time.NewTicker(7 * time.Second)
		registerTicker := time.NewTicker(time.Minute)

		for {
			select {
			case <-c.ctx.Done():
				keepAliveTicker.Stop()
				registerTicker.Stop()
				if err := c.client.Agent().ServiceDeregister(node.Id); err != nil {
					log.Error("err", "err", err)
				}
				log.Info("ServiceDeregister")
				return
			case <-keepAliveTicker.C:
				// 心跳
				c.client.Agent().PassTTL("service:"+node.Id, "")
			case <-registerTicker.C:
				// 兜底的
				if err = regFunc(hostIP, port); err != nil {
					log.Error("err", "err", err)
				}
			}
		}
	}()

	c.rsvc = s
	return nil
}

func (c *consulRegistry) Deregister(s *registry.Service, opts ...registry.DeregisterOption) error {
	log.Info("Deregister")
	c.cancel()
	return nil
}

func (c *consulRegistry) GetService(name string, opts ...registry.GetOption) (map[string]*registry.Service, error) {

	entryList, _, err := c.client.Health().Service(name, "", false, &consul.QueryOptions{})
	if err != nil {
		log.Error("err", "error", err)
		return nil, err
	}

	resultMap := make(map[string]*registry.Service)
	for _, entry := range entryList {
		for _, health := range entry.Checks {

			if health.ServiceName != name {
				continue
			}

			//log.Info("svr", "ip", entry.Service.Address, "port", entry.Service.Port,
			//	"ServiceName", health.ServiceName, "ServiceID", health.ServiceID, "status", health.Status)

			svc := &registry.Service{
				Endpoints: nil,
				Name:      name,
				Version:   "",
				Nodes:     make([]*registry.Node, 0, 1),
			}

			svc.Nodes = append(svc.Nodes, &registry.Node{
				Id:       entry.Service.ID,
				Address:  fmt.Sprintf("%s:%d", entry.Service.Address, entry.Service.Port), // ip地址和端口
				Metadata: make(map[string]string),
			})

			key := fmt.Sprintf("%s:%d", entry.Service.Address, entry.Service.Port)
			resultMap[key] = svc
		}
	}

	return resultMap, nil
}

func (c *consulRegistry) Watch(opts ...registry.WatchOption) (registry.Watcher, error) {
	return newConsulWatcher(c, opts...)
}

func (c *consulRegistry) String() string {
	return "consul"
}

func NewRegistry(opts ...registry.Option) (registry.Registry, error) {

	ctx, cancel := context.WithCancel(context.Background())

	cr := &consulRegistry{
		ctx:    ctx,
		cancel: cancel,
		opts:   registry.Options{},
		//register:    make(map[string]uint64),
		//lastChecked: make(map[string]time.Time),
		//queryOptions: &consul.QueryOptions{
		//	AllowStale: true,
		//},
	}
	if err := cr.Init(opts...); err != nil {
		return nil, err
	}

	return cr, nil
}
