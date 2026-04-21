package service

import (
	"context"
	"fmt"
	"strconv"

	v1 "finance/api/finance/v1"
	"finance/internal/biz"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// FinanceService 财务服务
type FinanceService struct {
	v1.UnimplementedFinanceServer
	uc  *biz.FinanceUsecase
	log *log.Helper
}

func getCurrentUserID(ctx context.Context) (uint32, bool) {
	md, ok := metadata.FromServerContext(ctx)
	if !ok {
		return 0, false
	}
	userIDStr := md.Get("x-md-user-id")
	if userIDStr == "" {
		return 0, false
	}
	uid, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return 0, false
	}
	return uint32(uid), true
}

func checkUserOwnership(ctx context.Context, reqUserID uint32) error {
	currentUserID, ok := getCurrentUserID(ctx)
	if !ok || reqUserID == 0 {
		return nil
	}
	if reqUserID != currentUserID {
		return fmt.Errorf("unauthorized: user_id mismatch")
	}
	return nil
}

func checkInternalService(ctx context.Context) error {
	md, ok := metadata.FromServerContext(ctx)
	if !ok {
		return fmt.Errorf("unauthorized: internal service only")
	}
	if md.Get("x-md-internal-service") == "" {
		return fmt.Errorf("unauthorized: internal service only")
	}
	return nil
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
	if err := checkUserOwnership(ctx, req.UserId); err != nil {
		return nil, err
	}
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
	if err := checkUserOwnership(ctx, req.UserId); err != nil {
		return nil, err
	}
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
	if err := checkUserOwnership(ctx, req.UserId); err != nil {
		return nil, err
	}
	checkIn, err := s.uc.CheckIn(ctx, &biz.CheckIn{
		UserID: req.UserId,
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
	if err := checkUserOwnership(ctx, req.UserId); err != nil {
		return nil, err
	}
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

// ListHashrateCompensations 查询算力补偿记录列表
func (s *FinanceService) ListHashrateCompensations(ctx context.Context, req *v1.ListHashrateCompensationsRequest) (*v1.ListHashrateCompensationsResponse, error) {
	if err := checkInternalService(ctx); err != nil {
		return nil, err
	}
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 100
	}
	records, err := s.uc.ListPendingHashrateCompensations(ctx, limit)
	if err != nil {
		return nil, err
	}

	var resp []*v1.HashrateCompensationInfo
	for _, r := range records {
		item := &v1.HashrateCompensationInfo{
			Id:         r.ID,
			UserId:     r.UserID,
			Amount:     r.Amount,
			RequestId:  r.RequestID,
			Reason:     r.Reason,
			Status:     int32(r.Status),
			RetryTimes: r.RetryTimes,
			CreatedAt:  timestamppb.New(r.CreatedAt),
		}
		if r.CompensatedAt != nil {
			item.CompensatedAt = timestamppb.New(*r.CompensatedAt)
		}
		resp = append(resp, item)
	}
	return &v1.ListHashrateCompensationsResponse{Records: resp}, nil
}

// CompensateHashrate 执行算力补偿
func (s *FinanceService) CompensateHashrate(ctx context.Context, req *v1.CompensateHashrateRequest) (*v1.CompensateHashrateResponse, error) {
	if err := checkInternalService(ctx); err != nil {
		return nil, err
	}
	records, err := s.uc.ListPendingHashrateCompensations(ctx, 1)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return &v1.CompensateHashrateResponse{Success: false, Message: "no pending compensation found"}, nil
	}

	var target *biz.HashrateCompensation
	for _, r := range records {
		if r.ID == req.Id {
			target = r
			break
		}
	}
	if target == nil {
		return &v1.CompensateHashrateResponse{Success: false, Message: "compensation record not found or already processed"}, nil
	}

	if err := s.uc.CompensateHashrate(ctx, target); err != nil {
		return &v1.CompensateHashrateResponse{Success: false, Message: err.Error()}, nil
	}
	return &v1.CompensateHashrateResponse{Success: true, Message: "compensation completed"}, nil
}

// CreateBalanceLog 创建余额变动记录
// 仅允许内部服务调用，外部用户应通过业务接口（如 Recharge/Withdraw）间接产生流水。
func (s *FinanceService) CreateBalanceLog(ctx context.Context, req *v1.CreateBalanceLogRequest) (*v1.BalanceLogInfo, error) {
	if err := checkInternalService(ctx); err != nil {
		return nil, err
	}
	if req.Type <= 0 || req.Amount <= 0 {
		return nil, fmt.Errorf("invalid balance log: type and amount must be greater than 0")
	}
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
