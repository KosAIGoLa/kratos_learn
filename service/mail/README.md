# Mail Service

高性能邮件发送服务，支持批量邮件发送、模板邮件、邮件队列和状态追踪。

## 功能特性

- **单封邮件发送** - 支持 HTML 和纯文本邮件
- **批量邮件发送** - 支持大批量邮件发送，带速率限制
- **邮件模板** - 支持变量替换的邮件模板
- **附件支持** - 支持多附件邮件
- **SMTP 支持** - 支持 TLS/SSL 加密连接
- **邮件状态追踪** - 追踪邮件发送、投递、打开状态
- **服务注册** - 支持 etcd 服务发现
- **限流控制** - 支持每秒/每日发送限制
- **重试机制** - 失败自动重试，指数退避

## 快速开始

### 安装依赖
```bash
make init
```

### 生成代码
```bash
# 生成 proto 文件
make api
make config

# 生成 wire
cd cmd/mail
wire
```

### 编译运行
```bash
# 编译
go build -o ./bin/ ./cmd/mail

# 运行 (修改 configs/config.yaml 中的 SMTP 配置)
./bin/mail -conf ./configs
```

## API 接口

### 发送单封邮件
```bash
POST /v1/mail/send
{
  "email": {
    "to": "user@example.com",
    "subject": "Hello",
    "body": "<h1>Hello World</h1>",
    "content_type": "text/html"
  }
}
```

### 批量发送邮件
```bash
POST /v1/mail/send-batch
{
  "emails": [
    {"to": "user1@example.com", "subject": "Test 1", "body": "Hello 1"},
    {"to": "user2@example.com", "subject": "Test 2", "body": "Hello 2"}
  ],
  "batch_size": 100,
  "rate_limit": "1s"
}
```

### 使用模板发送
```bash
POST /v1/mail/send
{
  "email": {
    "to": "user@example.com"
  },
  "template_id": "welcome",
  "template_data": {
    "name": "John",
    "company": "Example Inc"
  }
}
```

### 查询邮件状态
```bash
GET /v1/mail/status/{message_id}
```

### 查询批量状态
```bash
GET /v1/mail/batch/{batch_id}
```

### 获取模板列表
```bash
GET /v1/mail/templates
```

## 配置说明

编辑 `configs/config.yaml`:

```yaml
email:
  smtp:
    host: smtp.gmail.com
    port: 587
    username: your-email@gmail.com
    password: your-password
    use_tls: true
    from: noreply@example.com
    from_name: Mail Service
  rate_limit:
    max_per_second: 10
    max_per_day: 10000
    max_batch_size: 100
  retry:
    max_retries: 3
    initial_delay: 1s
    backoff_multiplier: 2.0
```

## Docker 部署

```bash
# 构建镜像
docker build -t mail-service .

# 运行
docker run -d -p 8000:8000 -p 9000:9000 \
  -v $(pwd)/configs:/data/conf \
  mail-service
```

## 项目结构

```
service/mail/
├── api/mail/v1/          # API proto 定义
├── cmd/mail/             # 服务入口
│   ├── main.go
│   └── wire.go
├── internal/
│   ├── biz/              # 业务逻辑
│   │   ├── email.go
│   │   ├── email_sender.go
│   │   └── email_usecase.go
│   ├── data/             # 数据层
│   │   └── email.go
│   ├── service/          # 服务实现
│   │   └── mail.go
│   └── server/           # 服务器
│       ├── grpc.go
│       └── http.go
└── configs/              # 配置文件
```

## 开发计划

- [x] 基础邮件发送 (SMTP)
- [x] 批量邮件发送
- [x] 邮件模板系统
- [x] 状态追踪
- [ ] 异步队列 (Redis/RabbitMQ)
- [ ] 邮件 Webhook 回调
- [ ] 邮件统计分析
- [ ] 模板编辑器

