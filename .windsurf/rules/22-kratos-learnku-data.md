---
trigger: manual
description: Kratos微服務實戰指南 (三) 數據層實現 (LearnKu文章64942)
globs: 
---

## Kratos 微服務實戰之用戶服務 (三) 數據層實現 (Data Layer)

參考文章: https://learnku.com/articles/64942

---

### 1. 修改數據層文件

#### 修改 `user/internal/data/data.go`

```go
package data

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/extra/redisotel"
	"github.com/go-redis/redis/v8"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	slog "log"
	"os"
	"time"
	"user/internal/conf"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewDB, NewRedis, NewUserRepo)

type Data struct {
	db  *gorm.DB
	rdb *redis.Client
}

// NewData .
func NewData(c *conf.Data, logger log.Logger, db *gorm.DB, rdb *redis.Client) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{db: db, rdb: rdb}, cleanup, nil
}

// NewDB .
func NewDB(c *conf.Data) *gorm.DB {
	// 終端打印輸入 sql 執行記錄
	newLogger := logger.New(
		slog.New(os.Stdout, "\r\n", slog.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // 慢查詢 SQL 閾值
			Colorful:      true,        // 禁用彩色打印
			LogLevel:      logger.Info, // Log level
		},
	)

	db, err := gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{
		Logger:                                   newLogger,
		DisableForeignKeyConstraintWhenMigrating: true,
		NamingStrategy: schema.NamingStrategy{
			//SingularTable: true, // 表名是否加 s
		},
	})
	if err != nil {
		log.Errorf("failed opening connection to sqlite: %v", err)
		panic("failed to connect database")
	}
	return db
}

func NewRedis(c *conf.Data) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:         c.Redis.Addr,
		Password:     c.Redis.Password,
		DB:           int(c.Redis.Db),
		DialTimeout:  c.Redis.DialTimeout.AsDuration(),
		WriteTimeout: c.Redis.WriteTimeout.AsDuration(),
		ReadTimeout:  c.Redis.ReadTimeout.AsDuration(),
	})
	rdb.AddHook(redisotel.TracingHook{})
	return rdb
}
```

> wire 概念如果不熟悉的話，請參看 [Wire 依賴注入](https://go-kratos.dev/docs/guide/wire)

---

#### 修改或刪除 `user/internal/data/greeter.go` 為 `user.go`

```go
package data

import (
	"context"
	"crypto/sha512"
	"fmt"
	"github.com/anaskhan96/go-password-encoder"
	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"time"
	"user/internal/biz"
)

// 定義數據表結構體
type User struct {
	ID          int64 `gorm:"primarykey"`
	Mobile      string `gorm:"index:idx_mobile;unique;type:varchar(11) comment '手機號碼，用戶唯一標識';not null"`
	Password    string `gorm:"type:varchar(100);not null "` // 用戶密碼的保存需要注意是否加密
	NickName    string `gorm:"type:varchar(25) comment '用戶暱稱'"`
	Birthday    *time.Time `gorm:"type:datetime comment '出生日日期'"`
	Gender      string `gorm:"column:gender;default:male;type:varchar(16) comment 'female:女,male:男'"`
	Role        int    `gorm:"column:role;default:1;type:int comment '1:普通用戶，2:管理員'"`
	CreatedAt   time.Time `gorm:"column:add_time"`
	UpdatedAt   time.Time `gorm:"column:update_time"`
	DeletedAt   gorm.DeletedAt
	IsDeletedAt bool
}

type userRepo struct {
	data *Data
	log  *log.Helper
}

// NewUserRepo . 這裡需要注意，上面 data 文件 wire 注入的是此方法，方法名不要寫錯了
func NewUserRepo(data *Data, logger log.Logger) biz.UserRepo {
	return &userRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// CreateUser .
func (r *userRepo) CreateUser(ctx context.Context, u *biz.User) (*biz.User, error) {
	var user User

	// 驗證是否已經創建
	result := r.data.db.Where(&User{Mobile: u.Mobile}).First(&user)
	if result.RowsAffected == 1 {
		return nil, status.Errorf(codes.AlreadyExists, "用戶已存在")
	}

	user.Mobile = u.Mobile
	user.NickName = u.NickName
	user.Password = encrypt(u.Password) // 密碼加密

	res := r.data.db.Create(&user)
	if res.Error != nil {
		return nil, status.Errorf(codes.Internal, res.Error.Error())
	}

	return &biz.User{
		ID:       user.ID,
		Mobile:   user.Mobile,
		Password: user.Password,
		NickName: user.NickName,
		Gender:   user.Gender,
		Role:     user.Role,
	}, nil
}

// Password encryption
func encrypt(psd string) string {
	options := &password.Options{SaltLen: 16, Iterations: 10000, KeyLen: 32, HashFunction: sha512.New}
	salt, encodedPwd := password.Encode(psd, options)
	return fmt.Sprintf("$pbkdf2-sha512$%s$%s", salt, encodedPwd)
}
```

---

### 下一步

完成數據層實現後，請參考：

- `23-kratos-learnku-biz-service.md` - 業務層與服務層
- `24-kratos-learnku-server.md` - 服務器配置與啟動
- `25-kratos-learnku-testing.md` - 測試

---

### 參考資源

- 源碼: https://github.com/aliliin/kratos-shop
- Wire 依賴注入: https://go-kratos.dev/docs/guide/wire
- GORM 文檔: https://gorm.io/
