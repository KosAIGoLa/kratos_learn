---
trigger: manual
description: Kratos 配置管理规范
globs: ["**/configs/*.yaml", "**/internal/conf/*.proto"]
---

# Kratos 配置管理规范

## 1. 配置文件 (configs/config.yaml)

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
    source: root:root@tcp(127.0.0.1:3306)/{db_name}?charset=utf8mb4&parseTime=True&loc=Local
  redis:
    addr: 127.0.0.1:6379
    db: 0
    dial_timeout: 1s
    read_timeout: 0.2s
    write_timeout: 0.2s

# 服务注册 (可选)
registry:
  etcd:
    endpoints:
      - 127.0.0.1:2379
    dial_timeout: 1s

# 链路追踪 (可选)
trace:
  endpoint: http://127.0.0.1:14268/api/traces
```

## 2. Proto 配置定义 (internal/conf/conf.proto)

```protobuf
syntax = "proto3";
package kratos.conf;

option go_package = "{service}/internal/conf;conf";

import "google/protobuf/duration.proto";

message Bootstrap {
  Server server = 1;
  Data data = 2;
  Registry registry = 3;
  Trace trace = 4;
  JWT jwt = 5;
}

message Server {
  message HTTP {
    string network = 1;
    string addr = 2;
    google.protobuf.Duration timeout = 3;
  }
  message GRPC {
    string network = 1;
    string addr = 2;
    google.protobuf.Duration timeout = 3;
  }
  HTTP http = 1;
  GRPC grpc = 2;
}

message Data {
  message Database {
    string driver = 1;
    string source = 2;
  }
  message Redis {
    string addr = 1;
    int32 db = 2;
    google.protobuf.Duration dial_timeout = 3;
    google.protobuf.Duration read_timeout = 4;
    google.protobuf.Duration write_timeout = 5;
  }
  Database database = 1;
  Redis redis = 2;
}

message Registry {
  message Etcd {
    repeated string endpoints = 1;
    google.protobuf.Duration dial_timeout = 2;
  }
  Etcd etcd = 1;
}

message Trace {
  string endpoint = 1;
}

message JWT {
  string secret = 1;
  google.protobuf.Duration access_token_ttl = 2;
  google.protobuf.Duration refresh_token_ttl = 3;
}
```

## 3. 生成配置代码

```bash
cd service/{name}
make config

# 输出: internal/conf/conf.pb.go
```

## 4. 配置加载 (main.go)

```go
func main() {
    flag.Parse()
    
    // 创建配置源
    c := config.New(
        config.WithSource(
            file.NewSource(flagconf),
        ),
    )
    defer c.Close()
    
    if err := c.Load(); err != nil {
        panic(err)
    }
    
    var bc conf.Bootstrap
    if err := c.Scan(&bc); err != nil {
        panic(err)
    }
    
    // 使用 bc.Server, bc.Data, bc.Registry...
}
```

## 5. 环境变量支持

配置文件支持环境变量替换：

```yaml
data:
  database:
    driver: mysql
    source: ${DB_DSN:root:root@tcp(127.0.0.1:3306)/default}
  redis:
    addr: ${REDIS_ADDR:127.0.0.1:6379}
```

## 下一步

- [数据层规范](04-data-layer.md) → 数据模型与仓储实现
