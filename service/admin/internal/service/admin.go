package service

import (
	"context"

	v1 "admin/api/admin/v1"
	"admin/internal/biz"
	"admin/internal/pkg/captcha"
	"admin/internal/pkg/jwt"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AdminService 管理服务
type AdminService struct {
	v1.UnimplementedAdminServer
	adminUC        *biz.AdminUsecase
	menuUC         *biz.MenuUsecase
	roleUC         *biz.RoleUsecase
	adminLogUC     *biz.AdminLogUsecase
	jwtManager     *jwt.JWTManager
	captchaManager *captcha.CaptchaManager
	log            *log.Helper
}

// NewAdminService 创建管理服务
func NewAdminService(adminUC *biz.AdminUsecase, menuUC *biz.MenuUsecase, roleUC *biz.RoleUsecase, adminLogUC *biz.AdminLogUsecase, jwtManager *jwt.JWTManager, captchaManager *captcha.CaptchaManager, logger log.Logger) *AdminService {
	return &AdminService{
		adminUC:        adminUC,
		menuUC:         menuUC,
		roleUC:         roleUC,
		adminLogUC:     adminLogUC,
		jwtManager:     jwtManager,
		captchaManager: captchaManager,
		log:            log.NewHelper(logger),
	}
}

// GetCaptcha 获取验证码
func (s *AdminService) GetCaptcha(ctx context.Context, req *v1.GetCaptchaRequest) (*v1.GetCaptchaResponse, error) {
	id, b64s, err := s.captchaManager.Generate()
	if err != nil {
		return nil, errors.New(500, "CAPTCHA_GENERATE_FAILED", "验证码生成失败")
	}
	return &v1.GetCaptchaResponse{
		CaptchaId:    id,
		CaptchaImage: b64s,
	}, nil
}

// Login 管理员登录
func (s *AdminService) Login(ctx context.Context, req *v1.AdminLoginRequest) (*v1.AdminLoginResponse, error) {
	if !s.captchaManager.Verify(req.CaptchaId, req.CaptchaCode, true) {
		return nil, errors.New(400, "CAPTCHA_VERIFY_FAILED", "验证码错误")
	}

	admin, err := s.adminUC.AdminLogin(ctx, req.Username, req.Password)
	if err != nil {
		return nil, errors.New(400, "LOGIN_FAILED", "用户名或密码错误")
	}

	token, err := s.jwtManager.GenerateToken(admin.ID, admin.Username)
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.jwtManager.GenerateRefreshToken(admin.ID, admin.Username)
	if err != nil {
		return nil, err
	}

	return &v1.AdminLoginResponse{
		Admin:        s.toProtoAdmin(admin),
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

// CreateAdmin 创建管理员
func (s *AdminService) CreateAdmin(ctx context.Context, req *v1.CreateAdminRequest) (*v1.AdminInfo, error) {
	admin, err := s.adminUC.CreateAdmin(ctx, &biz.Admin{
		Username: req.Username,
		Password: req.Password,
		Nickname: req.Nickname,
		Roles:    req.Roles,
		Status:   1,
	})
	if err != nil {
		return nil, err
	}

	s.log.Infof("管理员 %s 创建成功", req.Username)
	return s.toProtoAdmin(admin), nil
}

// GetAdmin 获取管理员
func (s *AdminService) GetAdmin(ctx context.Context, req *v1.GetAdminRequest) (*v1.AdminInfo, error) {
	admin, err := s.adminUC.GetAdmin(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return s.toProtoAdmin(admin), nil
}

// UpdateAdmin 更新管理员
func (s *AdminService) UpdateAdmin(ctx context.Context, req *v1.UpdateAdminRequest) (*v1.AdminInfo, error) {
	admin, err := s.adminUC.UpdateAdmin(ctx, &biz.Admin{
		ID:       req.Id,
		Nickname: req.Nickname,
		Password: req.Password,
		Status:   int8(req.Status),
		Roles:    req.Roles,
	})
	if err != nil {
		return nil, err
	}
	return s.toProtoAdmin(admin), nil
}

// DeleteAdmin 删除管理员
func (s *AdminService) DeleteAdmin(ctx context.Context, req *v1.DeleteAdminRequest) (*v1.DeleteAdminResponse, error) {
	err := s.adminUC.DeleteAdmin(ctx, req.Id)
	if err != nil {
		return &v1.DeleteAdminResponse{Success: false, Message: err.Error()}, err
	}
	return &v1.DeleteAdminResponse{Success: true, Message: "删除成功"}, nil
}

// ListAdmins 列出管理员
func (s *AdminService) ListAdmins(ctx context.Context, req *v1.ListAdminsRequest) (*v1.ListAdminsResponse, error) {
	admins, total, err := s.adminUC.ListAdmins(ctx, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, err
	}

	adminInfos := make([]*v1.AdminInfo, 0, len(admins))
	for _, a := range admins {
		adminInfos = append(adminInfos, s.toProtoAdmin(a))
	}

	return &v1.ListAdminsResponse{
		Total:  total,
		Admins: adminInfos,
	}, nil
}

func (s *AdminService) toProtoAdmin(a *biz.Admin) *v1.AdminInfo {
	return &v1.AdminInfo{
		Id:       a.ID,
		Username: a.Username,
		Nickname: a.Nickname,
		LastLoginAt: func() *timestamppb.Timestamp {
			if a.LastLoginAt != nil {
				return timestamppb.New(*a.LastLoginAt)
			}
			return nil
		}(),
		Status:    int32(a.Status),
		Roles:     a.Roles,
		CreatedAt: timestamppb.New(a.CreatedAt),
	}
}

// CreateMenu 创建菜单
func (s *AdminService) CreateMenu(ctx context.Context, req *v1.CreateMenuRequest) (*v1.MenuInfo, error) {
	menu, err := s.menuUC.CreateMenu(ctx, &biz.Menu{
		ParentID:   req.ParentId,
		Name:       req.Name,
		Path:       req.Path,
		Component:  req.Component,
		Permission: req.Permission,
		Icon:       req.Icon,
		Type:       int8(req.Type),
		Sort:       req.Sort,
		Status:     1,
	})
	if err != nil {
		return nil, err
	}
	return s.toProtoMenu(menu), nil
}

// GetMenu 获取菜单
func (s *AdminService) GetMenu(ctx context.Context, req *v1.GetMenuRequest) (*v1.MenuInfo, error) {
	menu, err := s.menuUC.GetMenu(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return s.toProtoMenu(menu), nil
}

// UpdateMenu 更新菜单
func (s *AdminService) UpdateMenu(ctx context.Context, req *v1.UpdateMenuRequest) (*v1.MenuInfo, error) {
	menu, err := s.menuUC.UpdateMenu(ctx, &biz.Menu{
		ID:         req.Id,
		ParentID:   req.ParentId,
		Name:       req.Name,
		Path:       req.Path,
		Component:  req.Component,
		Permission: req.Permission,
		Icon:       req.Icon,
		Type:       int8(req.Type),
		Sort:       req.Sort,
		Status:     int8(req.Status),
	})
	if err != nil {
		return nil, err
	}
	return s.toProtoMenu(menu), nil
}

// DeleteMenu 删除菜单
func (s *AdminService) DeleteMenu(ctx context.Context, req *v1.DeleteMenuRequest) (*v1.DeleteMenuResponse, error) {
	err := s.menuUC.DeleteMenu(ctx, req.Id)
	if err != nil {
		return &v1.DeleteMenuResponse{Success: false, Message: err.Error()}, err
	}
	return &v1.DeleteMenuResponse{Success: true, Message: "删除成功"}, nil
}

// ListMenus 列出菜单
func (s *AdminService) ListMenus(ctx context.Context, req *v1.ListMenusRequest) (*v1.ListMenusResponse, error) {
	menus, err := s.menuUC.ListMenus(ctx, int8(req.Status))
	if err != nil {
		return nil, err
	}

	return &v1.ListMenusResponse{Menus: s.buildProtoMenuTree(menus)}, nil
}

func (s *AdminService) toProtoMenu(m *biz.Menu) *v1.MenuInfo {
	return &v1.MenuInfo{
		Id:         m.ID,
		ParentId:   m.ParentID,
		Name:       m.Name,
		Path:       m.Path,
		Component:  m.Component,
		Permission: m.Permission,
		Icon:       m.Icon,
		Type:       int32(m.Type),
		Sort:       m.Sort,
		Status:     int32(m.Status),
	}
}

func (s *AdminService) buildProtoMenuTree(menus []*biz.Menu) []*v1.MenuInfo {
	menuMap := make(map[uint32]*v1.MenuInfo)
	var roots []*v1.MenuInfo

	for _, m := range menus {
		info := s.toProtoMenu(m)
		menuMap[info.Id] = info
	}

	for _, m := range menus {
		info := menuMap[m.ID]
		if m.ParentID == 0 {
			roots = append(roots, info)
		} else {
			if parent, ok := menuMap[m.ParentID]; ok {
				parent.Children = append(parent.Children, info)
			}
		}
	}

	return roots
}

// CreateRole 创建角色
func (s *AdminService) CreateRole(ctx context.Context, req *v1.CreateRoleRequest) (*v1.RoleInfo, error) {
	role, err := s.roleUC.CreateRole(ctx, &biz.Role{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		MenuIDs:     req.MenuIds,
		Status:      1,
	})
	if err != nil {
		return nil, err
	}
	return s.toProtoRole(role), nil
}

// GetRole 获取角色
func (s *AdminService) GetRole(ctx context.Context, req *v1.GetRoleRequest) (*v1.RoleInfo, error) {
	role, err := s.roleUC.GetRole(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return s.toProtoRole(role), nil
}

// UpdateRole 更新角色
func (s *AdminService) UpdateRole(ctx context.Context, req *v1.UpdateRoleRequest) (*v1.RoleInfo, error) {
	role, err := s.roleUC.UpdateRole(ctx, &biz.Role{
		ID:          req.Id,
		Name:        req.Name,
		Description: req.Description,
		Status:      int8(req.Status),
		MenuIDs:     req.MenuIds,
	})
	if err != nil {
		return nil, err
	}
	return s.toProtoRole(role), nil
}

// DeleteRole 删除角色
func (s *AdminService) DeleteRole(ctx context.Context, req *v1.DeleteRoleRequest) (*v1.DeleteRoleResponse, error) {
	err := s.roleUC.DeleteRole(ctx, req.Id)
	if err != nil {
		return &v1.DeleteRoleResponse{Success: false, Message: err.Error()}, err
	}
	return &v1.DeleteRoleResponse{Success: true, Message: "删除成功"}, nil
}

// ListRoles 列出角色
func (s *AdminService) ListRoles(ctx context.Context, req *v1.ListRolesRequest) (*v1.ListRolesResponse, error) {
	roles, err := s.roleUC.ListRoles(ctx, int8(req.Status))
	if err != nil {
		return nil, err
	}

	roleInfos := make([]*v1.RoleInfo, 0, len(roles))
	for _, r := range roles {
		roleInfos = append(roleInfos, s.toProtoRole(r))
	}
	return &v1.ListRolesResponse{Roles: roleInfos}, nil
}

func (s *AdminService) toProtoRole(r *biz.Role) *v1.RoleInfo {
	return &v1.RoleInfo{
		Id:          r.ID,
		Name:        r.Name,
		Code:        r.Code,
		Description: r.Description,
		Status:      int32(r.Status),
		MenuIds:     r.MenuIDs,
		CreatedAt:   timestamppb.New(r.CreatedAt),
	}
}

// ListAdminLogs 列出管理员操作日志
func (s *AdminService) ListAdminLogs(ctx context.Context, req *v1.ListAdminLogsRequest) (*v1.ListAdminLogsResponse, error) {
	logs, total, err := s.adminLogUC.ListLogs(ctx, req.AdminId, req.Module, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, err
	}

	logInfos := make([]*v1.LogInfo, 0, len(logs))
	for _, l := range logs {
		logInfos = append(logInfos, &v1.LogInfo{
			Id:            l.ID,
			AdminId:       l.AdminID,
			AdminUsername: l.AdminUsername,
			Action:        l.Action,
			Module:        l.Module,
			Description:   l.Description,
			RequestMethod: l.RequestMethod,
			RequestUrl:    l.RequestURL,
			ResponseCode:  l.ResponseCode,
			Ip:            l.IP,
			Status:        int32(l.Status),
			FailReason:    l.FailReason,
			CreatedAt:     timestamppb.New(l.CreatedAt),
		})
	}

	return &v1.ListAdminLogsResponse{
		Total: total,
		Logs:  logInfos,
	}, nil
}
