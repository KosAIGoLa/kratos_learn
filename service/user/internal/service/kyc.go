package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	v1 "user/api/user/v1"
	"user/internal/biz"
)

// KYCService 实名认证服务
type KYCService struct {
	v1.UnimplementedUserServer
	uc  *biz.KYCUsecase
	log *log.Helper
}

// NewKYCService 创建实名认证服务
func NewKYCService(uc *biz.KYCUsecase, logger log.Logger) *KYCService {
	return &KYCService{
		uc:  uc,
		log: log.NewHelper(logger),
	}
}

// SubmitKYC 提交实名认证
func (s *KYCService) SubmitKYC(ctx context.Context, req *v1.SubmitKYCRequest) (*v1.KYCInfo, error) {
	kyc, err := s.uc.SubmitKYC(ctx, &biz.KYCVerification{
		UserID:      req.UserId,
		Name:        req.Name,
		IDCard:      req.IdCard,
		IDCardFront: req.IdCardFront,
		IDCardBack:  req.IdCardBack,
	})
	if err != nil {
		return nil, err
	}
	return &v1.KYCInfo{
		Id:          kyc.ID,
		UserId:      kyc.UserID,
		Name:        kyc.Name,
		IdCard:      kyc.IDCard,
		IdCardFront: kyc.IDCardFront,
		IdCardBack:  kyc.IDCardBack,
		Status:      kyc.Status,
		Remark:      kyc.Remark,
	}, nil
}

// GetKYC 获取实名认证信息
func (s *KYCService) GetKYC(ctx context.Context, req *v1.GetKYCRequest) (*v1.KYCInfo, error) {
	kyc, err := s.uc.GetKYCByUserID(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	return &v1.KYCInfo{
		Id:          kyc.ID,
		UserId:      kyc.UserID,
		Name:        kyc.Name,
		IdCard:      kyc.IDCard,
		IdCardFront: kyc.IDCardFront,
		IdCardBack:  kyc.IDCardBack,
		Status:      kyc.Status,
		Remark:      kyc.Remark,
	}, nil
}
