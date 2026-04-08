package pdf_test

import (
	"testing"

	"github.com/flyfleet/fleet-core/internal/domains/billing/pdf"
	"github.com/stretchr/testify/assert"
)

func TestTemplateValidator_CheckSchema(t *testing.T) {
	v := pdf.NewTemplateValidator()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid field access",
			input:   "<h1>{{.Invoice.InvoiceNumber}}</h1>",
			wantErr: false,
		},
		{
			name:    "valid range list",
			input:   "{{range .Items}}<p>{{.Description}}</p>{{end}}",
			wantErr: false,
		},
		{
			name:    "invalid syntax",
			input:   "<h1>{{.Invoice.Number</h1>",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.CheckSchema(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
