package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	v1 "order/api/order/v1"
	"order/internal/biz"
)

// OrderService 订单服务
type OrderService struct {
	v1.UnimplementedOrderServer
	uc  *biz.OrderUsecase
	log *log.Helper
}

// NewOrderService 创建订单服务
func NewOrderService(uc *biz.OrderUsecase, logger log.Logger) *OrderService {
	return &OrderService{
		uc:  uc,
		log: log.NewHelper(logger),
	}
}

// CreateOrder 创建订单
func (s *OrderService) CreateOrder(ctx context.Context, req *v1.CreateOrderRequest) (*v1.OrderInfo, error) {
	order, err := s.uc.CreateOrder(ctx, &biz.Order{
		UserID:     req.UserId,
		ProductID:  req.ProductId,
		Quantity:   req.Quantity,
		InviteCode: req.InviteCode,
	})
	if err != nil {
		return nil, err
	}
	return s.toProtoOrder(order), nil
}

// GetOrder 获取订单
func (s *OrderService) GetOrder(ctx context.Context, req *v1.GetOrderRequest) (*v1.OrderInfo, error) {
	var order *biz.Order
	var err error

	switch req.Identifier.(type) {
	case *v1.GetOrderRequest_Id:
		order, err = s.uc.GetOrder(ctx, req.GetId())
	case *v1.GetOrderRequest_OrderNo:
		order, err = s.uc.GetOrderByOrderNo(ctx, req.GetOrderNo())
	}

	if err != nil {
		return nil, err
	}
	return s.toProtoOrder(order), nil
}

// ListOrders 获取订单列表
func (s *OrderService) ListOrders(ctx context.Context, req *v1.ListOrdersRequest) (*v1.ListOrdersResponse, error) {
	orders, total, err := s.uc.ListOrders(ctx, req.UserId, req.Status, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	var protoOrders []*v1.OrderInfo
	for _, o := range orders {
		protoOrders = append(protoOrders, s.toProtoOrder(o))
	}

	return &v1.ListOrdersResponse{
		Orders: protoOrders,
		Total:  total,
	}, nil
}

// CancelOrder 取消订单
func (s *OrderService) CancelOrder(ctx context.Context, req *v1.CancelOrderRequest) (*v1.OrderInfo, error) {
	if err := s.uc.CancelOrder(ctx, req.Id); err != nil {
		return nil, err
	}
	return s.GetOrder(ctx, &v1.GetOrderRequest{Identifier: &v1.GetOrderRequest_Id{Id: req.Id}})
}

// PayOrder 支付订单
func (s *OrderService) PayOrder(ctx context.Context, req *v1.PayOrderRequest) (*v1.PayOrderResponse, error) {
	// TODO: 调用支付服务
	return &v1.PayOrderResponse{
		OrderNo: req.OrderNo,
		PayUrl:  "https://example.com/pay",
		Status:  1,
	}, nil
}

func (s *OrderService) toProtoOrder(o *biz.Order) *v1.OrderInfo {
	return &v1.OrderInfo{
		Id:          o.ID,
		OrderNo:     o.OrderNo,
		UserId:      o.UserID,
		InviteCode:  o.InviteCode,
		Phone:       o.Phone,
		Name:        o.Name,
		ProductId:   o.ProductID,
		ProductName: o.ProductName,
		Amount:      o.Amount,
		Quantity:    o.Quantity,
		Status:      o.Status,
	}
}
