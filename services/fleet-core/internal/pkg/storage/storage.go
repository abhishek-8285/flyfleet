package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// StorageProvider defines the interface for file storage operations.
type StorageProvider interface {
	Save(ctx context.Context, path string, reader io.Reader) (string, error)
	Delete(ctx context.Context, path string) error
	GetURL(ctx context.Context, path string) (string, error)
}

// LocalStorage implements StorageProvider for the local filesystem.
type LocalStorage struct {
	BaseDir string
	BaseURL string
}

func NewLocalStorage(baseDir, baseURL string) (*LocalStorage, error) {
	// Ensure base directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}
	return &LocalStorage{BaseDir: baseDir, BaseURL: baseURL}, nil
}

func (s *LocalStorage) Save(ctx context.Context, path string, reader io.Reader) (string, error) {
	fullPath := filepath.Join(s.BaseDir, path)
	
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	return path, nil
}

func (s *LocalStorage) Delete(ctx context.Context, path string) error {
	fullPath := filepath.Join(s.BaseDir, path)
	return os.Remove(fullPath)
}

func (s *LocalStorage) GetURL(ctx context.Context, path string) (string, error) {
	// For local storage, we might return a relative URL or a full URL if BaseURL is set
	return fmt.Sprintf("%s/%s", s.BaseURL, path), nil
}
