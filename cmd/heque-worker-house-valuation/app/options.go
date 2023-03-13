package app

import (
	"errors"
	"net"

	"github.com/spf13/pflag"
)

// WorkerOptions runs a heque worker.
type WorkerOptions struct {
	BindAddress      net.IP
	BindPort         uint
	QueueName        string
	RedisAddress     string
	DebtdbAddress    string
	YunfangKeyID     string
	YunfangAccessKey string
	YunfangDomain    string
	IsMock           string
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
	fs.UintVar(&w.BindPort, "bind-port", 8086, ""+
		"The port on which to serve requests.")
	fs.StringVar(&w.QueueName, "queue-name", "evaluate_house", ""+
		"The name of queue.")
	fs.StringVar(&w.RedisAddress, "redis-address", "localhost:6379", ""+
		"The address of redis server.")
	fs.StringVar(&w.DebtdbAddress, "debtdb-graphql-address", "http://localhost:8081/graphql", ""+
		"The address of debtdb address.")
	fs.StringVar(&w.YunfangKeyID, "yunfang-keyid", "nosuchkeyid", ""+
		"The code of yunfang keyid.")
	fs.StringVar(&w.YunfangAccessKey, "yunfang-access-key", "nosuchaccesskey", ""+
		"The code of yunfang access key.")
	fs.StringVar(&w.YunfangDomain, "yunfang-domain", "nosuchdomain", ""+
		"The URI of yunfang domain.")
	fs.StringVar(&w.IsMock, "is-mock", "false", ""+
		"The mock of fahai-api server.")
}

// Validate checks WorkerOptions and return an error if it fails
func (w *WorkerOptions) Validate() error {
	if w.YunfangKeyID == "" {
		return errors.New("--yunfang-keyid must be specified")
	}

	if w.YunfangAccessKey == "" {
		return errors.New("--yunfang-access-key must be specified")
	}

	if w.YunfangDomain == "" {
		return errors.New("--yunfang-domain must be specified")
	}

	return nil
}
