package data

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"system/internal/biz"
)

// RiskControl 风控规则数据模型
type RiskControl struct {
	ID               uint32 `gorm:"primarykey"`
	Name             string `gorm:"type:varchar(50);not null"`
	Code             string `gorm:"uniqueIndex:idx_code;type:varchar(50);not null"`
	Type             string `gorm:"index:idx_type;type:varchar(30);not null"`
	Level            string `gorm:"index:idx_level;type:varchar(20);default:'medium'"`
	TriggerCondition string `gorm:"type:text;not null"`
	Action           string `gorm:"type:varchar(50);not null"`
	LimitValue       uint32
	TimeWindow       uint32 `gorm:"default:3600"`
	Enabled          int8   `gorm:"index:idx_enabled;default:1"`
	Description      string `gorm:"type:varchar(255)"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type riskControlRepo struct {
	data *Data
	log  *log.Helper
}

// NewRiskControlRepo 创建风控仓库
func NewRiskControlRepo(data *Data, logger log.Logger) biz.RiskControlRepo {
	return &riskControlRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *riskControlRepo) ListRiskControls(ctx context.Context, typ string, enabled int32) ([]*biz.RiskControl, error) {
	var controls []RiskControl
	query := r.data.db.Model(&RiskControl{})
	if typ != "" {
		query = query.Where("type = ?", typ)
	}
	if enabled >= 0 {
		query = query.Where("enabled = ?", enabled)
	}
	if err := query.Find(&controls).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}

	var bizControls []*biz.RiskControl
	for _, c := range controls {
		bizControls = append(bizControls, r.toBizControl(&c))
	}
	return bizControls, nil
}

func (r *riskControlRepo) GetRiskControl(ctx context.Context, id uint32) (*biz.RiskControl, error) {
	var control RiskControl
	if err := r.data.db.First(&control, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "风控规则不存在")
		}
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}
	return r.toBizControl(&control), nil
}

func (r *riskControlRepo) CreateRiskControl(ctx context.Context, c *biz.RiskControl) (*biz.RiskControl, error) {
	control := RiskControl{
		Name:             c.Name,
		Code:             c.Code,
		Type:             c.Type,
		Level:            c.Level,
		TriggerCondition: c.TriggerCondition,
		Action:           c.Action,
		LimitValue:       c.LimitValue,
		TimeWindow:       c.TimeWindow,
		Enabled:          1,
		Description:      c.Description,
	}
	if err := r.data.db.Create(&control).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}
	return r.toBizControl(&control), nil
}

func (r *riskControlRepo) CheckFrequency(ctx context.Context, key string, window uint32, limit uint32) (uint32, error) {
	pipe := r.data.rdb.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, time.Duration(window)*time.Second)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, status.Errorf(codes.Internal, "redis pipeline failed: %s", err.Error())
	}
	return uint32(incr.Val()), nil
}

func (r *riskControlRepo) UpdateRiskControl(ctx context.Context, c *biz.RiskControl) (*biz.RiskControl, error) {
	updates := map[string]interface{}{}
	if c.Name != "" {
		updates["name"] = c.Name
	}
	if c.Enabled >= 0 {
		updates["enabled"] = c.Enabled
	}
	if err := r.data.db.Model(&RiskControl{}).Where("id = ?", c.ID).Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.GetRiskControl(ctx, c.ID)
}

func (r *riskControlRepo) toBizControl(c *RiskControl) *biz.RiskControl {
	return &biz.RiskControl{
		ID:               c.ID,
		Name:             c.Name,
		Code:             c.Code,
		Type:             c.Type,
		Level:            c.Level,
		TriggerCondition: c.TriggerCondition,
		Action:           c.Action,
		LimitValue:       c.LimitValue,
		TimeWindow:       c.TimeWindow,
		Enabled:          int32(c.Enabled),
		Description:      c.Description,
	}
}
