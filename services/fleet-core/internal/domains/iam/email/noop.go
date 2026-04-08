package email

import (
	"context"

	"go.uber.org/zap"
)

// NoOpProvider logs emails instead of sending them (for development/testing)
type NoOpProvider struct {
	log *zap.Logger
}

// NewNoOpProvider creates a new no-op email provider
func NewNoOpProvider(log *zap.Logger) *NoOpProvider {
	return &NoOpProvider{log: log}
}

// Send logs the email instead of sending it
func (p *NoOpProvider) Send(ctx context.Context, to, subject, bodyHTML, bodyText string) error {
	p.log.Debug("[email:no-op] would send email",
		zap.String("to", to),
		zap.String("subject", subject),
	)
	return nil
}
