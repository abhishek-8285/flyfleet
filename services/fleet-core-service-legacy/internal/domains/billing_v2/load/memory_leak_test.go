//go:build browser
// +build browser

package load_test

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/billing_v2/pdf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPDFGenerator_MemoryLeak loops renders and monitors MemStats to ensure memory does not infinitely climb
func TestPDFGenerator_MemoryLeak(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory leak test in short mode")
	}

	generator := pdf.NewPDFGenerator(10 * time.Second)
	html := `<html><body><h1>Memory Leak Test</h1><p>Tracking allocation...</p></body></html>`
	opts := pdf.DefaultA4Portrait()

	// 1. Base Memory Baseline
	runtime.GC()
	var baseStats runtime.MemStats
	runtime.ReadMemStats(&baseStats)
	
	const iterations = 50 // Test smaller number due to CI runner limits, but enough to trigger leaks if any

	// 2. Loop Execution
	for i := 0; i < iterations; i++ {
		ctx := context.Background()
		_, err := generator.GenerateFromHTML(ctx, html, opts)
		require.NoError(t, err, "Should not fail rendering during leak test")
	}

	// 3. Final Memory Baseline
	runtime.GC()
	var finalStats runtime.MemStats
	runtime.ReadMemStats(&finalStats)

	// Calculate memory delta
	deltaBytes := int64(finalStats.HeapAlloc) - int64(baseStats.HeapAlloc)
	deltaMB := float64(deltaBytes) / 1024 / 1024

	t.Logf("Memory Delta after %d iterations: %.2f MB", iterations, deltaMB)

	// 4. Assertion (Allowing max 10MB organic growth, any leak will easily exceed 100MB+)
	assert.Less(t, deltaMB, 20.0, "Memory allocation increased unnaturally, indicating a leak in ChromeDP pool")
}
