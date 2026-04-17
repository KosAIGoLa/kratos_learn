package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	v1 "system/api/system/v1"
	"system/internal/biz"
)

// SystemService 系统服务
type SystemService struct {
	v1.UnimplementedSystemServer
	uc  *biz.SystemUsecase
	log *log.Helper
}

// NewSystemService 创建系统服务
func NewSystemService(uc *biz.SystemUsecase, logger log.Logger) *SystemService {
	return &SystemService{
		uc:  uc,
		log: log.NewHelper(logger),
	}
}

// GetConfig 获取配置
func (s *SystemService) GetConfig(ctx context.Context, req *v1.GetConfigRequest) (*v1.ConfigInfo, error) {
	config, err := s.uc.GetConfig(ctx, req.Key)
	if err != nil {
		return nil, err
	}
	return &v1.ConfigInfo{
		Id:          config.ID,
		Key:         config.Key,
		Value:       config.Value,
		Description: config.Description,
		Group:       config.Group,
	}, nil
}

// SetConfig 设置配置
func (s *SystemService) SetConfig(ctx context.Context, req *v1.SetConfigRequest) (*v1.ConfigInfo, error) {
	config, err := s.uc.SetConfig(ctx, &biz.SystemConfig{
		Key:         req.Key,
		Value:       req.Value,
		Description: req.Description,
		Group:       req.Group,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ConfigInfo{
		Id:          config.ID,
		Key:         config.Key,
		Value:       config.Value,
		Description: config.Description,
		Group:       config.Group,
	}, nil
}

// GetProfitSharingRules 获取分账规则
func (s *SystemService) GetProfitSharingRules(ctx context.Context, req *v1.GetProfitSharingRulesRequest) (*v1.GetProfitSharingRulesResponse, error) {
	rules, err := s.uc.GetProfitSharingRules(ctx, req.Type, req.Enabled)
	if err != nil {
		return nil, err
	}

	var protoRules []*v1.ProfitSharingRuleInfo
	for _, r := range rules {
		protoRules = append(protoRules, &v1.ProfitSharingRuleInfo{
			Id:          r.ID,
			Name:        r.Name,
			Type:        r.Type,
			Level:       r.Level,
			Ratio:       r.Ratio,
			FixedAmount: r.FixedAmount,
			Unit:        r.Unit,
			MinAmount:   r.MinAmount,
			MaxAmount:   r.MaxAmount,
			Sort:        r.Sort,
			Enabled:     r.Enabled,
			Description: r.Description,
		})
	}

	return &v1.GetProfitSharingRulesResponse{
		Rules: protoRules,
	}, nil
}

// CheckRisk 风控检查
func (s *SystemService) CheckRisk(ctx context.Context, req *v1.CheckRiskRequest) (*v1.CheckRiskResponse, error) {
	result, err := s.uc.CheckRisk(ctx, req.Type, req.UserId, req.Ip)
	if err != nil {
		return nil, err
	}
	return &v1.CheckRiskResponse{
		Pass:    result.Pass,
		Level:   result.Level,
		Action:  result.Action,
		Message: result.Message,
	}, nil
}

// GetDomains 获取域名列表
func (s *SystemService) GetDomains(ctx context.Context, req *v1.GetDomainsRequest) (*v1.GetDomainsResponse, error) {
	domains, err := s.uc.ListDomains(ctx, req.Enabled)
	if err != nil {
		return nil, err
	}

	var protoDomains []*v1.DomainInfo
	for _, d := range domains {
		protoDomains = append(protoDomains, &v1.DomainInfo{
			Id:      d.ID,
			Domain:  d.Domain,
			Enabled: d.Enabled,
		})
	}

	return &v1.GetDomainsResponse{
		Domains: protoDomains,
	}, nil
}

// CheckWhitelist 检查IP白名单
func (s *SystemService) CheckWhitelist(ctx context.Context, req *v1.CheckWhitelistRequest) (*v1.CheckWhitelistResponse, error) {
	whitelisted, err := s.uc.CheckWhitelist(ctx, req.Ip, req.Type)
	if err != nil {
		return nil, err
	}
	return &v1.CheckWhitelistResponse{
		Whitelisted: whitelisted,
	}, nil
}
