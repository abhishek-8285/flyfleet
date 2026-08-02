package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	appcrypto "github.com/gonotify/gonotify/internal/crypto"
)

var (
	ErrProviderNotFound       = errors.New("provider not found")
	ErrProviderConfigNotFound = errors.New("provider config not found")
)

type SystemProvider struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Channel      string         `json:"channel"`
	Name         string         `json:"name"`
	Capabilities map[string]any `json:"capabilities"`
	CreatedAt    time.Time      `json:"created_at"`
}

type ProviderConfigView struct {
	ID         string         `json:"id"`
	TenantID   string         `json:"tenant_id"`
	ProviderID string         `json:"provider_id"`
	Type       string         `json:"type"`
	Channel    string         `json:"channel"`
	Name       string         `json:"name"`
	Settings   map[string]any `json:"settings"`
	Enabled    bool           `json:"enabled"`
	Priority   int            `json:"priority"`
	Weight     int            `json:"weight"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

type CreateProviderConfigRequest struct {
	ProviderID  string           `json:"provider_id"`
	Name        string           `json:"name"`
	Settings    map[string]any   `json:"settings"`
	Credentials map[string]string `json:"credentials"`
	Priority    int              `json:"priority"`
	Weight      int              `json:"weight"`
}

type Repository struct {
	db            *pgxpool.Pool
	encryptionKey []byte
}

func NewRepository(db *pgxpool.Pool, encryptionKey []byte) *Repository {
	return &Repository{db: db, encryptionKey: encryptionKey}
}

func (r *Repository) ListSystemProviders(ctx context.Context) ([]SystemProvider, error) {
	rows, err := r.db.Query(
		ctx,
		`
		SELECT id::text, type, channel, name, capabilities, created_at
		FROM providers
		ORDER BY channel, type
		`,
	)
	if err != nil {
		return nil, fmt.Errorf("query system providers: %w", err)
	}
	defer rows.Close()

	list := make([]SystemProvider, 0)
	for rows.Next() {
		var p SystemProvider
		var capBytes []byte
		if err := rows.Scan(&p.ID, &p.Type, &p.Channel, &p.Name, &capBytes, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan provider: %w", err)
		}
		_ = json.Unmarshal(capBytes, &p.Capabilities)
		list = append(list, p)
	}
	return list, nil
}

func (r *Repository) CreateProviderConfig(
	ctx context.Context,
	tenantID string,
	req CreateProviderConfigRequest,
) (ProviderConfigView, error) {
	if req.Priority == 0 {
		req.Priority = 100
	}
	if req.Weight == 0 {
		req.Weight = 100
	}
	settingsBytes, err := json.Marshal(req.Settings)
	if err != nil {
		return ProviderConfigView{}, fmt.Errorf("marshal settings: %w", err)
	}

	var credentialsEncrypted []byte
	if len(req.Credentials) > 0 {
		credJSON, err := json.Marshal(req.Credentials)
		if err != nil {
			return ProviderConfigView{}, fmt.Errorf("marshal credentials: %w", err)
		}
		credentialsEncrypted, err = appcrypto.Encrypt(credJSON, r.encryptionKey)
		if err != nil {
			return ProviderConfigView{}, fmt.Errorf("encrypt credentials: %w", err)
		}
	}

	var view ProviderConfigView
	var viewSettingsBytes []byte

	err = r.db.QueryRow(
		ctx,
		`
		INSERT INTO provider_configs (
			tenant_id,
			provider_id,
			name,
			settings,
			credentials_encrypted,
			enabled,
			priority,
			weight
		)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5, true, $6, $7)
		RETURNING
			id::text,
			tenant_id::text,
			provider_id::text,
			name,
			settings,
			enabled,
			priority,
			weight,
			created_at,
			updated_at
		`,
		tenantID,
		req.ProviderID,
		req.Name,
		settingsBytes,
		credentialsEncrypted,
		req.Priority,
		req.Weight,
	).Scan(
		&view.ID,
		&view.TenantID,
		&view.ProviderID,
		&view.Name,
		&viewSettingsBytes,
		&view.Enabled,
		&view.Priority,
		&view.Weight,
		&view.CreatedAt,
		&view.UpdatedAt,
	)

	if err != nil {
		return ProviderConfigView{}, fmt.Errorf("insert provider config: %w", err)
	}

	_ = json.Unmarshal(viewSettingsBytes, &view.Settings)

	_ = r.db.QueryRow(ctx, `SELECT type, channel FROM providers WHERE id = $1::uuid`, req.ProviderID).Scan(&view.Type, &view.Channel)

	return view, nil
}

func (r *Repository) ListProviderConfigs(
	ctx context.Context,
	tenantID string,
) ([]ProviderConfigView, error) {
	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			pc.id::text,
			pc.tenant_id::text,
			pc.provider_id::text,
			p.type,
			p.channel,
			pc.name,
			pc.settings,
			pc.enabled,
			pc.priority,
			pc.weight,
			pc.created_at,
			pc.updated_at
		FROM provider_configs pc
		JOIN providers p ON p.id = pc.provider_id
		WHERE pc.tenant_id = $1::uuid
		ORDER BY p.channel, pc.priority ASC, pc.created_at DESC
		`,
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("query provider configs: %w", err)
	}
	defer rows.Close()

	list := make([]ProviderConfigView, 0)
	for rows.Next() {
		var pc ProviderConfigView
		var settingsBytes []byte
		if err := rows.Scan(
			&pc.ID,
			&pc.TenantID,
			&pc.ProviderID,
			&pc.Type,
			&pc.Channel,
			&pc.Name,
			&settingsBytes,
			&pc.Enabled,
			&pc.Priority,
			&pc.Weight,
			&pc.CreatedAt,
			&pc.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan provider config: %w", err)
		}
		_ = json.Unmarshal(settingsBytes, &pc.Settings)
		list = append(list, pc)
	}

	return list, nil
}

func (r *Repository) SetProviderConfigStatus(
	ctx context.Context,
	tenantID string,
	configID string,
	enabled bool,
) (ProviderConfigView, error) {
	var pc ProviderConfigView
	var settingsBytes []byte

	err := r.db.QueryRow(
		ctx,
		`
		UPDATE provider_configs
		SET enabled = $1
		WHERE tenant_id = $2::uuid AND id = $3::uuid
		RETURNING
			id::text,
			tenant_id::text,
			provider_id::text,
			name,
			settings,
			enabled,
			priority,
			weight,
			created_at,
			updated_at
		`,
		enabled,
		tenantID,
		configID,
	).Scan(
		&pc.ID,
		&pc.TenantID,
		&pc.ProviderID,
		&pc.Name,
		&settingsBytes,
		&pc.Enabled,
		&pc.Priority,
		&pc.Weight,
		&pc.CreatedAt,
		&pc.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return ProviderConfigView{}, ErrProviderConfigNotFound
	}
	if err != nil {
		return ProviderConfigView{}, fmt.Errorf("update provider config status: %w", err)
	}

	_ = json.Unmarshal(settingsBytes, &pc.Settings)
	_ = r.db.QueryRow(ctx, `SELECT type, channel FROM providers WHERE id = $1::uuid`, pc.ProviderID).Scan(&pc.Type, &pc.Channel)

	return pc, nil
}
