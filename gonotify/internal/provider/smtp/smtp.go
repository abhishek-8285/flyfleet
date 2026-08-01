package smtp

import (
	"bytes"
	"context"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"time"

	"github.com/gonotify/gonotify/internal/provider"
)

type Config struct {
	Host      string
	Port      int
	Username  string
	Password  string
	FromEmail string
	FromName  string
	UseTLS    bool
}

type Provider struct {
	cfg Config
}

func New(cfg Config) *Provider {
	return &Provider{
		cfg: cfg,
	}
}

func (p *Provider) Name() string {
	return "smtp"
}

func (p *Provider) Channel() provider.Channel {
	return provider.ChannelEmail
}

func (p *Provider) Send(ctx context.Context, req provider.Request) (provider.Result, error) {
	if req.To.Email == "" {
		return provider.Result{}, fmt.Errorf("recipient email is required")
	}

	fromAddress := req.From
	if fromAddress == "" {
		fromAddress = p.cfg.FromEmail
	}

	fromName := req.FromName
	if fromName == "" {
		fromName = p.cfg.FromName
	}

	from := mail.Address{
		Name:    fromName,
		Address: fromAddress,
	}

	to := mail.Address{
		Address: req.To.Email,
	}

	subject := mime.QEncoding.Encode("utf-8", req.Subject)
	message := buildMIME(from.String(), to.Address, subject, req.Text, req.HTML)

	addr := fmt.Sprintf("%s:%d", p.cfg.Host, p.cfg.Port)

	var auth smtp.Auth

	if p.cfg.Username != "" && p.cfg.Password != "" {
		auth = smtp.PlainAuth(
			"",
			p.cfg.Username,
			p.cfg.Password,
			p.cfg.Host,
		)
	}

	// TODO: replace with a context-aware SMTP client for production.
	// net/smtp.SendMail does not accept context directly.
	if err := smtp.SendMail(
		addr,
		auth,
		from.Address,
		[]string{to.Address},
		message,
	); err != nil {
		return provider.Result{}, fmt.Errorf("smtp send failed: %w", err)
	}

	return provider.Result{
		Status: "sent",
	}, nil
}

func (p *Provider) Health(ctx context.Context) error {
	dialer := net.Dialer{
		Timeout: 3 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", p.cfg.Host, p.cfg.Port)

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("smtp health check failed: %w", err)
	}

	return conn.Close()
}

func buildMIME(
	from string,
	to string,
	subject string,
	textBody string,
	htmlBody string,
) []byte {
	var buf bytes.Buffer

	boundary := fmt.Sprintf("gonotify_%d", time.Now().UnixNano())

	fmt.Fprintf(&buf, "From: %s\r\n", from)
	fmt.Fprintf(&buf, "To: %s\r\n", to)
	fmt.Fprintf(&buf, "Subject: %s\r\n", subject)
	fmt.Fprintf(&buf, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&buf, "Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary)
	fmt.Fprintf(&buf, "\r\n")

	if textBody != "" {
		fmt.Fprintf(&buf, "--%s\r\n", boundary)
		fmt.Fprintf(&buf, "Content-Type: text/plain; charset=UTF-8\r\n")
		fmt.Fprintf(&buf, "\r\n")
		fmt.Fprintf(&buf, "%s\r\n", textBody)
	}

	if htmlBody != "" {
		fmt.Fprintf(&buf, "--%s\r\n", boundary)
		fmt.Fprintf(&buf, "Content-Type: text/html; charset=UTF-8\r\n")
		fmt.Fprintf(&buf, "\r\n")
		fmt.Fprintf(&buf, "%s\r\n", htmlBody)
	}

	fmt.Fprintf(&buf, "--%s--\r\n", boundary)

	return buf.Bytes()
}
