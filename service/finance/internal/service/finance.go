package service

import (
	"context"
	"time"

	v1 "finance/api/finance/v1"
	"finance/internal/biz"
	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// FinanceService 财务服务
type FinanceService struct {
	v1.UnimplementedFinanceServer
	uc  *biz.FinanceUsecase
	log *log.Helper
}

// NewFinanceService 创建财务服务
func NewFinanceService(uc *biz.FinanceUsecase, logger log.Logger) *FinanceService {
	return &FinanceService{
		uc:  uc,
		log: log.NewHelper(logger),
	}
}

// Recharge 充值
func (s *FinanceService) Recharge(ctx context.Context, req *v1.RechargeRequest) (*v1.RechargeInfo, error) {
	recharge, err := s.uc.Recharge(ctx, &biz.Recharge{
		UserID: req.UserId,
		Amount: req.Amount,
	})
	if err != nil {
		return nil, err
	}
	return s.toProtoRecharge(recharge), nil
}

// GetRecharge 获取充值记录
func (s *FinanceService) GetRecharge(ctx context.Context, req *v1.GetRechargeRequest) (*v1.RechargeInfo, error) {
	recharge, err := s.uc.GetRechargeByOrderNo(ctx, req.OrderNo)
	if err != nil {
		return nil, err
	}
	return s.toProtoRecharge(recharge), nil
}

// ListRecharges 获取充值列表
func (s *FinanceService) ListRecharges(ctx context.Context, req *v1.ListRechargesRequest) (*v1.ListRechargesResponse, error) {
	recharges, total, err := s.uc.ListRecharges(ctx, req.UserId, req.Status, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	var protoRecharges []*v1.RechargeInfo
	for _, r := range recharges {
		protoRecharges = append(protoRecharges, s.toProtoRecharge(r))
	}

	return &v1.ListRechargesResponse{
		Recharges: protoRecharges,
		Total:     total,
	}, nil
}

// Withdraw 提现
func (s *FinanceService) Withdraw(ctx context.Context, req *v1.WithdrawRequest) (*v1.WithdrawalInfo, error) {
	withdrawal, err := s.uc.Withdraw(ctx, &biz.Withdrawal{
		UserID:   req.UserId,
		Amount:   req.Amount,
		BankCard: req.BankCard,
		BankName: req.BankName,
	})
	if err != nil {
		return nil, err
	}
	return s.toProtoWithdrawal(withdrawal), nil
}

// GetWithdrawal 获取提现记录
func (s *FinanceService) GetWithdrawal(ctx context.Context, req *v1.GetWithdrawalRequest) (*v1.WithdrawalInfo, error) {
	withdrawal, err := s.uc.GetWithdrawal(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return s.toProtoWithdrawal(withdrawal), nil
}

// ListWithdrawals 获取提现列表
func (s *FinanceService) ListWithdrawals(ctx context.Context, req *v1.ListWithdrawalsRequest) (*v1.ListWithdrawalsResponse, error) {
	withdrawals, total, err := s.uc.ListWithdrawals(ctx, req.UserId, req.Status, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	var protoWithdrawals []*v1.WithdrawalInfo
	for _, w := range withdrawals {
		protoWithdrawals = append(protoWithdrawals, s.toProtoWithdrawal(w))
	}

	return &v1.ListWithdrawalsResponse{
		Withdrawals: protoWithdrawals,
		Total:       total,
	}, nil
}

// CheckIn 签到
func (s *FinanceService) CheckIn(ctx context.Context, req *v1.CheckInRequest) (*v1.CheckInResponse, error) {
	checkIn, err := s.uc.CheckIn(ctx, &biz.CheckIn{
		UserID:          req.UserId,
		CheckInDate:     time.Now().Format("2006-01-02"),
		ConsecutiveDays: 1,
		RewardPoints:    10,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CheckInResponse{
		Success:         true,
		ConsecutiveDays: checkIn.ConsecutiveDays,
		RewardPoints:    checkIn.RewardPoints,
	}, nil
}

// GetUserBalance 获取用户余额
func (s *FinanceService) GetUserBalance(ctx context.Context, req *v1.GetUserBalanceRequest) (*v1.UserBalanceInfo, error) {
	// TODO: 从用户服务获取余额
	return &v1.UserBalanceInfo{
		UserId:     req.UserId,
		Balance:    0,
		WorkPoints: 0,
	}, nil
}

func (s *FinanceService) toProtoRecharge(r *biz.Recharge) *v1.RechargeInfo {
	return &v1.RechargeInfo{
		Id:         r.ID,
		UserId:     r.UserID,
		InviteCode: r.InviteCode,
		Phone:      r.Phone,
		Name:       r.Name,
		OrderNo:    r.OrderNo,
		Amount:     r.Amount,
		Status:     r.Status,
		CreatedAt:  timestamppb.New(r.CreatedAt),
	}
}

func (s *FinanceService) toProtoWithdrawal(w *biz.Withdrawal) *v1.WithdrawalInfo {
	return &v1.WithdrawalInfo{
		Id:       w.ID,
		UserId:   w.UserID,
		Phone:    w.Phone,
		Name:     w.Name,
		BankCard: w.BankCard,
		BankName: w.BankName,
		Amount:   w.Amount,
		Status:   w.Status,
		Remark:   w.Remark,
		ProcessedAt: func() *timestamppb.Timestamp {
			if w.ProcessedAt != nil {
				return timestamppb.New(*w.ProcessedAt)
			}
			return nil
		}(),
		CreatedAt: timestamppb.New(w.CreatedAt),
	}
}
