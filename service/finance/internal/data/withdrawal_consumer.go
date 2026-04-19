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
		// 生成分布式锁 Key (基于用户ID、银行卡号、金额和时间戳精确到分钟)
		lockKey := fmt.Sprintf("withdrawal:lock:%d:%s:%.2f:%s",
			msg.UserID, msg.BankCard, msg.Amount, time.Now().Format("200601021504"))

		// 尝试获取分布式锁 (过期时间30秒)
		result, err := c.data.rdb.SetArgs(ctx, lockKey, "1", redis.SetArgs{
			Mode: "NX",
			TTL:  30 * time.Second,
		}).Result()
		if err != nil {
			c.log.Errorf("failed to acquire lock: %v", err)
			return err
		}
		if result != "OK" {
			c.log.Warnf("duplicate withdrawal detected, skip processing: user_id=%d, amount=%.2f", msg.UserID, msg.Amount)
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
			// 释放锁
			c.data.rdb.Del(ctx, lockKey)
			c.log.Errorf("failed to create withdrawal record: %v", err)
			return err
		}

		// 处理成功后删除锁（或等待过期）
		c.data.rdb.Del(ctx, lockKey)

		c.log.Infof("withdrawal record created from queue: user_id=%d, amount=%.2f",
			msg.UserID, msg.Amount)
		return nil
	}

	return c.data.rabbitmq.ConsumeWithdrawals(handler)
}
