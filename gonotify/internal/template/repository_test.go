package template

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTemplateStructs(t *testing.T) {
	now := time.Now().UTC()
	tpl := Template{
		ID:          "tpl_123",
		TenantID:    "tenant_456",
		Key:         "order.shipped",
		Name:        "Order Shipped Template",
		Description: "Sent when order is shipped",
		Channel:     "email",
		Category:    "transactional",
		CreatedAt:   now,
		UpdatedAt:   now,
		Versions: []TemplateVersion{
			{
				ID:         "ver_1",
				TemplateID: "tpl_123",
				Version:    1,
				Subject:    "Order {{.order_id}} Shipped",
				HTMLBody:   "<p>Shipped!</p>",
				TextBody:   "Shipped!",
				Status:     "published",
				CreatedAt:  now,
				UpdatedAt:  now,
			},
		},
	}

	data, err := json.Marshal(tpl)
	if err != nil {
		t.Fatalf("json.Marshal(Template) unexpected error: %v", err)
	}

	var parsed Template
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json.Unmarshal(Template) unexpected error: %v", err)
	}

	if parsed.Key != "order.shipped" {
		t.Errorf("parsed.Key = %q; want 'order.shipped'", parsed.Key)
	}
	if len(parsed.Versions) != 1 || parsed.Versions[0].Status != "published" {
		t.Errorf("parsed.Versions = %v; want 1 version with published status", parsed.Versions)
	}
}

func TestTemplateErrors(t *testing.T) {
	if ErrTemplateNotFound.Error() != "template not found" {
		t.Errorf("ErrTemplateNotFound = %q", ErrTemplateNotFound.Error())
	}
	if ErrVersionNotFound.Error() != "template version not found" {
		t.Errorf("ErrVersionNotFound = %q", ErrVersionNotFound.Error())
	}
}
