package biz

import (
	"context"
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
	RelatedID  uint32
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
	RelatedID     uint32
	CreatedAt     time.Time
}

// CheckIn 签到领域模型
type CheckIn struct {
	ID              uint64
	UserID          uint32
	CheckInDate     string
	ConsecutiveDays uint32
	RewardPoints    float64
	CreatedAt       time.Time
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
}

// CheckInRepo 签到存储接口
type CheckInRepo interface {
	CheckIn(ctx context.Context, c *CheckIn) (*CheckIn, error)
	GetLastCheckIn(ctx context.Context, userID uint32) (*CheckIn, error)
}

// FinanceUsecase 财务用例
type FinanceUsecase struct {
	rechargeRepo           RechargeRepo
	withdrawalRepo         WithdrawalRepo
	withdrawalMessageQueue WithdrawalMessageQueue
	incomeLogRepo          IncomeLogRepo
	balanceLogRepo         BalanceLogRepo
	checkInRepo            CheckInRepo
	log                    *log.Helper
}

// NewFinanceUsecase 创建财务用例
func NewFinanceUsecase(
	rechargeRepo RechargeRepo,
	withdrawalRepo WithdrawalRepo,
	withdrawalMessageQueue WithdrawalMessageQueue,
	incomeLogRepo IncomeLogRepo,
	balanceLogRepo BalanceLogRepo,
	checkInRepo CheckInRepo,
	logger log.Logger,
) *FinanceUsecase {
	return &FinanceUsecase{
		rechargeRepo:           rechargeRepo,
		withdrawalRepo:         withdrawalRepo,
		withdrawalMessageQueue: withdrawalMessageQueue,
		incomeLogRepo:          incomeLogRepo,
		balanceLogRepo:         balanceLogRepo,
		checkInRepo:            checkInRepo,
		log:                    log.NewHelper(logger),
	}
}

// Recharge 充值
func (uc *FinanceUsecase) Recharge(ctx context.Context, r *Recharge) (*Recharge, error) {
	return uc.rechargeRepo.CreateRecharge(ctx, r)
}

// Withdraw 提现 - 推送到消息队列进行异步处理
func (uc *FinanceUsecase) Withdraw(ctx context.Context, w *Withdrawal) (*Withdrawal, error) {
	// 推送到 RabbitMQ 消息队列
	if err := uc.withdrawalMessageQueue.PublishWithdrawal(ctx, w); err != nil {
		return nil, err
	}
	uc.log.Infof("withdrawal request queued: user_id=%d, amount=%.2f", w.UserID, w.Amount)

	// 返回一个临时状态，实际记录由消费者创建
	return &Withdrawal{
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

// CheckIn 签到
func (uc *FinanceUsecase) CheckIn(ctx context.Context, c *CheckIn) (*CheckIn, error) {
	return uc.checkInRepo.CheckIn(ctx, c)
}

// GetUserBalance 获取用户余额
func (uc *FinanceUsecase) GetUserBalance(ctx context.Context, userID uint32) (float64, float64, error) {
	// TODO: 实现从用户服务获取余额的逻辑
	return 0, 0, nil
}
