package routing

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gonotify/gonotify/internal/provider"
)

// Repository provides database access for provider config routing.
type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// GetConfigs returns enabled provider configs for a given tenant and
// channel, ordered by priority (ascending) then weight (descending).
// This implements the routing query that lets the dispatcher try
// multiple providers in priority order before falling back.
func (r *Repository) GetConfigs(ctx context.Context, tenantID string, channel string) ([]provider.ProviderConfig, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			pc.id::text,
			p.type,
			p.channel,
			pc.name,
			pc.settings,
			pc.credentials_encrypted,
			pc.priority,
			pc.weight
		FROM provider_configs pc
		JOIN providers p ON p.id = pc.provider_id
		WHERE pc.tenant_id = $1::uuid
		  AND p.channel = $2
		  AND pc.enabled = true
		ORDER BY pc.priority ASC, pc.weight DESC
	`, tenantID, channel)
	if err != nil {
		return nil, fmt.Errorf("query provider configs: %w", err)
	}
	defer rows.Close()

	list := make([]provider.ProviderConfig, 0)
	for rows.Next() {
		var cfg provider.ProviderConfig
		var settingsBytes []byte

		if err := rows.Scan(
			&cfg.ID,
			&cfg.Type,
			&cfg.Channel,
			&cfg.Name,
			&settingsBytes,
			&cfg.CredentialsEncrypted,
			&cfg.Priority,
			&cfg.Weight,
		); err != nil {
			return nil, fmt.Errorf("scan provider config: %w", err)
		}

		if len(settingsBytes) > 0 {
			_ = json.Unmarshal(settingsBytes, &cfg.Settings)
		}

		list = append(list, cfg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate provider configs: %w", err)
	}

	return list, nil
}
