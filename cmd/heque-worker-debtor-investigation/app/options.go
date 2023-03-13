package app

import (
	"net"

	"github.com/spf13/pflag"
)

// WorkerOptions runs a heque worker.
type WorkerOptions struct {
	BindAddress          net.IP
	BindPort             uint
	QueueName            string
	RedisAddress         string
	DebtdbAddress        string
	CreditGatewayAddress string
	IsMock               string
}

// NewWorkerOptions creates a new WorkerOptions object with default parameters
func NewWorkerOptions() *WorkerOptions {
	w := WorkerOptions{}
	return &w
}

// AddFlags adds flags for a specific worker to the specified FlagSet
func (w *WorkerOptions) AddFlags(fs *pflag.FlagSet) {
	fs.IPVar(&w.BindAddress, "bind-address", net.ParseIP("127.0.0.1"), ""+
		"The IP address on which to listen for the --bind-port port.")
	fs.UintVar(&w.BindPort, "bind-port", 8087, ""+
		"The port on which to serve requests.")
	fs.StringVar(&w.QueueName, "queue-name", "investigate_debtor", ""+
		"The name of queue.")
	fs.StringVar(&w.RedisAddress, "redis-address", "localhost:6379", ""+
		"The address of redis server.")
	fs.StringVar(&w.DebtdbAddress, "debtdb-graphql-address", "http://localhost:8081/graphql", ""+
		"The address of debtdb address.")
	fs.StringVar(&w.CreditGatewayAddress, "credit-gateway-address", "http://localhost:8085/v1/graphql", ""+
		"The address of credit-gateway address.")
	fs.StringVar(&w.IsMock, "is-mock", "false", ""+
		"The mock of fahai-api server.")
}
