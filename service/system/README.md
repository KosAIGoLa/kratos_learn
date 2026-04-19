# 系统服务 (System Service)

## 服务概述

系统服务负责平台的基础配置管理、系统参数设置、公告管理、数据字典维护等核心配置功能。

## 核心功能模块

| 模块 | 功能说明 |
|------|----------|
| **系统配置** | 全局参数、业务规则配置 |
| **提现配置** | 最低提现额、手续费、次数限制 |
| **任务配置** | 签到奖励、邀请奖励、首充奖励 |
| **公告管理** | 系统公告、弹窗公告、通知消息 |
| **数据字典** | 常量配置、枚举值管理 |
| **地区配置** | 省市区数据、银行列表 |
| **敏感词管理** | 敏感词库、内容审核 |

## 配置分类

| 配置项 | 说明 |
|--------|------|
| `withdrawal.min_amount` | 最低提现金额 |
| `withdrawal.max_daily_count` | 每日最大提现次数 |
| `withdrawal.fee_rate` | 提现手续费率 |
| `task.checkin_points` | 签到奖励工分 |
| `task.invite_points` | 邀请奖励工分 |
| `task.first_recharge_points` | 首充奖励工分 |

## 目录结构

```
system/
├── api/                 # API接口定义
│   └── system/v1/       # 系统服务v1版本接口
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
| GetConfig | GET | 获取配置项 |
| UpdateConfig | PUT | 更新配置 |
| ListConfigs | GET | 配置列表 |
| GetWithdrawalConfig | GET | 提现配置 |
| UpdateWithdrawalConfig | PUT | 更新提现配置 |
| GetTaskConfig | GET | 任务配置 |
| UpdateTaskConfig | PUT | 更新任务配置 |
| ListNotices | GET | 公告列表 |
| CreateNotice | POST | 发布公告 |
| UpdateNotice | PUT | 更新公告 |
| DeleteNotice | DELETE | 删除公告 |
| GetDataDict | GET | 数据字典查询 |
| RefreshCache | POST | 刷新配置缓存 |

## 数据模型

| 实体 | 说明 |
|------|------|
| Config | 系统配置 |
| Notice | 系统公告 |
| DataDict | 数据字典 |
| SensitiveWord | 敏感词 |

## 依赖服务

无直接依赖，被其他服务依赖调用。

## 快速开始

```bash
make api
cd cmd/server && go run .
```

## Docker部署

```bash
docker build -t system-service:latest .
docker run -d -p 8006:8000 -p 9006:9000 \
  -v ./configs:/data/conf \
  --name system-service \
  system-service:latest
```

## 配置热更新

系统服务支持配置热更新，修改配置后调用 `RefreshCache` 接口即可生效，无需重启服务。
