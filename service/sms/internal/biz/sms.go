package biz

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"sms/internal/conf"
)

// SmsUsecase 短信业务逻辑层
type SmsUsecase struct {
	providers       map[string]SmsProvider
	providerList    []SmsProvider
	defaultProvider string
	retryTimes      int32
	enableLog       bool

	mu   sync.RWMutex
	repo SmsRepo
	log  *log.Helper
}

// SmsRepo 短信记录存储接口
type SmsRepo interface {
	SaveLog(ctx context.Context, log *SmsLog) error
	GetLog(ctx context.Context, id string) (*SmsLog, error)
	ListLogs(ctx context.Context, phone, status, providerId string, page, pageSize int32) ([]*SmsLog, int32, error)
}

// SmsLog 短信发送记录
type SmsLog struct {
	ID                string
	Phone             string
	TemplateCode      string
	TemplateParams    map[string]string
	Content           string
	ProviderID        string
	ProviderName      string
	Status            string // pending, sent, failed, delivered
	ProviderMessageID string
	ErrorMessage      string
	RetryCount        int32
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// NewSmsUsecase 创建短信业务逻辑实例
func NewSmsUsecase(repo SmsRepo, c *conf.Sms, logger log.Logger) (*SmsUsecase, error) {
	uc := &SmsUsecase{
		providers:       make(map[string]SmsProvider),
		providerList:    make([]SmsProvider, 0),
		defaultProvider: c.DefaultProvider,
		retryTimes:      c.RetryTimes,
		enableLog:       c.EnableLog,
		repo:            repo,
		log:             log.NewHelper(logger),
	}

	// 注册服务商工厂
	registry := NewProviderRegistry()
	registry.Register("mock", &MockProviderFactory{})
	registry.Register("console", &ConsoleProviderFactory{})
	registry.Register("twilio", &TwilioProviderFactory{})
	registry.Register("aliyun", &AliyunProviderFactory{})

	// 初始化配置的服务商
	for _, pc := range c.Providers {
		if !pc.Enabled {
			continue
		}

		provider, err := registry.Create(pc.Type, pc.Config)
		if err != nil {
			uc.log.Warnf("Failed to create provider %s: %v", pc.Id, err)
			continue
		}

		// 使用配置中的 ID 和名称覆盖
		if pc.Id != "" {
			// 注意：这里 provider 的 ID 已经在工厂中设置
			_ = provider
		}

		uc.providers[provider.ID()] = provider
		uc.providerList = append(uc.providerList, provider)
		uc.log.Infof("Registered SMS provider: %s (%s)", provider.ID(), provider.Name())
	}

	// 按优先级排序服务商列表
	sort.Slice(uc.providerList, func(i, j int) bool {
		return getProviderPriority(uc.providerList[i]) < getProviderPriority(uc.providerList[j])
	})

	if len(uc.providers) == 0 {
		return nil, fmt.Errorf("no SMS provider configured")
	}

	return uc, nil
}

// SendSms 发送短信
func (uc *SmsUsecase) SendSms(ctx context.Context, req *SendRequest, providerId string) (*SendResult, error) {
	// 选择服务商
	provider, err := uc.selectProvider(providerId)
	if err != nil {
		return nil, err
	}

	// 创建发送记录
	smsLog := &SmsLog{
		ID:             generateID(),
		Phone:          req.Phone,
		TemplateCode:   req.TemplateCode,
		TemplateParams: req.TemplateParams,
		Content:        req.Content,
		ProviderID:     provider.ID(),
		ProviderName:   provider.Name(),
		Status:         "pending",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 保存记录
	if uc.enableLog {
		if err := uc.repo.SaveLog(ctx, smsLog); err != nil {
			uc.log.Warnf("Failed to save SMS log: %v", err)
		}
	}

	// 发送短信
	result, err := uc.sendWithRetry(ctx, provider, req, smsLog)

	// 更新记录状态
	smsLog.Status = "sent"
	if err != nil {
		smsLog.Status = "failed"
		smsLog.ErrorMessage = err.Error()
	}
	if result != nil {
		smsLog.ProviderMessageID = result.MessageID
	}
	smsLog.UpdatedAt = time.Now()

	if uc.enableLog {
		if err := uc.repo.SaveLog(ctx, smsLog); err != nil {
			uc.log.Warnf("Failed to update SMS log: %v", err)
		}
	}

	return result, err
}

// sendWithRetry 带重试的发送
func (uc *SmsUsecase) sendWithRetry(ctx context.Context, provider SmsProvider, req *SendRequest, log *SmsLog) (*SendResult, error) {
	var lastErr error
	var result *SendResult

	for i := int32(0); i <= uc.retryTimes; i++ {
		if i > 0 {
			uc.log.Infof("Retrying SMS send to %s, attempt %d", req.Phone, i)
			time.Sleep(time.Second * time.Duration(i)) // 指数退避
		}

		result, lastErr = provider.SendSms(ctx, req)
		if lastErr == nil && result.Success {
			return result, nil
		}

		log.RetryCount = i + 1
	}

	return result, fmt.Errorf("failed after %d retries: %w", uc.retryTimes, lastErr)
}

// BatchSendSms 批量发送短信
func (uc *SmsUsecase) BatchSendSms(ctx context.Context, phones []string, req *SendRequest, providerId string) ([]*SendResult, error) {
	results := make([]*SendResult, 0, len(phones))

	for _, phone := range phones {
		singleReq := &SendRequest{
			Phone:          phone,
			TemplateCode:   req.TemplateCode,
			TemplateParams: req.TemplateParams,
			Content:        req.Content,
			SignName:       req.SignName,
		}

		result, err := uc.SendSms(ctx, singleReq, providerId)
		if err != nil {
			uc.log.Errorf("Failed to send SMS to %s: %v", phone, err)
		}
		results = append(results, result)
	}

	return results, nil
}

// selectProvider 选择短信服务商
func (uc *SmsUsecase) selectProvider(providerId string) (SmsProvider, error) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	// 如果指定了服务商ID，使用指定的
	if providerId != "" {
		provider, ok := uc.providers[providerId]
		if !ok {
			return nil, fmt.Errorf("provider %s not found", providerId)
		}
		if !provider.IsAvailable() {
			return nil, fmt.Errorf("provider %s is not available", providerId)
		}
		return provider, nil
	}

	// 使用默认服务商
	if uc.defaultProvider != "" {
		provider, ok := uc.providers[uc.defaultProvider]
		if ok && provider.IsAvailable() {
			return provider, nil
		}
	}

	// 选择第一个可用的服务商
	selector := &DefaultProviderSelector{}
	return selector.Select(uc.providerList)
}

// GetProviders 获取所有服务商
func (uc *SmsUsecase) GetProviders() []SmsProvider {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	providers := make([]SmsProvider, 0, len(uc.providerList))
	providers = append(providers, uc.providerList...)
	return providers
}

// GetDefaultProvider 获取默认服务商
func (uc *SmsUsecase) GetDefaultProvider() string {
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	return uc.defaultProvider
}

// SwitchProvider 切换默认服务商
func (uc *SmsUsecase) SwitchProvider(providerId string) error {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	provider, ok := uc.providers[providerId]
	if !ok {
		return fmt.Errorf("provider %s not found", providerId)
	}
	if !provider.IsAvailable() {
		return fmt.Errorf("provider %s is not available", providerId)
	}

	uc.defaultProvider = providerId
	uc.log.Infof("Switched default SMS provider to: %s", providerId)
	return nil
}

// GetSmsLog 获取短信发送记录
func (uc *SmsUsecase) GetSmsLog(ctx context.Context, id string) (*SmsLog, error) {
	return uc.repo.GetLog(ctx, id)
}

// ListSmsLogs 查询短信发送记录
func (uc *SmsUsecase) ListSmsLogs(ctx context.Context, phone, status, providerId string, page, pageSize int32) ([]*SmsLog, int32, error) {
	return uc.repo.ListLogs(ctx, phone, status, providerId, page, pageSize)
}

// getProviderPriority 获取服务商优先级
func getProviderPriority(p SmsProvider) int32 {
	// 这里可以通过接口扩展获取优先级
	// 简化实现，直接返回0
	return 0
}

// generateID 生成唯一ID
func generateID() string {
	return fmt.Sprintf("SMS%d", time.Now().UnixNano())
}
