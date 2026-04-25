package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// News 新闻领域模型
type News struct {
	ID          uint32
	Title       string
	Summary     string
	Content     string
	CoverImage  string
	Category    string
	Type        string
	Author      string
	AdminID     *uint32
	ViewCount   uint32
	Sort        int32
	Status      int32
	IsTop       int32
	PublishTime time.Time
	ExpireTime  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Banner 轮播图领域模型
type Banner struct {
	ID         uint32
	Title      string
	Image      string
	Link       string
	Type       string
	Position   string
	Sort       int32
	Status     int32
	StartTime  *time.Time
	EndTime    *time.Time
	ClickCount uint32
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewsRepo 新闻存储接口
type NewsRepo interface {
	ListNews(ctx context.Context, category, typ string, status int32, page, pageSize uint32) ([]*News, uint32, error)
	GetNews(ctx context.Context, id uint32) (*News, error)
	CreateNews(ctx context.Context, n *News) (*News, error)
	UpdateNews(ctx context.Context, n *News) (*News, error)
	DeleteNews(ctx context.Context, id uint32) error
	IncrementViewCount(ctx context.Context, id uint32) error
}

// BannerRepo 轮播图存储接口
type BannerRepo interface {
	ListBanners(ctx context.Context, typ, position string, status int32) ([]*Banner, error)
	GetBanner(ctx context.Context, id uint32) (*Banner, error)
	CreateBanner(ctx context.Context, b *Banner) (*Banner, error)
	UpdateBanner(ctx context.Context, b *Banner) (*Banner, error)
	DeleteBanner(ctx context.Context, id uint32) error
	RecordClick(ctx context.Context, id uint32) error
}

// ContentUsecase 内容用例
type ContentUsecase struct {
	newsRepo   NewsRepo
	bannerRepo BannerRepo
	log        *log.Helper
}

// NewContentUsecase 创建内容用例
func NewContentUsecase(newsRepo NewsRepo, bannerRepo BannerRepo, logger log.Logger) *ContentUsecase {
	return &ContentUsecase{
		newsRepo:   newsRepo,
		bannerRepo: bannerRepo,
		log:        log.NewHelper(logger),
	}
}

// ListNews 获取新闻列表
func (uc *ContentUsecase) ListNews(ctx context.Context, category, typ string, status int32, page, pageSize uint32) ([]*News, uint32, error) {
	return uc.newsRepo.ListNews(ctx, category, typ, status, page, pageSize)
}

// GetNews 获取新闻详情
func (uc *ContentUsecase) GetNews(ctx context.Context, id uint32) (*News, error) {
	// 增加浏览量
	_ = uc.newsRepo.IncrementViewCount(ctx, id)
	return uc.newsRepo.GetNews(ctx, id)
}

// CreateNews 创建新闻
func (uc *ContentUsecase) CreateNews(ctx context.Context, n *News) (*News, error) {
	return uc.newsRepo.CreateNews(ctx, n)
}

// UpdateNews 更新新闻
func (uc *ContentUsecase) UpdateNews(ctx context.Context, n *News) (*News, error) {
	return uc.newsRepo.UpdateNews(ctx, n)
}

// DeleteNews 删除新闻
func (uc *ContentUsecase) DeleteNews(ctx context.Context, id uint32) error {
	return uc.newsRepo.DeleteNews(ctx, id)
}

// ListBanners 获取轮播图列表
func (uc *ContentUsecase) ListBanners(ctx context.Context, typ, position string, status int32) ([]*Banner, error) {
	return uc.bannerRepo.ListBanners(ctx, typ, position, status)
}

// GetBanner 获取轮播图
func (uc *ContentUsecase) GetBanner(ctx context.Context, id uint32) (*Banner, error) {
	return uc.bannerRepo.GetBanner(ctx, id)
}

// CreateBanner 创建轮播图
func (uc *ContentUsecase) CreateBanner(ctx context.Context, b *Banner) (*Banner, error) {
	return uc.bannerRepo.CreateBanner(ctx, b)
}

// UpdateBanner 更新轮播图
func (uc *ContentUsecase) UpdateBanner(ctx context.Context, b *Banner) (*Banner, error) {
	return uc.bannerRepo.UpdateBanner(ctx, b)
}

// DeleteBanner 删除轮播图
func (uc *ContentUsecase) DeleteBanner(ctx context.Context, id uint32) error {
	return uc.bannerRepo.DeleteBanner(ctx, id)
}

// RecordClick 记录点击
func (uc *ContentUsecase) RecordClick(ctx context.Context, id uint32) error {
	return uc.bannerRepo.RecordClick(ctx, id)
}
