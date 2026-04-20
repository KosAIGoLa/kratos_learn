# ETCD 服務發現集成文檔

## 概述

所有微服務已集成 etcd 作為服務發現和註冊中心，使用 [go-kratos/kratos/contrib/registry/etcd](https://github.com/go-kratos/kratos/tree/main/contrib/registry/etcd)。

## 已集成服務

- ✅ **admin** - 管理服務 (端口: HTTP 8108, gRPC 9108)
- ✅ **user** - 用戶服務 (端口: HTTP 8107, gRPC 9107)
- ⏳ **content** - 內容服務 (端口: HTTP 8101, gRPC 9101)
- ⏳ **cron** - 定時任務服務 (端口: HTTP 8102, gRPC 9102)
- ⏳ **finance** - 財務服務 (端口: HTTP 8103, gRPC 9103)
- ⏳ **order** - 訂單服務 (端口: HTTP 8104, gRPC 9104)
- ⏳ **payment** - 支付服務 (端口: HTTP 8105, gRPC 9105)
- ⏳ **product** - 產品服務 (端口: HTTP 8106, gRPC 9106)
- ⏳ **system** - 系統服務 (端口: HTTP 8109, gRPC 9109)

## 配置說明

### etcd 配置

每個服務的 `configs/config.yaml` 中添加以下配置：

```yaml
registry:
  etcd:
    endpoints:
      - 127.0.0.1:2379
    dial_timeout: 5s
```

### 配置參數說明

- `endpoints`: etcd 服務器地址列表，支持多個節點
- `dial_timeout`: 連接超時時間

## 實現步驟

### 1. 更新 conf.proto

在 `internal/conf/conf.proto` 中添加：

```protobuf
message Bootstrap {
  Server server = 1;
  Data data = 2;
  // ... 其他配置
  Registry registry = N;  // 添加 registry 字段
}

message Registry {
  message Etcd {
    repeated string endpoints = 1;
    google.protobuf.Duration dial_timeout = 2;
  }
  Etcd etcd = 1;
}
```

### 2. 更新 go.mod

添加依賴：

```go
require (
    github.com/go-kratos/kratos/contrib/registry/etcd/v2 v2.0.0-20260404020628-f149714c1d54
    go.etcd.io/etcd/client/v3 v3.5.11
    // ... 其他依賴
)
```

### 3. 更新 server.go

在 `internal/server/server.go` 中添加：

```go
package server

import (
	"<service>/internal/conf"

	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/google/wire"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var ProviderSet = wire.NewSet(NewGRPCServer, NewHTTPServer, NewRegistrar)

func NewRegistrar(conf *conf.Registry) registry.Registrar {
	if conf == nil || conf.Etcd == nil {
		return nil
	}
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   conf.Etcd.Endpoints,
		DialTimeout: conf.Etcd.DialTimeout.AsDuration(),
	})
	if err != nil {
		panic(err)
	}
	return etcd.New(client)
}
```

### 4. 更新 wire.go

在 `cmd/<service>/wire.go` 中：

```go
func wireApp(*conf.Server, *conf.Data, *conf.Registry, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
```

### 5. 更新 main.go

在 `cmd/<service>/main.go` 中：

```go
import (
	// ... 其他導入
	"github.com/go-kratos/kratos/v2/registry"
)

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server, r registry.Registrar) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(gs, hs),
		kratos.Registrar(r),  // 添加註冊器
	)
}

func main() {
	// ... 配置加載
	app, cleanup, err := wireApp(bc.Server, bc.Data, bc.Registry, logger)
	// ... 其他代碼
}
```

### 6. 重新生成代碼

```bash
cd service/<service>

# 下載依賴
go mod tidy

# 重新生成 protobuf
make api

# 重新生成 wire
go generate ./...
```

## 服務發現使用

### 服務註冊

服務啟動時會自動註冊到 etcd：

- **服務名**: 通過 `kratos.Name()` 設置
- **服務 ID**: 通過 `kratos.ID()` 設置（默認為主機名）
- **服務地址**: 自動從 HTTP/gRPC 服務器配置中獲取

### 服務發現

其他服務可以通過 etcd 發現並連接：

```go
import (
	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// 創建 etcd 客戶端
client, err := clientv3.New(clientv3.Config{
	Endpoints: []string{"127.0.0.1:2379"},
})
if err != nil {
	panic(err)
}

// 創建服務發現
r := etcd.New(client)

// 連接到服務
conn, err := grpc.DialInsecure(
	context.Background(),
	grpc.WithEndpoint("discovery:///user"),  // discovery:/// 前綴表示使用服務發現
	grpc.WithDiscovery(r),
)
```

## 監控和管理

### 查看已註冊服務

使用 etcdctl 查看：

```bash
# 列出所有服務
etcdctl get --prefix /microservices/

# 查看特定服務
etcdctl get --prefix /microservices/user/
```

### 健康檢查

Kratos 會自動進行健康檢查：

- 服務正常運行時，會定期更新租約
- 服務停止時，會自動從 etcd 中移除註冊信息
- 租約過期時（默認 30 秒），服務會被自動移除

## 故障排查

### 1. 服務無法註冊

檢查 etcd 連接：

```bash
# 測試 etcd 連接
etcdctl endpoint health

# 檢查服務日誌
tail -f /var/log/<service>.log
```

### 2. 服務發現失敗

- 確認 etcd 配置正確
- 檢查網絡連接
- 驗證服務名稱是否正確

### 3. 連接超時

調整 `dial_timeout` 配置：

```yaml
registry:
  etcd:
    endpoints:
      - 127.0.0.1:2379
    dial_timeout: 10s  # 增加超時時間
```

## 生產環境建議

### 1. etcd 集群配置

使用多節點 etcd 集群：

```yaml
registry:
  etcd:
    endpoints:
      - etcd1.example.com:2379
      - etcd2.example.com:2379
      - etcd3.example.com:2379
    dial_timeout: 5s
```

### 2. 安全配置

啟用 TLS：

```yaml
registry:
  etcd:
    endpoints:
      - etcd.example.com:2379
    dial_timeout: 5s
    tls:
      cert_file: /path/to/cert.pem
      key_file: /path/to/key.pem
      ca_file: /path/to/ca.pem
```

### 3. 監控

- 監控 etcd 集群健康狀態
- 監控服務註冊/註銷事件
- 設置告警規則

## 測試

### 本地測試

1. 啟動 etcd：

```bash
docker run -d --name etcd \
  -p 2379:2379 \
  -p 2380:2380 \
  quay.io/coreos/etcd:latest \
  /usr/local/bin/etcd \
  --advertise-client-urls http://0.0.0.0:2379 \
  --listen-client-urls http://0.0.0.0:2379
```

2. 啟動服務：

```bash
cd service/user
go run cmd/user/main.go -conf configs
```

3. 驗證註冊：

```bash
etcdctl get --prefix /microservices/
```

## 參考資料

- [Kratos Registry](https://go-kratos.dev/docs/component/registry)
- [etcd Documentation](https://etcd.io/docs/)
- [go-kratos/kratos/contrib/registry/etcd](https://github.com/go-kratos/kratos/tree/main/contrib/registry/etcd)
