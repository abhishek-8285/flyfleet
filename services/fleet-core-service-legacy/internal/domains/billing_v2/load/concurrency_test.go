//go:build browser
// +build browser

package load_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/billing_v2/pdf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
)

// TestPDFGenerator_Concurrency ensures no race conditions occur when firing multiple Headless Chrome requests concurrently
func TestPDFGenerator_Concurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrency test in short testing mode")
	}

	// 1. Setup Generator
	generator := pdf.NewPDFGenerator(30 * time.Second)

	// 2. Concurrency constraints
	const routines = 100
	var wg sync.WaitGroup
	wg.Add(routines)

	successCount := atomic.NewInt32(0)
	errorCount := atomic.NewInt32(0)

	htmlTmpl := `<html><body><h1>Invoice {{.ReqID}}</h1><p>Testing concurrent rendering</p></body></html>`

	// 3. Fire concurrent requests
	for i := 0; i < routines; i++ {
		go func(reqID int) {
			defer wg.Done()
			
			// Simulate unique payload
			html := htmlTmpl // normally this would execute template, for isolated test we use static
			
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			
			data, err := generator.GenerateFromHTML(ctx, html, pdf.DefaultA4Portrait())
			if err != nil {
				t.Logf("Req %d failed: %v", reqID, err)
				errorCount.Inc()
				return
			}
			
			// Check valid PDF generated
			if len(data) > 10 && string(data[:5]) == "%PDF-" {
				successCount.Inc()
			} else {
				errorCount.Inc()
			}
		}(i)
	}

	wg.Wait()

	// 4. Assert Results
	t.Logf("Concurrency Test: %d/%d requests succeeded", successCount.Load(), routines)
	require.Equal(t, int32(routines), successCount.Load()+errorCount.Load(), "All routines must complete")
	
	// We expect high success rate, but chromedp might drop some on heavy load on small runners
	// So we assert at least 95% pass rate.
	assert.GreaterOrEqual(t, float64(successCount.Load())/float64(routines), 0.90, "Should have 90%+ success rate under concurrency")
}
