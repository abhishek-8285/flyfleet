package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

type AzureProvider struct {
	client    *azblob.Client
	container string
}

func NewAzureProvider(connectionString, container string) (*AzureProvider, error) {
	client, err := azblob.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid connection string: %w", err)
	}
	return &AzureProvider{client: client, container: container}, nil
}

func (a *AzureProvider) Upload(ctx context.Context, key string, data io.Reader) (string, error) {
	// Azure SDK requires ReadSeeker or Bytes. Buffering for simplicity as io.Reader isn't seekable.
	// For large files, we'd want a streaming approach, but this suffices for Phase 2 proof.
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, data); err != nil {
		return "", err
	}

	_, err := a.client.UploadBuffer(ctx, a.container, key, buf.Bytes(), nil)
	if err != nil {
		return "", fmt.Errorf("failed to upload to Azure: %w", err)
	}
	return fmt.Sprintf("azure://%s/%s", a.container, key), nil
}

func (a *AzureProvider) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	resp, err := a.client.DownloadStream(ctx, a.container, key, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to download from Azure: %w", err)
	}
	return resp.Body, nil
}

func (a *AzureProvider) Delete(ctx context.Context, key string) error {
	_, err := a.client.DeleteBlob(ctx, a.container, key, nil)
	if err != nil {
		return fmt.Errorf("failed to delete from Azure: %w", err)
	}
	return nil
}

func (a *AzureProvider) List(ctx context.Context, prefix string) ([]string, error) {
	var keys []string
	pager := a.client.NewListBlobsFlatPager(a.container, &azblob.ListBlobsFlatOptions{
		Prefix: &prefix,
	})

	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list blobs: %w", err)
		}
		for _, blob := range resp.Segment.BlobItems {
			keys = append(keys, *blob.Name)
		}
	}
	return keys, nil
}
