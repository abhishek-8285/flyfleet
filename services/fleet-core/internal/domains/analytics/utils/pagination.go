package utils

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"
)

// EncodeCursor encodes a created_at timestamp and an ID into a base64 cursor string.
// Format: "timestamp_rfc3339nano|id"
func EncodeCursor(createdAt time.Time, id string) string {
	raw := fmt.Sprintf("%s|%s", createdAt.Format(time.RFC3339Nano), id)
	return base64.StdEncoding.EncodeToString([]byte(raw))
}

// DecodeCursor parses a base64 cursor string back into a timestamp and an ID.
func DecodeCursor(cursor string) (time.Time, string, error) {
	if cursor == "" {
		return time.Time{}, "", fmt.Errorf("empty cursor")
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, "", fmt.Errorf("invalid cursor encoding: %w", err)
	}

	parts := strings.Split(string(decodedBytes), "|")
	if len(parts) != 2 {
		return time.Time{}, "", fmt.Errorf("invalid cursor format")
	}

	createdAt, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, "", fmt.Errorf("invalid cursor time format: %w", err)
	}

	return createdAt, parts[1], nil
}
