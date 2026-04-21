package storage_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"testing"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/billing/models"
	"github.com/flyfleet/fleet-core/internal/domains/billing/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestLocalStorage_CorruptionDetection(t *testing.T) {
	logger := zaptest.NewLogger(t)
	basePath := t.TempDir()

	store, err := storage.NewLocalStorage(basePath, logger)
	require.NoError(t, err)

	ctx := context.Background()
	inv := &models.Invoice{
		ID:            uuid.New(),
		TenantID:         uuid.New(),
		InvoiceNumber: "INV-CORRUPT-TEST",
		InvoiceDate:   time.Now(),
	}

	pdfData := []byte("%PDF-1.4\n%...\nHello World PDF Data")

	err = store.Save(ctx, inv, pdfData)
	require.NoError(t, err)

	// Calculate expected checksum
	h := sha256.New()
	h.Write(pdfData)
	expectedChecksum := hex.EncodeToString(h.Sum(nil))

	// Verify intact checksum
	valid, err := store.VerifyChecksum(ctx, inv, expectedChecksum)
	require.NoError(t, err, "Should verify intact checksum without error")
	assert.True(t, valid, "Checksum should be valid for intact file")

	// Corrupt the file
	filePath := store.GetPath(inv)
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	require.NoError(t, err, "Should open file for corruption")
	_, err = f.WriteString("corrupted")
	require.NoError(t, err, "Should write corrupted data")
	f.Close()

	// Verify checksum again
	valid, err = store.VerifyChecksum(ctx, inv, expectedChecksum)
	require.NoError(t, err, "Should handle verification of corrupted file without error")
	assert.False(t, valid, "Checksum should be invalid after file corruption")
}
