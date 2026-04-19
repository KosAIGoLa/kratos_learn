package data

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"finance/internal/biz"
)

// Withdrawal 提现数据模型
type Withdrawal struct {
	ID          uint64  `gorm:"primarykey"`
	UserID      uint32  `gorm:"index:idx_user_id;not null"`
	Phone       string  `gorm:"index:idx_phone;type:varchar(20);not null"`
	Name        string  `gorm:"type:varchar(50);not null"`
	BankCard    string  `gorm:"type:varchar(30);not null"`
	BankName    string  `gorm:"type:varchar(50)"`
	Amount      float64 `gorm:"type:decimal(10,2);not null"`
	Status      int8    `gorm:"index:idx_status;default:0"`
	Remark      string  `gorm:"type:varchar(255)"`
	ProcessedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type withdrawalRepo struct {
	data *Data
	log  *log.Helper
}

// NewWithdrawalRepo 创建提现仓库
func NewWithdrawalRepo(data *Data, logger log.Logger) biz.WithdrawalRepo {
	return &withdrawalRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *withdrawalRepo) CreateWithdrawal(ctx context.Context, w *biz.Withdrawal) (*biz.Withdrawal, error) {
	withdrawal := Withdrawal{
		UserID:    w.UserID,
		Phone:     w.Phone,
		Name:      w.Name,
		BankCard:  w.BankCard,
		BankName:  w.BankName,
		Amount:    w.Amount,
		Status:    0, // 待审核
		CreatedAt: time.Now(),
	}
	if err := r.data.db.Create(&withdrawal).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create withdrawal: %v", err)
	}
	return r.toBizWithdrawal(&withdrawal), nil
}

func (r *withdrawalRepo) GetWithdrawal(ctx context.Context, id uint64) (*biz.Withdrawal, error) {
	var withdrawal Withdrawal
	if err := r.data.db.First(&withdrawal, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "提现记录不存在")
		}
		return nil, status.Errorf(codes.Internal, "failed to get withdrawal: %v", err)
	}
	return r.toBizWithdrawal(&withdrawal), nil
}

func (r *withdrawalRepo) ListWithdrawals(ctx context.Context, userID uint32, statusFilter int32, page, pageSize uint32) ([]*biz.Withdrawal, uint32, error) {
	var withdrawals []Withdrawal
	var total int64

	query := r.data.db.Model(&Withdrawal{})
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if statusFilter >= 0 {
		query = query.Where("status = ?", statusFilter)
	}

	query.Count(&total)
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Limit(int(pageSize)).Offset(int(offset)).Find(&withdrawals).Error; err != nil {
		return nil, 0, status.Errorf(codes.Internal, "failed to query withdrawals: %v", err)
	}

	var bizWithdrawals []*biz.Withdrawal
	for _, w := range withdrawals {
		bizWithdrawals = append(bizWithdrawals, r.toBizWithdrawal(&w))
	}
	return bizWithdrawals, uint32(total), nil
}

func (r *withdrawalRepo) toBizWithdrawal(w *Withdrawal) *biz.Withdrawal {
	return &biz.Withdrawal{
		ID:          w.ID,
		UserID:      w.UserID,
		Phone:       w.Phone,
		Name:        w.Name,
		BankCard:    w.BankCard,
		BankName:    w.BankName,
		Amount:      w.Amount,
		Status:      int32(w.Status),
		Remark:      w.Remark,
		ProcessedAt: w.ProcessedAt,
		CreatedAt:   w.CreatedAt,
	}
}
