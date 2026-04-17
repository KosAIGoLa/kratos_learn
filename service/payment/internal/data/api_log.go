package data

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"payment/internal/biz"
)

// PaymentAPILog 支付API日志数据模型
type PaymentAPILog struct {
	ID           uint64 `gorm:"primarykey"`
	OrderNo      string `gorm:"index:idx_order_no;type:varchar(50);not null"`
	ChannelID    uint32 `gorm:"index:idx_channel_id;not null"`
	ChannelCode  int8   `gorm:"index:idx_channel_code;not null"`
	Action       int8   `gorm:"index:idx_action;not null"`
	RequestURL   string `gorm:"type:varchar(255);not null"`
	Status       int8   `gorm:"index:idx_status;default:0"`
	ErrorMsg     string `gorm:"type:varchar(500)"`
	RetryTimes   uint32 `gorm:"default:0"`
	RequestTime  time.Time
	ResponseTime *time.Time
	CreatedAt    time.Time
}

type apiLogRepo struct {
	data *Data
	log  *log.Helper
}

// NewAPILogRepo 创建API日志仓库
func NewAPILogRepo(data *Data, logger log.Logger) biz.APILogRepo {
	return &apiLogRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *apiLogRepo) ListAPILogs(ctx context.Context, orderNo string, channelID uint32, action, statusFilter int32, page, pageSize uint32) ([]*biz.PaymentAPILog, uint32, error) {
	var logs []PaymentAPILog
	var total int64

	query := r.data.db.Model(&PaymentAPILog{})
	if orderNo != "" {
		query = query.Where("order_no = ?", orderNo)
	}
	if channelID > 0 {
		query = query.Where("channel_id = ?", channelID)
	}
	if action > 0 {
		query = query.Where("action = ?", action)
	}
	if statusFilter >= 0 {
		query = query.Where("status = ?", statusFilter)
	}

	query.Count(&total)
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Limit(int(pageSize)).Offset(int(offset)).Find(&logs).Error; err != nil {
		return nil, 0, status.Errorf(codes.Internal, err.Error())
	}

	var bizLogs []*biz.PaymentAPILog
	for _, l := range logs {
		bizLogs = append(bizLogs, &biz.PaymentAPILog{
			ID:           l.ID,
			OrderNo:      l.OrderNo,
			ChannelID:    l.ChannelID,
			ChannelCode:  int32(l.ChannelCode),
			Action:       int32(l.Action),
			RequestURL:   l.RequestURL,
			Status:       int32(l.Status),
			ErrorMsg:     l.ErrorMsg,
			RetryTimes:   l.RetryTimes,
			RequestTime:  l.RequestTime,
			ResponseTime: l.ResponseTime,
			CreatedAt:    l.CreatedAt,
		})
	}
	return bizLogs, uint32(total), nil
}

func (r *apiLogRepo) LogRequest(ctx context.Context, log *biz.PaymentAPILog) (*biz.PaymentAPILog, error) {
	l := PaymentAPILog{
		OrderNo:     log.OrderNo,
		ChannelID:   log.ChannelID,
		ChannelCode: int8(log.ChannelCode),
		Action:      int8(log.Action),
		RequestURL:  log.RequestURL,
		Status:      0,
		RequestTime: time.Now(),
		CreatedAt:   time.Now(),
	}
	if err := r.data.db.Create(&l).Error; err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &biz.PaymentAPILog{
		ID:        l.ID,
		OrderNo:   l.OrderNo,
		Status:    int32(l.Status),
		CreatedAt: l.CreatedAt,
	}, nil
}

func (r *apiLogRepo) UpdateResponse(ctx context.Context, id uint64, statusValue int32, errorMsg string) error {
	now := time.Now()
	return r.data.db.Model(&PaymentAPILog{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":        statusValue,
		"error_msg":     errorMsg,
		"response_time": &now,
	}).Error
}
