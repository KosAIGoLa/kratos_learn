package data

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"system/internal/biz"
)

// SystemConfig 系统配置数据模型
type SystemConfig struct {
	ID          uint32 `gorm:"primarykey"`
	Key         string `gorm:"uniqueIndex:idx_key;type:varchar(50);not null"`
	Value       string `gorm:"type:varchar(500);not null"`
	Description string `gorm:"type:varchar(255)"`
	Group       string `gorm:"index:idx_group;type:varchar(30);default:'default'"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

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

func (r *configRepo) GetConfig(ctx context.Context, key string) (*biz.SystemConfig, error) {
	var config SystemConfig
	if err := r.data.db.Where("key = ?", key).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "配置不存在")
		}
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}
	return &biz.SystemConfig{
		ID:          config.ID,
		Key:         config.Key,
		Value:       config.Value,
		Description: config.Description,
		Group:       config.Group,
		CreatedAt:   config.CreatedAt,
	}, nil
}

func (r *configRepo) SetConfig(ctx context.Context, c *biz.SystemConfig) (*biz.SystemConfig, error) {
	config := SystemConfig{
		Key:         c.Key,
		Value:       c.Value,
		Description: c.Description,
		Group:       c.Group,
	}
	if err := r.data.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"value",
			"description",
			"group",
			"updated_at",
		}),
	}).Create(&config).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}
	if err := r.data.db.WithContext(ctx).Where("key = ?", c.Key).First(&config).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}
	return &biz.SystemConfig{
		ID:          config.ID,
		Key:         config.Key,
		Value:       config.Value,
		Description: config.Description,
		Group:       config.Group,
	}, nil
}

func (r *configRepo) ListConfigs(ctx context.Context, group string) ([]*biz.SystemConfig, error) {
	var configs []SystemConfig
	query := r.data.db.Model(&SystemConfig{})
	if group != "" {
		query = query.Where("group = ?", group)
	}
	if err := query.Find(&configs).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}

	var bizConfigs []*biz.SystemConfig
	for _, c := range configs {
		bizConfigs = append(bizConfigs, &biz.SystemConfig{
			ID:          c.ID,
			Key:         c.Key,
			Value:       c.Value,
			Description: c.Description,
			Group:       c.Group,
			CreatedAt:   c.CreatedAt,
		})
	}
	return bizConfigs, nil
}
