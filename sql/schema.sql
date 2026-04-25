-- 用户表 (基础)
CREATE TABLE IF NOT EXISTS `users` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `username` VARCHAR(50) NOT NULL UNIQUE COMMENT '用户账号',
    `invite_code` VARCHAR(20) NOT NULL UNIQUE COMMENT '邀请码',
    `phone` CHAR(11) NOT NULL UNIQUE COMMENT '手机号码',
    `password` VARCHAR(255) NOT NULL COMMENT '密码哈希',
    `payment_password` VARCHAR(255) DEFAULT NULL COMMENT '支付密码哈希',
    `name` VARCHAR(50) DEFAULT NULL COMMENT '姓名',
    `id_card` CHAR(18) DEFAULT NULL COMMENT '身份证号',
    `parent_id` INT UNSIGNED DEFAULT NULL COMMENT '上级用户ID',
    `balance` DECIMAL(15,2) DEFAULT 0.00 COMMENT '账户余额',
    `work_points` DECIMAL(15,2) DEFAULT 0.00 COMMENT '工分总额',
    -- 会员等级和工分相关字段 (cron service 同步)
    `member_level` VARCHAR(20) DEFAULT 'bronze' COMMENT '会员等级: bronze/silver/gold/platinum/diamond',
    `daily_work_points` DECIMAL(10,2) DEFAULT 0.00 COMMENT '当日工分',
    `monthly_work_points` DECIMAL(10,2) DEFAULT 0.00 COMMENT '当月累计工分',
    `daily_points_limit_reached` TINYINT DEFAULT 0 COMMENT '当日工分上限是否已达: 1是 0否',
    `member_expire_time` DATETIME DEFAULT NULL COMMENT '会员过期时间',
    `last_reset_time` DATETIME DEFAULT NULL COMMENT '最后重置时间(每日重置)',
    `level_update_time` DATETIME DEFAULT NULL COMMENT '等级更新时间',
    `status` TINYINT DEFAULT 1 COMMENT '状态: 1正常 0冻结',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` DATETIME DEFAULT NULL,

    INDEX `idx_username` (`username`),
    INDEX `idx_invite_code` (`invite_code`),
    INDEX `idx_phone` (`phone`),
    INDEX `idx_parent_id` (`parent_id`),
    INDEX `idx_member_level` (`member_level`),
    INDEX `idx_member_expire_time` (`member_expire_time`),
    INDEX `idx_status` (`status`),

    CONSTRAINT `fk_users_parent`
    FOREIGN KEY (`parent_id`) REFERENCES `users` (`id`)
                                                    ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 实名认证表
CREATE TABLE IF NOT EXISTS `kyc_verifications` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `user_id` INT UNSIGNED NOT NULL COMMENT '用户ID',
    `name` VARCHAR(50) NOT NULL COMMENT '姓名',
    `id_card` VARCHAR(18) NOT NULL COMMENT '身份证号',
    `id_card_front` VARCHAR(255) NOT NULL COMMENT '身份证正面',
    `id_card_back` VARCHAR(255) NOT NULL COMMENT '身份证反面',
    `status` TINYINT DEFAULT 0 COMMENT '状态: 0待审核 1通过 2拒绝',
    `remark` VARCHAR(255) DEFAULT NULL COMMENT '审核备注',
    `verified_at` DATETIME DEFAULT NULL COMMENT '审核时间',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY `uk_user_id` (`user_id`),
    INDEX `idx_status` (`status`),
    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='实名认证表';

-- 团队关系表
CREATE TABLE IF NOT EXISTS `team_relations` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `user_id` INT UNSIGNED NOT NULL COMMENT '用户ID',
    `parent_id` INT UNSIGNED DEFAULT NULL COMMENT '上级ID',
    `path` VARCHAR(500) NOT NULL COMMENT '层级路径 如: 1.2.3 (根到当前)',
    `level` INT UNSIGNED NOT NULL COMMENT '层级深度: 0根 1一级 2二级...',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY `uk_user_id` (`user_id`),
    INDEX `idx_parent_id` (`parent_id`),
    INDEX `idx_path` (`path`),
    INDEX `idx_level` (`level`),
    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE,
    FOREIGN KEY (`parent_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='团队关系表';

-- 矿机/算力产品表
CREATE TABLE IF NOT EXISTS `mining_machines` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(100) NOT NULL COMMENT '矿机名称',
    `model` VARCHAR(50) NOT NULL COMMENT '型号',
    `algorithm` VARCHAR(30) NOT NULL COMMENT '算法: SHA256/ETH/ETC/SCRYPT',
    `hashrate` DECIMAL(15,2) NOT NULL COMMENT '算力值(TH/s或MH/s)',
    `hashrate_unit` VARCHAR(10) NOT NULL COMMENT '算力单位: TH/s/MH/s/GH/s',
    `power_consumption` INT UNSIGNED COMMENT '功耗(W)',
    `duration_days` INT UNSIGNED NOT NULL COMMENT '有效期(天)',
    `price` DECIMAL(15,2) NOT NULL COMMENT '售价',
    `daily_reward_estimate` DECIMAL(15,6) COMMENT '预估日收益',
    `stock` INT UNSIGNED DEFAULT 0 COMMENT '库存数量',
    `status` TINYINT DEFAULT 1 COMMENT '状态: 1上架 0下架',
    `sort` INT UNSIGNED DEFAULT 0 COMMENT '排序',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX `idx_algorithm` (`algorithm`),
    INDEX `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='矿机/算力产品表';

-- 产品表
CREATE TABLE IF NOT EXISTS `products` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `machine_id` INT UNSIGNED DEFAULT NULL COMMENT '关联矿机ID(算力产品必填)',
    `name` VARCHAR(255) NOT NULL COMMENT '产品名称',
    `price` DECIMAL(10,2) NOT NULL COMMENT '产品价格',
    `description` TEXT DEFAULT NULL COMMENT '产品描述',
    `type` VARCHAR(20) DEFAULT 'mining' COMMENT '产品类型: mining算力/physical实物/service服务',
    `cycle` INT UNSIGNED DEFAULT 30 COMMENT '周期(天)',
    `productivity_rate` DECIMAL(5,4) DEFAULT 0.0100 COMMENT '生产力百分比(0-1)',
    `status` TINYINT DEFAULT 1 COMMENT '状态: 1上架 0下架',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX `idx_machine_id` (`machine_id`),
    INDEX `idx_type` (`type`),
    INDEX `idx_status` (`status`),
    FOREIGN KEY (`machine_id`) REFERENCES `mining_machines`(`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='产品表';

-- 订单表 (按月分区)
CREATE TABLE IF NOT EXISTS `orders` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT,
    `order_no` VARCHAR(50) NOT NULL UNIQUE COMMENT '订单号',
    `user_id` INT UNSIGNED NOT NULL COMMENT '用户ID',
    `invite_code` VARCHAR(20) NOT NULL COMMENT '邀请码',
    `phone` VARCHAR(20) NOT NULL COMMENT '手机号码',
    `name` VARCHAR(50) NOT NULL COMMENT '姓名',
    `product_id` INT UNSIGNED NOT NULL COMMENT '产品ID',
    `product_name` VARCHAR(255) NOT NULL COMMENT '产品名称',
    `amount` DECIMAL(10,2) NOT NULL COMMENT '购买金额',
    `quantity` INT UNSIGNED NOT NULL COMMENT '购买数量',
    `status` TINYINT DEFAULT 1 COMMENT '状态: 0待支付 1已完成 2已取消 3退款中 4已退款',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_order_no` (`order_no`),
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_invite_code` (`invite_code`),
    INDEX `idx_phone` (`phone`),
    INDEX `idx_created_at` (`created_at`),
    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE,
    FOREIGN KEY (`product_id`) REFERENCES `products`(`id`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单表';

-- 充值表 (按月分区)
CREATE TABLE IF NOT EXISTS `recharges` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT,
    `user_id` INT UNSIGNED NOT NULL COMMENT '用户ID',
    `invite_code` VARCHAR(20) NOT NULL COMMENT '邀请码',
    `phone` VARCHAR(20) NOT NULL COMMENT '手机号码',
    `name` VARCHAR(50) NOT NULL COMMENT '姓名',
    `order_no` VARCHAR(50) NOT NULL UNIQUE COMMENT '订单号',
    `amount` DECIMAL(10,2) NOT NULL COMMENT '充值金额',
    `status` TINYINT DEFAULT 0 COMMENT '状态: 0待支付 1成功 2失败',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_order_no` (`order_no`),
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_phone` (`phone`),
    INDEX `idx_created_at` (`created_at`),
    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='充值表';

-- 提现表 (按月分区)
CREATE TABLE IF NOT EXISTS `withdrawals` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT,
    `user_id` INT UNSIGNED NOT NULL COMMENT '用户ID',
    `phone` VARCHAR(20) NOT NULL COMMENT '手机号码',
    `name` VARCHAR(50) NOT NULL COMMENT '姓名',
    `bank_card` VARCHAR(30) NOT NULL COMMENT '银行卡号',
    `bank_name` VARCHAR(50) DEFAULT NULL COMMENT '银行名称',
    `amount` DECIMAL(10,2) NOT NULL COMMENT '提现金额',
    `status` TINYINT DEFAULT 0 COMMENT '状态: 0待审核 1通过 2拒绝',
    `remark` VARCHAR(255) DEFAULT NULL COMMENT '备注',
    `processed_at` DATETIME DEFAULT NULL COMMENT '处理时间',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_phone` (`phone`),
    INDEX `idx_status` (`status`),
    INDEX `idx_created_at` (`created_at`),
    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='提现表';

-- 收益明细表 (按月分区)
CREATE TABLE IF NOT EXISTS `income_logs` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT,
    `user_id` INT UNSIGNED NOT NULL COMMENT '用户ID',
    `phone` VARCHAR(20) NOT NULL COMMENT '手机号码',
    `name` VARCHAR(50) NOT NULL COMMENT '姓名',
    `source` VARCHAR(100) NOT NULL COMMENT '收益来源',
    `source_type` TINYINT NOT NULL COMMENT '来源类型: 1订单 2团队 3奖励',
    `amount` DECIMAL(10,2) NOT NULL COMMENT '收益金额',
    `related_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '关联订单/记录ID',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_phone` (`phone`),
    INDEX `idx_source_type` (`source_type`),
    INDEX `idx_created_at` (`created_at`),
    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='收益明细表';

-- 余额变动记录表 (按月分区)
CREATE TABLE IF NOT EXISTS `balance_logs` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT,
    `user_id` INT UNSIGNED NOT NULL COMMENT '用户ID',
    `type` TINYINT NOT NULL COMMENT '类型: 1充值 2提现 3收益 4扣除 5算力转余额',
    `amount` DECIMAL(15,2) NOT NULL COMMENT '变动金额',
    `before_balance` DECIMAL(15,2) NOT NULL COMMENT '变动前余额',
    `after_balance` DECIMAL(15,2) NOT NULL COMMENT '变动后余额',
    `remark` VARCHAR(255) DEFAULT NULL COMMENT '备注',
    `related_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '关联ID',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_type` (`type`),
    INDEX `idx_created_at` (`created_at`),
    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='余额变动记录表';

-- 管理员表
CREATE TABLE IF NOT EXISTS `admins` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `username` VARCHAR(50) NOT NULL UNIQUE COMMENT '管理员账号',
    `password` VARCHAR(255) NOT NULL COMMENT '密码哈希',
    `nickname` VARCHAR(50) DEFAULT NULL COMMENT '昵称',
    `last_login_at` DATETIME DEFAULT NULL COMMENT '最后登录时间',
    `status` TINYINT DEFAULT 1 COMMENT '状态: 1启用 0禁用',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员表';

-- 菜单表
CREATE TABLE IF NOT EXISTS `menus` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `parent_id` INT UNSIGNED DEFAULT 0 COMMENT '父菜单ID, 0为顶级',
    `name` VARCHAR(50) NOT NULL COMMENT '菜单名称',
    `path` VARCHAR(100) DEFAULT NULL COMMENT '路由路径',
    `component` VARCHAR(100) DEFAULT NULL COMMENT '组件路径',
    `permission` VARCHAR(100) DEFAULT NULL COMMENT '权限标识',
    `icon` VARCHAR(50) DEFAULT NULL COMMENT '图标',
    `type` TINYINT DEFAULT 1 COMMENT '类型: 1菜单 2按钮',
    `sort` INT UNSIGNED DEFAULT 0 COMMENT '排序',
    `status` TINYINT DEFAULT 1 COMMENT '状态: 1显示 0隐藏',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX `idx_parent_id` (`parent_id`),
    INDEX `idx_type` (`type`),
    INDEX `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='菜单表';

-- 角色表
CREATE TABLE IF NOT EXISTS `roles` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(50) NOT NULL UNIQUE COMMENT '角色名称',
    `code` VARCHAR(50) NOT NULL UNIQUE COMMENT '角色标识',
    `description` VARCHAR(255) DEFAULT NULL COMMENT '角色描述',
    `status` TINYINT DEFAULT 1 COMMENT '状态: 1启用 0禁用',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX `idx_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色表';

-- 角色-菜单关联表
CREATE TABLE IF NOT EXISTS `role_menus` (
    `role_id` INT UNSIGNED NOT NULL,
    `menu_id` INT UNSIGNED NOT NULL,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`role_id`, `menu_id`),
    FOREIGN KEY (`role_id`) REFERENCES `roles`(`id`) ON DELETE CASCADE,
    FOREIGN KEY (`menu_id`) REFERENCES `menus`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色菜单关联表';

-- 管理员-角色关联表
CREATE TABLE IF NOT EXISTS `admin_roles` (
    `admin_id` INT UNSIGNED NOT NULL,
    `role_id` INT UNSIGNED NOT NULL,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`admin_id`, `role_id`),
    FOREIGN KEY (`admin_id`) REFERENCES `admins`(`id`) ON DELETE CASCADE,
    FOREIGN KEY (`role_id`) REFERENCES `roles`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员角色关联表';

-- 支付渠道配置表
CREATE TABLE IF NOT EXISTS `payment_channels` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(50) NOT NULL COMMENT '渠道名称',
    `code` VARCHAR(30) NOT NULL UNIQUE COMMENT '渠道标识',
    `type` VARCHAR(20) NOT NULL COMMENT '类型: alipay/wechat/bank_card',
    `api_url` VARCHAR(255) NOT NULL COMMENT 'API接口地址',
    `api_key` VARCHAR(255) NOT NULL COMMENT 'API密钥',
    `api_secret` VARCHAR(255) DEFAULT NULL COMMENT 'API密钥加密串',
    `merchant_id` VARCHAR(100) DEFAULT NULL COMMENT '商户号',
    `app_id` VARCHAR(100) DEFAULT NULL COMMENT '应用ID',
    `notify_url` VARCHAR(255) DEFAULT NULL COMMENT '异步通知地址',
    `return_url` VARCHAR(255) DEFAULT NULL COMMENT '同步返回地址',
    `config` JSON DEFAULT NULL COMMENT '扩展配置',
    `min_amount` DECIMAL(10,2) DEFAULT 0.00 COMMENT '最小金额',
    `max_amount` DECIMAL(10,2) DEFAULT 999999.99 COMMENT '最大金额',
    `sort` INT UNSIGNED DEFAULT 0 COMMENT '排序',
    `is_default` TINYINT DEFAULT 0 COMMENT '默认渠道: 1是 0否',
    `status` TINYINT DEFAULT 1 COMMENT '状态: 1启用 0禁用',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX `idx_code` (`code`),
    INDEX `idx_type` (`type`),
    INDEX `idx_status` (`status`),
    INDEX `idx_is_default` (`is_default`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付渠道配置表';

-- 三方支付接口调用记录表 (按月分区)
CREATE TABLE IF NOT EXISTS `payment_api_logs` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT,
    `order_no` VARCHAR(50) NOT NULL COMMENT '业务订单号',
    `channel_id` INT UNSIGNED NOT NULL COMMENT '支付渠道ID',
    `channel_code` TINYINT NOT NULL COMMENT '渠道类型: 1支付宝 2微信 3银行卡',
    `action` TINYINT NOT NULL COMMENT '操作类型: 1支付 2查询 3退款 4回调',
    `request_url` VARCHAR(255) NOT NULL COMMENT '请求URL',
    `request_params` JSON COMMENT '请求参数',
    `request_headers` JSON COMMENT '请求头信息',
    `response_body` JSON COMMENT '响应内容',
    `response_code` VARCHAR(20) DEFAULT NULL COMMENT '三方返回状态码',
    `status` TINYINT DEFAULT 0 COMMENT '状态: 0处理中 1成功 2失败',
    `error_msg` VARCHAR(500) DEFAULT NULL COMMENT '错误信息',
    `retry_times` TINYINT UNSIGNED DEFAULT 0 COMMENT '重试次数',
    `request_time` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '请求时间',
    `response_time` DATETIME DEFAULT NULL COMMENT '响应时间',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    INDEX `idx_order_no` (`order_no`),
    INDEX `idx_channel_id` (`channel_id`),
    INDEX `idx_channel_code` (`channel_code`),
    INDEX `idx_action` (`action`),
    INDEX `idx_status` (`status`),
    INDEX `idx_request_time` (`request_time`),
    INDEX `idx_created_at` (`created_at`),
    FOREIGN KEY (`channel_id`) REFERENCES `payment_channels`(`id`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='三方支付接口调用记录表';

-- 任务表
CREATE TABLE IF NOT EXISTS `tasks` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(100) NOT NULL COMMENT '任务名称',
    `code` VARCHAR(50) NOT NULL UNIQUE COMMENT '任务代号',
    `work_points` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '工分',
    `enabled` TINYINT DEFAULT 1 COMMENT '是否启用: 1是 0否',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX `idx_code` (`code`),
    INDEX `idx_enabled` (`enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务表';

-- 系统配置表
CREATE TABLE IF NOT EXISTS `system_configs` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `key` VARCHAR(50) NOT NULL UNIQUE COMMENT '配置键',
    `value` VARCHAR(500) NOT NULL COMMENT '配置值',
    `description` VARCHAR(255) DEFAULT NULL COMMENT '配置说明',
    `group` VARCHAR(30) DEFAULT 'default' COMMENT '配置分组',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX `idx_key` (`key`),
    INDEX `idx_group` (`group`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统配置表';

-- 初始化配置: 限制提款最少工分
INSERT INTO `system_configs` (`key`, `value`, `description`, `group`) VALUES
('withdrawal.min_work_points', '100', '限制提款最少工分数', 'withdrawal')
ON DUPLICATE KEY UPDATE
`value` = VALUES(`value`),
`description` = VALUES(`description`),
`group` = VALUES(`group`);

-- 初始化配置: 邀请好友奖励工分
INSERT INTO `system_configs` (`key`, `value`, `description`, `group`) VALUES
('invite.reward_work_points', '10', '邀请好友奖励工分数', 'invite')
ON DUPLICATE KEY UPDATE
`value` = VALUES(`value`),
`description` = VALUES(`description`),
`group` = VALUES(`group`);

-- 初始化配置: 首充奖励工分
INSERT INTO `system_configs` (`key`, `value`, `description`, `group`) VALUES
('recharge.first_reward_work_points', '50', '首充奖励工分数', 'recharge')
ON DUPLICATE KEY UPDATE
`value` = VALUES(`value`),
`description` = VALUES(`description`),
`group` = VALUES(`group`);

-- 初始化配置: 单笔最大提取金额
INSERT INTO `system_configs` (`key`, `value`, `description`, `group`) VALUES
('withdrawal.max_amount', '5000', '单笔最大提取金额', 'withdrawal')
ON DUPLICATE KEY UPDATE
`value` = VALUES(`value`),
`description` = VALUES(`description`),
`group` = VALUES(`group`);

-- 分账规则表
CREATE TABLE IF NOT EXISTS `profit_sharing_rules` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(50) NOT NULL COMMENT '规则名称',
    `type` VARCHAR(30) NOT NULL COMMENT '分账类型: order订单/recharge充值/withdrawal提现/task任务',
    `level` TINYINT NOT NULL DEFAULT 0 COMMENT '分账层级: 0自己/1一级/2二级/3三级',
    `ratio` DECIMAL(5,4) NOT NULL COMMENT '分账比例 (0.0001-1.0000)',
    `fixed_amount` DECIMAL(10,2) DEFAULT 0.00 COMMENT '固定金额',
    `unit` VARCHAR(10) DEFAULT 'point' COMMENT '单位: point工分/amount金额',
    `min_amount` DECIMAL(10,2) DEFAULT 0.00 COMMENT '触发最低金额',
    `max_amount` DECIMAL(10,2) DEFAULT NULL COMMENT '触发最高金额',
    `sort` INT UNSIGNED DEFAULT 0 COMMENT '排序',
    `enabled` TINYINT DEFAULT 1 COMMENT '是否启用: 1启用 0禁用',
    `description` VARCHAR(255) DEFAULT NULL COMMENT '规则说明',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX `idx_type` (`type`),
    INDEX `idx_level` (`level`),
    INDEX `idx_enabled` (`enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分账规则表';

-- 初始化分账规则示例
INSERT INTO `profit_sharing_rules` (`name`, `type`, `level`, `ratio`, `fixed_amount`, `unit`, `description`)
SELECT '一级团队订单分润', 'order', 1, 0.1000, 0.00, 'point', '一级团队成员下单，上级获得10%工分'
WHERE NOT EXISTS (
    SELECT 1 FROM `profit_sharing_rules` WHERE `name` = '一级团队订单分润'
);

INSERT INTO `profit_sharing_rules` (`name`, `type`, `level`, `ratio`, `fixed_amount`, `unit`, `description`)
SELECT '二级团队订单分润', 'order', 2, 0.0500, 0.00, 'point', '二级团队成员下单，上上级获得5%工分'
WHERE NOT EXISTS (
    SELECT 1 FROM `profit_sharing_rules` WHERE `name` = '二级团队订单分润'
);

INSERT INTO `profit_sharing_rules` (`name`, `type`, `level`, `ratio`, `fixed_amount`, `unit`, `description`)
SELECT '一级团队充值分润', 'recharge', 1, 0.0500, 0.00, 'point', '一级团队成员充值，上级获得5%工分'
WHERE NOT EXISTS (
    SELECT 1 FROM `profit_sharing_rules` WHERE `name` = '一级团队充值分润'
);

-- 风控配置表
CREATE TABLE IF NOT EXISTS `risk_controls` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(50) NOT NULL COMMENT '规则名称',
    `code` VARCHAR(50) NOT NULL UNIQUE COMMENT '规则编码',
    `type` VARCHAR(30) NOT NULL COMMENT '风控类型: login登录/withdrawal提现/recharge充值/trade交易/ip_limitIP限制/device_limit设备限制',
    `level` VARCHAR(20) NOT NULL DEFAULT 'medium' COMMENT '风险等级: low低/medium中/high高/critical严重',
    `trigger_condition` TEXT NOT NULL COMMENT '触发条件配置(JSON字符串)',
    `action` VARCHAR(50) NOT NULL COMMENT '处置动作: block拦截/verify验证/sms通知/email通知/log记录',
    `limit_value` INT UNSIGNED DEFAULT NULL COMMENT '限制阈值',
    `time_window` INT UNSIGNED DEFAULT 3600 COMMENT '时间窗口(秒)',
    `enabled` TINYINT DEFAULT 1 COMMENT '是否启用: 1启用 0禁用',
    `description` VARCHAR(255) DEFAULT NULL COMMENT '规则描述',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX `idx_code` (`code`),
    INDEX `idx_type` (`type`),
    INDEX `idx_level` (`level`),
    INDEX `idx_enabled` (`enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='风控配置表';

-- 用户任务完成记录表
CREATE TABLE IF NOT EXISTS `user_tasks` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `user_id` INT UNSIGNED NOT NULL COMMENT '用户ID',
    `task_id` INT UNSIGNED NOT NULL COMMENT '任务ID',
    `completed_times` INT UNSIGNED DEFAULT 0 COMMENT '已完成次数',
    `last_completed_at` DATETIME DEFAULT NULL COMMENT '最后完成时间',
    `total_reward` DECIMAL(10,2) DEFAULT 0.00 COMMENT '累计获得奖励',
    `status` TINYINT DEFAULT 1 COMMENT '状态: 1进行中 2已完成 3已过期',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY `uk_user_task` (`user_id`, `task_id`),
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_task_id` (`task_id`),
    INDEX `idx_status` (`status`),
    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE,
    FOREIGN KEY (`task_id`) REFERENCES `tasks`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户任务完成记录表';

-- 会员权益表 (cron service 同步)
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

-- 用户每日任务完成记录表 (cron service 同步 - 每日重置)
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

-- 用户算力持有表
CREATE TABLE IF NOT EXISTS `user_hashrates` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `user_id` INT UNSIGNED NOT NULL COMMENT '用户ID',
    `machine_id` INT UNSIGNED NOT NULL COMMENT '矿机ID',
    `purchase_order_id` BIGINT UNSIGNED NOT NULL COMMENT '购买订单ID',
    `quantity` INT UNSIGNED NOT NULL COMMENT '购买数量',
    `total_hashrate` DECIMAL(15,2) NOT NULL COMMENT '总算力',
    `start_date` DATE NOT NULL COMMENT '生效日期',
    `end_date` DATE NOT NULL COMMENT '到期日期',
    `status` TINYINT DEFAULT 1 COMMENT '状态: 1生效中 2已过期 3已转让',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_machine_id` (`machine_id`),
    INDEX `idx_status` (`status`),
    INDEX `idx_end_date` (`end_date`),
    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE,
    FOREIGN KEY (`machine_id`) REFERENCES `mining_machines`(`id`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户算力持有表';

-- 用户登录日志表 (按月分区)
CREATE TABLE IF NOT EXISTS `login_logs` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT,
    `user_id` INT UNSIGNED COMMENT '用户ID(未登录时为NULL)',
    `phone` VARCHAR(20) COMMENT '登录手机号',
    `ip` VARCHAR(45) NOT NULL COMMENT 'IP地址',
    `device` VARCHAR(255) COMMENT '设备信息',
    `user_agent` VARCHAR(500) COMMENT '浏览器UA',
    `login_type` VARCHAR(20) DEFAULT 'password' COMMENT '登录方式: password/sms',
    `status` TINYINT DEFAULT 1 COMMENT '状态: 1成功 0失败',
    `fail_reason` VARCHAR(100) COMMENT '失败原因',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_ip` (`ip`),
    INDEX `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户登录日志表';

-- 后台管理员操作日志表
CREATE TABLE IF NOT EXISTS `admin_logs` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `admin_id` INT UNSIGNED NOT NULL COMMENT '管理员ID',
    `admin_username` VARCHAR(50) NOT NULL COMMENT '管理员账号',
    `action` VARCHAR(50) NOT NULL COMMENT '操作类型: login登录/create创建/update更新/delete删除/export导出/approve审核/reject拒绝',
    `module` VARCHAR(50) NOT NULL COMMENT '操作模块: user用户/order订单/product产品/withdrawal提现/recharge充值/task任务/config配置/risk风控',
    `description` VARCHAR(255) NOT NULL COMMENT '操作描述',
    `request_method` VARCHAR(10) COMMENT '请求方法: GET/POST/PUT/DELETE',
    `request_url` VARCHAR(255) COMMENT '请求URL',
    `request_params` TEXT COMMENT '请求参数(JSON字符串)',
    `response_code` INT COMMENT '响应状态码',
    `response_message` VARCHAR(255) COMMENT '响应消息',
    `ip` VARCHAR(45) NOT NULL COMMENT 'IP地址',
    `user_agent` VARCHAR(500) COMMENT '浏览器UA',
    `device` VARCHAR(255) COMMENT '设备信息',
    `target_id` INT UNSIGNED COMMENT '操作目标ID',
    `before_data` TEXT COMMENT '操作前数据(JSON字符串)',
    `after_data` TEXT COMMENT '操作后数据(JSON字符串)',
    `execution_time` INT UNSIGNED COMMENT '执行时间(毫秒)',
    `status` TINYINT DEFAULT 1 COMMENT '状态: 1成功 0失败',
    `fail_reason` VARCHAR(255) COMMENT '失败原因',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX `idx_admin_id` (`admin_id`),
    INDEX `idx_action` (`action`),
    INDEX `idx_module` (`module`),
    INDEX `idx_target_id` (`target_id`),
    INDEX `idx_status` (`status`),
    INDEX `idx_ip` (`ip`),
    INDEX `idx_created_at` (`created_at`),
    FOREIGN KEY (`admin_id`) REFERENCES `admins`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='后台管理员操作日志表';

-- 新闻公告表
CREATE TABLE IF NOT EXISTS `news` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `title` VARCHAR(200) NOT NULL COMMENT '新闻标题',
    `summary` VARCHAR(500) DEFAULT NULL COMMENT '新闻摘要',
    `content` TEXT NOT NULL COMMENT '新闻内容',
    `cover_image` VARCHAR(255) DEFAULT NULL COMMENT '封面图片',
    `category` VARCHAR(30) DEFAULT 'announcement' COMMENT '分类: announcement公告/activity活动/update更新/help帮助',
    `type` VARCHAR(20) DEFAULT 'normal' COMMENT '类型: normal普通/top置顶/hot热门',
    `author` VARCHAR(50) DEFAULT 'admin' COMMENT '发布者',
    `admin_id` INT UNSIGNED DEFAULT NULL COMMENT '发布管理员ID',
    `view_count` INT UNSIGNED DEFAULT 0 COMMENT '浏览次数',
    `sort` INT UNSIGNED DEFAULT 0 COMMENT '排序',
    `status` TINYINT DEFAULT 1 COMMENT '状态: 1发布 0草稿 2下架',
    `is_top` TINYINT DEFAULT 0 COMMENT '是否置顶: 1是 0否',
    `publish_time` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '发布时间',
    `expire_time` DATETIME DEFAULT NULL COMMENT '过期时间',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX `idx_category` (`category`),
    INDEX `idx_type` (`type`),
    INDEX `idx_status` (`status`),
    INDEX `idx_is_top` (`is_top`),
    INDEX `idx_publish_time` (`publish_time`),
    INDEX `idx_sort` (`sort`),
    FOREIGN KEY (`admin_id`) REFERENCES `admins`(`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='新闻公告表';

-- 多网域配置表
CREATE TABLE IF NOT EXISTS `domains` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `domain` VARCHAR(100) NOT NULL UNIQUE COMMENT '网域名称',
    `enabled` TINYINT DEFAULT 1 COMMENT '是否启用: 1启用 0禁用',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX `idx_domain` (`domain`),
    INDEX `idx_enabled` (`enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='多网域配置表';

-- IP白名单表
CREATE TABLE IF NOT EXISTS `whitelist_ips` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `ip` VARCHAR(45) NOT NULL COMMENT 'IP地址',
    `type` VARCHAR(20) DEFAULT 'admin' COMMENT '类型: admin后台/api接口/withdrawal提现',
    `description` VARCHAR(255) DEFAULT NULL COMMENT '备注说明',
    `enabled` TINYINT DEFAULT 1 COMMENT '是否启用: 1启用 0禁用',
    `expire_at` DATETIME DEFAULT NULL COMMENT '过期时间(NULL永久有效)',
    `created_by` INT UNSIGNED DEFAULT NULL COMMENT '创建者管理员ID',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY `uk_ip_type` (`ip`, `type`),
    INDEX `idx_ip` (`ip`),
    INDEX `idx_type` (`type`),
    INDEX `idx_enabled` (`enabled`),
    INDEX `idx_expire_at` (`expire_at`),
    FOREIGN KEY (`created_by`) REFERENCES `admins`(`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='IP白名单表';

-- 轮播图表
CREATE TABLE IF NOT EXISTS `banners` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `title` VARCHAR(100) DEFAULT NULL COMMENT '标题',
    `image` VARCHAR(255) NOT NULL COMMENT '图片地址',
    `link` VARCHAR(255) DEFAULT NULL COMMENT '跳转链接',
    `type` VARCHAR(20) DEFAULT 'pc' COMMENT '类型: pc电脑/h5手机/app应用',
    `position` VARCHAR(30) DEFAULT 'home' COMMENT '位置: home首页/user个人中心',
    `sort` INT UNSIGNED DEFAULT 0 COMMENT '排序',
    `status` TINYINT DEFAULT 1 COMMENT '状态: 1显示 0隐藏',
    `start_time` DATETIME DEFAULT NULL COMMENT '开始时间',
    `end_time` DATETIME DEFAULT NULL COMMENT '结束时间',
    `click_count` INT UNSIGNED DEFAULT 0 COMMENT '点击次数',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX `idx_type` (`type`),
    INDEX `idx_position` (`position`),
    INDEX `idx_status` (`status`),
    INDEX `idx_sort` (`sort`),
    INDEX `idx_start_time` (`start_time`),
    INDEX `idx_end_time` (`end_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='轮播图表';

-- 签到表 (按月分区)
CREATE TABLE IF NOT EXISTS `check_ins` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT,
    `user_id` INT UNSIGNED NOT NULL COMMENT '用户ID',
    `check_in_date` DATE NOT NULL COMMENT '签到日期',
    `consecutive_days` INT UNSIGNED DEFAULT 1 COMMENT '连续签到天数',
    `reward_points` DECIMAL(10,2) DEFAULT 0.00 COMMENT '奖励工分',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_date` (`user_id`, `check_in_date`),
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_check_in_date` (`check_in_date`),
    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='签到表';

-- SMS 短信发送记录表
CREATE TABLE IF NOT EXISTS `sms_logs` (
    `id` VARCHAR(64) NOT NULL COMMENT '记录ID',
    `phone` VARCHAR(20) NOT NULL COMMENT '手机号',
    `template_code` VARCHAR(64) DEFAULT NULL COMMENT '短信模板代码',
    `template_params` TEXT COMMENT '模板参数(JSON)',
    `content` TEXT COMMENT '短信内容',
    `provider_id` VARCHAR(64) NOT NULL COMMENT '服务商ID',
    `provider_name` VARCHAR(128) DEFAULT NULL COMMENT '服务商名称',
    `status` VARCHAR(20) NOT NULL COMMENT '状态: pending, sent, failed, delivered',
    `provider_message_id` VARCHAR(128) DEFAULT NULL COMMENT '服务商返回的消息ID',
    `error_message` TEXT COMMENT '错误信息',
    `retry_count` INT DEFAULT 0 COMMENT '重试次数',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    INDEX `idx_phone` (`phone`),
    INDEX `idx_status` (`status`),
    INDEX `idx_provider_id` (`provider_id`),
    INDEX `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='短信发送记录表';

-- 算力补偿记录表（由定时任务扫描并恢复算力）
CREATE TABLE IF NOT EXISTS `hashrate_compensation_records` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `user_id` INT UNSIGNED NOT NULL COMMENT '用户ID',
    `amount` DECIMAL(15,2) NOT NULL COMMENT '需恢复的算力金额',
    `request_id` VARCHAR(100) NOT NULL COMMENT '唯一请求ID',
    `reason` VARCHAR(255) DEFAULT NULL COMMENT '补偿原因',
    `status` TINYINT DEFAULT 0 COMMENT '状态: 0待补偿 1已补偿 2补偿失败',
    `retry_times` INT UNSIGNED DEFAULT 0 COMMENT '重试次数',
    `compensated_at` DATETIME DEFAULT NULL COMMENT '补偿时间',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY `uk_request_id` (`request_id`),
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_status` (`status`),
    INDEX `idx_created_at` (`created_at`),
    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='算力补偿记录表';

-- 监控告警日志表
CREATE TABLE IF NOT EXISTS `alert_logs` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `level` VARCHAR(20) NOT NULL COMMENT '告警级别: warning/error/critical',
    `module` VARCHAR(50) NOT NULL COMMENT '模块',
    `message` TEXT NOT NULL COMMENT '告警内容',
    `detail` TEXT DEFAULT NULL COMMENT '详细信息',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX `idx_level` (`level`),
    INDEX `idx_module` (`module`),
    INDEX `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='监控告警日志表';
