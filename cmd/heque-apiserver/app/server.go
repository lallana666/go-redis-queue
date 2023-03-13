package app

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"go.etcd.io/etcd/clientv3"

	"denggotech.cn/heque/heque/apiserver"
	utilflag "denggotech.cn/heque/heque/util/flag"
	utilredis "denggotech.cn/heque/heque/util/redis"
)

func NewAPIServerCommand() *cobra.Command {
	s := NewServerRunOptions()

	cmd := &cobra.Command{
		Use: "heque-apiserver",
		Long: `The heque API server validates and configures data
for the api objects which include jobs, queues, and others. The API 
Server services REST operations and provides the frontend to the
cluster's shared state through which all other components interact.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			utilflag.PrintFlags(cmd.Flags())

			if err := s.Validate(); err != nil {
				return err
			}

			return Run(s, apiserver.SetupSignalHandler())
		},
	}

	s.AddFlags(cmd.Flags())

	return cmd
}

// Run runs the specified APIServer. This should never exit.
func Run(s *ServerRunOptions, stopCh <-chan struct{}) error {
	storage, err := clientv3.New(clientv3.Config{
		Endpoints:   s.ETCDServers,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return err
	}
	defer storage.Close()

	// Initialize the redis
	err = utilredis.Init(s.RedisAddress)
	if err != nil {
		return err
	}

	srv := apiserver.NewAPIServer(&apiserver.Config{
		Storage: storage,
		Prefix:  s.Prefix,
	})
	return srv.ListenAndServe(fmt.Sprintf("%s:%d", s.BindAddress, s.BindPort))
}
