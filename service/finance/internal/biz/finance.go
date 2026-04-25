package biz

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// Recharge 充值领域模型
type Recharge struct {
	ID         uint64
	UserID     uint32
	InviteCode string
	Phone      string
	Name       string
	OrderNo    string
	Amount     float64
	Status     int32
	CreatedAt  time.Time
}

// Withdrawal 提现领域模型
type Withdrawal struct {
	ID          uint64
	RequestID   string
	UserID      uint32
	Phone       string
	Name        string
	BankCard    string
	BankName    string
	Amount      float64
	Status      int32
	Remark      string
	ProcessedAt *time.Time
	CreatedAt   time.Time
}

// IncomeLog 收益明细领域模型
type IncomeLog struct {
	ID         uint64
	UserID     uint32
	Phone      string
	Name       string
	Source     string
	SourceType int32
	Amount     float64
	RelatedID  uint64
	CreatedAt  time.Time
}

// BalanceLog 余额变动领域模型
type BalanceLog struct {
	ID            uint64
	UserID        uint32
	Type          int32
	Amount        float64
	BeforeBalance float64
	AfterBalance  float64
	Remark        string
	RelatedID     uint64
	CreatedAt     time.Time
}

const (
	BalanceLogTypeRecharge           int32 = 1
	BalanceLogTypeWithdraw           int32 = 2
	BalanceLogTypeIncome             int32 = 3
	BalanceLogTypeDeduction          int32 = 4
	BalanceLogTypeHashrateConversion int32 = 5
)

// CheckIn 签到领域模型
type CheckIn struct {
	ID              uint64
	UserID          uint32
	CheckInDate     string
	ConsecutiveDays uint32
	RewardPoints    float64
	CreatedAt       time.Time
}

// UserAsset 用户资产
type UserAsset struct {
	UserID     uint32
	Balance    float64
	WorkPoints float64
}

// HashrateConversion 算力转换结果
type HashrateConversion struct {
	UserID         uint32
	Amount         float64
	BeforeHashrate float64
	AfterHashrate  float64
	BeforeBalance  float64
	AfterBalance   float64
	Remark         string
	CreatedAt      time.Time
}

// HashrateCompensation 算力补偿记录
type HashrateCompensation struct {
	ID            uint64
	UserID        uint32
	Amount        float64
	RequestID     string
	Reason        string
	Status        int8
	RetryTimes    uint32
	CompensatedAt *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// HashrateCompensationRepo 算力补偿记录存储接口
type HashrateCompensationRepo interface {
	CreateCompensationRecord(ctx context.Context, record *HashrateCompensation) error
	ListPendingCompensations(ctx context.Context, limit int) ([]*HashrateCompensation, error)
	MarkCompensated(ctx context.Context, id uint64) error
	IncrementRetryTimes(ctx context.Context, id uint64) error
}

// RechargeRepo 充值存储接口
type RechargeRepo interface {
	CreateRecharge(ctx context.Context, r *Recharge) (*Recharge, error)
	GetRechargeByOrderNo(ctx context.Context, orderNo string) (*Recharge, error)
	ListRecharges(ctx context.Context, userID uint32, status int32, page, pageSize uint32) ([]*Recharge, uint32, error)
}

// WithdrawalRepo 提现存储接口
type WithdrawalRepo interface {
	CreateWithdrawal(ctx context.Context, w *Withdrawal) (*Withdrawal, error)
	GetWithdrawal(ctx context.Context, id uint64) (*Withdrawal, error)
	ListWithdrawals(ctx context.Context, userID uint32, status int32, page, pageSize uint32) ([]*Withdrawal, uint32, error)
}

// WithdrawalMessageQueue 提现消息队列接口
type WithdrawalMessageQueue interface {
	PublishWithdrawal(ctx context.Context, w *Withdrawal) error
}

// IncomeLogRepo 收益存储接口
type IncomeLogRepo interface {
	ListIncomeLogs(ctx context.Context, userID uint32, sourceType int32, page, pageSize uint32) ([]*IncomeLog, uint32, error)
}

// BalanceLogRepo 余额变动存储接口
type BalanceLogRepo interface {
	ListBalanceLogs(ctx context.Context, userID uint32, typ int32, page, pageSize uint32) ([]*BalanceLog, uint32, error)
	CreateBalanceLog(ctx context.Context, log *BalanceLog) (*BalanceLog, error)
}

// CheckInRepo 签到存储接口
type CheckInRepo interface {
	CheckIn(ctx context.Context, c *CheckIn) (*CheckIn, error)
	GetLastCheckIn(ctx context.Context, userID uint32) (*CheckIn, error)
}

// UserAssetRepo 用户资产存储接口
type UserAssetRepo interface {
	GetUserAsset(ctx context.Context, userID uint32) (*UserAsset, error)
	ConvertHashrate(ctx context.Context, userID uint32, amount float64) (*HashrateConversion, error)
	RestoreHashrate(ctx context.Context, userID uint32, amount float64, requestID string) error
}

// FinanceUsecase 财务用例
type FinanceUsecase struct {
	rechargeRepo             RechargeRepo
	withdrawalRepo           WithdrawalRepo
	withdrawalMessageQueue   WithdrawalMessageQueue
	incomeLogRepo            IncomeLogRepo
	balanceLogRepo           BalanceLogRepo
	checkInRepo              CheckInRepo
	userAssetRepo            UserAssetRepo
	hashrateCompensationRepo HashrateCompensationRepo
	log                      *log.Helper
}

// NewFinanceUsecase 创建财务用例
func NewFinanceUsecase(
	rechargeRepo RechargeRepo,
	withdrawalRepo WithdrawalRepo,
	withdrawalMessageQueue WithdrawalMessageQueue,
	incomeLogRepo IncomeLogRepo,
	balanceLogRepo BalanceLogRepo,
	checkInRepo CheckInRepo,
	userAssetRepo UserAssetRepo,
	hashrateCompensationRepo HashrateCompensationRepo,
	logger log.Logger,
) *FinanceUsecase {
	return &FinanceUsecase{
		rechargeRepo:             rechargeRepo,
		withdrawalRepo:           withdrawalRepo,
		withdrawalMessageQueue:   withdrawalMessageQueue,
		incomeLogRepo:            incomeLogRepo,
		balanceLogRepo:           balanceLogRepo,
		checkInRepo:              checkInRepo,
		userAssetRepo:            userAssetRepo,
		hashrateCompensationRepo: hashrateCompensationRepo,
		log:                      log.NewHelper(logger),
	}
}

// Recharge 充值
func (uc *FinanceUsecase) Recharge(ctx context.Context, r *Recharge) (*Recharge, error) {
	if r == nil {
		return nil, errors.New("recharge data is nil")
	}
	if r.Amount <= 0 {
		return nil, fmt.Errorf("recharge amount must be greater than 0, got %.2f", r.Amount)
	}
	return uc.rechargeRepo.CreateRecharge(ctx, r)
}

// Withdraw 提现 - 推送到消息队列进行异步处理
func (uc *FinanceUsecase) Withdraw(ctx context.Context, w *Withdrawal) (*Withdrawal, error) {
	if w == nil {
		return nil, errors.New("withdrawal data is nil")
	}
	if w.Amount <= 0 {
		return nil, fmt.Errorf("withdrawal amount must be greater than 0, got %.2f", w.Amount)
	}
	asset, err := uc.userAssetRepo.GetUserAsset(ctx, w.UserID)
	if err != nil {
		return nil, err
	}
	if asset.Balance < w.Amount {
		return nil, fmt.Errorf("insufficient balance: available %.2f, requested %.2f", asset.Balance, w.Amount)
	}

	// 生成唯一请求ID用于幂等控制
	w.RequestID = fmt.Sprintf("WDR-%d-%d", w.UserID, time.Now().UnixNano())

	// 推送到 RabbitMQ 消息队列
	if err := uc.withdrawalMessageQueue.PublishWithdrawal(ctx, w); err != nil {
		return nil, err
	}
	uc.log.Infof("withdrawal request queued: user_id=%d, amount=%.2f, request_id=%s", w.UserID, w.Amount, w.RequestID)

	// 返回一个临时状态，实际记录由消费者创建
	return &Withdrawal{
		RequestID: w.RequestID,
		UserID:    w.UserID,
		Amount:    w.Amount,
		BankCard:  w.BankCard,
		BankName:  w.BankName,
		Phone:     w.Phone,
		Name:      w.Name,
		Status:    0, // 待处理
		CreatedAt: time.Now(),
	}, nil
}

// GetRechargeByOrderNo 根据订单号获取充值记录
func (uc *FinanceUsecase) GetRechargeByOrderNo(ctx context.Context, orderNo string) (*Recharge, error) {
	return uc.rechargeRepo.GetRechargeByOrderNo(ctx, orderNo)
}

// ListRecharges 获取充值列表
func (uc *FinanceUsecase) ListRecharges(ctx context.Context, userID uint32, status int32, page, pageSize uint32) ([]*Recharge, uint32, error) {
	return uc.rechargeRepo.ListRecharges(ctx, userID, status, page, pageSize)
}

// GetWithdrawal 获取提现记录
func (uc *FinanceUsecase) GetWithdrawal(ctx context.Context, id uint64) (*Withdrawal, error) {
	return uc.withdrawalRepo.GetWithdrawal(ctx, id)
}

// ListWithdrawals 获取提现列表
func (uc *FinanceUsecase) ListWithdrawals(ctx context.Context, userID uint32, status int32, page, pageSize uint32) ([]*Withdrawal, uint32, error) {
	return uc.withdrawalRepo.ListWithdrawals(ctx, userID, status, page, pageSize)
}

// ListIncomeLogs 获取收益明细列表
func (uc *FinanceUsecase) ListIncomeLogs(ctx context.Context, userID uint32, sourceType int32, page, pageSize uint32) ([]*IncomeLog, uint32, error) {
	return uc.incomeLogRepo.ListIncomeLogs(ctx, userID, sourceType, page, pageSize)
}

// ListBalanceLogs 获取余额变动列表
func (uc *FinanceUsecase) ListBalanceLogs(ctx context.Context, userID uint32, typ int32, page, pageSize uint32) ([]*BalanceLog, uint32, error) {
	return uc.balanceLogRepo.ListBalanceLogs(ctx, userID, typ, page, pageSize)
}

// CreateBalanceLog 创建余额变动记录
func (uc *FinanceUsecase) CreateBalanceLog(ctx context.Context, log *BalanceLog) (*BalanceLog, error) {
	return uc.balanceLogRepo.CreateBalanceLog(ctx, log)
}

// CheckIn 签到，自动计算连续签到天数和奖励
func (uc *FinanceUsecase) CheckIn(ctx context.Context, c *CheckIn) (*CheckIn, error) {
	if c == nil {
		return nil, errors.New("check-in data is nil")
	}

	lastCheckIn, err := uc.checkInRepo.GetLastCheckIn(ctx, c.UserID)
	if err != nil {
		return nil, err
	}

	today := time.Now().Format("2006-01-02")
	c.CheckInDate = today

	if lastCheckIn != nil {
		lastDate, _ := time.Parse("2006-01-02", lastCheckIn.CheckInDate)
		yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		if lastDate.Format("2006-01-02") == yesterday {
			c.ConsecutiveDays = lastCheckIn.ConsecutiveDays + 1
		} else {
			c.ConsecutiveDays = 1
		}
	} else {
		c.ConsecutiveDays = 1
	}

	// 动态计算奖励
	switch {
	case c.ConsecutiveDays >= 7:
		c.RewardPoints = 20.0
	case c.ConsecutiveDays >= 3:
		c.RewardPoints = 15.0
	default:
		c.RewardPoints = 10.0
	}

	return uc.checkInRepo.CheckIn(ctx, c)
}

// GetUserBalance 获取用户余额
func (uc *FinanceUsecase) GetUserBalance(ctx context.Context, userID uint32) (float64, float64, error) {
	asset, err := uc.userAssetRepo.GetUserAsset(ctx, userID)
	if err != nil {
		return 0, 0, err
	}
	return asset.Balance, asset.WorkPoints, nil
}

// ConvertHashrate 手动将算力转换为余额
func (uc *FinanceUsecase) ConvertHashrate(ctx context.Context, userID uint32, amount float64) (*HashrateConversion, error) {
	return uc.userAssetRepo.ConvertHashrate(ctx, userID, amount)
}

// ListPendingHashrateCompensations 查询待补偿的算力记录
func (uc *FinanceUsecase) ListPendingHashrateCompensations(ctx context.Context, limit int) ([]*HashrateCompensation, error) {
	return uc.hashrateCompensationRepo.ListPendingCompensations(ctx, limit)
}

// CompensateHashrate 执行算力补偿（恢复算力并标记记录为已补偿）
func (uc *FinanceUsecase) CompensateHashrate(ctx context.Context, record *HashrateCompensation) error {
	// 1. 恢复用户算力
	if err := uc.userAssetRepo.RestoreHashrate(ctx, record.UserID, record.Amount, record.RequestID); err != nil {
		// 增加重试次数
		_ = uc.hashrateCompensationRepo.IncrementRetryTimes(ctx, record.ID)
		return err
	}

	// 2. 标记补偿记录为已补偿
	if err := uc.hashrateCompensationRepo.MarkCompensated(ctx, record.ID); err != nil {
		uc.log.Errorf("failed to mark compensation record %d as compensated: %v", record.ID, err)
		return err
	}

	uc.log.Infof("hashrate compensation completed: user=%d, amount=%.2f, request_id=%s", record.UserID, record.Amount, record.RequestID)
	return nil
}
