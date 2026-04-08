package pdf

import (
	"context"
	"github.com/flyfleet/fleet-core/internal/domains/billing/models"
)

// InvoiceEngine is the universal interface for any PDF generator
type InvoiceEngine interface {
	Generate(ctx context.Context, invoice *models.Invoice, items []models.InvoiceItem) ([]byte, error)
	Name() string
}

// Factory to create the requested engine
func NewEngineFactory(engineType string, chromeTimeout int) InvoiceEngine {
	if engineType == "high_speed" {
		return &NativeVectorEngineWrapper{}
	}
	// Default to Premium (Chromium)
	return &ChromiumEngineWrapper{
		generator: NewPDFGenerator(60),
	}
}

// --- ENGINE 1: Premium (Chromium) ---
type ChromiumEngineWrapper struct {
	generator *PDFGenerator
}

func (e *ChromiumEngineWrapper) Name() string { return "Chromium Premium" }
func (e *ChromiumEngineWrapper) Generate(ctx context.Context, invoice *models.Invoice, items []models.InvoiceItem) ([]byte, error) {
	html := GenerateInvoiceHTML(invoice, items)
	return e.generator.GenerateFromHTML(ctx, html, DefaultA4Portrait())
}

// --- ENGINE 2: High Speed (Native Go Vector) ---
// Note: This will be implemented using the vectorengine package we built
type NativeVectorEngineWrapper struct{}

func (e *NativeVectorEngineWrapper) Name() string { return "Native Vector High-Speed" }
func (e *NativeVectorEngineWrapper) Generate(ctx context.Context, invoice *models.Invoice, items []models.InvoiceItem) ([]byte, error) {
	// Logic to map models.Invoice to the Vector Layout JSON
	return nil, nil // Placeholder for now
}
