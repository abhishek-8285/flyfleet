package service

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/smtp"

	"github.com/flyfleet/fleet-core/internal/domains/iam/config"
	"go.uber.org/zap"
)

// EmailProvider defines the interface for sending emails.
type EmailProvider interface {
	Send(ctx context.Context, to, subject, bodyHTML, bodyText string) error
}

// EmailSender wraps an EmailProvider with FlyFleet-specific email templates.
type EmailSender struct {
	provider EmailProvider
	from     string
	fromName string
	baseURL  string
	log      *zap.Logger
}

func NewEmailSender(provider EmailProvider, cfg config.EmailConfig, log *zap.Logger) *EmailSender {
	return &EmailSender{
		provider: provider,
		from:     cfg.FromAddress,
		fromName: cfg.FromName,
		baseURL:  "https://app.flyfleet.io",
		log:      log,
	}
}

// SendPasswordReset sends a password reset email with a link containing the token.
func (s *EmailSender) SendPasswordReset(ctx context.Context, to, token string) error {
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

// SendInvitation sends an invitation email to join an organization.
func (s *EmailSender) SendInvitation(ctx context.Context, to, token, message, orgName, inviterName string) error {
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

// SendAccountLocked sends a notification that the account has been locked.
func (s *EmailSender) SendAccountLocked(ctx context.Context, to string) error {
	subject := "Your FlyFleet Account Has Been Locked"

	bodyHTML := fmt.Sprintf(`<!DOCTYPE html><html><body style="font-family:system-ui,sans-serif;max-width:480px;margin:0 auto;padding:24px;">
<h2 style="color:#d32f2f;">Account Locked</h2>
<p>Your FlyFleet account has been temporarily locked due to multiple failed login attempts.</p>
<p>For security reasons, your account will be automatically unlocked in 15 minutes. If you did not attempt to log in, please contact your organization administrator immediately.</p>
<p style="color:#666;font-size:14px;">If you need immediate assistance, reply to this email.</p>
</body></html>`)

	bodyText := "Your FlyFleet account has been temporarily locked due to multiple failed login attempts. It will be automatically unlocked in 15 minutes."

	return s.provider.Send(ctx, to, subject, bodyHTML, bodyText)
}

// ── SMTP Provider ──

type SMTPProvider struct {
	host     string
	port     int
	user     string
	password string
	from     string
	fromName string
	useTLS   bool
}

func NewSMTPProvider(cfg config.EmailConfig) *SMTPProvider {
	return &SMTPProvider{
		host:     cfg.SMTPHost,
		port:     cfg.SMTPPort,
		user:     cfg.SMTPUser,
		password: cfg.SMTPPassword,
		from:     cfg.FromAddress,
		fromName: cfg.FromName,
		useTLS:   true,
	}
}

func (p *SMTPProvider) Send(ctx context.Context, to, subject, bodyHTML, bodyText string) error {
	if p.host == "" {
		return errors.New("smtp host not configured")
	}

	addr := fmt.Sprintf("%s:%d", p.host, p.port)
	auth := smtp.PlainAuth("", p.user, p.password, p.host)

	msg := p.buildMessage(to, subject, bodyHTML, bodyText)

	var err error
	if p.useTLS {
		tlsConfig := &tls.Config{
			ServerName: p.host,
		}
		conn, dialErr := tls.Dial("tcp", addr, tlsConfig)
		if dialErr != nil {
			return fmt.Errorf("smtp tls dial: %w", dialErr)
		}
		client, clientErr := smtp.NewClient(conn, p.host)
		if clientErr != nil {
			return fmt.Errorf("smtp new client: %w", clientErr)
		}
		defer client.Close()

		if authErr := client.Auth(auth); authErr != nil {
			return fmt.Errorf("smtp auth: %w", authErr)
		}
		if fromErr := client.Mail(p.from); fromErr != nil {
			return fmt.Errorf("smtp mail from: %w", fromErr)
		}
		if rcptErr := client.Rcpt(to); rcptErr != nil {
			return fmt.Errorf("smtp rcpt to: %w", rcptErr)
		}
		w, writeErr := client.Data()
		if writeErr != nil {
			return fmt.Errorf("smtp data: %w", writeErr)
		}
		_, writeErr = w.Write([]byte(msg))
		if writeErr != nil {
			return fmt.Errorf("smtp write: %w", writeErr)
		}
		err = w.Close()
	} else {
		err = smtp.SendMail(addr, auth, p.from, []string{to}, []byte(msg))
	}

	if err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	return nil
}

func (p *SMTPProvider) buildMessage(to, subject, bodyHTML, bodyText string) string {
	msg := fmt.Sprintf(
		"From: %s <%s>\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: multipart/alternative; boundary=\"FLYFLEET_BOUNDARY\"\r\n\r\n",
		p.fromName, p.from, to, subject,
	)

	// Plain text part
	msg += "--FLYFLEET_BOUNDARY\r\n"
	msg += "Content-Type: text/plain; charset=UTF-8\r\n\r\n"
	msg += bodyText + "\r\n\r\n"

	// HTML part
	msg += "--FLYFLEET_BOUNDARY\r\n"
	msg += "Content-Type: text/html; charset=UTF-8\r\n\r\n"
	msg += bodyHTML + "\r\n\r\n"

	msg += "--FLYFLEET_BOUNDARY--\r\n"
	return msg
}

// ── NoOp Provider (for development / testing) ──

type NoOpEmailProvider struct {
	log *zap.Logger
}

func NewNoOpEmailProvider(log *zap.Logger) *NoOpEmailProvider {
	return &NoOpEmailProvider{log: log}
}

func (p *NoOpEmailProvider) Send(ctx context.Context, to, subject, bodyHTML, bodyText string) error {
	p.log.Debug("[email:no-op] would send email",
		zap.String("to", to),
		zap.String("subject", subject),
	)
	return nil
}
