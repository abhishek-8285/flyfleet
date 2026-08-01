package inbox

import (
	"encoding/json"
	"testing"
	"time"
)

func TestInboxItemSerialization(t *testing.T) {
	now := time.Now().UTC()
	item := Item{
		ID:         "inbox_123",
		Title:      "Order Shipped",
		Body:       "Your package is on the way",
		Category:   "order",
		ActionURL:  "https://gonotify.local/orders/123",
		ImageURL:   "https://gonotify.local/img/pkg.png",
		Data:       map[string]any{"tracking": "TRK987"},
		IsRead:     false,
		IsArchived: false,
		IsPinned:   true,
		CreatedAt:  now,
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("json.Marshal(Item) unexpected error: %v", err)
	}

	var parsed Item
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json.Unmarshal(Item) unexpected error: %v", err)
	}

	if parsed.ID != item.ID {
		t.Errorf("parsed.ID = %q; want %q", parsed.ID, item.ID)
	}
	if parsed.Title != item.Title {
		t.Errorf("parsed.Title = %q; want %q", parsed.Title, item.Title)
	}
	if !parsed.IsPinned {
		t.Errorf("parsed.IsPinned = false; want true")
	}
}

func TestNewRepository(t *testing.T) {
	repo := NewRepository(nil)
	if repo == nil {
		t.Fatal("NewRepository() returned nil")
	}
}
