---
trigger: manual
description: Kratos 数据层规范 (Data Layer)
globs: ["**/internal/data/*.go"]
---

# Kratos 数据层规范

## 1. 文件结构

```
internal/data/
├── data.go           # Data 结构体与 ProviderSet
├── user.go           # User 模型与仓储实现
├── team.go           # Team 模型与仓储 (可选)
└── system_config.go  # 系统配置仓储 (可选)
```

## 2. Data 结构体 (data.go)

```go
package data

import (
    "github.com/go-kratos/kratos/v2/log"
    "github.com/google/wire"
    "github.com/redis/go-redis/v9"
    "gorm.io/gorm"
    "{service}/internal/conf"
)

var ProviderSet = wire.NewSet(
    NewData,
    NewDB,
    NewRedis,
    NewUserRepo,
)

type Data struct {
    db  *gorm.DB
    rdb *redis.Client
}

func NewData(c *conf.Data, logger log.Logger, db *gorm.DB, rdb *redis.Client) (*Data, func(), error) {
    cleanup := func() {
        log.NewHelper(logger).Info("closing the data resources")
    }
    return &Data{db: db, rdb: rdb}, cleanup, nil
}

// NewDB 创建 MySQL 连接
func NewDB(c *conf.Data) *gorm.DB {
    db, err := gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    if err != nil {
        panic("failed to connect database")
    }
    return db
}

// NewRedis 创建 Redis 连接
func NewRedis(c *conf.Data) *redis.Client {
    return redis.NewClient(&redis.Options{
        Addr: c.Redis.Addr,
    })
}
```

## 3. 数据模型定义

```go
// User 数据模型 (GORM)
type User struct {
    ID         uint32         `gorm:"primarykey"`
    Username   string         `gorm:"uniqueIndex;size:50;not null"`
    Phone      string         `gorm:"uniqueIndex;size:11"`
    Password   string         `gorm:"size:255;not null"`
    ParentID   *uint32        `gorm:"index"`
    Balance    float64        `gorm:"decimal(15,2);default:0.00"`
    WorkPoints float64        `gorm:"decimal(15,2);default:0.00"`
    Status     int8           `gorm:"index;default:1"`
    CreatedAt  time.Time
    UpdatedAt  time.Time
    DeletedAt  gorm.DeletedAt `gorm:"index"`
}

// TableName 自定义表名
func (User) TableName() string {
    return "users"
}
```

## 4. 仓储实现

```go
// userRepo 仓储实现
type userRepo struct {
    data *Data
    log  *log.Helper
}

// NewUserRepo 创建仓储 (Wire Provider)
func NewUserRepo(data *Data, logger log.Logger) biz.UserRepo {
    return &userRepo{
        data: data,
        log:  log.NewHelper(logger),
    }
}

// CreateUser 创建用户
func (r *userRepo) CreateUser(ctx context.Context, u *biz.User) (*biz.User, error) {
    user := User{
        Username: u.Username,
        Phone:    u.Phone,
        Password: u.Password,
        ParentID: u.ParentID,
    }
    
    if err := r.data.db.Create(&user).Error; err != nil {
        return nil, status.Errorf(codes.Internal, "创建用户失败: %s", err)
    }
    
    return r.toBizUser(&user), nil
}

// GetUserByID 根据 ID 查询
func (r *userRepo) GetUserByID(ctx context.Context, id uint32) (*biz.User, error) {
    var user User
    if err := r.data.db.First(&user, id).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, status.Errorf(codes.NotFound, "用户不存在")
        }
        return nil, status.Errorf(codes.Internal, "查询失败: %s", err)
    }
    return r.toBizUser(&user), nil
}

// UpdateUser 更新用户
func (r *userRepo) UpdateUser(ctx context.Context, u *biz.User) (*biz.User, error) {
    updates := map[string]interface{}{
        "name":    u.Name,
        "id_card": u.IDCard,
    }
    
    if err := r.data.db.Model(&User{}).Where("id = ?", u.ID).Updates(updates).Error; err != nil {
        return nil, status.Errorf(codes.Internal, "更新失败: %s", err)
    }
    
    return r.GetUserByID(ctx, u.ID)
}

// 事务 + 行锁更新示例
func (r *userRepo) UpdateWorkPoints(ctx context.Context, userID uint32, points float64) error {
    return r.data.db.Transaction(func(tx *gorm.DB) error {
        var user User
        // 加行锁 FOR UPDATE
        if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
            Where("id = ?", userID).
            First(&user).Error; err != nil {
            return err
        }
        
        user.WorkPoints += points
        return tx.Save(&user).Error
    })
}

// 模型转换
func (r *userRepo) toBizUser(u *User) *biz.User {
    return &biz.User{
        ID:         u.ID,
        Username:   u.Username,
        Phone:      u.Phone,
        ParentID:   u.ParentID,
        Balance:    u.Balance,
        WorkPoints: u.WorkPoints,
        Status:     int32(u.Status),
        CreatedAt:  u.CreatedAt,
    }
}
```

## 5. 仓储接口 (在 biz 层定义)

```go
// UserRepo 用户仓储接口
// 定义在 internal/biz/user.go
type UserRepo interface {
    CreateUser(ctx context.Context, u *User) (*User, error)
    GetUserByID(ctx context.Context, id uint32) (*User, error)
    GetUserByPhone(ctx context.Context, phone string) (*User, error)
    UpdateUser(ctx context.Context, u *User) (*User, error)
    DeleteUser(ctx context.Context, id uint32) error
}
```

## 下一步

- [业务层规范](05-biz-layer.md) → 业务逻辑与用例
