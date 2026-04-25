# Report Service - 报表服务

基于 Kratos 框架实现的报表查询与 Excel 导出服务。

## 功能特性

- **订单报表** - 订单明细查询、统计与导出
- **用户报表** - 用户消费统计与活跃度分析
- **销售报表** - 按天/周/月分组的销售趋势分析
- **商品报表** - 商品销售排行与库存分析
- **Excel 导出** - 所有报表支持一键导出为 Excel
- **服务注册** - 支持 Etcd 服务注册与发现

## 快速开始

### 编译
```bash
cd service/report
make build
```

### 运行
```bash
./bin/report -conf ./configs
```

服务启动后：
- HTTP API: http://localhost:8007
- gRPC: localhost:9007

## API 接口

### 查询接口（支持分页）

| 报表类型 | 接口路径 | 方法 |
|---------|---------|------|
| 订单报表 | `/api/v1/reports/orders` | GET |
| 用户报表 | `/api/v1/reports/users` | GET |
| 销售报表 | `/api/v1/reports/sales` | GET |
| 商品报表 | `/api/v1/reports/products` | GET |

### 导出接口（Excel）

| 报表类型 | 接口路径 | 方法 |
|---------|---------|------|
| 订单报表 | `/api/v1/reports/orders/export` | GET |
| 用户报表 | `/api/v1/reports/users/export` | GET |
| 销售报表 | `/api/v1/reports/sales/export` | GET |
| 商品报表 | `/api/v1/reports/products/export` | GET |

### 使用示例

```bash
# 查询订单报表
curl "http://localhost:8007/api/v1/reports/orders?start_date=2024-01-01T00:00:00Z&end_date=2024-12-31T23:59:59Z&page=1&page_size=10"

# 导出订单报表为 Excel
curl "http://localhost:8007/api/v1/reports/orders/export?start_date=2024-01-01T00:00:00Z&end_date=2024-12-31T23:59:59Z" --output orders.xlsx

# 查询销售报表（按月分组）
curl "http://localhost:8007/api/v1/reports/sales?group_by=month&start_date=2024-01-01T00:00:00Z"
```

## 配置说明

配置文件位置：`configs/config.yaml`

```yaml
server:
  http:
    addr: 0.0.0.0:8007
    timeout: 1s
  grpc:
    addr: 0.0.0.0:9007
    timeout: 1s

data:
  database:
    driver: mysql
    source: root:root@tcp(127.0.0.1:3306)/kratos?parseTime=True&loc=Local
    slave_source: ""  # 可选：从库地址
  redis:
    addr: 127.0.0.1:6379
    password: ""
    db: 0

registry:
  etcd:
    endpoints:
      - 127.0.0.1:2379
    dial_timeout: 5s
```

## 技术栈

- **框架**: [Kratos](https://github.com/go-kratos/kratos) v2
- **数据库**: MySQL + GORM
- **缓存**: Redis
- **Excel**: excelize/v2
- **依赖注入**: Google Wire
- **服务注册**: Etcd
- **协议**: gRPC + HTTP

## 项目结构

```
service/report/
├── api/report/v1/          # API 定义 (protobuf)
│   └── report.proto
├── cmd/report/             # 入口文件
│   ├── main.go
│   └── wire.go
├── configs/                # 配置文件
│   └── config.yaml
├── internal/
│   ├── biz/               # 业务逻辑层
│   │   └── report.go
│   ├── data/              # 数据访问层
│   │   ├── data.go
│   │   └── report.go
│   ├── service/           # 服务层
│   │   └── report.go
│   ├── server/            # 服务器配置
│   │   ├── grpc.go
│   │   ├── http.go
│   │   └── server.go
│   └── conf/              # 配置定义
│       └── conf.proto
├── bin/                   # 编译输出
├── Makefile
└── README.md
```

## 开发命令

```bash
# 生成 API 代码
make api

# 生成配置代码
make config

# 生成依赖注入代码
cd cmd/report && GOWORK=off go generate

# 编译
make build

# 完整构建
make all
```

## Docker 部署

```bash
# 构建镜像
docker build -t report-service .

# 运行容器
docker run --rm -p 8007:8007 -p 9007:9007 \
  -v $(pwd)/configs:/data/conf \
  report-service
```

## 详细文档

更多详细信息请查看 [README_REPORT.md](./README_REPORT.md)

## 许可证

MIT License


