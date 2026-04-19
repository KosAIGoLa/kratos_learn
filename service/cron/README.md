# Cron Service - 定时任务调度服务

本服务负责系统中所有定时任务的调度和执行，基于 `robfig/cron` 库实现。

## 核心功能

### 1. 算力生产相关任务

| 任务名称 | 执行时间 | 功能说明 | 业务公式 |
|---------|---------|---------|---------|
| `hourly_hashrate_settlement` | 每小时 | 每小时算力结算 | 每小时算力 = 产品级别 × 生产力% |
| `daily_hashrate_settlement` | 每天 00:00 | 每日算力汇总结算 | 每日总算力 = Σ(每小时产出) |
| `hashrate_to_balance` | 每天 00:30 | 算力转可提现余额 | 可提现余额 = 算力 × 1:1 转换 |
| `expired_machine_cleanup` | 每天 01:00 | 过期机器清理 | - |

### 2. 每日签到和任务相关任务

| 任务名称 | 执行时间 | 功能说明 |
|---------|---------|---------|
| `daily_checkin_reset` | 每天 00:00 | 重置每日签到状态 |
| `daily_task_reward` | 每天 08:00 | 发放签到工分奖励 |
| `consecutive_checkin_check` | 每天 00:05 | 检查连续签到天数 |

**签到奖励公式**: 基础奖励(5工分) + 连续签到天数 × 0.5

### 3. 分账和结算任务

| 任务名称 | 执行时间 | 功能说明 |
|---------|---------|---------|
| `profit_sharing` | 每10分钟 | 执行订单/充值分润 |
| `invite_reward_settlement` | 每天 09:00 | 结算邀请奖励 |

**邀请奖励**: 一级邀请 10工分，二级邀请 5工分

### 4. 系统维护任务

| 任务名称 | 执行时间 | 功能说明 |
|---------|---------|---------|
| `health_check` | 每分钟 | 服务健康检查 |
| `data_cleanup` | 每天 02:00 | 数据清理 |
| `daily_report` | 每天 02:30 | 生成每日报表 |
| `withdrawal_limit_reset` | 每天 00:00 | 重置每日提现次数限制 |
| `user_daily_reset` | 每天 00:00 | 重置用户每日工分/任务状态 |
| `weekly_report` | 每周一 09:00 | 生成周报表 |
| `monthly_report` | 每月1号 03:00 | 生成月报表 |

## API 接口

### HTTP API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/cron/v1/tasks` | 列出所有定时任务 |
| POST | `/cron/v1/tasks` | 添加自定义定时任务 |
| DELETE | `/cron/v1/tasks/{name}` | 删除指定任务 |

### gRPC API

- `ListTasks` - 列出所有任务
- `AddTask` - 添加任务
- `RemoveTask` - 移除任务

## 项目结构

```
cron/
├── api/
│   └── cron/v1/
│       ├── cron.proto          # API 定义
│       ├── cron.pb.go          # Protobuf 生成代码
│       ├── cron_grpc.pb.go     # gRPC 生成代码
│       └── cron_http.pb.go     # HTTP 生成代码
├── cmd/cron/
│   ├── main.go                 # 服务入口
│   ├── wire.go                 # 依赖注入配置
│   └── wire_gen.go             # Wire 生成代码
├── internal/
│   ├── cron/
│   │   ├── provider.go         # 依赖注入 Provider
│   │   ├── scheduler.go        # Cron 调度器核心
│   │   └── tasks.go            # 所有定时任务实现
│   ├── service/
│   │   ├── service.go          # Service ProviderSet
│   │   └── cron.go             # Cron HTTP/gRPC 服务实现
│   ├── server/
│   │   ├── http.go             # HTTP 服务器
│   │   └── grpc.go             # gRPC 服务器
│   ├── conf/
│   │   └── conf.proto          # 配置定义
│   ├── data/
│   │   └── data.go             # 数据层
│   └── biz/
│       └── biz.go              # 业务层
├── configs/
│   └── config.yaml             # 配置文件
├── go.mod
├── go.sum
├── Makefile
└── README.md                   # 本文件
```

## 业务数据流

```
算力生产流程:
购买机器 → 每小时算力结算 → 24小时日结算 → 算力转可提现余额

签到任务流程:
每天0点重置 → 用户签到 → 每天8点发放奖励 → 检查连续签到

分润流程:
订单/充值 → 每10分钟分账 → 邀请奖励结算 → 工分发放
```

## 配置文件

```yaml
server:
  http:
    addr: 0.0.0.0:8109
    timeout: 1s
  grpc:
    addr: 0.0.0.0:9109
    timeout: 1s
data:
  database:
    driver: mysql
    source: root:root@tcp(localhost:3306)/kratos_learn
```

## 构建和运行

### 本地开发

```bash
# 编译 proto
make api

# 生成 wire 代码
go run github.com/google/wire/cmd/wire@latest ./cmd/cron

# 构建
go build -o bin/cron ./cmd/cron

# 运行
./bin/cron -conf ./configs
```

### Docker 构建

```bash
# 构建镜像
docker build -t kratos-cron .

# 运行容器
docker run -p 8109:8109 -p 9109:9109 kratos-cron
```

## 数据库表关联

本服务涉及的数据库表：

| 表名 | 说明 | 相关任务 |
|------|------|---------|
| `users` | 用户信息 | 所有用户相关任务 |
| `mining_machines` | 矿机/算力产品 | 算力生产任务 |
| `user_hashrates` | 用户算力持有 | 算力结算任务 |
| `member_benefits` | 会员权益 | 权益清理任务 |
| `user_daily_tasks` | 每日任务记录 | 任务重置/奖励 |
| `check_ins` | 签到记录 | 签到相关任务 |
| `orders` | 订单表 | 分账任务 |
| `recharges` | 充值表 | 分账任务 |
| `profit_sharing_rules` | 分账规则 | 分账任务 |

详见 `/sql/CRON_SQL_SYNC.md` 字段对照文档。

## 定时任务表达式说明

```
┌───────────── 秒 (0-59)
│ ┌───────────── 分钟 (0-59)
│ │ ┌───────────── 小时 (0-23)
│ │ │ ┌───────────── 日 (1-31)
│ │ │ │ ┌───────────── 月 (1-12)
│ │ │ │ │ ┌───────────── 星期 (0-6, 0=周日)
│ │ │ │ │ │
* * * * * *

示例:
0 0 * * * *     - 每小时执行
0 0 0 * * *     - 每天0点执行
0 */10 * * * *  - 每10分钟执行
0 0 9 * * 1     - 每周一9点执行
0 0 3 1 * *     - 每月1号3点执行
```

## 开发指南

### 添加新任务

1. 在 `internal/cron/tasks.go` 中添加任务实现函数
2. 在 `RegisterTasks()` 中注册任务
3. 如有需要，在数据库 SQL schema 中添加相关表/字段

示例:

```go
// 任务实现
func (tm *TaskManager) myNewTask() {
    tm.log.Info("执行新任务...")
    // 业务逻辑
}

// 注册任务
if err := tm.scheduler.AddTask("my_new_task", "0 0 * * * *", tm.myNewTask); err != nil {
    return err
}
```

### 测试任务

```bash
# 临时修改任务表达式为每分钟执行进行测试
# 在 RegisterTasks 中修改:
// if err := tm.scheduler.AddTask("test_task", "0 * * * * *", tm.testTask); err != nil {

# 查看日志输出
tail -f logs/cron.log
```

## 注意事项

1. **任务幂等性**: 所有任务应该支持重复执行而不产生副作用
2. **事务处理**: 涉及数据库操作的任务应该使用事务
3. **错误处理**: 任务失败应记录日志但不影响其他任务
4. **资源清理**: 及时清理过期数据避免表膨胀
5. **监控告警**: 关键任务应有监控和失败告警

## License

MIT License
