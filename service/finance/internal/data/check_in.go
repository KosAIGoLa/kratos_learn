package data

import (
	"context"
	"finance/internal/biz"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CheckIn 签到数据模型
type CheckIn struct {
	ID              uint64  `gorm:"primarykey"`
	UserID          uint32  `gorm:"index:idx_user_id;not null"`
	CheckInDate     string  `gorm:"index:idx_check_in_date;not null"`
	ConsecutiveDays uint32  `gorm:"default:1"`
	RewardPoints    float64 `gorm:"type:decimal(10,2);default:0.00"`
	CreatedAt       time.Time
}

type checkInRepo struct {
	data *Data
	log  *log.Helper
}

// NewCheckInRepo 创建签到仓库
func NewCheckInRepo(data *Data, logger log.Logger) biz.CheckInRepo {
	return &checkInRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *checkInRepo) CheckIn(ctx context.Context, c *biz.CheckIn) (*biz.CheckIn, error) {
	checkIn := CheckIn{
		UserID:          c.UserID,
		CheckInDate:     c.CheckInDate,
		ConsecutiveDays: c.ConsecutiveDays,
		RewardPoints:    c.RewardPoints,
		CreatedAt:       time.Now(),
	}
	if err := r.data.db.Create(&checkIn).Error; err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &biz.CheckIn{
		ID:              checkIn.ID,
		UserID:          checkIn.UserID,
		CheckInDate:     checkIn.CheckInDate,
		ConsecutiveDays: checkIn.ConsecutiveDays,
		RewardPoints:    checkIn.RewardPoints,
		CreatedAt:       checkIn.CreatedAt,
	}, nil
}

func (r *checkInRepo) GetLastCheckIn(ctx context.Context, userID uint32) (*biz.CheckIn, error) {
	var checkIn CheckIn
	if err := r.data.db.Where("user_id = ?", userID).Order("created_at DESC").First(&checkIn).Error; err != nil {
		return nil, err
	}
	return &biz.CheckIn{
		ID:              checkIn.ID,
		UserID:          checkIn.UserID,
		CheckInDate:     checkIn.CheckInDate,
		ConsecutiveDays: checkIn.ConsecutiveDays,
		RewardPoints:    checkIn.RewardPoints,
		CreatedAt:       checkIn.CreatedAt,
	}, nil
}
