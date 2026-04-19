# Cron Service 与 SQL Schema 字段同步文档

## 同步说明

本文档记录了 cron service 中的 Member 结构体与 SQL 数据库表的字段对应关系。

---

## 1. users 表字段对应

### Cron Member 结构体
```go
type Member struct {
    ID                      string
    Level                   string        // -> users.member_level
    DailyPoints             int           // -> users.daily_work_points
    MonthlyPoints           int           // -> users.monthly_work_points
    TotalPoints             int           // -> users.work_points
    DailyPointsLimitReached bool          // -> users.daily_points_limit_reached
    DailyTasks              []DailyTaskStatus  // -> user_daily_tasks 表
    Benefits                []MemberBenefit    // -> member_benefits 表
    ExpireTime              time.Time     // -> users.member_expire_time
    LastResetTime           time.Time     // -> users.last_reset_time
    LevelUpdateTime         time.Time     // -> users.level_update_time
}
```

### SQL users 表新增字段
```sql
-- 会员等级和工分相关字段 (cron service 同步)
`member_level` VARCHAR(20) DEFAULT 'bronze' COMMENT '会员等级: bronze/silver/gold/platinum/diamond',
`daily_work_points` DECIMAL(10,2) DEFAULT 0.00 COMMENT '当日工分',
`monthly_work_points` DECIMAL(10,2) DEFAULT 0.00 COMMENT '当月累计工分',
`daily_points_limit_reached` TINYINT DEFAULT 0 COMMENT '当日工分上限是否已达: 1是 0否',
`member_expire_time` DATETIME DEFAULT NULL COMMENT '会员过期时间',
`last_reset_time` DATETIME DEFAULT NULL COMMENT '最后重置时间(每日重置)',
`level_update_time` DATETIME DEFAULT NULL COMMENT '等级更新时间',
```

---

## 2. 新增关联表

### member_benefits 表 (会员权益)
对应 cron 中的 `MemberBenefit` 结构体

```sql
CREATE TABLE IF NOT EXISTS `member_benefits` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `user_id` INT UNSIGNED NOT NULL COMMENT '用户ID',
    `name` VARCHAR(100) NOT NULL COMMENT '权益名称',
    `benefit_type` VARCHAR(50) NOT NULL COMMENT '权益类型: discount折扣/free_shipping免邮/coupon优惠券/priority优先',
    `value` DECIMAL(10,2) DEFAULT 0.00 COMMENT '权益价值',
    `expire_time` DATETIME NOT NULL COMMENT '过期时间',
    `status` TINYINT DEFAULT 1 COMMENT '状态: 1有效 0已过期 2已使用',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_expire_time` (`expire_time`),
    INDEX `idx_status` (`status`),
    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会员权益表';
```

### user_daily_tasks 表 (用户每日任务)
对应 cron 中的 `DailyTaskStatus` 结构体

```sql
CREATE TABLE IF NOT EXISTS `user_daily_tasks` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `user_id` INT UNSIGNED NOT NULL COMMENT '用户ID',
    `task_id` INT UNSIGNED NOT NULL COMMENT '任务ID',
    `task_name` VARCHAR(100) NOT NULL COMMENT '任务名称',
    `completed` TINYINT DEFAULT 0 COMMENT '是否完成: 1是 0否',
    `completed_at` DATETIME DEFAULT NULL COMMENT '完成时间',
    `points` DECIMAL(10,2) DEFAULT 0.00 COMMENT '任务工分值',
    `task_date` DATE NOT NULL COMMENT '任务日期(每日重置)',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY `uk_user_task_date` (`user_id`, `task_id`, `task_date`),
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_task_id` (`task_id`),
    INDEX `idx_task_date` (`task_date`),
    INDEX `idx_completed` (`completed`),
    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE,
    FOREIGN KEY (`task_id`) REFERENCES `tasks`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户每日任务完成记录表';
```

---

## 3. 字段类型对照表

| Cron 字段 | Go 类型 | SQL 字段 | SQL 类型 | 说明 |
|-----------|---------|----------|----------|------|
| ID | string | id | INT UNSIGNED | 用户ID |
| Level | string | member_level | VARCHAR(20) | 会员等级 |
| DailyPoints | int | daily_work_points | DECIMAL(10,2) | 当日工分 |
| MonthlyPoints | int | monthly_work_points | DECIMAL(10,2) | 当月工分 |
| TotalPoints | int | work_points | DECIMAL(15,2) | 总工分 |
| DailyPointsLimitReached | bool | daily_points_limit_reached | TINYINT | 是否达上限 |
| ExpireTime | time.Time | member_expire_time | DATETIME | 会员过期时间 |
| LastResetTime | time.Time | last_reset_time | DATETIME | 最后重置时间 |
| LevelUpdateTime | time.Time | level_update_time | DATETIME | 等级更新时间 |
| Benefits[].Name | string | name | VARCHAR(100) | 权益名称 |
| Benefits[].ExpireTime | time.Time | expire_time | DATETIME | 权益过期时间 |
| Benefits[].Value | float64 | value | DECIMAL(10,2) | 权益价值 |
| DailyTasks[].TaskID | string | task_id | INT UNSIGNED | 任务ID |
| DailyTasks[].TaskName | string | task_name | VARCHAR(100) | 任务名称 |
| DailyTasks[].Completed | bool | completed | TINYINT | 是否完成 |
| DailyTasks[].CompletedAt | time.Time | completed_at | DATETIME | 完成时间 |
| DailyTasks[].Points | int | points | DECIMAL(10,2) | 任务工分 |

---

## 4. Cron 定时任务配置

| 任务名称 | 执行时间 | 操作表 | 操作说明 |
|----------|----------|--------|----------|
| member_daily_reset | 每天 03:00 | users, user_daily_tasks | 重置当日工分，清理昨日任务 |
| member_benefits_cleanup | 每天 03:30 | member_benefits | 清理过期权益 |
| member_level_check | 每天 04:00 | users | 检查更新会员等级 |
| expired_members_cleanup | 每天 01:00 | users | 清理过期会员 |
| member_monthly_settlement | 每月1号 05:00 | users | 月度工分结算 |

---

## 5. 数据同步建议

### 每日重置流程
```
1. member_daily_reset (03:00)
   - 将所有用户的 daily_work_points 累加到 monthly_work_points
   - 将 daily_work_points 重置为 0
   - 将 daily_points_limit_reached 设为 0
   - 删除/归档昨日的 user_daily_tasks 记录
   - 生成新的当日任务记录

2. member_benefits_cleanup (03:30)
   - 查询 member_benefits 中 expire_time < NOW() 的记录
   - 将 status 更新为 0 (已过期)

3. member_level_check (04:00)
   - 根据 monthly_work_points 计算新等级
   - 更新 users.member_level
   - 更新 users.level_update_time
```

### 月度结算流程
```
member_monthly_settlement (每月1号 05:00)
- 将 monthly_work_points 累加到 work_points
- 将 monthly_work_points 重置为 0
- 生成月度结算报告
```

---

## 6. 注意事项

1. **工分类型**: Cron 中使用 int，SQL 中使用 DECIMAL(10,2) 支持小数精度
2. **布尔类型**: Cron 中使用 bool，SQL 中使用 TINYINT (0/1)
3. **时间类型**: Cron 中使用 time.Time，SQL 中使用 DATETIME
4. **数组类型**: Cron 中的数组在 SQL 中拆分为独立表 (member_benefits, user_daily_tasks)
5. **每日任务**: user_daily_tasks 表使用 task_date 字段区分每日任务，支持历史查询
