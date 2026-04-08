package pdf_test

import (
	"bytes"
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/billing/pdf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPDFGenerator_BinaryValidation(t *testing.T) {
	// Check if chromium is available
	_, err := exec.LookPath("chromium")
	if err != nil {
		t.Skip("Chromium browser not found - skipping PDF generation test")
	}

	// Skip in short mode if chromedp takes too long, but we want this to run
	generator := pdf.NewPDFGenerator(15 * time.Second)

	html := `<html><body><h1>Binary Test</h1><p>Checking PDF Header</p></body></html>`
	opts := pdf.DefaultA4Portrait()

	ctx := context.Background()
	data, err := generator.GenerateFromHTML(ctx, html, opts)
	require.NoError(t, err, "Should generate PDF from HTML without error")

	// Validate binary prefix for PDF (%PDF-)
	assert.True(t, len(data) > 5, "PDF data is too short")
	assert.True(t, bytes.HasPrefix(data, []byte("%PDF-")), "Missing %PDF- header in generated output")
}
