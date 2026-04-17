package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// SystemConfig 系统配置领域模型
type SystemConfig struct {
	ID          uint32
	Key         string
	Value       string
	Description string
	Group       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ProfitSharingRule 分账规则领域模型
type ProfitSharingRule struct {
	ID          uint32
	Name        string
	Type        string
	Level       int32
	Ratio       float64
	FixedAmount float64
	Unit        string
	MinAmount   float64
	MaxAmount   float64
	Sort        int32
	Enabled     int32
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// RiskControl 风控规则领域模型
type RiskControl struct {
	ID               uint32
	Name             string
	Code             string
	Type             string
	Level            string
	TriggerCondition string
	Action           string
	LimitValue       uint32
	TimeWindow       uint32
	Enabled          int32
	Description      string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// Domain 域名领域模型
type Domain struct {
	ID        uint32
	Domain    string
	Enabled   int32
	CreatedAt time.Time
	UpdatedAt time.Time
}

// WhitelistIP IP白名单领域模型
type WhitelistIP struct {
	ID          uint32
	IP          string
	Type        string
	Description string
	Enabled     int32
	ExpireAt    *time.Time
	CreatedAt   time.Time
}

// RiskCheckResult 风控检查结果
type RiskCheckResult struct {
	Pass    bool
	Level   string
	Action  string
	Message string
}

// ConfigRepo 配置存储接口
type ConfigRepo interface {
	GetConfig(ctx context.Context, key string) (*SystemConfig, error)
	SetConfig(ctx context.Context, c *SystemConfig) (*SystemConfig, error)
	ListConfigs(ctx context.Context, group string) ([]*SystemConfig, error)
}

// ProfitSharingRepo 分账规则存储接口
type ProfitSharingRepo interface {
	ListRules(ctx context.Context, typ string, enabled int32) ([]*ProfitSharingRule, error)
	GetRule(ctx context.Context, id uint32) (*ProfitSharingRule, error)
	CreateRule(ctx context.Context, r *ProfitSharingRule) (*ProfitSharingRule, error)
	UpdateRule(ctx context.Context, r *ProfitSharingRule) (*ProfitSharingRule, error)
	DeleteRule(ctx context.Context, id uint32) error
}

// RiskControlRepo 风控存储接口
type RiskControlRepo interface {
	ListRiskControls(ctx context.Context, typ string, enabled int32) ([]*RiskControl, error)
	GetRiskControl(ctx context.Context, id uint32) (*RiskControl, error)
	CreateRiskControl(ctx context.Context, r *RiskControl) (*RiskControl, error)
	UpdateRiskControl(ctx context.Context, r *RiskControl) (*RiskControl, error)
}

// DomainRepo 域名存储接口
type DomainRepo interface {
	ListDomains(ctx context.Context, enabled int32) ([]*Domain, error)
	AddDomain(ctx context.Context, d *Domain) (*Domain, error)
}

// WhitelistRepo IP白名单存储接口
type WhitelistRepo interface {
	CheckWhitelist(ctx context.Context, ip, typ string) (bool, error)
}

// SystemUsecase 系统用例
type SystemUsecase struct {
	configRepo        ConfigRepo
	profitSharingRepo ProfitSharingRepo
	riskControlRepo   RiskControlRepo
	domainRepo        DomainRepo
	whitelistRepo     WhitelistRepo
	log               *log.Helper
}

// NewSystemUsecase 创建系统用例
func NewSystemUsecase(
	configRepo ConfigRepo,
	profitSharingRepo ProfitSharingRepo,
	riskControlRepo RiskControlRepo,
	domainRepo DomainRepo,
	whitelistRepo WhitelistRepo,
	logger log.Logger,
) *SystemUsecase {
	return &SystemUsecase{
		configRepo:        configRepo,
		profitSharingRepo: profitSharingRepo,
		riskControlRepo:   riskControlRepo,
		domainRepo:        domainRepo,
		whitelistRepo:     whitelistRepo,
		log:               log.NewHelper(logger),
	}
}

// GetConfig 获取配置
func (uc *SystemUsecase) GetConfig(ctx context.Context, key string) (*SystemConfig, error) {
	return uc.configRepo.GetConfig(ctx, key)
}

// SetConfig 设置配置
func (uc *SystemUsecase) SetConfig(ctx context.Context, c *SystemConfig) (*SystemConfig, error) {
	return uc.configRepo.SetConfig(ctx, c)
}

// ListConfigs 获取配置列表
func (uc *SystemUsecase) ListConfigs(ctx context.Context, group string) ([]*SystemConfig, error) {
	return uc.configRepo.ListConfigs(ctx, group)
}

// GetProfitSharingRules 获取分账规则列表
func (uc *SystemUsecase) GetProfitSharingRules(ctx context.Context, typ string, enabled int32) ([]*ProfitSharingRule, error) {
	return uc.profitSharingRepo.ListRules(ctx, typ, enabled)
}

// CreateProfitSharingRule 创建分账规则
func (uc *SystemUsecase) CreateProfitSharingRule(ctx context.Context, r *ProfitSharingRule) (*ProfitSharingRule, error) {
	return uc.profitSharingRepo.CreateRule(ctx, r)
}

// UpdateProfitSharingRule 更新分账规则
func (uc *SystemUsecase) UpdateProfitSharingRule(ctx context.Context, r *ProfitSharingRule) (*ProfitSharingRule, error) {
	return uc.profitSharingRepo.UpdateRule(ctx, r)
}

// DeleteProfitSharingRule 删除分账规则
func (uc *SystemUsecase) DeleteProfitSharingRule(ctx context.Context, id uint32) error {
	return uc.profitSharingRepo.DeleteRule(ctx, id)
}

// GetRiskControls 获取风控规则列表
func (uc *SystemUsecase) GetRiskControls(ctx context.Context, typ string, enabled int32) ([]*RiskControl, error) {
	return uc.riskControlRepo.ListRiskControls(ctx, typ, enabled)
}

// CreateRiskControl 创建风控规则
func (uc *SystemUsecase) CreateRiskControl(ctx context.Context, r *RiskControl) (*RiskControl, error) {
	return uc.riskControlRepo.CreateRiskControl(ctx, r)
}

// UpdateRiskControl 更新风控规则
func (uc *SystemUsecase) UpdateRiskControl(ctx context.Context, r *RiskControl) (*RiskControl, error) {
	return uc.riskControlRepo.UpdateRiskControl(ctx, r)
}

// ListDomains 获取域名列表
func (uc *SystemUsecase) ListDomains(ctx context.Context, enabled int32) ([]*Domain, error) {
	return uc.domainRepo.ListDomains(ctx, enabled)
}

// AddDomain 添加域名
func (uc *SystemUsecase) AddDomain(ctx context.Context, d *Domain) (*Domain, error) {
	return uc.domainRepo.AddDomain(ctx, d)
}

// CheckRisk 检查风险
func (uc *SystemUsecase) CheckRisk(ctx context.Context, typ string, userID uint32, ip string) (*RiskCheckResult, error) {
	// TODO: 实现风控检查逻辑
	return &RiskCheckResult{
		Pass:    true,
		Level:   "low",
		Action:  "allow",
		Message: "通过",
	}, nil
}

// CheckWhitelist 检查IP白名单
func (uc *SystemUsecase) CheckWhitelist(ctx context.Context, ip, typ string) (bool, error) {
	return uc.whitelistRepo.CheckWhitelist(ctx, ip, typ)
}
