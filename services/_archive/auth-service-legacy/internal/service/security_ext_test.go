package service

import (
	"context"
	"testing"

	"github.com/flyfleet/auth-service/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSecurityRepository implements repository methods for testing
type MockSecurityRepository struct {
	mock.Mock
}

func (m *MockSecurityRepository) ListIPRules(ctx context.Context, orgID uuid.UUID) ([]models.IPRule, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.IPRule), args.Error(1)
}

func (m *MockSecurityRepository) GetIPRuleByID(ctx context.Context, id uuid.UUID) (*models.IPRule, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.IPRule), args.Error(1)
}

func (m *MockSecurityRepository) CreateIPRule(ctx context.Context, rule *models.IPRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockSecurityRepository) DeleteIPRule(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSecurityRepository) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

// Test CIDR Validation
func TestValidateAndNormalizeCIDR(t *testing.T) {
	s := &AuthService{}

	tests := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{"Valid IPv4 CIDR", "192.168.1.0/24", "192.168.1.0/24", false},
		{"Valid IPv4 /32", "10.0.0.1/32", "10.0.0.1/32", false},
		{"Single IPv4", "192.168.1.1", "192.168.1.1/32", false},
		{"Single IPv6", "::1", "::1/128", false},
		{"Valid IPv6 CIDR", "2001:db8::/32", "2001:db8::/32", false},
		{"Whitespace trimmed", "  192.168.1.0/24  ", "192.168.1.0/24", false},
		{"Invalid IP", "not-an-ip", "", true},
		{"Invalid CIDR mask", "192.168.1.0/33", "", true},
		{"Empty string", "", "", true},
		{"Partial IP", "192.168", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := s.validateAndNormalizeCIDR(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// Test CreateIPRule validation
func TestCreateIPRule_Validation(t *testing.T) {
	ctx := context.Background()
	adminID := uuid.New()
	clientIP := "192.168.1.100"

	tests := []struct {
		name        string
		rule        *models.IPRule
		expectedErr error
	}{
		{
			name: "Missing OrgID",
			rule: &models.IPRule{
				OrgID: uuid.Nil,
				CIDR:  "10.0.0.1",
				Action: "BLOCK",
			},
			expectedErr: ErrInvalidOrgID,
		},
		{
			name: "Invalid CIDR",
			rule: &models.IPRule{
				OrgID: uuid.New(),
				CIDR:  "not-valid",
				Action: "BLOCK",
			},
			expectedErr: ErrInvalidCIDR,
		},
		{
			name: "Invalid action",
			rule: &models.IPRule{
				OrgID: uuid.New(),
				CIDR:  "10.0.0.1",
				Action: "INVALID",
			},
			expectedErr: ErrInvalidAction,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a minimal service without repo for validation tests
			s := &AuthService{}
			err := s.CreateIPRule(ctx, tt.rule, adminID, clientIP)
			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

// Test session revoke validation
func TestRevokeSession_Validation(t *testing.T) {
	ctx := context.Background()
	adminID := uuid.New()
	clientIP := "192.168.1.100"

	s := &AuthService{}

	err := s.RevokeSession(ctx, uuid.Nil, adminID, clientIP)
	assert.ErrorIs(t, err, ErrInvalidSessionID)
}

// Test RevokeAllSessionsForOrg validation
func TestRevokeAllSessionsForOrg_Validation(t *testing.T) {
	ctx := context.Background()
	adminID := uuid.New()
	clientIP := "192.168.1.100"

	s := &AuthService{}

	err := s.RevokeAllSessionsForOrg(ctx, uuid.Nil, adminID, clientIP)
	assert.ErrorIs(t, err, ErrInvalidOrgID)
}

// Test ListIPRules validation
func TestListIPRules_Validation(t *testing.T) {
	ctx := context.Background()
	s := &AuthService{}

	_, err := s.ListIPRules(ctx, uuid.Nil)
	assert.ErrorIs(t, err, ErrInvalidOrgID)
}

// Test Access Review validation
func TestListAccessReviews_Validation(t *testing.T) {
	ctx := context.Background()
	s := &AuthService{}

	_, _, err := s.ListAccessReviews(ctx, uuid.Nil, 1, 20)
	assert.ErrorIs(t, err, ErrInvalidOrgID)
}

func TestCreateAccessReview_Validation(t *testing.T) {
	ctx := context.Background()
	adminID := uuid.New()
	clientIP := "192.168.1.100"
	s := &AuthService{}

	// Missing org ID
	review := &models.AccessReview{
		OrgID: uuid.Nil,
		Title: "Q1 Review",
	}
	err := s.CreateAccessReview(ctx, review, adminID, clientIP)
	assert.ErrorIs(t, err, ErrInvalidOrgID)

	// Missing title
	review2 := &models.AccessReview{
		OrgID: uuid.New(),
		Title: "",
	}
	err = s.CreateAccessReview(ctx, review2, adminID, clientIP)
	assert.Error(t, err)
}

func TestUpdateReviewItem_Validation(t *testing.T) {
	ctx := context.Background()
	clientIP := "192.168.1.100"
	s := &AuthService{}

	// Invalid item ID
	err := s.UpdateReviewItem(ctx, uuid.Nil, "APPROVED", uuid.New(), "", clientIP)
	assert.Error(t, err)

	// Invalid status
	err = s.UpdateReviewItem(ctx, uuid.New(), "INVALID", uuid.New(), "", clientIP)
	assert.ErrorIs(t, err, ErrInvalidStatus)

	// Missing reviewer ID
	err = s.UpdateReviewItem(ctx, uuid.New(), "APPROVED", uuid.Nil, "", clientIP)
	assert.ErrorIs(t, err, ErrInvalidUserID)
}
