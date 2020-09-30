package registry

// Config .
type Config struct {
	RegistryName string // consul, etcd
	Addrs        []string
}
