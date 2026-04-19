# 财务服务 (Finance Service)

## 服务概述

财务服务负责平台所有资金相关业务，包括资金流水记录、用户资产管理、提现处理、工分管理、对账等。

## 核心功能模块

| 模块 | 功能说明 |
|------|----------|
| **资金管理** | 充值流水、提现流水、资金变动记录 |
| **资产管理** | 用户余额查询、资产调整、资产冻结 |
| **工分管理** | 工分发放、工分扣除、工分流水 |
| **提现处理** | 提现申请、审核流程、打款处理 |
| **对账管理** | 日终对账、差异处理、财务报表 |
| **算力转换** | 算力转余额、转换记录 |

## 资金类型

| 类型 | 收支 | 说明 |
|------|:----:|------|
| 用户充值 | 收入 | 用户充值金额 |
| 用户提现 | 支出 | 用户提现金额 |
| 算力转余额 | 收入 | 算力1:1转换 |
| 邀请奖励 | 收入 | 邀请返佣 |
| 工分抵扣 | - | 提现时抵扣 |

## 目录结构

```
finance/
├── api/                 # API接口定义
│   └── finance/v1/      # 财务服务v1版本接口
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
| GetUserAssets | GET | 查询用户资产 |
| ListTransactions | GET | 资金流水列表 |
| CreateTransaction | POST | 创建资金流水 |
| ApplyWithdrawal | POST | 申请提现 |
| ApproveWithdrawal | POST | 审核提现 |
| ProcessWithdrawal | POST | 处理打款 |
| ListWorkPoints | GET | 工分流水查询 |
| AddWorkPoints | POST | 增加工分 |
| DeductWorkPoints | POST | 扣除工分 |
| ConvertHashPower | POST | 算力转余额 |
| GetFinancialReport | GET | 财务报表 |

## 数据模型

| 实体 | 说明 |
|------|------|
| Transaction | 资金流水 |
| Withdrawal | 提现记录 |
| WorkPointLog | 工分流水 |
| HashPowerConversion | 算力转换 |
| UserAsset | 用户资产 |

## 依赖服务

| 服务 | 用途 |
|------|------|
| user | 用户信息/资产更新 |
| admin | 提现审批 |
| payment | 打款处理 |

## 快速开始

```bash
make api
cd cmd/server && go run .
```

## Docker部署

```bash
docker build -t finance-service:latest .
docker run -d -p 8005:8000 -p 9005:9000 \
  -v ./configs:/data/conf \
  --name finance-service \
  finance-service:latest
```

## 提现审核流程

```
申请 → 系统检查 → 待审核 → 人工审核 → 打款 → 完成
         ↓           ↓         ↓
       条件不符    审核拒绝   打款失败
```
