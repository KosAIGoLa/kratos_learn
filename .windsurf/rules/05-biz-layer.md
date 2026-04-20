---
trigger: manual
description: Kratos 业务层规范 (Biz Layer)
globs: ["**/internal/biz/*.go"]
---

# Kratos 业务层规范

## 1. 文件结构

```
internal/biz/
├── biz.go        # ProviderSet
├── user.go       # User 实体、仓储接口、用例
├── team.go       # Team 相关 (可选)
└── kyc.go        # KYC 相关 (可选)
```

## 2. ProviderSet (biz.go)

```go
package biz

import "github.com/google/wire"

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
    NewUserUsecase,
    NewTeamUsecase,
)
```

## 3. 实体定义 (Entity)

```go
// User 用户领域模型
type User struct {
    ID         uint32
    Username   string
    Phone      string
    Password   string
    Name       string
    IDCard     string
    ParentID   *uint32
    Balance    float64
    WorkPoints float64
    Status     int32
    CreatedAt  time.Time
    UpdatedAt  time.Time
}
```

## 4. 仓储接口定义

```go
// UserRepo 用户仓储接口
// 由 data 层实现
type UserRepo interface {
    CreateUser(ctx context.Context, u *User) (*User, error)
    GetUserByID(ctx context.Context, id uint32) (*User, error)
    GetUserByPhone(ctx context.Context, phone string) (*User, error)
    GetUserByUsername(ctx context.Context, username string) (*User, error)
    GetUserByInviteCode(ctx context.Context, inviteCode string) (*User, error)
    UpdateUser(ctx context.Context, u *User) (*User, error)
    DeleteUser(ctx context.Context, id uint32) error
    UpdateWorkPoints(ctx context.Context, userID uint32, points float64) error
}

// ConfigRepo 配置仓储接口
type ConfigRepo interface {
    GetConfig(ctx context.Context, key string) (*SystemConfig, error)
}

type SystemConfig struct {
    ID          uint32
    Key         string
    Value       string
    Description string
    Group       string
}
```

## 5. 用例实现 (Usecase)

```go
// UserUsecase 用户用例
type UserUsecase struct {
    repo       UserRepo
    configRepo ConfigRepo
    log        *log.Helper
}

// NewUserUsecase 创建用例 (Wire Provider)
func NewUserUsecase(repo UserRepo, configRepo ConfigRepo, logger log.Logger) *UserUsecase {
    return &UserUsecase{
        repo:       repo,
        configRepo: configRepo,
        log:        log.NewHelper(logger),
    }
}

// Register 用户注册
func (uc *UserUsecase) Register(ctx context.Context, u *User) (*User, error) {
    // 1. 参数校验
    if u.Username == "" || u.Phone == "" {
        return nil, errors.New(400, "PARAM_ERROR", "用户名和手机号不能为空")
    }
    
    // 2. 检查用户是否存在
    existing, _ := uc.repo.GetUserByPhone(ctx, u.Phone)
    if existing != nil {
        return nil, errors.New(400, "USER_EXISTS", "用户已存在")
    }
    
    // 3. 密码加密
    hashedPwd, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, err
    }
    u.Password = string(hashedPwd)
    
    // 4. 创建用户
    return uc.repo.CreateUser(ctx, u)
}

// GetUser 获取用户信息
func (uc *UserUsecase) GetUser(ctx context.Context, id uint32) (*User, error) {
    return uc.repo.GetUserByID(ctx, id)
}

// UpdateUser 更新用户信息
func (uc *UserUsecase) UpdateUser(ctx context.Context, u *User) (*User, error) {
    return uc.repo.UpdateUser(ctx, u)
}

// RewardInviteWorkPoints 邀请奖励工分
func (uc *UserUsecase) RewardInviteWorkPoints(ctx context.Context, parentID uint32) error {
    if parentID == 0 {
        return nil
    }
    
    // 获取奖励配置
    config, err := uc.configRepo.GetConfig(ctx, "invite.reward_work_points")
    if err != nil {
        uc.log.Warnf("获取奖励配置失败: %v", err)
        return nil
    }
    
    points, err := strconv.ParseFloat(config.Value, 64)
    if err != nil || points <= 0 {
        return nil
    }
    
    // 发放奖励
    if err := uc.repo.UpdateWorkPoints(ctx, parentID, points); err != nil {
        uc.log.Errorf("发放奖励失败: %v", err)
        return err
    }
    
    uc.log.Infof("邀请奖励已发放: parent_id=%d, points=%.2f", parentID, points)
    return nil
}
```

## 6. 业务错误处理

```go
import "github.com/go-kratos/kratos/v2/errors"

// 定义业务错误
var (
    ErrUserNotFound = errors.NotFound("USER_NOT_FOUND", "用户不存在")
    ErrUserExists   = errors.Conflict("USER_EXISTS", "用户已存在")
    ErrInvalidParam = errors.BadRequest("INVALID_PARAM", "参数错误")
)

// 在用例中使用
func (uc *UserUsecase) GetUser(ctx context.Context, id uint32) (*User, error) {
    user, err := uc.repo.GetUserByID(ctx, id)
    if err != nil {
        return nil, ErrUserNotFound
    }
    return user, nil
}
```

## 下一步

- [服务层规范](06-service-layer.md) → gRPC/HTTP 服务实现
