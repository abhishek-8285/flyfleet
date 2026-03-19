package email

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"go.uber.org/zap"
)

// EmailService handles sending invoice emails
type EmailService struct {
	apiKey    string
	fromEmail string
	fromName  string
	logger    *zap.Logger
}

// EmailConfig holds email configuration
type EmailConfig struct {
	APIKey    string `env:"SENDGRID_API_KEY"`
	FromEmail string `env:"FROM_EMAIL" default:"billing@flyfleet.com"`
	FromName  string `env:"FROM_NAME" default:"FlyFleet Billing"`
}

// NewEmailService creates a new email service
func NewEmailService(config EmailConfig, logger *zap.Logger) *EmailService {
	// Try to get API key from environment if not provided
	if config.APIKey == "" {
		config.APIKey = os.Getenv("SENDGRID_API_KEY")
	}
	if config.FromEmail == "" {
		config.FromEmail = getEnvOrDefault("FROM_EMAIL", "billing@flyfleet.com")
	}
	if config.FromName == "" {
		config.FromName = getEnvOrDefault("FROM_NAME", "FlyFleet Billing")
	}

	return &EmailService{
		apiKey:    config.APIKey,
		fromEmail: config.FromEmail,
		fromName:  config.FromName,
		logger:    logger,
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// InvoiceEmail represents an invoice email
type InvoiceEmail struct {
	To            string
	ToName        string
	Subject       string
	InvoiceNumber string
	CustomerName  string
	Amount        float64
	DueDate       string
	PDFData       []byte
	PDFFilename   string
}

// SendInvoice sends an invoice email with PDF attachment
func (s *EmailService) SendInvoice(ctx context.Context, email InvoiceEmail) error {
	if s.apiKey == "" {
		s.logger.Warn("SendGrid API key not configured, skipping email",
			zap.String("to", email.To),
			zap.String("invoice", email.InvoiceNumber),
		)
		return nil // Don't fail, just log warning
	}

	from := mail.NewEmail(s.fromName, s.fromEmail)
	to := mail.NewEmail(email.CustomerName, email.To)

	// Create HTML content
	htmlContent := s.createInvoiceEmailHTML(email)

	content := mail.NewContent("text/html", htmlContent)

	// Create email
	m := mail.NewV3MailInit(from, email.Subject, to, content)

	// Add CC if needed (accounting team)
	if ccEmail := os.Getenv("INVOICE_CC_EMAIL"); ccEmail != "" {
		cc := mail.NewEmail("FlyFleet Accounting", ccEmail)
		if len(m.Personalizations) > 0 {
			m.Personalizations[0].AddCCs(cc)
		}
	}

	// Attach PDF
	if len(email.PDFData) > 0 {
		attachment := mail.NewAttachment()
		attachment.SetContent(base64.StdEncoding.EncodeToString(email.PDFData))
		attachment.SetFilename(email.PDFFilename)
		attachment.SetType("application/pdf")
		attachment.SetDisposition("attachment")
		m.AddAttachment(attachment)
	}

	// Add tracking settings
	trackingSettings := mail.NewTrackingSettings()
	clickTracking := mail.NewClickTrackingSetting()
	clickTracking.SetEnable(true)
	trackingSettings.SetClickTracking(clickTracking)
	openTracking := mail.NewOpenTrackingSetting()
	openTracking.SetEnable(true)
	trackingSettings.SetOpenTracking(openTracking)
	m.SetTrackingSettings(trackingSettings)

	// Send email
	request := sendgrid.GetRequest(
		s.apiKey,
		"/v3/mail/send",
		"https://api.sendgrid.com",
	)
	request.Method = "POST"
	request.Body = mail.GetRequestBody(m)

	response, err := sendgrid.API(request)
	if err != nil {
		s.logger.Error("Failed to send invoice email",
			zap.String("to", email.To),
			zap.String("invoice", email.InvoiceNumber),
			zap.Error(err),
		)
		return fmt.Errorf("failed to send email: %w", err)
	}

	// Check response status code
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		s.logger.Error("SendGrid API error",
			zap.Int("status_code", response.StatusCode),
			zap.String("body", response.Body),
			zap.String("invoice", email.InvoiceNumber),
		)
		return fmt.Errorf("sendgrid error: status %d, body: %s", response.StatusCode, response.Body)
	}

	s.logger.Info("Invoice email sent successfully",
		zap.String("to", email.To),
		zap.String("invoice", email.InvoiceNumber),
		zap.Float64("amount", email.Amount),
	)

	return nil
}

// createInvoiceEmailHTML creates the HTML email body
func (s *EmailService) createInvoiceEmailHTML(email InvoiceEmail) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            line-height: 1.6;
            color: #1a1a1a;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }
        .header {
            background: linear-gradient(135deg, #3b82f6 0%%, #2563eb 100%%);
            color: white;
            padding: 30px;
            border-radius: 12px 12px 0 0;
            text-align: center;
        }
        .logo {
            font-size: 28px;
            font-weight: bold;
            margin-bottom: 10px;
        }
        .content {
            background: #ffffff;
            padding: 30px;
            border: 1px solid #e2e8f0;
            border-top: none;
            border-radius: 0 0 12px 12px;
        }
        .invoice-box {
            background: #f8fafc;
            border: 1px solid #e2e8f0;
            border-radius: 8px;
            padding: 20px;
            margin: 20px 0;
        }
        .invoice-number {
            font-size: 24px;
            font-weight: bold;
            color: #3b82f6;
            margin-bottom: 10px;
        }
        .amount {
            font-size: 32px;
            font-weight: bold;
            color: #10b981;
            margin: 15px 0;
        }
        .due-date {
            color: #ef4444;
            font-weight: 600;
        }
        .button {
            display: inline-block;
            background: #3b82f6;
            color: white;
            padding: 12px 30px;
            text-decoration: none;
            border-radius: 8px;
            font-weight: 600;
            margin: 20px 0;
        }
        .button:hover {
            background: #2563eb;
        }
        .footer {
            text-align: center;
            color: #64748b;
            font-size: 14px;
            margin-top: 30px;
            padding-top: 20px;
            border-top: 1px solid #e2e8f0;
        }
        .info-row {
            display: flex;
            justify-content: space-between;
            margin: 10px 0;
            padding: 10px 0;
            border-bottom: 1px solid #e2e8f0;
        }
        .info-label {
            color: #64748b;
            font-weight: 500;
        }
        .info-value {
            color: #1a1a1a;
            font-weight: 600;
        }
    </style>
</head>
<body>
    <div class="header">
        <div class="logo">🚀 FlyFleet</div>
        <div>Invoice Notification</div>
    </div>
    
    <div class="content">
        <p>Dear %s,</p>
        
        <p>Your invoice is ready. Please find the details below and the PDF invoice attached to this email.</p>
        
        <div class="invoice-box">
            <div class="invoice-number">📄 %s</div>
            
            <div class="info-row">
                <span class="info-label">Amount Due</span>
                <span class="amount">₹%.2f</span>
            </div>
            
            <div class="info-row">
                <span class="info-label">Due Date</span>
                <span class="due-date">%s</span>
            </div>
            
            <div class="info-row">
                <span class="info-label">Status</span>
                <span class="info-value">⏳ Pending Payment</span>
            </div>
        </div>
        
        <div style="text-align: center;">
            <a href="#" class="button">View Invoice</a>
        </div>
        
        <p style="margin-top: 20px;"><strong>Next Steps:</strong></p>
        <ol style="color: #64748b;">
            <li>Review the attached PDF invoice</li>
            <li>Click "View Invoice" to see details online</li>
            <li>Complete payment before the due date</li>
        </ol>
        
        <p style="margin-top: 20px;">If you have any questions about this invoice, please reply to this email or contact our support team.</p>
        
        <p>Thank you for your business!</p>
        
        <div class="footer">
            <p><strong>FlyFleet Technologies</strong></p>
            <p>123 Fleet Street, Bangalore, Karnataka 560001</p>
            <p>Email: billing@flyfleet.com | Phone: +91-80-1234-5678</p>
            <p style="margin-top: 15px; font-size: 12px;">
                This is an automated message. Please do not reply directly to this email.
            </p>
        </div>
    </div>
</body>
</html>
    `,
		email.CustomerName,
		email.InvoiceNumber,
		email.Amount,
		email.DueDate,
	)
}

// SendTestEmail sends a test email
func (s *EmailService) SendTestEmail(ctx context.Context, to string) error {
	email := InvoiceEmail{
		To:            to,
		ToName:        "Test User",
		Subject:       "FlyFleet Invoice System - Test Email",
		InvoiceNumber: "TEST-001",
		CustomerName:  "Test User",
		Amount:        0,
		DueDate:       "N/A",
		PDFData:       nil,
		PDFFilename:   "",
	}

	return s.SendInvoice(ctx, email)
}

// HealthCheck checks if email service is configured
func (s *EmailService) HealthCheck() error {
	if s.apiKey == "" {
		return fmt.Errorf("SendGrid API key not configured")
	}

	// Simple validation
	if !strings.Contains(s.fromEmail, "@") {
		return fmt.Errorf("invalid from email: %s", s.fromEmail)
	}

	return nil
}

// GetConfig returns current configuration (for debugging)
func (s *EmailService) GetConfig() map[string]string {
	apiKeyPreview := ""
	if s.apiKey != "" {
		apiKeyPreview = s.apiKey[:5] + "..." + s.apiKey[len(s.apiKey)-3:]
	}

	return map[string]string{
		"api_key_configured": apiKeyPreview,
		"from_email":         s.fromEmail,
		"from_name":          s.fromName,
	}
}
