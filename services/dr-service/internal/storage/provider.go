package storage

import (
	"context"
	"io"
)

// BackupProvider defines the interface for storing and retrieving backup data.
type BackupProvider interface {
	// Upload uploads data from the given reader to the storage provider.
	// Returns the location (e.g., S3 URL, file path) or an error.
	Upload(ctx context.Context, key string, data io.Reader) (string, error)

	// Download retrieves data for the given key.
	// Returns a ReadCloser that must be closed by the caller.
	Download(ctx context.Context, key string) (io.ReadCloser, error)

	// Delete removes the data at the given key.
	Delete(ctx context.Context, key string) error

	// List returns a list of keys with the given prefix.
	List(ctx context.Context, prefix string) ([]string, error)
}
