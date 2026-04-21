//go:build integration
// +build integration

package service

import (
	"context"
	"testing"

	"github.com/flyfleet/fleet-core/internal/domains/auth/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAuthService_RefreshToken_WithTenantID(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()

	userID := uuid.New()
	tenantID1 := uuid.New()
	tenantID2 := uuid.New()
	refreshTokenStr := "valid-refresh-token"

	mockRepo.GetRefreshTokenFn = func(ctx context.Context, token string) (*models.RefreshToken, error) {
		return &models.RefreshToken{
			ID:      uuid.New(),
			UserID:  userID,
			Token:   token,
			Revoked: false,
		}, nil
	}

	mockRepo.RevokeRefreshTokenFn = func(ctx context.Context, token string) error {
		return nil
	}

	mockRepo.GetUserByIDFn = func(ctx context.Context, id uuid.UUID) (*models.User, error) {
		return &models.User{
			ID:    id,
			Email: "test@example.com",
			TenantID: tenantID1,
			Role:  "user",
		}, nil
	}

	mockRepo.GetOrganizationsByEmailFn = func(ctx context.Context, email string) ([]models.Organization, error) {
		return []models.Organization{
			{ID: tenantID1, Name: "Org 1"},
			{ID: tenantID2, Name: "Org 2"},
		}, nil
	}

	mockRepo.GetUserPermissionsFn = func(ctx context.Context, userID, tenantID uuid.UUID) ([]string, error) {
		return []string{"read:trips"}, nil
	}

	// Case 1: Refresh with new Org ID (Org 2)
	events := make([]*models.SecurityEvent, 0)
	mockRepo.LogSecurityEventFn = func(ctx context.Context, event *models.SecurityEvent) error {
		events = append(events, event)
		return nil
	}

	resp, err := service.RefreshToken(ctx, refreshTokenStr, tenantID2.String(), "127.0.0.1", "test")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, tenantID2, resp.User.TenantID) // Should be updated to Org 2

	// Verify ORG_CONTEXT_SWITCH event
	foundSwitch := false
	for _, e := range events {
		if e.EventType == "ORG_CONTEXT_SWITCH" {
			foundSwitch = true
			assert.Contains(t, e.Metadata, tenantID1.String())
			assert.Contains(t, e.Metadata, tenantID2.String())
		}
	}
	assert.True(t, foundSwitch, "ORG_CONTEXT_SWITCH event not found")

	// Case 2: Refresh with unauthorized Org ID
	unauthorizedTenantID := uuid.New().String()
	_, err = service.RefreshToken(ctx, refreshTokenStr, unauthorizedTenantID, "127.0.0.1", "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized")

	// Case 3: Refresh without Org ID (should keep current/default)
	resp, err = service.RefreshToken(ctx, refreshTokenStr, "", "127.0.0.1", "test")
	assert.NoError(t, err)
	assert.Equal(t, tenantID1, resp.User.TenantID)
}

func TestAuthService_RevokeAllOrgTokens(t *testing.T) {
	service, mockRepo := setupTestService(t)
	ctx := context.Background()
	tenantID := uuid.New()

	called := false
	mockRepo.RevokeAllOrgTokensFn = func(ctx context.Context, id uuid.UUID) error {
		called = true
		assert.Equal(t, tenantID, id)
		return nil
	}

	err := service.repo.RevokeAllOrgTokens(ctx, tenantID)
	assert.NoError(t, err)
	assert.True(t, called)
}
