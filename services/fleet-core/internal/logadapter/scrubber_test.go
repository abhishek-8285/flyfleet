package logadapter

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestMaskSensitive(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no token",
			input:    "starting server on port 8080",
			expected: "starting server on port 8080",
		},
		{
			name:     "single token",
			input:    "received api key flc_1234567890abcdef",
			expected: "received api key flc_1234****",
		},
		{
			name:     "short token",
			input:    "key flc_abc",
			expected: "key flc_****",
		},
		{
			name:     "multiple tokens",
			input:    "keys flc_aaa and flc_bbb",
			expected: "keys flc_**** and flc_****",
		},
		{
			name:     "token in path",
			input:    "GET /api/v1/resource?token=flc_longtokensecret",
			expected: "GET /api/v1/resource?token=flc_long****",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, maskSensitive(tt.input))
		})
	}
}

func TestScrubberCore(t *testing.T) {
	var buf bytes.Buffer
	
	// Create a base core that writes to a buffer
	encoder := zapcore.NewJSONEncoder(zap.NewDevelopmentEncoderConfig())
	baseCore := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zap.DebugLevel)
	
	// Wrap it with our scrubber
	scrubber := NewScrubberCore(baseCore)
	logger := zap.New(scrubber)

	// Log a message with a token
	logger.Info("user login", zap.String("api_key", "flc_secret_token_123"))

	output := buf.String()
	assert.NotContains(t, output, "flc_secret_token_123")
	assert.Contains(t, output, "flc_secr****")
}
