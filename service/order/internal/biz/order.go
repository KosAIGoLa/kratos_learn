package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// Order 订单领域模型
type Order struct {
	ID          uint64
	OrderNo     string
	UserID      uint32
	InviteCode  string
	Phone       string
	Name        string
	ProductID   uint32
	ProductName string
	Amount      float64
	Quantity    uint32
	Status      int32
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// OrderRepo 订单存储接口
type OrderRepo interface {
	CreateOrder(ctx context.Context, o *Order) (*Order, error)
	GetOrderByID(ctx context.Context, id uint64) (*Order, error)
	GetOrderByOrderNo(ctx context.Context, orderNo string) (*Order, error)
	ListOrders(ctx context.Context, userID uint32, status int32, page, pageSize uint32) ([]*Order, uint32, error)
	UpdateOrderStatus(ctx context.Context, id uint64, status int32) error
	CancelOrder(ctx context.Context, id uint64) error
}

// OrderUsecase 订单用例
type OrderUsecase struct {
	repo OrderRepo
	log  *log.Helper
}

// NewOrderUsecase 创建订单用例
func NewOrderUsecase(repo OrderRepo, logger log.Logger) *OrderUsecase {
	return &OrderUsecase{repo: repo, log: log.NewHelper(logger)}
}

// CreateOrder 创建订单
func (uc *OrderUsecase) CreateOrder(ctx context.Context, o *Order) (*Order, error) {
	return uc.repo.CreateOrder(ctx, o)
}

// GetOrder 获取订单
func (uc *OrderUsecase) GetOrder(ctx context.Context, id uint64) (*Order, error) {
	return uc.repo.GetOrderByID(ctx, id)
}

// GetOrderByOrderNo 根据订单号获取订单
func (uc *OrderUsecase) GetOrderByOrderNo(ctx context.Context, orderNo string) (*Order, error) {
	return uc.repo.GetOrderByOrderNo(ctx, orderNo)
}

// ListOrders 获取订单列表
func (uc *OrderUsecase) ListOrders(ctx context.Context, userID uint32, status int32, page, pageSize uint32) ([]*Order, uint32, error) {
	return uc.repo.ListOrders(ctx, userID, status, page, pageSize)
}

// CancelOrder 取消订单
func (uc *OrderUsecase) CancelOrder(ctx context.Context, id uint64) error {
	return uc.repo.CancelOrder(ctx, id)
}

// PayOrder 支付订单
func (uc *OrderUsecase) PayOrder(ctx context.Context, orderNo string) (*Order, error) {
	order, err := uc.repo.GetOrderByOrderNo(ctx, orderNo)
	if err != nil {
		return nil, err
	}
	// TODO: 调用支付服务
	return order, nil
}
