# 订单服务 (Order Service)

## 服务概述

订单服务负责处理用户充值订单和产品购买订单的全流程管理，包括订单创建、支付状态跟踪、订单退款等。

## 核心功能模块

| 模块 | 功能说明 |
|------|----------|
| **充值订单** | 充值订单创建、支付回调、订单完成/失败 |
| **产品订单** | 产品购买、机器激活、到期处理 |
| **订单状态** | 待支付/支付成功/支付失败/已退款 |
| **退款管理** | 退款申请、退款审核、退款处理 |
| **订单统计** | 日/周/月订单统计、销售额分析 |
| **对账管理** | 支付通道对账、差异处理 |

## 订单状态流转

```
待支付 → 支付成功 → 订单完成
   ↓         ↓
支付失败   退款处理
```

## 目录结构

```
order/
├── api/                 # API接口定义
│   └── order/v1/        # 订单服务v1版本接口
├── cmd/                 # 服务入口
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
| CreateOrder | POST | 创建订单 |
| GetOrder | GET | 订单详情 |
| ListOrders | GET | 订单列表查询 |
| PayCallback | POST | 支付回调处理 |
| QueryOrderStatus | GET | 查询订单状态 |
| ApplyRefund | POST | 申请退款 |
| ProcessRefund | POST | 处理退款 |
| GetOrderStats | GET | 订单统计 |

## 数据模型

| 实体 | 说明 |
|------|------|
| Order | 订单主表 |
| RechargeOrder | 充值订单 |
| ProductOrder | 产品订单 |
| RefundRecord | 退款记录 |
| OrderLog | 订单日志 |

## 依赖服务

| 服务 | 用途 |
|------|------|
| payment | 支付处理 |
| user | 用户信息/资产更新 |
| product | 产品/库存查询 |
| finance | 资金流水记录 |

## 快速开始

```bash
make api
cd cmd/server && go run .
```

## Docker部署

```bash
docker build -t order-service:latest .
docker run -d -p 8003:8000 -p 9003:9000 \
  -v ./configs:/data/conf \
  --name order-service \
  order-service:latest
```
