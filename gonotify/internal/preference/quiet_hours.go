package preference

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// QuietHours stores a user's notification quiet-hours window.
type QuietHours struct {
	Enabled   bool   `json:"enabled"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Timezone  string `json:"timezone"`
}

// DefaultQuietHours returns the default quiet-hours configuration, matching the
// database defaults: disabled with a 22:00-08:00 UTC window.
func DefaultQuietHours() QuietHours {
	return QuietHours{
		Enabled:   false,
		StartTime: "22:00",
		EndTime:   "08:00",
		Timezone:  "UTC",
	}
}

// GetQuietHours fetches the quiet-hours configuration for an internal user.
// When no row exists the default (disabled) configuration is returned.
func (r *Repository) GetQuietHours(ctx context.Context, userID string) (QuietHours, error) {
	var qh QuietHours

	err := r.db.QueryRow(
		ctx,
		`
		SELECT enabled, start_time, end_time, timezone
		FROM quiet_hours
		WHERE user_id = $1::uuid
		`,
		userID,
	).Scan(&qh.Enabled, &qh.StartTime, &qh.EndTime, &qh.Timezone)

	if errors.Is(err, pgx.ErrNoRows) {
		return DefaultQuietHours(), nil
	}

	if err != nil {
		return QuietHours{}, fmt.Errorf("query quiet hours: %w", err)
	}

	return qh, nil
}

// SetQuietHours upserts the quiet-hours configuration for an internal user.
func (r *Repository) SetQuietHours(ctx context.Context, userID string, qh QuietHours) error {
	_, err := r.db.Exec(
		ctx,
		`
		INSERT INTO quiet_hours (
			user_id,
			enabled,
			start_time,
			end_time,
			timezone
		)
		VALUES (
			$1::uuid,
			$2,
			$3,
			$4,
			$5
		)
		ON CONFLICT (user_id) DO UPDATE SET
			enabled = EXCLUDED.enabled,
			start_time = EXCLUDED.start_time,
			end_time = EXCLUDED.end_time,
			timezone = EXCLUDED.timezone,
			updated_at = now()
		`,
		userID,
		qh.Enabled,
		qh.StartTime,
		qh.EndTime,
		qh.Timezone,
	)

	if err != nil {
		return fmt.Errorf("upsert quiet hours: %w", err)
	}

	return nil
}

// parseHHMM parses an "HH:MM" string into hour and minute integers.
func parseHHMM(s string) (hour int, minute int, ok bool) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return 0, 0, false
	}

	h, err := strconv.Atoi(parts[0])
	if err != nil || h < 0 || h > 23 {
		return 0, 0, false
	}

	m, err := strconv.Atoi(parts[1])
	if err != nil || m < 0 || m > 59 {
		return 0, 0, false
	}

	return h, m, true
}

// minutesFromMidnight converts an "HH:MM" string to minutes since midnight.
func minutesFromMidnight(s string) (int, bool) {
	h, m, ok := parseHHMM(s)
	if !ok {
		return 0, false
	}

	return h*60 + m, true
}

// loadTimezone resolves a timezone identifier, falling back to UTC when the
// identifier is empty or invalid.
func loadTimezone(tz string) *time.Location {
	if tz == "" {
		return time.UTC
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		return time.UTC
	}

	return loc
}

// IsActive reports whether now falls within the quiet-hours window, honoring the
// configured timezone and midnight-spanning windows (e.g. 22:00-08:00).
// The end boundary is exclusive.
func (qh QuietHours) IsActive(now time.Time) bool {
	if !qh.Enabled {
		return false
	}

	start, ok1 := minutesFromMidnight(qh.StartTime)
	end, ok2 := minutesFromMidnight(qh.EndTime)
	if !ok1 || !ok2 {
		return false
	}

	loc := loadTimezone(qh.Timezone)
	nowLocal := now.In(loc)
	t := nowLocal.Hour()*60 + nowLocal.Minute()

	switch {
	case start < end:
		// Normal window within a single day (e.g. 08:00-22:00).
		return t >= start && t < end
	case start > end:
		// Midnight-spanning window (e.g. 22:00-08:00).
		return t >= start || t < end
	default:
		// start == end: degenerate window, treat as inactive.
		return false
	}
}

// End returns the UTC time at which the current quiet-hours window ends.
// It should only be called when IsActive returns true.
func (qh QuietHours) End(now time.Time) time.Time {
	if !qh.Enabled {
		return now.UTC()
	}

	start, ok1 := minutesFromMidnight(qh.StartTime)
	endMin, ok2 := minutesFromMidnight(qh.EndTime)
	if !ok1 || !ok2 {
		return now.UTC()
	}

	loc := loadTimezone(qh.Timezone)
	nowLocal := now.In(loc)
	t := nowLocal.Hour()*60 + nowLocal.Minute()

	y, m, day := nowLocal.Date()
	end := time.Date(y, m, day, endMin/60, endMin%60, 0, 0, loc)

	if start > endMin && t >= start {
		// Midnight-spanning window and we are in the first segment
		// (after start, before midnight). The end boundary is tomorrow.
		end = end.Add(24 * time.Hour)
	}

	return end.UTC()
}

// Validate checks that the quiet-hours configuration is well-formed.
func (qh QuietHours) Validate() error {
	if !qh.Enabled {
		return nil
	}

	if _, _, ok := parseHHMM(qh.StartTime); !ok {
		return errors.New("invalid start_time: expected HH:MM")
	}

	if _, _, ok := parseHHMM(qh.EndTime); !ok {
		return errors.New("invalid end_time: expected HH:MM")
	}

	return validateTimezone(qh.Timezone)
}

func validateTimezone(tz string) error {
	if tz == "" || tz == "UTC" || tz == "Universal" {
		return nil
	}

	if _, err := time.LoadLocation(tz); err != nil {
		return fmt.Errorf("invalid timezone: %s", tz)
	}

	return nil
}
