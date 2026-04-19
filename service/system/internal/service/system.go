package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/timestamppb"
	v1 "system/api/system/v1"
	"system/internal/biz"
)

// SystemService 系统服务
type SystemService struct {
	v1.UnimplementedSystemServer
	uc    *biz.SystemUsecase
	logUC *biz.LogUsecase
	log   *log.Helper
}

// NewSystemService 创建系统服务
func NewSystemService(uc *biz.SystemUsecase, logUC *biz.LogUsecase, logger log.Logger) *SystemService {
	return &SystemService{
		uc:    uc,
		logUC: logUC,
		log:   log.NewHelper(logger),
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

// ==================== 系统日志 API ====================

// ListSystemLogs 查询系统日志列表
func (s *SystemService) ListSystemLogs(ctx context.Context, req *v1.ListSystemLogsRequest) (*v1.ListSystemLogsResponse, error) {
	logs, total, err := s.logUC.ListSystemLogs(ctx, req.Level, req.Module, req.OperatorId, req.StartTime, req.EndTime, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	var list []*v1.SystemLogInfo
	for _, l := range logs {
		list = append(list, &v1.SystemLogInfo{
			Id:           l.ID,
			Level:        l.Level,
			Module:       l.Module,
			Action:       l.Action,
			Message:      l.Message,
			OperatorId:   l.OperatorID,
			OperatorName: l.OperatorName,
			IpAddress:    l.IPAddress,
			UserAgent:    l.UserAgent,
			Metadata:     l.Metadata,
			CreatedAt:    timestamppb.New(l.CreatedAt),
		})
	}

	return &v1.ListSystemLogsResponse{
		Logs:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetSystemLog 获取单条系统日志
func (s *SystemService) GetSystemLog(ctx context.Context, req *v1.GetSystemLogRequest) (*v1.SystemLogInfo, error) {
	l, err := s.logUC.GetSystemLog(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &v1.SystemLogInfo{
		Id:           l.ID,
		Level:        l.Level,
		Module:       l.Module,
		Action:       l.Action,
		Message:      l.Message,
		OperatorId:   l.OperatorID,
		OperatorName: l.OperatorName,
		IpAddress:    l.IPAddress,
		UserAgent:    l.UserAgent,
		Metadata:     l.Metadata,
		CreatedAt:    timestamppb.New(l.CreatedAt),
	}, nil
}

// CreateSystemLog 创建系统日志
func (s *SystemService) CreateSystemLog(ctx context.Context, req *v1.CreateSystemLogRequest) (*v1.SystemLogInfo, error) {
	log := &biz.SystemLog{
		Level:        req.Level,
		Module:       req.Module,
		Action:       req.Action,
		Message:      req.Message,
		OperatorID:   req.OperatorId,
		OperatorName: req.OperatorName,
		IPAddress:    req.IpAddress,
		UserAgent:    req.UserAgent,
		Metadata:     req.Metadata,
	}

	if err := s.logUC.CreateSystemLog(ctx, log); err != nil {
		return nil, err
	}

	return &v1.SystemLogInfo{
		Id:           log.ID,
		Level:        log.Level,
		Module:       log.Module,
		Action:       log.Action,
		Message:      log.Message,
		OperatorId:   log.OperatorID,
		OperatorName: log.OperatorName,
		IpAddress:    log.IPAddress,
		UserAgent:    log.UserAgent,
		Metadata:     log.Metadata,
		CreatedAt:    timestamppb.New(log.CreatedAt),
	}, nil
}

// ==================== 用户日志 API ====================

// ListUserLogs 查询用户日志列表
func (s *SystemService) ListUserLogs(ctx context.Context, req *v1.ListUserLogsRequest) (*v1.ListUserLogsResponse, error) {
	logs, total, err := s.logUC.ListUserLogs(ctx, req.UserId, req.Action, req.Module, req.StartTime, req.EndTime, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	var list []*v1.UserLogInfo
	for _, l := range logs {
		list = append(list, &v1.UserLogInfo{
			Id:          l.ID,
			UserId:      l.UserID,
			Username:    l.Username,
			Action:      l.Action,
			Module:      l.Module,
			Description: l.Description,
			IpAddress:   l.IPAddress,
			DeviceInfo:  l.DeviceInfo,
			Metadata:    l.Metadata,
			CreatedAt:   timestamppb.New(l.CreatedAt),
		})
	}

	return &v1.ListUserLogsResponse{
		Logs:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetUserLog 获取单条用户日志
func (s *SystemService) GetUserLog(ctx context.Context, req *v1.GetUserLogRequest) (*v1.UserLogInfo, error) {
	l, err := s.logUC.GetUserLog(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &v1.UserLogInfo{
		Id:          l.ID,
		UserId:      l.UserID,
		Username:    l.Username,
		Action:      l.Action,
		Module:      l.Module,
		Description: l.Description,
		IpAddress:   l.IPAddress,
		DeviceInfo:  l.DeviceInfo,
		Metadata:    l.Metadata,
		CreatedAt:   timestamppb.New(l.CreatedAt),
	}, nil
}

// CreateUserLog 创建用户日志
func (s *SystemService) CreateUserLog(ctx context.Context, req *v1.CreateUserLogRequest) (*v1.UserLogInfo, error) {
	log := &biz.UserLog{
		UserID:      req.UserId,
		Username:    req.Username,
		Action:      req.Action,
		Module:      req.Module,
		Description: req.Description,
		IPAddress:   req.IpAddress,
		DeviceInfo:  req.DeviceInfo,
		Metadata:    req.Metadata,
	}

	if err := s.logUC.CreateUserLog(ctx, log); err != nil {
		return nil, err
	}

	return &v1.UserLogInfo{
		Id:          log.ID,
		UserId:      log.UserID,
		Username:    log.Username,
		Action:      log.Action,
		Module:      log.Module,
		Description: log.Description,
		IpAddress:   log.IPAddress,
		DeviceInfo:  log.DeviceInfo,
		Metadata:    log.Metadata,
		CreatedAt:   timestamppb.New(log.CreatedAt),
	}, nil
}
