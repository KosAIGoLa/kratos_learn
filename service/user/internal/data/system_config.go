package data

import (
	"context"
	"strconv"

	"user/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// SystemConfig 系统配置数据模型
type SystemConfig struct {
	ID          uint32 `gorm:"primarykey"`
	Key         string `gorm:"uniqueIndex:idx_key;type:varchar(50);not null"`
	Value       string `gorm:"type:varchar(500);not null"`
	Description string `gorm:"type:varchar(200)"`
	Group       string `gorm:"index:idx_group;type:varchar(50)"`
}

// configRepo 配置仓库
type configRepo struct {
	data *Data
	log  *log.Helper
}

// NewConfigRepo 创建配置仓库
func NewConfigRepo(data *Data, logger log.Logger) biz.ConfigRepo {
	return &configRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// GetConfig 根据 key 获取配置
func (r *configRepo) GetConfig(ctx context.Context, key string) (*biz.SystemConfig, error) {
	var config SystemConfig
	if err := r.data.db.Where("key = ?", key).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "配置不存在: %s", key)
		}
		return nil, status.Errorf(codes.Internal, "查询配置失败: %s", err.Error())
	}
	return &biz.SystemConfig{
		ID:          config.ID,
		Key:         config.Key,
		Value:       config.Value,
		Description: config.Description,
		Group:       config.Group,
	}, nil
}

// GetConfigValue 获取配置值（字符串）
func (r *configRepo) GetConfigValue(ctx context.Context, key string) (string, error) {
	config, err := r.GetConfig(ctx, key)
	if err != nil {
		return "", err
	}
	return config.Value, nil
}

// GetConfigInt 获取配置值（整数）
func (r *configRepo) GetConfigInt(ctx context.Context, key string) (int, error) {
	value, err := r.GetConfigValue(ctx, key)
	if err != nil {
		return 0, err
	}
	intVal, err := strconv.Atoi(value)
	if err != nil {
		return 0, status.Errorf(codes.Internal, "配置值转换失败: %s", err.Error())
	}
	return intVal, nil
}

// GetConfigFloat 获取配置值（浮点数）
func (r *configRepo) GetConfigFloat(ctx context.Context, key string) (float64, error) {
	value, err := r.GetConfigValue(ctx, key)
	if err != nil {
		return 0, err
	}
	floatVal, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, status.Errorf(codes.Internal, "配置值转换失败: %s", err.Error())
	}
	return floatVal, nil
}
