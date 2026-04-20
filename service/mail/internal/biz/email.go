package biz

import (
	"time"
)

// EmailStatus represents the status of an email
type EmailStatus string

const (
	EmailStatusPending    EmailStatus = "pending"
	EmailStatusSent       EmailStatus = "sent"
	EmailStatusDelivered  EmailStatus = "delivered"
	EmailStatusOpened     EmailStatus = "opened"
	EmailStatusClicked    EmailStatus = "clicked"
	EmailStatusFailed     EmailStatus = "failed"
	EmailStatusBounced    EmailStatus = "bounced"
	EmailStatusProcessing EmailStatus = "processing"
)

// Email represents an email entity
type Email struct {
	ID          string
	MessageID   string
	To          string
	From        string
	FromName    string
	Subject     string
	Body        string
	ContentType string
	Status      EmailStatus
	ErrorMsg    string
	SentAt      *time.Time
	DeliveredAt *time.Time
	OpenedAt    *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Attachment represents an email attachment
type Attachment struct {
	Filename    string
	Content     []byte
	ContentType string
}

// BatchStatus represents the status of a batch operation
type BatchStatus string

const (
	BatchStatusProcessing BatchStatus = "processing"
	BatchStatusCompleted  BatchStatus = "completed"
	BatchStatusFailed     BatchStatus = "failed"
)

// EmailBatch represents a batch of emails
type EmailBatch struct {
	ID          string
	Status      BatchStatus
	Total       int32
	Sent        int32
	Failed      int32
	Pending     int32
	CreatedAt   time.Time
	CompletedAt *time.Time
}

// EmailTemplate represents an email template
type EmailTemplate struct {
	ID          string
	Name        string
	Subject     string
	Content     string
	ContentType string
	Category    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// SendOptions represents options for sending an email
type SendOptions struct {
	TemplateID   string
	TemplateData map[string]string
	From         string
	FromName     string
	Attachments  []*Attachment
}

// BatchOptions represents options for batch sending
type BatchOptions struct {
	TemplateID   string
	TemplateData map[string]string
	From         string
	FromName     string
	BatchSize    int32
	RateLimit    time.Duration
}
