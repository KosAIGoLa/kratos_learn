package service

import (
	v1 "content/api/content/v1"
	"content/internal/biz"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ContentService 内容服务
type ContentService struct {
	v1.UnimplementedContentServer
	uc         *biz.ContentUsecase
	log        *log.Helper
	uploadPath string
	baseURL    string
}

// NewContentService 创建内容服务
func NewContentService(uc *biz.ContentUsecase, logger log.Logger) *ContentService {
	uploadPath := os.Getenv("UPLOAD_PATH")
	if uploadPath == "" {
		uploadPath = "./uploads"
	}
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8101"
	}
	// 确保上传目录存在
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		log.NewHelper(logger).Warnf("创建上传目录失败: %v", err)
	}
	return &ContentService{
		uc:         uc,
		log:        log.NewHelper(logger),
		uploadPath: uploadPath,
		baseURL:    baseURL,
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

// UploadImage 上传图片
func (s *ContentService) UploadImage(ctx context.Context, req *v1.UploadImageRequest) (*v1.UploadImageResponse, error) {
	// 从 context 中获取 multipart 文件 (Kratos HTTP 传输层会处理 multipart)
	// 这里 req.File 已经包含文件内容
	if len(req.File) == 0 {
		return &v1.UploadImageResponse{
			Success: false,
			Message: "请选择要上传的文件",
		}, nil
	}

	// 确定文件夹
	folder := req.Folder
	if folder == "" {
		folder = "images"
	}

	// 生成文件名
	ext := s.getFileExtension(req.File)
	if ext == "" {
		ext = ".jpg"
	}
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)

	// 创建文件夹
	dir := filepath.Join(s.uploadPath, folder)
	if err := os.MkdirAll(dir, 0755); err != nil {
		s.log.Errorf("创建目录失败: %v", err)
		return &v1.UploadImageResponse{
			Success: false,
			Message: "创建目录失败: " + err.Error(),
		}, nil
	}

	// 保存文件
	filepath := filepath.Join(dir, filename)
	file, err := os.Create(filepath)
	if err != nil {
		s.log.Errorf("创建文件失败: %v", err)
		return &v1.UploadImageResponse{
			Success: false,
			Message: "创建文件失败: " + err.Error(),
		}, nil
	}
	defer func() { _ = file.Close() }()

	if _, err := file.Write(req.File); err != nil {
		s.log.Errorf("写入文件失败: %v", err)
		return &v1.UploadImageResponse{
			Success: false,
			Message: "写入文件失败: " + err.Error(),
		}, nil
	}

	// 返回文件 URL
	fileURL := fmt.Sprintf("%s/uploads/%s/%s", s.baseURL, folder, filename)

	s.log.Infof("图片上传成功: %s", fileURL)

	return &v1.UploadImageResponse{
		Success:      true,
		Url:          fileURL,
		OriginalName: "", // 可以从 multipart 中获取
		FileName:     filename,
		FileSize:     int64(len(req.File)),
		Message:      "上传成功",
	}, nil
}

// getFileExtension 根据文件内容判断扩展名
func (s *ContentService) getFileExtension(data []byte) string {
	if len(data) < 4 {
		return ""
	}
	// 检查文件头
	if data[0] == 0xFF && data[1] == 0xD8 {
		return ".jpg"
	}
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return ".png"
	}
	if data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 {
		return ".gif"
	}
	if data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 {
		return ".webp"
	}
	return ".bin"
}
