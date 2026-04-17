package data

import (
	"content/internal/biz"
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// News 新闻数据模型
type News struct {
	ID          uint32    `gorm:"primarykey"`
	Title       string    `gorm:"type:varchar(200);not null"`
	Summary     string    `gorm:"type:varchar(500)"`
	Content     string    `gorm:"type:text;not null"`
	CoverImage  string    `gorm:"type:varchar(255)"`
	Category    string    `gorm:"index:idx_category;type:varchar(30);default:'announcement'"`
	Type        string    `gorm:"index:idx_type;type:varchar(20);default:'normal'"`
	Author      string    `gorm:"type:varchar(50);default:'admin'"`
	AdminID     uint32    `gorm:"index:idx_admin_id"`
	ViewCount   uint32    `gorm:"default:0"`
	Sort        int32     `gorm:"index:idx_sort;default:0"`
	Status      int8      `gorm:"index:idx_status;default:1"`
	IsTop       int8      `gorm:"index:idx_is_top;default:0"`
	PublishTime time.Time `gorm:"index:idx_publish_time"`
	ExpireTime  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type newsRepo struct {
	data *Data
	log  *log.Helper
}

// NewNewsRepo 创建新闻仓库
func NewNewsRepo(data *Data, logger log.Logger) biz.NewsRepo {
	return &newsRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *newsRepo) ListNews(ctx context.Context, category, typ string, newsStatus int32, page, pageSize uint32) ([]*biz.News, uint32, error) {
	var news []News
	var total int64

	query := r.data.db.Model(&News{})
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if typ != "" {
		query = query.Where("type = ?", typ)
	}
	if newsStatus >= 0 {
		query = query.Where("status = ?", newsStatus)
	}

	query.Count(&total)
	offset := (page - 1) * pageSize
	if err := query.Order("is_top DESC, sort ASC, publish_time DESC").
		Limit(int(pageSize)).Offset(int(offset)).Find(&news).Error; err != nil {
		return nil, 0, status.Errorf(codes.Internal, err.Error())
	}

	var bizNews []*biz.News
	for _, n := range news {
		bizNews = append(bizNews, r.toBizNews(&n))
	}
	return bizNews, uint32(total), nil
}

func (r *newsRepo) GetNews(ctx context.Context, id uint32) (*biz.News, error) {
	var news News
	if err := r.data.db.First(&news, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "新闻不存在")
		}
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return r.toBizNews(&news), nil
}

func (r *newsRepo) CreateNews(ctx context.Context, n *biz.News) (*biz.News, error) {
	news := News{
		Title:       n.Title,
		Summary:     n.Summary,
		Content:     n.Content,
		CoverImage:  n.CoverImage,
		Category:    n.Category,
		Type:        n.Type,
		Author:      n.Author,
		AdminID:     n.AdminID,
		Sort:        n.Sort,
		IsTop:       int8(n.IsTop),
		PublishTime: n.PublishTime,
		ExpireTime:  n.ExpireTime,
		Status:      1,
	}
	if err := r.data.db.Create(&news).Error; err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return r.toBizNews(&news), nil
}

func (r *newsRepo) UpdateNews(ctx context.Context, n *biz.News) (*biz.News, error) {
	updates := map[string]interface{}{}
	if n.Title != "" {
		updates["title"] = n.Title
	}
	if n.Content != "" {
		updates["content"] = n.Content
	}
	if n.Status >= 0 {
		updates["status"] = n.Status
	}
	if err := r.data.db.Model(&News{}).Where("id = ?", n.ID).Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.GetNews(ctx, n.ID)
}

func (r *newsRepo) DeleteNews(ctx context.Context, id uint32) error {
	return r.data.db.Delete(&News{}, id).Error
}

func (r *newsRepo) IncrementViewCount(ctx context.Context, id uint32) error {
	return r.data.db.Model(&News{}).Where("id = ?", id).UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

func (r *newsRepo) toBizNews(n *News) *biz.News {
	return &biz.News{
		ID:          n.ID,
		Title:       n.Title,
		Summary:     n.Summary,
		Content:     n.Content,
		CoverImage:  n.CoverImage,
		Category:    n.Category,
		Type:        n.Type,
		Author:      n.Author,
		ViewCount:   n.ViewCount,
		Sort:        n.Sort,
		Status:      int32(n.Status),
		IsTop:       int32(n.IsTop),
		PublishTime: n.PublishTime,
		ExpireTime:  n.ExpireTime,
	}
}
