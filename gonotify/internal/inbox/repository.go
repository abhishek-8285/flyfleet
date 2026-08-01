package inbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInboxItemNotFound = errors.New("inbox item not found")

type Item struct {
	ID             string         `json:"id"`
	NotificationID *string        `json:"notification_id,omitempty"`
	Title          string         `json:"title"`
	Body           string         `json:"body"`
	Category       string         `json:"category"`
	ActionURL      string         `json:"action_url"`
	ImageURL       string         `json:"image_url"`
	Data           map[string]any `json:"data"`
	IsRead         bool           `json:"is_read"`
	IsArchived     bool           `json:"is_archived"`
	IsPinned       bool           `json:"is_pinned"`
	ReadAt         *time.Time     `json:"read_at,omitempty"`
	ArchivedAt     *time.Time     `json:"archived_at,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
}

type CreateItemParams struct {
	TenantID       string
	UserID         string
	NotificationID *string
	Title          string
	Body           string
	Category       string
	ActionURL      string
	ImageURL       string
	Data           map[string]any
	IdempotencyKey string
}

type ListOptions struct {
	Unread   *bool
	Archived *bool
	Limit    int
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) CreateItemIdempotent(
	ctx context.Context,
	params CreateItemParams,
) (string, bool, error) {
	if params.Data == nil {
		params.Data = map[string]any{}
	}

	dataJSON, err := json.Marshal(params.Data)
	if err != nil {
		return "", false, fmt.Errorf("marshal inbox data: %w", err)
	}

	var itemID string

	err = r.db.QueryRow(
		ctx,
		`
		INSERT INTO inbox_items (
			tenant_id,
			user_id,
			notification_id,
			title,
			body,
			category,
			action_url,
			image_url,
			data,
			idempotency_key
		)
		VALUES (
			$1::uuid,
			$2::uuid,
			$3::uuid,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9::jsonb,
			$10
		)
		ON CONFLICT (tenant_id, idempotency_key) DO NOTHING
		RETURNING id::text
		`,
		params.TenantID,
		params.UserID,
		params.NotificationID,
		params.Title,
		params.Body,
		params.Category,
		params.ActionURL,
		params.ImageURL,
		dataJSON,
		params.IdempotencyKey,
	).Scan(&itemID)

	if err == nil {
		return itemID, true, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return "", false, fmt.Errorf("insert inbox item: %w", err)
	}

	err = r.db.QueryRow(
		ctx,
		`
		SELECT id::text
		FROM inbox_items
		WHERE tenant_id = $1::uuid
		  AND idempotency_key = $2
		`,
		params.TenantID,
		params.IdempotencyKey,
	).Scan(&itemID)

	if err != nil {
		return "", false, fmt.Errorf("fetch existing inbox item: %w", err)
	}

	return itemID, false, nil
}

func (r *Repository) List(
	ctx context.Context,
	tenantID string,
	userID string,
	opts ListOptions,
) ([]Item, error) {
	limit := opts.Limit

	if limit <= 0 {
		limit = 50
	}

	if limit > 200 {
		limit = 200
	}

	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			id::text,
			notification_id::text,
			title,
			COALESCE(body, ''),
			COALESCE(category, ''),
			COALESCE(action_url, ''),
			COALESCE(image_url, ''),
			data,
			is_read,
			is_archived,
			is_pinned,
			read_at,
			archived_at,
			created_at
		FROM inbox_items
		WHERE tenant_id = $1::uuid
		  AND user_id = $2::uuid
		  AND ($3::boolean IS NULL OR is_read = $3)
		  AND ($4::boolean IS NULL OR is_archived = $4)
		ORDER BY is_pinned DESC, created_at DESC
		LIMIT $5
		`,
		tenantID,
		userID,
		opts.Unread,
		opts.Archived,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("query inbox items: %w", err)
	}
	defer rows.Close()

	items := make([]Item, 0)

	for rows.Next() {
		var item Item
		var dataJSON []byte

		if err := rows.Scan(
			&item.ID,
			&item.NotificationID,
			&item.Title,
			&item.Body,
			&item.Category,
			&item.ActionURL,
			&item.ImageURL,
			&dataJSON,
			&item.IsRead,
			&item.IsArchived,
			&item.IsPinned,
			&item.ReadAt,
			&item.ArchivedAt,
			&item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan inbox item: %w", err)
		}

		if len(dataJSON) > 0 {
			if err := json.Unmarshal(dataJSON, &item.Data); err != nil {
				return nil, fmt.Errorf("unmarshal inbox data: %w", err)
			}
		}

		if item.Data == nil {
			item.Data = map[string]any{}
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate inbox items: %w", err)
	}

	return items, nil
}

func (r *Repository) MarkRead(
	ctx context.Context,
	tenantID string,
	userID string,
	itemID string,
) error {
	return r.updateState(
		ctx,
		tenantID,
		userID,
		itemID,
		`is_read = true, read_at = now()`,
	)
}

func (r *Repository) MarkUnread(
	ctx context.Context,
	tenantID string,
	userID string,
	itemID string,
) error {
	return r.updateState(
		ctx,
		tenantID,
		userID,
		itemID,
		`is_read = false, read_at = NULL`,
	)
}

func (r *Repository) Archive(
	ctx context.Context,
	tenantID string,
	userID string,
	itemID string,
) error {
	return r.updateState(
		ctx,
		tenantID,
		userID,
		itemID,
		`is_archived = true, archived_at = now()`,
	)
}

func (r *Repository) Unarchive(
	ctx context.Context,
	tenantID string,
	userID string,
	itemID string,
) error {
	return r.updateState(
		ctx,
		tenantID,
		userID,
		itemID,
		`is_archived = false, archived_at = NULL`,
	)
}

func (r *Repository) Pin(
	ctx context.Context,
	tenantID string,
	userID string,
	itemID string,
) error {
	return r.updateState(
		ctx,
		tenantID,
		userID,
		itemID,
		`is_pinned = true`,
	)
}

func (r *Repository) Unpin(
	ctx context.Context,
	tenantID string,
	userID string,
	itemID string,
) error {
	return r.updateState(
		ctx,
		tenantID,
		userID,
		itemID,
		`is_pinned = false`,
	)
}

func (r *Repository) updateState(
	ctx context.Context,
	tenantID string,
	userID string,
	itemID string,
	setClause string,
) error {
	query := fmt.Sprintf(
		`
		UPDATE inbox_items
		SET %s
		WHERE id = $1::uuid
		  AND tenant_id = $2::uuid
		  AND user_id = $3::uuid
		`,
		setClause,
	)

	tag, err := r.db.Exec(ctx, query, itemID, tenantID, userID)
	if err != nil {
		return fmt.Errorf("update inbox item: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrInboxItemNotFound
	}

	return nil
}
