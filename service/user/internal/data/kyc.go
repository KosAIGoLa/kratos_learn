package data

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"user/internal/biz"
)

// KYCVerification 实名认证数据模型
type KYCVerification struct {
	ID          uint32 `gorm:"primarykey"`
	UserID      uint32 `gorm:"uniqueIndex:uk_user_id;not null"`
	Name        string `gorm:"type:varchar(50);not null"`
	IDCard      string `gorm:"type:varchar(18);not null"`
	IDCardFront string `gorm:"type:varchar(255);not null"`
	IDCardBack  string `gorm:"type:varchar(255);not null"`
	Status      int8   `gorm:"index:idx_status;default:0"`
	Remark      string `gorm:"type:varchar(255)"`
	VerifiedAt  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type kycRepo struct {
	data *Data
	log  *log.Helper
}

// NewKYCRepo 创建实名认证仓库
func NewKYCRepo(data *Data, logger log.Logger) biz.KYCRepo {
	return &kycRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *kycRepo) CreateKYC(ctx context.Context, k *biz.KYCVerification) (*biz.KYCVerification, error) {
	kyc := KYCVerification{
		UserID:      k.UserID,
		Name:        k.Name,
		IDCard:      k.IDCard,
		IDCardFront: k.IDCardFront,
		IDCardBack:  k.IDCardBack,
		Status:      0,
	}
	if err := r.data.db.Create(&kyc).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}
	return r.toBizKYC(&kyc), nil
}

func (r *kycRepo) GetKYCByUserID(ctx context.Context, userID uint32) (*biz.KYCVerification, error) {
	var kyc KYCVerification
	if err := r.data.db.Where("user_id = ?", userID).First(&kyc).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "实名认证记录不存在")
		}
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}
	return r.toBizKYC(&kyc), nil
}

func (r *kycRepo) UpdateKYCStatus(ctx context.Context, id uint32, statusValue int32, remark string) error {
	now := time.Now()
	return r.data.db.Model(&KYCVerification{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":      statusValue,
		"remark":      remark,
		"verified_at": &now,
	}).Error
}

func (r *kycRepo) toBizKYC(k *KYCVerification) *biz.KYCVerification {
	return &biz.KYCVerification{
		ID:          k.ID,
		UserID:      k.UserID,
		Name:        k.Name,
		IDCard:      k.IDCard,
		IDCardFront: k.IDCardFront,
		IDCardBack:  k.IDCardBack,
		Status:      int32(k.Status),
		Remark:      k.Remark,
		VerifiedAt:  k.VerifiedAt,
		CreatedAt:   k.CreatedAt,
	}
}
