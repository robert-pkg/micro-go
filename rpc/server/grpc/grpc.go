package grpc

import (
	"fmt"
	"net"
	"time"

	"github.com/pkg/errors"

	"github.com/robert-pkg/micro-go/registry"
	"github.com/robert-pkg/micro-go/utils"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/robert-pkg/micro-go/log"

	grpc_go "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	_ "github.com/robert-pkg/micro-go/rpc/codec"
)

// Server .
type Server struct {
	srv *grpc_go.Server

	registry registry.Registry

	// registry service instance
	rsvc *registry.Service
}

// NewServer create Server
func NewServer(registry registry.Registry) *Server {
	s := &Server{
		registry: registry,
	}

	s.srv = grpc_go.NewServer(
		grpc_go.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	)

	return s
}

// Server return the grpc server for registering service.
func (s *Server) Server() *grpc_go.Server {
	return s.srv
}

// Shutdown .
func (s *Server) Shutdown() error {

	if s.registry != nil {
		s.registry.Deregister(s.rsvc)
	}

	return nil
}

// Start start server
func (s *Server) Start(serviceName string) error {

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

	log.Info("start", "serviceName", serviceName, "addr", addr)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	//数据上报
	grpc_prometheus.Register(s.srv)
	grpc_prometheus.EnableHandlingTimeHistogram()

	reflection.Register(s.srv)
	go func() {
		// Serve accepts incoming connections on the listener lis, creating a new
		// ServerTransport and service goroutine for each.
		// Serve will return a non-nil error unless Stop or GracefulStop is called.
		if err := s.srv.Serve(lis); err != nil {
			panic(err)
		}
	}()

	if err = s.register(serviceName, addr); err != nil {
		return err
	}

	return nil
}

func (s *Server) register(serviceName string, addr string) error {

	// register service
	node := &registry.Node{
		Id:       fmt.Sprintf("%s-%s", serviceName, addr),
		Address:  addr,
		Metadata: make(map[string]string),
	}

	node.Metadata["registry"] = s.registry.String()
	node.Metadata["server"] = s.String()
	node.Metadata["transport"] = s.String()
	node.Metadata["protocol"] = "grpc"

	service := &registry.Service{
		Name:      serviceName,
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

// String .
func (s *Server) String() string {
	return "grpc"
}
