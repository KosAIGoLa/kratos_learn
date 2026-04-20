package biz

import (
	"context"
	"fmt"
)

// SmsProvider 短信服务商接口
type SmsProvider interface {
	// ID 返回服务商唯一标识
	ID() string
	// Name 返回服务商名称
	Name() string
	// Type 返回服务商类型
	Type() string
	// SendSms 发送短信
	SendSms(ctx context.Context, req *SendRequest) (*SendResult, error)
	// IsAvailable 检查服务商是否可用
	IsAvailable() bool
}

// SendRequest 短信发送请求
type SendRequest struct {
	Phone          string
	TemplateCode   string
	TemplateParams map[string]string
	Content        string
	SignName       string
}

// SendResult 短信发送结果
type SendResult struct {
	Success      bool
	MessageID    string
	ProviderID   string
	ProviderName string
	ErrorMessage string
}

// ProviderFactory 短信服务商工厂
type ProviderFactory interface {
	CreateProvider(config map[string]string) (SmsProvider, error)
}

// ProviderRegistry 服务商注册表
type ProviderRegistry struct {
	factories map[string]ProviderFactory
}

// NewProviderRegistry 创建服务商注册表
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		factories: make(map[string]ProviderFactory),
	}
}

// Register 注册服务商工厂
func (r *ProviderRegistry) Register(providerType string, factory ProviderFactory) {
	r.factories[providerType] = factory
}

// Create 创建服务商实例
func (r *ProviderRegistry) Create(providerType string, config map[string]string) (SmsProvider, error) {
	factory, ok := r.factories[providerType]
	if !ok {
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}
	return factory.CreateProvider(config)
}

// ProviderSelector 服务商选择器
type ProviderSelector interface {
	// Select 根据策略选择一个服务商
	Select(providers []SmsProvider) (SmsProvider, error)
}

// DefaultProviderSelector 默认服务商选择器（按优先级选择第一个可用的）
type DefaultProviderSelector struct{}

// Select 选择服务商
func (s *DefaultProviderSelector) Select(providers []SmsProvider) (SmsProvider, error) {
	for _, p := range providers {
		if p.IsAvailable() {
			return p, nil
		}
	}
	return nil, fmt.Errorf("no available SMS provider")
}

// NamedProviderSelector 按名称选择指定服务商
type NamedProviderSelector struct {
	ProviderID string
}

// Select 选择指定名称的服务商
func (s *NamedProviderSelector) Select(providers []SmsProvider) (SmsProvider, error) {
	for _, p := range providers {
		if p.ID() == s.ProviderID && p.IsAvailable() {
			return p, nil
		}
	}
	return nil, fmt.Errorf("provider %s not found or not available", s.ProviderID)
}

// RoundRobinProviderSelector 轮询选择器
type RoundRobinProviderSelector struct {
	currentIndex int
}

// Select 轮询选择服务商
func (s *RoundRobinProviderSelector) Select(providers []SmsProvider) (SmsProvider, error) {
	availableProviders := make([]SmsProvider, 0)
	for _, p := range providers {
		if p.IsAvailable() {
			availableProviders = append(availableProviders, p)
		}
	}

	if len(availableProviders) == 0 {
		return nil, fmt.Errorf("no available SMS provider")
	}

	provider := availableProviders[s.currentIndex%len(availableProviders)]
	s.currentIndex++
	return provider, nil
}
