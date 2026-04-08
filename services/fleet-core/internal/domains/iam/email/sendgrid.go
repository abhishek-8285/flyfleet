package email

import (
	"context"
	"fmt"

	"github.com/flyfleet/fleet-core/internal/domains/iam/config"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"go.uber.org/zap"
)

// SendGridProvider implements EmailProvider using SendGrid API
type SendGridProvider struct {
	client   *sendgrid.Client
	from     mail.Email
	fromName string
	log      *zap.Logger
}

// NewSendGridProvider creates a new SendGrid email provider
func NewSendGridProvider(cfg config.EmailConfig, log *zap.Logger) *SendGridProvider {
	return &SendGridProvider{
		client:   sendgrid.NewSendClient(cfg.SendGridAPIKey),
		from:     mail.Email{Address: cfg.FromAddress, Name: cfg.FromName},
		fromName: cfg.FromName,
		log:      log,
	}
}

// Send sends an email via SendGrid
func (p *SendGridProvider) Send(ctx context.Context, to, subject, bodyHTML, bodyText string) error {
	toEmail := mail.Email{Address: to}

	message := mail.NewSingleEmail(&p.from, subject, &toEmail, bodyText, bodyHTML)

	response, err := p.client.SendWithContext(ctx, message)
	if err != nil {
		return fmt.Errorf("sendgrid send: %w", err)
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("sendgrid error: status=%d body=%s", response.StatusCode, response.Body)
	}

	p.log.Debug("email sent via sendgrid", zap.String("to", to), zap.String("subject", subject))
	return nil
}
