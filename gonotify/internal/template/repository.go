package template

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrTemplateNotFound = errors.New("template not found")
var ErrVersionNotFound = errors.New("template version not found")

type Template struct {
	ID          string            `json:"id"`
	TenantID    string            `json:"tenant_id"`
	Key         string            `json:"key"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Channel     string            `json:"channel"`
	Category    string            `json:"category"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Versions    []TemplateVersion `json:"versions,omitempty"`
}

type TemplateVersion struct {
	ID         string    `json:"id"`
	TemplateID string    `json:"template_id"`
	Version    int       `json:"version"`
	Subject    string    `json:"subject,omitempty"`
	HTMLBody   string    `json:"html_body,omitempty"`
	TextBody   string    `json:"text_body,omitempty"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ResolvedTemplate struct {
	TemplateID string
	VersionID  string
	Subject    string
	HTML       string
	Text       string
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) GetPublishedVersion(
	ctx context.Context,
	tenantID string,
	key string,
	channel string,
	locale string,
) (ResolvedTemplate, error) {
	_ = locale

	var tpl ResolvedTemplate

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			t.id::text,
			tv.id::text,
			COALESCE(tv.subject, ''),
			COALESCE(tv.html_body, ''),
			COALESCE(tv.text_body, '')
		FROM templates t
		JOIN template_versions tv ON tv.template_id = t.id
		WHERE t.tenant_id = $1::uuid
		  AND t.key = $2
		  AND t.channel = $3
		  AND tv.status = 'published'
		ORDER BY tv.version DESC
		LIMIT 1
		`,
		tenantID,
		key,
		channel,
	).Scan(
		&tpl.TemplateID,
		&tpl.VersionID,
		&tpl.Subject,
		&tpl.HTML,
		&tpl.Text,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return ResolvedTemplate{}, ErrTemplateNotFound
	}

	if err != nil {
		return ResolvedTemplate{}, fmt.Errorf("query published template: %w", err)
	}

	return tpl, nil
}

func (r *Repository) CreateTemplate(
	ctx context.Context,
	tenantID string,
	key string,
	name string,
	description string,
	channel string,
	category string,
) (Template, error) {
	if channel == "" {
		channel = "email"
	}

	if category == "" {
		category = "transactional"
	}

	var tpl Template

	err := r.db.QueryRow(
		ctx,
		`
		INSERT INTO templates (
			tenant_id,
			key,
			name,
			description,
			channel,
			category
		)
		VALUES ($1::uuid, $2, $3, NULLIF($4, ''), $5, $6)
		RETURNING
			id::text,
			tenant_id::text,
			key,
			name,
			COALESCE(description, ''),
			channel,
			category,
			created_at,
			updated_at
		`,
		tenantID,
		key,
		name,
		description,
		channel,
		category,
	).Scan(
		&tpl.ID,
		&tpl.TenantID,
		&tpl.Key,
		&tpl.Name,
		&tpl.Description,
		&tpl.Channel,
		&tpl.Category,
		&tpl.CreatedAt,
		&tpl.UpdatedAt,
	)

	if err != nil {
		return Template{}, fmt.Errorf("create template: %w", err)
	}

	return tpl, nil
}

func (r *Repository) ListTemplates(
	ctx context.Context,
	tenantID string,
) ([]Template, error) {
	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			id::text,
			tenant_id::text,
			key,
			name,
			COALESCE(description, ''),
			channel,
			category,
			created_at,
			updated_at
		FROM templates
		WHERE tenant_id = $1::uuid
		ORDER BY created_at DESC
		`,
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("list templates query: %w", err)
	}
	defer rows.Close()

	var templates []Template

	for rows.Next() {
		var tpl Template

		err := rows.Scan(
			&tpl.ID,
			&tpl.TenantID,
			&tpl.Key,
			&tpl.Name,
			&tpl.Description,
			&tpl.Channel,
			&tpl.Category,
			&tpl.CreatedAt,
			&tpl.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan template row: %w", err)
		}

		templates = append(templates, tpl)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("iterate templates rows: %w", rows.Err())
	}

	if templates == nil {
		templates = []Template{}
	}

	return templates, nil
}

func (r *Repository) GetTemplateByID(
	ctx context.Context,
	tenantID string,
	templateID string,
) (Template, error) {
	var tpl Template

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id::text,
			tenant_id::text,
			key,
			name,
			COALESCE(description, ''),
			channel,
			category,
			created_at,
			updated_at
		FROM templates
		WHERE tenant_id = $1::uuid
		  AND id = $2::uuid
		`,
		tenantID,
		templateID,
	).Scan(
		&tpl.ID,
		&tpl.TenantID,
		&tpl.Key,
		&tpl.Name,
		&tpl.Description,
		&tpl.Channel,
		&tpl.Category,
		&tpl.CreatedAt,
		&tpl.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Template{}, ErrTemplateNotFound
	}

	if err != nil {
		return Template{}, fmt.Errorf("get template by id: %w", err)
	}

	versions, err := r.ListTemplateVersions(ctx, templateID)
	if err != nil {
		return Template{}, err
	}

	tpl.Versions = versions

	return tpl, nil
}

func (r *Repository) CreateTemplateVersion(
	ctx context.Context,
	tenantID string,
	templateID string,
	subject string,
	htmlBody string,
	textBody string,
) (TemplateVersion, error) {
	var exists bool

	err := r.db.QueryRow(
		ctx,
		`
		SELECT EXISTS(
			SELECT 1 FROM templates
			WHERE id = $1::uuid AND tenant_id = $2::uuid
		)
		`,
		templateID,
		tenantID,
	).Scan(&exists)

	if err != nil {
		return TemplateVersion{}, fmt.Errorf("check template existence: %w", err)
	}

	if !exists {
		return TemplateVersion{}, ErrTemplateNotFound
	}

	var nextVersion int

	err = r.db.QueryRow(
		ctx,
		`
		SELECT COALESCE(MAX(version), 0) + 1
		FROM template_versions
		WHERE template_id = $1::uuid
		`,
		templateID,
	).Scan(&nextVersion)

	if err != nil {
		return TemplateVersion{}, fmt.Errorf("calculate next version: %w", err)
	}

	var ver TemplateVersion

	err = r.db.QueryRow(
		ctx,
		`
		INSERT INTO template_versions (
			template_id,
			version,
			subject,
			html_body,
			text_body,
			status
		)
		VALUES (
			$1::uuid,
			$2,
			NULLIF($3, ''),
			NULLIF($4, ''),
			NULLIF($5, ''),
			'draft'
		)
		RETURNING
			id::text,
			template_id::text,
			version,
			COALESCE(subject, ''),
			COALESCE(html_body, ''),
			COALESCE(text_body, ''),
			status,
			created_at,
			updated_at
		`,
		templateID,
		nextVersion,
		subject,
		htmlBody,
		textBody,
	).Scan(
		&ver.ID,
		&ver.TemplateID,
		&ver.Version,
		&ver.Subject,
		&ver.HTMLBody,
		&ver.TextBody,
		&ver.Status,
		&ver.CreatedAt,
		&ver.UpdatedAt,
	)

	if err != nil {
		return TemplateVersion{}, fmt.Errorf("insert template version: %w", err)
	}

	return ver, nil
}

func (r *Repository) PublishTemplateVersion(
	ctx context.Context,
	tenantID string,
	templateID string,
	versionID string,
) (TemplateVersion, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return TemplateVersion{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var exists bool

	err = tx.QueryRow(
		ctx,
		`
		SELECT EXISTS(
			SELECT 1 FROM templates
			WHERE id = $1::uuid AND tenant_id = $2::uuid
		)
		`,
		templateID,
		tenantID,
	).Scan(&exists)

	if err != nil {
		return TemplateVersion{}, fmt.Errorf("check template existence: %w", err)
	}

	if !exists {
		return TemplateVersion{}, ErrTemplateNotFound
	}

	_, err = tx.Exec(
		ctx,
		`
		UPDATE template_versions
		SET status = 'archived', updated_at = now()
		WHERE template_id = $1::uuid
		  AND status = 'published'
		`,
		templateID,
	)
	if err != nil {
		return TemplateVersion{}, fmt.Errorf("archive existing published versions: %w", err)
	}

	var ver TemplateVersion

	err = tx.QueryRow(
		ctx,
		`
		UPDATE template_versions
		SET status = 'published', updated_at = now()
		WHERE id = $1::uuid AND template_id = $2::uuid
		RETURNING
			id::text,
			template_id::text,
			version,
			COALESCE(subject, ''),
			COALESCE(html_body, ''),
			COALESCE(text_body, ''),
			status,
			created_at,
			updated_at
		`,
		versionID,
		templateID,
	).Scan(
		&ver.ID,
		&ver.TemplateID,
		&ver.Version,
		&ver.Subject,
		&ver.HTMLBody,
		&ver.TextBody,
		&ver.Status,
		&ver.CreatedAt,
		&ver.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return TemplateVersion{}, ErrVersionNotFound
	}

	if err != nil {
		return TemplateVersion{}, fmt.Errorf("update published status: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return TemplateVersion{}, fmt.Errorf("commit tx: %w", err)
	}

	return ver, nil
}

func (r *Repository) ListTemplateVersions(
	ctx context.Context,
	templateID string,
) ([]TemplateVersion, error) {
	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			id::text,
			template_id::text,
			version,
			COALESCE(subject, ''),
			COALESCE(html_body, ''),
			COALESCE(text_body, ''),
			status,
			created_at,
			updated_at
		FROM template_versions
		WHERE template_id = $1::uuid
		ORDER BY version DESC
		`,
		templateID,
	)
	if err != nil {
		return nil, fmt.Errorf("list template versions query: %w", err)
	}
	defer rows.Close()

	var versions []TemplateVersion

	for rows.Next() {
		var ver TemplateVersion

		err := rows.Scan(
			&ver.ID,
			&ver.TemplateID,
			&ver.Version,
			&ver.Subject,
			&ver.HTMLBody,
			&ver.TextBody,
			&ver.Status,
			&ver.CreatedAt,
			&ver.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan template version row: %w", err)
		}

		versions = append(versions, ver)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("iterate template version rows: %w", rows.Err())
	}

	if versions == nil {
		versions = []TemplateVersion{}
	}

	return versions, nil
}
