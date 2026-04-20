package biz

import (
	"context"
	"fmt"
	"time"
)

// MockProvider 模拟短信服务商（用于测试）
type MockProvider struct {
	id       string
	name     string
	priority int32
	config   map[string]string
}

// NewMockProvider 创建模拟短信服务商
func NewMockProvider(id, name string, priority int32, config map[string]string) *MockProvider {
	return &MockProvider{
		id:       id,
		name:     name,
		priority: priority,
		config:   config,
	}
}

func (p *MockProvider) ID() string {
	return p.id
}

func (p *MockProvider) Name() string {
	return p.name
}

func (p *MockProvider) Type() string {
	return "mock"
}

func (p *MockProvider) IsAvailable() bool {
	return true
}

func (p *MockProvider) SendSms(ctx context.Context, req *SendRequest) (*SendResult, error) {
	// 模拟发送延迟
	time.Sleep(100 * time.Millisecond)

	// 生成模拟消息ID
	messageID := fmt.Sprintf("MOCK-%d", time.Now().UnixNano())

	return &SendResult{
		Success:      true,
		MessageID:    messageID,
		ProviderID:   p.id,
		ProviderName: p.name,
	}, nil
}

// MockProviderFactory 模拟短信服务商工厂
type MockProviderFactory struct{}

func (f *MockProviderFactory) CreateProvider(config map[string]string) (SmsProvider, error) {
	id := config["id"]
	if id == "" {
		id = "mock"
	}
	name := config["name"]
	if name == "" {
		name = "Mock SMS"
	}

	return NewMockProvider(id, name, 0, config), nil
}

// ConsoleProvider 控制台输出短信服务商（用于开发调试）
type ConsoleProvider struct {
	id       string
	name     string
	priority int32
}

// NewConsoleProvider 创建控制台短信服务商
func NewConsoleProvider(id, name string, priority int32) *ConsoleProvider {
	return &ConsoleProvider{
		id:       id,
		name:     name,
		priority: priority,
	}
}

func (p *ConsoleProvider) ID() string {
	return p.id
}

func (p *ConsoleProvider) Name() string {
	return p.name
}

func (p *ConsoleProvider) Type() string {
	return "console"
}

func (p *ConsoleProvider) IsAvailable() bool {
	return true
}

func (p *ConsoleProvider) SendSms(ctx context.Context, req *SendRequest) (*SendResult, error) {
	fmt.Printf("\n========== SMS [ConsoleProvider] ==========\n")
	fmt.Printf("To: %s\n", req.Phone)
	fmt.Printf("Template: %s\n", req.TemplateCode)
	fmt.Printf("Params: %v\n", req.TemplateParams)
	fmt.Printf("Content: %s\n", req.Content)
	fmt.Printf("Sign: %s\n", req.SignName)
	fmt.Printf("========================================\n\n")

	messageID := fmt.Sprintf("CONSOLE-%d", time.Now().UnixNano())

	return &SendResult{
		Success:      true,
		MessageID:    messageID,
		ProviderID:   p.id,
		ProviderName: p.name,
	}, nil
}

// ConsoleProviderFactory 控制台短信服务商工厂
type ConsoleProviderFactory struct{}

func (f *ConsoleProviderFactory) CreateProvider(config map[string]string) (SmsProvider, error) {
	id := config["id"]
	if id == "" {
		id = "console"
	}
	name := config["name"]
	if name == "" {
		name = "Console SMS"
	}

	return NewConsoleProvider(id, name, 0), nil
}
