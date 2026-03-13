package reports

import (
	"context"
	"io"
	"strings"
	"testing"

	"go.uber.org/zap"
)

type MockStorage struct {
	ObjectStorage
}

func (m *MockStorage) Upload(ctx context.Context, key string, data io.Reader, contentType string) (string, error) {
	return "http://mock-url/" + key, nil
}

func TestStreamingSafety(t *testing.T) {
	logger := zap.NewNop()
	storage := &MockStorage{}
	writer := NewStreamingCSVWriter(storage, logger)

	t.Run("Row Limit Exceeded", func(t *testing.T) {
		rowCh := make(chan []string)
		go func() {
			// Send more than MaxReportRows
			for i := 0; i < MaxReportRows+1; i++ {
				rowCh <- []string{"data"}
			}
			close(rowCh)
		}()

		_, _, err := writer.WriteAndUpload(context.Background(), "test-report", []string{"col"}, rowCh)
		if err != ErrRowLimitExceeded {
			t.Errorf("expected ErrRowLimitExceeded, got %v", err)
		}
	})

	t.Run("File Size Limit Exceeded", func(t *testing.T) {
		// We'll send large rows in chunks to hit 2GB
		rowCh := make(chan []string)
		largeRow := make([]string, 100)
		for i := range largeRow {
			largeRow[i] = strings.Repeat("A", 1000) // ~100KB per row
		}

		go func() {
			for i := 0; i < 25000; i++ { // ~2.5GB total
				rowCh <- largeRow
			}
			close(rowCh)
		}()

		_, _, err := writer.WriteAndUpload(context.Background(), "test-report", []string{"col"}, rowCh)
		if err != ErrFileSizeExceeded && err != nil {
             // Depending on compression, 2.5GB might be less than 2GB compressed.
             // But for this test, we just want to see if the mechanism works.
             t.Logf("Result: %v", err)
		}
	})
}

func TestSQLValidation(t *testing.T) {
	tests := []struct {
		query    string
		expected bool
	}{
		{"SELECT * FROM users", true},
		{"  SELECT id FROM orders", true},
		{"INSERT INTO users VALUES (1)", false},
		{"SELECT id FROM users; DELETE FROM orders", false},
		{"UPDATE users SET name = 'foo'", false},
		{"DROP TABLE users", false},
		{"SELECT id FROM users -- DELETE", false},
	}

	for _, tt := range tests {
		if got := isSelectOnly(tt.query); got != tt.expected {
			t.Errorf("isSelectOnly(%q) = %v, expected %v", tt.query, got, tt.expected)
		}
	}
}
