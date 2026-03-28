//go:build integration

package storage_test

import (
	"context"
	"testing"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/billing/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/billing/storage"
	"github.com/flyfleet/fleet-core-service/internal/domains/billing/testutil"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestS3Storage_MinioIntegration(t *testing.T) {
	ctx := context.Background()
	endpoint, accessKey, secretKey, terminate := testutil.SetupMinioContainer(ctx, t)
	defer terminate()

	client, err := storage.NewMinioClient(endpoint, accessKey, secretKey, false)
	require.NoError(t, err)

	bucketName := "test-invoices"
	err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	require.NoError(t, err)

	store := storage.NewS3Storage(client, bucketName, "us-east-1")

	inv := &models.Invoice{
		ID:            uuid.New(),
		OrgID:         uuid.New(),
		InvoiceNumber: "INV-MINIO-001",
		InvoiceDate:   time.Now(),
	}

	pdfData := []byte("%PDF-1.4\n%...\nMinIO Integration PDF Data")

	// 1. Upload Test
	err = store.Save(ctx, inv, pdfData)
	require.NoError(t, err, "Should save PDF to MinIO without error")

	// 2. Retrieval Test
	retrievedData, err := store.Get(ctx, inv)
	require.NoError(t, err, "Should retrieve PDF from MinIO without error")
	assert.Equal(t, pdfData, retrievedData, "Retrieved data should match uploaded data")

	// 3. Presigned URL Test
	urlStr, err := store.GetPresignedURL(ctx, inv, 1*time.Hour)
	require.NoError(t, err, "Should generate presigned URL without error")
	assert.Contains(t, urlStr, endpoint, "Presigned URL should contain MinIO endpoint")
	assert.Contains(t, urlStr, bucketName, "Presigned URL should contain bucket name")
	assert.Contains(t, urlStr, inv.InvoiceNumber, "Presigned URL should contain invoice number")
}

func TestS3Storage_SignedURLExpiration(t *testing.T) {
	ctx := context.Background()
	endpoint, accessKey, secretKey, terminate := testutil.SetupMinioContainer(ctx, t)
	defer terminate()

	client, err := storage.NewMinioClient(endpoint, accessKey, secretKey, false)
	require.NoError(t, err)

	bucketName := "test-expiry"
	err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	require.NoError(t, err)

	store := storage.NewS3Storage(client, bucketName, "us-east-1")

	inv := &models.Invoice{
		ID:            uuid.New(),
		OrgID:         uuid.New(),
		InvoiceNumber: "INV-EXP-TEST",
		InvoiceDate:   time.Now(),
	}

	// 1. Upload Test File
	err = store.Save(ctx, inv, []byte("test"))
	require.NoError(t, err)

	// 2. Generate URLs with different expiries
	duration := 5 * time.Minute
	urlStr, err := store.GetPresignedURL(ctx, inv, duration)
	require.NoError(t, err)

	// Ensure X-Amz-Expires matches (5 min = 300 seconds)
	assert.Contains(t, urlStr, "X-Amz-Expires=300")

	// Create another URL that expires immediately
	// Minio/S3 requires min 1 second expiry
	shortDuration := 1 * time.Second
	shortUrlStr, err := store.GetPresignedURL(ctx, inv, shortDuration)
	require.NoError(t, err)
	assert.Contains(t, shortUrlStr, "X-Amz-Expires=1")
}
