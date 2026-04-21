package data

import (
	"context"
	"time"

	slog "log"
	"os"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"

	"system/internal/conf"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewDB, NewRedis, NewMongoDB, NewConfigRepo, NewProfitSharingRepo, NewRiskControlRepo, NewDomainRepo, NewWhitelistRepo, NewSystemLogRepo, NewUserLogRepo)

// Data .
type Data struct {
	db  *gorm.DB
	rdb *redis.Client
	mdb *mongo.Database
}

// NewData .
func NewData(c *conf.Data, logger log.Logger, db *gorm.DB, rdb *redis.Client, mdb *mongo.Database) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
		if mdb != nil {
			_ = mdb.Client().Disconnect(context.Background())
		}
	}
	return &Data{db: db, rdb: rdb, mdb: mdb}, cleanup, nil
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
	if c.Database != nil && c.Database.SlaveSource != "" {
		if err := db.Use(dbresolver.Register(dbresolver.Config{
			Replicas: []gorm.Dialector{mysql.Open(c.Database.SlaveSource)},
		})); err != nil {
			panic("failed to register dbresolver: " + err.Error())
		}
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

// NewMongoDB 创建MongoDB连接
func NewMongoDB(c *conf.Data) *mongo.Database {
	if c.Mongodb == nil || c.Mongodb.Uri == "" {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.Mongodb.ConnectTimeout.AsDuration())
	defer cancel()

	clientOptions := options.Client().ApplyURI(c.Mongodb.Uri)

	// 设置认证信息 (如果配置了用户名密码)
	if c.Mongodb.Username != "" && c.Mongodb.Password != "" {
		auth := options.Credential{
			Username:   c.Mongodb.Username,
			Password:   c.Mongodb.Password,
			AuthSource: c.Mongodb.Database, // 认证数据库
		}
		clientOptions.SetAuth(auth)
	}

	if c.Mongodb.MaxPoolSize > 0 {
		clientOptions.SetMaxPoolSize(uint64(c.Mongodb.MaxPoolSize))
	}
	if c.Mongodb.MinPoolSize > 0 {
		clientOptions.SetMinPoolSize(uint64(c.Mongodb.MinPoolSize))
	}

	// mongo-driver v2: Connect 不需要 context 作为第一个参数
	client, err := mongo.Connect(clientOptions)
	if err != nil {
		panic("failed to connect mongodb: " + err.Error())
	}

	// 验证连接
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		panic("failed to ping mongodb: " + err.Error())
	}

	return client.Database(c.Mongodb.Database)
}
