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

// Domain 域名数据模型
type Domain struct {
	ID        uint32 `gorm:"primarykey"`
	Domain    string `gorm:"uniqueIndex:idx_domain;type:varchar(100);not null"`
	Enabled   int8   `gorm:"index:idx_enabled;default:1"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type domainRepo struct {
	data *Data
	log  *log.Helper
}

// NewDomainRepo 创建域名仓库
func NewDomainRepo(data *Data, logger log.Logger) biz.DomainRepo {
	return &domainRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *domainRepo) ListDomains(ctx context.Context, enabled int32) ([]*biz.Domain, error) {
	var domains []Domain
	query := r.data.db.Model(&Domain{})
	if enabled >= 0 {
		query = query.Where("enabled = ?", enabled)
	}
	if err := query.Find(&domains).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}

	var bizDomains []*biz.Domain
	for _, d := range domains {
		bizDomains = append(bizDomains, &biz.Domain{
			ID:        d.ID,
			Domain:    d.Domain,
			Enabled:   int32(d.Enabled),
			CreatedAt: d.CreatedAt,
		})
	}
	return bizDomains, nil
}

func (r *domainRepo) AddDomain(ctx context.Context, d *biz.Domain) (*biz.Domain, error) {
	domain := Domain{
		Domain:  d.Domain,
		Enabled: 1,
	}
	if err := r.data.db.Create(&domain).Error; err != nil {
		if err == gorm.ErrDuplicatedKey {
			return nil, status.Errorf(codes.AlreadyExists, "域名已存在")
		}
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}
	return &biz.Domain{
		ID:        domain.ID,
		Domain:    domain.Domain,
		Enabled:   int32(domain.Enabled),
		CreatedAt: domain.CreatedAt,
	}, nil
}
