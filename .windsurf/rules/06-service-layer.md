---
trigger: manual
description: Kratos 服务层规范 (Service Layer)
globs: ["**/internal/service/*.go"]
---

# Kratos 服务层规范

## 1. 文件结构

```
internal/service/
├── service.go      # ProviderSet
└── user.go         # User 服务实现
```

## 2. ProviderSet (service.go)

```go
package service

import "github.com/google/wire"

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewUserService)
```

## 3. 服务实现

```go
package service

import (
    "context"
    
    v1 "{service}/api/{service}/v1"
    "{service}/internal/biz"
    
    "github.com/go-kratos/kratos/v2/errors"
    "github.com/go-kratos/kratos/v2/log"
)

// UserService 用户服务
type UserService struct {
    v1.UnimplementedUserServer
    uc  *biz.UserUsecase
    log *log.Helper
}

// NewUserService 创建服务 (Wire Provider)
func NewUserService(uc *biz.UserUsecase, logger log.Logger) *UserService {
    return &UserService{
        uc:  uc,
        log: log.NewHelper(logger),
    }
}

// Register 用户注册
func (s *UserService) Register(ctx context.Context, req *v1.RegisterRequest) (*v1.LoginResponse, error) {
    // 1. 参数校验
    if req.Username == "" || req.Password == "" {
        return nil, errors.BadRequest("PARAM_ERROR", "用户名和密码不能为空")
    }
    
    // 2. 解析邀请码获取 parent_id
    var parentID uint32
    if req.InviteCode != "" {
        parentUser, err := s.uc.GetUserByInviteCode(ctx, req.InviteCode)
        if err == nil {
            parentID = parentUser.ID
        }
    }
    
    // 3. 调用用例创建用户
    user, err := s.uc.Register(ctx, &biz.User{
        Username: req.Username,
        Phone:    req.Phone,
        Password: req.Password,
        ParentID: &parentID,
    })
    if err != nil {
        return nil, err
    }
    
    // 4. 异步发放邀请奖励
    if parentID > 0 {
        go func() {
            if err := s.uc.RewardInviteWorkPoints(context.Background(), parentID); err != nil {
                s.log.Errorf("邀请奖励发放失败: %v", err)
            }
        }()
    }
    
    // 5. 生成 Token
    token, _ := s.jwtManager.GenerateToken(user.ID, user.Username)
    
    // 6. 返回响应
    return &v1.LoginResponse{
        User:  s.toProtoUser(user),
        Token: token,
    }, nil
}

// GetUser 获取用户信息
func (s *UserService) GetUser(ctx context.Context, req *v1.GetUserRequest) (*v1.UserInfo, error) {
    user, err := s.uc.GetUser(ctx, req.Id)
    if err != nil {
        return nil, err
    }
    return s.toProtoUser(user), nil
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(ctx context.Context, req *v1.UpdateUserRequest) (*v1.UserInfo, error) {
    user, err := s.uc.UpdateUser(ctx, &biz.User{
        ID:     req.Id,
        Name:   req.Name,
        IDCard: req.IdCard,
    })
    if err != nil {
        return nil, err
    }
    return s.toProtoUser(user), nil
}

// toProtoUser 转换为 protobuf 格式
func (s *UserService) toProtoUser(u *biz.User) *v1.UserInfo {
    return &v1.UserInfo{
        Id:       u.ID,
        Username: u.Username,
        Phone:    u.Phone,
        Name:     u.Name,
        ParentId: func() uint32 {
            if u.ParentID != nil {
                return *u.ParentID
            }
            return 0
        }(),
        Balance:    u.Balance,
        WorkPoints: u.WorkPoints,
        Status:     u.Status,
    }
}
```

## 4. 服务层职责

| 职责 | 说明 | 示例 |
|------|------|------|
| 参数校验 | 检查请求参数合法性 | 非空、格式、范围检查 |
| 协议转换 | proto ↔ biz 模型转换 | `toProtoUser()`, `toBizUser()` |
| 调用用例 | 执行业务逻辑 | `uc.Register()`, `uc.GetUser()` |
| 错误处理 | 返回 gRPC 错误码 | `errors.BadRequest()`, `errors.NotFound()` |
| 响应组装 | 构建 proto 响应 | `&v1.UserInfo{...}` |

## 5. 错误码规范

```go
import "github.com/go-kratos/kratos/v2/errors"

// 400 Bad Request
errors.BadRequest("PARAM_ERROR", "参数错误")
errors.BadRequest("INVALID_PHONE", "手机号格式错误")

// 401 Unauthorized
errors.Unauthorized("TOKEN_EXPIRED", "token 已过期")

// 403 Forbidden
errors.Forbidden("NO_PERMISSION", "无权限")

// 404 Not Found
errors.NotFound("USER_NOT_FOUND", "用户不存在")

// 409 Conflict
errors.Conflict("USER_EXISTS", "用户已存在")

// 500 Internal Server Error
errors.InternalServer("SYSTEM_ERROR", "系统错误")
```

## 下一步

- [服务器配置](07-server-setup.md) → 服务器启动与注册
