package reports

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DeliveryMethod defines how completed reports are sent to the requester.
type DeliveryMethod string

const (
	DeliveryDownloadLink DeliveryMethod = "DOWNLOAD_LINK"
	DeliveryEmail        DeliveryMethod = "EMAIL"
	DeliveryWebhook      DeliveryMethod = "WEBHOOK"
)

// DeliveryOptions holds config for each delivery channel.
type DeliveryOptions struct {
	Method       DeliveryMethod `json:"method"`
	EmailAddress string         `json:"email_address,omitempty"`
	WebhookURL   string         `json:"webhook_url,omitempty"`
}

// DeliveryService dispatches completed report notifications.
type DeliveryService struct {
	logger *zap.Logger
}

func NewDeliveryService(logger *zap.Logger) *DeliveryService {
	return &DeliveryService{logger: logger}
}

// Deliver sends a notification for a completed report.
func (d *DeliveryService) Deliver(ctx context.Context, reportID uuid.UUID, fileURL string, opts *DeliveryOptions) error {
	if opts == nil {
		return nil // No delivery config, download-link is implicit
	}

	switch opts.Method {
	case DeliveryEmail:
		return d.sendEmail(ctx, reportID, fileURL, opts.EmailAddress)
	case DeliveryWebhook:
		return d.sendWebhook(ctx, reportID, fileURL, opts.WebhookURL)
	default:
		d.logger.Info("Report available for download", zap.String("url", fileURL))
		return nil
	}
}

func (d *DeliveryService) sendEmail(_ context.Context, reportID uuid.UUID, fileURL, email string) error {
	if email == "" {
		return fmt.Errorf("email address is empty")
	}
	// In production, integrate with SendGrid or SES here.
	d.logger.Info("Sending report via email",
		zap.String("report_id", reportID.String()),
		zap.String("email", email),
		zap.String("file_url", fileURL))
	return nil
}

func (d *DeliveryService) sendWebhook(_ context.Context, reportID uuid.UUID, fileURL, webhookURL string) error {
	if webhookURL == "" {
		return fmt.Errorf("webhook URL is empty")
	}
	// In production, make an HTTP POST with report metadata + download link.
	d.logger.Info("Sending report via webhook",
		zap.String("report_id", reportID.String()),
		zap.String("webhook_url", webhookURL),
		zap.String("file_url", fileURL),
		zap.Time("delivered_at", time.Now()))
	return nil
}
