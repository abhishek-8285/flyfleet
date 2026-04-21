package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/integrations/models"
	"github.com/flyfleet/fleet-core/internal/domains/integrations/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type IntegrationsService interface {
	// Integrations
	CreateIntegration(ctx context.Context, tenantID uuid.UUID, name, iType, provider string, config any) (*models.Integration, error)
	GetIntegration(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*models.Integration, error)
	ListIntegrations(ctx context.Context, tenantID uuid.UUID) ([]*models.Integration, error)
	UpdateIntegration(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, name, iType, provider, status string, config any) error
	DeleteIntegration(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error
	TestConnection(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (bool, string, error)

	// API Clients
	CreateAPIClient(ctx context.Context, tenantID uuid.UUID, name string, scopes []string, expiryDays int) (*models.APIClient, string, error)
	RotateAPIClientKey(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (string, error)
	ListAPIClients(ctx context.Context, tenantID uuid.UUID) ([]*models.APIClient, error)
	DeleteAPIClient(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error

	// Webhooks
	CreateWebhook(ctx context.Context, tenantID uuid.UUID, url string, events []string) (*models.Webhook, error)
	ListWebhooks(ctx context.Context, tenantID uuid.UUID) ([]*models.Webhook, error)
	UpdateWebhook(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, url string, events []string, status string) error
	DeleteWebhook(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error

	// Logs
	ListLogs(ctx context.Context, tenantID uuid.UUID, integrationID *uuid.UUID, limit, offset int) ([]*models.IntegrationLog, error)
}

type integrationsService struct {
	repo       repository.Repository
	encryptKey string
	logger     *zap.Logger
}

func NewIntegrationsService(repo repository.Repository, encryptKey string, logger *zap.Logger) IntegrationsService {
	return &integrationsService{
		repo:       repo,
		encryptKey: encryptKey,
		logger:     logger,
	}
}

func (s *integrationsService) CreateIntegration(ctx context.Context, tenantID uuid.UUID, name, iType, provider string, config any) (*models.Integration, error) {
	userID := s.getUserID(ctx)
	integration := &models.Integration{
		ID:        uuid.New(),
		TenantID:     tenantID,
		Name:      name,
		Type:      iType,
		Provider:  provider,
		Status:    "active",
		Config:    config, // Encrypting during repository save or here if strictly needed
		CreatedBy: userID,
		UpdatedBy: userID,
	}
	err := s.repo.CreateIntegration(ctx, integration)
	return integration, err
}

func (s *integrationsService) GetIntegration(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*models.Integration, error) {
	return s.repo.GetIntegration(ctx, id, tenantID)
}

func (s *integrationsService) ListIntegrations(ctx context.Context, tenantID uuid.UUID) ([]*models.Integration, error) {
	return s.repo.ListIntegrations(ctx, tenantID)
}

func (s *integrationsService) UpdateIntegration(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, name, iType, provider, status string, config any) error {
	userID := s.getUserID(ctx)
	integration := &models.Integration{
		ID:        id,
		TenantID:     tenantID,
		Name:      name,
		Type:      iType,
		Provider:  provider,
		Status:    status,
		Config:    config,
		UpdatedBy: userID,
	}
	return s.repo.UpdateIntegration(ctx, integration)
}

func (s *integrationsService) DeleteIntegration(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	return s.repo.DeleteIntegration(ctx, id, tenantID)
}

func (s *integrationsService) TestConnection(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (bool, string, error) {
	// Mock implementation for now
	time.Sleep(1 * time.Second)
	return true, "Connection successful", nil
}

func (s *integrationsService) CreateAPIClient(ctx context.Context, tenantID uuid.UUID, name string, scopes []string, expiryDays int) (*models.APIClient, string, error) {
	rawKey := generateSecureKey(32)
	hash, _ := bcrypt.GenerateFromPassword([]byte(rawKey), bcrypt.DefaultCost)

	var expiry *time.Time
	if expiryDays > 0 {
		t := time.Now().AddDate(0, 0, expiryDays)
		expiry = &t
	}

	userID := s.getUserID(ctx)
	u := uuid.New()
	client := &models.APIClient{
		ID:               uuid.New(),
		TenantID:            tenantID,
		ClientID:         "cli_" + hex.EncodeToString(u[:8]),
		ClientSecretHash: string(hash),
		Name:             name,
		Scopes:           scopes,
		Expiry:           expiry,
		Status:           "active",
		CreatedBy:        userID,
		UpdatedBy:        userID,
	}

	err := s.repo.CreateAPIClient(ctx, client)
	return client, rawKey, err
}

func (s *integrationsService) RotateAPIClientKey(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (string, error) {
	client, err := s.repo.GetAPIClient(ctx, id, tenantID)
	if err != nil {
		return "", err
	}

	rawKey := generateSecureKey(32)
	hash, _ := bcrypt.GenerateFromPassword([]byte(rawKey), bcrypt.DefaultCost)

	client.ClientSecretHash = string(hash)
	client.UpdatedBy = s.getUserID(ctx)
	return rawKey, s.repo.UpdateAPIClient(ctx, client)
}

func (s *integrationsService) ListAPIClients(ctx context.Context, tenantID uuid.UUID) ([]*models.APIClient, error) {
	return s.repo.ListAPIClients(ctx, tenantID)
}

func (s *integrationsService) DeleteAPIClient(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	return s.repo.DeleteAPIClient(ctx, id, tenantID)
}

func (s *integrationsService) CreateWebhook(ctx context.Context, tenantID uuid.UUID, url string, events []string) (*models.Webhook, error) {
	userID := s.getUserID(ctx)
	webhook := &models.Webhook{
		ID:        uuid.New(),
		TenantID:     tenantID,
		URL:       url,
		Events:    events,
		Secret:    "whsec_" + generateSecureKey(16),
		Status:    "active",
		CreatedBy: userID,
		UpdatedBy: userID,
	}
	err := s.repo.CreateWebhook(ctx, webhook)
	return webhook, err
}

func (s *integrationsService) ListWebhooks(ctx context.Context, tenantID uuid.UUID) ([]*models.Webhook, error) {
	return s.repo.ListWebhooks(ctx, tenantID)
}

func (s *integrationsService) UpdateWebhook(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, url string, events []string, status string) error {
	userID := s.getUserID(ctx)
	webhook := &models.Webhook{
		ID:        id,
		TenantID:     tenantID,
		URL:       url,
		Events:    events,
		Status:    status,
		UpdatedBy: userID,
	}
	return s.repo.UpdateWebhook(ctx, webhook)
}

func (s *integrationsService) getUserID(ctx context.Context) uuid.UUID {
	val := ctx.Value("user_id")
	if val == nil {
		return uuid.Nil
	}
	uidStr, ok := val.(string)
	if !ok {
		return uuid.Nil
	}
	uid, _ := uuid.Parse(uidStr)
	return uid
}

func (s *integrationsService) DeleteWebhook(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	return s.repo.DeleteWebhook(ctx, id, tenantID)
}

func (s *integrationsService) ListLogs(ctx context.Context, tenantID uuid.UUID, integrationID *uuid.UUID, limit, offset int) ([]*models.IntegrationLog, error) {
	return s.repo.ListIntegrationLogs(ctx, tenantID, integrationID, limit, offset)
}

func generateSecureKey(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
