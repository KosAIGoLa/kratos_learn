package biz

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// AliyunProvider 阿里云短信服务商
type AliyunProvider struct {
	id       string
	name     string
	priority int32
	enabled  bool

	// 阿里云配置
	accessKeyId     string
	accessKeySecret string
	signName        string
	endpoint        string

	httpClient *http.Client
}

// NewAliyunProvider 创建阿里云短信服务商
func NewAliyunProvider(id, name string, priority int32, enabled bool, config map[string]string) (*AliyunProvider, error) {
	accessKeyId := config["access_key_id"]
	accessKeySecret := config["access_key_secret"]
	signName := config["sign_name"]

	if accessKeyId == "" || accessKeySecret == "" {
		return nil, fmt.Errorf("aliyun: access_key_id and access_key_secret are required")
	}

	endpoint := config["endpoint"]
	if endpoint == "" {
		endpoint = "dysmsapi.aliyuncs.com"
	}

	return &AliyunProvider{
		id:              id,
		name:            name,
		priority:        priority,
		enabled:         enabled,
		accessKeyId:     accessKeyId,
		accessKeySecret: accessKeySecret,
		signName:        signName,
		endpoint:        endpoint,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (p *AliyunProvider) ID() string {
	return p.id
}

func (p *AliyunProvider) Name() string {
	return p.name
}

func (p *AliyunProvider) Type() string {
	return "aliyun"
}

func (p *AliyunProvider) IsAvailable() bool {
	return p.enabled && p.accessKeyId != "" && p.accessKeySecret != ""
}

func (p *AliyunProvider) SendSms(ctx context.Context, req *SendRequest) (*SendResult, error) {
	// 构建请求参数
	params := map[string]string{
		"AccessKeyId":      p.accessKeyId,
		"Action":           "SendSms",
		"PhoneNumbers":     req.Phone,
		"SignName":         p.getSignName(req.SignName),
		"TemplateCode":     req.TemplateCode,
		"Format":           "JSON",
		"Version":          "2017-05-25",
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureVersion": "1.0",
		"SignatureNonce":   fmt.Sprintf("%d", time.Now().UnixNano()),
	}

	// 序列化模板参数
	if len(req.TemplateParams) > 0 {
		templateParamJson, _ := json.Marshal(req.TemplateParams)
		params["TemplateParam"] = string(templateParamJson)
	}

	// 计算签名
	signature := p.calculateSignature(params)
	params["Signature"] = signature

	// 构建请求 URL
	apiURL := fmt.Sprintf("https://%s/?%s", p.endpoint, p.encodeParams(params))

	// 创建 HTTP 请求
	httpReq, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return &SendResult{
			Success:      false,
			ProviderID:   p.id,
			ProviderName: p.name,
			ErrorMessage: err.Error(),
		}, err
	}

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
	defer func() {
		_ = resp.Body.Close()
	}()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &SendResult{
			Success:      false,
			ProviderID:   p.id,
			ProviderName: p.name,
			ErrorMessage: err.Error(),
		}, err
	}

	// 解析响应
	var result struct {
		Code      string `json:"Code"`
		Message   string `json:"Message"`
		RequestId string `json:"RequestId"`
		BizId     string `json:"BizId"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return &SendResult{
			Success:      false,
			ProviderID:   p.id,
			ProviderName: p.name,
			ErrorMessage: err.Error(),
		}, err
	}

	// 检查响应状态
	if result.Code != "OK" {
		return &SendResult{
			Success:      false,
			ProviderID:   p.id,
			ProviderName: p.name,
			ErrorMessage: result.Message,
		}, fmt.Errorf("aliyun sms error: %s", result.Message)
	}

	return &SendResult{
		Success:      true,
		MessageID:    result.BizId,
		ProviderID:   p.id,
		ProviderName: p.name,
	}, nil
}

// getSignName 获取短信签名
func (p *AliyunProvider) getSignName(signName string) string {
	if signName != "" {
		return signName
	}
	return p.signName
}

// calculateSignature 计算阿里云签名
func (p *AliyunProvider) calculateSignature(params map[string]string) string {
	// 排序参数
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建签名字符串
	var signParts []string
	for _, k := range keys {
		signParts = append(signParts, fmt.Sprintf("%s=%s", k, params[k]))
	}
	signString := strings.Join(signParts, "&")

	// 添加 HTTP 方法
	signString = "GET&%2F&" + url.QueryEscape(signString)

	// 计算 HMAC-SHA1
	h := hmac.New(sha1.New, []byte(p.accessKeySecret+"&"))
	h.Write([]byte(signString))

	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// encodeParams URL 编码参数
func (p *AliyunProvider) encodeParams(params map[string]string) string {
	var parts []string
	for k, v := range params {
		parts = append(parts, fmt.Sprintf("%s=%s", k, url.QueryEscape(v)))
	}
	return strings.Join(parts, "&")
}

// AliyunProviderFactory 阿里云短信服务商工厂
type AliyunProviderFactory struct{}

func (f *AliyunProviderFactory) CreateProvider(config map[string]string) (SmsProvider, error) {
	id := config["id"]
	if id == "" {
		id = "aliyun"
	}
	name := config["name"]
	if name == "" {
		name = "阿里云短信"
	}

	// 解析优先级
	priority := int32(0)
	if p := config["priority"]; p != "" {
		parsed, err := strconv.ParseInt(p, 10, 32)
		if err == nil {
			priority = int32(parsed)
		}
	}

	// 解析启用状态
	enabled := true
	if e := config["enabled"]; e == "false" {
		enabled = false
	}

	return NewAliyunProvider(id, name, priority, enabled, config)
}
