package smtp

import (
	"context"
	"strings"
	"testing"

	"github.com/gonotify/gonotify/internal/provider"
)

func TestSMTPMetadata(t *testing.T) {
	p := New(Config{
		Host:      "localhost",
		Port:      1025,
		FromEmail: "test@gonotify.local",
		FromName:  "GoNotify Test",
	})

	if p.Name() != "smtp" {
		t.Errorf("Name() = %q; want 'smtp'", p.Name())
	}

	if p.Channel() != provider.ChannelEmail {
		t.Errorf("Channel() = %v; want %v", p.Channel(), provider.ChannelEmail)
	}
}

func TestSMTPSendValidation(t *testing.T) {
	p := New(Config{
		Host:      "localhost",
		Port:      1025,
		FromEmail: "test@gonotify.local",
	})

	ctx := context.Background()

	// Missing recipient email
	_, err := p.Send(ctx, provider.Request{
		Subject: "Test Email",
		Text:    "Hello World",
	})
	if err == nil {
		t.Error("Send() expected error for missing recipient email, got nil")
	}

	// Unreachable SMTP server
	pBad := New(Config{
		Host:      "127.0.0.1",
		Port:      65530, // Unused port
		Username:  "user",
		Password:  "pass",
		FromEmail: "test@gonotify.local",
	})

	_, err = pBad.Send(ctx, provider.Request{
		To:      provider.Recipient{Email: "user@example.com"},
		From:    "custom@gonotify.local",
		FromName: "Custom Sender",
		Subject: "Test Email",
		Text:    "Hello World",
	})
	if err == nil {
		t.Error("Send() expected error for unreachable SMTP server, got nil")
	}
}

func TestSMTPHealthUnreachable(t *testing.T) {
	pBad := New(Config{
		Host: "127.0.0.1",
		Port: 65530,
	})

	err := pBad.Health(context.Background())
	if err == nil {
		t.Error("Health() expected error for closed port, got nil")
	}
}

func TestBuildMIME(t *testing.T) {
	mimeData := buildMIME(
		"sender@gonotify.local",
		"recipient@example.com",
		"Test Subject",
		"Plain text content",
		"<p>HTML content</p>",
	)

	body := string(mimeData)

	if !strings.Contains(body, "From: sender@gonotify.local") {
		t.Error("buildMIME missing From header")
	}
	if !strings.Contains(body, "To: recipient@example.com") {
		t.Error("buildMIME missing To header")
	}
	if !strings.Contains(body, "Subject: Test Subject") {
		t.Error("buildMIME missing Subject header")
	}
	if !strings.Contains(body, "Plain text content") {
		t.Error("buildMIME missing text body")
	}
	if !strings.Contains(body, "HTML content") {
		t.Error("buildMIME missing html body")
	}
}
