package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TwilioProvider Twilio 短信服务商
type TwilioProvider struct {
	id       string
	name     string
	priority int32
	enabled  bool

	// Twilio 配置
	accountSid string
	authToken  string
	fromNumber string

	httpClient *http.Client
}

// NewTwilioProvider 创建 Twilio 短信服务商
func NewTwilioProvider(id, name string, priority int32, enabled bool, config map[string]string) (*TwilioProvider, error) {
	accountSid := config["account_sid"]
	authToken := config["auth_token"]
	fromNumber := config["from_number"]

	if accountSid == "" || authToken == "" || fromNumber == "" {
		return nil, fmt.Errorf("twilio: account_sid, auth_token and from_number are required")
	}

	return &TwilioProvider{
		id:         id,
		name:       name,
		priority:   priority,
		enabled:    enabled,
		accountSid: accountSid,
		authToken:  authToken,
		fromNumber: fromNumber,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (p *TwilioProvider) ID() string {
	return p.id
}

func (p *TwilioProvider) Name() string {
	return p.name
}

func (p *TwilioProvider) Type() string {
	return "twilio"
}

func (p *TwilioProvider) IsAvailable() bool {
	return p.enabled && p.accountSid != "" && p.authToken != ""
}

func (p *TwilioProvider) SendSms(ctx context.Context, req *SendRequest) (*SendResult, error) {
	// 构建 Twilio API URL
	apiURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", p.accountSid)

	// 准备请求参数
	data := url.Values{}
	data.Set("To", req.Phone)
	data.Set("From", p.fromNumber)

	// 如果有模板参数，格式化内容
	if req.Content != "" {
		data.Set("Body", req.Content)
	} else {
		data.Set("Body", fmt.Sprintf("Template: %s, Params: %v", req.TemplateCode, req.TemplateParams))
	}

	// 创建 HTTP 请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return &SendResult{
			Success:      false,
			ProviderID:   p.id,
			ProviderName: p.name,
			ErrorMessage: err.Error(),
		}, err
	}

	// 设置认证和头部
	httpReq.SetBasicAuth(p.accountSid, p.authToken)
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 发送请求
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return &SendResult{
			Success:      false,
			ProviderID:   p.id,
			ProviderName: p.name,
			ErrorMessage: err.Error(),
		}, err
	}
	defer resp.Body.Close()

	// 解析响应
	var result struct {
		SID          string `json:"sid"`
		Status       string `json:"status"`
		ErrorMessage string `json:"error_message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &SendResult{
			Success:      false,
			ProviderID:   p.id,
			ProviderName: p.name,
			ErrorMessage: err.Error(),
		}, err
	}

	// 检查响应状态
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return &SendResult{
			Success:      false,
			ProviderID:   p.id,
			ProviderName: p.name,
			ErrorMessage: result.ErrorMessage,
		}, fmt.Errorf("twilio api error: %s", result.ErrorMessage)
	}

	return &SendResult{
		Success:      true,
		MessageID:    result.SID,
		ProviderID:   p.id,
		ProviderName: p.name,
	}, nil
}

// TwilioProviderFactory Twilio 短信服务商工厂
type TwilioProviderFactory struct{}

func (f *TwilioProviderFactory) CreateProvider(config map[string]string) (SmsProvider, error) {
	id := config["id"]
	if id == "" {
		id = "twilio"
	}
	name := config["name"]
	if name == "" {
		name = "Twilio"
	}

	// 解析优先级
	priority := int32(0)
	if p := config["priority"]; p != "" {
		fmt.Sscanf(p, "%d", &priority)
	}

	// 解析启用状态
	enabled := true
	if e := config["enabled"]; e == "false" {
		enabled = false
	}

	return NewTwilioProvider(id, name, priority, enabled, config)
}
