package provider

import (
	"encoding/json"
	"testing"
)

func TestProviderStructSerialization(t *testing.T) {
	req := Request{
		TenantID:       "tenant_1",
		NotificationID: "notif_1",
		Channel:        ChannelEmail,
		To: Recipient{
			UserID: "usr_1",
			Email:  "user@example.com",
			Phone:  "+14155552671",
		},
		From:     "noreply@example.com",
		FromName: "GoNotify",
		Subject:  "Welcome",
		Text:     "Welcome to GoNotify",
		HTML:     "<p>Welcome to GoNotify</p>",
		Attachments: []Attachment{
			{
				Filename:    "welcome.pdf",
				ContentType: "application/pdf",
				URL:         "https://example.com/welcome.pdf",
			},
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal(Request) unexpected error: %v", err)
	}

	var parsed Request
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json.Unmarshal(Request) unexpected error: %v", err)
	}

	if parsed.TenantID != req.TenantID || parsed.To.Email != req.To.Email {
		t.Errorf("parsed = %v; want %v", parsed, req)
	}

	res := Result{
		ProviderMessageID: "msg_123",
		Status:            "sent",
		Raw:               map[string]any{"ok": true},
	}

	resData, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("json.Marshal(Result) unexpected error: %v", err)
	}

	var parsedRes Result
	if err := json.Unmarshal(resData, &parsedRes); err != nil {
		t.Fatalf("json.Unmarshal(Result) unexpected error: %v", err)
	}

	if parsedRes.ProviderMessageID != res.ProviderMessageID || parsedRes.Status != res.Status {
		t.Errorf("parsedRes = %v; want %v", parsedRes, res)
	}
}
