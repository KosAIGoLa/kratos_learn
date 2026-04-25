package data

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"finance/internal/biz"
)

// Recharge 充值数据模型
type Recharge struct {
	ID         uint64  `gorm:"primarykey"`
	UserID     uint32  `gorm:"index:idx_user_id;not null"`
	InviteCode string  `gorm:"type:varchar(20);not null"`
	Phone      string  `gorm:"index:idx_phone;type:varchar(20);not null"`
	Name       string  `gorm:"type:varchar(50);not null"`
	OrderNo    string  `gorm:"uniqueIndex:uk_order_no;type:varchar(50);not null"`
	Amount     float64 `gorm:"type:decimal(10,2);not null"`
	Status     int8    `gorm:"index:idx_status;default:0"`
	CreatedAt  time.Time
}

type rechargeRepo struct {
	data *Data
	log  *log.Helper
}

// NewRechargeRepo 创建充值仓库
func NewRechargeRepo(data *Data, logger log.Logger) biz.RechargeRepo {
	return &rechargeRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *rechargeRepo) CreateRecharge(ctx context.Context, re *biz.Recharge) (*biz.Recharge, error) {
	orderNo := re.OrderNo
	if orderNo == "" {
		orderNo = generateRechargeOrderNo()
	}
	recharge := Recharge{
		UserID:     re.UserID,
		InviteCode: re.InviteCode,
		Phone:      re.Phone,
		Name:       re.Name,
		OrderNo:    orderNo,
		Amount:     re.Amount,
		Status:     0, // 待支付
		CreatedAt:  time.Now(),
	}
	if err := r.data.db.Create(&recharge).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create recharge: %v", err)
	}
	return r.toBizRecharge(&recharge), nil
}

func (r *rechargeRepo) GetRechargeByOrderNo(ctx context.Context, orderNo string) (*biz.Recharge, error) {
	var recharge Recharge
	if err := r.data.db.Where("order_no = ?", orderNo).First(&recharge).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "充值记录不存在")
		}
		return nil, status.Errorf(codes.Internal, "failed to get recharge: %v", err)
	}
	return r.toBizRecharge(&recharge), nil
}

func (r *rechargeRepo) ListRecharges(ctx context.Context, userID uint32, statusFilter int32, page, pageSize uint32) ([]*biz.Recharge, uint32, error) {
	if page == 0 || pageSize == 0 {
		return nil, 0, status.Errorf(codes.InvalidArgument, "page and pageSize must be greater than 0")
	}

	var recharges []Recharge
	var total int64

	query := r.data.db.Model(&Recharge{})
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if statusFilter >= 0 {
		query = query.Where("status = ?", statusFilter)
	}

	query.Count(&total)
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Limit(int(pageSize)).Offset(int(offset)).Find(&recharges).Error; err != nil {
		return nil, 0, status.Errorf(codes.Internal, "failed to query recharges: %v", err)
	}

	var bizRecharges []*biz.Recharge
	for _, re := range recharges {
		bizRecharges = append(bizRecharges, r.toBizRecharge(&re))
	}
	return bizRecharges, uint32(total), nil
}

func (r *rechargeRepo) toBizRecharge(re *Recharge) *biz.Recharge {
	return &biz.Recharge{
		ID:         re.ID,
		UserID:     re.UserID,
		InviteCode: re.InviteCode,
		Phone:      re.Phone,
		Name:       re.Name,
		OrderNo:    re.OrderNo,
		Amount:     re.Amount,
		Status:     int32(re.Status),
		CreatedAt:  re.CreatedAt,
	}
}

func generateRechargeOrderNo() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("REC%d%s", time.Now().UnixNano(), hex.EncodeToString(b))
}
