# Admin Support Module — Technical Documentation

## Overview

The Admin Support Module is the **operational command center** of the FlyFleet platform, providing comprehensive tenant governance, support operations, incident management, and emergency system control.

**Module ID**: 1  
**Service Port**: Integrated in fleet-core-service (monolith)  
**Route Prefix**: `/admin`

---

## Architecture

```
admin_support
├── organization          # Tenant lifecycle management
├── admin_users           # Platform admin accounts
├── rbac                  # Role-based access control
├── support_tickets       # Customer support system
├── incidents             # Platform incident management
├── system_flags          # Feature flags & toggles
├── maintenance           # Scheduled maintenance windows
├── impersonation         # Admin impersonation sessions
├── audit_logs            # Immutable audit trail
└── workers               # Background monitoring
```

---

## Database Schema

### Core Tables

| Table | Purpose | Key Fields |
|-------|---------|------------|
| `organizations` | Tenant records | id, name, plan_type, status, region |
| `organization_users` | Org-level users | org_id, email, role, status |
| `admin_users` | Platform admins | email, role, mfa_enabled |
| `admin_permissions` | Granular permissions | name (unique) |
| `admin_role_permissions` | Role-permission mapping | role, permission |
| `support_tickets` | Support requests | org_id, subject, priority, status, sla_deadline |
| `support_ticket_messages` | Ticket conversation | ticket_id, sender_id, message |
| `support_ticket_attachments` | Ticket files | ticket_id, file_url |
| `incidents` | Platform issues | severity, status, description |
| `incident_updates` | Incident timeline | incident_id, update_message |
| `system_flags` | Feature toggles | key, value |
| `maintenance_windows` | Scheduled maintenance | service, start_time, end_time |
| `impersonation_sessions` | Admin impersonation | admin_id, user_id, expires_at |
| `admin_audit_logs` | Audit trail | actor_id, action, resource_type, before/after |

### Status Enums

**Organization Status**: `ACTIVE`, `SUSPENDED`, `DELETED`  
**Ticket Priority**: `LOW`, `MEDIUM`, `HIGH`, `CRITICAL`  
**Ticket Status**: `OPEN`, `IN_PROGRESS`, `WAITING_CUSTOMER`, `RESOLVED`, `CLOSED`, `ESCALATED`  
**Incident Severity**: `SEV1`, `SEV2`, `SEV3`, `SEV4`  
**Incident Status**: `OPEN`, `MITIGATING`, `RESOLVED`, `POSTMORTEM_PENDING`  
**Admin Roles**: `SUPER_ADMIN`, `PLATFORM_ADMIN`, `SUPPORT_AGENT`, `FINANCE_ADMIN`, `READ_ONLY_AUDITOR`

---

## API Endpoints

### Organization Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/admin/orgs` | Create organization |
| `GET` | `/admin/orgs` | List all organizations |
| `GET` | `/admin/orgs/{id}` | Get organization details |
| `PATCH` | `/admin/orgs/{id}` | Update organization |
| `POST` | `/admin/orgs/{id}/suspend` | Suspend tenant |
| `POST` | `/admin/orgs/{id}/activate` | Activate tenant |
| `DELETE` | `/admin/orgs/{id}` | Delete organization (soft) |
| `POST` | `/admin/orgs/{id}/users` | Add organization user |
| `GET` | `/admin/orgs/{id}/users` | List organization users |

### Support Tickets

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/admin/tickets` | Create ticket |
| `GET` | `/admin/tickets` | List tickets (query: `organization_id`) |
| `GET` | `/admin/tickets/{id}` | Get ticket details |
| `POST` | `/admin/tickets/{id}/reply` | Add message |
| `POST` | `/admin/tickets/{id}/assign` | Assign to admin |
| `POST` | `/admin/tickets/{id}/close` | Resolve ticket |
| `POST` | `/admin/tickets/{id}/attachments` | Add attachment |
| `GET` | `/admin/tickets/{id}/attachments` | List attachments |

### Incident Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/admin/incidents` | Declare incident |
| `GET` | `/admin/incidents` | List all incidents |
| `POST` | `/admin/incidents/{id}/update` | Post timeline update |
| `POST` | `/admin/incidents/{id}/resolve` | Resolve incident |

### System Controls

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/admin/system/flag` | Set system flag |
| `GET` | `/admin/system/flags` | List all flags |
| `POST` | `/admin/system/maintenance` | Start maintenance window |
| `GET` | `/admin/maintenance` | List maintenance windows |

### RBAC Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/admin/rbac/permissions` | List all permissions |
| `GET` | `/admin/rbac/roles/{role}/permissions` | Get role permissions |
| `POST` | `/admin/rbac/roles/{role}/permissions` | Set role permissions |

### Impersonation

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/admin/impersonate/start` | Start impersonation |
| `POST` | `/admin/impersonate/stop` | Stop impersonation |

---

## SLA Definitions

| Priority | Response Time | Escalation |
|----------|--------------|------------|
| `CRITICAL` | 1 hour | Auto-escalate + PagerDuty |
| `HIGH` | 6 hours | Auto-escalate |
| `MEDIUM` | 24 hours | Email notification |
| `LOW` | 72 hours | None |

**SLA Monitor**: Runs every 5 minutes, detects violations, auto-escalates tickets.

---

## Background Workers

### 1. SLA Monitor Worker

**Interval**: 5 minutes  
**Purpose**: Detect and escalate tickets exceeding SLA deadlines

```go
worker := NewSLAMonitorWorker(repo, metrics, logger)
go worker.Start(ctx)
```

**Actions**:
- Scans all open tickets
- Checks `sla_deadline` vs current time
- Auto-escalates violated tickets to `ESCALATED` status
- Posts system message to ticket
- Increments `admin_support_sla_exceeded_total` metric

---

### 2. Incident Monitor Worker

**Interval**: 1 minute  
**Purpose**: Alert on unresolved SEV1/SEV2 incidents

```go
worker := NewIncidentWorker(repo, metrics, logger)
go worker.Start(ctx)
```

**Actions**:
- Monitors unresolved incidents
- SEV1 → PagerDuty alert (every minute)
- SEV2 → Email/Slack notification

---

### 3. Maintenance Window Worker

**Interval**: 1 minute  
**Purpose**: Activate/deactivate maintenance windows

```go
worker := NewMaintenanceWorker(repo, metrics, logger)
go worker.Start(ctx)
```

**Actions**:
- Activates maintenance flags when window starts
- Clears flags when window expires
- Sets `maintenance_{service}` system flags

---

## Security Model

### Authentication & Authorization

- **MFA Required**: All admin users must have MFA enabled
- **IP Allowlist**: Admin endpoints restricted to known IPs
- **Session Timeout**: 30 minutes for impersonation, 8 hours for normal admin
- **Rate Limiting**: Applied to sensitive operations

### RBAC Permissions

| Permission | Description | Roles |
|------------|-------------|-------|
| `org.manage` | Manage organizations | SUPER_ADMIN, PLATFORM_ADMIN, FINANCE_ADMIN |
| `org.suspend` | Suspend tenants | SUPER_ADMIN, PLATFORM_ADMIN |
| `org.delete` | Delete organizations | SUPER_ADMIN |
| `ticket.manage` | Manage tickets | SUPER_ADMIN, PLATFORM_ADMIN, SUPPORT_AGENT |
| `ticket.assign` | Assign tickets | SUPER_ADMIN, PLATFORM_ADMIN, SUPPORT_AGENT |
| `ticket.close` | Close tickets | SUPER_ADMIN, PLATFORM_ADMIN, SUPPORT_AGENT |
| `incident.manage` | Manage incidents | SUPER_ADMIN, PLATFORM_ADMIN |
| `incident.declare` | Declare incidents | SUPER_ADMIN, PLATFORM_ADMIN, SUPPORT_AGENT |
| `incident.resolve` | Resolve incidents | SUPER_ADMIN, PLATFORM_ADMIN |
| `system.override` | System overrides | SUPER_ADMIN |
| `system.flag.manage` | Manage flags | SUPER_ADMIN, PLATFORM_ADMIN |
| `system.maintenance` | Schedule maintenance | SUPER_ADMIN, PLATFORM_ADMIN |
| `dr.execute` | Disaster recovery | SUPER_ADMIN |
| `impersonation.start` | Start impersonation | SUPER_ADMIN, PLATFORM_ADMIN |
| `audit.view` | View audit logs | All admin roles |
| `admin.manage` | Manage admins | SUPER_ADMIN |

### High-Risk Operations (Two-Person Approval Required)

- DR failover execution
- Tenant deletion
- Billing override
- System-wide flag changes

---

## Observability

### Metrics (Prometheus)

| Metric | Type | Description |
|--------|------|-------------|
| `admin_support_tickets_open_total` | Gauge | Current open tickets |
| `admin_support_tickets_resolved_total` | Counter | Total resolved tickets |
| `admin_support_incidents_total{severity}` | CounterVec | Incidents by severity |
| `admin_support_actions_total{action,resource_type}` | CounterVec | Admin actions |
| `admin_support_impersonation_sessions_active` | Gauge | Active impersonations |
| `admin_support_sla_exceeded_total` | Counter | SLA violations |

### Logs

- `admin_activity_logs` — All admin actions
- `incident_timeline` — Incident updates
- `ticket_activity` — Ticket messages and status changes
- `impersonation_sessions` — Session start/stop

---

## Usage Examples

### Create Organization

```bash
curl -X POST http://localhost:8081/admin/orgs \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Logistics",
    "plan_type": "enterprise",
    "region": "us-east-1"
  }'
```

### Suspend Organization

```bash
curl -X POST http://localhost:8081/admin/orgs/{id}/suspend \
  -H "Authorization: Bearer <token>"
```

### Create Support Ticket

```bash
curl -X POST http://localhost:8081/admin/tickets \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "organization_id": "uuid",
    "subject": "Vehicle tracking issue",
    "priority": "HIGH"
  }'
```

### Declare SEV1 Incident

```bash
curl -X POST http://localhost:8081/admin/incidents \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "severity": "SEV1",
    "description": "GPS tracking service outage"
  }'
```

### Set System Flag

```bash
curl -X POST http://localhost:8081/admin/system/flag \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "key": "tracking_pause",
    "value": "true"
  }'
```

### Start Impersonation

```bash
curl -X POST http://localhost:8081/admin/impersonate/start \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "uuid"
  }'
```

---

## Migration

Apply the database schema:

```bash
psql -U fleet_user -d fleet_db -f services/fleet-core-service-legacy/internal/domains/admin_support/migrations/001_admin_support_schema.sql
```

---

## Testing

```bash
cd services/fleet-core-service-legacy
go test ./internal/domains/admin_support/... -v
```

---

## Future Enhancements

- [ ] PagerDuty integration for SEV1 incidents
- [ ] Email/Slack notifications for SEV2
- [ ] Two-person approval workflow for high-risk operations
- [ ] Advanced audit log querying and export
- [ ] Custom SLA definitions per organization plan
- [ ] Automated postmortem generation
- [ ] DR failover automation
- [ ] Admin activity anomaly detection

---

## Related Modules

- **Alert Service** — Receives incident alerts
- **Integrations Service** — Third-party notification providers
- **Reports Service** — Admin activity reports
- **Governance Service** — Compliance and policy enforcement

---

## Support

For issues or questions, create a ticket via the internal support system or contact the platform team.
