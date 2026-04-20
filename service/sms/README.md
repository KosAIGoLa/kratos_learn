# SMS 短信服务

支持多短信服务商弹性配置和动态选择的微服务。

## 功能特性

- ✅ 支持多短信服务商配置（Twilio、阿里云、腾讯云等）
- ✅ 动态选择短信服务商（按名称、优先级、轮询）
- ✅ 发送失败自动重试机制
- ✅ 完整的短信发送记录和查询
- ✅ 支持批量发送短信
- ✅ 运行时切换默认服务商
- ✅ RESTful API 和 gRPC 双协议支持

## 架构设计

```
sms/
├── api/sms/v1/          # Protobuf API 定义
│   └── sms.proto        # 短信服务接口
├── internal/
│   ├── biz/
│   │   ├── provider.go       # 服务商接口定义
│   │   ├── provider_mock.go  # Mock/Console 实现
│   │   ├── provider_twilio.go # Twilio 实现
│   │   ├── provider_aliyun.go # 阿里云实现
│   │   └── sms.go            # 业务逻辑层
│   ├── data/
│   │   └── sms.go            # 数据存储层
│   ├── service/
│   │   └── sms.go            # API 服务实现
│   └── conf/
│       └── conf.proto        # 配置定义
└── configs/
    └── config.yaml           # 配置文件示例
```

## 支持的短信服务商

| 类型 | 说明 | 状态 |
|------|------|------|
| console | 控制台输出（开发调试） | ✅ 可用 |
| mock | 模拟发送（单元测试） | ✅ 可用 |
| twilio | Twilio SMS | ✅ 可用 |
| aliyun | 阿里云短信 | ✅ 可用 |

## 快速开始

### 1. 配置服务商

编辑 `configs/config.yaml`：

```yaml
sms:
  default_provider: console  # 默认服务商
  retry_times: 2              # 重试次数
  enable_log: true            # 启用记录
  providers:
    # 控制台输出（开发调试）
    - id: console
      name: 控制台输出
      type: console
      enabled: true
      priority: 1
    
    # Twilio（生产环境）
    - id: twilio
      name: Twilio
      type: twilio
      enabled: false
      priority: 2
      config:
        account_sid: your_account_sid
        auth_token: your_auth_token
        from_number: your_phone_number
```

### 2. 启动服务

```bash
# 生成 protobuf 代码
cd service/sms
make api

# 生成 wire 依赖注入
cd cmd/sms
go generate ./...

# 运行服务
cd ../..
go run cmd/sms/main.go -conf ./configs
```

### 3. 发送短信

```bash
curl -X POST http://localhost:8110/v1/sms/send \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "+8613800138000",
    "template_code": "SMS_123456",
    "template_params": {
      "code": "123456"
    }
  }'
```

### 4. 切换服务商

```bash
curl -X POST http://localhost:8110/v1/sms/provider/switch \
  -H "Content-Type: application/json" \
  -d '{
    "provider_id": "twilio"
  }'
```

## API 接口

### 发送短信
```
POST /v1/sms/send
```

### 批量发送
```
POST /v1/sms/batch-send
```

### 获取服务商列表
```
GET /v1/sms/providers
```

### 切换默认服务商
```
POST /v1/sms/provider/switch
```

### 查询发送记录
```
GET /v1/sms/logs?page=1&page_size=20
```

## 扩展新的短信服务商

实现 `SmsProvider` 接口：

```go
type SmsProvider interface {
    ID() string
    Name() string
    Type() string
    SendSms(ctx context.Context, req *SendRequest) (*SendResult, error)
    IsAvailable() bool
}
```

注册到工厂：

```go
registry.Register("your_provider", &YourProviderFactory{})
```

## Docker

```bash
# 构建
docker build -f service/sms/Dockerfile -t sms-service .

# 运行
docker run -p 8110:8110 -p 9110:9110 -v /path/to/configs:/data/conf sms-service
```

## 端口

- HTTP: 8110
- gRPC: 9110

