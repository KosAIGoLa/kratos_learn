---
trigger: manual
description: Kratos微服務實戰指南 (二) 配置管理 (LearnKu文章64942)
globs: 
---

## Kratos 微服務實戰之用戶服務 (二) 配置管理

參考文章: https://learnku.com/articles/64942

---

### 1. 修改配置文件

#### 修改 `user/configs/config.yaml` 文件

```yaml
server:
  http:
    addr: 0.0.0.0:8000
    timeout: 1s
  grpc:
    addr: 0.0.0.0:50051
    timeout: 1s

data:
  database:
    driver: mysql
    source: root:root@tcp(127.0.0.1:3306)/shop_user?charset=utf8mb4&parseTime=True&loc=Local
  redis:
    addr: 127.0.0.1:6379
    dial_timeout: 1s
    read_timeout: 0.2s
    write_timeout: 0.2s

trace:
  endpoint: http://127.0.0.1:14268/api/traces
```

> 具體連接 mysql、redis 的參數填寫自己本機的，本項目用到的是 gorm。
> trace 是以後要用到的鏈路追踪的參數，先定義了。

---

#### 新建 `user/configs/registry.yaml` 文件

引入 [consul](https://www.consul.io/) 服務：

```yaml
# 這裡引入了 consul 的服務註冊與發現，先把配置加入進去
consul:
  address: 127.0.0.1:8500
  scheme: http
```

---

#### 修改 `user/internal/conf/conf.proto` 配置文件

在文件底部新增 consul 和 trace 的配置信息：

```protobuf
message Trace {
  string endpoint = 1;
}

message Registry {
  message Consul {
    string address = 1;
    string scheme  = 2;
  }
  Consul consul = 1;
}
```

完整的 Bootstrap 配置：

```protobuf
message Bootstrap {
  Server   server   = 1;
  Data     data     = 2;
  Registry registry = 3;
  Trace    trace    = 4;
}
```

---

#### 新生成 `conf.pb.go` 文件

在 `service/user` 目錄下，執行命令：

```bash
make config
```

---

#### 安裝 consul 服務工具

使用 docker 工具進行創建：

```bash
docker run -d -p 8500:8500 -p 8300:8300 -p 8301:8301 -p 8302:8302 -p 8600:8600/udp consul consul agent -dev -client=0.0.0.0
```

瀏覽器訪問 `http://127.0.0.1:8500/ui/dc1/services` 測試是否安裝成功。

---

### 下一步

完成配置管理後，請參考：

- `22-kratos-learnku-data.md` - 數據層實現
- `23-kratos-learnku-biz-service.md` - 業務層與服務層
- `24-kratos-learnku-server.md` - 服務器配置與啟動
- `25-kratos-learnku-testing.md` - 測試

---

### 參考資源

- 源碼: https://github.com/aliliin/kratos-shop
- Consul: https://www.consul.io/
- Wire 依賴注入: https://go-kratos.dev/blog/go-project-wire
