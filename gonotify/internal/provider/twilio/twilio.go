package twilio

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gonotify/gonotify/internal/provider"
)

type Config struct {
	AccountSID string
	AuthToken  string
	From       string
}

type Provider struct {
	cfg    Config
	client *http.Client
}

func New(cfg Config) *Provider {
	return &Provider{
		cfg: cfg,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (p *Provider) Name() string {
	return "twilio"
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

	from := req.From
	if from == "" {
		from = p.cfg.From
	}

	if from == "" {
		return provider.Result{}, fmt.Errorf("twilio from number is required")
	}

	if p.cfg.AccountSID == "" || p.cfg.AuthToken == "" {
		return provider.Result{}, fmt.Errorf("twilio credentials are required")
	}

	endpoint := fmt.Sprintf(
		"https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json",
		p.cfg.AccountSID,
	)

	form := url.Values{}
	form.Set("From", from)
	form.Set("To", req.To.Phone)
	form.Set("Body", req.Text)

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return provider.Result{}, fmt.Errorf("create twilio request: %w", err)
	}

	httpReq.SetBasicAuth(p.cfg.AccountSID, p.cfg.AuthToken)
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return provider.Result{}, fmt.Errorf("twilio request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return provider.Result{}, fmt.Errorf("read twilio response: %w", err)
	}

	var parsed struct {
		SID          string `json:"sid"`
		Status       string `json:"status"`
		ErrorCode    int    `json:"error_code"`
		ErrorMessage string `json:"error_message"`
	}

	if err := json.Unmarshal(body, &parsed); err != nil {
		return provider.Result{}, fmt.Errorf("decode twilio response: %w", err)
	}

	if resp.StatusCode >= 300 {
		return provider.Result{}, fmt.Errorf(
			"twilio error: status=%d code=%d message=%s",
			resp.StatusCode,
			parsed.ErrorCode,
			parsed.ErrorMessage,
		)
	}

	return provider.Result{
		ProviderMessageID: parsed.SID,
		Status:            parsed.Status,
		Raw: map[string]any{
			"provider": "twilio",
			"sid":      parsed.SID,
			"status":   parsed.Status,
		},
	}, nil
}

func (p *Provider) Health(ctx context.Context) error {
	// For Phase 2C, keep health simple.
	// Later, check credentials and account quota.
	return nil
}
