package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/crypto/bcrypt"
)

// Admin 管理员领域模型
type Admin struct {
	ID          uint32
	Username    string
	Password    string
	Nickname    string
	LastLoginAt *time.Time
	Status      int8
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Roles       []string // 角色代码列表
}

// Menu 菜单领域模型
type Menu struct {
	ID         uint32
	ParentID   uint32
	Name       string
	Path       string
	Component  string
	Permission string
	Icon       string
	Type       int8 // 1菜单 2按钮
	Sort       uint32
	Status     int8 // 1显示 0隐藏
	Children   []*Menu
	CreatedAt  time.Time
}

// Role 角色领域模型
type Role struct {
	ID          uint32
	Name        string
	Code        string
	Description string
	Status      int8
	MenuIDs     []uint32 // 关联菜单ID列表
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// AdminLog 管理员操作日志领域模型
type AdminLog struct {
	ID              uint32
	AdminID         uint32
	AdminUsername   string
	Action          string
	Module          string
	Description     string
	RequestMethod   string
	RequestURL      string
	RequestParams   string
	ResponseCode    int32
	ResponseMessage string
	IP              string
	UserAgent       string
	Device          string
	TargetID        uint32
	BeforeData      string
	AfterData       string
	ExecutionTime   uint32
	Status          int8
	FailReason      string
	CreatedAt       time.Time
}

// AdminRepo 管理员存储接口
type AdminRepo interface {
	CreateAdmin(ctx context.Context, a *Admin) (*Admin, error)
	GetAdminByID(ctx context.Context, id uint32) (*Admin, error)
	GetAdminByUsername(ctx context.Context, username string) (*Admin, error)
	UpdateAdmin(ctx context.Context, a *Admin) (*Admin, error)
	DeleteAdmin(ctx context.Context, id uint32) error
	ListAdmins(ctx context.Context, page, pageSize int) ([]*Admin, int64, error)
	UpdateLastLogin(ctx context.Context, id uint32, loginTime time.Time) error
}

// MenuRepo 菜单存储接口
type MenuRepo interface {
	CreateMenu(ctx context.Context, m *Menu) (*Menu, error)
	GetMenuByID(ctx context.Context, id uint32) (*Menu, error)
	UpdateMenu(ctx context.Context, m *Menu) (*Menu, error)
	DeleteMenu(ctx context.Context, id uint32) error
	ListMenus(ctx context.Context, status int8) ([]*Menu, error)
	GetMenusByRoleIDs(ctx context.Context, roleIDs []uint32) ([]*Menu, error)
}

// RoleRepo 角色存储接口
type RoleRepo interface {
	CreateRole(ctx context.Context, r *Role) (*Role, error)
	GetRoleByID(ctx context.Context, id uint32) (*Role, error)
	GetRoleByCode(ctx context.Context, code string) (*Role, error)
	UpdateRole(ctx context.Context, r *Role) (*Role, error)
	DeleteRole(ctx context.Context, id uint32) error
	ListRoles(ctx context.Context, status int8) ([]*Role, error)
	AssignRoleMenus(ctx context.Context, roleID uint32, menuIDs []uint32) error
	GetRoleMenuIDs(ctx context.Context, roleID uint32) ([]uint32, error)
}

// AdminRoleRepo 管理员角色关联存储接口
type AdminRoleRepo interface {
	AssignAdminRoles(ctx context.Context, adminID uint32, roleIDs []uint32) error
	GetAdminRoleIDs(ctx context.Context, adminID uint32) ([]uint32, error)
	GetAdminRoleCodes(ctx context.Context, adminID uint32) ([]string, error)
}

// AdminLogRepo 管理员日志存储接口
type AdminLogRepo interface {
	CreateLog(ctx context.Context, log *AdminLog) error
	ListLogs(ctx context.Context, adminID uint32, module string, page, pageSize int) ([]*AdminLog, int64, error)
}

// AdminUsecase 管理员用例
type AdminUsecase struct {
	adminRepo     AdminRepo
	roleRepo      RoleRepo
	adminRoleRepo AdminRoleRepo
	log           *log.Helper
}

// NewAdminUsecase 创建管理员用例
func NewAdminUsecase(adminRepo AdminRepo, roleRepo RoleRepo, adminRoleRepo AdminRoleRepo, logger log.Logger) *AdminUsecase {
	return &AdminUsecase{
		adminRepo:     adminRepo,
		roleRepo:      roleRepo,
		adminRoleRepo: adminRoleRepo,
		log:           log.NewHelper(logger),
	}
}

// HashPassword 使用 bcrypt 哈希密码
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// VerifyPassword 验证 bcrypt 密码
func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// CreateAdmin 创建管理员
func (uc *AdminUsecase) CreateAdmin(ctx context.Context, a *Admin) (*Admin, error) {
	hashedPwd, err := HashPassword(a.Password)
	if err != nil {
		return nil, err
	}
	a.Password = hashedPwd

	admin, err := uc.adminRepo.CreateAdmin(ctx, a)
	if err != nil {
		return nil, err
	}

	// 分配角色
	if len(a.Roles) > 0 {
		roleIDs := make([]uint32, 0, len(a.Roles))
		for _, code := range a.Roles {
			role, err := uc.roleRepo.GetRoleByCode(ctx, code)
			if err != nil {
				continue
			}
			roleIDs = append(roleIDs, role.ID)
		}
		if err := uc.adminRoleRepo.AssignAdminRoles(ctx, admin.ID, roleIDs); err != nil {
			uc.log.Warnf("assign admin roles failed: %v", err)
		}
	}

	return admin, nil
}

// GetAdmin 获取管理员信息
func (uc *AdminUsecase) GetAdmin(ctx context.Context, id uint32) (*Admin, error) {
	admin, err := uc.adminRepo.GetAdminByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 获取角色代码
	roleCodes, err := uc.adminRoleRepo.GetAdminRoleCodes(ctx, id)
	if err == nil {
		admin.Roles = roleCodes
	}

	return admin, nil
}

// GetAdminByUsername 根据用户名获取管理员
func (uc *AdminUsecase) GetAdminByUsername(ctx context.Context, username string) (*Admin, error) {
	return uc.adminRepo.GetAdminByUsername(ctx, username)
}

// UpdateAdmin 更新管理员
func (uc *AdminUsecase) UpdateAdmin(ctx context.Context, a *Admin) (*Admin, error) {
	if a.Password != "" {
		hashedPwd, err := HashPassword(a.Password)
		if err != nil {
			return nil, err
		}
		a.Password = hashedPwd
	}

	admin, err := uc.adminRepo.UpdateAdmin(ctx, a)
	if err != nil {
		return nil, err
	}

	// 更新角色
	if len(a.Roles) > 0 {
		roleIDs := make([]uint32, 0, len(a.Roles))
		for _, code := range a.Roles {
			role, err := uc.roleRepo.GetRoleByCode(ctx, code)
			if err != nil {
				continue
			}
			roleIDs = append(roleIDs, role.ID)
		}
		if err := uc.adminRoleRepo.AssignAdminRoles(ctx, a.ID, roleIDs); err != nil {
			uc.log.Warnf("assign admin roles failed: %v", err)
		}
	}

	return admin, nil
}

// DeleteAdmin 删除管理员
func (uc *AdminUsecase) DeleteAdmin(ctx context.Context, id uint32) error {
	return uc.adminRepo.DeleteAdmin(ctx, id)
}

// ListAdmins 列出管理员
func (uc *AdminUsecase) ListAdmins(ctx context.Context, page, pageSize int) ([]*Admin, int64, error) {
	return uc.adminRepo.ListAdmins(ctx, page, pageSize)
}

// AdminLogin 管理员登录
func (uc *AdminUsecase) AdminLogin(ctx context.Context, username, password string) (*Admin, error) {
	admin, err := uc.adminRepo.GetAdminByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	if err := VerifyPassword(admin.Password, password); err != nil {
		return nil, err
	}

	if admin.Status != 1 {
		return nil, err
	}

	// 更新最后登录时间
	now := time.Now()
	if err := uc.adminRepo.UpdateLastLogin(ctx, admin.ID, now); err != nil {
		uc.log.Warnf("update last login time failed: %v", err)
	}

	// 获取角色
	roleCodes, err := uc.adminRoleRepo.GetAdminRoleCodes(ctx, admin.ID)
	if err == nil {
		admin.Roles = roleCodes
	}

	return admin, nil
}

// MenuUsecase 菜单用例
type MenuUsecase struct {
	repo MenuRepo
	log  *log.Helper
}

// NewMenuUsecase 创建菜单用例
func NewMenuUsecase(repo MenuRepo, logger log.Logger) *MenuUsecase {
	return &MenuUsecase{repo: repo, log: log.NewHelper(logger)}
}

// CreateMenu 创建菜单
func (uc *MenuUsecase) CreateMenu(ctx context.Context, m *Menu) (*Menu, error) {
	return uc.repo.CreateMenu(ctx, m)
}

// GetMenu 获取菜单
func (uc *MenuUsecase) GetMenu(ctx context.Context, id uint32) (*Menu, error) {
	return uc.repo.GetMenuByID(ctx, id)
}

// UpdateMenu 更新菜单
func (uc *MenuUsecase) UpdateMenu(ctx context.Context, m *Menu) (*Menu, error) {
	return uc.repo.UpdateMenu(ctx, m)
}

// DeleteMenu 删除菜单
func (uc *MenuUsecase) DeleteMenu(ctx context.Context, id uint32) error {
	return uc.repo.DeleteMenu(ctx, id)
}

// ListMenus 列出菜单
func (uc *MenuUsecase) ListMenus(ctx context.Context, status int8) ([]*Menu, error) {
	return uc.repo.ListMenus(ctx, status)
}

// GetMenusByRoleIDs 根据角色ID获取菜单
func (uc *MenuUsecase) GetMenusByRoleIDs(ctx context.Context, roleIDs []uint32) ([]*Menu, error) {
	return uc.repo.GetMenusByRoleIDs(ctx, roleIDs)
}

// RoleUsecase 角色用例
type RoleUsecase struct {
	repo          RoleRepo
	menuRepo      MenuRepo
	adminRoleRepo AdminRoleRepo
	log           *log.Helper
}

// NewRoleUsecase 创建角色用例
func NewRoleUsecase(repo RoleRepo, menuRepo MenuRepo, adminRoleRepo AdminRoleRepo, logger log.Logger) *RoleUsecase {
	return &RoleUsecase{
		repo:          repo,
		menuRepo:      menuRepo,
		adminRoleRepo: adminRoleRepo,
		log:           log.NewHelper(logger),
	}
}

// CreateRole 创建角色
func (uc *RoleUsecase) CreateRole(ctx context.Context, r *Role) (*Role, error) {
	role, err := uc.repo.CreateRole(ctx, r)
	if err != nil {
		return nil, err
	}

	// 分配菜单权限
	if len(r.MenuIDs) > 0 {
		if err := uc.repo.AssignRoleMenus(ctx, role.ID, r.MenuIDs); err != nil {
			uc.log.Warnf("assign role menus failed: %v", err)
		}
	}

	return role, nil
}

// GetRole 获取角色
func (uc *RoleUsecase) GetRole(ctx context.Context, id uint32) (*Role, error) {
	role, err := uc.repo.GetRoleByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 获取关联菜单ID
	menuIDs, err := uc.repo.GetRoleMenuIDs(ctx, id)
	if err == nil {
		role.MenuIDs = menuIDs
	}

	return role, nil
}

// GetRoleByCode 根据代码获取角色
func (uc *RoleUsecase) GetRoleByCode(ctx context.Context, code string) (*Role, error) {
	return uc.repo.GetRoleByCode(ctx, code)
}

// UpdateRole 更新角色
func (uc *RoleUsecase) UpdateRole(ctx context.Context, r *Role) (*Role, error) {
	role, err := uc.repo.UpdateRole(ctx, r)
	if err != nil {
		return nil, err
	}

	// 更新菜单权限
	if len(r.MenuIDs) > 0 {
		if err := uc.repo.AssignRoleMenus(ctx, r.ID, r.MenuIDs); err != nil {
			uc.log.Warnf("assign role menus failed: %v", err)
		}
	}

	return role, nil
}

// DeleteRole 删除角色
func (uc *RoleUsecase) DeleteRole(ctx context.Context, id uint32) error {
	return uc.repo.DeleteRole(ctx, id)
}

// ListRoles 列出角色
func (uc *RoleUsecase) ListRoles(ctx context.Context, status int8) ([]*Role, error) {
	return uc.repo.ListRoles(ctx, status)
}

// AdminLogUsecase 管理员日志用例
type AdminLogUsecase struct {
	repo AdminLogRepo
	log  *log.Helper
}

// NewAdminLogUsecase 创建日志用例
func NewAdminLogUsecase(repo AdminLogRepo, logger log.Logger) *AdminLogUsecase {
	return &AdminLogUsecase{repo: repo, log: log.NewHelper(logger)}
}

// CreateLog 创建日志
func (uc *AdminLogUsecase) CreateLog(ctx context.Context, log *AdminLog) error {
	return uc.repo.CreateLog(ctx, log)
}

// ListLogs 列出日志
func (uc *AdminLogUsecase) ListLogs(ctx context.Context, adminID uint32, module string, page, pageSize int) ([]*AdminLog, int64, error) {
	return uc.repo.ListLogs(ctx, adminID, module, page, pageSize)
}
