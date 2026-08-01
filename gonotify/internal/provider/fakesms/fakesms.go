package fakesms

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gonotify/gonotify/internal/provider"
)

type Provider struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Provider {
	return &Provider{
		db: db,
	}
}

func (p *Provider) Name() string {
	return "fake_sms"
}

func (p *Provider) Channel() provider.Channel {
	return provider.ChannelSMS
}

func (p *Provider) Send(ctx context.Context, req provider.Request) (provider.Result, error) {
	if req.To.Phone == "" {
		return provider.Result{}, fmt.Errorf("recipient phone is required")
	}

	if req.Text == "" {
		return provider.Result{}, fmt.Errorf("sms body is required")
	}

	var messageID string

	err := p.db.QueryRow(
		ctx,
		`
		INSERT INTO fake_sms_outbox (
			tenant_id,
			notification_id,
			to_phone,
			body
		)
		VALUES (
			$1::uuid,
			NULLIF($2, '')::uuid,
			$3,
			$4
		)
		RETURNING id::text
		`,
		req.TenantID,
		req.NotificationID,
		req.To.Phone,
		req.Text,
	).Scan(&messageID)

	if err != nil {
		return provider.Result{}, fmt.Errorf("fake sms insert failed: %w", err)
	}

	return provider.Result{
		ProviderMessageID: messageID,
		Status:            "sent",
		Raw: map[string]any{
			"provider": "fake_sms",
		},
	}, nil
}

func (p *Provider) Health(ctx context.Context) error {
	return nil
}
