package data

import (
	"context"
	"finance/internal/biz"
	"finance/internal/conf"
	slog "log"
	"os"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/extra/redisotel/v9"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData, NewDB, NewRedis, NewRabbitMQClient, NewUserClient,
	NewRechargeRepo, NewWithdrawalRepo, NewIncomeLogRepo, NewBalanceLogRepo, NewCheckInRepo, NewUserAssetRepo,
	// 绑定 RabbitMQClient 到 biz.WithdrawalMessageQueue 接口
	wire.Bind(new(biz.WithdrawalMessageQueue), new(*RabbitMQClient)),
)

// Data .
type Data struct {
	db       *gorm.DB
	rdb      *redis.Client
	rabbitmq *RabbitMQClient
	user     *UserClient
}

// NewData .
func NewData(c *conf.Data, logger log.Logger, db *gorm.DB, rdb *redis.Client, rabbitmq *RabbitMQClient, userClient *UserClient) (*Data, func(), error) {
	data := &Data{db: db, rdb: rdb, rabbitmq: rabbitmq, user: userClient}

	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
		if rabbitmq != nil {
			rabbitmq.Close()
		}
	}

	// 启动 RabbitMQ 消费者
	if rabbitmq != nil {
		consumer := NewWithdrawalConsumer(data, logger)
		if err := consumer.Start(context.Background()); err != nil {
			cleanup()
			return nil, nil, err
		}
	}

	return data, cleanup, nil
}

func (d *Data) UserClient() *UserClient {
	return d.user
}

// NewDB .
func NewDB(c *conf.Data) *gorm.DB {
	newLogger := logger.New(
		slog.New(os.Stdout, "\r\n", slog.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			Colorful:      true,
			LogLevel:      logger.Info,
		},
	)
	db, err := gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{
		Logger:                                   newLogger,
		DisableForeignKeyConstraintWhenMigrating: true,
		NamingStrategy:                           schema.NamingStrategy{},
	})
	if err != nil {
		panic("failed to connect database")
	}
	return db
}

// NewRedis .
func NewRedis(c *conf.Data) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:         c.Redis.Addr,
		Password:     c.Redis.Password,
		DB:           int(c.Redis.Db),
		DialTimeout:  c.Redis.DialTimeout.AsDuration(),
		WriteTimeout: c.Redis.WriteTimeout.AsDuration(),
		ReadTimeout:  c.Redis.ReadTimeout.AsDuration(),
	})
	if err := redisotel.InstrumentTracing(rdb); err != nil {
		panic(err)
	}
	return rdb
}
