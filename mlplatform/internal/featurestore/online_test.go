package featurestore

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

func TestRedisOnlineStore_RoundTrip(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	defer mr.Close()

	fs := NewRedisOnlineStore(mr.Addr(), "", 0)
	ctx := context.Background()

	fv := &FeatureVector{
		EntityType: "driver",
		EntityID:   "d1",
		Features: []FeatureValue{{
			Name:      "acceptance_rate",
			Value:     0.95,
			Timestamp: time.Now().UTC(),
		}},
	}
	if err := fs.PutOnlineFeatures(ctx, fv); err != nil {
		t.Fatalf("put: %v", err)
	}

	got, err := fs.GetOnlineFeatures(ctx, "driver", "d1", []string{"acceptance_rate"})
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(got.Features) != 1 {
		t.Fatalf("expected 1 feature, got %d", len(got.Features))
	}
	if got.Features[0].Name != "acceptance_rate" {
		t.Fatalf("expected feature acceptance_rate, got %q", got.Features[0].Name)
	}
}
