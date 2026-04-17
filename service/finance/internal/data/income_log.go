package data

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"finance/internal/biz"
)

// IncomeLog 收益明细数据模型
type IncomeLog struct {
	ID         uint64  `gorm:"primarykey"`
	UserID     uint32  `gorm:"index:idx_user_id;not null"`
	Phone      string  `gorm:"index:idx_phone;type:varchar(20);not null"`
	Name       string  `gorm:"type:varchar(50);not null"`
	Source     string  `gorm:"type:varchar(100);not null"`
	SourceType int8    `gorm:"index:idx_source_type;not null"`
	Amount     float64 `gorm:"type:decimal(10,2);not null"`
	RelatedID  uint32  `gorm:"index:idx_related_id"`
	CreatedAt  time.Time
}

type incomeLogRepo struct {
	data *Data
	log  *log.Helper
}

// NewIncomeLogRepo 创建收益日志仓库
func NewIncomeLogRepo(data *Data, logger log.Logger) biz.IncomeLogRepo {
	return &incomeLogRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *incomeLogRepo) ListIncomeLogs(ctx context.Context, userID uint32, sourceType int32, page, pageSize uint32) ([]*biz.IncomeLog, uint32, error) {
	var logs []IncomeLog
	var total int64

	query := r.data.db.Model(&IncomeLog{})
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if sourceType > 0 {
		query = query.Where("source_type = ?", sourceType)
	}

	query.Count(&total)
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Limit(int(pageSize)).Offset(int(offset)).Find(&logs).Error; err != nil {
		return nil, 0, status.Errorf(codes.Internal, err.Error())
	}

	var bizLogs []*biz.IncomeLog
	for _, l := range logs {
		bizLogs = append(bizLogs, &biz.IncomeLog{
			ID:         l.ID,
			UserID:     l.UserID,
			Phone:      l.Phone,
			Name:       l.Name,
			Source:     l.Source,
			SourceType: int32(l.SourceType),
			Amount:     l.Amount,
			RelatedID:  l.RelatedID,
			CreatedAt:  l.CreatedAt,
		})
	}
	return bizLogs, uint32(total), nil
}
