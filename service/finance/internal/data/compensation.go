package data

import (
	"context"
	"finance/internal/biz"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// HashrateCompensationRecord 算力补偿记录 GORM 模型
type HashrateCompensationRecord struct {
	ID            uint64     `gorm:"primarykey"`
	UserID        uint32     `gorm:"index:idx_user_id;not null"`
	Amount        float64    `gorm:"column:amount;type:decimal(15,2);not null"`
	RequestID     string     `gorm:"column:request_id;type:varchar(100);not null;uniqueIndex:uk_request_id"`
	Reason        string     `gorm:"column:reason;type:varchar(255)"`
	Status        int8       `gorm:"column:status;index:idx_status;default:0"`
	RetryTimes    uint32     `gorm:"column:retry_times;default:0"`
	CompensatedAt *time.Time `gorm:"column:compensated_at"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (HashrateCompensationRecord) TableName() string {
	return "hashrate_compensation_records"
}

type hashrateCompensationRepo struct {
	data *Data
	log  *log.Helper
}

// NewHashrateCompensationRepo 创建算力补偿记录仓库
func NewHashrateCompensationRepo(data *Data, logger log.Logger) biz.HashrateCompensationRepo {
	return &hashrateCompensationRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *hashrateCompensationRepo) CreateCompensationRecord(ctx context.Context, record *biz.HashrateCompensation) error {
	m := &HashrateCompensationRecord{
		UserID:    record.UserID,
		Amount:    record.Amount,
		RequestID: record.RequestID,
		Reason:    record.Reason,
		Status:    record.Status,
	}
	return r.data.db.WithContext(ctx).Create(m).Error
}

func (r *hashrateCompensationRepo) ListPendingCompensations(ctx context.Context, limit int) ([]*biz.HashrateCompensation, error) {
	var records []HashrateCompensationRecord
	if err := r.data.db.WithContext(ctx).
		Where("status = ?", 0).
		Order("created_at ASC").
		Limit(limit).
		Find(&records).Error; err != nil {
		return nil, err
	}

	var result []*biz.HashrateCompensation
	for _, rec := range records {
		result = append(result, toBizHashrateCompensation(&rec))
	}
	return result, nil
}

func (r *hashrateCompensationRepo) MarkCompensated(ctx context.Context, id uint64) error {
	now := time.Now()
	return r.data.db.WithContext(ctx).
		Model(&HashrateCompensationRecord{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":         1,
			"compensated_at": &now,
		}).Error
}

func (r *hashrateCompensationRepo) IncrementRetryTimes(ctx context.Context, id uint64) error {
	return r.data.db.WithContext(ctx).
		Model(&HashrateCompensationRecord{}).
		Where("id = ?", id).
		UpdateColumn("retry_times", gorm.Expr("retry_times + 1")).Error
}

func toBizHashrateCompensation(rec *HashrateCompensationRecord) *biz.HashrateCompensation {
	return &biz.HashrateCompensation{
		ID:            rec.ID,
		UserID:        rec.UserID,
		Amount:        rec.Amount,
		RequestID:     rec.RequestID,
		Reason:        rec.Reason,
		Status:        rec.Status,
		RetryTimes:    rec.RetryTimes,
		CompensatedAt: rec.CompensatedAt,
		CreatedAt:     rec.CreatedAt,
		UpdatedAt:     rec.UpdatedAt,
	}
}
