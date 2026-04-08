package email

import (
	"context"
	"fmt"

	"github.com/flyfleet/fleet-core/internal/domains/iam/config"
	"go.uber.org/zap"
)

// Provider interface for sending emails
type Provider interface {
	Send(ctx context.Context, to, subject, bodyHTML, bodyText string) error
}

// EmailService handles sending emails with the configured provider
type EmailService struct {
	provider Provider
	from     string
	fromName string
	baseURL  string
	log      *zap.Logger
}

// NewEmailService creates an email service with the appropriate provider based on config
func NewEmailService(cfg config.EmailConfig, log *zap.Logger) *EmailService {
	var provider Provider

	switch cfg.Provider {
	case "sendgrid":
		if cfg.SendGridAPIKey != "" {
			provider = NewSendGridProvider(cfg, log)
			log.Info("using sendgrid email provider")
		} else {
			log.Warn("sendgrid configured but no API key, falling back to no-op")
			provider = NewNoOpProvider(log)
		}
	case "smtp":
		if cfg.SMTPHost != "" {
			provider = NewSMTPProvider(cfg)
			log.Info("using smtp email provider")
		} else {
			log.Warn("smtp configured but no host, falling back to no-op")
			provider = NewNoOpProvider(log)
		}
	default:
		log.Warn("unknown email provider, using no-op", zap.String("provider", cfg.Provider))
		provider = NewNoOpProvider(log)
	}

	return &EmailService{
		provider: provider,
		from:     cfg.FromAddress,
		fromName: cfg.FromName,
		baseURL:  "https://app.flyfleet.io",
		log:      log,
	}
}

// SendInvitation sends an invitation email
func (s *EmailService) SendInvitation(ctx context.Context, to, token, message, orgName, inviterName string) error {
	subject := fmt.Sprintf("You're invited to join %s on FlyFleet", orgName)
	acceptURL := fmt.Sprintf("%s/auth/invite?token=%s", s.baseURL, token)

	messageHTML := ""
	if message != "" {
		messageHTML = fmt.Sprintf("<p style=\"background:#f4f4f8;padding:12px;border-radius:6px;font-style:italic;\">%s</p>", message)
	}

	bodyHTML := fmt.Sprintf(`<!DOCTYPE html><html><body style="font-family:system-ui,sans-serif;max-width:480px;margin:0 auto;padding:24px;">
<h2 style="color:#1a1a2e;">Join %s on FlyFleet</h2>
<p><strong>%s</strong> has invited you to join their fleet management workspace.</p>
%s
<p><a href="%s" style="display:inline-block;background:#4361ee;color:#fff;padding:12px 24px;text-decoration:none;border-radius:6px;">Accept Invitation</a></p>
<p style="color:#666;font-size:14px;">This invitation expires in 7 days.</p>
</body></html>`, orgName, inviterName, messageHTML, acceptURL)

	bodyText := fmt.Sprintf("%s has invited you to join %s on FlyFleet.\n\nAccept the invitation here: %s\n\nThis invitation expires in 7 days.", inviterName, orgName, acceptURL)

	return s.provider.Send(ctx, to, subject, bodyHTML, bodyText)
}

// SendPasswordReset sends a password reset email
func (s *EmailService) SendPasswordReset(ctx context.Context, to, token string) error {
	subject := "Reset Your FlyFleet Password"
	resetURL := fmt.Sprintf("%s/auth/reset?token=%s", s.baseURL, token)

	bodyHTML := fmt.Sprintf(`<!DOCTYPE html><html><body style="font-family:system-ui,sans-serif;max-width:480px;margin:0 auto;padding:24px;">
<h2 style="color:#1a1a2e;">Reset Your Password</h2>
<p>You requested a password reset for your FlyFleet account. Click the button below to set a new password:</p>
<p><a href="%s" style="display:inline-block;background:#4361ee;color:#fff;padding:12px 24px;text-decoration:none;border-radius:6px;">Reset Password</a></p>
<p style="color:#666;font-size:14px;">This link expires in 30 minutes. If you didn't request this, you can safely ignore this email.</p>
</body></html>`, resetURL)

	bodyText := fmt.Sprintf("Reset your FlyFleet password by visiting: %s\n\nThis link expires in 30 minutes.", resetURL)

	return s.provider.Send(ctx, to, subject, bodyHTML, bodyText)
}

// SendAccountLocked sends a notification that the account has been locked
func (s *EmailService) SendAccountLocked(ctx context.Context, to string) error {
	subject := "Your FlyFleet Account Has Been Locked"

	bodyHTML := `<!DOCTYPE html><html><body style="font-family:system-ui,sans-serif;max-width:480px;margin:0 auto;padding:24px;">
<h2 style="color:#d32f2f;">Account Locked</h2>
<p>Your FlyFleet account has been temporarily locked due to multiple failed login attempts.</p>
<p>For security reasons, your account will be automatically unlocked in 15 minutes. If you did not attempt to log in, please contact your organization administrator immediately.</p>
<p style="color:#666;font-size:14px;">If you need immediate assistance, reply to this email.</p>
</body></html>`

	bodyText := "Your FlyFleet account has been temporarily locked due to multiple failed login attempts. It will be automatically unlocked in 15 minutes."

	return s.provider.Send(ctx, to, subject, bodyHTML, bodyText)
}
