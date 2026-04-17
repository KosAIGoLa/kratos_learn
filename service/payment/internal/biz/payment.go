package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// PaymentChannel 支付渠道领域模型
type PaymentChannel struct {
	ID         uint32
	Name       string
	Code       string
	Type       string
	APIURL     string
	APIKey     string
	APISecret  string
	MerchantID string
	AppID      string
	NotifyURL  string
	ReturnURL  string
	MinAmount  float64
	MaxAmount  float64
	Sort       int32
	IsDefault  int32
	Status     int32
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// PaymentAPILog 支付API调用日志领域模型
type PaymentAPILog struct {
	ID           uint64
	OrderNo      string
	ChannelID    uint32
	ChannelCode  int32
	Action       int32
	RequestURL   string
	Status       int32
	ErrorMsg     string
	RetryTimes   uint32
	RequestTime  time.Time
	ResponseTime *time.Time
	CreatedAt    time.Time
}

// PaymentResult 支付结果
type PaymentResult struct {
	OrderNo        string
	ChannelOrderNo string
	Amount         float64
	Status         int32
	PayURL         string
}

// ChannelRepo 支付渠道存储接口
type ChannelRepo interface {
	ListChannels(ctx context.Context, typ string, status int32) ([]*PaymentChannel, error)
	GetChannel(ctx context.Context, id uint32) (*PaymentChannel, error)
	CreateChannel(ctx context.Context, c *PaymentChannel) (*PaymentChannel, error)
	UpdateChannel(ctx context.Context, c *PaymentChannel) (*PaymentChannel, error)
	DeleteChannel(ctx context.Context, id uint32) error
	GetChannelByCode(ctx context.Context, code string) (*PaymentChannel, error)
}

// APILogRepo API日志存储接口
type APILogRepo interface {
	ListAPILogs(ctx context.Context, orderNo string, channelID uint32, action, status int32, page, pageSize uint32) ([]*PaymentAPILog, uint32, error)
	LogRequest(ctx context.Context, log *PaymentAPILog) (*PaymentAPILog, error)
	UpdateResponse(ctx context.Context, id uint64, status int32, errorMsg string) error
}

// PaymentUsecase 支付用例
type PaymentUsecase struct {
	channelRepo ChannelRepo
	apiLogRepo  APILogRepo
	log         *log.Helper
}

// NewPaymentUsecase 创建支付用例
func NewPaymentUsecase(channelRepo ChannelRepo, apiLogRepo APILogRepo, logger log.Logger) *PaymentUsecase {
	return &PaymentUsecase{
		channelRepo: channelRepo,
		apiLogRepo:  apiLogRepo,
		log:         log.NewHelper(logger),
	}
}

// ListChannels 获取支付渠道列表
func (uc *PaymentUsecase) ListChannels(ctx context.Context, typ string, status int32) ([]*PaymentChannel, error) {
	return uc.channelRepo.ListChannels(ctx, typ, status)
}

// GetChannel 获取支付渠道
func (uc *PaymentUsecase) GetChannel(ctx context.Context, id uint32) (*PaymentChannel, error) {
	return uc.channelRepo.GetChannel(ctx, id)
}

// CreateChannel 创建支付渠道
func (uc *PaymentUsecase) CreateChannel(ctx context.Context, c *PaymentChannel) (*PaymentChannel, error) {
	return uc.channelRepo.CreateChannel(ctx, c)
}

// UpdateChannel 更新支付渠道
func (uc *PaymentUsecase) UpdateChannel(ctx context.Context, c *PaymentChannel) (*PaymentChannel, error) {
	return uc.channelRepo.UpdateChannel(ctx, c)
}

// DeleteChannel 删除支付渠道
func (uc *PaymentUsecase) DeleteChannel(ctx context.Context, id uint32) error {
	return uc.channelRepo.DeleteChannel(ctx, id)
}

// ListAPILogs 获取API日志列表
func (uc *PaymentUsecase) ListAPILogs(ctx context.Context, orderNo string, channelID uint32, action, status int32, page, pageSize uint32) ([]*PaymentAPILog, uint32, error) {
	return uc.apiLogRepo.ListAPILogs(ctx, orderNo, channelID, action, status, page, pageSize)
}

// ProcessPayment 处理支付
func (uc *PaymentUsecase) ProcessPayment(ctx context.Context, orderNo string, amount float64, channelCode string) (*PaymentResult, error) {
	channel, err := uc.channelRepo.GetChannelByCode(ctx, channelCode)
	if err != nil {
		return nil, err
	}

	// TODO: 调用渠道支付接口
	result := &PaymentResult{
		OrderNo: orderNo,
		Amount:  amount,
		Status:  0, // 处理中
	}

	// 记录API日志
	_ = channel // 使用channel

	return result, nil
}

// HandleCallback 处理支付回调
func (uc *PaymentUsecase) HandleCallback(ctx context.Context, channelCode string, orderNo string) (*PaymentResult, error) {
	// TODO: 验证签名，处理回调
	return &PaymentResult{
		OrderNo: orderNo,
		Status:  1, // 成功
	}, nil
}
