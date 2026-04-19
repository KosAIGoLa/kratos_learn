# 支付服务 (Payment Service)

## 服务概述

支付服务负责处理所有支付相关业务，包括多渠道支付接入、支付请求处理、支付回调处理、退款处理等。

## 核心功能模块

| 模块 | 功能说明 |
|------|----------|
| **支付通道** | 支付宝、微信、快捷支付、银行卡支付 |
| **支付请求** | 创建支付订单、获取支付参数 |
| **回调处理** | 异步回调接收、签名验证、状态通知 |
| **退款处理** | 原路退款、批量退款、退款查询 |
| **对账管理** | 日终对账、差异处理、对账报表 |
| **风控管理** | 支付风控、异常检测、限额控制 |

## 支持的支付方式

| 通道 | 支持功能 |
|------|----------|
| 支付宝 | 扫码支付、H5支付、APP支付 |
| 微信支付 | 扫码支付、JSAPI支付、APP支付 |
| 快捷支付 | 银行卡快捷支付 |
| 银行卡 | 网银支付 |

## 目录结构

```
payment/
├── api/                 # API接口定义
│   └── payment/v1/      # 支付服务v1版本接口
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
| CreatePayment | POST | 创建支付 |
| GetPaymentParams | GET | 获取支付参数 |
| ProcessCallback | POST | 处理支付回调 |
| QueryPayment | GET | 查询支付状态 |
| CreateRefund | POST | 创建退款 |
| QueryRefund | GET | 查询退款状态 |
| GetReconciliation | GET | 获取对账数据 |

## 数据模型

| 实体 | 说明 |
|------|------|
| Payment | 支付记录 |
| Refund | 退款记录 |
| Channel | 支付通道配置 |
| Reconciliation | 对账记录 |

## 依赖服务

| 服务 | 用途 |
|------|------|
| order | 订单状态同步 |
| finance | 资金流水记录 |

## 快速开始

```bash
make api
cd cmd/server && go run .
```

## Docker部署

```bash
docker build -t payment-service:latest .
docker run -d -p 8004:8000 -p 9004:9000 \
  -v ./configs:/data/conf \
  --name payment-service \
  payment-service:latest
```

## 配置说明

支付通道配置需在 `configs/` 中配置各渠道的密钥信息：
- 支付宝: app_id, private_key, public_key
- 微信: mch_id, app_id, api_key
