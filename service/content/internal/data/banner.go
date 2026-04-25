package data

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"content/internal/biz"
)

// Banner 轮播图数据模型
type Banner struct {
	ID         uint32     `gorm:"primarykey"`
	Title      string     `gorm:"type:varchar(100)"`
	Image      string     `gorm:"type:varchar(255);not null"`
	Link       string     `gorm:"type:varchar(255)"`
	Type       string     `gorm:"index:idx_type;type:varchar(20);default:'pc'"`
	Position   string     `gorm:"index:idx_position;type:varchar(30);default:'home'"`
	Sort       int32      `gorm:"index:idx_sort;default:0"`
	Status     int8       `gorm:"index:idx_status;default:1"`
	StartTime  *time.Time `gorm:"index:idx_start_time"`
	EndTime    *time.Time `gorm:"index:idx_end_time"`
	ClickCount uint32     `gorm:"default:0"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type bannerRepo struct {
	data *Data
	log  *log.Helper
}

// NewBannerRepo 创建轮播图仓库
func NewBannerRepo(data *Data, logger log.Logger) biz.BannerRepo {
	return &bannerRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *bannerRepo) ListBanners(ctx context.Context, typ, position string, statusFilter int32) ([]*biz.Banner, error) {
	var banners []Banner
	query := r.data.db.Model(&Banner{})
	if typ != "" {
		query = query.Where("type = ?", typ)
	}
	if position != "" {
		query = query.Where("position = ?", position)
	}
	if statusFilter >= 0 {
		query = query.Where("status = ?", statusFilter)
	}
	now := time.Now()
	query = query.Where("(start_time IS NULL OR start_time <= ?) AND (end_time IS NULL OR end_time >= ?)", now, now)

	if err := query.Order("sort ASC").Find(&banners).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query banners: %v", err)
	}

	var bizBanners []*biz.Banner
	for _, b := range banners {
		bizBanners = append(bizBanners, r.toBizBanner(&b))
	}
	return bizBanners, nil
}

func (r *bannerRepo) GetBanner(ctx context.Context, id uint32) (*biz.Banner, error) {
	var banner Banner
	if err := r.data.db.First(&banner, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "轮播图不存在")
		}
		return nil, status.Errorf(codes.Internal, "failed to get banner: %v", err)
	}
	return r.toBizBanner(&banner), nil
}

func (r *bannerRepo) CreateBanner(ctx context.Context, b *biz.Banner) (*biz.Banner, error) {
	banner := Banner{
		Title:     b.Title,
		Image:     b.Image,
		Link:      b.Link,
		Type:      b.Type,
		Position:  b.Position,
		Sort:      b.Sort,
		StartTime: b.StartTime,
		EndTime:   b.EndTime,
		Status:    1,
	}
	if err := r.data.db.Create(&banner).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create banner: %v", err)
	}
	return r.toBizBanner(&banner), nil
}

func (r *bannerRepo) UpdateBanner(ctx context.Context, b *biz.Banner) (*biz.Banner, error) {
	updates := map[string]interface{}{}
	if b.Title != "" {
		updates["title"] = b.Title
	}
	if b.Image != "" {
		updates["image"] = b.Image
	}
	if b.Link != "" {
		updates["link"] = b.Link
	}
	if b.Sort != 0 {
		updates["sort"] = b.Sort
	}
	if b.Status >= 0 {
		updates["status"] = b.Status
	}
	if err := r.data.db.Model(&Banner{}).Where("id = ?", b.ID).Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.GetBanner(ctx, b.ID)
}

func (r *bannerRepo) DeleteBanner(ctx context.Context, id uint32) error {
	return r.data.db.Delete(&Banner{}, id).Error
}

func (r *bannerRepo) RecordClick(ctx context.Context, id uint32) error {
	return r.data.db.Model(&Banner{}).Where("id = ?", id).UpdateColumn("click_count", gorm.Expr("click_count + 1")).Error
}

func (r *bannerRepo) toBizBanner(b *Banner) *biz.Banner {
	return &biz.Banner{
		ID:         b.ID,
		Title:      b.Title,
		Image:      b.Image,
		Link:       b.Link,
		Type:       b.Type,
		Position:   b.Position,
		Sort:       b.Sort,
		Status:     int32(b.Status),
		StartTime:  b.StartTime,
		EndTime:    b.EndTime,
		ClickCount: b.ClickCount,
	}
}
