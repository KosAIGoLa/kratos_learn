# 挖礦/算力平台微服務

基於 [Kratos](https://go-kratos.dev/) 框架的 Go 微服務架構，提供完整的挖礦/算力平台後端服務。

---

## 技術棧

| 組件 | 技術 |
|------|------|
| 框架 | [Kratos](https://go-kratos.dev/) v2.7.2 |
| 通信 | gRPC + HTTP/REST |
| 數據庫 | MySQL + GORM |
| 文檔數據庫 | MongoDB |
| 緩存 | Redis |
| 消息隊列 | RabbitMQ |
| 配置 | Protobuf + YAML |
| 依賴注入 | Google Wire |
| 服務發現 | etcd |
| API 文檔 | Swagger/OpenAPI v2 |
| 鏈路追蹤 | OpenTelemetry + Jaeger |

---

## 服務清單

| 服務 | HTTP 端口 | gRPC 端口 | 說明 |
|------|-----------|-----------|------|
| content | 8101 | 9101 | 內容管理（新聞、輪播圖） |
| finance | 8102 | 9102 | 財務管理（充值、提現、收益） |
| order | 8103 | 9103 | 訂單管理 |
| payment | 8104 | 9104 | 支付處理 |
| product | 8105 | 9105 | 產品管理（礦機、算力產品） |
| system | 8106 | 9106 | 系統配置（分潤規則、風控、域名） |
| user | 8107 | 9107 | 用戶管理（註冊、登錄、KYC、團隊） |
| admin | 8108 | 9108 | 管理員後台（管理員、角色、菜單、日誌） |
| mail | 8109 | 9109 | 郵件服務 |
| sms | 8110 | 9110 | 短信服務 |
| cron | 8111 | 9111 | 定時任務服務 |
| report | 8007 | 9007 | 報表服務（訂單、用戶、銷售、商品報表 + Excel 導出） |

---

## 項目結構

```
my-front-app/
├── service/                    # 微服務實現
│   ├── content/               # 內容服務
│   │   ├── api/content/v1/    # Protobuf API 定義
│   │   ├── cmd/server/        # 服務入口
│   │   ├── configs/           # 配置文件
│   │   └── internal/
│   │       ├── biz/           # 業務邏輯層
│   │       ├── data/          # 數據訪問層
│   │       ├── server/        # HTTP/gRPC 服務器
│   │       └── service/       # 服務實現層
│   ├── finance/               # 財務服務
│   ├── order/                 # 訂單服務
│   ├── payment/               # 支付服務
│   ├── product/               # 產品服務
│   ├── system/                # 系統服務
│   ├── user/                  # 用戶服務
│   ├── admin/                 # 管理員服務（含菜單、角色、日誌管理）
│   ├── mail/                  # 郵件服務
│   ├── sms/                   # 短信服務
│   ├── cron/                  # 定時任務服務
│   └── report/                # 報表服務
├── third_party/               # 第三方 proto 文件
├── schema.sql                 # 數據庫結構
├── all_post.http              # HTTP 測試請求集合
├── test_grpc.sh              # gRPC 測試腳本
└── go.mod                     # Go 依賴管理
```

---

## 快速開始

### 環境要求

- Go 1.21+
- MySQL 8.0+
- MongoDB 5.0+
- Redis 6.0+
- RabbitMQ 3.8+
- Protobuf 編譯器

### 1. 安裝依賴

```bash
go mod tidy
go mod download
```

### 2. 初始化數據庫

```bash
mysql -u root -p < schema.sql
```

### 3. 生成 API 代碼

```bash
# 生成指定服務的 API
cd service/content && kratos proto api api/content/v1/content.proto

# 或生成所有服務
make api
```

### 4. 生成 Wire 依賴注入

```bash
cd service/content/cmd/server && wire
cd service/finance/cmd/server && wire
# ... 其他服務
```

### 5. 運行服務

```bash
# 進入服務目錄並運行
cd service/content && go run cmd/server/main.go
cd service/user && go run cmd/server/main.go
# ... 其他服務
```

---

## 數據庫結構

| 服務 | 主要數據表 |
|------|-----------|
| user | `users`, `kyc_verifications`, `team_relations`, `login_logs` |
| product | `products`, `mining_machines` |
| order | `orders` |
| finance | `recharges`, `withdrawals`, `income_logs`, `balance_logs`, `check_ins`, `user_hashrates` |
| content | `news`, `banners` |
| admin | `admins`, `menus`, `roles`, `role_menus`, `admin_roles`, `admin_logs` |
| payment | `payment_channels`, `payment_api_logs` |
| system | `system_configs`, `profit_sharing_rules`, `risk_controls`, `domains`, `whitelist_ips` |
| report | 複用 `orders`, `users`, `products` 等表進行報表統計 |

---

## API 測試

項目包含完整的 HTTP 測試文件 `all_post.http`，可在 IDE 中直接運行：

```
all_post.http
├── Content Service (8101)    # 橫幅、新聞管理
├── User Service (8107)        # 註冊、登錄、KYC、團隊
├── Product Service (8105)     # 產品、任務管理
├── Order Service (8103)       # 訂單創建、支付
├── Finance Service (8102)     # 充值、提現、收益
├── Payment Service (8104)     # 支付渠道、回調
├── System Service (8106)      # 配置、分潤、風控
├── Admin Service (8108)       # 管理員、角色、菜單、日誌
├── Mail Service (8109)        # 郵件發送服務
├── SMS Service (8110)         # 短信發送服務
├── Report Service (8007)      # 報表服務
└── Swagger UI 訪問地址        # http://localhost:{port}/q/
```

各服務 Swagger UI 地址：
- Content: http://localhost:8101/q/
- Finance: http://localhost:8102/q/
- Order: http://localhost:8103/q/
- Payment: http://localhost:8104/q/
- Product: http://localhost:8105/q/
- System: http://localhost:8106/q/
- User: http://localhost:8107/q/
- Admin: http://localhost:8108/q/
- Mail: http://localhost:8109/q/
- SMS: http://localhost:8110/q/
- Report: http://localhost:8007/q/

### gRPC 測試

項目提供 `test_grpc.sh` 腳本用於測試 gRPC 服務：

**安裝 grpcurl：**
```bash
brew install grpcurl
```

**使用方法：**
```bash
# 測試所有服務健康狀態
./test_grpc.sh all

# 列出所有服務與方法
./test_grpc.sh list

# 測試指定服務（帶示例請求）
./test_grpc.sh user
./test_grpc.sh content
./test_grpc.sh admin

# 調用指定方法
./test_grpc.sh user Login '{"phone": "13800138000", "password": "123456"}'
./test_grpc.sh admin Login '{"username": "admin", "password": "admin123"}'
./test_grpc.sh content ListBanners '{"type": "home", "status": 1}'
./test_grpc.sh product ListProducts '{"status": 1, "page": 1, "page_size": 10}'
```

**gRPC 端口：**
- Content: `localhost:9101`
- Finance: `localhost:9102`
- Order: `localhost:9103`
- Payment: `localhost:9104`
- Product: `localhost:9105`
- System: `localhost:9106`
- User: `localhost:9107`
- Admin: `localhost:9108`
- Mail: `localhost:9109`
- SMS: `localhost:9110`
- Cron: `localhost:9111`
- Report: `localhost:9007`

---

## 開發規範

參見 `.windsurf/rules/` 目錄下的開發規範：
- `01-overview.md` - 項目概覽
- `02-api-definition.md` - API 定義規範
- `03-configuration.md` - 配置管理規範
- `04-data-layer.md` - 數據層規範
- `05-biz-layer.md` - 業務層規範
- `06-service-layer.md` - 服務層規範
- `07-server-setup.md` - 服務器設置規範
- `08-testing.md` - 測試規範

---

## 核心功能模塊

### 用戶系統
- 用戶註冊/登錄（JWT Token + Refresh Token）
- KYC 身份認證
- 邀請碼與團隊關係

### 產品系統
- 礦機管理
- 算力產品配置
- 任務系統

### 財務系統
- 充值/提現（提現基於 RabbitMQ 異步處理）
- 收益計算與發放
- 餘額變動記錄
- 每日簽到

#### 提現流程（異步處理）
```
用戶發起提現 → API 接收請求 → 推送 RabbitMQ → 返回「處理中」
                                              ↓
                                       消費者異步處理
                                              ↓
                                       創建提現記錄（MySQL）
```
- **隊列**: `withdrawal.queue`
- **交換器**: `finance.exchange`
- **路由鍵**: `withdrawal`

### 服務發現

基於 **etcd** 的服務註冊與發現，支持多服務實例動態管理。

**配置示例：**
```yaml
registry:
  etcd:
    endpoints:
      - 127.0.0.1:2379
    dial_timeout: 5s
```

**查看已註冊服務：**
```bash
etcdctl get --prefix /microservices/
```

詳見 [ETCD_SERVICE_DISCOVERY.md](documents/ETCD_SERVICE_DISCOVERY.mdOVERY.md)

### 訂單系統
- 產品購買
- 訂單狀態管理

### 支付系統
- 多支付渠道支持
- 支付回調處理

### 內容系統
- 新聞公告
- 輪播圖管理

### 系統配置
- 分潤規則配置
- 風控規則
- 域名白名單

### 報表系統
- 訂單報表（明細查詢、統計、Excel 導出）
- 用戶報表（消費統計、活躍度分析）
- 銷售報表（日/周/月趨勢分析）
- 商品報表（銷售排行、庫存分析）

### 管理員系統
- 管理員登錄（JWT Token）
- 管理員管理（創建、編輯、刪除）
- 角色管理（角色創建、權限分配）
- 菜單管理（菜單結構、權限標識）
- 操作日誌查詢

### 通知服務
- 郵件服務（SMTP 配置、郵件模板）
- 短信服務（多提供商支持：Twilio、阿里雲等）

### 鏈路追蹤 (OpenTelemetry)
基於 **OpenTelemetry** 的分布式鏈路追蹤，支持跨服務請求追蹤。

**配置示例：**
```yaml
trace:
  endpoint: http://127.0.0.1:14268/api/traces
  sample_rate: 1.0
```

**查看鏈路：**
- 訪問 Jaeger UI: http://localhost:16686
