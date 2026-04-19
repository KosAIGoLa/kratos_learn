# 管理员服务 (Admin Service)

## 服务概述

管理员服务负责后台管理系统的权限控制、角色管理、管理员账号管理及操作日志记录，保障系统的安全性和可追溯性。

## 核心功能模块

| 模块 | 功能说明 |
|------|----------|
| **管理员管理** | 管理员CRUD、账号启用/禁用、密码重置 |
| **角色管理** | 角色创建、权限分配、角色层级 |
| **权限控制** | 菜单权限、操作权限、数据权限 |
| **登录认证** | 后台登录、Token管理、登录日志 |
| **操作日志** | 操作记录、日志查询、审计追踪 |
| **审批管理** | 提现审批、KYC审批、任务审批 |

## 目录结构

```
admin/
├── api/                 # API接口定义 (Protobuf)
│   └── admin/v1/        # 管理员服务v1版本接口
├── cmd/                 # 服务入口
│   └── server/          # HTTP/gRPC服务启动
├── internal/            # 内部实现
│   ├── biz/             # 业务逻辑层
│   ├── data/            # 数据访问层
│   ├── service/         # 服务实现层
│   └── server/          # 服务器配置
├── configs/             # 配置文件
└── third_party/         # 第三方依赖
```

## 主要API接口

| 接口 | 方法 | 说明 |
|------|------|------|
| AdminLogin | POST | 管理员登录 |
| GetAdminInfo | GET | 获取管理员信息 |
| ListAdmins | GET | 管理员列表查询 |
| CreateAdmin | POST | 创建管理员 |
| UpdateAdmin | PUT | 更新管理员 |
| DeleteAdmin | DELETE | 删除管理员 |
| ListRoles | GET | 角色列表 |
| CreateRole | POST | 创建角色 |
| UpdateRole | PUT | 更新角色权限 |
| ListPermissions | GET | 权限列表 |
| ListOperationLogs | GET | 操作日志查询 |
| ApproveWithdrawal | POST | 审批提现申请 |
| ApproveKYC | POST | 审批KYC认证 |

## 数据模型

| 实体 | 说明 |
|------|------|
| Admin | 管理员账号 |
| Role | 角色 |
| Permission | 权限 |
| AdminRole | 管理员角色关联 |
| OperationLog | 操作日志 |
| ApprovalRecord | 审批记录 |

## 权限等级

| 等级 | 角色 | 权限范围 |
|------|------|----------|
| Level 1 | 超级管理员 | 全部权限 |
| Level 2 | 运营管理员 | 用户/订单/提现审核 |
| Level 3 | 财务管理员 | 财务相关/报表导出 |
| Level 4 | 客服管理员 | 查看/处理投诉 |
| Level 5 | 只读管理员 | 仅查看 |

## 依赖服务

| 服务 | 用途 |
|------|------|
| user | 查询用户信息 |
| finance | 资金流水确认 |
| system | 系统配置 |

## 快速开始

```bash
# 生成API代码
make api

# 运行服务
cd cmd/server && go run .
```

## Docker部署

```bash
docker build -t admin-service:latest .
docker run -d -p 8001:8000 -p 9001:9000 \
  -v ./configs:/data/conf \
  --name admin-service \
  admin-service:latest
```

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| ADMIN_HTTP_PORT | HTTP端口 | 8000 |
| ADMIN_GRPC_PORT | gRPC端口 | 9000 |
| JWT_SECRET | JWT密钥 | - |
| TOKEN_EXPIRE | Token过期时间(小时) | 24 |
