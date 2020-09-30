package consul

import (
	"github.com/pkg/errors"
	"github.com/robert-pkg/micro-go/registry"
)

// InitRegistry .
func InitRegistry(c *registry.Config) registry.Registry {

	// 根据系统配置，使用etcd， consul
	registry, err := NewRegistry()
	if err != nil {
		panic(errors.Wrap(err, "grpc server start fail"))
	}

	return registry
}

// InitRegistryAsDefault .
func InitRegistryAsDefault(c *registry.Config) registry.Registry {
	registry.DefaultRegistry = InitRegistry(c)
	return registry.DefaultRegistry
}
