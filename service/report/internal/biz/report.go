package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// OrderReportItem 订单报表项
type OrderReportItem struct {
	OrderID     uint64
	OrderNo     string
	UserID      uint32
	UserName    string
	ProductName string
	Amount      float64
	Status      int32
	CreatedAt   time.Time
}

// UserReportItem 用户报表项
type UserReportItem struct {
	UserID      uint32
	Username    string
	Email       string
	Phone       string
	OrderCount  uint32
	TotalAmount float64
	CreatedAt   time.Time
	LastLogin   *time.Time
}

// SalesReportItem 销售报表项
type SalesReportItem struct {
	Period       string
	SalesAmount  float64
	OrderCount   uint32
	ProductCount uint32
}

// ProductReportItem 商品报表项
type ProductReportItem struct {
	ProductID     uint32
	ProductName   string
	SalesCount    uint32
	SalesAmount   float64
	StockQuantity uint32
}

// ReportRepo 报表仓储接口
type ReportRepo interface {
	// 订单报表
	GetOrderReport(ctx context.Context, startDate, endDate *time.Time, status int32, page, pageSize uint32) ([]*OrderReportItem, uint32, float64, uint32, error)
	GetAllOrderReport(ctx context.Context, startDate, endDate *time.Time, status int32) ([]*OrderReportItem, float64, uint32, error)

	// 用户报表
	GetUserReport(ctx context.Context, startDate, endDate *time.Time, page, pageSize uint32) ([]*UserReportItem, uint32, uint32, uint32, error)
	GetAllUserReport(ctx context.Context, startDate, endDate *time.Time) ([]*UserReportItem, uint32, uint32, error)

	// 销售报表
	GetSalesReport(ctx context.Context, startDate, endDate *time.Time, groupBy string) ([]*SalesReportItem, float64, uint32, error)

	// 商品报表
	GetProductReport(ctx context.Context, startDate, endDate *time.Time, page, pageSize uint32) ([]*ProductReportItem, uint32, error)
	GetAllProductReport(ctx context.Context, startDate, endDate *time.Time) ([]*ProductReportItem, error)
}

// ReportUsecase 报表用例
type ReportUsecase struct {
	repo ReportRepo
	log  *log.Helper
}

// NewReportUsecase 创建报表用例
func NewReportUsecase(repo ReportRepo, logger log.Logger) *ReportUsecase {
	return &ReportUsecase{repo: repo, log: log.NewHelper(logger)}
}

// GetOrderReport 获取订单报表
func (uc *ReportUsecase) GetOrderReport(ctx context.Context, startDate, endDate *time.Time, status int32, page, pageSize uint32) ([]*OrderReportItem, uint32, float64, uint32, error) {
	return uc.repo.GetOrderReport(ctx, startDate, endDate, status, page, pageSize)
}

// GetAllOrderReport 获取所有订单报表 (用于导出)
func (uc *ReportUsecase) GetAllOrderReport(ctx context.Context, startDate, endDate *time.Time, status int32) ([]*OrderReportItem, float64, uint32, error) {
	return uc.repo.GetAllOrderReport(ctx, startDate, endDate, status)
}

// GetUserReport 获取用户报表
func (uc *ReportUsecase) GetUserReport(ctx context.Context, startDate, endDate *time.Time, page, pageSize uint32) ([]*UserReportItem, uint32, uint32, uint32, error) {
	return uc.repo.GetUserReport(ctx, startDate, endDate, page, pageSize)
}

// GetAllUserReport 获取所有用户报表 (用于导出)
func (uc *ReportUsecase) GetAllUserReport(ctx context.Context, startDate, endDate *time.Time) ([]*UserReportItem, uint32, uint32, error) {
	return uc.repo.GetAllUserReport(ctx, startDate, endDate)
}

// GetSalesReport 获取销售报表
func (uc *ReportUsecase) GetSalesReport(ctx context.Context, startDate, endDate *time.Time, groupBy string) ([]*SalesReportItem, float64, uint32, error) {
	return uc.repo.GetSalesReport(ctx, startDate, endDate, groupBy)
}

// GetProductReport 获取商品报表
func (uc *ReportUsecase) GetProductReport(ctx context.Context, startDate, endDate *time.Time, page, pageSize uint32) ([]*ProductReportItem, uint32, error) {
	return uc.repo.GetProductReport(ctx, startDate, endDate, page, pageSize)
}

// GetAllProductReport 获取所有商品报表 (用于导出)
func (uc *ReportUsecase) GetAllProductReport(ctx context.Context, startDate, endDate *time.Time) ([]*ProductReportItem, error) {
	return uc.repo.GetAllProductReport(ctx, startDate, endDate)
}
