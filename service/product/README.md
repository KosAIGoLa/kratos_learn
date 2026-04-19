# 产品服务 (Product Service)

## 服务概述

产品服务负责CPU产品的全生命周期管理，包括产品配置、库存管理、上下架控制及销售统计。

## 核心功能模块

| 模块 | 功能说明 |
|------|----------|
| **产品管理** | 产品CRUD、配置维护、上下架控制 |
| **库存管理** | 库存增减、库存预警、库存日志 |
| **算力配置** | 产品级别、生产力%、有效期配置 |
| **价格管理** | 产品价格、促销价格 |
| **销售统计** | 销量统计、销售额统计、产品排行 |
| **产品分类** | 产品分类、标签管理 |

## 产品算力公式

```
每小时算力产出 = 产品级别 × 生产力%
每日算力产出 = 每小时产出 × 24小时
```

## 目录结构

```
product/
├── api/                 # API接口定义
│   └── product/v1/      # 产品服务v1版本接口
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
| ListProducts | GET | 产品列表查询 |
| GetProduct | GET | 产品详情 |
| CreateProduct | POST | 创建产品 |
| UpdateProduct | PUT | 更新产品 |
| DeleteProduct | DELETE | 删除产品 |
| UpdateStock | PUT | 更新库存 |
| GetStockLog | GET | 库存变动日志 |
| ToggleProductStatus | PUT | 上下架切换 |
| GetSalesStats | GET | 销售统计 |

## 数据模型

| 实体 | 说明 |
|------|------|
| Product | 产品信息 |
| ProductCategory | 产品分类 |
| Stock | 库存 |
| StockLog | 库存日志 |
| SalesStats | 销售统计 |

## 依赖服务

| 服务 | 用途 |
|------|------|
| order | 订单关联查询 |
| user | 购买用户统计 |

## 快速开始

```bash
make api
cd cmd/server && go run .
```

## Docker部署

```bash
docker build -t product-service:latest .
docker run -d -p 8002:8000 -p 9002:9000 \
  -v ./configs:/data/conf \
  --name product-service \
  product-service:latest
```
