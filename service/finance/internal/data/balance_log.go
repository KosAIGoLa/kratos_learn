package data

import (
	"context"
	"finance/internal/biz"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// BalanceLog 余额变动数据模型
type BalanceLog struct {
	ID            uint64  `gorm:"primarykey"`
	UserID        uint32  `gorm:"index:idx_user_id;not null"`
	Type          int8    `gorm:"index:idx_type;not null"`
	Amount        float64 `gorm:"type:decimal(15,2);not null"`
	BeforeBalance float64 `gorm:"type:decimal(15,2);not null"`
	AfterBalance  float64 `gorm:"type:decimal(15,2);not null"`
	Remark        string  `gorm:"type:varchar(255)"`
	RelatedID     uint32  `gorm:"index:idx_related_id"`
	CreatedAt     time.Time
}

type balanceLogRepo struct {
	data *Data
	log  *log.Helper
}

// NewBalanceLogRepo 创建余额日志仓库
func NewBalanceLogRepo(data *Data, logger log.Logger) biz.BalanceLogRepo {
	return &balanceLogRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *balanceLogRepo) ListBalanceLogs(ctx context.Context, userID uint32, typ int32, page, pageSize uint32) ([]*biz.BalanceLog, uint32, error) {
	var logs []BalanceLog
	var total int64

	query := r.data.db.Model(&BalanceLog{})
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if typ > 0 {
		query = query.Where("type = ?", typ)
	}

	query.Count(&total)
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Limit(int(pageSize)).Offset(int(offset)).Find(&logs).Error; err != nil {
		return nil, 0, status.Errorf(codes.Internal, err.Error())
	}

	var bizLogs []*biz.BalanceLog
	for _, l := range logs {
		bizLogs = append(bizLogs, &biz.BalanceLog{
			ID:            l.ID,
			UserID:        l.UserID,
			Type:          int32(l.Type),
			Amount:        l.Amount,
			BeforeBalance: l.BeforeBalance,
			AfterBalance:  l.AfterBalance,
			Remark:        l.Remark,
			RelatedID:     l.RelatedID,
			CreatedAt:     l.CreatedAt,
		})
	}
	return bizLogs, uint32(total), nil
}
