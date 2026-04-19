package biz

import (
	"context"
	"time"
)

// SystemLog 系统日志实体
type SystemLog struct {
	ID           string
	Level        string            // debug/info/warn/error/fatal
	Module       string            // 模块名称
	Action       string            // 操作类型
	Message      string            // 日志内容
	OperatorID   string            // 操作者ID
	OperatorName string            // 操作者名称
	IPAddress    string            // IP地址
	UserAgent    string            // 用户代理
	Metadata     map[string]string // 额外元数据
	CreatedAt    time.Time
}

// UserLog 用户日志实体
type UserLog struct {
	ID          string
	UserID      uint32            // 用户ID
	Username    string            // 用户名
	Action      string            // 操作类型
	Module      string            // 模块
	Description string            // 操作描述
	IPAddress   string            // IP地址
	DeviceInfo  string            // 设备信息
	Metadata    map[string]string // 额外数据
	CreatedAt   time.Time
}

// SystemLogRepo 系统日志存储接口
type SystemLogRepo interface {
	// ListSystemLogs 查询系统日志列表
	ListSystemLogs(ctx context.Context, level, module, operatorID, startTime, endTime string, page, pageSize int32) ([]*SystemLog, int32, error)
	// GetSystemLog 获取单条系统日志
	GetSystemLog(ctx context.Context, id string) (*SystemLog, error)
	// CreateSystemLog 创建系统日志
	CreateSystemLog(ctx context.Context, log *SystemLog) error
}

// UserLogRepo 用户日志存储接口
type UserLogRepo interface {
	// ListUserLogs 查询用户日志列表
	ListUserLogs(ctx context.Context, userID uint32, action, module, startTime, endTime string, page, pageSize int32) ([]*UserLog, int32, error)
	// GetUserLog 获取单条用户日志
	GetUserLog(ctx context.Context, id string) (*UserLog, error)
	// CreateUserLog 创建用户日志
	CreateUserLog(ctx context.Context, log *UserLog) error
}

// LogUsecase 日志用例
type LogUsecase struct {
	systemLogRepo SystemLogRepo
	userLogRepo   UserLogRepo
}

// NewLogUsecase 创建日志用例
func NewLogUsecase(systemLogRepo SystemLogRepo, userLogRepo UserLogRepo) *LogUsecase {
	return &LogUsecase{
		systemLogRepo: systemLogRepo,
		userLogRepo:   userLogRepo,
	}
}

// ListSystemLogs 查询系统日志
func (uc *LogUsecase) ListSystemLogs(ctx context.Context, level, module, operatorID, startTime, endTime string, page, pageSize int32) ([]*SystemLog, int32, error) {
	return uc.systemLogRepo.ListSystemLogs(ctx, level, module, operatorID, startTime, endTime, page, pageSize)
}

// GetSystemLog 获取系统日志
func (uc *LogUsecase) GetSystemLog(ctx context.Context, id string) (*SystemLog, error) {
	return uc.systemLogRepo.GetSystemLog(ctx, id)
}

// CreateSystemLog 创建系统日志
func (uc *LogUsecase) CreateSystemLog(ctx context.Context, log *SystemLog) error {
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}
	return uc.systemLogRepo.CreateSystemLog(ctx, log)
}

// ListUserLogs 查询用户日志
func (uc *LogUsecase) ListUserLogs(ctx context.Context, userID uint32, action, module, startTime, endTime string, page, pageSize int32) ([]*UserLog, int32, error) {
	return uc.userLogRepo.ListUserLogs(ctx, userID, action, module, startTime, endTime, page, pageSize)
}

// GetUserLog 获取用户日志
func (uc *LogUsecase) GetUserLog(ctx context.Context, id string) (*UserLog, error) {
	return uc.userLogRepo.GetUserLog(ctx, id)
}

// CreateUserLog 创建用户日志
func (uc *LogUsecase) CreateUserLog(ctx context.Context, log *UserLog) error {
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}
	return uc.userLogRepo.CreateUserLog(ctx, log)
}
