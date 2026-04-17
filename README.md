# 挖礦/算力平台微服務

基於 Kratos 框架的微服務架構實現，依照 schema.sql 數據庫結構設計。

## 服務架構

```
my-front-app/
├── api/                          # API 定義 (Protobuf)
│   └── service/
│       ├── user/v1/              # 用戶服務 API
│       ├── product/v1/           # 產品服務 API
│       ├── order/v1/             # 訂單服務 API
│       ├── finance/v1/            # 財務服務 API
│       ├── task/v1/               # 任務服務 API
│       ├── admin/v1/              # 管理員服務 API
│       ├── payment/v1/            # 支付服務 API
│       └── system/v1/              # 系統服務 API
├── service/                      # 微服務實現
│   ├── user/                     # 用戶服務 (端口: 8001/9001)
│   │   ├── cmd/server/           # 服務入口
│   │   └── internal/
│   │       ├── biz/              # 業務邏輯
│   │       ├── data/             # 數據訪問
│   │       ├── service/          # 服務實現
│   │       └── server/           # HTTP/gRPC 服務器
│   ├── product/                  # 產品服務 (端口: 8002/9002)
│   ├── order/                    # 訂單服務 (端口: 8003/9003)
│   ├── finance/                  # 財務服務 (端口: 8004/9004)
│   ├── task/                     # 任務服務 (端口: 8005/9005)
│   ├── admin/                    # 管理員服務 (端口: 8006/9006)
│   ├── payment/                  # 支付服務 (端口: 8007/9007)
│   └── system/                   # 系統服務 (端口: 8008/9008)
├── configs/                      # 配置文件
├── third_party/                # 第三方 proto 文件
└── schema.sql                  # 數據庫結構
```

## 數據庫結構對應

| 服務 | 數據表 |
|------|--------|
| user | users, kyc_verifications, team_relations, login_logs |
| product | products, mining_machines |
| order | orders |
| finance | recharges, withdrawals, income_logs, balance_logs, user_hashrates |
| task | tasks, user_tasks, check_ins |
| admin | admins, menus, roles, role_menus, admin_roles, admin_logs, news, banners |
| payment | payment_channels, payment_api_logs |
| system | system_configs, profit_sharing_rules, risk_controls, domains, whitelist_ips |

## 快速開始

### 1. 初始化環境

```bash
make init
```

### 2. 生成 API 代碼

```bash
make api
```

### 3. 生成 Wire 依賴注入

```bash
make wire
```

### 4. 運行服務

```bash
# 運行用戶服務
make run-user

# 運行產品服務
make run-product

# 運行訂單服務
make run-order

# 運行財務服務
make run-finance

# 運行任務服務
make run-task

# 運行管理員服務
make run-admin

# 運行支付服務
make run-payment

# 運行系統服務
make run-system
```

### 5. 構建所有服務

```bash
make build
```

## API 端點

### 用戶服務 (user)
- POST   /api/v1/user/register       # 註冊
- POST   /api/v1/user/login           # 登錄
- GET    /api/v1/user/{id}            # 獲取用戶
- PUT    /api/v1/user/{id}            # 更新用戶
- GET    /api/v1/users                # 用戶列表
- POST   /api/v1/user/kyc             # 提交KYC
- GET    /api/v1/user/kyc             # 獲取KYC
- GET    /api/v1/user/team            # 團隊關係
- GET    /api/v1/user/subordinates    # 下級列表

### 產品服務 (product)
- GET    /api/v1/products             # 產品列表
- GET    /api/v1/products/{id}        # 產品詳情
- POST   /api/v1/products             # 創建產品
- PUT    /api/v1/products/{id}        # 更新產品
- DELETE /api/v1/products/{id}        # 刪除產品
- GET    /api/v1/mining-machines      # 礦機列表
- GET    /api/v1/mining-machines/{id} # 礦機詳情

### 訂單服務 (order)
- POST   /api/v1/orders               # 創建訂單
- GET    /api/v1/orders/{id}          # 訂單詳情
- GET    /api/v1/orders               # 訂單列表
- PUT    /api/v1/orders/{id}/cancel    # 取消訂單
- PUT    /api/v1/orders/{id}/confirm # 確認訂單
- PUT    /api/v1/orders/{id}/refund   # 退款申請

### 財務服務 (finance)
- POST   /api/v1/recharges            # 充值
- GET    /api/v1/recharges/{id}       # 充值詳情
- GET    /api/v1/recharges            # 充值記錄
- POST   /api/v1/withdrawals          # 申請提現
- GET    /api/v1/withdrawals/{id}     # 提現詳情
- GET    /api/v1/withdrawals          # 提現記錄
- GET    /api/v1/incomes              # 收益記錄
- GET    /api/v1/balance-logs         # 餘額變動
- GET    /api/v1/finance/overview     # 財務概覽

### 任務服務 (task)
- GET    /api/v1/tasks                # 任務列表
- GET    /api/v1/tasks/{id}           # 任務詳情
- POST   /api/v1/tasks/{id}/complete  # 完成任務
- GET    /api/v1/user-tasks           # 用戶任務
- POST   /api/v1/check-in             # 簽到
- GET    /api/v1/check-ins            # 簽到記錄

### 管理員服務 (admin)
- POST   /api/v1/admin/login          # 管理員登錄
- GET    /api/v1/admins               # 管理員列表
- POST   /api/v1/admins               # 創建管理員
- GET    /api/v1/menus                # 菜單列表
- GET    /api/v1/roles                # 角色列表
- PUT    /api/v1/kyc/{id}/review      # 審核KYC
- PUT    /api/v1/withdrawals/{id}/review # 審核提現
- GET    /api/v1/admin-logs           # 操作日誌

### 支付服務 (payment)
- GET    /api/v1/payment-channels     # 支付渠道
- POST   /api/v1/payments             # 創建支付
- GET    /api/v1/payments/{order_no}  # 查詢支付
- POST   /api/v1/payments/callback/{channel_code} # 支付回調

### 系統服務 (system)
- GET    /api/v1/configs              # 配置列表
- PUT    /api/v1/configs              # 設置配置
- GET    /api/v1/profit-sharing-rules # 分賬規則
- GET    /api/v1/risk-controls        # 風控規則
- GET    /api/v1/domains              # 域名列表

## 技術棧

- **框架**: [Kratos](https://go-kratos.dev/) - Go 微服務框架
- **通信**: gRPC + HTTP
- **數據庫**: MySQL (GORM)
- **緩存**: Redis
- **配置**: Protobuf
- **依賴注入**: Wire

## 項目結構

遵循 Kratos 標準項目結構：
- `api/` - API 定義 (Protobuf)
- `service/{name}/cmd/` - 應用程序入口
- `service/{name}/internal/biz/` - 業務邏輯層
- `service/{name}/internal/data/` - 數據訪問層
- `service/{name}/internal/service/` - 服務層實現
- `service/{name}/internal/server/` - 服務器配置

## 開發規範

參見 `.windsurf/skills/my-new-skill` 中的 Kratos 開發技能指南。
