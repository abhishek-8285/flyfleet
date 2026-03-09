package service_test

import (
	"encoding/json"
	"testing"
)

type MaskingService interface {
	MaskData(data interface{}, rules json.RawMessage) interface{}
}

type dummyMasker struct{}

func (m *dummyMasker) MaskData(data interface{}, rules json.RawMessage) interface{} {
	// Dummy implementation that just returns "MASKED" for any string
	return "MASKED"
}

func TestMaskingService(t *testing.T) {
	masker := &dummyMasker{}
	
	rules := json.RawMessage(`{"pii": true}`)
	original := "sensitive_email@example.com"
	
	result := masker.MaskData(original, rules)
	if result != "MASKED" {
		t.Errorf("Expected 'MASKED', got %v", result)
	}
}
