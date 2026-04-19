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
| 配置 | Protobuf + YAML |
| 依賴注入 | Google Wire |
| API 文檔 | Swagger/OpenAPI v2 |

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
| admin | 8000 | 9000 | 管理員後台（管理員、角色、菜單、日誌） |

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
│   └── admin/                 # 管理員服務（含菜單、角色、日誌管理）
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
├── Admin Service (8000)       # 管理員、角色、菜單、日誌
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
- Admin: http://localhost:8000/q/

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

---

## 開發規範

參見 `.windsurf/rules/` 目錄下的開發規範：
- `01-overview.md` - 項目概覽
- `02-project-structure.md` - 項目結構規範
- `03-coding-style.md` - 編碼風格
- `04-database.md` - 數據庫規範
- `05-api-design.md` - API 設計規範
- `06-kratos-pattern.md` - Kratos 模式
- `07-error-handling.md` - 錯誤處理
- `08-testing.md` - 測試規範
- `09-security.md` - 安全規範
- `10-deployment.md` - 部署規範
- `11-common-commands.md` - 常用命令
- `12-troubleshooting.md` - 問題排查
- `13-kratos-layout.md` - Kratos 佈局規範

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
- 充值/提現
- 收益計算與發放
- 餘額變動記錄
- 每日簽到

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

### 管理員系統
- 管理員登錄（JWT Token）
- 管理員管理（創建、編輯、刪除）
- 角色管理（角色創建、權限分配）
- 菜單管理（菜單結構、權限標識）
- 操作日誌查詢
