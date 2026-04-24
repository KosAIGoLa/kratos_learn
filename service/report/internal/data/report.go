package data

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"report/internal/biz"
)

type reportRepo struct {
	data *Data
	log  *log.Helper
}

// NewReportRepo 创建报表仓库
func NewReportRepo(data *Data, logger log.Logger) biz.ReportRepo {
	return &reportRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// GetOrderReport 获取订单报表
func (r *reportRepo) GetOrderReport(ctx context.Context, startDate, endDate *time.Time, statusFilter int32, page, pageSize uint32) ([]*biz.OrderReportItem, uint32, float64, uint32, error) {
	var results []struct {
		OrderID     uint64
		OrderNo     string
		UserID      uint32
		UserName    string
		ProductName string
		Amount      float64
		Status      int8
		CreatedAt   time.Time
	}
	var total int64
	var totalAmount float64
	var totalOrders uint32

	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 10
	}

	query := r.data.db.Table("orders").
		Select("orders.id as order_id, orders.order_no, orders.user_id, orders.name as user_name, orders.product_name, orders.amount, orders.status, orders.created_at")

	if startDate != nil {
		query = query.Where("orders.created_at >= ?", startDate)
	}
	if endDate != nil {
		query = query.Where("orders.created_at <= ?", endDate)
	}
	if statusFilter >= 0 {
		query = query.Where("orders.status = ?", statusFilter)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, 0, 0, status.Errorf(codes.Internal, "%s", err.Error())
	}

	offset := (page - 1) * pageSize
	if err := query.Order("orders.created_at DESC").Limit(int(pageSize)).Offset(int(offset)).Find(&results).Error; err != nil {
		return nil, 0, 0, 0, status.Errorf(codes.Internal, "%s", err.Error())
	}

	var sumQuery = r.data.db.Table("orders").Select("COALESCE(SUM(amount), 0) as total_amount, COUNT(*) as total_orders")
	if startDate != nil {
		sumQuery = sumQuery.Where("created_at >= ?", startDate)
	}
	if endDate != nil {
		sumQuery = sumQuery.Where("created_at <= ?", endDate)
	}
	if statusFilter >= 0 {
		sumQuery = sumQuery.Where("status = ?", statusFilter)
	}

	var sumResult struct {
		TotalAmount float64
		TotalOrders uint32
	}
	if err := sumQuery.Scan(&sumResult).Error; err != nil {
		return nil, 0, 0, 0, status.Errorf(codes.Internal, "%s", err.Error())
	}
	totalAmount = sumResult.TotalAmount
	totalOrders = sumResult.TotalOrders

	var items []*biz.OrderReportItem
	for _, r := range results {
		items = append(items, &biz.OrderReportItem{
			OrderID:     r.OrderID,
			OrderNo:     r.OrderNo,
			UserID:      r.UserID,
			UserName:    r.UserName,
			ProductName: r.ProductName,
			Amount:      r.Amount,
			Status:      int32(r.Status),
			CreatedAt:   r.CreatedAt,
		})
	}

	return items, uint32(total), totalAmount, totalOrders, nil
}

// GetAllOrderReport 获取所有订单报表 (用于导出)
func (r *reportRepo) GetAllOrderReport(ctx context.Context, startDate, endDate *time.Time, statusFilter int32) ([]*biz.OrderReportItem, float64, uint32, error) {
	var results []struct {
		OrderID     uint64
		OrderNo     string
		UserID      uint32
		UserName    string
		ProductName string
		Amount      float64
		Status      int8
		CreatedAt   time.Time
	}

	query := r.data.db.Table("orders").
		Select("orders.id as order_id, orders.order_no, orders.user_id, orders.name as user_name, orders.product_name, orders.amount, orders.status, orders.created_at")

	if startDate != nil {
		query = query.Where("orders.created_at >= ?", startDate)
	}
	if endDate != nil {
		query = query.Where("orders.created_at <= ?", endDate)
	}
	if statusFilter >= 0 {
		query = query.Where("orders.status = ?", statusFilter)
	}

	if err := query.Order("orders.created_at DESC").Find(&results).Error; err != nil {
		return nil, 0, 0, status.Errorf(codes.Internal, "%s", err.Error())
	}

	var totalAmount float64
	var items []*biz.OrderReportItem
	for _, r := range results {
		items = append(items, &biz.OrderReportItem{
			OrderID:     r.OrderID,
			OrderNo:     r.OrderNo,
			UserID:      r.UserID,
			UserName:    r.UserName,
			ProductName: r.ProductName,
			Amount:      r.Amount,
			Status:      int32(r.Status),
			CreatedAt:   r.CreatedAt,
		})
		totalAmount += r.Amount
	}

	return items, totalAmount, uint32(len(items)), nil
}

// GetUserReport 获取用户报表
func (r *reportRepo) GetUserReport(ctx context.Context, startDate, endDate *time.Time, page, pageSize uint32) ([]*biz.UserReportItem, uint32, uint32, uint32, error) {
	var results []struct {
		UserID      uint32
		Username    string
		Email       string
		Phone       string
		OrderCount  uint32
		TotalAmount float64
		CreatedAt   time.Time
		LastLogin   *time.Time
	}
	var total int64

	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 10
	}

	query := r.data.db.Table("users").
		Select("users.id as user_id, users.username, users.email, users.phone, COALESCE(COUNT(orders.id), 0) as order_count, COALESCE(SUM(orders.amount), 0) as total_amount, users.created_at, users.last_login_at as last_login").
		Joins("LEFT JOIN orders ON users.id = orders.user_id").
		Group("users.id")

	if startDate != nil {
		query = query.Where("users.created_at >= ?", startDate)
	}
	if endDate != nil {
		query = query.Where("users.created_at <= ?", endDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, 0, 0, status.Errorf(codes.Internal, "%s", err.Error())
	}

	offset := (page - 1) * pageSize
	if err := query.Order("users.created_at DESC").Limit(int(pageSize)).Offset(int(offset)).Find(&results).Error; err != nil {
		return nil, 0, 0, 0, status.Errorf(codes.Internal, "%s", err.Error())
	}

	var totalUsers uint32
	var activeUsers uint32
	if err := r.data.db.Table("users").Count(&total).Error; err == nil {
		totalUsers = uint32(total)
	}
	if err := r.data.db.Table("users").Where("last_login_at IS NOT NULL AND last_login_at >= ?", time.Now().AddDate(0, 0, -30)).Count(&total).Error; err == nil {
		activeUsers = uint32(total)
	}

	var items []*biz.UserReportItem
	for _, r := range results {
		items = append(items, &biz.UserReportItem{
			UserID:      r.UserID,
			Username:    r.Username,
			Email:       r.Email,
			Phone:       r.Phone,
			OrderCount:  r.OrderCount,
			TotalAmount: r.TotalAmount,
			CreatedAt:   r.CreatedAt,
			LastLogin:   r.LastLogin,
		})
	}

	return items, uint32(total), totalUsers, activeUsers, nil
}

// GetAllUserReport 获取所有用户报表 (用于导出)
func (r *reportRepo) GetAllUserReport(ctx context.Context, startDate, endDate *time.Time) ([]*biz.UserReportItem, uint32, uint32, error) {
	var results []struct {
		UserID      uint32
		Username    string
		Email       string
		Phone       string
		OrderCount  uint32
		TotalAmount float64
		CreatedAt   time.Time
		LastLogin   *time.Time
	}
	var total int64

	query := r.data.db.Table("users").
		Select("users.id as user_id, users.username, users.email, users.phone, COALESCE(COUNT(orders.id), 0) as order_count, COALESCE(SUM(orders.amount), 0) as total_amount, users.created_at, users.last_login_at as last_login").
		Joins("LEFT JOIN orders ON users.id = orders.user_id").
		Group("users.id")

	if startDate != nil {
		query = query.Where("users.created_at >= ?", startDate)
	}
	if endDate != nil {
		query = query.Where("users.created_at <= ?", endDate)
	}

	if err := query.Order("users.created_at DESC").Find(&results).Error; err != nil {
		return nil, 0, 0, status.Errorf(codes.Internal, "%s", err.Error())
	}

	var totalUsers uint32
	var activeUsers uint32
	if err := r.data.db.Table("users").Count(&total).Error; err == nil {
		totalUsers = uint32(total)
	}
	if err := r.data.db.Table("users").Where("last_login_at IS NOT NULL AND last_login_at >= ?", time.Now().AddDate(0, 0, -30)).Count(&total).Error; err == nil {
		activeUsers = uint32(total)
	}

	var items []*biz.UserReportItem
	for _, r := range results {
		items = append(items, &biz.UserReportItem{
			UserID:      r.UserID,
			Username:    r.Username,
			Email:       r.Email,
			Phone:       r.Phone,
			OrderCount:  r.OrderCount,
			TotalAmount: r.TotalAmount,
			CreatedAt:   r.CreatedAt,
			LastLogin:   r.LastLogin,
		})
	}

	return items, totalUsers, activeUsers, nil
}

// GetSalesReport 获取销售报表
func (r *reportRepo) GetSalesReport(ctx context.Context, startDate, endDate *time.Time, groupBy string) ([]*biz.SalesReportItem, float64, uint32, error) {
	var results []struct {
		Period       string
		SalesAmount  float64
		OrderCount   uint32
		ProductCount uint32
	}

	var dateFormat string
	switch groupBy {
	case "day":
		dateFormat = "%Y-%m-%d"
	case "week":
		dateFormat = "%Y-%u"
	case "month":
		dateFormat = "%Y-%m"
	default:
		dateFormat = "%Y-%m-%d"
	}

	query := r.data.db.Table("orders").
		Select(fmt.Sprintf("DATE_FORMAT(created_at, '%s') as period, SUM(amount) as sales_amount, COUNT(*) as order_count, COUNT(DISTINCT product_id) as product_count", dateFormat)).
		Group("period").
		Order("period DESC")

	if startDate != nil {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", endDate)
	}

	if err := query.Find(&results).Error; err != nil {
		return nil, 0, 0, status.Errorf(codes.Internal, "%s", err.Error())
	}

	var totalSales float64
	var totalOrders uint32
	var items []*biz.SalesReportItem
	for _, r := range results {
		items = append(items, &biz.SalesReportItem{
			Period:       r.Period,
			SalesAmount:  r.SalesAmount,
			OrderCount:   r.OrderCount,
			ProductCount: r.ProductCount,
		})
		totalSales += r.SalesAmount
		totalOrders += r.OrderCount
	}

	return items, totalSales, totalOrders, nil
}

// GetProductReport 获取商品报表
func (r *reportRepo) GetProductReport(ctx context.Context, startDate, endDate *time.Time, page, pageSize uint32) ([]*biz.ProductReportItem, uint32, error) {
	var results []struct {
		ProductID     uint32
		ProductName   string
		SalesCount    uint32
		SalesAmount   float64
		StockQuantity uint32
	}
	var total int64

	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 10
	}

	query := r.data.db.Table("products").
		Select("products.id as product_id, products.name as product_name, COALESCE(SUM(orders.quantity), 0) as sales_count, COALESCE(SUM(orders.amount), 0) as sales_amount, products.stock as stock_quantity").
		Joins("LEFT JOIN orders ON products.id = orders.product_id")

	if startDate != nil || endDate != nil {
		if startDate != nil {
			query = query.Where("orders.created_at >= ?", startDate)
		}
		if endDate != nil {
			query = query.Where("orders.created_at <= ?", endDate)
		}
	}

	query = query.Group("products.id")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, status.Errorf(codes.Internal, "%s", err.Error())
	}

	offset := (page - 1) * pageSize
	if err := query.Order("sales_amount DESC").Limit(int(pageSize)).Offset(int(offset)).Find(&results).Error; err != nil {
		return nil, 0, status.Errorf(codes.Internal, "%s", err.Error())
	}

	var items []*biz.ProductReportItem
	for _, r := range results {
		items = append(items, &biz.ProductReportItem{
			ProductID:     r.ProductID,
			ProductName:   r.ProductName,
			SalesCount:    r.SalesCount,
			SalesAmount:   r.SalesAmount,
			StockQuantity: r.StockQuantity,
		})
	}

	return items, uint32(total), nil
}

// GetAllProductReport 获取所有商品报表 (用于导出)
func (r *reportRepo) GetAllProductReport(ctx context.Context, startDate, endDate *time.Time) ([]*biz.ProductReportItem, error) {
	var results []struct {
		ProductID     uint32
		ProductName   string
		SalesCount    uint32
		SalesAmount   float64
		StockQuantity uint32
	}

	query := r.data.db.Table("products").
		Select("products.id as product_id, products.name as product_name, COALESCE(SUM(orders.quantity), 0) as sales_count, COALESCE(SUM(orders.amount), 0) as sales_amount, products.stock as stock_quantity").
		Joins("LEFT JOIN orders ON products.id = orders.product_id")

	if startDate != nil || endDate != nil {
		if startDate != nil {
			query = query.Where("orders.created_at >= ?", startDate)
		}
		if endDate != nil {
			query = query.Where("orders.created_at <= ?", endDate)
		}
	}

	query = query.Group("products.id")

	if err := query.Order("sales_amount DESC").Find(&results).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}

	var items []*biz.ProductReportItem
	for _, r := range results {
		items = append(items, &biz.ProductReportItem{
			ProductID:     r.ProductID,
			ProductName:   r.ProductName,
			SalesCount:    r.SalesCount,
			SalesAmount:   r.SalesAmount,
			StockQuantity: r.StockQuantity,
		})
	}

	return items, nil
}
