package service

import (
	v1 "content/api/content/v1"
	"content/internal/biz"
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ContentService 内容服务
type ContentService struct {
	v1.UnimplementedContentServer
	uc  *biz.ContentUsecase
	log *log.Helper
}

// NewContentService 创建内容服务
func NewContentService(uc *biz.ContentUsecase, logger log.Logger) *ContentService {
	return &ContentService{
		uc:  uc,
		log: log.NewHelper(logger),
	}
}

// ListNews 获取新闻列表
func (s *ContentService) ListNews(ctx context.Context, req *v1.ListNewsRequest) (*v1.ListNewsResponse, error) {
	news, total, err := s.uc.ListNews(ctx, req.Category, req.Type, req.Status, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	var protoNews []*v1.NewsInfo
	for _, n := range news {
		protoNews = append(protoNews, s.toProtoNews(n))
	}

	return &v1.ListNewsResponse{
		News:  protoNews,
		Total: total,
	}, nil
}

// GetNews 获取新闻详情
func (s *ContentService) GetNews(ctx context.Context, req *v1.GetNewsRequest) (*v1.NewsInfo, error) {
	news, err := s.uc.GetNews(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return s.toProtoNews(news), nil
}

// CreateNews 创建新闻
func (s *ContentService) CreateNews(ctx context.Context, req *v1.CreateNewsRequest) (*v1.NewsInfo, error) {
	news, err := s.uc.CreateNews(ctx, &biz.News{
		Title:       req.Title,
		Summary:     req.Summary,
		Content:     req.Content,
		CoverImage:  req.CoverImage,
		Category:    req.Category,
		Type:        req.Type,
		Author:      req.Author,
		AdminID:     req.AdminId,
		Sort:        req.Sort,
		IsTop:       req.IsTop,
		PublishTime: req.PublishTime.AsTime(),
	})
	if err != nil {
		return nil, err
	}
	return s.toProtoNews(news), nil
}

// UpdateNews 更新新闻
func (s *ContentService) UpdateNews(ctx context.Context, req *v1.UpdateNewsRequest) (*v1.NewsInfo, error) {
	news, err := s.uc.UpdateNews(ctx, &biz.News{
		ID:         req.Id,
		Title:      req.Title,
		Summary:    req.Summary,
		Content:    req.Content,
		CoverImage: req.CoverImage,
		Status:     req.Status,
		IsTop:      req.IsTop,
	})
	if err != nil {
		return nil, err
	}
	return s.toProtoNews(news), nil
}

// DeleteNews 删除新闻
func (s *ContentService) DeleteNews(ctx context.Context, req *v1.DeleteNewsRequest) (*v1.DeleteNewsResponse, error) {
	if err := s.uc.DeleteNews(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.DeleteNewsResponse{Success: true}, nil
}

// ListBanners 获取轮播图列表
func (s *ContentService) ListBanners(ctx context.Context, req *v1.ListBannersRequest) (*v1.ListBannersResponse, error) {
	banners, err := s.uc.ListBanners(ctx, req.Type, req.Position, req.Status)
	if err != nil {
		return nil, err
	}

	var protoBanners []*v1.BannerInfo
	for _, b := range banners {
		protoBanners = append(protoBanners, s.toProtoBanner(b))
	}

	return &v1.ListBannersResponse{
		Banners: protoBanners,
	}, nil
}

// GetBanner 获取轮播图
func (s *ContentService) GetBanner(ctx context.Context, req *v1.GetBannerRequest) (*v1.BannerInfo, error) {
	banner, err := s.uc.GetBanner(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return s.toProtoBanner(banner), nil
}

// CreateBanner 创建轮播图
func (s *ContentService) CreateBanner(ctx context.Context, req *v1.CreateBannerRequest) (*v1.BannerInfo, error) {
	banner, err := s.uc.CreateBanner(ctx, &biz.Banner{
		Title:    req.Title,
		Image:    req.Image,
		Link:     req.Link,
		Type:     req.Type,
		Position: req.Position,
		Sort:     req.Sort,
		StartTime: func() *time.Time {
			if req.StartTime != nil {
				t := req.StartTime.AsTime()
				return &t
			}
			return nil
		}(),
		EndTime: func() *time.Time {
			if req.EndTime != nil {
				t := req.EndTime.AsTime()
				return &t
			}
			return nil
		}(),
	})
	if err != nil {
		return nil, err
	}
	return s.toProtoBanner(banner), nil
}

// UpdateBanner 更新轮播图
func (s *ContentService) UpdateBanner(ctx context.Context, req *v1.UpdateBannerRequest) (*v1.BannerInfo, error) {
	banner, err := s.uc.UpdateBanner(ctx, &biz.Banner{
		ID:     req.Id,
		Title:  req.Title,
		Image:  req.Image,
		Link:   req.Link,
		Sort:   req.Sort,
		Status: req.Status,
	})
	if err != nil {
		return nil, err
	}
	return s.toProtoBanner(banner), nil
}

// DeleteBanner 删除轮播图
func (s *ContentService) DeleteBanner(ctx context.Context, req *v1.DeleteBannerRequest) (*v1.DeleteBannerResponse, error) {
	if err := s.uc.DeleteBanner(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.DeleteBannerResponse{Success: true}, nil
}

// RecordClick 记录点击
func (s *ContentService) RecordClick(ctx context.Context, req *v1.RecordClickRequest) (*v1.RecordClickResponse, error) {
	if err := s.uc.RecordClick(ctx, req.BannerId); err != nil {
		return nil, err
	}
	return &v1.RecordClickResponse{Success: true}, nil
}

func (s *ContentService) toProtoNews(n *biz.News) *v1.NewsInfo {
	return &v1.NewsInfo{
		Id:          n.ID,
		Title:       n.Title,
		Summary:     n.Summary,
		Content:     n.Content,
		CoverImage:  n.CoverImage,
		Category:    n.Category,
		Type:        n.Type,
		Author:      n.Author,
		ViewCount:   n.ViewCount,
		Sort:        n.Sort,
		Status:      n.Status,
		IsTop:       n.IsTop,
		PublishTime: timestamppb.New(n.PublishTime),
		ExpireTime: func() *timestamppb.Timestamp {
			if n.ExpireTime != nil {
				return timestamppb.New(*n.ExpireTime)
			}
			return nil
		}(),
	}
}

func (s *ContentService) toProtoBanner(b *biz.Banner) *v1.BannerInfo {
	return &v1.BannerInfo{
		Id:       b.ID,
		Title:    b.Title,
		Image:    b.Image,
		Link:     b.Link,
		Type:     b.Type,
		Position: b.Position,
		Sort:     b.Sort,
		Status:   b.Status,
		StartTime: func() *timestamppb.Timestamp {
			if b.StartTime != nil {
				return timestamppb.New(*b.StartTime)
			}
			return nil
		}(),
		EndTime: func() *timestamppb.Timestamp {
			if b.EndTime != nil {
				return timestamppb.New(*b.EndTime)
			}
			return nil
		}(),
		ClickCount: b.ClickCount,
	}
}
