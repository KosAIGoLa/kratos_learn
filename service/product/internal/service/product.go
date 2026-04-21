package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	v1 "product/api/product/v1"
	"product/internal/biz"
)

// ProductService 产品服务
type ProductService struct {
	v1.UnimplementedProductServer
	uc  *biz.ProductUsecase
	log *log.Helper
}

// NewProductService 创建产品服务
func NewProductService(uc *biz.ProductUsecase, logger log.Logger) *ProductService {
	return &ProductService{
		uc:  uc,
		log: log.NewHelper(logger),
	}
}

// ListProducts 获取产品列表
func (s *ProductService) ListProducts(ctx context.Context, req *v1.ListProductsRequest) (*v1.ListProductsResponse, error) {
	products, total, err := s.uc.ListProducts(ctx, req.Type, req.Status, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	var protoProducts []*v1.ProductInfo
	for _, p := range products {
		protoProducts = append(protoProducts, s.toProtoProduct(p))
	}

	return &v1.ListProductsResponse{
		Products: protoProducts,
		Total:    total,
	}, nil
}

// GetProduct 获取产品详情
func (s *ProductService) GetProduct(ctx context.Context, req *v1.GetProductRequest) (*v1.ProductInfo, error) {
	product, err := s.uc.GetProduct(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return s.toProtoProduct(product), nil
}

// CreateProduct 创建产品
func (s *ProductService) CreateProduct(ctx context.Context, req *v1.CreateProductRequest) (*v1.ProductInfo, error) {
	product, err := s.uc.CreateProduct(ctx, &biz.Product{
		MachineID: func() *uint32 {
			if req.MachineId > 0 {
				v := req.MachineId
				return &v
			}
			return nil
		}(),
		Name:             req.Name,
		Price:            req.Price,
		Description:      req.Description,
		Type:             req.Type,
		Cycle:            req.Cycle,
		ProductivityRate: req.ProductivityRate,
	})
	if err != nil {
		return nil, err
	}
	return s.toProtoProduct(product), nil
}

// UpdateProduct 更新产品
func (s *ProductService) UpdateProduct(ctx context.Context, req *v1.UpdateProductRequest) (*v1.ProductInfo, error) {
	product, err := s.uc.UpdateProduct(ctx, &biz.Product{
		ID:          req.Id,
		Name:        req.Name,
		Price:       req.Price,
		Description: req.Description,
		Status:      req.Status,
	})
	if err != nil {
		return nil, err
	}
	return s.toProtoProduct(product), nil
}

// DeleteProduct 删除产品
func (s *ProductService) DeleteProduct(ctx context.Context, req *v1.DeleteProductRequest) (*v1.DeleteProductResponse, error) {
	err := s.uc.DeleteProduct(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.DeleteProductResponse{Success: true}, nil
}

func (s *ProductService) toProtoProduct(p *biz.Product) *v1.ProductInfo {
	return &v1.ProductInfo{
		Id: p.ID,
		MachineId: func() uint32 {
			if p.MachineID != nil {
				return *p.MachineID
			}
			return 0
		}(),
		Name:             p.Name,
		Price:            p.Price,
		Description:      p.Description,
		Type:             p.Type,
		Cycle:            p.Cycle,
		ProductivityRate: p.ProductivityRate,
		Status:           p.Status,
	}
}
