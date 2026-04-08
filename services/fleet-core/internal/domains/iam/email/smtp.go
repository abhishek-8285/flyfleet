package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"

	"github.com/flyfleet/fleet-core/internal/domains/iam/config"
)

// SMTPProvider implements EmailProvider using SMTP
type SMTPProvider struct {
	host     string
	port     int
	user     string
	password string
	from     string
	fromName string
	useTLS   bool
}

// NewSMTPProvider creates a new SMTP email provider
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

// Send sends an email via SMTP
func (p *SMTPProvider) Send(ctx context.Context, to, subject, bodyHTML, bodyText string) error {
	if p.host == "" {
		return fmt.Errorf("smtp host not configured")
	}

	addr := fmt.Sprintf("%s:%d", p.host, p.port)
	auth := smtp.PlainAuth("", p.user, p.password, p.host)

	msg := p.buildMessage(to, subject, bodyHTML, bodyText)

	if p.useTLS {
		tlsConfig := &tls.Config{ServerName: p.host}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("smtp tls dial: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, p.host)
		if err != nil {
			return fmt.Errorf("smtp new client: %w", err)
		}
		defer client.Close()

		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
		if err := client.Mail(p.from); err != nil {
			return fmt.Errorf("smtp mail from: %w", err)
		}
		if err := client.Rcpt(to); err != nil {
			return fmt.Errorf("smtp rcpt: %w", err)
		}
		w, err := client.Data()
		if err != nil {
			return fmt.Errorf("smtp data: %w", err)
		}
		if _, err := w.Write([]byte(msg)); err != nil {
			return fmt.Errorf("smtp write: %w", err)
		}
		if err := w.Close(); err != nil {
			return fmt.Errorf("smtp close: %w", err)
		}
	} else {
		if err := smtp.SendMail(addr, auth, p.from, []string{to}, []byte(msg)); err != nil {
			return fmt.Errorf("smtp send: %w", err)
		}
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

	msg += "--FLYFLEET_BOUNDARY\r\n"
	msg += "Content-Type: text/plain; charset=UTF-8\r\n\r\n"
	msg += bodyText + "\r\n\r\n"

	msg += "--FLYFLEET_BOUNDARY\r\n"
	msg += "Content-Type: text/html; charset=UTF-8\r\n\r\n"
	msg += bodyHTML + "\r\n\r\n"

	msg += "--FLYFLEET_BOUNDARY--\r\n"
	return msg
}
