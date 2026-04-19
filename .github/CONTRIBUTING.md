# 貢獻指南

感謝您對本項目的關注！我們歡迎所有形式的貢獻。

## 行為準則

請閱讀並遵守我們的 [行為準則](CODE_OF_CONDUCT.md)。

## 如何貢獻

### 報告 Bug

在報告 Bug 之前，請：

1. 檢查 [Issues](../../issues) 確保問題尚未被報告
2. 使用最新版本測試問題是否仍然存在
3. 使用 Bug 報告模板創建新 Issue

### 建議新功能

1. 檢查 [Issues](../../issues) 確保功能尚未被建議
2. 使用功能請求模板創建新 Issue
3. 清楚描述功能的用例和預期行為

### 提交 Pull Request

#### 開發流程

1. **Fork 項目**
   ```bash
   git clone https://github.com/your-username/my-front-app.git
   cd my-front-app
   ```

2. **創建分支**
   ```bash
   git checkout -b feature/your-feature-name
   # 或
   git checkout -b fix/your-bug-fix
   ```

3. **安裝依賴**
   ```bash
   cd service/[service-name]
   go mod download
   ```

4. **進行更改**
   - 遵循項目的編碼規範
   - 添加必要的測試
   - 更新相關文檔

5. **運行測試**
   ```bash
   go test ./...
   go test -race ./...
   go test -coverprofile=coverage.out ./...
   ```

6. **運行 Lint**
   ```bash
   golangci-lint run
   ```

7. **提交更改**
   ```bash
   git add .
   git commit -m "feat(service): add new feature"
   ```

   提交消息格式：
   - `feat(scope): 新功能`
   - `fix(scope): Bug 修復`
   - `docs(scope): 文檔更新`
   - `style(scope): 代碼格式化`
   - `refactor(scope): 代碼重構`
   - `test(scope): 測試相關`
   - `chore(scope): 構建/工具相關`

8. **推送到 Fork**
   ```bash
   git push origin feature/your-feature-name
   ```

9. **創建 Pull Request**
   - 使用 PR 模板
   - 清楚描述變更
   - 鏈接相關 Issue
   - 等待審查

#### PR 審查流程

1. 自動化檢查必須通過（CI、測試、Lint）
2. 至少需要一位維護者的批准
3. 解決所有審查意見
4. 保持 PR 更新（rebase 或 merge main）

### 編碼規範

#### Go 代碼風格

- 遵循 [Effective Go](https://golang.org/doc/effective_go.html)
- 使用 `gofmt` 格式化代碼
- 使用 `golangci-lint` 檢查代碼
- 變量名使用駝峰命名法
- 常量使用大寫字母和下劃線

#### 項目結構

遵循 Kratos 項目佈局：

```
service/[service-name]/
├── api/                # Protobuf API 定義
├── cmd/                # 服務入口
├── configs/            # 配置文件
├── internal/
│   ├── biz/           # 業務邏輯層
│   ├── data/          # 數據訪問層
│   ├── server/        # HTTP/gRPC 服務器
│   └── service/       # 服務實現層
└── test/              # 測試文件
```

#### 測試要求

- 所有新功能必須包含單元測試
- 測試覆蓋率應保持在 80% 以上
- 使用表驅動測試
- Mock 外部依賴

示例：

```go
func TestUserService_Register(t *testing.T) {
    tests := []struct {
        name    string
        input   *RegisterRequest
        want    *RegisterResponse
        wantErr bool
    }{
        {
            name: "valid registration",
            input: &RegisterRequest{
                Phone:    "13800138000",
                Password: "password123",
            },
            want: &RegisterResponse{
                UserId: 1,
            },
            wantErr: false,
        },
        // 更多測試用例...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 測試邏輯
        })
    }
}
```

#### 文檔要求

- 所有公開的函數和類型必須有註釋
- 複雜的邏輯需要添加說明
- 更新 README.md（如果需要）
- 更新 API 文檔（如果 API 有變更）

### Proto 文件規範

- 使用 Protocol Buffers v3
- 遵循 [Google API 設計指南](https://cloud.google.com/apis/design)
- 使用語義化版本控制（v1, v2, etc.）
- 添加適當的註釋

### 數據庫遷移

- 使用 GORM AutoMigrate 或手動遷移
- 提供回滾方案
- 在 PR 中說明數據庫變更

## 開發環境設置

### 必需工具

```bash
# 安裝 Go
brew install go

# 安裝 Protobuf 編譯器
brew install protobuf

# 安裝 Kratos CLI
go install github.com/go-kratos/kratos/cmd/kratos/v2@latest

# 安裝 Wire
go install github.com/google/wire/cmd/wire@latest

# 安裝 golangci-lint
brew install golangci-lint
```

### 本地服務

```bash
# 啟動 MySQL
docker run -d --name mysql \
  -e MYSQL_ROOT_PASSWORD=root \
  -e MYSQL_DATABASE=test_db \
  -p 3306:3306 \
  mysql:8.0

# 啟動 Redis
docker run -d --name redis \
  -p 6379:6379 \
  redis:7-alpine

# 啟動 RabbitMQ
docker run -d --name rabbitmq \
  -p 5672:5672 \
  -p 15672:15672 \
  rabbitmq:3.12-management-alpine
```

## 獲取幫助

- 查看 [文檔](../../wiki)
- 在 [Discussions](../../discussions) 提問
- 加入我們的社區頻道

## 許可證

通過貢獻，您同意您的貢獻將在與項目相同的許可證下授權。
