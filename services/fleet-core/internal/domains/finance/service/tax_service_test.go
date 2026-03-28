package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateGST(t *testing.T) {
	s := NewTaxCalculationService()

	t.Run("Intra-state calculation", func(t *testing.T) {
		amount := 1000.0
		rate := 18.0
		result := s.CalculateGST(amount, rate, false)

		assert.Equal(t, 1180.0, result.TotalAmount)
		assert.Equal(t, 180.0, result.GSTAmount)
		assert.Equal(t, 90.0, result.CGST)
		assert.Equal(t, 90.0, result.SGST)
		assert.Equal(t, 0.0, result.IGST)
		assert.Equal(t, 1000.0, result.BaseAmount)
	})

	t.Run("Inter-state calculation", func(t *testing.T) {
		amount := 1000.0
		rate := 18.0
		result := s.CalculateGST(amount, rate, true)

		assert.Equal(t, 1180.0, result.TotalAmount)
		assert.Equal(t, 180.0, result.GSTAmount)
		assert.Equal(t, 0.0, result.CGST)
		assert.Equal(t, 0.0, result.SGST)
		assert.Equal(t, 180.0, result.IGST)
		assert.Equal(t, 1000.0, result.BaseAmount)
	})
}

func TestValidateGSTIN(t *testing.T) {
	s := NewTaxCalculationService()

	tests := []struct {
		gstin    string
		expected bool
	}{
		{"27AAAAA0000A1Z5", true},
		{"07AAAAA0000A1Z5", true},
		{"12345", false},
		{"invalidgstin123", false},
		{"", false},
		{"27AAAAA0000A1Z", false},   // too short
		{"27AAAAA0000A1Z55", false}, // too long
		{" 27AAAAA0000A1Z5 ", true}, // whitespace handling
	}

	for _, tt := range tests {
		result := s.ValidateGSTIN(tt.gstin)
		assert.Equal(t, tt.expected, result, "GSTIN: %s", tt.gstin)
	}
}
