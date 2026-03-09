package providers

import (
	"context"
	"fmt"

	"github.com/flyfleet/fleet-core-service/internal/domains/alerts_v2/models"
)

type ProviderClient interface {
	Send(ctx context.Context, log *models.NotificationLog) (providerResponse map[string]any, err error)
	Name() string
}

type ProviderFactory interface {
	GetClient(channel models.Channel) (ProviderClient, error)
}

type providerFactory struct {
	twilio   ProviderClient
	sendgrid ProviderClient
	firebase ProviderClient
	webhook  ProviderClient
}

func NewProviderFactory(twilio, sendgrid, firebase, webhook ProviderClient) ProviderFactory {
	return &providerFactory{
		twilio:   twilio,
		sendgrid: sendgrid,
		firebase: firebase,
		webhook:  webhook,
	}
}

func (f *providerFactory) GetClient(channel models.Channel) (ProviderClient, error) {
	switch channel {
	case models.ChannelEmail:
		return f.sendgrid, nil
	case models.ChannelSMS:
		return f.twilio, nil
	case models.ChannelPush:
		return f.firebase, nil
	case models.ChannelWebhook:
		return f.webhook, nil
	case models.ChannelInApp:
		return nil, fmt.Errorf("in-app notifications are handled by the delivery engine internally")
	default:
		return nil, fmt.Errorf("unsupported channel: %s", channel)
	}
}
