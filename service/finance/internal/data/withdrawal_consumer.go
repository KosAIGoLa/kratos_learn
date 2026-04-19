package data

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
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

		err := c.data.db.WithContext(ctx).Create(withdrawal).Error
		if err != nil {
			c.log.Errorf("failed to create withdrawal record: %v", err)
			return err
		}

		c.log.Infof("withdrawal record created from queue: user_id=%d, amount=%.2f",
			msg.UserID, msg.Amount)
		return nil
	}

	return c.data.rabbitmq.ConsumeWithdrawals(handler)
}
