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

// ListIncomeLogs 获取收益明细列表
func (s *FinanceService) ListIncomeLogs(ctx context.Context, req *v1.ListIncomeLogsRequest) (*v1.ListIncomeLogsResponse, error) {
	logs, total, err := s.uc.ListIncomeLogs(ctx, req.UserId, req.SourceType, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	resp := &v1.ListIncomeLogsResponse{Total: total}
	for _, item := range logs {
		resp.IncomeLogs = append(resp.IncomeLogs, &v1.IncomeLogInfo{
			Id:         item.ID,
			UserId:     item.UserID,
			Phone:      item.Phone,
			Name:       item.Name,
			Source:     item.Source,
			SourceType: item.SourceType,
			Amount:     item.Amount,
			RelatedId:  item.RelatedID,
			CreatedAt:  timestamppb.New(item.CreatedAt),
		})
	}
	return resp, nil
}

// ListBalanceLogs 获取余额变动列表
func (s *FinanceService) ListBalanceLogs(ctx context.Context, req *v1.ListBalanceLogsRequest) (*v1.ListBalanceLogsResponse, error) {
	logs, total, err := s.uc.ListBalanceLogs(ctx, req.UserId, req.Type, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	resp := &v1.ListBalanceLogsResponse{Total: total}
	for _, item := range logs {
		resp.BalanceLogs = append(resp.BalanceLogs, s.toProtoBalanceLog(item))
	}
	return resp, nil
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
	balance, workPoints, err := s.uc.GetUserBalance(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	return &v1.UserBalanceInfo{
		UserId:     req.UserId,
		Balance:    balance,
		WorkPoints: workPoints,
	}, nil
}

// ConvertHashPower 手动算力转余额
func (s *FinanceService) ConvertHashPower(ctx context.Context, req *v1.ConvertHashPowerRequest) (*v1.HashPowerConversionInfo, error) {
	conversion, err := s.uc.ConvertHashrate(ctx, req.UserId, req.Amount)
	if err != nil {
		return nil, err
	}

	return &v1.HashPowerConversionInfo{
		UserId:         conversion.UserID,
		Amount:         conversion.Amount,
		BeforeHashrate: conversion.BeforeHashrate,
		AfterHashrate:  conversion.AfterHashrate,
		BeforeBalance:  conversion.BeforeBalance,
		AfterBalance:   conversion.AfterBalance,
		Remark:         conversion.Remark,
		CreatedAt:      timestamppb.New(conversion.CreatedAt),
	}, nil
}

// CreateBalanceLog 创建余额变动记录
func (s *FinanceService) CreateBalanceLog(ctx context.Context, req *v1.CreateBalanceLogRequest) (*v1.BalanceLogInfo, error) {
	log, err := s.uc.CreateBalanceLog(ctx, &biz.BalanceLog{
		UserID:        req.UserId,
		Type:          req.Type,
		Amount:        req.Amount,
		BeforeBalance: req.BeforeBalance,
		AfterBalance:  req.AfterBalance,
		Remark:        req.Remark,
		RelatedID:     req.RelatedId,
	})
	if err != nil {
		return nil, err
	}
	return s.toProtoBalanceLog(log), nil
}

func (s *FinanceService) toProtoBalanceLog(l *biz.BalanceLog) *v1.BalanceLogInfo {
	return &v1.BalanceLogInfo{
		Id:            l.ID,
		UserId:        l.UserID,
		Type:          l.Type,
		Amount:        l.Amount,
		BeforeBalance: l.BeforeBalance,
		AfterBalance:  l.AfterBalance,
		Remark:        l.Remark,
		RelatedId:     l.RelatedID,
		CreatedAt:     timestamppb.New(l.CreatedAt),
	}
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
