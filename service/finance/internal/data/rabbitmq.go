package data

import (
	"context"
	"encoding/json"
	"finance/internal/biz"
	"finance/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/rabbitmq/amqp091-go"
)

// RabbitMQClient RabbitMQ 客户端
type RabbitMQClient struct {
	conn       *amqp091.Connection
	channel    *amqp091.Channel
	exchange   string
	queue      string
	routingKey string
	log        *log.Helper
}

// WithdrawalMessage 提款消息
type WithdrawalMessage struct {
	UserID   uint32  `json:"user_id"`
	Amount   float64 `json:"amount"`
	BankCard string  `json:"bank_card"`
	BankName string  `json:"bank_name"`
	Phone    string  `json:"phone"`
	Name     string  `json:"name"`
}

// NewRabbitMQClient 创建 RabbitMQ 客户端
func NewRabbitMQClient(c *conf.Data, logger log.Logger) (*RabbitMQClient, func(), error) {
	if c.Rabbitmq == nil || c.Rabbitmq.Addr == "" {
		log.NewHelper(logger).Warn("rabbitmq not configured")
		return nil, func() {}, nil
	}

	conn, err := amqp091.Dial(c.Rabbitmq.Addr)
	if err != nil {
		return nil, nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, nil, err
	}

	exchange := c.Rabbitmq.Exchange
	if exchange == "" {
		exchange = "finance.exchange"
	}
	queue := c.Rabbitmq.Queue
	if queue == "" {
		queue = "withdrawal.queue"
	}
	routingKey := c.Rabbitmq.RoutingKey
	if routingKey == "" {
		routingKey = "withdrawal"
	}

	// 声明交换器
	err = ch.ExchangeDeclare(
		exchange, // name
		"direct", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, nil, err
	}

	// 声明队列
	_, err = ch.QueueDeclare(
		queue, // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, nil, err
	}

	// 绑定队列到交换器
	err = ch.QueueBind(
		queue,      // queue name
		routingKey, // routing key
		exchange,   // exchange
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, nil, err
	}

	client := &RabbitMQClient{
		conn:       conn,
		channel:    ch,
		exchange:   exchange,
		queue:      queue,
		routingKey: routingKey,
		log:        log.NewHelper(logger),
	}

	cleanup := func() {
		client.Close()
	}

	client.log.Info("rabbitmq client initialized")
	return client, cleanup, nil
}

// PublishWithdrawal 发布提款消息 (实现 biz.WithdrawalMessageQueue 接口)
func (c *RabbitMQClient) PublishWithdrawal(ctx context.Context, w *biz.Withdrawal) error {
	if c == nil || c.channel == nil {
		return nil
	}

	msg := &WithdrawalMessage{
		UserID:   w.UserID,
		Amount:   w.Amount,
		BankCard: w.BankCard,
		BankName: w.BankName,
		Phone:    w.Phone,
		Name:     w.Name,
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	err = c.channel.PublishWithContext(
		ctx,
		c.exchange,   // exchange
		c.routingKey, // routing key
		false,        // mandatory
		false,        // immediate
		amqp091.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp091.Persistent, // 持久化消息
			Body:         body,
		},
	)
	if err != nil {
		c.log.Errorf("failed to publish withdrawal message: %v", err)
		return err
	}

	c.log.Infof("withdrawal message published: user_id=%d, amount=%.2f", msg.UserID, msg.Amount)
	return nil
}

// ConsumeWithdrawals 消费提款消息
func (c *RabbitMQClient) ConsumeWithdrawals(handler func(*WithdrawalMessage) error) error {
	if c == nil || c.channel == nil {
		return nil
	}

	msgs, err := c.channel.Consume(
		c.queue, // queue
		"",      // consumer
		false,   // auto-ack (关闭自动确认，手动确认)
		false,   // exclusive
		false,   // no-local
		false,   // no-wait
		nil,     // args
	)
	if err != nil {
		return err
	}

	go func() {
		for msg := range msgs {
			var withdrawalMsg WithdrawalMessage
			if err := json.Unmarshal(msg.Body, &withdrawalMsg); err != nil {
				c.log.Errorf("failed to unmarshal withdrawal message: %v", err)
				_ = msg.Nack(false, false) // 拒绝消息，不重入队
				continue
			}

			c.log.Infof("processing withdrawal message: user_id=%d, amount=%.2f",
				withdrawalMsg.UserID, withdrawalMsg.Amount)

			if err := handler(&withdrawalMsg); err != nil {
				c.log.Errorf("failed to process withdrawal: %v", err)
				_ = msg.Nack(false, true) // 重入队，稍后重试
				continue
			}

			_ = msg.Ack(false) // 确认消息已处理
			c.log.Infof("withdrawal message processed successfully: user_id=%d", withdrawalMsg.UserID)
		}
	}()

	c.log.Info("withdrawal consumer started")
	return nil
}

// Close 关闭连接
func (c *RabbitMQClient) Close() {
	if c.channel != nil {
		_ = c.channel.Close()
	}
	if c.conn != nil {
		_ = c.conn.Close()
	}
}
