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

// ProfitSharingRule 分账规则数据模型
type ProfitSharingRule struct {
	ID          uint32  `gorm:"primarykey"`
	Name        string  `gorm:"type:varchar(50);not null"`
	Type        string  `gorm:"index:idx_type;type:varchar(30);not null"`
	Level       int8    `gorm:"index:idx_level;not null"`
	Ratio       float64 `gorm:"type:decimal(5,4);not null"`
	FixedAmount float64 `gorm:"type:decimal(10,2);default:0.00"`
	Unit        string  `gorm:"type:varchar(10);default:'point'"`
	MinAmount   float64 `gorm:"type:decimal(10,2);default:0.00"`
	MaxAmount   float64 `gorm:"type:decimal(10,2)"`
	Sort        int32   `gorm:"default:0"`
	Enabled     int8    `gorm:"index:idx_enabled;default:1"`
	Description string  `gorm:"type:varchar(255)"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type profitSharingRepo struct {
	data *Data
	log  *log.Helper
}

// NewProfitSharingRepo 创建分账规则仓库
func NewProfitSharingRepo(data *Data, logger log.Logger) biz.ProfitSharingRepo {
	return &profitSharingRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *profitSharingRepo) ListRules(ctx context.Context, typ string, enabled int32) ([]*biz.ProfitSharingRule, error) {
	var rules []ProfitSharingRule
	query := r.data.db.Model(&ProfitSharingRule{})
	if typ != "" {
		query = query.Where("type = ?", typ)
	}
	if enabled >= 0 {
		query = query.Where("enabled = ?", enabled)
	}
	if err := query.Order("sort ASC").Find(&rules).Error; err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	var bizRules []*biz.ProfitSharingRule
	for _, rule := range rules {
		bizRules = append(bizRules, r.toBizRule(&rule))
	}
	return bizRules, nil
}

func (r *profitSharingRepo) GetRule(ctx context.Context, id uint32) (*biz.ProfitSharingRule, error) {
	var rule ProfitSharingRule
	if err := r.data.db.First(&rule, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "分账规则不存在")
		}
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return r.toBizRule(&rule), nil
}

func (r *profitSharingRepo) CreateRule(ctx context.Context, rule *biz.ProfitSharingRule) (*biz.ProfitSharingRule, error) {
	r2 := ProfitSharingRule{
		Name:        rule.Name,
		Type:        rule.Type,
		Level:       int8(rule.Level),
		Ratio:       rule.Ratio,
		FixedAmount: rule.FixedAmount,
		Unit:        rule.Unit,
		MinAmount:   rule.MinAmount,
		MaxAmount:   rule.MaxAmount,
		Sort:        rule.Sort,
		Enabled:     1,
		Description: rule.Description,
	}
	if err := r.data.db.Create(&r2).Error; err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return r.toBizRule(&r2), nil
}

func (r *profitSharingRepo) UpdateRule(ctx context.Context, rule *biz.ProfitSharingRule) (*biz.ProfitSharingRule, error) {
	updates := map[string]interface{}{}
	if rule.Name != "" {
		updates["name"] = rule.Name
	}
	if rule.Ratio > 0 {
		updates["ratio"] = rule.Ratio
	}
	if rule.Enabled >= 0 {
		updates["enabled"] = rule.Enabled
	}
	if err := r.data.db.Model(&ProfitSharingRule{}).Where("id = ?", rule.ID).Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.GetRule(ctx, rule.ID)
}

func (r *profitSharingRepo) DeleteRule(ctx context.Context, id uint32) error {
	return r.data.db.Delete(&ProfitSharingRule{}, id).Error
}

func (r *profitSharingRepo) toBizRule(rule *ProfitSharingRule) *biz.ProfitSharingRule {
	return &biz.ProfitSharingRule{
		ID:          rule.ID,
		Name:        rule.Name,
		Type:        rule.Type,
		Level:       int32(rule.Level),
		Ratio:       rule.Ratio,
		FixedAmount: rule.FixedAmount,
		Unit:        rule.Unit,
		MinAmount:   rule.MinAmount,
		MaxAmount:   rule.MaxAmount,
		Sort:        rule.Sort,
		Enabled:     int32(rule.Enabled),
		Description: rule.Description,
	}
}
