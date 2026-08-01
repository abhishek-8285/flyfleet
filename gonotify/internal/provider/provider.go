package provider

import "context"

type Channel string

const (
	ChannelEmail Channel = "email"
	ChannelSMS   Channel = "sms"
	ChannelPush  Channel = "push"
	ChannelInApp Channel = "in_app"
)

type Recipient struct {
	UserID      string         `json:"user_id,omitempty"`
	Email       string         `json:"email,omitempty"`
	Phone       string         `json:"phone,omitempty"`
	DeviceToken string         `json:"device_token,omitempty"`
	Locale      string         `json:"locale,omitempty"`
	Timezone    string         `json:"timezone,omitempty"`
	Attributes  map[string]any `json:"attributes,omitempty"`
}

type Attachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	URL         string `json:"url,omitempty"`
	Content     []byte `json:"-"`
}

type Request struct {
	TenantID       string            `json:"tenant_id"`
	NotificationID string            `json:"notification_id"`
	Channel        Channel           `json:"channel"`
	To             Recipient         `json:"to"`
	From           string            `json:"from,omitempty"`
	FromName       string            `json:"from_name,omitempty"`
	ReplyTo        string            `json:"reply_to,omitempty"`
	Subject        string            `json:"subject"`
	Text           string            `json:"text"`
	HTML           string            `json:"html"`
	Data           map[string]any    `json:"data,omitempty"`
	Attachments    []Attachment      `json:"attachments,omitempty"`
	Headers        map[string]string `json:"headers,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

type Result struct {
	ProviderMessageID string         `json:"provider_message_id,omitempty"`
	Status            string         `json:"status"`
	Raw               map[string]any `json:"raw,omitempty"`
}

type Provider interface {
	Name() string
	Channel() Channel
	Send(ctx context.Context, req Request) (Result, error)
	Health(ctx context.Context) error
}
