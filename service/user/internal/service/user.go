package service

import (
	"context"

	v1 "user/api/user/v1"
	"user/internal/biz"
	"user/internal/pkg/captcha"
	"user/internal/pkg/jwt"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

// UserService 用户服务
type UserService struct {
	v1.UnimplementedUserServer
	uc             *biz.UserUsecase
	jwtManager     *jwt.JWTManager
	captchaManager *captcha.CaptchaManager
	log            *log.Helper
}

// NewUserService 创建用户服务
func NewUserService(uc *biz.UserUsecase, jwtManager *jwt.JWTManager, captchaManager *captcha.CaptchaManager, logger log.Logger) *UserService {
	return &UserService{
		uc:             uc,
		jwtManager:     jwtManager,
		captchaManager: captchaManager,
		log:            log.NewHelper(logger),
	}
}

// Register 用户注册
func (s *UserService) Register(ctx context.Context, req *v1.RegisterRequest) (*v1.LoginResponse, error) {

	hashedPassword, err := biz.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	var parentId uint32 = 0
	if req.InviteCode != "" {
		parentUser, err := s.uc.GetUserByInviteCode(ctx, req.InviteCode)
		if err == nil {
			parentId = parentUser.ID
		}
	}

	user, err := s.uc.Register(ctx, &biz.User{
		Username: req.Username,
		Phone:    req.Phone,
		ParentID: &parentId,
		//Password:   req.Password,
		Password:   hashedPassword,
		InviteCode: req.InviteCode,
	})
	if err != nil {
		return nil, err
	}

	// 给邀请人奖励工分（异步执行，不影响注册流程）
	if parentId > 0 {
		go func() {
			if err := s.uc.RewardInviteWorkPoints(context.Background(), parentId); err != nil {
				s.log.Errorf("邀请奖励发放失败: parent_id=%d, err=%v", parentId, err)
			}
		}()
	}

	// 生成 token 和 refresh_token
	token, err := s.jwtManager.GenerateToken(user.ID, user.Username, user.Phone)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, user.Username, user.Phone)
	if err != nil {
		return nil, err
	}

	return &v1.LoginResponse{
		User:         s.toProtoUser(user),
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

// GetCaptcha 获取验证码
func (s *UserService) GetCaptcha(ctx context.Context, req *v1.GetCaptchaRequest) (*v1.GetCaptchaResponse, error) {
	id, b64s, err := s.captchaManager.Generate()
	if err != nil {
		return nil, errors.New(500, "CAPTCHA_GENERATE_FAILED", "验证码生成失败")
	}
	return &v1.GetCaptchaResponse{
		CaptchaId:    id,
		CaptchaImage: b64s,
	}, nil
}

// Login 用户登录
func (s *UserService) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	if !s.captchaManager.Verify(req.CaptchaId, req.CaptchaCode, true) {
		return nil, errors.New(400, "CAPTCHA_VERIFY_FAILED", "验证码错误")
	}

	user, err := s.uc.GetUserByPhone(ctx, req.Phone)
	if err != nil {
		return nil, err
	}

	// 驗證密碼
	if err := biz.VerifyPassword(user.Password, req.Password); err != nil {
		return nil, errors.New(400, "INVALID_PASSWORD", "密碼錯誤")
	}

	// 生成 token 和 refresh_token
	token, err := s.jwtManager.GenerateToken(user.ID, user.Username, user.Phone)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, user.Username, user.Phone)
	if err != nil {
		return nil, err
	}

	return &v1.LoginResponse{
		User:         s.toProtoUser(user),
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshToken 刷新 token
func (s *UserService) RefreshToken(ctx context.Context, req *v1.RefreshTokenRequest) (*v1.LoginResponse, error) {
	// 解析 refresh_token
	claims, err := s.jwtManager.ParseRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, err
	}

	// 獲取用戶信息
	user, err := s.uc.GetUser(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	// 生成新的 token 和 refresh_token
	token, err := s.jwtManager.GenerateToken(user.ID, user.Username, user.Phone)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, user.Username, user.Phone)
	if err != nil {
		return nil, err
	}

	return &v1.LoginResponse{
		User:         s.toProtoUser(user),
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

// GetUser 获取用户信息
func (s *UserService) GetUser(ctx context.Context, req *v1.GetUserRequest) (*v1.UserInfo, error) {
	user, err := s.uc.GetUser(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return s.toProtoUser(user), nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(ctx context.Context, req *v1.DeleteUserRequest) (*v1.DeleteUserResponse, error) {
	err := s.uc.DeleteUser(ctx, req.Id)
	if err != nil {
		return &v1.DeleteUserResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}
	return &v1.DeleteUserResponse{
		Success: true,
		Message: "User deleted successfully",
	}, nil
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(ctx context.Context, req *v1.UpdateUserRequest) (*v1.UserInfo, error) {
	user, err := s.uc.UpdateUser(ctx, &biz.User{
		ID:     req.Id,
		Name:   req.Name,
		IDCard: req.IdCard,
	})
	if err != nil {
		return nil, err
	}
	return s.toProtoUser(user), nil
}

// AdjustUserAsset 调整用户资产
func (s *UserService) AdjustUserAsset(ctx context.Context, req *v1.AdjustUserAssetRequest) (*v1.UserInfo, error) {
	user, err := s.uc.AdjustUserAsset(ctx, req.UserId, req.BalanceDelta, req.WorkPointsDelta)
	if err != nil {
		return nil, err
	}
	return s.toProtoUser(user), nil
}

func (s *UserService) toProtoUser(u *biz.User) *v1.UserInfo {
	return &v1.UserInfo{
		Id:         u.ID,
		Username:   u.Username,
		InviteCode: u.InviteCode,
		Phone:      u.Phone,
		Name:       u.Name,
		IdCard:     u.IDCard,
		ParentId: func() uint32 {
			if u.ParentID != nil {
				return *u.ParentID
			}
			return 0
		}(),
		Balance:    u.Balance,
		WorkPoints: u.WorkPoints,
		Status:     u.Status,
	}
}
