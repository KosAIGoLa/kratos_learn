package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	v1 "payment/api/payment/v1"
	"payment/internal/biz"
)

// PaymentService 支付服务
type PaymentService struct {
	v1.UnimplementedPaymentServer
	uc  *biz.PaymentUsecase
	log *log.Helper
}

// NewPaymentService 创建支付服务
func NewPaymentService(uc *biz.PaymentUsecase, logger log.Logger) *PaymentService {
	return &PaymentService{
		uc:  uc,
		log: log.NewHelper(logger),
	}
}

// ListChannels 获取支付渠道列表
func (s *PaymentService) ListChannels(ctx context.Context, req *v1.ListChannelsRequest) (*v1.ListChannelsResponse, error) {
	channels, err := s.uc.ListChannels(ctx, req.Type, req.Status)
	if err != nil {
		return nil, err
	}

	var protoChannels []*v1.ChannelInfo
	for _, c := range channels {
		protoChannels = append(protoChannels, s.toProtoChannel(c))
	}

	return &v1.ListChannelsResponse{
		Channels: protoChannels,
	}, nil
}

// GetChannel 获取支付渠道
func (s *PaymentService) GetChannel(ctx context.Context, req *v1.GetChannelRequest) (*v1.ChannelInfo, error) {
	channel, err := s.uc.GetChannel(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return s.toProtoChannel(channel), nil
}

// CreateChannel 创建支付渠道
func (s *PaymentService) CreateChannel(ctx context.Context, req *v1.CreateChannelRequest) (*v1.ChannelInfo, error) {
	channel, err := s.uc.CreateChannel(ctx, &biz.PaymentChannel{
		Name:       req.Name,
		Code:       req.Code,
		Type:       req.Type,
		APIURL:     req.ApiUrl,
		APIKey:     req.ApiKey,
		APISecret:  req.ApiSecret,
		MerchantID: req.MerchantId,
		AppID:      req.AppId,
		NotifyURL:  req.NotifyUrl,
		ReturnURL:  req.ReturnUrl,
		MinAmount:  req.MinAmount,
		MaxAmount:  req.MaxAmount,
	})
	if err != nil {
		return nil, err
	}
	return s.toProtoChannel(channel), nil
}

// UpdateChannel 更新支付渠道
func (s *PaymentService) UpdateChannel(ctx context.Context, req *v1.UpdateChannelRequest) (*v1.ChannelInfo, error) {
	channel, err := s.uc.UpdateChannel(ctx, &biz.PaymentChannel{
		ID:     req.Id,
		Name:   req.Name,
		APIURL: req.ApiUrl,
		Status: req.Status,
	})
	if err != nil {
		return nil, err
	}
	return s.toProtoChannel(channel), nil
}

// DeleteChannel 删除支付渠道
func (s *PaymentService) DeleteChannel(ctx context.Context, req *v1.DeleteChannelRequest) (*v1.DeleteChannelResponse, error) {
	if err := s.uc.DeleteChannel(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.DeleteChannelResponse{Success: true}, nil
}

// ProcessPayment 处理支付
func (s *PaymentService) ProcessPayment(ctx context.Context, req *v1.ProcessPaymentRequest) (*v1.PaymentResponse, error) {
	result, err := s.uc.ProcessPayment(ctx, req.OrderNo, req.Amount, req.ChannelCode)
	if err != nil {
		return nil, err
	}
	return &v1.PaymentResponse{
		OrderNo: result.OrderNo,
		Amount:  result.Amount,
		Status:  result.Status,
		PayUrl:  result.PayURL,
	}, nil
}

// HandleCallback 处理支付回调
func (s *PaymentService) HandleCallback(ctx context.Context, req *v1.HandleCallbackRequest) (*v1.HandleCallbackResponse, error) {
	result, err := s.uc.HandleCallback(ctx, req.ChannelCode, req.OrderNo)
	if err != nil {
		return &v1.HandleCallbackResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}
	return &v1.HandleCallbackResponse{
		Success: result.Status == 1,
		Message: "处理成功",
	}, nil
}

func (s *PaymentService) toProtoChannel(c *biz.PaymentChannel) *v1.ChannelInfo {
	return &v1.ChannelInfo{
		Id:         c.ID,
		Name:       c.Name,
		Code:       c.Code,
		Type:       c.Type,
		ApiUrl:     c.APIURL,
		MerchantId: c.MerchantID,
		AppId:      c.AppID,
		NotifyUrl:  c.NotifyURL,
		ReturnUrl:  c.ReturnURL,
		MinAmount:  c.MinAmount,
		MaxAmount:  c.MaxAmount,
		Sort:       c.Sort,
		IsDefault:  c.IsDefault,
		Status:     c.Status,
	}
}
