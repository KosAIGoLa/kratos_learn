# 内容服务 (Content Service)

## 服务概述

内容服务负责平台的内容管理，包括帮助文档、用户协议、常见问题、推广素材等内容的维护和展示。

## 核心功能模块

| 模块 | 功能说明 |
|------|----------|
| **帮助中心** | 帮助文档分类、文章内容管理 |
| **用户协议** | 服务协议、隐私政策、风险声明 |
| **常见问题** | FAQ分类、问题解答管理 |
| **推广素材** | 推广文案、海报图片、分享链接 |
| **Banner管理** | 轮播图配置、跳转链接 |
| **消息模板** | 短信模板、邮件模板、推送模板 |

## 内容类型

| 类型 | 说明 |
|------|------|
| `help` | 帮助文档 |
| `agreement` | 用户协议 |
| `faq` | 常见问题 |
| `promotion` | 推广素材 |
| `banner` | 轮播图 |
| `template` | 消息模板 |

## 目录结构

```
content/
├── api/                 # API接口定义
│   └── content/v1/      # 内容服务v1版本接口
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
| ListCategories | GET | 内容分类列表 |
| GetCategory | GET | 分类详情 |
| CreateCategory | POST | 创建分类 |
| ListContents | GET | 内容列表 |
| GetContent | GET | 内容详情 |
| CreateContent | POST | 创建内容 |
| UpdateContent | PUT | 更新内容 |
| DeleteContent | DELETE | 删除内容 |
| ListBanners | GET | Banner列表 |
| CreateBanner | POST | 添加Banner |
| UpdateBanner | PUT | 更新Banner |
| DeleteBanner | DELETE | 删除Banner |
| GetTemplate | GET | 获取消息模板 |
| UpdateTemplate | PUT | 更新消息模板 |
| GetPromotionMaterial | GET | 获取推广素材 |

## 数据模型

| 实体 | 说明 |
|------|------|
| Category | 内容分类 |
| Content | 内容主体 |
| Banner | 轮播图 |
| Template | 消息模板 |
| Promotion | 推广素材 |

## 依赖服务

无直接依赖。

## 快速开始

```bash
make api
cd cmd/server && go run .
```

## Docker部署

```bash
docker build -t content-service:latest .
docker run -d -p 8007:8000 -p 9007:9000 \
  -v ./configs:/data/conf \
  --name content-service \
  content-service:latest
```

## 内容发布流程

```
草稿 → 审核中 → 已发布 → 已下架
  ↓         ↓
保存    审核拒绝
```
