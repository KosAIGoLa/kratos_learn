# Report Service - 报表服务

## 功能概述

报表服务提供了完整的报表查询和 Excel 导出功能，支持以下报表类型：

### 1. 订单报表 (Order Report)
- **查询接口**: `GET /api/v1/reports/orders`
- **导出接口**: `GET /api/v1/reports/orders/export`
- **功能**: 
  - 按时间范围、状态筛选订单
  - 分页查询
  - 统计总金额、总订单数
  - 导出为 Excel 文件

### 2. 用户报表 (User Report)
- **查询接口**: `GET /api/v1/reports/users`
- **导出接口**: `GET /api/v1/reports/users/export`
- **功能**:
  - 按注册时间筛选用户
  - 显示用户订单数、消费金额
  - 统计总用户数、活跃用户数
  - 导出为 Excel 文件

### 3. 销售报表 (Sales Report)
- **查询接口**: `GET /api/v1/reports/sales`
- **导出接口**: `GET /api/v1/reports/sales/export`
- **功能**:
  - 按天/周/月分组统计
  - 显示销售金额、订单数、商品数
  - 计算平均订单金额
  - 导出为 Excel 文件

### 4. 商品报表 (Product Report)
- **查询接口**: `GET /api/v1/reports/products`
- **导出接口**: `GET /api/v1/reports/products/export`
- **功能**:
  - 按时间范围统计商品销售
  - 显示销售数量、金额、库存
  - 按销售金额排序
  - 导出为 Excel 文件

## API 示例

### 查询订单报表
```bash
curl -X GET "http://localhost:8007/api/v1/reports/orders?start_date=2024-01-01T00:00:00Z&end_date=2024-12-31T23:59:59Z&status=1&page=1&page_size=10"
```

### 导出订单报表
```bash
curl -X GET "http://localhost:8007/api/v1/reports/orders/export?start_date=2024-01-01T00:00:00Z&end_date=2024-12-31T23:59:59Z&status=1&format=xlsx" --output order_report.xlsx
```

### 查询销售报表（按天分组）
```bash
curl -X GET "http://localhost:8007/api/v1/reports/sales?start_date=2024-01-01T00:00:00Z&end_date=2024-12-31T23:59:59Z&group_by=day"
```

### 导出用户报表
```bash
curl -X GET "http://localhost:8007/api/v1/reports/users/export?start_date=2024-01-01T00:00:00Z&end_date=2024-12-31T23:59:59Z&format=xlsx" --output user_report.xlsx
```

## 配置说明

### 服务端口
- HTTP: `8007`
- gRPC: `9007`

### 数据库配置
在 `configs/config.yaml` 中配置：
```yaml
data:
  database:
    driver: mysql
    source: root:root@tcp(127.0.0.1:3306)/kratos?parseTime=True&loc=Local
    slave_source: ""  # 可选：从库地址
  redis:
    network: tcp
    addr: 127.0.0.1:6379
    password: ""
    db: 0
    dial_timeout: 1s
    read_timeout: 0.2s
    write_timeout: 0.2s
```

### 服务注册 (Etcd)
在 `configs/config.yaml` 中配置 Etcd：
```yaml
registry:
  etcd:
    endpoints:
      - 127.0.0.1:2379
    dial_timeout: 5s
```

## 数据库表依赖

报表服务需要以下数据库表：

1. **orders** - 订单表
   - id, order_no, user_id, name, product_id, product_name, amount, quantity, status, created_at

2. **users** - 用户表
   - id, username, email, phone, created_at, last_login_at

3. **products** - 商品表
   - id, name, stock

## 编译和运行

### 编译
```bash
cd service/report
make build
```

### 运行
```bash
./bin/report -conf ./configs
```

### 生成 API 代码
```bash
make api
```

### 生成依赖注入代码
```bash
cd cmd/report
GOWORK=off go generate
```

## 技术栈

- **框架**: Kratos v2
- **数据库**: MySQL (GORM)
- **缓存**: Redis
- **Excel**: excelize/v2
- **依赖注入**: Wire
- **协议**: gRPC + HTTP

## 项目结构

```
service/report/
├── api/report/v1/          # API 定义 (protobuf)
├── cmd/report/             # 入口文件
├── configs/                # 配置文件
├── internal/
│   ├── biz/               # 业务逻辑层
│   │   └── report.go      # 报表用例
│   ├── data/              # 数据访问层
│   │   ├── data.go        # 数据库/Redis 初始化
│   │   └── report.go      # 报表数据查询
│   ├── service/           # 服务层
│   │   └── report.go      # gRPC/HTTP 服务实现 + Excel 导出
│   ├── server/            # 服务器配置
│   └── conf/              # 配置定义
└── bin/                   # 编译输出
```

## Excel 导出功能

所有报表都支持导出为 Excel 格式，特点：

- ✅ 自动生成表头（加粗、灰色背景）
- ✅ 数据按列对齐
- ✅ 底部汇总行（总计、平均值等）
- ✅ 时间格式化（YYYY-MM-DD HH:mm:ss）
- ✅ 状态文本转换（如：0→待支付，1→已支付）
- ✅ 文件名包含时间戳
- ✅ 直接返回文件二进制数据

## 注意事项

1. 确保数据库表结构与代码中的查询匹配
2. 大数据量导出时注意内存使用
3. 建议使用从库进行报表查询，避免影响主库性能
4. Excel 导出接口返回的是文件二进制数据，需要正确处理响应头
5. 时间参数使用 RFC3339 格式（如：2024-01-01T00:00:00Z）
