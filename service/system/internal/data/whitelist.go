package data

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"system/internal/biz"
)

// WhitelistIP IP白名单数据模型
type WhitelistIP struct {
	ID          uint32     `gorm:"primarykey"`
	IP          string     `gorm:"index:idx_ip;type:varchar(45);not null"`
	Type        string     `gorm:"index:idx_type;type:varchar(20);default:'admin'"`
	Description string     `gorm:"type:varchar(255)"`
	Enabled     int8       `gorm:"index:idx_enabled;default:1"`
	ExpireAt    *time.Time `gorm:"index:idx_expire_at"`
	CreatedBy   uint32
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type whitelistRepo struct {
	data *Data
	log  *log.Helper
}

// NewWhitelistRepo 创建IP白名单仓库
func NewWhitelistRepo(data *Data, logger log.Logger) biz.WhitelistRepo {
	return &whitelistRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *whitelistRepo) CheckWhitelist(ctx context.Context, ip, typ string) (bool, error) {
	var count int64
	now := time.Now()

	query := r.data.db.Model(&WhitelistIP{}).
		Where("ip = ? AND type = ? AND enabled = 1", ip, typ).
		Where("expire_at IS NULL OR expire_at > ?", now)

	if err := query.Count(&count).Error; err != nil {
		return false, status.Errorf(codes.Internal, "%s", err.Error())
	}

	return count > 0, nil
}
