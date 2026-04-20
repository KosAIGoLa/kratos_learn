---
trigger: manual
description: Kratos 微服务开发规范概览
globs: ["**/service/**"]
---

# Kratos 微服务开发规范

> 基于 [Kratos](https://go-kratos.dev/) 框架的微服务项目开发规范

## 目录结构

```
service/{service-name}/
├── api/                    # API 定义 (protobuf)
│   └── {name}/v1/
│       └── {name}.proto
├── cmd/{name}/             # 启动入口
│   ├── main.go
│   ├── wire.go
│   └── wire_gen.go
├── configs/                # 配置文件
│   └── config.yaml
├── internal/
│   ├── biz/               # 业务逻辑层 (Business)
│   ├── data/              # 数据访问层 (Data)
│   ├── service/           # 服务层 (Service)
│   ├── server/            # 服务器配置 (Server)
│   └── conf/              # 配置结构定义
└── go.mod
```

## 开发流程

| 顺序 | 阶段 | 文件 | 说明 |
|------|------|------|------|
| 1 | [项目初始化](01-project-init.md) | - | 创建服务、定义 API |
| 2 | [API 定义](02-api-definition.md) | `api/{name}/v1/*.proto` | protobuf 接口定义 |
| 3 | [配置管理](03-configuration.md) | `configs/`, `internal/conf/` | 配置结构与加载 |
| 4 | [数据层](04-data-layer.md) | `internal/data/` | GORM 模型与仓储 |
| 5 | [业务层](05-biz-layer.md) | `internal/biz/` | 业务逻辑与用例 |
| 6 | [服务层](06-service-layer.md) | `internal/service/` | gRPC/HTTP 服务实现 |
| 7 | [服务器配置](07-server-setup.md) | `internal/server/`, `cmd/` | 服务器启动与注册 |
| 8 | [测试](08-testing.md) | `test/`, `*_test.go` | 单元测试与集成测试 |

## 核心依赖

```go
// go.mod
go 1.21

require (
    github.com/go-kratos/kratos/v2 v2.7.0
    github.com/google/wire v0.5.0
    gorm.io/gorm v1.25.0
    gorm.io/driver/mysql v1.5.0
    github.com/redis/go-redis/v9 v9.0.0
)
```

## 常用命令

```bash
# 创建新服务
kratos new {service-name}

# 添加 proto
kratos proto add api/{name}/v1/{name}.proto

# 生成 service 骨架
kratos proto server api/{name}/v1/{name}.proto -t internal/service

# 生成 proto 代码
make api

# 生成配置代码
make config

# 生成 wire 依赖注入
make wire

# 启动服务
kratos run
```

## 架构分层

```
┌─────────────────────────────────────┐
│           Service Layer             │  ← gRPC/HTTP 接口实现
│     (internal/service/*.go)         │
├─────────────────────────────────────┤
│           Biz Layer                 │  ← 业务逻辑、用例
│       (internal/biz/*.go)           │
├─────────────────────────────────────┤
│           Data Layer                │  ← 数据访问、仓储实现
│       (internal/data/*.go)          │
├─────────────────────────────────────┤
│           Config Layer              │  ← 配置定义与加载
│       (internal/conf/*.go)          │
└─────────────────────────────────────┘
```

## 参考资料

- [Kratos 官方文档](https://go-kratos.dev/docs/)
- [Wire 依赖注入](https://go-kratos.dev/docs/guide/wire)
- [项目布局](https://go-kratos.dev/docs/getting-started/layout)
- [GORM 文档](https://gorm.io/)
