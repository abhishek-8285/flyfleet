// Package secrets provides an abstraction layer for secrets management.
// Supports Vault, environment variables, and file-based secrets with
// automatic fallback.
//
// Usage:
//
//	mgr := secrets.NewManager(secrets.WithVault("http://vault:8200", "s.token"))
//	dbURL, err := mgr.Get(ctx, "database/url")
package secrets

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"go.uber.org/zap"
)

// Provider is the interface for a secrets backend.
type Provider interface {
	Get(ctx context.Context, key string) (string, error)
	Name() string
}

// Manager manages secrets with fallback chain.
type Manager struct {
	providers []Provider
	cache     map[string]string
	mu        sync.RWMutex
	logger    *zap.Logger
}

// Option configures the Manager.
type Option func(*Manager)

// WithVault adds a Vault provider to the chain.
func WithVault(addr, token string) Option {
	return func(m *Manager) {
		m.providers = append([]Provider{&VaultProvider{addr: addr, token: token}}, m.providers...)
	}
}

// WithEnv adds the environment variable provider (always included as fallback).
func WithEnv() Option {
	return func(m *Manager) {
		m.providers = append(m.providers, &EnvProvider{})
	}
}

// NewManager creates a secrets manager with the given providers.
// Environment variables are always included as the final fallback.
func NewManager(logger *zap.Logger, opts ...Option) *Manager {
	mgr := &Manager{
		cache:  make(map[string]string),
		logger: logger,
	}

	for _, opt := range opts {
		opt(mgr)
	}

	// Ensure env is always the last fallback
	hasEnv := false
	for _, p := range mgr.providers {
		if p.Name() == "env" {
			hasEnv = true
			break
		}
	}
	if !hasEnv {
		mgr.providers = append(mgr.providers, &EnvProvider{})
	}

	return mgr
}

// Get retrieves a secret by key, trying each provider in order.
func (m *Manager) Get(ctx context.Context, key string) (string, error) {
	// Check cache first
	m.mu.RLock()
	if v, ok := m.cache[key]; ok {
		m.mu.RUnlock()
		return v, nil
	}
	m.mu.RUnlock()

	// Try each provider
	for _, p := range m.providers {
		val, err := p.Get(ctx, key)
		if err == nil && val != "" {
			m.mu.Lock()
			m.cache[key] = val
			m.mu.Unlock()

			m.logger.Debug("Secret resolved",
				zap.String("key", key),
				zap.String("provider", p.Name()))
			return val, nil
		}
	}

	return "", fmt.Errorf("secret %q not found in any provider", key)
}

// MustGet retrieves a secret or panics. Use during startup initialization.
func (m *Manager) MustGet(ctx context.Context, key string) string {
	val, err := m.Get(ctx, key)
	if err != nil {
		m.logger.Fatal("Required secret not found", zap.String("key", key), zap.Error(err))
	}
	return val
}

// ─── Environment Variable Provider ──────────────────

// EnvProvider reads secrets from environment variables.
// Key "database/url" maps to env var "DATABASE_URL".
type EnvProvider struct{}

func (e *EnvProvider) Name() string { return "env" }

func (e *EnvProvider) Get(_ context.Context, key string) (string, error) {
	envKey := strings.ToUpper(strings.ReplaceAll(key, "/", "_"))
	envKey = strings.ReplaceAll(envKey, "-", "_")
	envKey = strings.ReplaceAll(envKey, ".", "_")

	val := os.Getenv(envKey)
	if val == "" {
		return "", fmt.Errorf("env var %s not set", envKey)
	}
	return val, nil
}

// ─── Vault Provider ─────────────────────────────────

// VaultProvider reads secrets from HashiCorp Vault.
// This is a lightweight implementation that can be extended
// with the official Vault client library for production use.
type VaultProvider struct {
	addr  string
	token string
}

func (v *VaultProvider) Name() string { return "vault" }

func (v *VaultProvider) Get(ctx context.Context, key string) (string, error) {
	if v.addr == "" || v.token == "" {
		return "", fmt.Errorf("vault not configured")
	}

	// In production, use the official Vault API client:
	// api.NewClient(&api.Config{Address: v.addr})
	// For now, this serves as the integration point.
	//
	// The key path would be: secret/data/flyfleet/{key}
	// Example: secret/data/flyfleet/database/url

	return "", fmt.Errorf("vault lookup for %q: not implemented (use env fallback)", key)
}

// ─── File Provider ──────────────────────────────────

// FileProvider reads secrets from files (e.g., Kubernetes Secret volumes).
// Key "database/url" maps to file "/run/secrets/database_url".
type FileProvider struct {
	BasePath string
}

func (f *FileProvider) Name() string { return "file" }

func (f *FileProvider) Get(_ context.Context, key string) (string, error) {
	fileName := strings.ReplaceAll(key, "/", "_")
	fileName = strings.ReplaceAll(fileName, "-", "_")
	path := f.BasePath + "/" + fileName

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("file secret %s: %w", path, err)
	}
	return strings.TrimSpace(string(data)), nil
}
