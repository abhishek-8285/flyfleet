package featurestore

import (
	"context"
	"time"
)

// FeatureValue is a single feature observation.
type FeatureValue struct {
	Name      string      `json:"name"`
	Value     interface{} `json:"value"`
	Timestamp time.Time   `json:"timestamp"`
}

// FeatureVector is a set of features for one entity.
type FeatureVector struct {
	EntityID   string         `json:"entity_id"`
	EntityType string         `json:"entity_type"` // "driver", "rider", "zone"
	Features   []FeatureValue `json:"features"`
}

// Store is the unified feature store interface.
type Store interface {
	// Online serving (low-latency, current state)
	GetOnlineFeatures(ctx context.Context, entityType, entityID string, featureNames []string) (*FeatureVector, error)
	PutOnlineFeatures(ctx context.Context, fv *FeatureVector) error

	// Offline store (historical, for training)
	GetOfflineFeatures(ctx context.Context, entityType string, entityIDs []string, featureNames []string, start, end time.Time) ([]*FeatureVector, error)
	WriteOfflineFeatures(ctx context.Context, fvs []*FeatureVector) error

	// Point-in-time join for training datasets
	PointInTimeJoin(ctx context.Context, entityTable string, timestamps []time.Time, featureNames []string) ([]*FeatureVector, error)

	// Registry
	RegisterFeature(ctx context.Context, def *FeatureDefinition) error
	ListFeatures(ctx context.Context, entityType string) ([]*FeatureDefinition, error)
}

// FeatureDefinition describes a feature's schema.
type FeatureDefinition struct {
	Name        string            `json:"name"`
	EntityType  string            `json:"entity_type"`
	DType       string            `json:"dtype"` // "float64", "int64", "string", "bool"
	Description string            `json:"description"`
	Source      string            `json:"source"` // "stream", "batch", "push"
	Tags        map[string]string `json:"tags"`
	CreatedAt   time.Time         `json:"created_at"`
}
