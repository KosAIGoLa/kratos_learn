package biz

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	"net/smtp"
	"net/textproto"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// EmailSender is the interface for sending emails
type EmailSender interface {
	Send(ctx context.Context, email *Email, opts *SendOptions) error
	SendBatch(ctx context.Context, emails []*Email, opts *BatchOptions) (*EmailBatch, error)
}

// SMTPConfig holds SMTP configuration
type SMTPConfig struct {
	Host            string
	Port            int
	Username        string
	Password        string
	UseTLS          bool
	UseSSL          bool
	Timeout         time.Duration
	DefaultFrom     string
	DefaultFromName string
	MaxRetries      int
	RateLimitPerSec int
}

// SMTPSender implements EmailSender using SMTP
type SMTPSender struct {
	config *SMTPConfig
	logger *log.Helper
}

// NewSMTPSender creates a new SMTP sender
func NewSMTPSender(config *SMTPConfig, logger log.Logger) *SMTPSender {
	return &SMTPSender{
		config: config,
		logger: log.NewHelper(logger),
	}
}

// Send sends a single email via SMTP
func (s *SMTPSender) Send(ctx context.Context, email *Email, opts *SendOptions) error {
	if opts == nil {
		opts = &SendOptions{}
	}

	// Apply template if provided
	if opts.TemplateID != "" {
		// Template will be loaded and applied
		// For now, just log it
		s.logger.Infof("Applying template %s", opts.TemplateID)
	}

	// Set sender
	from := opts.From
	if from == "" {
		from = s.config.DefaultFrom
	}
	fromName := opts.FromName
	if fromName == "" {
		fromName = s.config.DefaultFromName
	}

	// Build email content
	content, err := s.buildEmailContent(email, from, fromName, opts.Attachments)
	if err != nil {
		return fmt.Errorf("failed to build email content: %w", err)
	}

	// Send via SMTP with retry
	return s.sendWithRetry(ctx, email.To, content)
}

// SendBatch sends multiple emails in batch
func (s *SMTPSender) SendBatch(ctx context.Context, emails []*Email, opts *BatchOptions) (*EmailBatch, error) {
	if opts == nil {
		opts = &BatchOptions{
			BatchSize: 100,
			RateLimit: time.Second,
		}
	}

	batch := &EmailBatch{
		ID:        generateBatchID(),
		Status:    BatchStatusProcessing,
		Total:     int32(len(emails)),
		Pending:   int32(len(emails)),
		CreatedAt: time.Now(),
	}

	// Process in batches
	batchSize := int(opts.BatchSize)
	if batchSize <= 0 {
		batchSize = 100
	}

	rateLimiter := time.NewTicker(opts.RateLimit)
	defer rateLimiter.Stop()

	for i := 0; i < len(emails); i += batchSize {
		end := i + batchSize
		if end > len(emails) {
			end = len(emails)
		}

		batchEmails := emails[i:end]
		for _, email := range batchEmails {
			select {
			case <-ctx.Done():
				batch.Status = BatchStatusFailed
				return batch, ctx.Err()
			case <-rateLimiter.C:
				opts := &SendOptions{
					TemplateID:   opts.TemplateID,
					TemplateData: opts.TemplateData,
					From:         opts.From,
					FromName:     opts.FromName,
				}

				if err := s.Send(ctx, email, opts); err != nil {
					s.logger.Errorf("Failed to send email to %s: %v", email.To, err)
					email.Status = EmailStatusFailed
					email.ErrorMsg = err.Error()
					batch.Failed++
				} else {
					email.Status = EmailStatusSent
					batch.Sent++
				}
				batch.Pending--
			}
		}
	}

	now := time.Now()
	batch.CompletedAt = &now
	if batch.Failed == batch.Total {
		batch.Status = BatchStatusFailed
	} else {
		batch.Status = BatchStatusCompleted
	}

	return batch, nil
}

// buildEmailContent builds the email content with MIME format
func (s *SMTPSender) buildEmailContent(email *Email, from, fromName string, attachments []*Attachment) ([]byte, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Headers
	headers := textproto.MIMEHeader{}
	if fromName != "" {
		headers.Set("From", fmt.Sprintf("%s <%s>", fromName, from))
	} else {
		headers.Set("From", from)
	}
	headers.Set("To", email.To)
	headers.Set("Subject", encodeSubject(email.Subject))
	headers.Set("MIME-Version", "1.0")
	headers.Set("Content-Type", fmt.Sprintf("multipart/mixed; boundary=%s", writer.Boundary()))
	headers.Set("Date", time.Now().Format(time.RFC1123))

	// Write headers
	for key, values := range headers {
		for _, value := range values {
			fmt.Fprintf(&buf, "%s: %s\r\n", key, value)
		}
	}
	fmt.Fprintf(&buf, "\r\n")

	// Body part
	bodyPart, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Type":              {fmt.Sprintf("%s; charset=UTF-8", email.ContentType)},
		"Content-Transfer-Encoding": {"quoted-printable"},
	})
	if err != nil {
		return nil, err
	}

	qpWriter := quotedprintable.NewWriter(bodyPart)
	_, err = qpWriter.Write([]byte(email.Body))
	if err != nil {
		return nil, err
	}
	qpWriter.Close()

	// Attachments
	for _, att := range attachments {
		header := textproto.MIMEHeader{
			"Content-Type":              {fmt.Sprintf("%s; name=%q", att.ContentType, att.Filename)},
			"Content-Disposition":       {fmt.Sprintf("attachment; filename=%q", att.Filename)},
			"Content-Transfer-Encoding": {"base64"},
		}
		part, err := writer.CreatePart(header)
		if err != nil {
			return nil, err
		}
		_, err = part.Write(att.Content)
		if err != nil {
			return nil, err
		}
	}

	writer.Close()
	return buf.Bytes(), nil
}

// sendWithRetry sends email with retry logic
func (s *SMTPSender) sendWithRetry(ctx context.Context, to string, content []byte) error {
	maxRetries := s.config.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			time.Sleep(time.Duration(i) * time.Second)
		}

		lastErr = s.sendSMTP(to, content)
		if lastErr == nil {
			return nil
		}

		s.logger.Warnf("Send attempt %d failed: %v", i+1, lastErr)
	}

	return fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

// sendSMTP performs the actual SMTP sending
func (s *SMTPSender) sendSMTP(to string, content []byte) error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	// Connect
	conn, err := net.DialTimeout("tcp", addr, s.config.Timeout)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	// TLS
	if s.config.UseTLS {
		tlsConfig := &tls.Config{
			ServerName: s.config.Host,
		}
		tlsConn := tls.Client(conn, tlsConfig)
		err = tlsConn.Handshake()
		if err != nil {
			return fmt.Errorf("TLS handshake failed: %w", err)
		}
		conn = tlsConn
	}

	// SMTP client
	client, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Auth
	if s.config.Username != "" && s.config.Password != "" {
		auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP auth failed: %w", err)
		}
	}

	// Mail and RCPT
	from := s.config.DefaultFrom
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("RCPT TO failed: %w", err)
	}

	// Data
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA failed: %w", err)
	}
	_, err = w.Write(content)
	if err != nil {
		return fmt.Errorf("failed to write content: %w", err)
	}
	w.Close()

	return client.Quit()
}

// encodeSubject encodes the subject using MIME encoding
func encodeSubject(subject string) string {
	needsEncoding := false
	for _, r := range subject {
		if r > 127 {
			needsEncoding = true
			break
		}
	}

	if !needsEncoding {
		return subject
	}

	var buf bytes.Buffer
	qpWriter := quotedprintable.NewWriter(&buf)
	qpWriter.Write([]byte(subject))
	qpWriter.Close()
	return fmt.Sprintf("=?UTF-8?q?%s?=", strings.ReplaceAll(buf.String(), " ", "_"))
}

// generateBatchID generates a unique batch ID
func generateBatchID() string {
	return fmt.Sprintf("batch_%d", time.Now().UnixNano())
}
