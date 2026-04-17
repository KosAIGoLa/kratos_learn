package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/crypto/bcrypt"
)

// User 用户领域模型
type User struct {
	ID         uint32
	Username   string
	InviteCode string
	Phone      string
	Password   string
	Name       string
	IDCard     string
	ParentID   *uint32
	Balance    float64
	WorkPoints float64
	Status     int32
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// KYCVerification 实名认证领域模型
type KYCVerification struct {
	ID          uint32
	UserID      uint32
	Name        string
	IDCard      string
	IDCardFront string
	IDCardBack  string
	Status      int32
	Remark      string
	VerifiedAt  *time.Time
	CreatedAt   time.Time
}

// TeamRelation 团队关系领域模型
type TeamRelation struct {
	ID        uint32
	UserID    uint32
	ParentID  *uint32
	Path      string
	Level     uint32
	CreatedAt time.Time
}

// UserRepo 用户存储接口
type UserRepo interface {
	CreateUser(ctx context.Context, u *User) (*User, error)
	GetUserByID(ctx context.Context, id uint32) (*User, error)
	GetUserByPhone(ctx context.Context, phone string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	GetUserByInviteCode(ctx context.Context, inviteCode string) (*User, error)
	UpdateUser(ctx context.Context, u *User) (*User, error)
	DeleteUser(ctx context.Context, id uint32) error
}

// KYCRepo 实名认证存储接口
type KYCRepo interface {
	CreateKYC(ctx context.Context, k *KYCVerification) (*KYCVerification, error)
	GetKYCByUserID(ctx context.Context, userID uint32) (*KYCVerification, error)
	UpdateKYCStatus(ctx context.Context, id uint32, status int32, remark string) error
}

// TeamRepo 团队关系存储接口
type TeamRepo interface {
	CreateTeamRelation(ctx context.Context, t *TeamRelation) (*TeamRelation, error)
	GetTeamRelationByUserID(ctx context.Context, userID uint32) (*TeamRelation, error)
	GetTeamMembers(ctx context.Context, parentID uint32, level int32) ([]*User, error)
	GetAllDescendants(ctx context.Context, userID uint32) ([]*User, error)
}

// UserUsecase 用户用例
type UserUsecase struct {
	repo UserRepo
	log  *log.Helper
}

// NewUserUsecase 创建用户用例
func NewUserUsecase(repo UserRepo, logger log.Logger) *UserUsecase {
	return &UserUsecase{repo: repo, log: log.NewHelper(logger)}
}

// Register 用户注册
func (uc *UserUsecase) Register(ctx context.Context, u *User) (*User, error) {
	// 對密碼進行 bcrypt 哈希
	return uc.repo.CreateUser(ctx, u)
}

// HashPassword 使用 bcrypt 哈希密碼
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// VerifyPassword 驗證 bcrypt 密碼
func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// GetUser 获取用户信息
func (uc *UserUsecase) GetUser(ctx context.Context, id uint32) (*User, error) {
	return uc.repo.GetUserByID(ctx, id)
}

// GetUserByPhone 根据手机号获取用户
func (uc *UserUsecase) GetUserByPhone(ctx context.Context, phone string) (*User, error) {
	return uc.repo.GetUserByPhone(ctx, phone)
}

// GetUserByInviteCode 根据邀请码获取用户
func (uc *UserUsecase) GetUserByInviteCode(ctx context.Context, inviteCode string) (*User, error) {
	return uc.repo.GetUserByInviteCode(ctx, inviteCode)
}

// UpdateUser 更新用户信息
func (uc *UserUsecase) UpdateUser(ctx context.Context, u *User) (*User, error) {
	return uc.repo.UpdateUser(ctx, u)
}

// DeleteUser 删除用户
func (uc *UserUsecase) DeleteUser(ctx context.Context, id uint32) error {
	return uc.repo.DeleteUser(ctx, id)
}

// KYCUsecase 实名认证用例
type KYCUsecase struct {
	repo KYCRepo
	log  *log.Helper
}

// NewKYCUsecase 创建实名认证用例
func NewKYCUsecase(repo KYCRepo, logger log.Logger) *KYCUsecase {
	return &KYCUsecase{repo: repo, log: log.NewHelper(logger)}
}

// GetKYCByUserID 根据用户ID获取实名认证
func (uc *KYCUsecase) GetKYCByUserID(ctx context.Context, userID uint32) (*KYCVerification, error) {
	return uc.repo.GetKYCByUserID(ctx, userID)
}

// SubmitKYC 提交实名认证
func (uc *KYCUsecase) SubmitKYC(ctx context.Context, k *KYCVerification) (*KYCVerification, error) {
	return uc.repo.CreateKYC(ctx, k)
}

// TeamUsecase 团队关系用例
type TeamUsecase struct {
	repo TeamRepo
	log  *log.Helper
}

// NewTeamUsecase 创建团队关系用例
func NewTeamUsecase(repo TeamRepo, logger log.Logger) *TeamUsecase {
	return &TeamUsecase{repo: repo, log: log.NewHelper(logger)}
}

// GetTeamRelationByUserID 根据用户ID获取团队关系
func (uc *TeamUsecase) GetTeamRelationByUserID(ctx context.Context, userID uint32) (*TeamRelation, error) {
	return uc.repo.GetTeamRelationByUserID(ctx, userID)
}

// GetTeamMembers 获取团队成员
func (uc *TeamUsecase) GetTeamMembers(ctx context.Context, parentID uint32, level int32) ([]*User, error) {
	return uc.repo.GetTeamMembers(ctx, parentID, level)
}
