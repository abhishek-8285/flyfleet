package pdf_test

import (
	"testing"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/billing/pdf"
	"github.com/stretchr/testify/assert"
)

func TestTemplateService_SanitizeHTML(t *testing.T) {
	s := pdf.NewTemplateService(nil)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean html",
			input:    "<div>Hello</div>",
			expected: "<div>Hello</div>",
		},
		{
			name:     "script tag",
			input:    "<div><script>alert(1)</script></div>",
			expected: "<div>alert(1)</div>", // script tags are removed, content remains
		},
		{
			name:     "iframe tag",
			input:    "<iframe src='malicious.com'></iframe>",
			expected: "", // iframe tag and content removed entirely (it's a void element)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.SanitizeHTML(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTemplateService_SafeFuncMap(t *testing.T) {
	s := pdf.NewTemplateService(nil)
	funcs := s.SafeFuncMap()

	t.Run("formatCurrency", func(t *testing.T) {
		f := funcs["formatCurrency"].(func(float64, string) string)
		assert.Equal(t, "USD 100.50", f(100.5, "USD"))
	})

	t.Run("formatDate", func(t *testing.T) {
		f := funcs["formatDate"].(func(time.Time) string)
		testTime := time.Date(2026, 3, 6, 0, 0, 0, 0, time.UTC)
		assert.Equal(t, "Mar 06, 2026", f(testTime))
	})

	t.Run("add", func(t *testing.T) {
		f := funcs["add"].(func(int, int) int)
		assert.Equal(t, 15, f(10, 5))
	})
}
