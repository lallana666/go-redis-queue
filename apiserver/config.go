package apiserver

import "go.etcd.io/etcd/clientv3"

// Config is a structure used to configure an APIServer.
type Config struct {
	// Prefix is the prefix of queue
	Prefix string
	// ETCDServers is List of etcd servers to connect with (ip:port), comma separated.
	ETCDServers []string
	// Storage is client of etcd
	Storage *clientv3.Client
}
