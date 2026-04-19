package data

import (
	"admin/internal/biz"
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// Admin 管理员数据模型
type Admin struct {
	ID          uint32 `gorm:"primarykey"`
	Username    string `gorm:"uniqueIndex:idx_username;type:varchar(50);not null"`
	Password    string `gorm:"type:varchar(255);not null"`
	Nickname    string `gorm:"type:varchar(50)"`
	LastLoginAt *time.Time
	Status      int8 `gorm:"index:idx_status;default:1"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Menu 菜单数据模型
type Menu struct {
	ID         uint32  `gorm:"primarykey"`
	ParentID   uint32  `gorm:"index:idx_parent_id;default:0"`
	Name       string  `gorm:"type:varchar(50);not null"`
	Path       string  `gorm:"type:varchar(100)"`
	Component  string  `gorm:"type:varchar(100)"`
	Permission string  `gorm:"type:varchar(100)"`
	Icon       string  `gorm:"type:varchar(50)"`
	Type       int8    `gorm:"index:idx_type;default:1"`
	Sort       uint32  `gorm:"default:0"`
	Status     int8    `gorm:"index:idx_status;default:1"`
	Children   []*Menu `gorm:"-"` // 不存储到数据库
	CreatedAt  time.Time
}

// Role 角色数据模型
type Role struct {
	ID          uint32 `gorm:"primarykey"`
	Name        string `gorm:"uniqueIndex:idx_name;type:varchar(50);not null"`
	Code        string `gorm:"uniqueIndex:idx_code;type:varchar(50);not null"`
	Description string `gorm:"type:varchar(255)"`
	Status      int8   `gorm:"default:1"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// RoleMenu 角色菜单关联表
type RoleMenu struct {
	RoleID    uint32 `gorm:"primaryKey"`
	MenuID    uint32 `gorm:"primaryKey"`
	CreatedAt time.Time
}

// AdminRole 管理员角色关联表
type AdminRole struct {
	AdminID   uint32 `gorm:"primaryKey"`
	RoleID    uint32 `gorm:"primaryKey"`
	CreatedAt time.Time
}

// AdminLog 管理员操作日志数据模型
type AdminLog struct {
	ID              uint32 `gorm:"primarykey"`
	AdminID         uint32 `gorm:"index:idx_admin_id;not null"`
	AdminUsername   string `gorm:"type:varchar(50);not null"`
	Action          string `gorm:"type:varchar(50);not null"`
	Module          string `gorm:"index:idx_module;type:varchar(50);not null"`
	Description     string `gorm:"type:varchar(255);not null"`
	RequestMethod   string `gorm:"type:varchar(10)"`
	RequestURL      string `gorm:"type:varchar(255)"`
	RequestParams   string `gorm:"type:text"`
	ResponseCode    int32
	ResponseMessage string `gorm:"type:varchar(255)"`
	IP              string `gorm:"index:idx_ip;type:varchar(45);not null"`
	UserAgent       string `gorm:"type:varchar(500)"`
	Device          string `gorm:"type:varchar(255)"`
	TargetID        uint32 `gorm:"index:idx_target_id"`
	BeforeData      string `gorm:"type:text"`
	AfterData       string `gorm:"type:text"`
	ExecutionTime   uint32
	Status          int8   `gorm:"index:idx_status;default:1"`
	FailReason      string `gorm:"type:varchar(255)"`
	CreatedAt       time.Time
}

type adminRepo struct {
	data *Data
	log  *log.Helper
}

// NewAdminRepo 创建管理员仓库
func NewAdminRepo(data *Data, logger log.Logger) biz.AdminRepo {
	return &adminRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *adminRepo) CreateAdmin(ctx context.Context, a *biz.Admin) (*biz.Admin, error) {
	admin := Admin{
		Username: a.Username,
		Password: a.Password,
		Nickname: a.Nickname,
		Status:   a.Status,
	}
	if err := r.data.db.Create(&admin).Error; err != nil {
		return nil, grpcstatus.Errorf(codes.Internal, "%s", err.Error())
	}
	return r.toBizAdmin(&admin), nil
}

func (r *adminRepo) GetAdminByID(ctx context.Context, id uint32) (*biz.Admin, error) {
	var admin Admin
	if err := r.data.db.First(&admin, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, grpcstatus.Errorf(codes.NotFound, "管理员不存在")
		}
		return nil, grpcstatus.Errorf(codes.Internal, "%s", err.Error())
	}
	return r.toBizAdmin(&admin), nil
}

func (r *adminRepo) GetAdminByUsername(ctx context.Context, username string) (*biz.Admin, error) {
	var admin Admin
	if err := r.data.db.Where("username = ?", username).First(&admin).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, grpcstatus.Errorf(codes.NotFound, "管理员不存在")
		}
		return nil, grpcstatus.Errorf(codes.Internal, "%s", err.Error())
	}
	return r.toBizAdmin(&admin), nil
}

func (r *adminRepo) UpdateAdmin(ctx context.Context, a *biz.Admin) (*biz.Admin, error) {
	updates := map[string]interface{}{
		"nickname": a.Nickname,
		"status":   a.Status,
	}
	if a.Password != "" {
		updates["password"] = a.Password
	}

	if err := r.data.db.Model(&Admin{}).Where("id = ?", a.ID).Updates(updates).Error; err != nil {
		return nil, grpcstatus.Errorf(codes.Internal, "%s", err.Error())
	}
	return r.GetAdminByID(ctx, a.ID)
}

func (r *adminRepo) DeleteAdmin(ctx context.Context, id uint32) error {
	if err := r.data.db.Delete(&Admin{}, id).Error; err != nil {
		return grpcstatus.Errorf(codes.Internal, "%s", err.Error())
	}
	return nil
}

func (r *adminRepo) ListAdmins(ctx context.Context, page, pageSize int) ([]*biz.Admin, int64, error) {
	var admins []*Admin
	var total int64

	offset := (page - 1) * pageSize
	if err := r.data.db.Model(&Admin{}).Count(&total).Error; err != nil {
		return nil, 0, grpcstatus.Errorf(codes.Internal, "%s", err.Error())
	}

	if err := r.data.db.Limit(pageSize).Offset(offset).Find(&admins).Error; err != nil {
		return nil, 0, grpcstatus.Errorf(codes.Internal, "%s", err.Error())
	}

	bizAdmins := make([]*biz.Admin, 0, len(admins))
	for _, a := range admins {
		bizAdmins = append(bizAdmins, r.toBizAdmin(a))
	}
	return bizAdmins, total, nil
}

func (r *adminRepo) UpdateLastLogin(ctx context.Context, id uint32, loginTime time.Time) error {
	return r.data.db.Model(&Admin{}).Where("id = ?", id).Update("last_login_at", loginTime).Error
}

func (r *adminRepo) toBizAdmin(a *Admin) *biz.Admin {
	return &biz.Admin{
		ID:          a.ID,
		Username:    a.Username,
		Password:    a.Password,
		Nickname:    a.Nickname,
		LastLoginAt: a.LastLoginAt,
		Status:      a.Status,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}

type menuRepo struct {
	data *Data
	log  *log.Helper
}

// NewMenuRepo 创建菜单仓库
func NewMenuRepo(data *Data, logger log.Logger) biz.MenuRepo {
	return &menuRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *menuRepo) CreateMenu(ctx context.Context, m *biz.Menu) (*biz.Menu, error) {
	menu := Menu{
		ParentID:   m.ParentID,
		Name:       m.Name,
		Path:       m.Path,
		Component:  m.Component,
		Permission: m.Permission,
		Icon:       m.Icon,
		Type:       m.Type,
		Sort:       m.Sort,
		Status:     m.Status,
	}
	if err := r.data.db.Create(&menu).Error; err != nil {
		return nil, grpcstatus.Errorf(codes.Internal, "%s", err.Error())
	}
	return r.toBizMenu(&menu), nil
}

func (r *menuRepo) GetMenuByID(ctx context.Context, id uint32) (*biz.Menu, error) {
	var menu Menu
	if err := r.data.db.First(&menu, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, grpcstatus.Errorf(codes.NotFound, "菜单不存在")
		}
		return nil, grpcstatus.Errorf(codes.Internal, "%s", err.Error())
	}
	return r.toBizMenu(&menu), nil
}

func (r *menuRepo) UpdateMenu(ctx context.Context, m *biz.Menu) (*biz.Menu, error) {
	updates := map[string]interface{}{
		"parent_id":  m.ParentID,
		"name":       m.Name,
		"path":       m.Path,
		"component":  m.Component,
		"permission": m.Permission,
		"icon":       m.Icon,
		"type":       m.Type,
		"sort":       m.Sort,
		"status":     m.Status,
	}

	if err := r.data.db.Model(&Menu{}).Where("id = ?", m.ID).Updates(updates).Error; err != nil {
		return nil, grpcstatus.Errorf(codes.Internal, "%s", err.Error())
	}
	return r.GetMenuByID(ctx, m.ID)
}

func (r *menuRepo) DeleteMenu(ctx context.Context, id uint32) error {
	if err := r.data.db.Delete(&Menu{}, id).Error; err != nil {
		return grpcstatus.Errorf(codes.Internal, "%s", err.Error())
	}
	return nil
}

func (r *menuRepo) ListMenus(ctx context.Context, statusFilter int8) ([]*biz.Menu, error) {
	var menus []*Menu
	query := r.data.db
	if statusFilter >= 0 {
		query = query.Where("status = ?", statusFilter)
	}
	if err := query.Order("sort asc, id asc").Find(&menus).Error; err != nil {
		return nil, grpcstatus.Errorf(codes.Internal, "%s", err.Error())
	}

	bizMenus := make([]*biz.Menu, 0, len(menus))
	for _, m := range menus {
		bizMenus = append(bizMenus, r.toBizMenu(m))
	}
	return bizMenus, nil
}

func (r *menuRepo) GetMenusByRoleIDs(ctx context.Context, roleIDs []uint32) ([]*biz.Menu, error) {
	var menuIDs []uint32
	if err := r.data.db.Model(&RoleMenu{}).Where("role_id IN ?", roleIDs).Pluck("menu_id", &menuIDs).Error; err != nil {
		return nil, grpcstatus.Errorf(codes.Internal, "%s", err.Error())
	}

	if len(menuIDs) == 0 {
		return []*biz.Menu{}, nil
	}

	var menus []*Menu
	if err := r.data.db.Where("id IN ? AND status = ?", menuIDs, 1).Order("sort asc, id asc").Find(&menus).Error; err != nil {
		return nil, grpcstatus.Errorf(codes.Internal, "%s", err.Error())
	}

	bizMenus := make([]*biz.Menu, 0, len(menus))
	for _, m := range menus {
		bizMenus = append(bizMenus, r.toBizMenu(m))
	}
	return bizMenus, nil
}

func (r *menuRepo) toBizMenu(m *Menu) *biz.Menu {
	return &biz.Menu{
		ID:         m.ID,
		ParentID:   m.ParentID,
		Name:       m.Name,
		Path:       m.Path,
		Component:  m.Component,
		Permission: m.Permission,
		Icon:       m.Icon,
		Type:       m.Type,
		Sort:       m.Sort,
		Status:     m.Status,
		CreatedAt:  m.CreatedAt,
	}
}

type roleRepo struct {
	data *Data
	log  *log.Helper
}

// NewRoleRepo 创建角色仓库
func NewRoleRepo(data *Data, logger log.Logger) biz.RoleRepo {
	return &roleRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *roleRepo) CreateRole(ctx context.Context, role *biz.Role) (*biz.Role, error) {
	r2 := Role{
		Name:        role.Name,
		Code:        role.Code,
		Description: role.Description,
		Status:      role.Status,
	}
	if err := r.data.db.Create(&r2).Error; err != nil {
		return nil, grpcstatus.Errorf(codes.Internal, "%s", err.Error())
	}
	return r.toBizRole(&r2), nil
}

func (r *roleRepo) GetRoleByID(ctx context.Context, id uint32) (*biz.Role, error) {
	var role Role
	if err := r.data.db.First(&role, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, grpcstatus.Errorf(codes.NotFound, "角色不存在")
		}
		return nil, grpcstatus.Errorf(codes.Internal, "%s", err.Error())
	}
	return r.toBizRole(&role), nil
}

func (r *roleRepo) GetRoleByCode(ctx context.Context, code string) (*biz.Role, error) {
	var role Role
	if err := r.data.db.Where("code = ?", code).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, grpcstatus.Errorf(codes.NotFound, "角色不存在")
		}
		return nil, grpcstatus.Errorf(codes.Internal, "%s", err.Error())
	}
	return r.toBizRole(&role), nil
}

func (r *roleRepo) UpdateRole(ctx context.Context, role *biz.Role) (*biz.Role, error) {
	updates := map[string]interface{}{
		"name":        role.Name,
		"description": role.Description,
		"status":      role.Status,
	}

	if err := r.data.db.Model(&Role{}).Where("id = ?", role.ID).Updates(updates).Error; err != nil {
		return nil, grpcstatus.Errorf(codes.Internal, "%s", err.Error())
	}
	return r.GetRoleByID(ctx, role.ID)
}

func (r *roleRepo) DeleteRole(ctx context.Context, id uint32) error {
	if err := r.data.db.Delete(&Role{}, id).Error; err != nil {
		return grpcstatus.Errorf(codes.Internal, "%s", err.Error())
	}
	return nil
}

func (r *roleRepo) ListRoles(ctx context.Context, statusFilter int8) ([]*biz.Role, error) {
	var roles []*Role
	query := r.data.db
	if statusFilter >= 0 {
		query = query.Where("status = ?", statusFilter)
	}
	if err := query.Find(&roles).Error; err != nil {
		return nil, grpcstatus.Errorf(codes.Internal, "%s", err.Error())
	}

	bizRoles := make([]*biz.Role, 0, len(roles))
	for _, role := range roles {
		bizRoles = append(bizRoles, r.toBizRole(role))
	}
	return bizRoles, nil
}

func (r *roleRepo) AssignRoleMenus(ctx context.Context, roleID uint32, menuIDs []uint32) error {
	// 删除旧关联
	if err := r.data.db.Where("role_id = ?", roleID).Delete(&RoleMenu{}).Error; err != nil {
		return err
	}

	// 添加新关联
	if len(menuIDs) > 0 {
		roleMenus := make([]RoleMenu, 0, len(menuIDs))
		for _, menuID := range menuIDs {
			roleMenus = append(roleMenus, RoleMenu{
				RoleID:    roleID,
				MenuID:    menuID,
				CreatedAt: time.Now(),
			})
		}
		if err := r.data.db.Create(&roleMenus).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *roleRepo) GetRoleMenuIDs(ctx context.Context, roleID uint32) ([]uint32, error) {
	var menuIDs []uint32
	if err := r.data.db.Model(&RoleMenu{}).Where("role_id = ?", roleID).Pluck("menu_id", &menuIDs).Error; err != nil {
		return nil, err
	}
	return menuIDs, nil
}

func (r *roleRepo) toBizRole(role *Role) *biz.Role {
	return &biz.Role{
		ID:          role.ID,
		Name:        role.Name,
		Code:        role.Code,
		Description: role.Description,
		Status:      role.Status,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}
}

type adminRoleRepo struct {
	data *Data
	log  *log.Helper
}

// NewAdminRoleRepo 创建管理员角色关联仓库
func NewAdminRoleRepo(data *Data, logger log.Logger) biz.AdminRoleRepo {
	return &adminRoleRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *adminRoleRepo) AssignAdminRoles(ctx context.Context, adminID uint32, roleIDs []uint32) error {
	// 删除旧关联
	if err := r.data.db.Where("admin_id = ?", adminID).Delete(&AdminRole{}).Error; err != nil {
		return err
	}

	// 添加新关联
	if len(roleIDs) > 0 {
		adminRoles := make([]AdminRole, 0, len(roleIDs))
		for _, roleID := range roleIDs {
			adminRoles = append(adminRoles, AdminRole{
				AdminID:   adminID,
				RoleID:    roleID,
				CreatedAt: time.Now(),
			})
		}
		if err := r.data.db.Create(&adminRoles).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *adminRoleRepo) GetAdminRoleIDs(ctx context.Context, adminID uint32) ([]uint32, error) {
	var roleIDs []uint32
	if err := r.data.db.Model(&AdminRole{}).Where("admin_id = ?", adminID).Pluck("role_id", &roleIDs).Error; err != nil {
		return nil, err
	}
	return roleIDs, nil
}

func (r *adminRoleRepo) GetAdminRoleCodes(ctx context.Context, adminID uint32) ([]string, error) {
	var codes []string
	if err := r.data.db.Model(&AdminRole{}).
		Select("roles.code").
		Joins("JOIN roles ON admin_roles.role_id = roles.id").
		Where("admin_roles.admin_id = ?", adminID).
		Pluck("roles.code", &codes).Error; err != nil {
		return nil, err
	}
	return codes, nil
}

type adminLogRepo struct {
	data *Data
	log  *log.Helper
}

// NewAdminLogRepo 创建管理员日志仓库
func NewAdminLogRepo(data *Data, logger log.Logger) biz.AdminLogRepo {
	return &adminLogRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *adminLogRepo) CreateLog(ctx context.Context, log *biz.AdminLog) error {
	l := AdminLog{
		AdminID:         log.AdminID,
		AdminUsername:   log.AdminUsername,
		Action:          log.Action,
		Module:          log.Module,
		Description:     log.Description,
		RequestMethod:   log.RequestMethod,
		RequestURL:      log.RequestURL,
		RequestParams:   log.RequestParams,
		ResponseCode:    log.ResponseCode,
		ResponseMessage: log.ResponseMessage,
		IP:              log.IP,
		UserAgent:       log.UserAgent,
		Device:          log.Device,
		TargetID:        log.TargetID,
		BeforeData:      log.BeforeData,
		AfterData:       log.AfterData,
		ExecutionTime:   log.ExecutionTime,
		Status:          log.Status,
		FailReason:      log.FailReason,
	}
	if err := r.data.db.Create(&l).Error; err != nil {
		return grpcstatus.Errorf(codes.Internal, "%s", err.Error())
	}
	return nil
}

func (r *adminLogRepo) ListLogs(ctx context.Context, adminID uint32, module string, page, pageSize int) ([]*biz.AdminLog, int64, error) {
	var logs []*AdminLog
	var total int64

	query := r.data.db.Model(&AdminLog{})
	if adminID > 0 {
		query = query.Where("admin_id = ?", adminID)
	}
	if module != "" {
		query = query.Where("module = ?", module)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, grpcstatus.Errorf(codes.Internal, "%s", err.Error())
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at desc").Limit(pageSize).Offset(offset).Find(&logs).Error; err != nil {
		return nil, 0, grpcstatus.Errorf(codes.Internal, "%s", err.Error())
	}

	bizLogs := make([]*biz.AdminLog, 0, len(logs))
	for _, l := range logs {
		bizLogs = append(bizLogs, r.toBizAdminLog(l))
	}
	return bizLogs, total, nil
}

func (r *adminLogRepo) toBizAdminLog(l *AdminLog) *biz.AdminLog {
	return &biz.AdminLog{
		ID:              l.ID,
		AdminID:         l.AdminID,
		AdminUsername:   l.AdminUsername,
		Action:          l.Action,
		Module:          l.Module,
		Description:     l.Description,
		RequestMethod:   l.RequestMethod,
		RequestURL:      l.RequestURL,
		RequestParams:   l.RequestParams,
		ResponseCode:    l.ResponseCode,
		ResponseMessage: l.ResponseMessage,
		IP:              l.IP,
		UserAgent:       l.UserAgent,
		Device:          l.Device,
		TargetID:        l.TargetID,
		BeforeData:      l.BeforeData,
		AfterData:       l.AfterData,
		ExecutionTime:   l.ExecutionTime,
		Status:          l.Status,
		FailReason:      l.FailReason,
		CreatedAt:       l.CreatedAt,
	}
}
