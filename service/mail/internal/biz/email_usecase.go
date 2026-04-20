package biz

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// EmailRepo is the email repository interface
type EmailRepo interface {
	Save(ctx context.Context, email *Email) error
	Update(ctx context.Context, email *Email) error
	FindByID(ctx context.Context, id string) (*Email, error)
	FindByMessageID(ctx context.Context, messageID string) (*Email, error)
	SaveBatch(ctx context.Context, batch *EmailBatch) error
	UpdateBatch(ctx context.Context, batch *EmailBatch) error
	FindBatchByID(ctx context.Context, id string) (*EmailBatch, error)
	ListTemplates(ctx context.Context, category string, page, pageSize int) ([]*EmailTemplate, int32, error)
	FindTemplateByID(ctx context.Context, id string) (*EmailTemplate, error)
}

// EmailUsecase is the email usecase
type EmailUsecase struct {
	repo   EmailRepo
	sender EmailSender
	logger *log.Helper
	config *SMTPConfig
}

// NewEmailUsecase creates a new email usecase
func NewEmailUsecase(repo EmailRepo, sender EmailSender, config *SMTPConfig, logger log.Logger) *EmailUsecase {
	return &EmailUsecase{
		repo:   repo,
		sender: sender,
		logger: log.NewHelper(logger),
		config: config,
	}
}

// SendEmail sends a single email
func (uc *EmailUsecase) SendEmail(ctx context.Context, email *Email, opts *SendOptions) (string, error) {
	uc.logger.Infof("Sending email to %s with subject: %s", email.To, email.Subject)

	// Generate message ID
	email.MessageID = generateMessageID()
	email.Status = EmailStatusPending
	email.CreatedAt = time.Now()
	email.UpdatedAt = time.Now()

	// Save to repository
	if err := uc.repo.Save(ctx, email); err != nil {
		return "", fmt.Errorf("failed to save email: %w", err)
	}

	// Send via sender
	if err := uc.sender.Send(ctx, email, opts); err != nil {
		email.Status = EmailStatusFailed
		email.ErrorMsg = err.Error()
		email.UpdatedAt = time.Now()
		if updateErr := uc.repo.Update(ctx, email); updateErr != nil {
			uc.logger.Errorf("Failed to persist failed email status for %s: %v", email.MessageID, updateErr)
		}
		return email.MessageID, fmt.Errorf("failed to send email: %w", err)
	}

	// Update status
	email.Status = EmailStatusSent
	now := time.Now()
	email.SentAt = &now
	email.UpdatedAt = time.Now()
	if err := uc.repo.Update(ctx, email); err != nil {
		uc.logger.Errorf("Failed to update email status: %v", err)
	}

	return email.MessageID, nil
}

// SendBatchEmail sends emails in batch
func (uc *EmailUsecase) SendBatchEmail(ctx context.Context, emails []*Email, opts *BatchOptions) (*EmailBatch, error) {
	uc.logger.Infof("Starting batch email send for %d emails", len(emails))

	if opts == nil {
		opts = &BatchOptions{
			BatchSize: 100,
			RateLimit: time.Second,
		}
	}

	// Initialize emails
	for _, email := range emails {
		email.MessageID = generateMessageID()
		email.Status = EmailStatusPending
		email.CreatedAt = time.Now()
		email.UpdatedAt = time.Now()
	}

	// Send batch
	batch, err := uc.sender.SendBatch(ctx, emails, opts)
	if err != nil {
		return nil, fmt.Errorf("batch send failed: %w", err)
	}

	// Save batch to repository
	if err := uc.repo.SaveBatch(ctx, batch); err != nil {
		uc.logger.Errorf("Failed to save batch: %v", err)
	}

	// Update individual emails
	for _, email := range emails {
		if err := uc.repo.Update(ctx, email); err != nil {
			uc.logger.Errorf("Failed to update email %s: %v", email.MessageID, err)
		}
	}

	return batch, nil
}

// GetEmailStatus gets the status of an email
func (uc *EmailUsecase) GetEmailStatus(ctx context.Context, messageID string) (*Email, error) {
	email, err := uc.repo.FindByMessageID(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to find email: %w", err)
	}
	return email, nil
}

// GetBatchStatus gets the status of a batch
func (uc *EmailUsecase) GetBatchStatus(ctx context.Context, batchID string) (*EmailBatch, error) {
	batch, err := uc.repo.FindBatchByID(ctx, batchID)
	if err != nil {
		return nil, fmt.Errorf("failed to find batch: %w", err)
	}
	return batch, nil
}

// ListTemplates lists email templates
func (uc *EmailUsecase) ListTemplates(ctx context.Context, category string, page, pageSize int) ([]*EmailTemplate, int32, error) {
	return uc.repo.ListTemplates(ctx, category, page, pageSize)
}

// ApplyTemplate applies a template to email content
func (uc *EmailUsecase) ApplyTemplate(ctx context.Context, templateID string, data map[string]string) (*EmailTemplate, error) {
	template, err := uc.repo.FindTemplateByID(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}

	// Simple variable substitution
	for key, value := range data {
		placeholder := fmt.Sprintf("{{%s}}", key)
		template.Content = replaceAll(template.Content, placeholder, value)
		template.Subject = replaceAll(template.Subject, placeholder, value)
	}

	return template, nil
}

// UpdateEmailStatus updates email status (for webhooks/callbacks)
func (uc *EmailUsecase) UpdateEmailStatus(ctx context.Context, messageID string, status EmailStatus, errorMsg string) error {
	email, err := uc.repo.FindByMessageID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("email not found: %w", err)
	}

	email.Status = status
	if errorMsg != "" {
		email.ErrorMsg = errorMsg
	}

	now := time.Now()
	switch status {
	case EmailStatusDelivered:
		email.DeliveredAt = &now
	case EmailStatusOpened:
		email.OpenedAt = &now
	}
	email.UpdatedAt = now

	return uc.repo.Update(ctx, email)
}

// generateMessageID generates a unique message ID
func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}

// replaceAll replaces all occurrences of old with new
func replaceAll(s, old, new string) string {
	// Simple string replacement
	result := ""
	start := 0
	for {
		idx := 0
		for i := start; i <= len(s)-len(old); i++ {
			if s[i:i+len(old)] == old {
				idx = i
				break
			}
		}
		if idx == 0 {
			result += s[start:]
			break
		}
		result += s[start:idx] + new
		start = idx + len(old)
	}
	return result
}
