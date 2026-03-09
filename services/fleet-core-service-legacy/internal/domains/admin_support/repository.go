package admin_support

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Repository defines the data access layer for the Admin Support Module.
type Repository interface {
	// Organizations
	CreateOrganization(ctx context.Context, org *Organization) error
	GetOrganization(ctx context.Context, id uuid.UUID) (*Organization, error)
	UpdateOrganization(ctx context.Context, org *Organization) error
	ListOrganizations(ctx context.Context) ([]*Organization, error)

	// Organization Users
	CreateOrganizationUser(ctx context.Context, user *OrganizationUser) error
	GetOrganizationUser(ctx context.Context, id uuid.UUID) (*OrganizationUser, error)
	ListOrganizationUsersByOrg(ctx context.Context, orgID uuid.UUID) ([]*OrganizationUser, error)

	// Admin Users
	CreateAdminUser(ctx context.Context, user *AdminUser) error
	GetAdminUserByEmail(ctx context.Context, email string) (*AdminUser, error)
	ListAdminUsers(ctx context.Context) ([]*AdminUser, error)

	// Admin Permissions & RBAC
	ListAdminPermissions(ctx context.Context) ([]*AdminPermission, error)
	GetRolePermissions(ctx context.Context, role string) ([]string, error)
	SetRolePermissions(ctx context.Context, role string, permissions []string) error

	// Support Tickets
	CreateTicket(ctx context.Context, ticket *SupportTicket) error
	GetTicket(ctx context.Context, id uuid.UUID) (*SupportTicket, error)
	UpdateTicket(ctx context.Context, ticket *SupportTicket) error
	ListTicketsByOrg(ctx context.Context, orgID uuid.UUID) ([]*SupportTicket, error)
	ListAllTickets(ctx context.Context) ([]*SupportTicket, error)
	CreateTicketMessage(ctx context.Context, msg *TicketMessage) error
	ListTicketMessages(ctx context.Context, ticketID uuid.UUID) ([]*TicketMessage, error)
	CreateTicketAttachment(ctx context.Context, att *SupportTicketAttachment) error
	ListTicketAttachments(ctx context.Context, ticketID uuid.UUID) ([]*SupportTicketAttachment, error)

	// Incidents
	CreateIncident(ctx context.Context, inc *Incident) error
	GetIncident(ctx context.Context, id uuid.UUID) (*Incident, error)
	UpdateIncident(ctx context.Context, inc *Incident) error
	CreateIncidentUpdate(ctx context.Context, update *IncidentUpdate) error
	ListAllIncidents(ctx context.Context) ([]*Incident, error)

	// System Flags & Maintenance
	SetSystemFlag(ctx context.Context, flag *SystemFlag) error
	GetSystemFlag(ctx context.Context, key string) (*SystemFlag, error)
	ListSystemFlags(ctx context.Context) ([]*SystemFlag, error)
	CreateMaintenanceWindow(ctx context.Context, window *MaintenanceWindow) error
	ListMaintenanceWindows(ctx context.Context) ([]*MaintenanceWindow, error)
	ListActiveMaintenanceWindows(ctx context.Context) ([]*MaintenanceWindow, error)
	DeactivateExpiredMaintenanceWindows(ctx context.Context) (int64, error)

	// Worker Queries
	ListSLABreachedTickets(ctx context.Context) ([]*SupportTicket, error)
	ListUnresolvedSEV1Incidents(ctx context.Context) ([]*Incident, error)

	// Audit & Impersonation
	CreateAuditLog(ctx context.Context, log *AuditLog) error
	CreateImpersonationSession(ctx context.Context, session *ImpersonationSession) error
	UpdateImpersonationSession(ctx context.Context, session *ImpersonationSession) error
	GetImpersonationSession(ctx context.Context, id uuid.UUID) (*ImpersonationSession, error)

	// Approvals (Two-Person)
	CreateAdminOperationApproval(ctx context.Context, app *AdminOperationApproval) error
	GetAdminOperationApproval(ctx context.Context, id uuid.UUID) (*AdminOperationApproval, error)
	UpdateAdminOperationApproval(ctx context.Context, app *AdminOperationApproval) error
	ListPendingApprovals(ctx context.Context) ([]*AdminOperationApproval, error)

	// Admin Search Sync State
	GetAdminSearchSyncState(ctx context.Context, domain string) (*AdminSearchSyncState, error)
	UpdateAdminSearchSyncState(ctx context.Context, state *AdminSearchSyncState) error
}

type repository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewRepository(db *pgxpool.Pool, logger *zap.Logger) Repository {
	return &repository{
		db:     db,
		logger: logger,
	}
}

// ─── Organizations ────────────────────────────────────────────────────────────

func (r *repository) CreateOrganization(ctx context.Context, org *Organization) error {
	query := `
		INSERT INTO organizations (id, name, plan_type, status, region, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query, org.ID, org.Name, org.PlanType, org.Status, org.Region, org.CreatedAt, org.UpdatedAt)
	return err
}

func (r *repository) GetOrganization(ctx context.Context, id uuid.UUID) (*Organization, error) {
	query := `SELECT id, name, plan_type, status, region, created_at, updated_at FROM organizations WHERE id = $1`
	var org Organization
	err := r.db.QueryRow(ctx, query, id).Scan(
		&org.ID, &org.Name, &org.PlanType, &org.Status, &org.Region, &org.CreatedAt, &org.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *repository) UpdateOrganization(ctx context.Context, org *Organization) error {
	query := `
		UPDATE organizations 
		SET name = $1, plan_type = $2, status = $3, region = $4, updated_at = NOW()
		WHERE id = $5
	`
	_, err := r.db.Exec(ctx, query, org.Name, org.PlanType, org.Status, org.Region, org.ID)
	return err
}

func (r *repository) ListOrganizations(ctx context.Context) ([]*Organization, error) {
	query := `SELECT id, name, plan_type, status, region, created_at, updated_at FROM organizations`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []*Organization
	for rows.Next() {
		var org Organization
		if err := rows.Scan(&org.ID, &org.Name, &org.PlanType, &org.Status, &org.Region, &org.CreatedAt, &org.UpdatedAt); err != nil {
			return nil, err
		}
		orgs = append(orgs, &org)
	}
	return orgs, nil
}

// ─── Organization Users ───────────────────────────────────────────────────────

func (r *repository) CreateOrganizationUser(ctx context.Context, u *OrganizationUser) error {
	query := `INSERT INTO organization_users (id, organization_id, email, role, status, created_at) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(ctx, query, u.ID, u.OrganizationID, u.Email, u.Role, u.Status, u.CreatedAt)
	return err
}

func (r *repository) GetOrganizationUser(ctx context.Context, id uuid.UUID) (*OrganizationUser, error) {
	query := `SELECT id, organization_id, email, role, status, created_at FROM organization_users WHERE id = $1`
	var u OrganizationUser
	err := r.db.QueryRow(ctx, query, id).Scan(&u.ID, &u.OrganizationID, &u.Email, &u.Role, &u.Status, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *repository) ListOrganizationUsersByOrg(ctx context.Context, orgID uuid.UUID) ([]*OrganizationUser, error) {
	query := `SELECT id, organization_id, email, role, status, created_at FROM organization_users WHERE organization_id = $1`
	rows, err := r.db.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*OrganizationUser
	for rows.Next() {
		var u OrganizationUser
		if err := rows.Scan(&u.ID, &u.OrganizationID, &u.Email, &u.Role, &u.Status, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, &u)
	}
	return users, nil
}

// ─── Admin Users ──────────────────────────────────────────────────────────────

func (r *repository) CreateAdminUser(ctx context.Context, user *AdminUser) error {
	query := `INSERT INTO admin_users (id, email, role, mfa_enabled, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(ctx, query, user.ID, user.Email, user.Role, user.MFAEnabled, user.CreatedAt)
	return err
}

func (r *repository) GetAdminUserByEmail(ctx context.Context, email string) (*AdminUser, error) {
	query := `SELECT id, email, role, mfa_enabled, created_at FROM admin_users WHERE email = $1`
	var user AdminUser
	err := r.db.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.Role, &user.MFAEnabled, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) ListAdminUsers(ctx context.Context) ([]*AdminUser, error) {
	query := `SELECT id, email, role, mfa_enabled, created_at FROM admin_users`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*AdminUser
	for rows.Next() {
		var user AdminUser
		if err := rows.Scan(&user.ID, &user.Email, &user.Role, &user.MFAEnabled, &user.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	return users, nil
}

// ─── Admin Permissions & RBAC ─────────────────────────────────────────────────

func (r *repository) ListAdminPermissions(ctx context.Context) ([]*AdminPermission, error) {
	query := `SELECT id, name FROM admin_permissions`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []*AdminPermission
	for rows.Next() {
		var p AdminPermission
		if err := rows.Scan(&p.ID, &p.Name); err != nil {
			return nil, err
		}
		perms = append(perms, &p)
	}
	return perms, nil
}

func (r *repository) GetRolePermissions(ctx context.Context, role string) ([]string, error) {
	query := `SELECT permission FROM admin_role_permissions WHERE role = $1`
	rows, err := r.db.Query(ctx, query, role)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, nil
}

func (r *repository) SetRolePermissions(ctx context.Context, role string, permissions []string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Delete existing
	if _, err := tx.Exec(ctx, `DELETE FROM admin_role_permissions WHERE role = $1`, role); err != nil {
		return err
	}

	// Insert new
	for _, p := range permissions {
		if _, err := tx.Exec(ctx, `INSERT INTO admin_role_permissions (role, permission) VALUES ($1, $2)`, role, p); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// ─── Support Tickets ──────────────────────────────────────────────────────────

func (r *repository) CreateTicket(ctx context.Context, t *SupportTicket) error {
	query := `
		INSERT INTO support_tickets (id, organization_id, subject, priority, status, created_by, assigned_to, sla_deadline, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.Exec(ctx, query, t.ID, t.OrganizationID, t.Subject, t.Priority, t.Status, t.CreatedBy, t.AssignedTo, t.SLADeadline, t.CreatedAt, t.UpdatedAt)
	return err
}

func (r *repository) GetTicket(ctx context.Context, id uuid.UUID) (*SupportTicket, error) {
	query := `SELECT id, organization_id, subject, priority, status, created_by, assigned_to, sla_deadline, created_at, updated_at FROM support_tickets WHERE id = $1`
	var t SupportTicket
	err := r.db.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.OrganizationID, &t.Subject, &t.Priority, &t.Status, &t.CreatedBy, &t.AssignedTo, &t.SLADeadline, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *repository) UpdateTicket(ctx context.Context, t *SupportTicket) error {
	query := `
		UPDATE support_tickets 
		SET priority = $1, status = $2, assigned_to = $3, updated_at = NOW()
		WHERE id = $4
	`
	_, err := r.db.Exec(ctx, query, t.Priority, t.Status, t.AssignedTo, t.ID)
	return err
}

func (r *repository) ListTicketsByOrg(ctx context.Context, orgID uuid.UUID) ([]*SupportTicket, error) {
	query := `SELECT id, organization_id, subject, priority, status, created_by, assigned_to, sla_deadline, created_at, updated_at FROM support_tickets WHERE organization_id = $1`
	rows, err := r.db.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []*SupportTicket
	for rows.Next() {
		var t SupportTicket
		if err := rows.Scan(&t.ID, &t.OrganizationID, &t.Subject, &t.Priority, &t.Status, &t.CreatedBy, &t.AssignedTo, &t.SLADeadline, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tickets = append(tickets, &t)
	}
	return tickets, nil
}

func (r *repository) ListAllTickets(ctx context.Context) ([]*SupportTicket, error) {
	query := `SELECT id, organization_id, subject, priority, status, created_by, assigned_to, sla_deadline, created_at, updated_at FROM support_tickets`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []*SupportTicket
	for rows.Next() {
		var t SupportTicket
		if err := rows.Scan(&t.ID, &t.OrganizationID, &t.Subject, &t.Priority, &t.Status, &t.CreatedBy, &t.AssignedTo, &t.SLADeadline, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tickets = append(tickets, &t)
	}
	return tickets, nil
}

func (r *repository) CreateTicketMessage(ctx context.Context, m *TicketMessage) error {
	query := `INSERT INTO support_ticket_messages (id, ticket_id, sender_id, message, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(ctx, query, m.ID, m.TicketID, m.SenderID, m.Message, m.CreatedAt)
	return err
}

func (r *repository) ListTicketMessages(ctx context.Context, ticketID uuid.UUID) ([]*TicketMessage, error) {
	query := `SELECT id, ticket_id, sender_id, message, created_at FROM support_ticket_messages WHERE ticket_id = $1 ORDER BY created_at ASC`
	rows, err := r.db.Query(ctx, query, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []*TicketMessage
	for rows.Next() {
		var m TicketMessage
		if err := rows.Scan(&m.ID, &m.TicketID, &m.SenderID, &m.Message, &m.CreatedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, &m)
	}
	return msgs, nil
}

func (r *repository) CreateTicketAttachment(ctx context.Context, a *SupportTicketAttachment) error {
	query := `INSERT INTO support_ticket_attachments (id, ticket_id, file_url, uploaded_by, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(ctx, query, a.ID, a.TicketID, a.FileURL, a.UploadedBy, a.CreatedAt)
	return err
}

func (r *repository) ListTicketAttachments(ctx context.Context, ticketID uuid.UUID) ([]*SupportTicketAttachment, error) {
	query := `SELECT id, ticket_id, file_url, uploaded_by, created_at FROM support_ticket_attachments WHERE ticket_id = $1`
	rows, err := r.db.Query(ctx, query, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var atts []*SupportTicketAttachment
	for rows.Next() {
		var a SupportTicketAttachment
		if err := rows.Scan(&a.ID, &a.TicketID, &a.FileURL, &a.UploadedBy, &a.CreatedAt); err != nil {
			return nil, err
		}
		atts = append(atts, &a)
	}
	return atts, nil
}

// ─── Incidents ──────────────────────────────────────────────────────────────

func (r *repository) CreateIncident(ctx context.Context, inc *Incident) error {
	query := `INSERT INTO incidents (id, severity, status, description, detected_at, resolved_at, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.Exec(ctx, query, inc.ID, inc.Severity, inc.Status, inc.Description, inc.DetectedAt, inc.ResolvedAt, inc.CreatedAt, inc.UpdatedAt)
	return err
}

func (r *repository) GetIncident(ctx context.Context, id uuid.UUID) (*Incident, error) {
	query := `SELECT id, severity, status, description, detected_at, resolved_at, created_at, updated_at FROM incidents WHERE id = $1`
	var inc Incident
	err := r.db.QueryRow(ctx, query, id).Scan(&inc.ID, &inc.Severity, &inc.Status, &inc.Description, &inc.DetectedAt, &inc.ResolvedAt, &inc.CreatedAt, &inc.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &inc, nil
}

func (r *repository) UpdateIncident(ctx context.Context, inc *Incident) error {
	query := `UPDATE incidents SET severity = $1, status = $2, resolved_at = $3, updated_at = NOW() WHERE id = $4`
	_, err := r.db.Exec(ctx, query, inc.Severity, inc.Status, inc.ResolvedAt, inc.ID)
	return err
}

func (r *repository) CreateIncidentUpdate(ctx context.Context, u *IncidentUpdate) error {
	query := `INSERT INTO incident_updates (id, incident_id, update_message, created_at) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(ctx, query, u.ID, u.IncidentID, u.UpdateMessage, u.CreatedAt)
	return err
}

// ─── System Flags ──────────────────────────────────────────────────────────

func (r *repository) SetSystemFlag(ctx context.Context, f *SystemFlag) error {
	query := `
		INSERT INTO system_flags (key, value, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()
	`
	_, err := r.db.Exec(ctx, query, f.Key, f.Value)
	return err
}

func (r *repository) GetSystemFlag(ctx context.Context, key string) (*SystemFlag, error) {
	query := `SELECT key, value, updated_at FROM system_flags WHERE key = $1`
	var f SystemFlag
	err := r.db.QueryRow(ctx, query, key).Scan(&f.Key, &f.Value, &f.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *repository) ListSystemFlags(ctx context.Context) ([]*SystemFlag, error) {
	query := `SELECT key, value, updated_at FROM system_flags ORDER BY key`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flags []*SystemFlag
	for rows.Next() {
		var f SystemFlag
		if err := rows.Scan(&f.Key, &f.Value, &f.UpdatedAt); err != nil {
			return nil, err
		}
		flags = append(flags, &f)
	}
	return flags, nil
}

// ─── Maintenance Windows ───────────────────────────────────────────────────

func (r *repository) CreateMaintenanceWindow(ctx context.Context, w *MaintenanceWindow) error {
	query := `INSERT INTO maintenance_windows (id, service, start_time, end_time, description, created_at) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(ctx, query, w.ID, w.Service, w.StartTime, w.EndTime, w.Description, w.CreatedAt)
	return err
}

func (r *repository) ListMaintenanceWindows(ctx context.Context) ([]*MaintenanceWindow, error) {
	query := `SELECT id, service, start_time, end_time, description, created_at FROM maintenance_windows ORDER BY start_time DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var windows []*MaintenanceWindow
	for rows.Next() {
		var w MaintenanceWindow
		if err := rows.Scan(&w.ID, &w.Service, &w.StartTime, &w.EndTime, &w.Description, &w.CreatedAt); err != nil {
			return nil, err
		}
		windows = append(windows, &w)
	}
	return windows, nil
}

func (r *repository) ListActiveMaintenanceWindows(ctx context.Context) ([]*MaintenanceWindow, error) {
	query := `SELECT id, service, start_time, end_time, description, created_at FROM maintenance_windows WHERE start_time <= NOW() AND end_time > NOW()`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var windows []*MaintenanceWindow
	for rows.Next() {
		var w MaintenanceWindow
		if err := rows.Scan(&w.ID, &w.Service, &w.StartTime, &w.EndTime, &w.Description, &w.CreatedAt); err != nil {
			return nil, err
		}
		windows = append(windows, &w)
	}
	return windows, nil
}

func (r *repository) DeactivateExpiredMaintenanceWindows(ctx context.Context) (int64, error) {
	// Maintenance windows are time-bounded; we just query by end_time.
	// A future migration can add an 'active' flag for explicit deactivation.
	// For now, return the count of windows that have just expired.
	query := `SELECT COUNT(*) FROM maintenance_windows WHERE end_time <= NOW() AND end_time > NOW() - INTERVAL '1 minute'`
	var count int64
	err := r.db.QueryRow(ctx, query).Scan(&count)
	return count, err
}

// ─── Worker Queries ──────────────────────────────────────────────────────────

func (r *repository) ListSLABreachedTickets(ctx context.Context) ([]*SupportTicket, error) {
	query := `
		SELECT id, organization_id, subject, priority, status, created_by, assigned_to, sla_deadline, created_at, updated_at
		FROM support_tickets
		WHERE sla_deadline < NOW()
		  AND status NOT IN ('RESOLVED', 'CLOSED')
		ORDER BY sla_deadline ASC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []*SupportTicket
	for rows.Next() {
		var t SupportTicket
		if err := rows.Scan(&t.ID, &t.OrganizationID, &t.Subject, &t.Priority, &t.Status, &t.CreatedBy, &t.AssignedTo, &t.SLADeadline, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tickets = append(tickets, &t)
	}
	return tickets, nil
}

func (r *repository) ListUnresolvedSEV1Incidents(ctx context.Context) ([]*Incident, error) {
	query := `
		SELECT id, severity, status, description, detected_at, resolved_at, created_at, updated_at
		FROM incidents
		WHERE severity = 'SEV1'
		  AND status NOT IN ('RESOLVED', 'POSTMORTEM_PENDING')
		ORDER BY detected_at ASC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incidents []*Incident
	for rows.Next() {
		var inc Incident
		if err := rows.Scan(&inc.ID, &inc.Severity, &inc.Status, &inc.Description, &inc.DetectedAt, &inc.ResolvedAt, &inc.CreatedAt, &inc.UpdatedAt); err != nil {
			return nil, err
		}
		incidents = append(incidents, &inc)
	}
	return incidents, nil
}

// ─── Audit & Impersonation ────────────────────────────────────────────────

func (r *repository) CreateAuditLog(ctx context.Context, l *AuditLog) error {
	query := `
		INSERT INTO admin_audit_logs (id, actor_id, action, resource_type, resource_id, before_state, after_state, ip_address, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query, l.ID, l.ActorID, l.Action, l.ResourceType, l.ResourceID, l.BeforeState, l.AfterState, l.IPAddress, l.CreatedAt)
	return err
}

func (r *repository) CreateImpersonationSession(ctx context.Context, s *ImpersonationSession) error {
	query := `
		INSERT INTO impersonation_sessions (id, admin_id, user_id, started_at, expires_at, audit_log_id)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query, s.ID, s.AdminID, s.UserID, s.StartedAt, s.ExpiresAt, s.AuditLogID)
	return err
}

func (r *repository) UpdateImpersonationSession(ctx context.Context, s *ImpersonationSession) error {
	query := `UPDATE impersonation_sessions SET expires_at = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, s.ExpiresAt, s.ID)
	return err
}

func (r *repository) GetImpersonationSession(ctx context.Context, id uuid.UUID) (*ImpersonationSession, error) {
	query := `SELECT id, admin_id, user_id, started_at, expires_at, audit_log_id FROM impersonation_sessions WHERE id = $1`
	var s ImpersonationSession
	err := r.db.QueryRow(ctx, query, id).Scan(&s.ID, &s.AdminID, &s.UserID, &s.StartedAt, &s.ExpiresAt, &s.AuditLogID)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// ─── Approvals ──────────────────────────────────────────────────────────────

func (r *repository) CreateAdminOperationApproval(ctx context.Context, app *AdminOperationApproval) error {
	query := `INSERT INTO admin_operation_approvals (id, operation_type, target_id, payload, requested_by, status, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Exec(ctx, query, app.ID, app.OperationType, app.TargetID, app.Payload, app.RequestedBy, app.Status, app.CreatedAt)
	return err
}

func (r *repository) GetAdminOperationApproval(ctx context.Context, id uuid.UUID) (*AdminOperationApproval, error) {
	query := `SELECT id, operation_type, target_id, payload, requested_by, status, approved_by, created_at, resolved_at FROM admin_operation_approvals WHERE id = $1`
	var a AdminOperationApproval
	err := r.db.QueryRow(ctx, query, id).Scan(&a.ID, &a.OperationType, &a.TargetID, &a.Payload, &a.RequestedBy, &a.Status, &a.ApprovedBy, &a.CreatedAt, &a.ResolvedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *repository) UpdateAdminOperationApproval(ctx context.Context, app *AdminOperationApproval) error {
	query := `UPDATE admin_operation_approvals SET status = $1, approved_by = $2, resolved_at = $3 WHERE id = $4`
	_, err := r.db.Exec(ctx, query, app.Status, app.ApprovedBy, app.ResolvedAt, app.ID)
	return err
}

func (r *repository) ListPendingApprovals(ctx context.Context) ([]*AdminOperationApproval, error) {
	query := `SELECT id, operation_type, target_id, payload, requested_by, status, approved_by, created_at, resolved_at FROM admin_operation_approvals WHERE status = 'PENDING' ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var approvals []*AdminOperationApproval
	for rows.Next() {
		var a AdminOperationApproval
		if err := rows.Scan(&a.ID, &a.OperationType, &a.TargetID, &a.Payload, &a.RequestedBy, &a.Status, &a.ApprovedBy, &a.CreatedAt, &a.ResolvedAt); err != nil {
			return nil, err
		}
		approvals = append(approvals, &a)
	}
	return approvals, nil
}

// ─── Search Sync ────────────────────────────────────────────────────────────

func (r *repository) GetAdminSearchSyncState(ctx context.Context, domain string) (*AdminSearchSyncState, error) {
	query := `SELECT domain, last_synced_at FROM admin_search_sync_state WHERE domain = $1`
	var s AdminSearchSyncState
	err := r.db.QueryRow(ctx, query, domain).Scan(&s.Domain, &s.LastSyncedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *repository) UpdateAdminSearchSyncState(ctx context.Context, state *AdminSearchSyncState) error {
	query := `UPDATE admin_search_sync_state SET last_synced_at = $1 WHERE domain = $2`
	_, err := r.db.Exec(ctx, query, state.LastSyncedAt, state.Domain)
	return err
}
