package utils

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
)

// CompressJSON marshals the given struct to JSON and compresses it using Gzip
func CompressJSON(data interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal to JSON: %w", err)
	}

	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	
	if _, err := gz.Write(jsonData); err != nil {
		_ = gz.Close()
		return nil, fmt.Errorf("failed to write compressed data: %w", err)
	}
	
	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}
	
	return b.Bytes(), nil
}

// DecompressJSON decompresses Gzip data and unmarshals it into the given interface
func DecompressJSON(compressedData []byte, v interface{}) error {
	b := bytes.NewReader(compressedData)
	gz, err := gzip.NewReader(b)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gz.Close()

	decompressedData, err := io.ReadAll(gz)
	if err != nil {
		return fmt.Errorf("failed to read decompressed data: %w", err)
	}

	if err := json.Unmarshal(decompressedData, v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}
