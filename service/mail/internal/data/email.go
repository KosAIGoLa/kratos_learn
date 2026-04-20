package data

import (
	"context"
	"fmt"
	"sync"
	"time"

	"mail/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

// emailRepo implements biz.EmailRepo
type emailRepo struct {
	data   *Data
	logger *log.Helper

	// In-memory storage (replace with database in production)
	emails        map[string]*biz.Email // key: email ID
	emailsByMsgID map[string]*biz.Email // key: message ID
	batches       map[string]*biz.EmailBatch
	templates     map[string]*biz.EmailTemplate
	mu            sync.RWMutex
}

// NewEmailRepo creates a new email repository
func NewEmailRepo(data *Data, logger log.Logger) biz.EmailRepo {
	repo := &emailRepo{
		data:          data,
		logger:        log.NewHelper(logger),
		emails:        make(map[string]*biz.Email),
		emailsByMsgID: make(map[string]*biz.Email),
		batches:       make(map[string]*biz.EmailBatch),
		templates:     make(map[string]*biz.EmailTemplate),
	}

	// Initialize with some default templates
	repo.initDefaultTemplates()

	return repo
}

// initDefaultTemplates initializes default email templates
func (r *emailRepo) initDefaultTemplates() {
	now := time.Now()
	r.templates["welcome"] = &biz.EmailTemplate{
		ID:          "welcome",
		Name:        "Welcome Email",
		Subject:     "Welcome to {{company}}!",
		Content:     "<h1>Welcome {{name}}!</h1><p>Thank you for joining {{company}}.</p>",
		ContentType: "text/html",
		Category:    "onboarding",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	r.templates["reset_password"] = &biz.EmailTemplate{
		ID:          "reset_password",
		Name:        "Password Reset",
		Subject:     "Reset your password",
		Content:     "<p>Click <a href='{{reset_link}}'>here</a> to reset your password.</p>",
		ContentType: "text/html",
		Category:    "security",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	r.templates["notification"] = &biz.EmailTemplate{
		ID:          "notification",
		Name:        "General Notification",
		Subject:     "{{subject}}",
		Content:     "<p>{{message}}</p>",
		ContentType: "text/html",
		Category:    "general",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Save saves an email
func (r *emailRepo) Save(ctx context.Context, email *biz.Email) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if email.ID == "" {
		email.ID = fmt.Sprintf("email_%d", time.Now().UnixNano())
	}

	r.emails[email.ID] = email
	r.emailsByMsgID[email.MessageID] = email

	r.logger.Infof("Saved email: %s (message_id: %s)", email.ID, email.MessageID)
	return nil
}

// Update updates an email
func (r *emailRepo) Update(ctx context.Context, email *biz.Email) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if existing, ok := r.emails[email.ID]; ok {
		// Update fields
		existing.Status = email.Status
		existing.ErrorMsg = email.ErrorMsg
		if email.SentAt != nil {
			existing.SentAt = email.SentAt
		}
		if email.DeliveredAt != nil {
			existing.DeliveredAt = email.DeliveredAt
		}
		if email.OpenedAt != nil {
			existing.OpenedAt = email.OpenedAt
		}
		existing.UpdatedAt = time.Now()
		r.logger.Infof("Updated email: %s (status: %s)", email.ID, email.Status)
		return nil
	}

	return fmt.Errorf("email not found: %s", email.ID)
}

// FindByID finds an email by ID
func (r *emailRepo) FindByID(ctx context.Context, id string) (*biz.Email, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if email, ok := r.emails[id]; ok {
		return email, nil
	}

	return nil, fmt.Errorf("email not found: %s", id)
}

// FindByMessageID finds an email by message ID
func (r *emailRepo) FindByMessageID(ctx context.Context, messageID string) (*biz.Email, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if email, ok := r.emailsByMsgID[messageID]; ok {
		return email, nil
	}

	return nil, fmt.Errorf("email not found with message_id: %s", messageID)
}

// SaveBatch saves a batch
func (r *emailRepo) SaveBatch(ctx context.Context, batch *biz.EmailBatch) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.batches[batch.ID] = batch
	r.logger.Infof("Saved batch: %s (total: %d)", batch.ID, batch.Total)
	return nil
}

// UpdateBatch updates a batch
func (r *emailRepo) UpdateBatch(ctx context.Context, batch *biz.EmailBatch) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if existing, ok := r.batches[batch.ID]; ok {
		existing.Status = batch.Status
		existing.Sent = batch.Sent
		existing.Failed = batch.Failed
		existing.Pending = batch.Pending
		if batch.CompletedAt != nil {
			existing.CompletedAt = batch.CompletedAt
		}
		r.logger.Infof("Updated batch: %s (status: %s, sent: %d, failed: %d)",
			batch.ID, batch.Status, batch.Sent, batch.Failed)
		return nil
	}

	return fmt.Errorf("batch not found: %s", batch.ID)
}

// FindBatchByID finds a batch by ID
func (r *emailRepo) FindBatchByID(ctx context.Context, id string) (*biz.EmailBatch, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if batch, ok := r.batches[id]; ok {
		return batch, nil
	}

	return nil, fmt.Errorf("batch not found: %s", id)
}

// ListTemplates lists templates with pagination
func (r *emailRepo) ListTemplates(ctx context.Context, category string, page, pageSize int) ([]*biz.EmailTemplate, int32, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*biz.EmailTemplate
	for _, template := range r.templates {
		if category == "" || template.Category == category {
			result = append(result, template)
		}
	}

	total := int32(len(result))

	// Pagination
	if page > 0 && pageSize > 0 {
		start := (page - 1) * pageSize
		end := start + pageSize
		if start > len(result) {
			return []*biz.EmailTemplate{}, total, nil
		}
		if end > len(result) {
			end = len(result)
		}
		result = result[start:end]
	}

	return result, total, nil
}

// FindTemplateByID finds a template by ID
func (r *emailRepo) FindTemplateByID(ctx context.Context, id string) (*biz.EmailTemplate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if template, ok := r.templates[id]; ok {
		return template, nil
	}

	return nil, fmt.Errorf("template not found: %s", id)
}
