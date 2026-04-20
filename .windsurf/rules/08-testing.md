---
trigger: manual
description: Kratos 测试规范
globs: ["**/*_test.go", "**/test/*.go"]
---

# Kratos 测试规范

## 1. 测试文件结构

```
service/{name}/
├── test/
│   └── grpc_test.go      # gRPC 集成测试
├── internal/
│   ├── biz/
│   │   └── user_test.go  # 业务层单元测试
│   ├── data/
│   │   └── user_test.go  # 数据层单元测试
│   └── service/
│       └── user_test.go  # 服务层单元测试
```

## 2. 单元测试 (biz)

```go
package biz

import (
    "context"
    "testing"
    
    "github.com/go-kratos/kratos/v2/log"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// Mock UserRepo
type mockUserRepo struct {
    mock.Mock
}

func (m *mockUserRepo) CreateUser(ctx context.Context, u *User) (*User, error) {
    args := m.Called(ctx, u)
    return args.Get(0).(*User), args.Error(1)
}

func (m *mockUserRepo) GetUserByID(ctx context.Context, id uint32) (*User, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(*User), args.Error(1)
}

// Test 注册
func TestUserUsecase_Register(t *testing.T) {
    // Arrange
    repo := new(mockUserRepo)
    uc := NewUserUsecase(repo, nil, log.DefaultLogger)
    
    ctx := context.Background()
    user := &User{
        Username: "test",
        Phone:    "13800138000",
        Password: "123456",
    }
    
    repo.On("GetUserByPhone", ctx, user.Phone).Return(nil, nil)
    repo.On("CreateUser", ctx, mock.Anything).Return(&User{ID: 1}, nil)
    
    // Act
    result, err := uc.Register(ctx, user)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, uint32(1), result.ID)
    repo.AssertExpectations(t)
}
```

## 3. 集成测试 (gRPC)

```go
package test

import (
    "context"
    "testing"
    
    v1 "{service}/api/{service}/v1"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

var (
    userClient v1.UserClient
    conn       *grpc.ClientConn
)

func init() {
    var err error
    conn, err = grpc.Dial("127.0.0.1:50051", 
        grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        panic(err)
    }
    userClient = v1.NewUserClient(conn)
}

func TestCreateUser(t *testing.T) {
    ctx := context.Background()
    
    resp, err := userClient.Register(ctx, &v1.RegisterRequest{
        Username: "test_user",
        Phone:    "13888888888",
        Password: "password123",
    })
    
    if err != nil {
        t.Fatalf("create user failed: %v", err)
    }
    
    t.Logf("created user id: %d", resp.User.Id)
}

func TestGetUser(t *testing.T) {
    ctx := context.Background()
    
    resp, err := userClient.GetUser(ctx, &v1.GetUserRequest{
        Id: 1,
    })
    
    if err != nil {
        t.Fatalf("get user failed: %v", err)
    }
    
    t.Logf("user: %+v", resp)
}
```

## 4. 测试命令

```bash
# 运行所有测试
go test ./...

# 运行特定包测试
go test ./internal/biz/...

# 运行集成测试 (需要服务启动)
go test ./test/...

# 带覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 5. Makefile 测试任务

```makefile
.PHONY: test test-cover

test:
	go test -v ./...

test-cover:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
```

## 6. 测试数据库

```go
// test_helper.go
package test

import (
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
    dsn := "root:root@tcp(127.0.0.1:3306)/test_db?charset=utf8mb4"
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        panic(err)
    }
    
    // 迁移表
    db.AutoMigrate(&User{})
    
    return db
}

func teardownTestDB(db *gorm.DB) {
    // 清理数据
    db.Exec("TRUNCATE TABLE users")
}
```

## 参考资料

- [Go 测试](https://golang.org/pkg/testing/)
- [Testify](https://github.com/stretchr/testify)
- [Kratos 测试示例](https://github.com/go-kratos/kratos/tree/main/examples)
