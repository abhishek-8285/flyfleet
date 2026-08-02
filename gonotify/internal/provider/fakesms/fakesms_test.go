package fakesms

import (
	"context"
	"testing"

	"github.com/gonotify/gonotify/internal/provider"
)

func TestFakeSMSMetadata(t *testing.T) {
	p := New(Config{})

	if p.Name() != "fake_sms" {
		t.Errorf("Name() = %q; want 'fake_sms'", p.Name())
	}

	if p.Channel() != provider.ChannelSMS {
		t.Errorf("Channel() = %v; want %v", p.Channel(), provider.ChannelSMS)
	}

	if err := p.Health(context.Background()); err != nil {
		t.Errorf("Health() returned unexpected error: %v", err)
	}
}

func TestFakeSMSSendValidation(t *testing.T) {
	p := New(Config{})
	ctx := context.Background()

	// Missing phone
	_, err := p.Send(ctx, provider.Request{
		Text: "Hello world",
	})
	if err == nil {
		t.Error("Send() expected error for missing phone, got nil")
	}

	// Missing text
	_, err = p.Send(ctx, provider.Request{
		To: provider.Recipient{Phone: "+14155552671"},
	})
	if err == nil {
		t.Error("Send() expected error for missing text, got nil")
	}
}
