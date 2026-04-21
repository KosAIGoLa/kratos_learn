package data

import (
	"context"
	"finance/internal/biz"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type User struct {
	ID         uint32  `gorm:"primarykey"`
	Balance    float64 `gorm:"type:decimal(15,2)"`
	WorkPoints float64 `gorm:"column:work_points;type:decimal(15,2)"`
}

func (User) TableName() string {
	return "users"
}

type UserHashrate struct {
	ID            uint32  `gorm:"primarykey"`
	UserID        uint32  `gorm:"index:idx_user_id;not null"`
	TotalHashrate float64 `gorm:"column:total_hashrate;type:decimal(15,2);not null"`
	Status        int8    `gorm:"index:idx_status;default:1"`
}

func (UserHashrate) TableName() string {
	return "user_hashrates"
}

type userAssetRepo struct {
	data *Data
	log  *log.Helper
}

// NewUserAssetRepo 创建用户资产仓库
func NewUserAssetRepo(data *Data, logger log.Logger) biz.UserAssetRepo {
	return &userAssetRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *userAssetRepo) GetUserAsset(ctx context.Context, userID uint32) (*biz.UserAsset, error) {
	var user User
	if err := r.data.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "user not found: %d", userID)
		}
		return nil, status.Errorf(codes.Internal, "failed to query user asset: %v", err)
	}

	return &biz.UserAsset{
		UserID:     user.ID,
		Balance:    user.Balance,
		WorkPoints: user.WorkPoints,
	}, nil
}

func (r *userAssetRepo) ConvertHashrate(ctx context.Context, userID uint32, amount float64) (*biz.HashrateConversion, error) {
	if amount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "conversion amount must be greater than 0")
	}

	var result *biz.HashrateConversion
	err := r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return status.Errorf(codes.NotFound, "user not found: %d", userID)
			}
			return status.Errorf(codes.Internal, "failed to lock user asset: %v", err)
		}

		var rows []UserHashrate
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND status = ? AND total_hashrate > 0", userID, 1).
			Order("id ASC").
			Find(&rows).Error; err != nil {
			return status.Errorf(codes.Internal, "failed to query user hashrate: %v", err)
		}

		available := 0.0
		for _, row := range rows {
			available += row.TotalHashrate
		}

		if available < amount {
			return status.Errorf(codes.FailedPrecondition, "insufficient hashrate: available %.2f, requested %.2f", available, amount)
		}

		convertedTotal, err := r.getConvertedHashrateTotal(ctx, tx, userID)
		if err != nil {
			return err
		}

		generatedTotal := available + convertedTotal
		convertibleLimit := generatedTotal - convertedTotal
		if amount > convertibleLimit {
			return status.Errorf(
				codes.FailedPrecondition,
				"conversion amount exceeds generated hashrate: generated %.2f, converted %.2f, remaining %.2f, requested %.2f",
				generatedTotal,
				convertedTotal,
				convertibleLimit,
				amount,
			)
		}

		remaining := amount
		for _, row := range rows {
			if remaining <= 0 {
				break
			}

			deduct := row.TotalHashrate
			if deduct > remaining {
				deduct = remaining
			}

			nextHashrate := row.TotalHashrate - deduct
			if err := tx.Model(&UserHashrate{}).
				Where("id = ?", row.ID).
				Update("total_hashrate", nextHashrate).Error; err != nil {
				return status.Errorf(codes.Internal, "failed to update user hashrate: %v", err)
			}

			remaining -= deduct
		}

		beforeBalance := user.Balance
		afterBalance := beforeBalance + amount
		if err := tx.Model(&User{}).
			Where("id = ?", userID).
			Update("balance", afterBalance).Error; err != nil {
			return status.Errorf(codes.Internal, "failed to update user balance: %v", err)
		}

		remark := buildConvertedHashrateRemark(amount)
		balanceLog := &BalanceLog{
			UserID:        userID,
			Type:          int8(biz.BalanceLogTypeHashrateConversion),
			Amount:        amount,
			BeforeBalance: beforeBalance,
			AfterBalance:  afterBalance,
			Remark:        remark,
		}
		if err := tx.Create(balanceLog).Error; err != nil {
			return status.Errorf(codes.Internal, "failed to create balance log: %v", err)
		}

		result = &biz.HashrateConversion{
			UserID:         userID,
			Amount:         amount,
			BeforeHashrate: available,
			AfterHashrate:  available - amount,
			BeforeBalance:  beforeBalance,
			AfterBalance:   afterBalance,
			Remark:         remark,
			CreatedAt:      balanceLog.CreatedAt,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *userAssetRepo) getConvertedHashrateTotal(ctx context.Context, tx *gorm.DB, userID uint32) (float64, error) {
	var total float64
	err := tx.WithContext(ctx).
		Model(&BalanceLog{}).
		Where("user_id = ? AND type = ?", userID, biz.BalanceLogTypeHashrateConversion).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	if err != nil {
		return 0, status.Errorf(codes.Internal, "failed to query converted hashrate total: %v", err)
	}
	return total, nil
}

const convertedHashrateRemarkPrefix = "手动算力转换"

func buildConvertedHashrateRemark(amount float64) string {
	return fmt.Sprintf("%s %.2f", convertedHashrateRemarkPrefix, amount)
}
