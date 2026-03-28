package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/billing/models"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// ObjectStorage defines the interface for persisting documents (S3, Minio, Local)
type ObjectStorage interface {
	Save(ctx context.Context, invoice *models.Invoice, data []byte) error
	Get(ctx context.Context, invoice *models.Invoice) ([]byte, error)
	GetPresignedURL(ctx context.Context, invoice *models.Invoice, expires time.Duration) (string, error)
	Delete(ctx context.Context, invoice *models.Invoice) error
	GetStorageStats(ctx context.Context) (*StorageStats, error)
	CleanupOldFiles(ctx context.Context, olderThan time.Duration) (int, error)
	HealthCheck(ctx context.Context) error
	VerifyChecksum(ctx context.Context, invoice *models.Invoice, expectedChecksum string) (bool, error)
}

// StorageStats represents storage statistics
type StorageStats struct {
	TotalSizeBytes int64   `json:"total_size_bytes"`
	TotalSizeMB    float64 `json:"total_size_mb"`
	FileCount      int64   `json:"file_count"`
	BasePath       string  `json:"base_path"`
}

// S3Storage implementation using minio-go
type S3Storage struct {
	bucket string
	region string
	client *minio.Client
}

func NewS3Storage(client *minio.Client, bucket, region string) *S3Storage {
	return &S3Storage{
		bucket: bucket,
		region: region,
		client: client,
	}
}

// NewMinioClient is a helper to create the client
func NewMinioClient(endpoint, accessKeyID, secretAccessKey string, useSSL bool) (*minio.Client, error) {
	return minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
}

func (s *S3Storage) getKey(invoice *models.Invoice) string {
	year := invoice.InvoiceDate.Format("2006")
	month := invoice.InvoiceDate.Format("01")
	return fmt.Sprintf("invoices/%s/%s/%s/%s.pdf", invoice.OrgID.String(), year, month, invoice.InvoiceNumber)
}

func (s *S3Storage) Save(ctx context.Context, invoice *models.Invoice, data []byte) error {
	key := s.getKey(invoice)

	reader := bytes.NewReader(data)
	_, err := s.client.PutObject(ctx, s.bucket, key, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: "application/pdf",
	})
	return err
}

func (s *S3Storage) Get(ctx context.Context, invoice *models.Invoice) ([]byte, error) {
	key := s.getKey(invoice)

	object, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer object.Close()

	stat, err := object.Stat()
	if err != nil {
		return nil, err
	}

	data := make([]byte, stat.Size)
	_, err = object.Read(data)
	if err != nil && err.Error() != "EOF" {
		return nil, err
	}

	return data, nil
}

func (s *S3Storage) GetPresignedURL(ctx context.Context, invoice *models.Invoice, expires time.Duration) (string, error) {
	key := s.getKey(invoice)
	url, err := s.client.PresignedGetObject(ctx, s.bucket, key, expires, nil)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}

func (s *S3Storage) GetStorageStats(ctx context.Context) (*StorageStats, error) {
	return &StorageStats{}, nil
}

func (s *S3Storage) CleanupOldFiles(ctx context.Context, olderThan time.Duration) (int, error) {
	return 0, nil
}

func (s *S3Storage) Delete(ctx context.Context, invoice *models.Invoice) error {
	key := s.getKey(invoice)
	return s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
}

func (s *S3Storage) HealthCheck(ctx context.Context) error {
	_, err := s.client.BucketExists(ctx, s.bucket)
	return err
}

func (s *S3Storage) VerifyChecksum(ctx context.Context, invoice *models.Invoice, expectedChecksum string) (bool, error) {
	data, err := s.Get(ctx, invoice)
	if err != nil {
		return false, err
	}

	h := sha256.New()
	h.Write(data)
	actualChecksum := hex.EncodeToString(h.Sum(nil))

	return actualChecksum == expectedChecksum, nil
}
