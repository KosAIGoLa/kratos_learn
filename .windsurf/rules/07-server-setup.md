---
trigger: manual
description: Kratos 服务器配置与启动规范
globs: ["**/internal/server/*.go", "**/cmd/**/main.go"]
---

# Kratos 服务器配置与启动规范

## 1. 文件结构

```
internal/server/
├── server.go      # ProviderSet + Registrar
├── grpc.go        # gRPC 服务器
└── http.go        # HTTP 服务器 (可选)

cmd/{name}/
├── main.go        # 启动入口
├── wire.go        # Wire 注入定义
└── wire_gen.go    # 生成代码
```

## 2. ProviderSet (server.go)

```go
package server

import (
    "github.com/go-kratos/kratos/v2/registry"
    "github.com/google/wire"
    "{service}/internal/conf"
    
    etcd "github.com/go-kratos/kratos/contrib/registry/etcd/v2"
    clientv3 "go.etcd.io/etcd/client/v3"
)

var ProviderSet = wire.NewSet(NewGRPCServer, NewRegistrar)

// NewRegistrar 服务注册 (etcd)
func NewRegistrar(conf *conf.Registry) registry.Registrar {
    if conf == nil || conf.Etcd == nil {
        return nil
    }
    
    client, err := clientv3.New(clientv3.Config{
        Endpoints:   conf.Etcd.Endpoints,
        DialTimeout: conf.Etcd.DialTimeout.AsDuration(),
    })
    if err != nil {
        panic(err)
    }
    
    return etcd.New(client)
}
```

## 3. gRPC 服务器 (grpc.go)

```go
package server

import (
    "github.com/go-kratos/kratos/v2/log"
    "github.com/go-kratos/kratos/v2/middleware/recovery"
    "github.com/go-kratos/kratos/v2/transport/grpc"
    
    v1 "{service}/api/{service}/v1"
    "{service}/internal/conf"
    "{service}/internal/service"
)

// NewGRPCServer 创建 gRPC 服务器
func NewGRPCServer(c *conf.Server, svc *service.UserService, logger log.Logger) *grpc.Server {
    var opts = []grpc.ServerOption{
        grpc.Middleware(
            recovery.Recovery(),
        ),
    }
    
    if c.Grpc.Network != "" {
        opts = append(opts, grpc.Network(c.Grpc.Network))
    }
    if c.Grpc.Addr != "" {
        opts = append(opts, grpc.Address(c.Grpc.Addr))
    }
    if c.Grpc.Timeout != nil {
        opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
    }
    
    srv := grpc.NewServer(opts...)
    v1.RegisterUserServer(srv, svc)
    
    return srv
}
```

## 4. Wire 注入 (wire.go)

```go
//go:build wireinject
// +build wireinject

package main

import (
    "github.com/go-kratos/kratos/v2"
    "github.com/go-kratos/kratos/v2/log"
    "github.com/google/wire"
    "{service}/internal/biz"
    "{service}/internal/conf"
    "{service}/internal/data"
    "{service}/internal/server"
    "{service}/internal/service"
)

// wireApp 初始化应用
func wireApp(*conf.Server, *conf.Data, *conf.Registry, log.Logger) (*kratos.App, func(), error) {
    panic(wire.Build(
        server.ProviderSet,
        data.ProviderSet,
        biz.ProviderSet,
        service.ProviderSet,
        newApp,
    ))
}
```

## 5. 启动入口 (main.go)

```go
package main

import (
    "flag"
    "os"
    
    "github.com/go-kratos/kratos/v2"
    "github.com/go-kratos/kratos/v2/config"
    "github.com/go-kratos/kratos/v2/config/file"
    "github.com/go-kratos/kratos/v2/log"
    "github.com/go-kratos/kratos/v2/registry"
    "github.com/go-kratos/kratos/v2/transport/grpc"
    "{service}/internal/conf"
)

var (
    Name    = "{service}.service"
    Version = "v1.0.0"
    flagconf string
)

func init() {
    flag.StringVar(&flagconf, "conf", "../../configs", "config path")
}

// newApp 创建 Kratos 应用
func newApp(logger log.Logger, gs *grpc.Server, rr registry.Registrar) *kratos.App {
    return kratos.New(
        kratos.ID(os.Hostname()),
        kratos.Name(Name),
        kratos.Version(Version),
        kratos.Metadata(map[string]string{}),
        kratos.Logger(logger),
        kratos.Server(gs),
        kratos.Registrar(rr),
    )
}

func main() {
    flag.Parse()
    
    // 初始化日志
    logger := log.With(log.NewStdLogger(os.Stdout),
        "service.name", Name,
        "service.version", Version,
    )
    
    // 加载配置
    c := config.New(config.WithSource(file.NewSource(flagconf)))
    defer c.Close()
    
    if err := c.Load(); err != nil {
        panic(err)
    }
    
    var bc conf.Bootstrap
    if err := c.Scan(&bc); err != nil {
        panic(err)
    }
    
    // Wire 注入创建应用
    app, cleanup, err := wireApp(bc.Server, bc.Data, bc.Registry, logger)
    if err != nil {
        panic(err)
    }
    defer cleanup()
    
    // 启动服务
    if err := app.Run(); err != nil {
        panic(err)
    }
}
```

## 6. 启动命令

```bash
# 生成 Wire 代码
cd cmd/{name} && wire

# 或从根目录
make wire

# 启动服务
cd cmd/{name} && go run .

# 或
kratos run

# 指定配置文件
go run . -conf ../../configs
```

## 下一步

- [测试规范](08-testing.md) → 单元测试与集成测试
