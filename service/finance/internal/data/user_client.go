package data

import (
	"context"
	"time"

	"finance/internal/conf"
	userv1 "user/api/user/v1"

	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type UserClient struct {
	client userv1.UserClient
	log    *log.Helper
}

func NewUserClient(registryConf *conf.Registry, logger log.Logger) (*UserClient, error) {
	helper := log.NewHelper(logger)
	if registryConf == nil || registryConf.Etcd == nil || len(registryConf.Etcd.Endpoints) == 0 {
		helper.Warn("user client registry config not found, using empty client")
		return &UserClient{log: helper}, nil
	}

	dialTimeout := 5 * time.Second
	if registryConf.Etcd.DialTimeout != nil {
		dialTimeout = registryConf.Etcd.DialTimeout.AsDuration()
	}

	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints:   registryConf.Etcd.Endpoints,
		DialTimeout: dialTimeout,
	})
	if err != nil {
		return nil, err
	}

	discovery := etcd.New(etcdClient)
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///user"),
		grpc.WithDiscovery(discovery),
		grpc.WithTimeout(5*time.Second),
		grpc.WithMiddleware(recovery.Recovery()),
	)
	if err != nil {
		return nil, err
	}

	return &UserClient{
		client: userv1.NewUserClient(conn),
		log:    helper,
	}, nil
}

func (c *UserClient) GetUser(ctx context.Context, userID uint32) (*userv1.UserInfo, error) {
	if c.client == nil {
		return nil, nil
	}
	return c.client.GetUser(ctx, &userv1.GetUserRequest{Id: userID})
}

func (c *UserClient) AdjustUserAsset(ctx context.Context, userID uint32, balanceDelta, workPointsDelta float64, reason, requestID string) (*userv1.UserInfo, error) {
	if c.client == nil {
		return nil, nil
	}
	return c.client.AdjustUserAsset(ctx, &userv1.AdjustUserAssetRequest{
		UserId:          userID,
		BalanceDelta:    balanceDelta,
		WorkPointsDelta: workPointsDelta,
		Reason:          reason,
		RequestId:       requestID,
	})
}
