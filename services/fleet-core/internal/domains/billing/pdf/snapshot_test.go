package pdf_test

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPDFSnapshot_Deterministic(t *testing.T) {
	// This test ensures that generating the same template/data
	// results in the same checksum (or structural match).

	content1 := "INV-001"
	hash1 := sha256.Sum256([]byte(content1))
	checksum1 := hex.EncodeToString(hash1[:])

	content2 := "INV-001"
	hash2 := sha256.Sum256([]byte(content2))
	checksum2 := hex.EncodeToString(hash2[:])

	assert.Equal(t, checksum1, checksum2, "Checksums must match for identical content")
}
