package warehouse

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/flyfleet/mlplatform/internal/featurestore"
)

type ClickHouseStore struct {
	conn driver.Conn
}

func NewClickHouse(dsn string) (*ClickHouseStore, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{dsn},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
	})
	if err != nil {
		return nil, err
	}
	return &ClickHouseStore{conn: conn}, nil
}

// Migrate creates the core tables.
func (s *ClickHouseStore) Migrate(ctx context.Context) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS events (
			id          String,
			type        LowCardinality(String),
			source      LowCardinality(String),
			subject     String,
			timestamp   DateTime64(3),
			data        String,
			metadata    Map(String, String),
			trace_id    String
		) ENGINE = MergeTree()
		PARTITION BY toYYYYMMDD(timestamp)
		ORDER BY (type, subject, timestamp)`,

		`CREATE TABLE IF NOT EXISTS features (
			entity_type LowCardinality(String),
			entity_id   String,
			name        LowCardinality(String),
			value       Float64,
			timestamp   DateTime64(3)
		) ENGINE = MergeTree()
		PARTITION BY (entity_type, toYYYYMMDD(timestamp))
		ORDER BY (entity_type, entity_id, name, timestamp)`,

		`CREATE TABLE IF NOT EXISTS predictions (
			id          String,
			model_name  LowCardinality(String),
			model_version String,
			entity_id   String,
			score       Float64,
			features    Map(String, Float64),
			timestamp   DateTime64(3),
			experiment  String
		) ENGINE = MergeTree()
		PARTITION BY toYYYYMMDD(timestamp)
		ORDER BY (model_name, entity_id, timestamp)`,

		`CREATE MATERIALIZED VIEW IF NOT EXISTS events_hourly_mv
		ENGINE = SummingMergeTree()
		PARTITION BY toYYYYMMDD(hour)
		ORDER BY (type, hour)
		AS SELECT
			type,
			toStartOfHour(timestamp) AS hour,
			count() AS cnt
		FROM events
		GROUP BY type, hour`,
	}

	for _, q := range queries {
		if err := s.conn.Exec(ctx, q); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	return nil
}

// WriteEvents batch-inserts raw events.
func (s *ClickHouseStore) WriteEvents(ctx context.Context, events []EventRow) error {
	batch, err := s.conn.PrepareBatch(ctx, "INSERT INTO events")
	if err != nil {
		return err
	}
	for _, e := range events {
		if err := batch.Append(e.ID, e.Type, e.Source, e.Subject, e.Timestamp, e.Data, e.Metadata, e.TraceID); err != nil {
			return err
		}
	}
	return batch.Send()
}

type EventRow struct {
	ID        string
	Type      string
	Source    string
	Subject   string
	Timestamp time.Time
	Data      string
	Metadata  map[string]string
	TraceID   string
}

// GetOfflineFeatures implements the featurestore offline interface.
func (s *ClickHouseStore) GetOfflineFeatures(ctx context.Context, entityType string, entityIDs []string, featureNames []string, start, end time.Time) ([]*featurestore.FeatureVector, error) {
	query := `
		SELECT entity_id, name, value, timestamp
		FROM features
		WHERE entity_type = ?
		  AND entity_id IN (?)
		  AND name IN (?)
		  AND timestamp BETWEEN ? AND ?
		ORDER BY entity_id, name, timestamp
	`

	rows, err := s.conn.Query(ctx, query, entityType, entityIDs, featureNames, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vectorMap := make(map[string]*featurestore.FeatureVector)
	for rows.Next() {
		var entityID, name string
		var value float64
		var ts time.Time
		if err := rows.Scan(&entityID, &name, &value, &ts); err != nil {
			return nil, err
		}
		fv, ok := vectorMap[entityID]
		if !ok {
			fv = &featurestore.FeatureVector{
				EntityID:   entityID,
				EntityType: entityType,
			}
			vectorMap[entityID] = fv
		}
		fv.Features = append(fv.Features, featurestore.FeatureValue{
			Name:      name,
			Value:     value,
			Timestamp: ts,
		})
	}

	result := make([]*featurestore.FeatureVector, 0, len(vectorMap))
	for _, fv := range vectorMap {
		result = append(result, fv)
	}
	return result, nil
}

// PointInTimeJoin retrieves the latest feature value before each timestamp.
func (s *ClickHouseStore) PointInTimeJoin(ctx context.Context, entityType string, entityIDs []string, timestamps []time.Time, featureNames []string) ([]*featurestore.FeatureVector, error) {
	query := `
		SELECT entity_id, name, argMax(value, timestamp) AS value, max(timestamp) AS ts
		FROM features
		WHERE entity_type = ?
		  AND entity_id IN (?)
		  AND name IN (?)
		  AND timestamp <= ?
		GROUP BY entity_id, name
	`

	var results []*featurestore.FeatureVector
	for i, ts := range timestamps {
		rows, err := s.conn.Query(ctx, query, entityType, entityIDs[i:i+1], featureNames, ts)
		if err != nil {
			return nil, err
		}

		fv := &featurestore.FeatureVector{
			EntityID:   entityIDs[i],
			EntityType: entityType,
		}
		for rows.Next() {
			var name string
			var value float64
			var fts time.Time
			rows.Scan(&name, &value, &fts)
			fv.Features = append(fv.Features, featurestore.FeatureValue{
				Name: name, Value: value, Timestamp: fts,
			})
		}
		rows.Close()
		results = append(results, fv)
	}
	return results, nil
}

// WriteFeatures batch-inserts computed features.
func (s *ClickHouseStore) WriteFeatures(ctx context.Context, fvs []*featurestore.FeatureVector) error {
	batch, err := s.conn.PrepareBatch(ctx, "INSERT INTO features")
	if err != nil {
		return err
	}
	for _, fv := range fvs {
		for _, f := range fv.Features {
			val, _ := f.Value.(float64)
			batch.Append(fv.EntityType, fv.EntityID, f.Name, val, f.Timestamp)
		}
	}
	return batch.Send()
}

// Query executes a raw SQL statement and returns streamable rows. It is the
// read path used by batch activities (e.g. computing daily features, hourly
// aggregates, drift sampling) that operate directly on the warehouse.
func (s *ClickHouseStore) Query(ctx context.Context, query string, args ...any) (driver.Rows, error) {
	return s.conn.Query(ctx, query, args...)
}

// CountEvents returns the number of events of a given type.
func (s *ClickHouseStore) CountEvents(ctx context.Context, eventType string) (int64, error) {
	var count int64
	err := s.conn.QueryRow(ctx,
		"SELECT count() FROM events WHERE type = ?", eventType,
	).Scan(&count)
	return count, err
}
