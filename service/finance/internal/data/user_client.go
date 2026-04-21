package data

import (
	"context"
	"fmt"
	"time"

	"finance/internal/conf"
	userv1 "user/api/user/v1"

	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	kratosgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

type UserClient struct {
	client userv1.UserClient
	conn   *grpc.ClientConn
	etcd   *clientv3.Client
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
	conn, err := kratosgrpc.DialInsecure(
		context.Background(),
		kratosgrpc.WithEndpoint("discovery:///user"),
		kratosgrpc.WithDiscovery(discovery),
		kratosgrpc.WithTimeout(5*time.Second),
		kratosgrpc.WithMiddleware(recovery.Recovery()),
	)
	if err != nil {
		return nil, err
	}

	return &UserClient{
		client: userv1.NewUserClient(conn),
		conn:   conn,
		etcd:   etcdClient,
		log:    helper,
	}, nil
}

func (c *UserClient) Close() error {
	if c.conn != nil {
		_ = c.conn.Close()
	}
	if c.etcd != nil {
		_ = c.etcd.Close()
	}
	return nil
}

func (c *UserClient) GetUser(ctx context.Context, userID uint32) (*userv1.UserInfo, error) {
	if c.client == nil {
		return nil, fmt.Errorf("user grpc client not initialized")
	}
	return c.client.GetUser(ctx, &userv1.GetUserRequest{Id: userID})
}

func (c *UserClient) AdjustUserAsset(ctx context.Context, userID uint32, balanceDelta, workPointsDelta float64, reason, requestID string) (*userv1.UserInfo, error) {
	if c.client == nil {
		return nil, fmt.Errorf("user grpc client not initialized")
	}
	return c.client.AdjustUserAsset(ctx, &userv1.AdjustUserAssetRequest{
		UserId:          userID,
		BalanceDelta:    balanceDelta,
		WorkPointsDelta: workPointsDelta,
		Reason:          reason,
		RequestId:       requestID,
	})
}
