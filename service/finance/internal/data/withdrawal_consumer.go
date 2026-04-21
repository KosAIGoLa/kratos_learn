package data

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

// WithdrawalConsumer 提款消费者
type WithdrawalConsumer struct {
	data *Data
	log  *log.Helper
}

// NewWithdrawalConsumer 创建提款消费者
func NewWithdrawalConsumer(data *Data, logger log.Logger) *WithdrawalConsumer {
	return &WithdrawalConsumer{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// Start 启动消费者
func (c *WithdrawalConsumer) Start(ctx context.Context) error {
	if c.data.rabbitmq == nil {
		c.log.Warn("rabbitmq not configured, consumer not started")
		return nil
	}

	handler := func(msg *WithdrawalMessage) error {
		if msg.RequestID == "" {
			c.log.Errorf("withdrawal message missing request_id, cannot ensure idempotency")
			return fmt.Errorf("missing request_id")
		}

		// 基于请求ID生成幂等键，精确识别每笔独立提现。
		lockKey := fmt.Sprintf("withdrawal:idempotency:%s", msg.RequestID)

		// 尝试获取幂等锁，相同 request_id 只允许处理一次。
		result, err := c.data.rdb.SetArgs(ctx, lockKey, "1", redis.SetArgs{
			Mode: "NX",
			TTL:  24 * time.Hour,
		}).Result()
		if err != nil {
			c.log.Errorf("failed to acquire lock: %v", err)
			return err
		}
		if result != "OK" {
			c.log.Warnf("duplicate withdrawal detected, skip processing: request_id=%s", msg.RequestID)
			return nil // 重复请求，直接返回成功，不处理
		}

		// 创建提款记录
		withdrawal := &Withdrawal{
			UserID:    msg.UserID,
			Phone:     msg.Phone,
			Name:      msg.Name,
			BankCard:  msg.BankCard,
			BankName:  msg.BankName,
			Amount:    msg.Amount,
			Status:    0, // 待审核
			CreatedAt: time.Now(),
		}

		err = c.data.db.WithContext(ctx).Create(withdrawal).Error
		if err != nil {
			// 入库失败时释放锁，允许消息重试。
			c.data.rdb.Del(ctx, lockKey)
			c.log.Errorf("failed to create withdrawal record: %v", err)
			return err
		}

		c.log.Infof("withdrawal record created from queue: user_id=%d, amount=%.2f",
			msg.UserID, msg.Amount)
		return nil
	}

	return c.data.rabbitmq.ConsumeWithdrawals(handler)
}
