package data

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"product/internal/biz"
)

// Product 产品数据模型
type Product struct {
	ID               uint32  `gorm:"primarykey"`
	MachineID        *uint32 `gorm:"index:idx_machine_id"`
	Name             string  `gorm:"type:varchar(255);not null"`
	Price            float64 `gorm:"type:decimal(10,2);not null"`
	Description      string  `gorm:"type:text"`
	Type             string  `gorm:"index:idx_type;type:varchar(20);default:'mining'"`
	Cycle            uint32  `gorm:"default:30"`
	ProductivityRate float64 `gorm:"type:decimal(5,4);default:0.0100"`
	Status           int8    `gorm:"index:idx_status;default:1"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type productRepo struct {
	data *Data
	log  *log.Helper
}

// NewProductRepo 创建产品仓库
func NewProductRepo(data *Data, logger log.Logger) biz.ProductRepo {
	return &productRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *productRepo) ListProducts(ctx context.Context, typ string, statusFilter int32, page, pageSize uint32) ([]*biz.Product, uint32, error) {
	var products []Product
	var total int64

	query := r.data.db.Model(&Product{})
	if typ != "" {
		query = query.Where("type = ?", typ)
	}
	if statusFilter >= 0 {
		query = query.Where("status = ?", statusFilter)
	}

	query.Count(&total)
	offset := (page - 1) * pageSize
	if err := query.Limit(int(pageSize)).Offset(int(offset)).Find(&products).Error; err != nil {
		return nil, 0, status.Errorf(codes.Internal, "%s", err.Error())
	}

	var bizProducts []*biz.Product
	for _, p := range products {
		bizProducts = append(bizProducts, r.toBizProduct(&p))
	}
	return bizProducts, uint32(total), nil
}

func (r *productRepo) GetProduct(ctx context.Context, id uint32) (*biz.Product, error) {
	var product Product
	if err := r.data.db.First(&product, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "产品不存在")
		}
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}
	return r.toBizProduct(&product), nil
}

func (r *productRepo) CreateProduct(ctx context.Context, p *biz.Product) (*biz.Product, error) {
	product := Product{
		MachineID:        p.MachineID,
		Name:             p.Name,
		Price:            p.Price,
		Description:      p.Description,
		Type:             p.Type,
		Cycle:            p.Cycle,
		ProductivityRate: p.ProductivityRate,
		Status:           1,
	}
	if err := r.data.db.Create(&product).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}
	return r.toBizProduct(&product), nil
}

func (r *productRepo) UpdateProduct(ctx context.Context, p *biz.Product) (*biz.Product, error) {
	updates := map[string]interface{}{}
	if p.Name != "" {
		updates["name"] = p.Name
	}
	if p.Price > 0 {
		updates["price"] = p.Price
	}
	if p.Status != 0 {
		updates["status"] = p.Status
	}
	if err := r.data.db.Model(&Product{}).Where("id = ?", p.ID).Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.GetProduct(ctx, p.ID)
}

func (r *productRepo) DeleteProduct(ctx context.Context, id uint32) error {
	if err := r.data.db.Delete(&Product{}, id).Error; err != nil {
		return status.Errorf(codes.Internal, "%s", err.Error())
	}
	return nil
}

func (r *productRepo) toBizProduct(p *Product) *biz.Product {
	return &biz.Product{
		ID:               p.ID,
		MachineID:        p.MachineID,
		Name:             p.Name,
		Price:            p.Price,
		Description:      p.Description,
		Type:             p.Type,
		Cycle:            p.Cycle,
		ProductivityRate: p.ProductivityRate,
		Status:           int32(p.Status),
		CreatedAt:        p.CreatedAt,
	}
}
