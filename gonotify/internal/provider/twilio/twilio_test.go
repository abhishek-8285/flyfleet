package twilio

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/gonotify/gonotify/internal/provider"
)

type mockTransport struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}

func TestTwilioMetadata(t *testing.T) {
	p := New(Config{
		AccountSID: "ACtest123",
		AuthToken:  "authtoken123",
		From:       "+15005550006",
	})

	if p.Name() != "twilio" {
		t.Errorf("Name() = %q; want 'twilio'", p.Name())
	}

	if p.Channel() != provider.ChannelSMS {
		t.Errorf("Channel() = %v; want %v", p.Channel(), provider.ChannelSMS)
	}

	if err := p.Health(context.Background()); err != nil {
		t.Errorf("Health() returned unexpected error: %v", err)
	}
}

func TestTwilioSendValidation(t *testing.T) {
	p := New(Config{
		AccountSID: "ACtest123",
		AuthToken:  "authtoken123",
		From:       "+15005550006",
	})

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

	// Missing From number (req.From and cfg.From empty)
	pNoFrom := New(Config{
		AccountSID: "ACtest123",
		AuthToken:  "authtoken123",
	})
	_, err = pNoFrom.Send(ctx, provider.Request{
		To:   provider.Recipient{Phone: "+14155552671"},
		Text: "Hello world",
	})
	if err == nil {
		t.Error("Send() expected error for missing from number, got nil")
	}

	// Missing Twilio credentials
	pNoCreds := New(Config{})
	_, err = pNoCreds.Send(ctx, provider.Request{
		To:   provider.Recipient{Phone: "+14155552671"},
		Text: "Hello world",
		From: "+15005550006",
	})
	if err == nil {
		t.Error("Send() expected error for missing credentials, got nil")
	}
}

func TestTwilioSendSuccess(t *testing.T) {
	p := New(Config{
		AccountSID: "ACtest123",
		AuthToken:  "authtoken123",
		From:       "+15005550006",
	})

	p.client.Transport = &mockTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			// Verify basic auth
			username, password, ok := req.BasicAuth()
			if !ok || username != "ACtest123" || password != "authtoken123" {
				t.Errorf("BasicAuth failed: username=%s, ok=%v", username, ok)
			}

			// Verify headers
			if req.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
				t.Errorf("Content-Type header = %q; want application/x-www-form-urlencoded", req.Header.Get("Content-Type"))
			}

			jsonResp := `{"sid":"SM123456","status":"queued","error_code":0,"error_message":""}`
			return &http.Response{
				StatusCode: http.StatusCreated,
				Body:       io.NopCloser(bytes.NewBufferString(jsonResp)),
				Header:     make(http.Header),
			}, nil
		},
	}

	res, err := p.Send(context.Background(), provider.Request{
		To:   provider.Recipient{Phone: "+14155552671"},
		Text: "Test message",
	})
	if err != nil {
		t.Fatalf("Send() unexpected error: %v", err)
	}

	if res.ProviderMessageID != "SM123456" {
		t.Errorf("ProviderMessageID = %q; want SM123456", res.ProviderMessageID)
	}

	if res.Status != "queued" {
		t.Errorf("Status = %q; want queued", res.Status)
	}
}

func TestTwilioSendAPIError(t *testing.T) {
	p := New(Config{
		AccountSID: "ACtest123",
		AuthToken:  "authtoken123",
		From:       "+15005550006",
	})

	p.client.Transport = &mockTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			jsonResp := `{"sid":"","status":"","error_code":21614,"error_message":"Invalid To Phone Number"}`
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(bytes.NewBufferString(jsonResp)),
				Header:     make(http.Header),
			}, nil
		},
	}

	_, err := p.Send(context.Background(), provider.Request{
		To:   provider.Recipient{Phone: "invalid_phone"},
		Text: "Test message",
	})
	if err == nil {
		t.Fatal("Send() expected API error response, got nil")
	}
}

func TestTwilioSendInvalidJSON(t *testing.T) {
	p := New(Config{
		AccountSID: "ACtest123",
		AuthToken:  "authtoken123",
		From:       "+15005550006",
	})

	p.client.Transport = &mockTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString("not_valid_json")),
				Header:     make(http.Header),
			}, nil
		},
	}

	_, err := p.Send(context.Background(), provider.Request{
		To:   provider.Recipient{Phone: "+14155552671"},
		Text: "Test message",
	})
	if err == nil {
		t.Fatal("Send() expected error for invalid JSON body, got nil")
	}
}
