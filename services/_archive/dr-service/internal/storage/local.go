package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type LocalProvider struct {
	baseDir string
}

func NewLocalProvider(baseDir string) (*LocalProvider, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}
	return &LocalProvider{baseDir: baseDir}, nil
}

func (l *LocalProvider) Upload(ctx context.Context, key string, data io.Reader) (string, error) {
	fullPath := filepath.Join(l.baseDir, key)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory for key: %w", err)
	}

	out, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, data); err != nil {
		return "", fmt.Errorf("failed to write data: %w", err)
	}

	return fullPath, nil
}

func (l *LocalProvider) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	fullPath := filepath.Join(l.baseDir, key)
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return file, nil
}

func (l *LocalProvider) Delete(ctx context.Context, key string) error {
	fullPath := filepath.Join(l.baseDir, key)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove file: %w", err)
	}
	return nil
}

func (l *LocalProvider) List(ctx context.Context, prefix string) ([]string, error) {
	var keys []string
	root := filepath.Join(l.baseDir, prefix)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if !info.IsDir() {
			relPath, err := filepath.Rel(l.baseDir, path)
			if err == nil {
				keys = append(keys, relPath)
			}
		}
		return nil
	})
	return keys, err
}
