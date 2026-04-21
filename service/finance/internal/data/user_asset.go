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
	data             *Data
	compensationRepo biz.HashrateCompensationRepo
	log              *log.Helper
}

// NewUserAssetRepo 创建用户资产仓库
func NewUserAssetRepo(data *Data, compensationRepo biz.HashrateCompensationRepo, logger log.Logger) biz.UserAssetRepo {
	return &userAssetRepo{
		data:             data,
		compensationRepo: compensationRepo,
		log:              log.NewHelper(logger),
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

// ConvertHashrate 手动将算力转换为余额
// gRPC 调用（AdjustUserAsset）已移至本地 DB 事务外，避免分布式事务不一致。
// 若本地事务成功但 gRPC 失败，算力已扣但余额未增加，需要补偿机制。
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

	var available float64
	requestID := fmt.Sprintf("hashrate-convert-%d-%d", userID, time.Now().UnixNano())
	reason := buildConvertedHashrateRemark(amount)

	// Step 1: 本地事务仅扣减算力，不涉及跨服务调用
	err = r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var rows []UserHashrate
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND status = ? AND total_hashrate > 0", userID, 1).
			Order("id ASC").
			Find(&rows).Error; err != nil {
			return status.Errorf(codes.Internal, "failed to query user hashrate: %v", err)
		}

		available = 0.0
		for _, row := range rows {
			available += row.TotalHashrate
		}

		if available < amount {
			return status.Errorf(codes.FailedPrecondition, "insufficient hashrate: available %.2f, requested %.2f", available, amount)
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
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Step 2: 跨服务 gRPC 调用（在本地事务外执行）
	updatedUser, err := r.data.UserClient().AdjustUserAsset(ctx, userID, amount, 0, reason, requestID)
	if err != nil {
		// gRPC 失败：算力已扣但余额未增，写入补偿记录表，由定时任务扫描并恢复算力
		compErr := r.compensationRepo.CreateCompensationRecord(ctx, &biz.HashrateCompensation{
			UserID:    userID,
			Amount:    amount,
			RequestID: requestID,
			Reason:    reason,
			Status:    0,
		})
		if compErr != nil {
			r.log.Errorf("failed to create hashrate compensation record for user %d: %v", userID, compErr)
		}
		return nil, status.Errorf(codes.Internal, "user asset adjustment failed after hashrate deducted, compensation record created: %v", err)
	}
	if updatedUser == nil {
		return nil, status.Error(codes.Internal, "user asset adjustment returned nil response")
	}

	// Step 3: 创建余额日志（在事务外，失败不阻塞主流程）
	balanceLog := &BalanceLog{
		UserID:        userID,
		Type:          int8(biz.BalanceLogTypeHashrateConversion),
		Amount:        amount,
		BeforeBalance: userAsset.Balance,
		AfterBalance:  updatedUser.Balance,
		Remark:        reason,
		CreatedAt:     time.Now(),
	}
	if err := r.data.db.WithContext(ctx).Create(balanceLog).Error; err != nil {
		// 日志创建失败不影响核心数据一致性，记录告警日志和监控告警表
		r.log.Errorf("[ALERT] balance log creation failed after hashrate conversion: user_id=%d, amount=%.2f, err=%v", userID, amount, err)
		alert := map[string]interface{}{
			"level":   "warning",
			"module":  "finance.hashrate_conversion",
			"message": fmt.Sprintf("balance log creation failed: user_id=%d, amount=%.2f", userID, amount),
			"detail":  fmt.Sprintf("error: %v", err),
		}
		if dbErr := r.data.db.WithContext(ctx).Table("alert_logs").Create(alert).Error; dbErr != nil {
			r.log.Errorf("[ALERT] failed to write alert log: %v", dbErr)
		}
	}

	return &biz.HashrateConversion{
		UserID:         userID,
		Amount:         amount,
		BeforeHashrate: available,
		AfterHashrate:  available - amount,
		BeforeBalance:  userAsset.Balance,
		AfterBalance:   updatedUser.Balance,
		Remark:         reason,
		CreatedAt:      balanceLog.CreatedAt,
	}, nil
}

const convertedHashrateRemarkPrefix = "手动算力转换"

func (r *userAssetRepo) RestoreHashrate(ctx context.Context, userID uint32, amount float64, requestID string) error {
	var row UserHashrate
	if err := r.data.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, 1).
		Order("id ASC").
		First(&row).Error; err != nil {
		return status.Errorf(codes.Internal, "failed to find user hashrate record for restore: %v", err)
	}

	nextHashrate := row.TotalHashrate + amount
	if err := r.data.db.WithContext(ctx).
		Model(&UserHashrate{}).
		Where("id = ?", row.ID).
		Update("total_hashrate", nextHashrate).Error; err != nil {
		return status.Errorf(codes.Internal, "failed to restore user hashrate: %v", err)
	}

	r.log.Infof("hashrate restored for user %d, request %s, amount %.2f", userID, requestID, amount)
	return nil
}

func buildConvertedHashrateRemark(amount float64) string {
	return fmt.Sprintf("%s %.2f", convertedHashrateRemarkPrefix, amount)
}
