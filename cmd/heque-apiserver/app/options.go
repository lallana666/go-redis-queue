package app

import (
	"net"

	"github.com/spf13/pflag"
)

// ServerRunOptions runs a heque api server.
type ServerRunOptions struct {
	BindAddress  net.IP
	BindPort     uint
	ETCDServers  []string
	Prefix       string
	RedisAddress string
}

// NewServerRunOptions creates a new ServerRunOptions object with default parameters
func NewServerRunOptions() *ServerRunOptions {
	s := ServerRunOptions{}
	return &s
}

// AddFlags adds flags for a specific APIServer to the specified FlagSet
func (s *ServerRunOptions) AddFlags(fs *pflag.FlagSet) {
	fs.IPVar(&s.BindAddress, "bind-address", net.ParseIP("0.0.0.0"), ""+
		"The IP address on which to listen for the --bind-port port. If blank, all interfaces "+
		"will be used (0.0.0.0 for all IPv4 interfaces and :: for all IPv6 interfaces).")
	fs.UintVar(&s.BindPort, "bind-port", 8086, ""+
		"The port on which to serve requests.")
	fs.StringSliceVar(&s.ETCDServers, "etcd-servers", []string{"localhost:2379"}, ""+
		"List of etcd servers to connect with (ip:port), comma separated.")
	fs.StringVar(&s.Prefix, "prefix", "/registry", ""+
		"The prefix of queue.")
	fs.StringVar(&s.RedisAddress, "redis-address", "localhost:6379", ""+
		"The address of redis server.")
}

// Validate checks ServerRunOptions and return an error if it fails
func (s *ServerRunOptions) Validate() error {
	return nil
}
