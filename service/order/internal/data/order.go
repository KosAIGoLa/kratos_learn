package data

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"order/internal/biz"
)

// Order 订单数据模型
type Order struct {
	ID          uint64  `gorm:"primarykey"`
	OrderNo     string  `gorm:"uniqueIndex:uk_order_no;type:varchar(50);not null"`
	UserID      uint32  `gorm:"index:idx_user_id;not null"`
	InviteCode  string  `gorm:"type:varchar(20);not null"`
	Phone       string  `gorm:"index:idx_phone;type:varchar(20);not null"`
	Name        string  `gorm:"type:varchar(50);not null"`
	ProductID   uint32  `gorm:"not null"`
	ProductName string  `gorm:"type:varchar(255);not null"`
	Amount      float64 `gorm:"type:decimal(10,2);not null"`
	Quantity    uint32  `gorm:"not null"`
	Status      int8    `gorm:"index:idx_status;default:0"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type orderRepo struct {
	data *Data
	log  *log.Helper
}

// NewOrderRepo 创建订单仓库
func NewOrderRepo(data *Data, logger log.Logger) biz.OrderRepo {
	return &orderRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *orderRepo) CreateOrder(ctx context.Context, o *biz.Order) (*biz.Order, error) {
	order := Order{
		OrderNo:     generateOrderNo(),
		UserID:      o.UserID,
		InviteCode:  o.InviteCode,
		Phone:       o.Phone,
		Name:        o.Name,
		ProductID:   o.ProductID,
		ProductName: o.ProductName,
		Amount:      o.Amount,
		Quantity:    o.Quantity,
		Status:      0, // 待支付
	}
	if err := r.data.db.Create(&order).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}
	return r.toBizOrder(&order), nil
}

func (r *orderRepo) GetOrderByID(ctx context.Context, id uint64) (*biz.Order, error) {
	var order Order
	if err := r.data.db.First(&order, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "订单不存在")
		}
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}
	return r.toBizOrder(&order), nil
}

func (r *orderRepo) GetOrderByOrderNo(ctx context.Context, orderNo string) (*biz.Order, error) {
	var order Order
	if err := r.data.db.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		return nil, err
	}
	return r.toBizOrder(&order), nil
}

func (r *orderRepo) ListOrders(ctx context.Context, userID uint32, statusFilter int32, page, pageSize uint32) ([]*biz.Order, uint32, error) {
	var orders []Order
	var total int64

	query := r.data.db.Model(&Order{})
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if statusFilter >= 0 {
		query = query.Where("status = ?", statusFilter)
	}

	query.Count(&total)
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Limit(int(pageSize)).Offset(int(offset)).Find(&orders).Error; err != nil {
		return nil, 0, status.Errorf(codes.Internal, "%s", err.Error())
	}

	var bizOrders []*biz.Order
	for _, o := range orders {
		bizOrders = append(bizOrders, r.toBizOrder(&o))
	}
	return bizOrders, uint32(total), nil
}

func (r *orderRepo) UpdateOrderStatus(ctx context.Context, id uint64, statusValue int32) error {
	return r.data.db.Model(&Order{}).Where("id = ?", id).Update("status", statusValue).Error
}

func (r *orderRepo) CancelOrder(ctx context.Context, id uint64) error {
	return r.data.db.Model(&Order{}).Where("id = ?", id).Update("status", 2).Error
}

func (r *orderRepo) toBizOrder(o *Order) *biz.Order {
	return &biz.Order{
		ID:          o.ID,
		OrderNo:     o.OrderNo,
		UserID:      o.UserID,
		InviteCode:  o.InviteCode,
		Phone:       o.Phone,
		Name:        o.Name,
		ProductID:   o.ProductID,
		ProductName: o.ProductName,
		Amount:      o.Amount,
		Quantity:    o.Quantity,
		Status:      int32(o.Status),
		CreatedAt:   o.CreatedAt,
		UpdatedAt:   o.UpdatedAt,
	}
}

func generateOrderNo() string {
	return fmt.Sprintf("ORD%d", time.Now().UnixNano())
}
