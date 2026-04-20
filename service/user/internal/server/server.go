package server

import (
	"user/internal/conf"

	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/google/wire"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(NewGRPCServer, NewHTTPServer, NewRegistrar)

// NewRegistrar creates a new etcd registrar
func NewRegistrar(conf *conf.Registry) registry.Registrar {
	if conf == nil || conf.Etcd == nil {
		return nil
	}
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   conf.Etcd.Endpoints,
		DialTimeout: conf.Etcd.DialTimeout.AsDuration(),
	})
	if err != nil {
		panic(err)
	}
	return etcd.New(client)
}
