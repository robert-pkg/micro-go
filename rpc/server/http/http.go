package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/robert-pkg/micro-go/log"
	"github.com/robert-pkg/micro-go/registry"
	"github.com/robert-pkg/micro-go/trace"
	"github.com/robert-pkg/micro-go/utils"
)

// Server .
type Server struct {
	engine  *gin.Engine
	httpSvr *http.Server

	serviceName      string
	shortServiceName string

	registry registry.Registry

	// registry service instance
	rsvc *registry.Service
}

// NewServer create Server
func NewServer(registry registry.Registry, serviceName string) *Server {
	s := &Server{
		registry:    registry,
		serviceName: serviceName,
	}

	if len(serviceName) > 0 {
		nameList := strings.Split(serviceName, ".")
		if len(nameList) > 0 {
			s.shortServiceName = nameList[len(nameList)-1]
		}
	}

	engine := gin.New()
	engine.Use(gin.Recovery())

	if true {
		m := make([]gin.HandlerFunc, 0, 2)
		m = append(m, logger())

		tracer := opentracing.GlobalTracer()
		if tracer != nil {
			m = append(m, trace.SetUpTraceForGinServer())
		}

		engine.Use(m...)
	}

	s.engine = engine

	return s
}

// Shutdown .
func (s *Server) Shutdown() error {

	if s.registry != nil {
		s.registry.Deregister(s.rsvc)
	}

	// 关闭 http 服务器(优雅退出)
	if s.httpSvr != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.httpSvr.Shutdown(ctx); err != nil {
			log.Error("Server shutdown error", "err", err)
		}
	}

	return nil
}

// GetShortServiceName .
func (s *Server) GetShortServiceName() string {
	return s.shortServiceName
}

// Start start server
func (s *Server) Start() error {

	var addr string
	if true {
		ipList, err := utils.LocalInternalIP()
		if err != nil {
			return err
		}

		if len(ipList) <= 0 {
			return errors.New("无可用IP地址")
		}

		port := utils.RandPort()

		addr = fmt.Sprintf("%s:%d", ipList[0], port)
	}

	s.httpSvr = &http.Server{
		Addr:    addr,
		Handler: s.engine,
	}

	go func() {
		if err := s.httpSvr.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	log.Info("start http server", "serviceName", s.serviceName, "addr", addr)
	if err := s.register(addr); err != nil {
		return err
	}

	return nil
}

func (s *Server) register(addr string) error {

	// register service
	node := &registry.Node{
		Id:       fmt.Sprintf("%s-%s", s.serviceName, addr),
		Address:  addr,
		Metadata: make(map[string]string),
	}

	node.Metadata["registry"] = s.registry.String()
	node.Metadata["server"] = s.String()
	node.Metadata["transport"] = s.String()
	node.Metadata["protocol"] = "http"

	service := &registry.Service{
		Name:      s.serviceName,
		Version:   "latest",
		Nodes:     []*registry.Node{node},
		Endpoints: make([]*registry.Endpoint, 0),
	}

	s.rsvc = service

	regFunc := func(service *registry.Service) error {
		var regErr error

		for i := 0; i < 3; i++ {
			// set the ttl
			//rOpts := []registry.RegisterOption{registry.RegisterTTL(config.RegisterTTL)}
			// attempt to register
			if err := s.registry.Register(service); err != nil {
				// set the error
				regErr = err
				// backoff then retry
				time.Sleep(time.Second * 3)
				continue
			} else {
				// success so nil error
				regErr = nil
				break
			}
		}

		return regErr
	}

	if err := regFunc(s.rsvc); err != nil {
		return errors.Wrap(err, "服务注册失败")
	}

	return nil
}

func (s *Server) GetEngine() *gin.Engine {
	return s.engine
}

//

// String .
func (s *Server) String() string {
	return "http"
}
