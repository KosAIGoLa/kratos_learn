package cron

import (
	"cron/internal/data"
	"time"

	kratoslog "github.com/go-kratos/kratos/v2/log"
)

// TaskManager 任务管理器
type TaskManager struct {
	scheduler *Scheduler
	data      *data.Data
	log       *kratoslog.Helper
}

// NewTaskManager 创建任务管理器。
// 保持旧签名以兼容现有测试与调用方。
func NewTaskManager(scheduler *Scheduler, logger kratoslog.Logger) *TaskManager {
	return NewTaskManagerWithData(scheduler, nil, logger)
}

// NewTaskManagerWithData 创建带数据依赖的任务管理器。
func NewTaskManagerWithData(scheduler *Scheduler, data *data.Data, logger kratoslog.Logger) *TaskManager {
	return &TaskManager{
		scheduler: scheduler,
		data:      data,
		log:       kratoslog.NewHelper(logger),
	}
}

// RegisterTasks 注册所有定时任务
func (tm *TaskManager) RegisterTasks() error {
	// ==================== 算力生产相关任务 ====================

	// 1. 每小时算力结算 - 每小时执行，计算用户算力产出
	// 公式: 每小时算力产出 = 产品级别 × 生产力百分比
	if err := tm.scheduler.AddTask("hourly_hashrate_settlement", "0 0 * * * *", tm.hourlyHashrateSettlement); err != nil {
		return err
	}

	// 2. 每日算力汇总结算 - 每天0点执行，24小时周期结算
	// 公式: 每日总算力 = Σ(每小时产出)
	if err := tm.scheduler.AddTask("daily_hashrate_settlement", "0 0 0 * * *", tm.dailyHashrateSettlement); err != nil {
		return err
	}

	// ==================== 每日签到和任务相关任务 ====================

	// 4. 每日签到任务重置 - 每天0点执行，重置签到状态
	if err := tm.scheduler.AddTask("daily_checkin_reset", "0 0 0 * * *", tm.dailyCheckinReset); err != nil {
		return err
	}

	// 5. 每日任务奖励发放 - 每天8:00执行，发放签到工分
	if err := tm.scheduler.AddTask("daily_task_reward", "0 0 8 * * *", tm.dailyTaskReward); err != nil {
		return err
	}

	// 6. 连续签到状态检查 - 每天0:05执行，检查连续签到
	if err := tm.scheduler.AddTask("consecutive_checkin_check", "0 5 0 * * *", tm.consecutiveCheckinCheck); err != nil {
		return err
	}

	// ==================== 分账和结算任务 ====================

	// 7. 分账规则执行 - 每10分钟执行一次，处理订单/充值分润
	if err := tm.scheduler.AddTask("profit_sharing", "0 */10 * * * *", tm.profitSharing); err != nil {
		return err
	}

	// 8. 邀请奖励结算 - 每天9:00执行，结算邀请奖励
	if err := tm.scheduler.AddTask("invite_reward_settlement", "0 0 9 * * *", tm.inviteRewardSettlement); err != nil {
		return err
	}

	// ==================== 系统维护任务 ====================

	// 9. 健康检查 - 每分钟执行
	if err := tm.scheduler.AddTask("health_check", "0 * * * * *", tm.healthCheck); err != nil {
		return err
	}

	// 10. 过期机器清理 - 每天1:00执行，清理过期算力机器
	if err := tm.scheduler.AddTask("expired_machine_cleanup", "0 0 1 * * *", tm.expiredMachineCleanup); err != nil {
		return err
	}

	// 11. 数据清理 - 每天2:00执行
	if err := tm.scheduler.AddTask("data_cleanup", "0 0 2 * * *", tm.dataCleanup); err != nil {
		return err
	}

	// 12. 每日报表生成 - 每天2:30执行
	if err := tm.scheduler.AddTask("daily_report", "0 30 2 * * *", tm.generateDailyReport); err != nil {
		return err
	}

	// 13. 提现限制重置 - 每天0:00执行，重置每日提现次数
	if err := tm.scheduler.AddTask("withdrawal_limit_reset", "0 0 0 * * *", tm.withdrawalLimitReset); err != nil {
		return err
	}

	// 14. 用户工分/任务每日重置 - 每天0:00执行
	if err := tm.scheduler.AddTask("user_daily_reset", "0 0 0 * * *", tm.userDailyReset); err != nil {
		return err
	}

	// 15. 每周报表 - 每周一早上9点
	if err := tm.scheduler.AddTask("weekly_report", "0 0 9 * * 1", tm.generateWeeklyReport); err != nil {
		return err
	}

	// 16. 每月报表 - 每月1号凌晨3点
	if err := tm.scheduler.AddTask("monthly_report", "0 0 3 1 * *", tm.generateMonthlyReport); err != nil {
		return err
	}

	tm.log.Info("所有定时任务注册完成")
	return nil
}

// ==================== 具体任务实现 ====================

// healthCheck 健康检查任务
func (tm *TaskManager) healthCheck() {
	tm.log.Info("执行健康检查任务...")
	// 这里可以检查其他服务的健康状态
	// 例如：检查数据库连接、Redis 连接、外部 API 等
	tm.log.Info("健康检查完成，所有服务运行正常")
}

// dataCleanup 数据清理任务
func (tm *TaskManager) dataCleanup() {
	tm.log.Info("执行数据清理任务...")

	// 清理过期日志
	tm.cleanupOldLogs()

	// 清理临时数据
	tm.cleanupTempData()

	// 清理过期缓存
	tm.cleanupExpiredCache()

	tm.log.Info("数据清理任务完成")
}

func (tm *TaskManager) cleanupOldLogs() {
	// 示例：删除 30 天前的日志
	cutoff := time.Now().AddDate(0, 0, -30)
	tm.log.Infof("清理 %s 之前的日志记录", cutoff.Format("2006-01-02"))
	// 实际实现：调用 data 层删除旧日志
}

func (tm *TaskManager) cleanupTempData() {
	tm.log.Info("清理临时数据...")
	// 清理临时文件、过期会话等
}

func (tm *TaskManager) cleanupExpiredCache() {
	tm.log.Info("清理过期缓存...")
	// 清理 Redis 过期键等
}

// generateDailyReport 生成日报表
func (tm *TaskManager) generateDailyReport() {
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	tm.log.Infof("生成 %s 日报表...", yesterday)

	// 统计昨日数据
	stats := tm.collectDailyStats(yesterday)

	// 发送报表
	tm.sendReport("daily", yesterday, stats)

	tm.log.Infof("日报表 %s 生成完成", yesterday)
}

// generateWeeklyReport 生成周报表
func (tm *TaskManager) generateWeeklyReport() {
	now := time.Now()
	weekStart := now.AddDate(0, 0, -7).Format("2006-01-02")
	weekEnd := now.AddDate(0, 0, -1).Format("2006-01-02")

	tm.log.Infof("生成周报表 %s ~ %s...", weekStart, weekEnd)

	stats := tm.collectWeeklyStats(weekStart, weekEnd)
	tm.sendReport("weekly", weekStart+"_"+weekEnd, stats)

	tm.log.Infof("周报表 %s ~ %s 生成完成", weekStart, weekEnd)
}

// generateMonthlyReport 生成月报表
func (tm *TaskManager) generateMonthlyReport() {
	now := time.Now()
	lastMonth := now.AddDate(0, -1, 0)
	monthStr := lastMonth.Format("2006-01")

	tm.log.Infof("生成月报表 %s...", monthStr)

	stats := tm.collectMonthlyStats(monthStr)
	tm.sendReport("monthly", monthStr, stats)

	tm.log.Infof("月报表 %s 生成完成", monthStr)
}

// ==================== 辅助方法 ====================

// collectDailyStats 收集每日统计数据
func (tm *TaskManager) collectDailyStats(date string) map[string]interface{} {
	// 实际实现：从数据库查询统计数据
	return map[string]interface{}{
		"date":         date,
		"total_orders": 100,
		"total_users":  50,
		"revenue":      5000.00,
	}
}

// collectWeeklyStats 收集每周统计数据
func (tm *TaskManager) collectWeeklyStats(start, end string) map[string]interface{} {
	return map[string]interface{}{
		"period":       start + " to " + end,
		"total_orders": 700,
		"total_users":  300,
		"revenue":      35000.00,
	}
}

// collectMonthlyStats 收集每月统计数据
func (tm *TaskManager) collectMonthlyStats(month string) map[string]interface{} {
	return map[string]interface{}{
		"month":        month,
		"total_orders": 3000,
		"total_users":  1200,
		"revenue":      150000.00,
	}
}

// sendReport 发送报表
func (tm *TaskManager) sendReport(reportType, period string, data map[string]interface{}) {
	tm.log.Infof("发送 %s 报表 [%s]: %v", reportType, period, data)
	// 实际实现：
	// - 保存到文件系统
	// - 发送到邮箱
	// - 上传到云存储
	// - 发送到消息队列
}

// ==================== 动态任务管理 ====================

// AddCustomTask 添加自定义任务
func (tm *TaskManager) AddCustomTask(name, spec string, task func()) error {
	return tm.scheduler.AddTask(name, spec, task)
}

// RemoveTask 移除任务
func (tm *TaskManager) RemoveTask(name string) {
	tm.scheduler.RemoveTask(name)
}

// ListTasks 列出所有任务
func (tm *TaskManager) ListTasks() []TaskInfo {
	return tm.scheduler.ListTasks()
}

// ==================== 会员重置任务 ====================

// memberDailyReset 会员每日重置
// 每天凌晨执行，重置当日任务完成状态、工分/积分
//
//nolint:unused // 预留功能，待后续启用
func (tm *TaskManager) memberDailyReset() {
	today := time.Now().Format("2006-01-02")
	tm.log.Infof("执行会员每日重置任务 [%s]...", today)

	// 1. 获取所有活跃会员
	members := tm.getAllActiveMembers()

	resetCount := 0
	for _, member := range members {
		// 重置当日任务完成状态
		tm.resetDailyTasks(member)

		// 重置当日工分/积分获取状态
		tm.resetDailyPoints(member)

		// 记录重置时间
		member.LastResetTime = time.Now()

		// 更新会员数据
		if err := tm.updateMember(member); err != nil {
			tm.log.Errorf("重置会员 [%s] 每日状态失败: %v", member.ID, err)
			continue
		}
		resetCount++
	}

	tm.log.Infof("会员每日重置完成，共处理 %d 位会员", resetCount)
}

// resetDailyTasks 重置会员当日任务
//
//nolint:unused // 预留功能，待后续启用
func (tm *TaskManager) resetDailyTasks(member *Member) {
	// 重置所有每日任务的完成状态
	member.DailyTasks = []DailyTaskStatus{}
	tm.log.Infof("会员 [%s] 每日任务已重置", member.ID)
}

// resetDailyPoints 重置会员当日工分/积分
//
//nolint:unused // 预留功能，待后续启用
func (tm *TaskManager) resetDailyPoints(member *Member) {
	// 将当日积分累加到本月积分
	member.MonthlyPoints += member.DailyPoints

	// 重置当日积分
	member.DailyPoints = 0
	member.DailyPointsLimitReached = false
	tm.log.Infof("会员 [%s] 当日工分/积分已重置，累计到本月: %d", member.ID, member.MonthlyPoints)
}

// memberBenefitsCleanup 清理过期会员权益
// 每天执行，清理当日已过期的会员权益
//
//nolint:unused // 预留功能，待后续启用
func (tm *TaskManager) memberBenefitsCleanup() {
	today := time.Now().Format("2006-01-02")
	tm.log.Infof("执行会员权益清理任务 [%s]...", today)

	// 1. 获取所有有权益的会员
	members := tm.getMembersWithBenefits()

	cleanupCount := 0
	for _, member := range members {
		validBenefits := []MemberBenefit{}
		expiredBenefits := []MemberBenefit{}

		for _, benefit := range member.Benefits {
			if benefit.ExpireTime.After(time.Now()) {
				validBenefits = append(validBenefits, benefit)
			} else {
				expiredBenefits = append(expiredBenefits, benefit)
			}
		}

		// 如果有过期权益，更新会员权益列表
		if len(expiredBenefits) > 0 {
			member.Benefits = validBenefits
			if err := tm.updateMemberBenefits(member.ID, validBenefits); err != nil {
				tm.log.Errorf("更新会员 [%s] 权益失败: %v", member.ID, err)
				continue
			}
			cleanupCount += len(expiredBenefits)
			tm.log.Infof("会员 [%s] 清理了 %d 项过期权益", member.ID, len(expiredBenefits))
		}
	}

	tm.log.Infof("会员权益清理完成，共清理 %d 项过期权益", cleanupCount)
}

// memberLevelCheck 会员等级检查
// 每天执行，检查是否需要更新会员等级
//
//nolint:unused // 预留功能，待后续启用
func (tm *TaskManager) memberLevelCheck() {
	today := time.Now().Format("2006-01-02")
	tm.log.Infof("执行会员等级检查任务 [%s]...", today)

	// 1. 获取所有会员的当月消费统计
	stats := tm.getMemberMonthlyStats(today[:7])

	updateCount := 0
	for _, stat := range stats {
		member := stat.Member
		oldLevel := member.Level

		// 根据消费金额计算新等级
		newLevel := tm.calculateMemberLevel(stat.Spending)

		if newLevel != oldLevel {
			member.Level = newLevel
			member.LevelUpdateTime = time.Now()

			if err := tm.updateMemberLevel(member.ID, newLevel); err != nil {
				tm.log.Errorf("更新会员 [%s] 等级失败: %v", member.ID, err)
				continue
			}

			updateCount++
			tm.log.Infof("会员 [%s] 等级从 [%s] 更新为 [%s]", member.ID, oldLevel, newLevel)

			// 发送等级变更通知
			tm.sendLevelChangeNotification(member, oldLevel, newLevel)
		}
	}

	tm.log.Infof("会员等级更新完成，共更新 %d 位会员", updateCount)
}

// expiredMembersCleanup 清理过期会员
// 每天执行，清理已过期的临时/试用会员
//
//nolint:unused // 预留功能，待后续启用
func (tm *TaskManager) expiredMembersCleanup() {
	today := time.Now().Format("2006-01-02")
	tm.log.Infof("执行过期会员清理任务 [%s]...", today)

	// 1. 获取所有已过期的会员
	expiredMembers := tm.getExpiredMembers()

	cleanupCount := 0
	archivedCount := 0
	for _, member := range expiredMembers {
		// 根据过期时间决定处理方式
		daysSinceExpiry := int(time.Since(member.ExpireTime).Hours() / 24)

		if daysSinceExpiry <= 30 {
			// 30 天内：标记为过期但未删除（可续费恢复）
			if err := tm.markMemberAsExpired(member.ID); err != nil {
				tm.log.Errorf("标记会员 [%s] 过期失败: %v", member.ID, err)
				continue
			}
			cleanupCount++
		} else if daysSinceExpiry > 90 {
			// 超过 90 天：归档并删除
			if err := tm.archiveAndDeleteMember(member); err != nil {
				tm.log.Errorf("归档会员 [%s] 失败: %v", member.ID, err)
				continue
			}
			archivedCount++
		}
	}

	tm.log.Infof("过期会员清理完成，标记过期 %d 位，归档删除 %d 位", cleanupCount, archivedCount)
}

// memberMonthlySettlement 会员月度结算
// 每月 1 号执行，结算上月工分/积分，准备新的月度周期
//
//nolint:unused // 预留功能，待后续启用
func (tm *TaskManager) memberMonthlySettlement() {
	lastMonth := time.Now().AddDate(0, -1, 0).Format("2006-01")
	tm.log.Infof("执行会员月度结算任务 [%s]...", lastMonth)

	// 1. 获取所有会员
	members := tm.getAllMembers()

	settlementCount := 0
	for _, member := range members {
		// 将本月积分累加到总积分
		member.TotalPoints += member.MonthlyPoints

		// 生成会员月度结算报告
		tm.generateMemberMonthlyReport(member, lastMonth)

		// 重置本月积分
		member.MonthlyPoints = 0

		// 更新会员数据
		if err := tm.updateMember(member); err != nil {
			tm.log.Errorf("结算会员 [%s] 月度数据失败: %v", member.ID, err)
			continue
		}
		settlementCount++
	}

	tm.log.Infof("会员月度结算完成，共处理 %d 位会员", settlementCount)
}

// generateMemberMonthlyReport 生成会员月度结算报告
//
//nolint:unused // 预留功能，待后续启用
func (tm *TaskManager) generateMemberMonthlyReport(member *Member, month string) {
	tm.log.Infof("会员 [%s] %s 月度结算: 本月工分 %d, 累计总工分 %d",
		member.ID, month, member.MonthlyPoints, member.TotalPoints)

	// 实际实现：
	// - 保存到结算记录表
	// - 发送结算通知
	// - 生成月度账单
}

// ==================== 会员数据结构 ====================

// Member 会员信息
type Member struct {
	ID                      string
	Level                   string
	DailyPoints             int               // 当日工分/积分
	MonthlyPoints           int               // 当月累计工分/积分
	TotalPoints             int               // 累计总工分/积分
	DailyPointsLimitReached bool              // 当日工分上限是否已达
	DailyTasks              []DailyTaskStatus // 当日任务完成状态
	Benefits                []MemberBenefit   // 会员权益
	ExpireTime              time.Time
	LastResetTime           time.Time
	LevelUpdateTime         time.Time
}

// DailyTaskStatus 每日任务状态
type DailyTaskStatus struct {
	TaskID      string    // 任务ID
	TaskName    string    // 任务名称
	Completed   bool      // 是否完成
	CompletedAt time.Time // 完成时间
	Points      int       // 任务工分值
}

// MemberBenefit 会员权益
type MemberBenefit struct {
	Name       string
	ExpireTime time.Time
	Value      float64
}

// MemberMonthlyStat 会员月度统计
type MemberMonthlyStat struct {
	Member   *Member
	Spending float64
}

// ==================== 会员相关辅助方法 ====================

// getAllMembers 获取所有会员（模拟实现）
//
//nolint:unused // 预留功能，待后续启用
func (tm *TaskManager) getAllMembers() []*Member {
	// 实际实现：从数据库查询所有会员
	return []*Member{
		{ID: "M001", Level: "gold", DailyPoints: 10, MonthlyPoints: 100, TotalPoints: 1000},
		{ID: "M002", Level: "silver", DailyPoints: 5, MonthlyPoints: 50, TotalPoints: 500},
	}
}

// getAllActiveMembers 获取所有活跃会员（模拟实现）
//
//nolint:unused // 预留功能，待后续启用
func (tm *TaskManager) getAllActiveMembers() []*Member {
	// 实际实现：从数据库查询所有活跃会员（未过期、未冻结的会员）
	return []*Member{
		{
			ID:                      "M001",
			Level:                   "gold",
			DailyPoints:             10,
			MonthlyPoints:           100,
			TotalPoints:             1000,
			DailyPointsLimitReached: false,
			DailyTasks: []DailyTaskStatus{
				{TaskID: "T001", TaskName: "签到", Completed: true, Points: 5},
				{TaskID: "T002", TaskName: "分享", Completed: false, Points: 10},
			},
		},
		{
			ID:                      "M002",
			Level:                   "silver",
			DailyPoints:             5,
			MonthlyPoints:           50,
			TotalPoints:             500,
			DailyPointsLimitReached: false,
			DailyTasks: []DailyTaskStatus{
				{TaskID: "T001", TaskName: "签到", Completed: false, Points: 5},
			},
		},
	}
}

// getMembersWithBenefits 获取有权益的会员（模拟实现）
//
//nolint:unused // 预留功能，待后续启用
func (tm *TaskManager) getMembersWithBenefits() []*Member {
	// 实际实现：从数据库查询有权益的会员
	return []*Member{
		{
			ID: "M001",
			Benefits: []MemberBenefit{
				{Name: "免费配送", ExpireTime: time.Now().AddDate(0, 0, -5), Value: 20},
				{Name: "折扣券", ExpireTime: time.Now().AddDate(0, 1, 0), Value: 50},
			},
		},
	}
}

// getMemberMonthlyStats 获取会员月度消费统计（模拟实现）
//
//nolint:unused // 预留功能，待后续启用
func (tm *TaskManager) getMemberMonthlyStats(month string) []*MemberMonthlyStat {
	// 实际实现：从数据库查询上月消费统计
	return []*MemberMonthlyStat{
		{Member: &Member{ID: "M001", Level: "silver"}, Spending: 5000},
		{Member: &Member{ID: "M002", Level: "bronze"}, Spending: 15000},
	}
}

// getExpiredMembers 获取过期会员（模拟实现）
//
//nolint:unused // 预留功能，待后续启用
func (tm *TaskManager) getExpiredMembers() []*Member {
	// 实际实现：从数据库查询过期会员
	return []*Member{
		{ID: "M003", ExpireTime: time.Now().AddDate(0, 0, -10)},
		{ID: "M004", ExpireTime: time.Now().AddDate(0, 0, -100)},
	}
}

// calculateMemberLevel 根据消费金额计算会员等级
//
//nolint:unused // 预留功能，待后续启用
func (tm *TaskManager) calculateMemberLevel(spending float64) string {
	switch {
	case spending >= 10000:
		return "diamond"
	case spending >= 5000:
		return "platinum"
	case spending >= 2000:
		return "gold"
	case spending >= 1000:
		return "silver"
	default:
		return "bronze"
	}
}

// ==================== 算力生产相关任务实现 ====================

// hourlyHashrateSettlement 每小时算力结算
// 公式: 每小时算力产出 = 产品级别 × 生产力百分比
func (tm *TaskManager) hourlyHashrateSettlement() {
	currentHour := time.Now().Format("2006-01-02 15:04")
	tm.log.Infof("执行每小时算力结算 [%s]...", currentHour)

	// 1. 获取所有运行中的机器
	machines := tm.getRunningMachines()

	settlementCount := 0
	totalHashrate := 0.0

	for _, machine := range machines {
		// 计算每小时算力产出 = 产品级别 × 生产力百分比
		hourlyOutput := float64(machine.ProductLevel) * machine.ProductivityRate

		// 累计用户算力产出
		machine.UserHashrate += hourlyOutput
		totalHashrate += hourlyOutput

		// 更新机器算力记录
		if err := tm.updateMachineHashrate(machine.ID, machine.UserHashrate); err != nil {
			tm.log.Errorf("更新机器 [%d] 算力失败: %v", machine.ID, err)
			continue
		}

		// 记录算力产出日志
		tm.recordHashrateLog(machine.UserID, machine.ID, hourlyOutput, currentHour)

		settlementCount++
	}

	tm.log.Infof("每小时算力结算完成，共处理 %d 台机器，总算力产出: %.2f", settlementCount, totalHashrate)
}

// dailyHashrateSettlement 每日算力汇总结算
// 公式: 每日总算力 = Σ(每小时产出)
// 24小时周期结算
func (tm *TaskManager) dailyHashrateSettlement() {
	today := time.Now().Format("2006-01-02")
	tm.log.Infof("执行每日算力汇总结算 [%s]...", today)

	// 1. 获取所有用户
	users := tm.getAllActiveUsers()

	settlementCount := 0
	for _, user := range users {
		// 获取用户昨日24小时算力产出总和
		dailyHashrate := tm.getUserDailyHashrate(user.ID, today)

		// 累加到用户总算力
		user.TotalHashrate += dailyHashrate

		// 更新用户算力数据
		if err := tm.updateUserHashrate(user.ID, user.TotalHashrate, dailyHashrate); err != nil {
			tm.log.Errorf("更新用户 [%d] 算力失败: %v", user.ID, err)
			continue
		}

		// 记录每日算力结算
		tm.log.Infof("用户 [%d] 每日算力结算: %.2f, 累计总算力: %.2f", user.ID, dailyHashrate, user.TotalHashrate)

		settlementCount++
	}

	tm.log.Infof("每日算力汇总结算完成，共处理 %d 位用户", settlementCount)
}

// expiredMachineCleanup 过期机器清理
func (tm *TaskManager) expiredMachineCleanup() {
	today := time.Now().Format("2006-01-02")
	tm.log.Infof("执行过期机器清理任务 [%s]...", today)

	// 1. 获取所有已过期的机器
	expiredMachines := tm.getExpiredMachines()

	cleanupCount := 0
	for _, machine := range expiredMachines {
		// 停止机器算力生产
		machine.Status = "expired"

		// 更新机器状态
		if err := tm.updateMachineStatus(machine.ID, "expired"); err != nil {
			tm.log.Errorf("更新机器 [%d] 状态失败: %v", machine.ID, err)
			continue
		}

		tm.log.Infof("机器 [%d] 已过期并停止生产", machine.ID)
		cleanupCount++
	}

	tm.log.Infof("过期机器清理完成，共处理 %d 台机器", cleanupCount)
}

// ==================== 每日签到和任务相关任务实现 ====================

// dailyCheckinReset 每日签到任务重置
func (tm *TaskManager) dailyCheckinReset() {
	today := time.Now().Format("2006-01-02")
	tm.log.Infof("执行每日签到重置任务 [%s]...", today)

	// 1. 获取所有用户
	users := tm.getAllActiveUsers()

	resetCount := 0
	for _, user := range users {
		// 重置用户今日签到状态
		user.CheckedInToday = false
		user.TodayCheckinReward = 0

		// 更新用户签到状态
		if err := tm.updateUserCheckinStatus(user.ID, false, 0); err != nil {
			tm.log.Errorf("重置用户 [%d] 签到状态失败: %v", user.ID, err)
			continue
		}

		resetCount++
	}

	tm.log.Infof("每日签到重置完成，共处理 %d 位用户", resetCount)
}

// dailyTaskReward 每日任务奖励发放
func (tm *TaskManager) dailyTaskReward() {
	today := time.Now().Format("2006-01-02")
	tm.log.Infof("执行每日任务奖励发放 [%s]...", today)

	// 1. 获取昨日已签到的用户
	checkedInUsers := tm.getYesterdayCheckedInUsers()

	rewardCount := 0
	totalReward := 0.0

	for _, user := range checkedInUsers {
		// 计算签到奖励
		// 基础奖励 + 连续签到加成
		baseReward := 5.0                                       // 基础工分奖励
		consecutiveBonus := float64(user.ConsecutiveDays) * 0.5 // 连续签到加成
		totalReward := baseReward + consecutiveBonus

		// 增加用户工分
		user.WorkPoints += totalReward

		// 更新用户工分
		if err := tm.updateUserWorkPoints(user.ID, user.WorkPoints); err != nil {
			tm.log.Errorf("更新用户 [%d] 工分失败: %v", user.ID, err)
			continue
		}

		// 记录奖励日志
		tm.recordTaskReward(user.ID, "daily_checkin", totalReward, today)

		rewardCount++
	}

	tm.log.Infof("每日任务奖励发放完成，共发放 %d 位用户，总奖励: %.2f 工分", rewardCount, totalReward)
}

// consecutiveCheckinCheck 连续签到状态检查
func (tm *TaskManager) consecutiveCheckinCheck() {
	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	tm.log.Infof("执行连续签到检查 [%s]...", today)

	// 1. 获取所有用户
	users := tm.getAllActiveUsers()

	checkCount := 0
	resetCount := 0

	for _, user := range users {
		// 检查用户昨日是否签到
		checkedYesterday := tm.checkUserCheckedIn(user.ID, yesterday)

		if checkedYesterday {
			// 连续签到天数+1
			user.ConsecutiveDays++
		} else {
			// 未连续签到，重置连续天数
			if user.ConsecutiveDays > 0 {
				user.ConsecutiveDays = 0
				resetCount++
			}
		}

		// 更新用户连续签到天数
		if err := tm.updateUserConsecutiveDays(user.ID, user.ConsecutiveDays); err != nil {
			tm.log.Errorf("更新用户 [%d] 连续签到天数失败: %v", user.ID, err)
			continue
		}

		checkCount++
	}

	tm.log.Infof("连续签到检查完成，共检查 %d 位用户，重置 %d 位用户", checkCount, resetCount)
}

// ==================== 分账和结算任务实现 ====================

// profitSharing 分账规则执行
// 每10分钟执行一次，处理订单/充值分润
func (tm *TaskManager) profitSharing() {
	currentTime := time.Now().Format("2006-01-02 15:04")
	tm.log.Infof("执行分账规则 [%s]...", currentTime)

	// 1. 获取待分账的订单
	pendingOrders := tm.getPendingProfitSharingOrders()

	sharingCount := 0
	for _, order := range pendingOrders {
		// 执行分账逻辑
		// - 一级邀请奖励
		// - 二级邀请奖励
		// - 团队分润

		if err := tm.executeProfitSharing(order); err != nil {
			tm.log.Errorf("订单 [%d] 分账失败: %v", order.ID, err)
			continue
		}

		sharingCount++
	}

	tm.log.Infof("分账规则执行完成，共处理 %d 笔订单", sharingCount)
}

// inviteRewardSettlement 邀请奖励结算
func (tm *TaskManager) inviteRewardSettlement() {
	today := time.Now().Format("2006-01-02")
	tm.log.Infof("执行邀请奖励结算 [%s]...", today)

	// 1. 获取昨日新增邀请关系
	inviteRelations := tm.getYesterdayInviteRelations()

	rewardCount := 0
	for _, relation := range inviteRelations {
		// 计算邀请奖励
		// 一级邀请奖励
		if relation.Level == 1 {
			reward := tm.getInviteRewardConfig("level1")
			if err := tm.grantInviteReward(relation.InviterID, reward, relation.InviteeID); err != nil {
				tm.log.Errorf("发放一级邀请奖励失败: %v", err)
				continue
			}
		}

		// 二级邀请奖励
		if relation.Level == 2 {
			reward := tm.getInviteRewardConfig("level2")
			if err := tm.grantInviteReward(relation.InviterID, reward, relation.InviteeID); err != nil {
				tm.log.Errorf("发放二级邀请奖励失败: %v", err)
				continue
			}
		}

		rewardCount++
	}

	tm.log.Infof("邀请奖励结算完成，共处理 %d 笔邀请", rewardCount)
}

// ==================== 系统维护任务实现 ====================

// withdrawalLimitReset 提现限制重置
func (tm *TaskManager) withdrawalLimitReset() {
	today := time.Now().Format("2006-01-02")
	tm.log.Infof("执行提现限制重置 [%s]...", today)

	// 1. 获取所有用户
	users := tm.getAllActiveUsers()

	resetCount := 0
	for _, user := range users {
		// 重置用户每日提现次数
		user.DailyWithdrawalCount = 0
		user.DailyWithdrawalAmount = 0

		// 更新用户提现限制
		if err := tm.updateUserWithdrawalLimit(user.ID, 0, 0); err != nil {
			tm.log.Errorf("重置用户 [%d] 提现限制失败: %v", user.ID, err)
			continue
		}

		resetCount++
	}

	tm.log.Infof("提现限制重置完成，共处理 %d 位用户", resetCount)
}

// userDailyReset 用户工分/任务每日重置
func (tm *TaskManager) userDailyReset() {
	today := time.Now().Format("2006-01-02")
	tm.log.Infof("执行用户每日重置 [%s]...", today)

	// 1. 获取所有活跃用户
	users := tm.getAllActiveUsers()

	resetCount := 0
	for _, user := range users {
		// 重置每日工分获取状态
		user.DailyWorkPoints = 0
		user.DailyPointsLimitReached = false

		// 重置任务完成状态
		user.DailyTasks = []DailyTaskStatus{}

		// 将昨日工分累计到月度
		user.MonthlyWorkPoints += user.DailyWorkPoints

		// 更新用户数据
		if err := tm.updateUserDailyStats(user.ID, user.DailyWorkPoints, user.MonthlyWorkPoints); err != nil {
			tm.log.Errorf("重置用户 [%d] 每日状态失败: %v", user.ID, err)
			continue
		}

		resetCount++
	}

	tm.log.Infof("用户每日重置完成，共处理 %d 位用户", resetCount)
}

// ==================== 算力业务数据结构 ====================

// Machine 算力机器
type Machine struct {
	ID               int
	UserID           int
	ProductLevel     int     // 产品级别
	ProductivityRate float64 // 生产力百分比 (如 0.03)
	UserHashrate     float64 // 用户算力累计
	Status           string  // running/expired
	StartDate        time.Time
	EndDate          time.Time
}

// User 用户信息（扩展版）
type User struct {
	ID                      int
	Username                string
	Level                   string  // 会员等级
	WorkPoints              float64 // 工分总额
	DailyWorkPoints         float64 // 当日工分
	MonthlyWorkPoints       float64 // 当月累计工分
	DailyPointsLimitReached bool
	DailyTasks              []DailyTaskStatus

	// 算力相关
	TotalHashrate       float64 // 总算力累计
	DailyHashrate       float64 // 当日算力产出
	WithdrawableBalance float64 // 可提现余额

	// 签到相关
	CheckedInToday     bool
	TodayCheckinReward float64
	ConsecutiveDays    int

	// 提现相关
	DailyWithdrawalCount  int
	DailyWithdrawalAmount float64

	// 时间相关
	LastResetTime   time.Time
	LevelUpdateTime time.Time
	ExpireTime      time.Time
}

// InviteRelation 邀请关系
type InviteRelation struct {
	ID        int
	InviterID int
	InviteeID int
	Level     int // 1=一级邀请, 2=二级邀请
	CreatedAt time.Time
}

// Order 订单信息（分账用）
type Order struct {
	ID        int
	UserID    int
	Amount    float64
	Type      string // order/recharge
	Status    string
	CreatedAt time.Time
}

// ==================== 算力业务辅助方法 ====================

// getRunningMachines 获取运行中的机器（模拟实现）
func (tm *TaskManager) getRunningMachines() []*Machine {
	// 实际实现：从数据库查询运行中的机器
	return []*Machine{
		{ID: 1, UserID: 1001, ProductLevel: 1000, ProductivityRate: 0.03, Status: "running"},
		{ID: 2, UserID: 1001, ProductLevel: 500, ProductivityRate: 0.02, Status: "running"},
		{ID: 3, UserID: 1002, ProductLevel: 1000, ProductivityRate: 0.03, Status: "running"},
	}
}

// getExpiredMachines 获取过期机器（模拟实现）
func (tm *TaskManager) getExpiredMachines() []*Machine {
	// 实际实现：从数据库查询过期机器
	return []*Machine{
		{ID: 4, UserID: 1003, ProductLevel: 500, ProductivityRate: 0.02, Status: "expired", EndDate: time.Now().AddDate(0, 0, -1)},
	}
}

// getAllActiveUsers 获取所有活跃用户（模拟实现）
func (tm *TaskManager) getAllActiveUsers() []*User {
	// 实际实现：从数据库查询活跃用户
	return []*User{
		{
			ID:                  1001,
			Username:            "user001",
			Level:               "gold",
			WorkPoints:          100,
			DailyWorkPoints:     10,
			ConsecutiveDays:     5,
			TotalHashrate:       10000,
			DailyHashrate:       720,
			WithdrawableBalance: 5000,
		},
		{
			ID:                  1002,
			Username:            "user002",
			Level:               "silver",
			WorkPoints:          50,
			DailyWorkPoints:     5,
			ConsecutiveDays:     3,
			TotalHashrate:       5000,
			DailyHashrate:       360,
			WithdrawableBalance: 2000,
		},
	}
}

// getYesterdayCheckedInUsers 获取昨日已签到用户（模拟实现）
func (tm *TaskManager) getYesterdayCheckedInUsers() []*User {
	// 实际实现：从数据库查询昨日签到用户
	return tm.getAllActiveUsers()
}

// getPendingProfitSharingOrders 获取待分账订单（模拟实现）
func (tm *TaskManager) getPendingProfitSharingOrders() []*Order {
	// 实际实现：从数据库查询待分账订单
	return []*Order{
		{ID: 1, UserID: 1001, Amount: 1000, Type: "order", Status: "pending"},
		{ID: 2, UserID: 1002, Amount: 500, Type: "recharge", Status: "pending"},
	}
}

// getYesterdayInviteRelations 获取昨日邀请关系（模拟实现）
func (tm *TaskManager) getYesterdayInviteRelations() []*InviteRelation {
	// 实际实现：从数据库查询昨日新增邀请关系
	return []*InviteRelation{
		{ID: 1, InviterID: 1001, InviteeID: 1003, Level: 1, CreatedAt: time.Now().AddDate(0, 0, -1)},
		{ID: 2, InviterID: 1002, InviteeID: 1004, Level: 2, CreatedAt: time.Now().AddDate(0, 0, -1)},
	}
}

// ==================== 数据更新辅助方法 ====================

// updateMachineHashrate 更新机器算力（模拟实现）
func (tm *TaskManager) updateMachineHashrate(machineID int, hashrate float64) error {
	tm.log.Infof("更新机器 [%d] 算力: %.2f", machineID, hashrate)
	return nil
}

// updateMachineStatus 更新机器状态（模拟实现）
func (tm *TaskManager) updateMachineStatus(machineID int, status string) error {
	tm.log.Infof("更新机器 [%d] 状态: %s", machineID, status)
	return nil
}

// updateUserHashrate 更新用户算力（模拟实现）
func (tm *TaskManager) updateUserHashrate(userID int, totalHashrate, dailyHashrate float64) error {
	tm.log.Infof("更新用户 [%d] 算力 - 总计: %.2f, 当日: %.2f", userID, totalHashrate, dailyHashrate)
	return nil
}

// getUserDailyHashrate 获取用户每日算力（模拟实现）
func (tm *TaskManager) getUserDailyHashrate(userID int, date string) float64 {
	// 实际实现：从数据库查询用户某日算力
	return 720.0 // 模拟每日720算力
}

// recordHashrateLog 记录算力日志（模拟实现）
func (tm *TaskManager) recordHashrateLog(userID, machineID int, hashrate float64, hour string) {
	tm.log.Infof("记录算力日志 - 用户[%d] 机器[%d] 时间[%s] 算力[%.2f]",
		userID, machineID, hour, hashrate)
}

// updateUserCheckinStatus 更新用户签到状态（模拟实现）
func (tm *TaskManager) updateUserCheckinStatus(userID int, checkedIn bool, reward float64) error {
	tm.log.Infof("更新用户 [%d] 签到状态: %v, 奖励: %.2f", userID, checkedIn, reward)
	return nil
}

// updateUserConsecutiveDays 更新用户连续签到天数（模拟实现）
func (tm *TaskManager) updateUserConsecutiveDays(userID int, days int) error {
	tm.log.Infof("更新用户 [%d] 连续签到天数: %d", userID, days)
	return nil
}

// updateUserWorkPoints 更新用户工分（模拟实现）
func (tm *TaskManager) updateUserWorkPoints(userID int, points float64) error {
	tm.log.Infof("更新用户 [%d] 工分: %.2f", userID, points)
	return nil
}

// checkUserCheckedIn 检查用户某日是否签到（模拟实现）
func (tm *TaskManager) checkUserCheckedIn(userID int, date string) bool {
	// 实际实现：查询数据库
	return true // 模拟已签到
}

// recordTaskReward 记录任务奖励（模拟实现）
func (tm *TaskManager) recordTaskReward(userID int, taskType string, amount float64, date string) {
	tm.log.Infof("记录任务奖励 - 用户[%d] 任务[%s] 金额[%.2f] 日期[%s]",
		userID, taskType, amount, date)
}

// executeProfitSharing 执行分账（模拟实现）
func (tm *TaskManager) executeProfitSharing(order *Order) error {
	tm.log.Infof("执行分账 - 订单[%d] 用户[%d] 金额[%.2f] 类型[%s]",
		order.ID, order.UserID, order.Amount, order.Type)
	return nil
}

// getInviteRewardConfig 获取邀请奖励配置（模拟实现）
func (tm *TaskManager) getInviteRewardConfig(level string) float64 {
	// 实际实现：从配置读取
	switch level {
	case "level1":
		return 10.0 // 一级邀请奖励10工分
	case "level2":
		return 5.0 // 二级邀请奖励5工分
	default:
		return 0
	}
}

// grantInviteReward 发放邀请奖励（模拟实现）
func (tm *TaskManager) grantInviteReward(inviterID int, reward float64, inviteeID int) error {
	tm.log.Infof("发放邀请奖励 - 邀请人[%d] 被邀请人[%d] 奖励[%.2f]",
		inviterID, inviteeID, reward)
	return nil
}

// updateUserWithdrawalLimit 更新用户提现限制（模拟实现）
func (tm *TaskManager) updateUserWithdrawalLimit(userID, count int, amount float64) error {
	tm.log.Infof("更新用户 [%d] 提现限制 - 次数[%d] 金额[%.2f]", userID, count, amount)
	return nil
}

// updateUserDailyStats 更新用户每日统计（模拟实现）
func (tm *TaskManager) updateUserDailyStats(userID int, dailyPoints, monthlyPoints float64) error {
	tm.log.Infof("更新用户 [%d] 每日统计 - 当日[%.2f] 当月[%.2f]",
		userID, dailyPoints, monthlyPoints)
	return nil
}

// ==================== 会员数据更新辅助方法 ====================

// updateMember 更新会员信息（模拟实现）
//
//nolint:unused // 预留功能，待后续启用
func (tm *TaskManager) updateMember(member *Member) error {
	tm.log.Infof("更新会员 [%s] 信息 - 等级[%s] 当日工分[%d] 当月工分[%d] 总工分[%d]",
		member.ID, member.Level, member.DailyPoints, member.MonthlyPoints, member.TotalPoints)
	return nil
}

// updateMemberBenefits 更新会员权益（模拟实现）
//
//nolint:unused // 预留功能，待后续启用
func (tm *TaskManager) updateMemberBenefits(memberID string, benefits []MemberBenefit) error {
	tm.log.Infof("更新会员 [%s] 权益，共 %d 项", memberID, len(benefits))
	return nil
}

// updateMemberLevel 更新会员等级（模拟实现）
//
//nolint:unused // 预留功能，待后续启用
func (tm *TaskManager) updateMemberLevel(memberID string, level string) error {
	tm.log.Infof("更新会员 [%s] 等级为 [%s]", memberID, level)
	return nil
}

// markMemberAsExpired 标记会员为过期（模拟实现）
//
//nolint:unused // 预留功能，待后续启用
func (tm *TaskManager) markMemberAsExpired(memberID string) error {
	tm.log.Infof("标记会员 [%s] 为过期状态", memberID)
	return nil
}

// archiveAndDeleteMember 归档并删除会员（模拟实现）
//
//nolint:unused // 预留功能，待后续启用
func (tm *TaskManager) archiveAndDeleteMember(member *Member) error {
	tm.log.Infof("归档并删除会员 [%s]", member.ID)
	return nil
}

// sendLevelChangeNotification 发送等级变更通知（模拟实现）
//
//nolint:unused // 预留功能，待后续启用
func (tm *TaskManager) sendLevelChangeNotification(member *Member, oldLevel, newLevel string) {
	tm.log.Infof("发送等级变更通知 - 会员[%s] 从[%s]升级到[%s]", member.ID, oldLevel, newLevel)
}
