package data

import (
	"context"
	"time"
	"user/internal/pkg/invite"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"user/internal/biz"
)

// User 用户数据模型
type User struct {
	ID              uint32  `gorm:"primarykey"`
	Username        string  `gorm:"uniqueIndex:idx_username;type:varchar(50);not null"`
	InviteCode      string  `gorm:"uniqueIndex:idx_invite_code;type:varchar(20);not null"`
	Phone           string  `gorm:"uniqueIndex:idx_phone;type:char(11);not null"`
	Password        string  `gorm:"type:varchar(255);not null"`
	PaymentPassword string  `gorm:"type:varchar(255)"`
	Name            string  `gorm:"type:varchar(50)"`
	IDCard          string  `gorm:"type:char(18)"`
	ParentID        *uint32 `gorm:"index:idx_parent_id"`
	Balance         float64 `gorm:"type:decimal(15,2);default:0.00"`
	WorkPoints      float64 `gorm:"type:decimal(15,2);default:0.00"`
	Status          int8    `gorm:"index:idx_status;default:1"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

type userRepo struct {
	data *Data
	log  *log.Helper
}

// NewUserRepo 创建用户仓库
func NewUserRepo(data *Data, logger log.Logger) biz.UserRepo {
	return &userRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *userRepo) CreateUser(ctx context.Context, u *biz.User) (*biz.User, error) {
	if u == nil {
		return nil, status.Errorf(codes.InvalidArgument, "user is nil")
	}
	//產生邀請碼
	inviteCode, _ := invite.GenerateInviteCode(6)
	user := User{
		Username:   u.Username,
		InviteCode: inviteCode,
		Phone:      u.Phone,
		Password:   u.Password,
		ParentID:   u.ParentID,
		Status:     1,
	}

	if err := r.data.db.Create(&user).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}
	return r.toBizUser(&user), nil
}

func (r *userRepo) GetUserByID(ctx context.Context, id uint32) (*biz.User, error) {
	var user User
	if err := r.data.db.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "用户不存在")
		}
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}
	return r.toBizUser(&user), nil
}

func (r *userRepo) GetUserByPhone(ctx context.Context, phone string) (*biz.User, error) {
	var user User
	if err := r.data.db.Where("phone = ?", phone).First(&user).Error; err != nil {
		return nil, err
	}
	return r.toBizUser(&user), nil
}

func (r *userRepo) GetUserByUsername(ctx context.Context, username string) (*biz.User, error) {
	var user User
	if err := r.data.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return r.toBizUser(&user), nil
}

func (r *userRepo) GetUserByInviteCode(ctx context.Context, inviteCode string) (*biz.User, error) {
	var user User
	if err := r.data.db.Where("invite_code = ?", inviteCode).First(&user).Error; err != nil {
		return nil, err
	}
	return r.toBizUser(&user), nil
}

func (r *userRepo) UpdateUser(ctx context.Context, u *biz.User) (*biz.User, error) {
	updates := map[string]interface{}{}
	if u.Name != "" {
		updates["name"] = u.Name
	}
	if u.IDCard != "" {
		updates["id_card"] = u.IDCard
	}
	if err := r.data.db.Model(&User{}).Where("id = ?", u.ID).Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.GetUserByID(ctx, u.ID)
}

func (r *userRepo) AdjustUserAsset(ctx context.Context, userID uint32, balanceDelta, workPointsDelta float64) (*biz.User, error) {
	var updated *biz.User
	err := r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user User
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", userID).
			First(&user).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return status.Errorf(codes.NotFound, "用户不存在: %d", userID)
			}
			return status.Errorf(codes.Internal, "查询用户失败: %s", err.Error())
		}

		nextBalance := user.Balance + balanceDelta
		nextWorkPoints := user.WorkPoints + workPointsDelta
		if nextBalance < 0 {
			return status.Errorf(codes.FailedPrecondition, "余额不足: current=%.2f delta=%.2f", user.Balance, balanceDelta)
		}
		if nextWorkPoints < 0 {
			return status.Errorf(codes.FailedPrecondition, "工分不足: current=%.2f delta=%.2f", user.WorkPoints, workPointsDelta)
		}

		user.Balance = nextBalance
		user.WorkPoints = nextWorkPoints
		if err := tx.Save(&user).Error; err != nil {
			return status.Errorf(codes.Internal, "更新用户资产失败: %s", err.Error())
		}

		updated = r.toBizUser(&user)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (r *userRepo) DeleteUser(ctx context.Context, id uint32) error {
	if err := r.data.db.Delete(&User{}, id).Error; err != nil {
		return status.Errorf(codes.Internal, "%s", err.Error())
	}
	return nil
}

// UpdateWorkPoints 更新用户工分（增加）- 使用事务和行锁
func (r *userRepo) UpdateWorkPoints(ctx context.Context, userID uint32, points float64) error {
	return r.data.db.Transaction(func(tx *gorm.DB) error {
		// 1. 查询用户并加行锁 (FOR UPDATE)
		var user User
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", userID).
			First(&user).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return status.Errorf(codes.NotFound, "用户不存在: %d", userID)
			}
			return status.Errorf(codes.Internal, "查询用户失败: %s", err.Error())
		}

		// 2. 更新工分（原子增加）
		user.WorkPoints += points
		if err := tx.Save(&user).Error; err != nil {
			return status.Errorf(codes.Internal, "更新工分失败: %s", err.Error())
		}

		r.log.Infof("工分更新成功: user_id=%d, added=%.2f, new_total=%.2f", userID, points, user.WorkPoints)
		return nil
	})
}

func (r *userRepo) toBizUser(u *User) *biz.User {
	return &biz.User{
		ID:         u.ID,
		Username:   u.Username,
		InviteCode: u.InviteCode,
		Phone:      u.Phone,
		Password:   u.Password,
		Name:       u.Name,
		IDCard:     u.IDCard,
		ParentID:   u.ParentID,
		Balance:    u.Balance,
		WorkPoints: u.WorkPoints,
		Status:     int32(u.Status),
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
	}
}
