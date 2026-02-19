package pagination

import (
	"encoding/json"
	"testing"
)

func TestParams_Normalize(t *testing.T) {
	tests := []struct {
		name     string
		input    Params
		wantPage int
		wantLim  int
	}{
		{"defaults", Params{}, 1, 20},
		{"negative page", Params{Page: -5, Limit: 10}, 1, 10},
		{"zero limit", Params{Page: 3, Limit: 0}, 3, 20},
		{"over max limit", Params{Page: 1, Limit: 200}, 1, 100},
		{"valid", Params{Page: 5, Limit: 50}, 5, 50},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.input.Normalize()
			if tt.input.Page != tt.wantPage {
				t.Errorf("Page = %d, want %d", tt.input.Page, tt.wantPage)
			}
			if tt.input.Limit != tt.wantLim {
				t.Errorf("Limit = %d, want %d", tt.input.Limit, tt.wantLim)
			}
		})
	}
}

func TestParams_Offset(t *testing.T) {
	tests := []struct {
		page, limit, want int
	}{
		{1, 20, 0},
		{2, 20, 20},
		{3, 10, 20},
		{1, 100, 0},
	}
	for _, tt := range tests {
		p := Params{Page: tt.page, Limit: tt.limit}
		if got := p.Offset(); got != tt.want {
			t.Errorf("Offset(%d,%d) = %d, want %d", tt.page, tt.limit, got, tt.want)
		}
	}
}

func TestNewMeta(t *testing.T) {
	tests := []struct {
		name       string
		total      int
		page       int
		limit      int
		wantPages  int
		wantNext   bool
		wantPrev   bool
	}{
		{"empty results", 0, 1, 20, 0, false, false},
		{"single page", 5, 1, 20, 1, false, false},
		{"first of many", 50, 1, 20, 3, true, false},
		{"middle page", 50, 2, 20, 3, true, true},
		{"last page", 50, 3, 20, 3, false, true},
		{"exact fit", 40, 2, 20, 2, false, true},
		{"one item", 1, 1, 20, 1, false, false},
		{"over page", 5, 10, 20, 1, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMeta(tt.total, tt.page, tt.limit)
			if m.TotalPages != tt.wantPages {
				t.Errorf("TotalPages = %d, want %d", m.TotalPages, tt.wantPages)
			}
			if m.HasNext != tt.wantNext {
				t.Errorf("HasNext = %v, want %v", m.HasNext, tt.wantNext)
			}
			if m.HasPrev != tt.wantPrev {
				t.Errorf("HasPrev = %v, want %v", m.HasPrev, tt.wantPrev)
			}
			if m.Total != tt.total {
				t.Errorf("Total = %d, want %d", m.Total, tt.total)
			}
		})
	}
}

func TestNewResponse_NilData(t *testing.T) {
	resp := NewResponse[string](nil, 0, 1, 20)
	if resp.Data == nil {
		t.Error("Data should be empty slice, not nil")
	}
	if len(resp.Data) != 0 {
		t.Errorf("Data length = %d, want 0", len(resp.Data))
	}
}

func TestNewResponse_JSONStructure(t *testing.T) {
	type Item struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	items := []Item{{ID: "1", Name: "a"}, {ID: "2", Name: "b"}}
	resp := NewResponse(items, 50, 1, 20)

	b, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if _, ok := raw["data"]; !ok {
		t.Error("missing 'data' key in JSON")
	}
	if _, ok := raw["meta"]; !ok {
		t.Error("missing 'meta' key in JSON")
	}

	// Verify meta fields
	var meta Meta
	if err := json.Unmarshal(raw["meta"], &meta); err != nil {
		t.Fatalf("Unmarshal meta error: %v", err)
	}
	if meta.Total != 50 || meta.Page != 1 || meta.Limit != 20 || meta.TotalPages != 3 {
		t.Errorf("unexpected meta values: %+v", meta)
	}
	if !meta.HasNext {
		t.Error("HasNext should be true")
	}
	if meta.HasPrev {
		t.Error("HasPrev should be false")
	}
}

func TestApplyToQuery(t *testing.T) {
	p := Params{Page: 3, Limit: 25}
	q, args := ApplyToQuery("SELECT * FROM drivers WHERE org_id = $1", p, 2)

	expected := "SELECT * FROM drivers WHERE org_id = $1 LIMIT $2 OFFSET $3"
	if q != expected {
		t.Errorf("query = %q, want %q", q, expected)
	}
	if len(args) != 2 {
		t.Fatalf("args length = %d, want 2", len(args))
	}
	if args[0] != 25 {
		t.Errorf("limit arg = %v, want 25", args[0])
	}
	if args[1] != 50 { // (3-1) * 25
		t.Errorf("offset arg = %v, want 50", args[1])
	}
}
