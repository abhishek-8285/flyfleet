package pdfoptimizer

import (
	"bytes"
	"fmt"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// Optimize takes raw PDF bytes and performs extreme industrial compression
func Optimize(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}

	// Read the original bytes
	rs := bytes.NewReader(data)

	// Configure for maximum "squeezing"
	conf := model.NewDefaultConfiguration()

	var buf bytes.Buffer
	// Use pdfcpu API to optimize the buffer
	err := api.Optimize(rs, &buf, conf)
	if err != nil {
		return nil, fmt.Errorf("pdfcpu optimization failed: %w", err)
	}

	return buf.Bytes(), nil
}
