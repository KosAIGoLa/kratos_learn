package service

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/xuri/excelize/v2"
	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "report/api/report/v1"
	"report/internal/biz"
)

// ReportService 报表服务
type ReportService struct {
	v1.UnimplementedReportServer
	uc  *biz.ReportUsecase
	log *log.Helper
}

// NewReportService 创建报表服务
func NewReportService(uc *biz.ReportUsecase, logger log.Logger) *ReportService {
	return &ReportService{
		uc:  uc,
		log: log.NewHelper(logger),
	}
}

// GetOrderReport 获取订单报表
func (s *ReportService) GetOrderReport(ctx context.Context, req *v1.GetOrderReportRequest) (*v1.OrderReportResponse, error) {
	var startDate, endDate *time.Time
	if req.StartDate != nil {
		t := req.StartDate.AsTime()
		startDate = &t
	}
	if req.EndDate != nil {
		t := req.EndDate.AsTime()
		endDate = &t
	}

	items, total, totalAmount, totalOrders, err := s.uc.GetOrderReport(ctx, startDate, endDate, req.Status, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	var protoItems []*v1.OrderReportItem
	for _, item := range items {
		protoItems = append(protoItems, &v1.OrderReportItem{
			OrderId:     item.OrderID,
			OrderNo:     item.OrderNo,
			UserId:      item.UserID,
			UserName:    item.UserName,
			ProductName: item.ProductName,
			Amount:      item.Amount,
			Status:      item.Status,
			CreatedAt:   timestamppb.New(item.CreatedAt),
		})
	}

	return &v1.OrderReportResponse{
		Items:       protoItems,
		Total:       total,
		TotalAmount: totalAmount,
		TotalOrders: totalOrders,
	}, nil
}

// GetUserReport 获取用户报表
func (s *ReportService) GetUserReport(ctx context.Context, req *v1.GetUserReportRequest) (*v1.UserReportResponse, error) {
	var startDate, endDate *time.Time
	if req.StartDate != nil {
		t := req.StartDate.AsTime()
		startDate = &t
	}
	if req.EndDate != nil {
		t := req.EndDate.AsTime()
		endDate = &t
	}

	items, total, totalUsers, activeUsers, err := s.uc.GetUserReport(ctx, startDate, endDate, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	var protoItems []*v1.UserReportItem
	for _, item := range items {
		var lastLogin *timestamppb.Timestamp
		if item.LastLogin != nil {
			lastLogin = timestamppb.New(*item.LastLogin)
		}
		protoItems = append(protoItems, &v1.UserReportItem{
			UserId:      item.UserID,
			Username:    item.Username,
			Email:       item.Email,
			Phone:       item.Phone,
			OrderCount:  item.OrderCount,
			TotalAmount: item.TotalAmount,
			CreatedAt:   timestamppb.New(item.CreatedAt),
			LastLogin:   lastLogin,
		})
	}

	return &v1.UserReportResponse{
		Items:       protoItems,
		Total:       total,
		TotalUsers:  totalUsers,
		ActiveUsers: activeUsers,
	}, nil
}

// GetSalesReport 获取销售报表
func (s *ReportService) GetSalesReport(ctx context.Context, req *v1.GetSalesReportRequest) (*v1.SalesReportResponse, error) {
	var startDate, endDate *time.Time
	if req.StartDate != nil {
		t := req.StartDate.AsTime()
		startDate = &t
	}
	if req.EndDate != nil {
		t := req.EndDate.AsTime()
		endDate = &t
	}

	groupBy := req.GroupBy
	if groupBy == "" {
		groupBy = "day"
	}

	items, totalSales, totalOrders, err := s.uc.GetSalesReport(ctx, startDate, endDate, groupBy)
	if err != nil {
		return nil, err
	}

	var protoItems []*v1.SalesReportItem
	for _, item := range items {
		protoItems = append(protoItems, &v1.SalesReportItem{
			Period:       item.Period,
			SalesAmount:  item.SalesAmount,
			OrderCount:   item.OrderCount,
			ProductCount: item.ProductCount,
		})
	}

	var avgOrderValue float64
	if totalOrders > 0 {
		avgOrderValue = totalSales / float64(totalOrders)
	}

	return &v1.SalesReportResponse{
		Items:         protoItems,
		TotalSales:    totalSales,
		TotalOrders:   totalOrders,
		AvgOrderValue: avgOrderValue,
	}, nil
}

// GetProductReport 获取商品报表
func (s *ReportService) GetProductReport(ctx context.Context, req *v1.GetProductReportRequest) (*v1.ProductReportResponse, error) {
	var startDate, endDate *time.Time
	if req.StartDate != nil {
		t := req.StartDate.AsTime()
		startDate = &t
	}
	if req.EndDate != nil {
		t := req.EndDate.AsTime()
		endDate = &t
	}

	items, total, err := s.uc.GetProductReport(ctx, startDate, endDate, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	var protoItems []*v1.ProductReportItem
	for _, item := range items {
		protoItems = append(protoItems, &v1.ProductReportItem{
			ProductId:     item.ProductID,
			ProductName:   item.ProductName,
			SalesCount:    item.SalesCount,
			SalesAmount:   item.SalesAmount,
			StockQuantity: item.StockQuantity,
		})
	}

	return &v1.ProductReportResponse{
		Items: protoItems,
		Total: total,
	}, nil
}

// ExportOrderReport 导出订单报表到Excel
func (s *ReportService) ExportOrderReport(ctx context.Context, req *v1.ExportOrderReportRequest) (*v1.ExportReportResponse, error) {
	var startDate, endDate *time.Time
	if req.StartDate != nil {
		t := req.StartDate.AsTime()
		startDate = &t
	}
	if req.EndDate != nil {
		t := req.EndDate.AsTime()
		endDate = &t
	}

	items, totalAmount, totalOrders, err := s.uc.GetAllOrderReport(ctx, startDate, endDate, req.Status)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	defer f.Close()

	sheetName := "订单报表"
	index, _ := f.NewSheet(sheetName)
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	headers := []string{"订单ID", "订单号", "用户ID", "用户名", "商品名称", "金额", "状态", "创建时间"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
	})
	f.SetCellStyle(sheetName, "A1", fmt.Sprintf("%c1", 'A'+len(headers)-1), style)

	for i, item := range items {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), item.OrderID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), item.OrderNo)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), item.UserID)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), item.UserName)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), item.ProductName)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), item.Amount)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), s.getOrderStatusText(item.Status))
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), item.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	summaryRow := len(items) + 3
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryRow), "汇总")
	f.SetCellValue(sheetName, fmt.Sprintf("E%d", summaryRow), "总订单数:")
	f.SetCellValue(sheetName, fmt.Sprintf("F%d", summaryRow), totalOrders)
	f.SetCellValue(sheetName, fmt.Sprintf("G%d", summaryRow), "总金额:")
	f.SetCellValue(sheetName, fmt.Sprintf("H%d", summaryRow), totalAmount)

	fileData, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	fileName := fmt.Sprintf("order_report_%s.xlsx", time.Now().Format("20060102150405"))

	return &v1.ExportReportResponse{
		FileName: fileName,
		FileSize: int64(fileData.Len()),
		FileData: fileData.Bytes(),
	}, nil
}

// ExportUserReport 导出用户报表到Excel
func (s *ReportService) ExportUserReport(ctx context.Context, req *v1.ExportUserReportRequest) (*v1.ExportReportResponse, error) {
	var startDate, endDate *time.Time
	if req.StartDate != nil {
		t := req.StartDate.AsTime()
		startDate = &t
	}
	if req.EndDate != nil {
		t := req.EndDate.AsTime()
		endDate = &t
	}

	items, totalUsers, activeUsers, err := s.uc.GetAllUserReport(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	defer f.Close()

	sheetName := "用户报表"
	index, _ := f.NewSheet(sheetName)
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	headers := []string{"用户ID", "用户名", "邮箱", "手机号", "订单数", "总消费金额", "注册时间", "最后登录"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
	})
	f.SetCellStyle(sheetName, "A1", fmt.Sprintf("%c1", 'A'+len(headers)-1), style)

	for i, item := range items {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), item.UserID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), item.Username)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), item.Email)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), item.Phone)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), item.OrderCount)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), item.TotalAmount)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), item.CreatedAt.Format("2006-01-02 15:04:05"))
		if item.LastLogin != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), item.LastLogin.Format("2006-01-02 15:04:05"))
		} else {
			f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), "从未登录")
		}
	}

	summaryRow := len(items) + 3
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryRow), "汇总")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", summaryRow), "总用户数:")
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", summaryRow), totalUsers)
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", summaryRow), "活跃用户数:")
	f.SetCellValue(sheetName, fmt.Sprintf("E%d", summaryRow), activeUsers)

	fileData, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	fileName := fmt.Sprintf("user_report_%s.xlsx", time.Now().Format("20060102150405"))

	return &v1.ExportReportResponse{
		FileName: fileName,
		FileSize: int64(fileData.Len()),
		FileData: fileData.Bytes(),
	}, nil
}

// ExportSalesReport 导出销售报表到Excel
func (s *ReportService) ExportSalesReport(ctx context.Context, req *v1.ExportSalesReportRequest) (*v1.ExportReportResponse, error) {
	var startDate, endDate *time.Time
	if req.StartDate != nil {
		t := req.StartDate.AsTime()
		startDate = &t
	}
	if req.EndDate != nil {
		t := req.EndDate.AsTime()
		endDate = &t
	}

	groupBy := req.GroupBy
	if groupBy == "" {
		groupBy = "day"
	}

	items, totalSales, totalOrders, err := s.uc.GetSalesReport(ctx, startDate, endDate, groupBy)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	defer f.Close()

	sheetName := "销售报表"
	index, _ := f.NewSheet(sheetName)
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	headers := []string{"时间周期", "销售金额", "订单数", "商品数"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
	})
	f.SetCellStyle(sheetName, "A1", fmt.Sprintf("%c1", 'A'+len(headers)-1), style)

	for i, item := range items {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), item.Period)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), item.SalesAmount)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), item.OrderCount)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), item.ProductCount)
	}

	summaryRow := len(items) + 3
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryRow), "汇总")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", summaryRow), totalSales)
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", summaryRow), totalOrders)
	var avgOrderValue float64
	if totalOrders > 0 {
		avgOrderValue = totalSales / float64(totalOrders)
	}
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", summaryRow), fmt.Sprintf("平均订单金额: %.2f", avgOrderValue))

	fileData, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	fileName := fmt.Sprintf("sales_report_%s.xlsx", time.Now().Format("20060102150405"))

	return &v1.ExportReportResponse{
		FileName: fileName,
		FileSize: int64(fileData.Len()),
		FileData: fileData.Bytes(),
	}, nil
}

// ExportProductReport 导出商品报表到Excel
func (s *ReportService) ExportProductReport(ctx context.Context, req *v1.ExportProductReportRequest) (*v1.ExportReportResponse, error) {
	var startDate, endDate *time.Time
	if req.StartDate != nil {
		t := req.StartDate.AsTime()
		startDate = &t
	}
	if req.EndDate != nil {
		t := req.EndDate.AsTime()
		endDate = &t
	}

	items, err := s.uc.GetAllProductReport(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	defer f.Close()

	sheetName := "商品报表"
	index, _ := f.NewSheet(sheetName)
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	headers := []string{"商品ID", "商品名称", "销售数量", "销售金额", "库存数量"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
	})
	f.SetCellStyle(sheetName, "A1", fmt.Sprintf("%c1", 'A'+len(headers)-1), style)

	var totalSalesCount uint32
	var totalSalesAmount float64
	for i, item := range items {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), item.ProductID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), item.ProductName)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), item.SalesCount)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), item.SalesAmount)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), item.StockQuantity)
		totalSalesCount += item.SalesCount
		totalSalesAmount += item.SalesAmount
	}

	summaryRow := len(items) + 3
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryRow), "汇总")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", summaryRow), "总销售数量:")
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", summaryRow), totalSalesCount)
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", summaryRow), totalSalesAmount)

	fileData, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	fileName := fmt.Sprintf("product_report_%s.xlsx", time.Now().Format("20060102150405"))

	return &v1.ExportReportResponse{
		FileName: fileName,
		FileSize: int64(fileData.Len()),
		FileData: fileData.Bytes(),
	}, nil
}

func (s *ReportService) getOrderStatusText(status int32) string {
	switch status {
	case 0:
		return "待支付"
	case 1:
		return "已支付"
	case 2:
		return "已取消"
	case 3:
		return "已完成"
	default:
		return "未知"
	}
}
