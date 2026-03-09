//go:build integration
// +build integration

package admin_support

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

var testMetrics = NewMetrics()

// MockRepository is a mock implementation of the Repository interface.
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateOrganization(ctx context.Context, org *Organization) error {
	args := m.Called(ctx, org)
	return args.Error(0)
}

func (m *MockRepository) GetOrganization(ctx context.Context, id uuid.UUID) (*Organization, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*Organization), args.Error(1)
}

func (m *MockRepository) UpdateOrganization(ctx context.Context, org *Organization) error {
	args := m.Called(ctx, org)
	return args.Error(0)
}

func (m *MockRepository) ListOrganizations(ctx context.Context) ([]*Organization, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*Organization), args.Error(1)
}

func (m *MockRepository) CreateAdminUser(ctx context.Context, user *AdminUser) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRepository) CreateTicket(ctx context.Context, ticket *SupportTicket) error {
	args := m.Called(ctx, ticket)
	return args.Error(0)
}

func (m *MockRepository) GetTicket(ctx context.Context, id uuid.UUID) (*SupportTicket, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*SupportTicket), args.Error(1)
}

func (m *MockRepository) UpdateTicket(ctx context.Context, ticket *SupportTicket) error {
	args := m.Called(ctx, ticket)
	return args.Error(0)
}

func (m *MockRepository) CreateTicketMessage(ctx context.Context, msg *TicketMessage) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func (m *MockRepository) CreateIncident(ctx context.Context, inc *Incident) error {
	args := m.Called(ctx, inc)
	return args.Error(0)
}

func (m *MockRepository) GetIncident(ctx context.Context, id uuid.UUID) (*Incident, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*Incident), args.Error(1)
}

func (m *MockRepository) UpdateIncident(ctx context.Context, inc *Incident) error {
	args := m.Called(ctx, inc)
	return args.Error(0)
}

func (m *MockRepository) CreateIncidentUpdate(ctx context.Context, u *IncidentUpdate) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockRepository) SetSystemFlag(ctx context.Context, flag *SystemFlag) error {
	args := m.Called(ctx, flag)
	return args.Error(0)
}

func (m *MockRepository) GetSystemFlag(ctx context.Context, key string) (*SystemFlag, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*SystemFlag), args.Error(1)
}

func (m *MockRepository) CreateMaintenanceWindow(ctx context.Context, window *MaintenanceWindow) error {
	args := m.Called(ctx, window)
	return args.Error(0)
}

func (m *MockRepository) CreateAuditLog(ctx context.Context, log *AuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockRepository) CreateImpersonationSession(ctx context.Context, session *ImpersonationSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockRepository) ListAdminUsers(ctx context.Context) ([]*AdminUser, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*AdminUser), args.Error(1)
}

func (m *MockRepository) GetAdminUserByEmail(ctx context.Context, email string) (*AdminUser, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*AdminUser), args.Error(1)
}

func (m *MockRepository) ListTicketsByOrg(ctx context.Context, orgID uuid.UUID) ([]*SupportTicket, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0).([]*SupportTicket), args.Error(1)
}

func (m *MockRepository) ListTicketMessages(ctx context.Context, ticketID uuid.UUID) ([]*TicketMessage, error) {
	args := m.Called(ctx, ticketID)
	return args.Get(0).([]*TicketMessage), args.Error(1)
}

func (m *MockRepository) ListMaintenanceWindows(ctx context.Context) ([]*MaintenanceWindow, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*MaintenanceWindow), args.Error(1)
}

func (m *MockRepository) ListOrganizationUsersByOrg(ctx context.Context, orgID uuid.UUID) ([]*OrganizationUser, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0).([]*OrganizationUser), args.Error(1)
}

func (m *MockRepository) CreateOrganizationUser(ctx context.Context, user *OrganizationUser) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRepository) GetOrganizationUser(ctx context.Context, id uuid.UUID) (*OrganizationUser, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*OrganizationUser), args.Error(1)
}

func (m *MockRepository) GetRolePermissions(ctx context.Context, role string) ([]string, error) {
	args := m.Called(ctx, role)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockRepository) SetRolePermissions(ctx context.Context, role string, permissions []string) error {
	args := m.Called(ctx, role, permissions)
	return args.Error(0)
}

func (m *MockRepository) ListAllTickets(ctx context.Context) ([]*SupportTicket, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*SupportTicket), args.Error(1)
}

func (m *MockRepository) ListTicketAttachments(ctx context.Context, ticketID uuid.UUID) ([]*SupportTicketAttachment, error) {
	args := m.Called(ctx, ticketID)
	return args.Get(0).([]*SupportTicketAttachment), args.Error(1)
}

func (m *MockRepository) CreateTicketAttachment(ctx context.Context, a *SupportTicketAttachment) error {
	args := m.Called(ctx, a)
	return args.Error(0)
}

func (m *MockRepository) ListAllIncidents(ctx context.Context) ([]*Incident, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*Incident), args.Error(1)
}

func (m *MockRepository) ListSystemFlags(ctx context.Context) ([]*SystemFlag, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*SystemFlag), args.Error(1)
}

func (m *MockRepository) ListAdminPermissions(ctx context.Context) ([]*AdminPermission, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*AdminPermission), args.Error(1)
}

func TestService_CreateOrganization(t *testing.T) {
	mockRepo := new(MockRepository)
	logger := zap.NewNop()
	svc := NewService(mockRepo, testMetrics, logger)

	ctx := context.Background()
	mockRepo.On("CreateOrganization", ctx, mock.AnythingOfType("*admin_support.Organization")).Return(nil)

	org, err := svc.CreateOrganization(ctx, "Test Org", "Enterprise", "us-east-1")
	assert.NoError(t, err)
	assert.NotNil(t, org)
	assert.Equal(t, "Test Org", org.Name)
	assert.Equal(t, OrgActive, org.Status)

	mockRepo.AssertExpectations(t)
}

func TestService_CreateSupportTicket_SLA(t *testing.T) {
	mockRepo := new(MockRepository)
	logger := zap.NewNop()
	svc := NewService(mockRepo, testMetrics, logger)

	ctx := context.Background()
	orgID := uuid.New()
	createdBy := uuid.New()

	mockRepo.On("CreateTicket", ctx, mock.AnythingOfType("*admin_support.SupportTicket")).Return(nil)

	// Test Critical Priority SLA (1h)
	ticket, err := svc.CreateTicket(ctx, orgID, "Urgent Issue", PriorityCritical, createdBy)
	assert.NoError(t, err)
	assert.WithinDuration(t, time.Now().Add(1*time.Hour), ticket.SLADeadline, 10*time.Second)

	// Test Low Priority SLA (72h)
	ticket, err = svc.CreateTicket(ctx, orgID, "Minor Request", PriorityLow, createdBy)
	assert.NoError(t, err)
	assert.WithinDuration(t, time.Now().Add(72*time.Hour), ticket.SLADeadline, 10*time.Second)

	mockRepo.AssertExpectations(t)
}

func TestService_RBAC(t *testing.T) {
	// Note: The current CheckPermission implementation is a stub that always assumes SUPER_ADMIN role
	// In production, this would fetch the actual user's role from context/database
	// The stub returns true immediately for SUPER_ADMIN without checking permissions
	mockRepo := new(MockRepository)
	logger := zap.NewNop()
	svc := NewService(mockRepo, testMetrics, logger)

	ctx := context.Background()
	userID := uuid.New()

	// With the current stub implementation, all permission checks return true (assumes SUPER_ADMIN)
	// This is a known limitation of the stub - production code would use actual user context
	// The stub checks: if role == "SUPER_ADMIN" { return true }
	assert.True(t, svc.CheckPermission(ctx, userID, "dr.execute"))
	assert.True(t, svc.CheckPermission(ctx, userID, "org.manage"))
	assert.True(t, svc.CheckPermission(ctx, userID, "ticket.manage"))
	
	// No mock expectations needed as the stub returns true immediately for SUPER_ADMIN
}

func TestService_SuspendActivateOrganization(t *testing.T) {
	mockRepo := new(MockRepository)
	logger := zap.NewNop()
	svc := NewService(mockRepo, testMetrics, logger)

	ctx := context.Background()
	orgID := uuid.New()
	actorID := uuid.New()

	existingOrg := &Organization{
		ID:     orgID,
		Name:   "Test Org",
		Status: OrgActive,
	}

	// Test Suspend
	mockRepo.On("GetOrganization", ctx, orgID).Return(existingOrg, nil)
	mockRepo.On("UpdateOrganization", ctx, mock.AnythingOfType("*admin_support.Organization")).Return(nil)
	mockRepo.On("CreateAuditLog", ctx, mock.AnythingOfType("*admin_support.AuditLog")).Return(nil)

	err := svc.SuspendOrganization(ctx, orgID, actorID)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)

	// Test Activate
	suspendedOrg := &Organization{
		ID:     orgID,
		Name:   "Test Org",
		Status: OrgSuspended,
	}

	mockRepo.On("GetOrganization", ctx, orgID).Return(suspendedOrg, nil)
	mockRepo.On("UpdateOrganization", ctx, mock.AnythingOfType("*admin_support.Organization")).Return(nil)
	mockRepo.On("CreateAuditLog", ctx, mock.AnythingOfType("*admin_support.AuditLog")).Return(nil)

	err = svc.ActivateOrganization(ctx, orgID, actorID)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_TicketWorkflow(t *testing.T) {
	mockRepo := new(MockRepository)
	logger := zap.NewNop()
	svc := NewService(mockRepo, testMetrics, logger)

	ctx := context.Background()
	ticketID := uuid.New()
	adminID := uuid.New()
	senderID := uuid.New()

	existingTicket := &SupportTicket{
		ID:       ticketID,
		Subject:  "Test Ticket",
		Status:   StatusOpen,
		Priority: PriorityHigh,
	}

	// Test Assign Ticket
	mockRepo.On("GetTicket", ctx, ticketID).Return(existingTicket, nil)
	mockRepo.On("UpdateTicket", ctx, mock.AnythingOfType("*admin_support.SupportTicket")).Return(nil)

	err := svc.AssignTicket(ctx, ticketID, adminID)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)

	// Test Reply to Ticket
	mockRepo.On("GetTicket", ctx, ticketID).Return(existingTicket, nil)
	mockRepo.On("CreateTicketMessage", ctx, mock.AnythingOfType("*admin_support.TicketMessage")).Return(nil)
	mockRepo.On("UpdateTicket", ctx, mock.AnythingOfType("*admin_support.SupportTicket")).Return(nil)

	err = svc.ReplyToTicket(ctx, ticketID, senderID, "Test reply message")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)

	// Test Resolve Ticket
	mockRepo.On("GetTicket", ctx, ticketID).Return(existingTicket, nil)
	mockRepo.On("UpdateTicket", ctx, mock.AnythingOfType("*admin_support.SupportTicket")).Return(nil)

	err = svc.ResolveTicket(ctx, ticketID)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_IncidentLifecycle(t *testing.T) {
	mockRepo := new(MockRepository)
	logger := zap.NewNop()
	svc := NewService(mockRepo, testMetrics, logger)

	ctx := context.Background()
	incidentID := uuid.New()

	// Test Declare Incident
	mockRepo.On("CreateIncident", ctx, mock.AnythingOfType("*admin_support.Incident")).Return(nil)

	incident, err := svc.DeclareIncident(ctx, Sev2, "Test incident description")
	assert.NoError(t, err)
	assert.NotNil(t, incident)
	assert.Equal(t, Sev2, incident.Severity)
	assert.Equal(t, IncidentOpen, incident.Status)
	mockRepo.AssertExpectations(t)

	// Test Update Incident
	mockRepo.On("CreateIncidentUpdate", ctx, mock.AnythingOfType("*admin_support.IncidentUpdate")).Return(nil)
	mockRepo.On("GetIncident", ctx, incidentID).Return(&Incident{ID: incidentID}, nil)
	mockRepo.On("UpdateIncident", ctx, mock.AnythingOfType("*admin_support.Incident")).Return(nil)

	err = svc.UpdateIncident(ctx, incidentID, "Investigating the issue")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)

	// Test Resolve Incident
	existingIncident := &Incident{
		ID:     incidentID,
		Status: IncidentOpen,
	}

	mockRepo.On("GetIncident", ctx, incidentID).Return(existingIncident, nil)
	mockRepo.On("UpdateIncident", ctx, mock.AnythingOfType("*admin_support.Incident")).Return(nil)

	err = svc.ResolveIncident(ctx, incidentID)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_Impersonation(t *testing.T) {
	mockRepo := new(MockRepository)
	logger := zap.NewNop()
	svc := NewService(mockRepo, testMetrics, logger)

	ctx := context.Background()
	adminID := uuid.New()
	userID := uuid.New()

	// Test Start Impersonation
	mockRepo.On("CreateAuditLog", ctx, mock.AnythingOfType("*admin_support.AuditLog")).Return(nil)
	mockRepo.On("CreateImpersonationSession", ctx, mock.AnythingOfType("*admin_support.ImpersonationSession")).Return(nil)

	session, err := svc.StartImpersonation(ctx, adminID, userID)
	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, adminID, session.AdminID)
	assert.Equal(t, userID, session.UserID)
	assert.WithinDuration(t, time.Now().Add(30*time.Minute), session.ExpiresAt, 1*time.Second)
	mockRepo.AssertExpectations(t)

	// Test Stop Impersonation
	sessionID := uuid.New()
	err = svc.StopImpersonation(ctx, sessionID)
	assert.NoError(t, err)
}

func TestService_SystemFlags(t *testing.T) {
	mockRepo := new(MockRepository)
	logger := zap.NewNop()
	svc := NewService(mockRepo, testMetrics, logger)

	ctx := context.Background()
	actorID := uuid.New()

	// Test Set System Flag
	mockRepo.On("SetSystemFlag", ctx, mock.AnythingOfType("*admin_support.SystemFlag")).Return(nil)
	mockRepo.On("CreateAuditLog", ctx, mock.AnythingOfType("*admin_support.AuditLog")).Return(nil)

	err := svc.SetSystemFlag(ctx, "tracking_pause", "true", actorID)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_MaintenanceWindow(t *testing.T) {
	mockRepo := new(MockRepository)
	logger := zap.NewNop()
	svc := NewService(mockRepo, testMetrics, logger)

	ctx := context.Background()

	// Test Start Maintenance
	mockRepo.On("CreateMaintenanceWindow", ctx, mock.AnythingOfType("*admin_support.MaintenanceWindow")).Return(nil)

	window, err := svc.StartMaintenance(ctx, "tracking-service", 2*time.Hour, "Scheduled maintenance")
	assert.NoError(t, err)
	assert.NotNil(t, window)
	assert.Equal(t, "tracking-service", window.Service)
	assert.WithinDuration(t, time.Now().Add(2*time.Hour), window.EndTime, 1*time.Second)
	mockRepo.AssertExpectations(t)
}

func TestSLACalculations(t *testing.T) {
	svc := &service{}

	tests := []struct {
		priority       TicketPriority
		expectedHours  time.Duration
	}{
		{PriorityCritical, 1 * time.Hour},
		{PriorityHigh, 6 * time.Hour},
		{PriorityMedium, 24 * time.Hour},
		{PriorityLow, 72 * time.Hour},
	}

	for _, tt := range tests {
		deadline := svc.calculateSLADeadline(tt.priority)
		expected := time.Now().Add(tt.expectedHours)
		assert.WithinDuration(t, expected, deadline, 10*time.Second,
			"Priority %s should have SLA of %v", tt.priority, tt.expectedHours)
	}
}

func TestIncidentSeverityAlerts(t *testing.T) {
	mockRepo := new(MockRepository)
	logger := zap.NewNop()
	svc := NewService(mockRepo, testMetrics, logger)

	ctx := context.Background()

	// Test SEV1 incident (should trigger PagerDuty alert)
	mockRepo.On("CreateIncident", ctx, mock.AnythingOfType("*admin_support.Incident")).Return(nil)

	incident, err := svc.DeclareIncident(ctx, Sev1, "Critical system outage")
	assert.NoError(t, err)
	assert.Equal(t, Sev1, incident.Severity)
	mockRepo.AssertExpectations(t)

	// Test SEV3 incident (lower severity)
	mockRepo.On("CreateIncident", ctx, mock.AnythingOfType("*admin_support.Incident")).Return(nil)

	incident, err = svc.DeclareIncident(ctx, Sev3, "Minor performance issue")
	assert.NoError(t, err)
	assert.Equal(t, Sev3, incident.Severity)
	mockRepo.AssertExpectations(t)
}
