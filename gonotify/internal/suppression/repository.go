package suppression

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrSuppressionNotFound = errors.New("suppression not found")
	ErrDuplicateSuppression = errors.New("suppression already exists")
)

var e164Regex = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)

// ValidChannels returns the set of channels that can be suppressed.
func ValidChannels() map[string]bool {
	return map[string]bool{
		"email":  true,
		"sms":    true,
		"push":   true,
		"in_app": true,
	}
}

// IsValidChannel reports whether the given channel can be suppressed.
func IsValidChannel(channel string) bool {
	return ValidChannels()[channel]
}

// Suppression represents a suppression record.
type Suppression struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Channel   string    `json:"channel"`
	Address   string    `json:"address"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// List returns all suppressions for the given tenant.
func (r *Repository) List(ctx context.Context, tenantID string) ([]Suppression, error) {
	rows, err := r.db.Query(
		ctx,
		`
		SELECT id::text, channel, address, reason, created_at
		FROM suppressions
		WHERE tenant_id = $1::uuid
		ORDER BY created_at DESC
		`,
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("query suppressions: %w", err)
	}
	defer rows.Close()

	list := make([]Suppression, 0)

	for rows.Next() {
		var s Suppression
		if err := rows.Scan(
			&s.ID,
			&s.Channel,
			&s.Address,
			&s.Reason,
			&s.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan suppression: %w", err)
		}

		list = append(list, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate suppressions: %w", err)
	}

	return list, nil
}

// Get fetches a single suppression by ID. Returns ErrSuppressionNotFound if
// the suppression does not exist or does not belong to the given tenant.
func (r *Repository) Get(
	ctx context.Context,
	tenantID string,
	id string,
) (*Suppression, error) {
	var s Suppression

	err := r.db.QueryRow(
		ctx,
		`
		SELECT id::text, channel, address, reason, created_at
		FROM suppressions
		WHERE tenant_id = $1::uuid
		  AND id = $2::uuid
		`,
		tenantID,
		id,
	).Scan(&s.ID, &s.Channel, &s.Address, &s.Reason, &s.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrSuppressionNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("query suppression: %w", err)
	}

	s.TenantID = tenantID

	return &s, nil
}

// Create inserts a new suppression. Returns ErrDuplicateSuppression if the
// exact (tenant_id, channel, address) tuple already exists.
func (r *Repository) Create(
	ctx context.Context,
	tenantID string,
	channel string,
	address string,
	reason string,
) (*Suppression, error) {
	normalizedAddr := normalizeAddress(channel, address)

	_, err := r.db.Exec(
		ctx,
		`
		INSERT INTO suppressions (
			tenant_id,
			channel,
			address,
			reason
		)
		VALUES (
			$1::uuid,
			$2,
			$3,
			$4
		)
		`,
		tenantID,
		channel,
		normalizedAddr,
		reason,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDuplicateSuppression
		}

		return nil, fmt.Errorf("insert suppression: %w", err)
	}

	var s Suppression

	err = r.db.QueryRow(
		ctx,
		`
		SELECT id::text, channel, address, reason, created_at
		FROM suppressions
		WHERE tenant_id = $1::uuid
		  AND channel = $2
		  AND address = $3
		`,
		tenantID,
		channel,
		normalizedAddr,
	).Scan(&s.ID, &s.Channel, &s.Address, &s.Reason, &s.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("query created suppression: %w", err)
	}

	s.TenantID = tenantID

	return &s, nil
}

// Delete removes a suppression by ID. Returns ErrSuppressionNotFound if the
// suppression does not exist or does not belong to the given tenant.
func (r *Repository) Delete(ctx context.Context, tenantID string, id string) error {
	result, err := r.db.Exec(
		ctx,
		`
		DELETE FROM suppressions
		WHERE tenant_id = $1::uuid
		  AND id = $2::uuid
		`,
		tenantID,
		id,
	)
	if err != nil {
		return fmt.Errorf("delete suppression: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrSuppressionNotFound
	}

	return nil
}

// IsSuppressed reports whether the given address is suppressed for the
// specified tenant and channel.
func (r *Repository) IsSuppressed(
	ctx context.Context,
	tenantID string,
	channel string,
	address string,
) (bool, error) {
	normalizedAddr := normalizeAddress(channel, address)

	var id string

	err := r.db.QueryRow(
		ctx,
		`
		SELECT id::text
		FROM suppressions
		WHERE tenant_id = $1::uuid
		  AND channel = $2
		  AND address = $3
		`,
		tenantID,
		channel,
		normalizedAddr,
	).Scan(&id)

	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("query suppression: %w", err)
	}

	return true, nil
}

// ValidateAddress checks that an address is appropriate for the given channel.
func ValidateAddress(channel string, address string) error {
	if address == "" {
		return errors.New("address is required")
	}

	switch channel {
	case "email":
		if !strings.Contains(address, "@") {
			return fmt.Errorf("invalid email address: %s", address)
		}

		if _, err := mail.ParseAddress(address); err != nil {
			return fmt.Errorf("invalid email address: %s", address)
		}

	case "sms":
		normalized := strings.TrimSpace(address)
		normalized = strings.NewReplacer(" ", "", "-", "", ".", "", "(", "", ")", "").Replace(normalized)
		if !e164Regex.MatchString(normalized) {
			return fmt.Errorf("invalid phone number (E.164): %s", address)
		}

	case "push", "in_app":
		// Any non-empty value is acceptable.
	}

	return nil
}

// normalizeAddress lowercases email addresses so that duplicate inserts
// for the same recipient are prevented regardless of casing.
func normalizeAddress(channel string, address string) string {
	if channel == "email" {
		return strings.ToLower(strings.TrimSpace(address))
	}

	return strings.TrimSpace(address)
}

// isUniqueViolation reports whether the error is a PostgreSQL unique
// constraint violation.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}

	type pgxErr interface {
		SQLState() string
	}

	// Walk the error chain looking for a SQLState of 23505.
	for e := err; e != nil; e = errors.Unwrap(e) {
		if pe, ok := e.(pgxErr); ok && pe.SQLState() == "23505" {
			return true
		}
	}

	// Fall back to string matching for robustness.
	msg := err.Error()
	return strings.Contains(msg, "unique_violation") ||
		strings.Contains(msg, "duplicate key") ||
		strings.Contains(msg, "23505")
}
