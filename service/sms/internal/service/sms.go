package service

import (
	"context"

	v1 "sms/api/sms/v1"
	"sms/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SmsService 短信服务实现
type SmsService struct {
	v1.UnimplementedSmsServiceServer
	uc  *biz.SmsUsecase
	log *log.Helper
}

// NewSmsService 创建短信服务实例
func NewSmsService(uc *biz.SmsUsecase, logger log.Logger) *SmsService {
	return &SmsService{
		uc:  uc,
		log: log.NewHelper(logger),
	}
}

// SendSms 发送短信
func (s *SmsService) SendSms(ctx context.Context, req *v1.SendSmsRequest) (*v1.SendSmsReply, error) {
	s.log.Infof("Sending SMS to %s via provider %s", req.Phone, req.ProviderId)

	// 构建发送请求
	sendReq := &biz.SendRequest{
		Phone:          req.Phone,
		TemplateCode:   req.TemplateCode,
		TemplateParams: req.TemplateParams,
		SignName:       req.SignName,
	}

	// 发送短信
	result, err := s.uc.SendSms(ctx, sendReq, req.ProviderId)
	if err != nil {
		s.log.Errorf("Failed to send SMS: %v", err)
		return &v1.SendSmsReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.SendSmsReply{
		Success:           result.Success,
		Id:                result.MessageID,
		ProviderMessageId: result.MessageID,
		ProviderId:        result.ProviderID,
		Message:           "SMS sent successfully",
	}, nil
}

// BatchSendSms 批量发送短信
func (s *SmsService) BatchSendSms(ctx context.Context, req *v1.BatchSendSmsRequest) (*v1.BatchSendSmsReply, error) {
	s.log.Infof("Batch sending SMS to %d phones via provider %s", len(req.Phones), req.ProviderId)

	// 构建发送请求
	sendReq := &biz.SendRequest{
		TemplateCode:   req.TemplateCode,
		TemplateParams: req.TemplateParams,
		SignName:       req.SignName,
	}

	// 批量发送
	results, err := s.uc.BatchSendSms(ctx, req.Phones, sendReq, req.ProviderId)
	if err != nil {
		s.log.Errorf("Failed to batch send SMS: %v", err)
		return nil, err
	}

	// 构建响应
	reply := &v1.BatchSendSmsReply{
		Results: make([]*v1.BatchSendSmsReply_Result, 0, len(results)),
	}

	for i, result := range results {
		if i < len(req.Phones) {
			resultItem := &v1.BatchSendSmsReply_Result{
				Phone:     req.Phones[i],
				Success:   result != nil && result.Success,
				MessageId: "",
			}
			if result != nil {
				resultItem.MessageId = result.MessageID
				if !result.Success {
					resultItem.Error = result.ErrorMessage
				}
			}
			reply.Results = append(reply.Results, resultItem)

			if resultItem.Success {
				reply.SuccessCount++
			} else {
				reply.FailCount++
			}
		}
	}

	return reply, nil
}

// GetSmsLog 获取短信发送记录
func (s *SmsService) GetSmsLog(ctx context.Context, req *v1.GetSmsLogRequest) (*v1.SmsLog, error) {
	log, err := s.uc.GetSmsLog(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	if log == nil {
		return nil, nil // 或者返回 not found 错误
	}

	return s.toProtoLog(log), nil
}

// ListSmsLogs 查询短信发送记录列表
func (s *SmsService) ListSmsLogs(ctx context.Context, req *v1.ListSmsLogsRequest) (*v1.ListSmsLogsReply, error) {
	logs, total, err := s.uc.ListSmsLogs(ctx, req.Phone, req.Status, req.ProviderId, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	protoLogs := make([]*v1.SmsLog, 0, len(logs))
	for _, log := range logs {
		protoLogs = append(protoLogs, s.toProtoLog(log))
	}

	return &v1.ListSmsLogsReply{
		Logs:  protoLogs,
		Total: total,
	}, nil
}

// GetProviders 获取所有短信服务商
func (s *SmsService) GetProviders(ctx context.Context, req *v1.GetProvidersRequest) (*v1.GetProvidersReply, error) {
	providers := s.uc.GetProviders()
	defaultProvider := s.uc.GetDefaultProvider()

	protoProviders := make([]*v1.SmsProviderInfo, 0, len(providers))
	for _, p := range providers {
		info := &v1.SmsProviderInfo{
			Id:        p.ID(),
			Name:      p.Name(),
			Type:      p.Type(),
			Enabled:   p.IsAvailable(),
			IsDefault: p.ID() == defaultProvider,
		}
		protoProviders = append(protoProviders, info)
	}

	return &v1.GetProvidersReply{
		Providers: protoProviders,
	}, nil
}

// SwitchProvider 切换默认短信服务商
func (s *SmsService) SwitchProvider(ctx context.Context, req *v1.SwitchProviderRequest) (*v1.SwitchProviderReply, error) {
	if err := s.uc.SwitchProvider(req.ProviderId); err != nil {
		return &v1.SwitchProviderReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 获取切换后的服务商信息
	providers := s.uc.GetProviders()
	var currentProvider *v1.SmsProviderInfo
	for _, p := range providers {
		if p.ID() == req.ProviderId {
			currentProvider = &v1.SmsProviderInfo{
				Id:        p.ID(),
				Name:      p.Name(),
				Type:      p.Type(),
				Enabled:   p.IsAvailable(),
				IsDefault: true,
			}
			break
		}
	}

	return &v1.SwitchProviderReply{
		Success:         true,
		Message:         "Provider switched successfully",
		CurrentProvider: currentProvider,
	}, nil
}

// toProtoLog 转换为 proto 对象
func (s *SmsService) toProtoLog(log *biz.SmsLog) *v1.SmsLog {
	return &v1.SmsLog{
		Id:                log.ID,
		Phone:             log.Phone,
		TemplateCode:      log.TemplateCode,
		TemplateParams:    log.TemplateParams,
		Content:           log.Content,
		ProviderId:        log.ProviderID,
		ProviderName:      log.ProviderName,
		Status:            log.Status,
		ProviderMessageId: log.ProviderMessageID,
		ErrorMessage:      log.ErrorMessage,
		RetryCount:        log.RetryCount,
		CreatedAt:         timestamppb.New(log.CreatedAt).String(),
		UpdatedAt:         timestamppb.New(log.UpdatedAt).String(),
	}
}
