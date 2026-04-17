package data

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"payment/internal/biz"
)

// PaymentChannel 支付渠道数据模型
type PaymentChannel struct {
	ID         uint32  `gorm:"primarykey"`
	Name       string  `gorm:"type:varchar(50);not null"`
	Code       string  `gorm:"uniqueIndex:idx_code;type:varchar(30);not null"`
	Type       string  `gorm:"index:idx_type;type:varchar(20);not null"`
	APIURL     string  `gorm:"type:varchar(255);not null"`
	APIKey     string  `gorm:"type:varchar(255);not null"`
	APISecret  string  `gorm:"type:varchar(255)"`
	MerchantID string  `gorm:"type:varchar(100)"`
	AppID      string  `gorm:"type:varchar(100)"`
	NotifyURL  string  `gorm:"type:varchar(255)"`
	ReturnURL  string  `gorm:"type:varchar(255)"`
	MinAmount  float64 `gorm:"type:decimal(10,2);default:0.00"`
	MaxAmount  float64 `gorm:"type:decimal(10,2);default:999999.99"`
	Sort       int32   `gorm:"default:0"`
	IsDefault  int8    `gorm:"index:idx_is_default;default:0"`
	Status     int8    `gorm:"index:idx_status;default:1"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type channelRepo struct {
	data *Data
	log  *log.Helper
}

// NewChannelRepo 创建渠道仓库
func NewChannelRepo(data *Data, logger log.Logger) biz.ChannelRepo {
	return &channelRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *channelRepo) ListChannels(ctx context.Context, typ string, statusFilter int32) ([]*biz.PaymentChannel, error) {
	var channels []PaymentChannel
	query := r.data.db.Model(&PaymentChannel{})
	if typ != "" {
		query = query.Where("type = ?", typ)
	}
	if statusFilter >= 0 {
		query = query.Where("status = ?", statusFilter)
	}
	if err := query.Order("sort ASC").Find(&channels).Error; err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	var bizChannels []*biz.PaymentChannel
	for _, c := range channels {
		bizChannels = append(bizChannels, r.toBizChannel(&c))
	}
	return bizChannels, nil
}

func (r *channelRepo) GetChannel(ctx context.Context, id uint32) (*biz.PaymentChannel, error) {
	var channel PaymentChannel
	if err := r.data.db.First(&channel, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "支付渠道不存在")
		}
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return r.toBizChannel(&channel), nil
}

func (r *channelRepo) CreateChannel(ctx context.Context, c *biz.PaymentChannel) (*biz.PaymentChannel, error) {
	channel := PaymentChannel{
		Name:       c.Name,
		Code:       c.Code,
		Type:       c.Type,
		APIURL:     c.APIURL,
		APIKey:     c.APIKey,
		APISecret:  c.APISecret,
		MerchantID: c.MerchantID,
		AppID:      c.AppID,
		NotifyURL:  c.NotifyURL,
		ReturnURL:  c.ReturnURL,
		MinAmount:  c.MinAmount,
		MaxAmount:  c.MaxAmount,
		Status:     1,
	}
	if err := r.data.db.Create(&channel).Error; err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return r.toBizChannel(&channel), nil
}

func (r *channelRepo) UpdateChannel(ctx context.Context, c *biz.PaymentChannel) (*biz.PaymentChannel, error) {
	updates := map[string]interface{}{}
	if c.Name != "" {
		updates["name"] = c.Name
	}
	if c.APIURL != "" {
		updates["api_url"] = c.APIURL
	}
	if c.Status != 0 {
		updates["status"] = c.Status
	}
	if err := r.data.db.Model(&PaymentChannel{}).Where("id = ?", c.ID).Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.GetChannel(ctx, c.ID)
}

func (r *channelRepo) DeleteChannel(ctx context.Context, id uint32) error {
	return r.data.db.Delete(&PaymentChannel{}, id).Error
}

func (r *channelRepo) GetChannelByCode(ctx context.Context, code string) (*biz.PaymentChannel, error) {
	var channel PaymentChannel
	if err := r.data.db.Where("code = ?", code).First(&channel).Error; err != nil {
		return nil, err
	}
	return r.toBizChannel(&channel), nil
}

func (r *channelRepo) toBizChannel(c *PaymentChannel) *biz.PaymentChannel {
	return &biz.PaymentChannel{
		ID:         c.ID,
		Name:       c.Name,
		Code:       c.Code,
		Type:       c.Type,
		APIURL:     c.APIURL,
		APIKey:     c.APIKey,
		APISecret:  c.APISecret,
		MerchantID: c.MerchantID,
		AppID:      c.AppID,
		NotifyURL:  c.NotifyURL,
		ReturnURL:  c.ReturnURL,
		MinAmount:  c.MinAmount,
		MaxAmount:  c.MaxAmount,
		Sort:       c.Sort,
		IsDefault:  int32(c.IsDefault),
		Status:     int32(c.Status),
	}
}
