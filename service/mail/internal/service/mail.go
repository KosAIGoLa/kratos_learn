package service

import (
	"context"
	"time"

	pb "mail/api/mail/v1"
	"mail/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MailService implements the mail gRPC service
type MailService struct {
	pb.UnimplementedMailServer

	uc     *biz.EmailUsecase
	logger *log.Helper
}

// NewMailService creates a new mail service
func NewMailService(uc *biz.EmailUsecase, logger log.Logger) *MailService {
	return &MailService{
		uc:     uc,
		logger: log.NewHelper(logger),
	}
}

// SendEmail implements the SendEmail RPC
func (s *MailService) SendEmail(ctx context.Context, req *pb.SendEmailRequest) (*pb.SendEmailResponse, error) {
	s.logger.Infof("SendEmail called: to=%s, subject=%s", req.Email.To, req.Email.Subject)

	// Convert proto email to biz email
	email := &biz.Email{
		To:          req.Email.To,
		Subject:     req.Email.Subject,
		Body:        req.Email.Body,
		ContentType: req.Email.ContentType,
	}

	// Convert attachments
	var attachments []*biz.Attachment
	for _, att := range req.Email.Attachments {
		attachments = append(attachments, &biz.Attachment{
			Filename:    att.Filename,
			Content:     att.Content,
			ContentType: att.ContentType,
		})
	}

	// Create send options
	opts := &biz.SendOptions{
		TemplateID:   req.TemplateId,
		TemplateData: req.TemplateData,
		From:         req.From,
		FromName:     req.FromName,
		Attachments:  attachments,
	}

	// Call usecase
	messageID, err := s.uc.SendEmail(ctx, email, opts)
	if err != nil {
		return nil, err
	}

	return &pb.SendEmailResponse{
		MessageId: messageID,
		Status:    string(biz.EmailStatusPending),
		CreatedAt: timestamppb.New(time.Now()),
	}, nil
}

// SendBatchEmail implements the SendBatchEmail RPC
func (s *MailService) SendBatchEmail(ctx context.Context, req *pb.SendBatchEmailRequest) (*pb.SendBatchEmailResponse, error) {
	s.logger.Infof("SendBatchEmail called: count=%d", len(req.Emails))

	// Convert proto emails to biz emails
	var emails []*biz.Email
	for _, e := range req.Emails {
		emails = append(emails, &biz.Email{
			To:          e.To,
			Subject:     e.Subject,
			Body:        e.Body,
			ContentType: e.ContentType,
		})
	}

	// Create batch options
	rateLimit := time.Second
	if req.RateLimit != nil {
		rateLimit = req.RateLimit.AsDuration()
	}

	opts := &biz.BatchOptions{
		TemplateID:   req.TemplateId,
		TemplateData: req.TemplateData,
		From:         req.From,
		FromName:     req.FromName,
		BatchSize:    req.BatchSize,
		RateLimit:    rateLimit,
	}

	// Call usecase
	batch, err := s.uc.SendBatchEmail(ctx, emails, opts)
	if err != nil {
		return nil, err
	}

	return &pb.SendBatchEmailResponse{
		BatchId:   batch.ID,
		Total:     batch.Total,
		Accepted:  batch.Sent,
		Rejected:  batch.Failed,
		Status:    string(batch.Status),
		CreatedAt: timestamppb.New(batch.CreatedAt),
	}, nil
}

// GetEmailStatus implements the GetEmailStatus RPC
func (s *MailService) GetEmailStatus(ctx context.Context, req *pb.GetEmailStatusRequest) (*pb.GetEmailStatusResponse, error) {
	s.logger.Infof("GetEmailStatus called: message_id=%s", req.MessageId)

	email, err := s.uc.GetEmailStatus(ctx, req.MessageId)
	if err != nil {
		return nil, err
	}

	resp := &pb.GetEmailStatusResponse{
		MessageId:    email.MessageID,
		Status:       string(email.Status),
		ErrorMessage: email.ErrorMsg,
	}

	if email.SentAt != nil {
		resp.SentAt = timestamppb.New(*email.SentAt)
	}
	if email.DeliveredAt != nil {
		resp.DeliveredAt = timestamppb.New(*email.DeliveredAt)
	}
	if email.OpenedAt != nil {
		resp.OpenedAt = timestamppb.New(*email.OpenedAt)
	}

	return resp, nil
}

// ListEmailTemplates implements the ListEmailTemplates RPC
func (s *MailService) ListEmailTemplates(ctx context.Context, req *pb.ListEmailTemplatesRequest) (*pb.ListEmailTemplatesResponse, error) {
	s.logger.Infof("ListEmailTemplates called: page=%d, category=%s", req.Page, req.Category)

	page := int(req.Page)
	if page <= 0 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 10
	}

	templates, total, err := s.uc.ListTemplates(ctx, req.Category, page, pageSize)
	if err != nil {
		return nil, err
	}

	var pbTemplates []*pb.ListEmailTemplatesResponse_Template
	for _, t := range templates {
		pbTemplates = append(pbTemplates, &pb.ListEmailTemplatesResponse_Template{
			Id:          t.ID,
			Name:        t.Name,
			Subject:     t.Subject,
			Content:     t.Content,
			ContentType: t.ContentType,
			Category:    t.Category,
			CreatedAt:   timestamppb.New(t.CreatedAt),
			UpdatedAt:   timestamppb.New(t.UpdatedAt),
		})
	}

	return &pb.ListEmailTemplatesResponse{
		Templates: pbTemplates,
		Total:     total,
	}, nil
}

// GetBatchStatus implements the GetBatchStatus RPC
func (s *MailService) GetBatchStatus(ctx context.Context, req *pb.GetBatchStatusRequest) (*pb.GetBatchStatusResponse, error) {
	s.logger.Infof("GetBatchStatus called: batch_id=%s", req.BatchId)

	batch, err := s.uc.GetBatchStatus(ctx, req.BatchId)
	if err != nil {
		return nil, err
	}

	resp := &pb.GetBatchStatusResponse{
		BatchId:   batch.ID,
		Status:    string(batch.Status),
		Total:     batch.Total,
		Sent:      batch.Sent,
		Failed:    batch.Failed,
		Pending:   batch.Pending,
		CreatedAt: timestamppb.New(batch.CreatedAt),
	}

	if batch.CompletedAt != nil {
		resp.CompletedAt = timestamppb.New(*batch.CompletedAt)
	}

	return resp, nil
}
