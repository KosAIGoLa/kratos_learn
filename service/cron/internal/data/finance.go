package data

import (
	"context"
	"fmt"
	"time"

	"cron/internal/conf"
	financev1 "finance/api/finance/v1"

	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// FinanceClient finance 服务客户端
type FinanceClient struct {
	client financev1.FinanceClient
	log    *log.Helper
}

// NewFinanceClient 创建 finance 服务客户端
func NewFinanceClient(bc *conf.Bootstrap, logger log.Logger) (*FinanceClient, error) {
	if bc == nil || bc.Client == nil || bc.Client.Finance == nil {
		log.NewHelper(logger).Warn("finance client config not found, using default")
		// 返回空客户端，实际使用时需要配置
		return &FinanceClient{
			client: nil,
			log:    log.NewHelper(logger),
		}, nil
	}

	// 配置 etcd 注册中心
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"}, // 默认 etcd 地址
	})
	if err != nil {
		return nil, err
	}

	discovery := etcd.New(etcdClient)

	// 创建 gRPC 连接
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///finance"),
		grpc.WithDiscovery(discovery),
		grpc.WithTimeout(5*time.Second),
		grpc.WithMiddleware(recovery.Recovery()),
	)
	if err != nil {
		return nil, err
	}

	client := financev1.NewFinanceClient(conn)
	return &FinanceClient{
		client: client,
		log:    log.NewHelper(logger),
	}, nil
}

// CreateBalanceLog 创建余额变动记录
func (c *FinanceClient) CreateBalanceLog(ctx context.Context, userID uint32, amount, beforeBalance, afterBalance float64, remark string) error {
	if c.client == nil {
		c.log.Warn("finance client not initialized, skipping CreateBalanceLog")
		return nil
	}

	_, err := c.client.CreateBalanceLog(ctx, &financev1.CreateBalanceLogRequest{
		UserId:        userID,
		Type:          1, // 类型 1 = 算力转换
		Amount:        amount,
		BeforeBalance: beforeBalance,
		AfterBalance:  afterBalance,
		Remark:        remark,
	})
	if err != nil {
		c.log.Errorf("create balance log failed: %v", err)
		return err
	}
	return nil
}

// GetUserBalance 获取用户余额
func (c *FinanceClient) GetUserBalance(ctx context.Context, userID uint32) (float64, error) {
	if c.client == nil {
		return 0, nil
	}

	resp, err := c.client.GetUserBalance(ctx, &financev1.GetUserBalanceRequest{
		UserId: userID,
	})
	if err != nil {
		c.log.Errorf("get user balance failed: %v", err)
		return 0, err
	}
	return resp.Balance, nil
}

// ListPendingHashrateCompensations 查询待补偿的算力记录
func (c *FinanceClient) ListPendingHashrateCompensations(ctx context.Context, limit int) ([]*financev1.HashrateCompensationInfo, error) {
	if c.client == nil {
		c.log.Warn("finance client not initialized, skipping ListPendingHashrateCompensations")
		return nil, nil
	}

	resp, err := c.client.ListHashrateCompensations(ctx, &financev1.ListHashrateCompensationsRequest{
		Status: 0,
		Limit:  uint32(limit),
	})
	if err != nil {
		c.log.Errorf("list pending hashrate compensations failed: %v", err)
		return nil, err
	}
	return resp.Records, nil
}

// CompensateHashrate 执行算力补偿
func (c *FinanceClient) CompensateHashrate(ctx context.Context, id uint64) error {
	if c.client == nil {
		c.log.Warn("finance client not initialized, skipping CompensateHashrate")
		return nil
	}

	resp, err := c.client.CompensateHashrate(ctx, &financev1.CompensateHashrateRequest{
		Id: id,
	})
	if err != nil {
		c.log.Errorf("compensate hashrate failed: %v", err)
		return err
	}
	if !resp.Success {
		return fmt.Errorf("compensate hashrate failed: %s", resp.Message)
	}
	return nil
}
