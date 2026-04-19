# 用户服务 (User Service)

## 服务概述

用户服务负责平台用户全生命周期管理，包括用户注册、登录、资料管理、资产查询、邀请关系维护等核心功能。

## 核心功能模块

| 模块 | 功能说明 |
|------|----------|
| **用户注册/登录** | 手机号注册、短信验证码、密码登录、Token管理 |
| **用户资料管理** | 用户信息修改、头像上传、密码修改 |
| **资产管理** | 充值余额、可提现余额、工分数量查询 |
| **邀请关系** | 邀请码生成、邀请层级关系、邀请统计 |
| **KYC认证** | 实名认证提交、身份证信息验证、审核状态查询 |
| **机器管理** | 用户机器列表、算力产出查询、机器状态 |

## 目录结构

```
user/
├── api/                 # API接口定义 (Protobuf)
│   └── user/v1/         # 用户服务v1版本接口
├── cmd/                 # 服务入口
│   └── server/          # HTTP/gRPC服务启动
├── internal/            # 内部实现
│   ├── biz/             # 业务逻辑层 (Usecase)
│   ├── data/            # 数据访问层 (Repository)
│   ├── service/         # 服务实现层 (Service)
│   └── server/          # 服务器配置 (HTTP/gRPC)
├── configs/             # 配置文件
└── third_party/         # 第三方依赖
```

## 主要API接口

| 接口 | 方法 | 说明 |
|------|------|------|
| Register | POST | 用户注册 |
| Login | POST | 用户登录 |
| GetUserInfo | GET | 获取用户信息 |
| UpdateUser | PUT | 更新用户信息 |
| GetUserAssets | GET | 查询用户资产 |
| GetInviteInfo | GET | 查询邀请信息 |
| SubmitKYC | POST | 提交KYC认证 |
| GetKYCStatus | GET | 查询KYC状态 |
| GetUserMachines | GET | 获取用户机器列表 |

## 数据模型

| 实体 | 说明 |
|------|------|
| User | 用户基础信息 |
| UserAsset | 用户资产 (余额、工分) |
| InviteRelation | 邀请关系 |
| KYCInfo | KYC认证信息 |
| UserMachine | 用户机器关联 |

## 依赖服务

| 服务 | 用途 |
|------|------|
| system | 获取系统配置 |
| finance | 查询资金流水 |

## 快速开始

```bash
# 生成API代码
make api

# 运行服务
cd cmd/server && go run .

# 或使用配置文件
./bin/server -conf ../../configs
```

## Docker部署

```bash
# 构建镜像
docker build -t user-service:latest .

# 运行容器
docker run -d -p 8000:8000 -p 9000:9000 \
  -v ./configs:/data/conf \
  --name user-service \
  user-service:latest
```

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| USER_HTTP_PORT | HTTP服务端口 | 8000 |
| USER_GRPC_PORT | gRPC服务端口 | 9000 |
| USER_DATABASE_DSN | 数据库连接串 | - |
| REDIS_ADDR | Redis地址 | localhost:6379 |
