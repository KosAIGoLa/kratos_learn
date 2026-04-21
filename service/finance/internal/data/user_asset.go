package data

import (
	"context"
	"finance/internal/biz"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

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
	if r.data.UserClient() == nil {
		return nil, status.Error(codes.FailedPrecondition, "user client not initialized")
	}

	user, err := r.data.UserClient().GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %d", userID)
	}

	return &biz.UserAsset{
		UserID:     user.Id,
		Balance:    user.Balance,
		WorkPoints: user.WorkPoints,
	}, nil
}

func (r *userAssetRepo) ConvertHashrate(ctx context.Context, userID uint32, amount float64) (*biz.HashrateConversion, error) {
	if amount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "conversion amount must be greater than 0")
	}
	if r.data.UserClient() == nil {
		return nil, status.Error(codes.FailedPrecondition, "user client not initialized")
	}

	userAsset, err := r.GetUserAsset(ctx, userID)
	if err != nil {
		return nil, err
	}

	var result *biz.HashrateConversion
	requestID := fmt.Sprintf("hashrate-convert-%d-%d", userID, time.Now().UnixNano())
	reason := buildConvertedHashrateRemark(amount)

	err = r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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

		updatedUser, err := r.data.UserClient().AdjustUserAsset(ctx, userID, amount, 0, reason, requestID)
		if err != nil {
			return err
		}
		if updatedUser == nil {
			return status.Error(codes.Internal, "user asset adjustment returned nil response")
		}

		balanceLog := &BalanceLog{
			UserID:        userID,
			Type:          int8(biz.BalanceLogTypeHashrateConversion),
			Amount:        amount,
			BeforeBalance: userAsset.Balance,
			AfterBalance:  updatedUser.Balance,
			Remark:        reason,
		}
		if err := tx.Create(balanceLog).Error; err != nil {
			return status.Errorf(codes.Internal, "failed to create balance log: %v", err)
		}

		result = &biz.HashrateConversion{
			UserID:         userID,
			Amount:         amount,
			BeforeHashrate: available,
			AfterHashrate:  available - amount,
			BeforeBalance:  userAsset.Balance,
			AfterBalance:   updatedUser.Balance,
			Remark:         reason,
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
