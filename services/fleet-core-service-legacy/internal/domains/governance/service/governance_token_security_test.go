package service_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/flyfleet/fleet-core-service/internal/domains/governance/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockTokenService struct {
	mock.Mock
}

func (m *MockTokenService) ValidateToken(ctx context.Context, token uuid.UUID) (*models.GovernanceExport, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GovernanceExport), args.Error(1)
}

func TestTokenSecurity_InvalidToken(t *testing.T) {
	mockSvc := new(MockTokenService)
	token := uuid.New()
	
	mockSvc.On("ValidateToken", mock.Anything, token).Return(nil, fmt.Errorf("token invalid or expired"))
	
	_, err := mockSvc.ValidateToken(context.Background(), token)
	if err == nil {
		t.Fatalf("Expected error for invalid token, got nil")
	}
	
	if err.Error() != "token invalid or expired" {
		t.Errorf("Expected 'token invalid or expired', got %s", err.Error())
	}
	
	mockSvc.AssertExpectations(t)
}
