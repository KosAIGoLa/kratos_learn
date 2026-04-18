---
trigger: manual
description: Kratos微服務實戰指南 (四) 業務層與服務層 (LearnKu文章64942)
globs: 
---

## Kratos 微服務實戰之用戶服務 (四) 業務層 (Biz) 與服務層 (Service)

參考文章: https://learnku.com/articles/64942

---

### 1. 實現業務層 (Biz)

#### 修改或刪除 `user/internal/biz/greeter.go` 為 `user.go`

```go
package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
)

// 定義返回數據結構體
type User struct {
	ID       int64
	Mobile   string
	Password string
	NickName string
	Birthday int64
	Gender   string
	Role     int
}

type UserRepo interface {
	CreateUser(context.Context, *User) (*User, error)
}

type UserUsecase struct {
	repo UserRepo
	log  *log.Helper
}

func NewUserUsecase(repo UserRepo, logger log.Logger) *UserUsecase {
	return &UserUsecase{repo: repo, log: log.NewHelper(logger)}
}

func (uc *UserUsecase) Create(ctx context.Context, u *User) (*User, error) {
	return uc.repo.CreateUser(ctx, u)
}
```

---

#### 修改 `user/internal/biz/biz.go` 文件

```go
package biz

import "github.com/google/wire"

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(NewUserUsecase)
```

---

### 2. 實現服務層 (Service)

#### 修改或刪除 `user/internal/service/greeter.go` 為 `user.go`

```go
package service

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	v1 "user/api/user/v1"
	"user/internal/biz"
)

type UserService struct {
	v1.UnimplementedUserServer
	uc  *biz.UserUsecase
	log *log.Helper
}

// NewUserService new a greeter service.
func NewUserService(uc *biz.UserUsecase, logger log.Logger) *UserService {
	return &UserService{uc: uc, log: log.NewHelper(logger)}
}

// CreateUser create a user
func (u *UserService) CreateUser(ctx context.Context, req *v1.CreateUserInfo) (*v1.UserInfoResponse, error) {
	user, err := u.uc.Create(ctx, &biz.User{
		Mobile:   req.Mobile,
		Password: req.Password,
		NickName: req.NickName,
	})
	if err != nil {
		return nil, err
	}

	userInfoRsp := v1.UserInfoResponse{
		Id:       user.ID,
		Mobile:   user.Mobile,
		Password: user.Password,
		NickName: user.NickName,
		Gender:   user.Gender,
		Role:     int32(user.Role),
		Birthday: user.Birthday,
	}
	return &userInfoRsp, nil
}
```

---

#### 修改 `user/internal/service/service.go` 文件

```go
package service

import "github.com/google/wire"

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewUserService)
```

---

### 下一步

完成業務層與服務層後，請參考：

- `24-kratos-learnku-server.md` - 服務器配置與啟動
- `25-kratos-learnku-testing.md` - 測試

---

### 參考資源

- 源碼: https://github.com/aliliin/kratos-shop
- Wire 依賴注入: https://go-kratos.dev/blog/go-project-wire
- Kratos 官方文檔: https://go-kratos.dev/docs/getting-started/start
