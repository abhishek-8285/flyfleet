# FlyFleet Enterprise — Full System PRD

**Version**: 2.0  
**Date**: April 3, 2026  
**Status**: Draft — Validated Against Complete Codebase + Engineering Audit  
**Scope**: Entire platform (392 DB tables, 27 backend domains, 180+ frontend pages, 3 frontend apps)  
**Validation**: Senior Backend Engineering Review + RICE Prioritization (18 items scored)

---

## 1. System Overview

### 1.1 What FlyFleet Is
FlyFleet is a **multi-tenant SaaS fleet management platform** for Indian logistics companies. It unifies vehicle tracking, driver management, dispatch, billing, compliance, maintenance, emergency response, and governance into a single enterprise dashboard.

### 1.2 Architecture Summary
```
┌─────────────────────────────────────────────────────────────┐
│                    Frontend Applications                     │
│  dashboard-web (React/Vite) │ landing-page (Next.js)        │
│  mobile (React Native)                                       │
└────────────────────────┬────────────────────────────────────┘
                         │ HTTPS + JWT (RS256)
┌────────────────────────▼────────────────────────────────────┐
│              Fleet-Core Monolith (Go / Chi)                  │
│  27 domains │ 11 API route groups │ JWKS Auth │ RLS Tenant  │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────┐
│              PostgreSQL 15 (392 tables)                      │
│  Row-Level Security │ Partitioned telemetry │ PostGIS        │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────┐
│              Redis 7 │ NATS │ Shared Pool Architecture       │
│  Sessions │ Rate Limits │ Pub/Sub │ Cache │ Dead Letter      │
└─────────────────────────────────────────────────────────────┘
```

### 1.3 Current State
| Aspect | Status |
|--------|--------|
| **Active backend** | Single monolith (`fleet-core`) — 27 domains consolidated |
| **Archived services** | 35 microservices moved to `_archive/` (consolidated into monolith) |
| **Active frontend** | `dashboard-web` (React/Vite) — 180+ pages |
| **Other frontends** | `landing-page` (Next.js), `mobile` (React Native) — exist but not wired |
| **Database** | 392 tables, partitioned telemetry, PostGIS spatial support |
| **Auth** | JWKS + RSA-256, refresh token rotation, MFA, RBAC (6 roles) |
| **Multi-tenant** | Row-Level Security on all org-scoped tables |

---

## 2. Problem Definition

### 2.1 Customer Problem

**Who**: Fleet operators in India managing 10-500 vehicles — logistics companies, delivery fleets, transport agencies, e-commerce delivery partners

**What**: 
- Can't see where vehicles are in real-time (rely on phone calls to drivers)
- Dispatch is manual (WhatsApp groups, phone calls, paper logs)
- Billing is error-prone (Excel sheets, manual calculations, 15-20% revenue leakage)
- Compliance is reactive (FastTag reconciliation, e-way bills — penalties after the fact)
- No audit trail for disputes (who did what, when, why)

**When**: Daily operations — dispatchers need this 8-12 hours/day, fleet managers review weekly

**Where**: Desktop dashboard (office operations) + mobile (drivers on road, future phase)

**Why**: Existing solutions are either:
- Too expensive (Samsara: $20+/vehicle/month = ₹1,600+/vehicle/month)
- Too basic (GPS-only trackers with no billing/dispatch/compliance)
- Not India-specific (no FastTag, e-way bill, Vahan/Sarathi integration)
- Poor UX (legacy enterprise software from 2010s)

**Impact per customer**:
- 15-20% revenue leakage from billing errors and untracked deliveries
- 2-3 hours/day wasted on manual dispatch coordination
- ₹10-50K/month in avoidable compliance penalties (FastTag mismatches, missing e-way bills)
- 30-45 minute response time to emergencies (accidents, breakdowns)

### 2.2 Market Opportunity

| Metric | Value | Source |
|--------|-------|--------|
| TAM (India fleet management) | ₹8,000 Cr ($1B) | Industry reports |
| SAM (SME fleets 10-500 vehicles) | ₹2,400 Cr ($300M) | Addressable segment |
| SOM (Year 1 target) | ₹60 Cr ($7.5M) | 500 orgs × ₹10K/month avg |
| Growth rate | 18% CAGR | Indian logistics market |

**Competition**:
| Competitor | Price/vehicle/mo | Gap |
|------------|-----------------|-----|
| Samsara | ₹1,600+ | Too expensive for Indian SMEs |
| LocoNav | ₹500 | No billing/dispatch, GPS-only |
| Fleetx.io | ₹800 | Poor UX, no multi-tenant SaaS |
| **FlyFleet** | **₹300** | **Unified platform, India-native, modern UX** |

### 2.3 Business Case

**Revenue Model**: Per-vehicle subscription + transaction fees
| Tier | Price | Features |
|------|-------|----------|
| Starter | ₹300/vehicle/mo | Tracking + basic dispatch |
| Professional | ₹500/vehicle/mo | + Billing + Notifications |
| Enterprise | ₹800/vehicle/mo | + Compliance + API access |

**Projected MRR**:
| Phase | Timeline | Customers | Vehicles | MRR |
|-------|----------|-----------|----------|-----|
| Alpha | Q2 2026 | 3 | 50 | ₹15K |
| Beta | Q3 2026 | 20 | 500 | ₹2.5L |
| Launch | Q4 2026 | 100 | 3,000 | ₹15L |
| Scale | Q2 2027 | 500 | 15,000 | ₹75L |

**Risk if we don't build**: 
- Competitors will fill the gap (LocoNav adding billing, Fleetx improving UX)
- Indian logistics market growing 18% YoY — window is 12-18 months
- First-mover advantage in unified platform (tracking + billing + compliance)

---

## 3. Solution Overview

### 3.1 Proposed Solution
A web-based fleet management dashboard with:
1. **Real-time vehicle tracking** on interactive map
2. **Dispatch workflow** — assign driver + vehicle + route
3. **Billing & invoicing** — auto-generate invoices per delivery
4. **India compliance** — FastTag reconciliation, e-way bill generation
5. **Multi-tenant** — one platform, multiple organizations

### 3.2 In Scope (Q1-Q4)
- ✅ IAM / User Management (Q2) — login, roles, org switching, password reset
- ✅ Live Vehicle Map (Q3) — Leaflet/MapLibre with real-time positions
- ✅ Dispatch Flow (Q3) — assign → track → complete delivery
- ✅ Billing/Invoices (Q1) — create, send, record payment, PDF export
- ✅ Notifications (Q4) — email/SMS for geofence, delays, due invoices
- ✅ Payment Gateway (Q4) — Razorpay integration for invoice payments

### 3.3 Out of Scope
- ❌ Driver mobile app (Phase 2)
- ❌ AI route optimization (Phase 3)
- ❌ Predictive maintenance (Phase 3)
- ❌ Multi-country compliance (Phase 3)
- ❌ White-label reseller portal (Phase 3)

### 3.4 MVP Definition (Q3 Target)
**Core features**:
1. User can log in and see their organization
2. User can see vehicles on a map with last-known position
3. User can create a delivery assignment (vehicle + driver + route)
4. User can generate an invoice for a completed delivery

**Success criteria**:
- 3 beta customers actively using daily
- <2s map load time
- <5s GPS position update latency
- Zero auth/session-loss bugs

**Learning goals**:
- Do fleet operators actually use the map daily?
- Is dispatch workflow intuitive without training?
- What's the #1 complaint? (prioritize Phase 2)

---

## 4. User Stories & Requirements

### 4.1 User Stories

#### IAM / Auth (Q2 — P0)
```
As a fleet operator admin
I want to log in securely and manage my team's access
So that only authorized people can see my fleet data

Acceptance Criteria:
- [ ] Login with email + password returns JWT + refresh token
- [ ] Refresh token rotation works without session loss
- [ ] Role-based access control (super_admin, org_admin, dispatcher, driver)
- [ ] User CRUD — create, update, deactivate users
- [ ] Password reset via email
- [ ] MFA optional but supported
- [ ] Session management — view active sessions, revoke
```

#### Vehicle Tracking (Q3 — P0)
```
As a dispatcher
I want to see all my vehicles on a map with live positions
So that I know where every vehicle is right now

Acceptance Criteria:
- [ ] Map loads with all organization vehicles in <2s
- [ ] Vehicle markers update every 5 seconds
- [ ] Clicking a vehicle shows: driver name, speed, last update, status
- [ ] Geofence alerts when vehicle enters/exits zone
- [ ] Vehicle history trail for last 24 hours
```

#### Dispatch Flow (Q3 — P0)
```
As a dispatcher
I want to assign a vehicle + driver to a delivery route
So that the delivery is tracked from start to completion

Acceptance Criteria:
- [ ] Create delivery: pickup location → drop location → assigned vehicle/driver
- [ ] Dispatcher sees delivery status: pending → in-transit → delivered
- [ ] Driver can mark delivery complete (via web or future mobile app)
- [ ] Auto-generate delivery receipt with timestamp + location
```

#### Billing/Invoices (Q1 — P0)
```
As a fleet operator
I want to generate invoices for deliveries automatically
So that I bill customers accurately and get paid faster

Acceptance Criteria:
- [ ] Auto-create invoice from completed delivery
- [ ] Edit invoice: add line items, adjust amounts, apply tax
- [ ] Send invoice via email with PDF attachment
- [ ] Record payment (full/partial)
- [ ] Dashboard shows: total invoiced, outstanding, overdue
- [ ] Razorpay payment link in email invoice
```

### 4.2 Functional Requirements

| ID | Requirement | Priority | Notes |
|----|------------|----------|-------|
| FR1 | User authentication with JWT + refresh tokens | P0 | Fixed token reuse bug |
| FR2 | Role-based access control (RBAC) | P0 | 6 roles defined |
| FR3 | User CRUD with org context | P0 | IAM routes were commented out |
| FR4 | Vehicle position ingestion API | P0 | Accept GPS data from devices |
| FR5 | Interactive map with vehicle markers | P0 | Leaflet + WebSocket |
| FR6 | Delivery creation workflow | P0 | Vehicle + driver + route |
| FR7 | Invoice generation with PDF | P0 | Already implemented, needs wiring |
| FR8 | Payment recording + Razorpay | P1 | Phase 4 |
| FR9 | Email notifications for deliveries | P1 | Phase 3 |
| FR10 | Geofence alerts | P2 | Phase 3 |
| FR11 | FastTag reconciliation | P2 | India-specific |
| FR12 | E-way bill generation | P2 | India-specific |

### 4.3 Non-Functional Requirements

| Category | Requirement | Target | Current |
|----------|------------|--------|---------|
| **Performance** | Dashboard load time | <2s P95 | 170ms ✅ |
| **Performance** | GPS position update latency | <5s | N/A (no real GPS) |
| **Performance** | Invoice PDF generation | <3s | N/A (endpoint exists) |
| **Performance** | Login response time | <200ms | 531ms ⚠️ |
| **Scalability** | Concurrent users per org | 50 | Untested |
| **Scalability** | Vehicles per org | 500 | Untested |
| **Scalability** | GPS updates/sec | 10,000 | Untested |
| **Security** | Auth success rate | >99.5% | ~100% ✅ |
| **Security** | Token reuse incidents | 0/month | 0 (fixed) ✅ |
| **Security** | Rate limiting coverage | 100% endpoints | 10% (4/41) ❌ |
| **Security** | TLS/HTTPS | Enforced | HTTP only ❌ |
| **Reliability** | Uptime | 99.9% | Untested |
| **Reliability** | Error rate | <1% of requests | 0% (dev) ✅ |
| **Usability** | Mobile-responsive dashboard | Yes | Partial |
| **Compliance** | Data residency | India (Mumbai) | Localhost (dev) |

---

## 5. Design & User Experience

### 5.1 Design Principles
1. **Operator-first** — Dispatchers use this 8-12 hours/day. Every click matters.
2. **Real-time by default** — If it's not live, it's not useful for fleet ops.
3. **India-native** — FastTag, e-way bills, Hindi support, ₹ pricing.
4. **Progressive disclosure** — Simple dashboard for daily use, deep dive for admins.

### 5.2 Information Architecture
```
Dashboard (home)
├── Map (live vehicles)
├── Dispatch (active deliveries)
├── Vehicles (fleet overview)
├── Drivers (team management)
├── Trips (history + analytics)
├── Billing (invoices + payments)
├── Fuel (logs + reconciliation)
├── Maintenance (work orders)
├── Alerts (geofence, delays)
├── Reports (export + schedule)
├── Settings (org, users, roles)
└── Admin (security, RBAC, platform config)
```

### 5.3 Key User Flows
```
Login → Dashboard → Map → Click vehicle → See driver + speed + location
Login → Dispatch → Create trip → Assign vehicle + driver → Track → Complete → Invoice
Login → Billing → View invoices → Send → Customer pays via Razorpay → Record payment
```

---

## 6. Go-to-Market Strategy

### 6.1 Launch Plan

| Phase | Timeline | Target | Success Criteria |
|-------|----------|--------|-----------------|
| **Alpha** | Q2 2026 | 3 beta customers (friends/family fleets) | Daily active use, <2% error rate |
| **Beta** | Q3 2026 | 20 customers (Delhi NCR logistics) | ₹2.5L MRR, NPS >40 |
| **Soft Launch** | Q4 2026 | 100 customers (paid) | ₹15L MRR, <1% churn |
| **Full Launch** | Q1 2027 | 500 customers | ₹75L MRR, 18% MoM growth |

### 6.2 Pricing Strategy

| Tier | Price | Features | Target Segment |
|------|-------|----------|---------------|
| **Starter** | ₹300/vehicle/mo | Tracking + basic dispatch | Small fleets (10-50 vehicles) |
| **Professional** | ₹500/vehicle/mo | + Billing + Notifications | Medium fleets (50-200 vehicles) |
| **Enterprise** | ₹800/vehicle/mo | + Compliance + API access | Large fleets (200-500 vehicles) |

### 6.3 Success Metrics

| Metric | Alpha (Q2) | Beta (Q3) | Launch (Q4) | Scale (Q1 2027) |
|--------|------------|-----------|-------------|-----------------|
| Active orgs | 3 | 20 | 100 | 500 |
| Vehicles tracked | 50 | 500 | 3,000 | 15,000 |
| Daily active users | 5 | 40 | 200 | 1,000 |
| Invoices generated/mo | 100 | 2,000 | 15,000 | 75,000 |
| MRR | ₹15K | ₹2.5L | ₹15L | ₹75L |
| Error rate | <5% | <2% | <1% | <0.5% |
| NPS | — | >40 | >50 | >60 |

---

## 7. Risks & Mitigations

| Risk | Probability | Impact | Mitigation | Owner |
|------|------------|--------|------------|-------|
| GPS device integration complexity | High | High | Start with mock data in Q3, integrate real devices in Q5 | Engineering |
| User adoption too slow | Medium | High | Beta program with hands-on onboarding for first 20 customers | Product |
| Scope creep | High | Medium | Weekly stakeholder reviews, strict P0/P1/P2 prioritization | Product |
| Tech debt accumulation | High | Medium | 20% capacity reserved for refactoring (already biting us) | Engineering |
| Competitor launches similar product | Medium | High | Speed to market — ship Q3 alpha before LocoNav adds billing | Product |
| Payment gateway compliance delays | Low | Medium | Start Razorpay integration in Q3, not Q4 | Engineering |
| Security breach (no TLS) | High | Critical | Q1: TLS enforcement, rate limiting, SQL audit | Engineering |
| Data loss (no backups tested) | Medium | Critical | Q2: DR test with restore verification | Engineering |

---

## 8. Timeline & Milestones (Combined Product + Engineering)

Based on RICE prioritization of 18 items (product features + engineering gaps):

| Quarter | Capacity | Deliverables | RICE Score | Success Criteria |
|---------|----------|--------------|------------|-----------------|
| **Q1** | 11/15 pm | TLS/HTTPS Enforcement | 10,000 | HTTPS enforced, HSTS headers |
| | | Unified Health Endpoint | 6,667 | Single /health checks all deps |
| | | SQL Injection Audit | 6,000 | 42 queries audited, 0 vulns |
| **Q2** | 13/15 pm | Complete IAM / User Management | 4,000 | User CRUD, no session bugs |
| | | Rate Limiting Coverage | 4,000 | 41/41 endpoint groups covered |
| | | Login Performance Optimization | 3,333 | Login <200ms P50 |
| **Q3** | 13/15 pm | Live Vehicle Tracking on Map | 2,308 | 3 beta customers using daily |
| **Q4** | 8/15 pm | API Versioning Strategy | 1,667 | /api/v2/ strategy documented |
| | | Notification System | 1,600 | Email alerts, <5% bounce rate |
| **Q5** | 13/15 pm | GPS Telematics Integration | 1,292 | Real device data flowing |
| **Q6** | 13/15 pm | Working Dispatch Flow | 1,231 | 50 deliveries tracked |
| **Q7** | 11/15 pm | Multi-Org Onboarding Flow | 1,067 | Self-service org creation |
| | | Real Payment Gateway Integration | 1,000 | Razorpay live, ₹5L processed |
| **Q8** | 13/15 pm | Fleet Analytics Dashboard | 800 | Real metrics, not stub data |
| | | India Compliance APIs | 750 | FastTag + e-way bill live |
| **Q9** | 13/15 pm | Driver Mobile App | 738 | Android app, 50 active drivers |
| **Q10** | 8/15 pm | Test Coverage Report | 667 | 80%+ coverage, gaps documented |
| | | OpenAPI/Swagger Spec | 600 | API docs auto-generated |

---

## 9. Team & Resources

### 9.1 Team Structure
| Role | Count | Responsibility |
|------|-------|----------------|
| Backend Engineer (Go) | 2 | APIs, services, database, security |
| Frontend Engineer (React) | 1 | Dashboard, map UI, forms |
| DevOps | 0.5 | CI/CD, infrastructure, monitoring |
| Product | 1 | PRD, prioritization, customer feedback |
| **Total** | **4.5 FTE** | |

### 9.2 Budget (Monthly)
| Item | Cost |
|------|------|
| Cloud infrastructure (AWS/GCP) | ₹15,000 |
| Razorpay fees | 2% per transaction |
| Email/SMS (SendGrid + MSG91) | ₹5,000 |
| Domain + SSL | ₹2,000 |
| **Total** | **~₹22,000/month** |

---

## 10. Backend Domains (27 Active)

### 2.1 Core Domains

| Domain | Tables | Status | Description |
|--------|--------|--------|-------------|
| **auth** | 15+ | ✅ Live | Users, roles, permissions, MFA, sessions, refresh tokens, signing keys, OIDC providers, invites, password resets |
| **vehicle** | 12+ | ✅ Live | Vehicles, specs, assignments, events, compliance, financials, inspections, downtime, lifecycle metrics |
| **driver** | 10+ | ✅ Live | Driver profiles, onboarding, shifts, training, payroll, ratings, safety records, mobile sync |
| **trip** | 5+ | ✅ Live | Trips, trip settings, profitability, admin trip management, telemetry |
| **tracking** | 8+ | ✅ Live | Telemetry ingestion, GPS locations, OBD readings, vehicle telemetry, analytics, anomalies, outbox |
| **route** | 4+ | ✅ Live | Route optimization jobs, requests, results |

### 2.2 Business Domains

| Domain | Tables | Status | Description |
|--------|--------|--------|-------------|
| **billing** | 15+ | ✅ Live (endpoints) | Invoices, invoice items, payments, settings, metrics, templates, sequences, coupons, subscriptions, pricing components, tax rates, revenue schedules, accounting exports |
| **fuel** | 8+ | ⚠️ Partial | Fuel logs, transactions, readings, vendors, settings, audit logs, card reconciliation, efficiency history |
| **maintenance** | 6+ | ⚠️ Partial | Work orders, parts, technicians, vendors, windows, schedules, settings |
| **finance** | 5+ | ⚠️ Partial | Expense items, reports, policies, audit logs, retention policies |
| **geofencing** | 3+ | ⚠️ Partial | Geofences, assignments, events |

### 2.3 Compliance & Governance

| Domain | Tables | Status | Description |
|--------|--------|--------|-------------|
| **india** | 6+ | ⚠️ Stub | FastTag transactions, toll transactions, e-way bills, Vahan/Sarathi integration, GST compliance, local permits |
| **governance** | 7+ | ✅ Live | DSAR requests, compliance reports, exports, audit logs, settings, download tokens, export lineage |
| **compliance** | 3+ | ⚠️ Partial | Vehicle compliance, authority zones, authority delivery logs |

### 2.4 Platform & Infrastructure

| Domain | Tables | Status | Description |
|--------|--------|--------|-------------|
| **alert** | 12+ | ✅ Live | Alert rules, groups, thresholds, v2 templates, logs, DLQ, user preferences |
| **analytics** | 6+ | ⚠️ Partial | Jobs, quotas, access logs, query logs, query plans |
| **monitoring** | 5+ | ✅ Live | Health checks, SLA metrics, synthetic checks, error budgets, service SLOs |
| **diagnostics** | 6+ | ⚠️ Partial | Device health, telemetry, commands, firmware rollouts, packages |
| **integrations** | 4+ | ⚠️ Partial | Integration configs, logs, webhooks, ERP configs |
| **settings** | 4+ | ⚠️ Partial | Settings audit logs, feature flags, platform settings, system limits |
| **reports** | 3+ | ⚠️ Partial | Report definitions, schedules, execution history |
| **video** | 4+ | ⚠️ Stub | Video streams, recordings, analytics, retention policies |
| **emergency** | 9+ | ⚠️ Partial | Incidents, commands, contacts, escalations, response units, status, vehicle state, snapshots, archive |
| **customer** | 3+ | ⚠️ Stub | Customer profiles, shipments, proof of delivery |
| **platform_compat** | 0 | ✅ Stub | Compatibility layer for legacy API migration |
| **admin_support** | 8+ | ⚠️ Partial | Organizations, DR backups/restores, support tickets, incidents, emergency response |

### 2.5 Cross-Cutting

| Domain | Tables | Status | Description |
|--------|--------|--------|-------------|
| **security** | 20+ | ✅ Live | Security events, audit logs, auth audit, MFA events/failures/lockouts, threat alerts, IP rules, impersonation sessions |
| **audit** | 6+ | ✅ Live | Audit logs, policies, settings, entity history, evidence exports, legal holds |
| **notifications** | 5+ | ⚠️ Partial | Channels, events, logs, preferences, settings, global defaults |

---

## 11. Database Schema (392 Tables by Domain)

### 3.1 Auth & Identity (15 tables)
`users`, `roles`, `permissions`, `user_roles`, `role_permissions`, `admin_permissions`, `admin_role_permissions`, `refresh_tokens`, `auth_tokens`, `mfa_codes`, `mfa_events`, `mfa_failures`, `mfa_lockouts`, `mfa_recovery_codes`, `mfa_users`, `mfa_policies`, `mfa_policy_history`, `mfa_exceptions`, `mfa_org_overrides`, `password_reset_tokens`, `invites`, `signing_keys`, `oidc_providers`, `oauth_clients`, `service_accounts`, `api_keys`, `api_clients`, `user_oidc_links`, `user_links`, `organization_memberships`, `organization_users`, `user_organizations`

### 3.2 Vehicles & Tracking (20+ tables)
`vehicles`, `vehicles_legacy`, `vehicle_specs`, `vehicle_assignments`, `vehicle_compliance`, `vehicle_events`, `vehicle_financials`, `vehicle_inspections`, `vehicle_lifecycle_metrics`, `vehicle_downtime`, `vehicle_daily_stats`, `vehicle_verifications`, `vehicle_telemetry`, `telemetry`, `telemetry_points` (partitioned by month), `telemetry_events` (partitioned by month), `telemetry_analytics`, `telemetry_anomalies`, `telemetry_outbox`, `gps_locations`, `obd_readings`, `tracking_config`, `fleet_analytics`, `fleet_hourly_stats`, `materialized_fleet_reports`

### 3.3 Drivers & Trips (10+ tables)
`trips`, `trip_settings`, `trip_profitability`, `driver_profiles`, `driver_onboarding`, `driver_shifts`, `driver_training`, `driver_payroll`, `driver_ratings`, `driver_safety`, `driver_mobile_sync`

### 3.4 Billing & Finance (25+ tables)
`billing_invoices`, `invoice_items`, `invoice_sequence`, `invoice_permissions`, `invoice_audit_log`, `invoice_approvals`, `invoice_approval_policies`, `invoice_approval_matrix`, `user_invoice_access`, `billing_settings`, `billing_configurations`, `billing_plans`, `billing_subscriptions`, `billing_pricing_components`, `billing_tax_rates`, `billing_coupons`, `billing_credit_transactions`, `billing_customer_balances`, `billing_hardware_products`, `billing_invoice_documents`, `billing_invoice_templates`, `billing_invoice_sequences`, `billing_accounting_exports`, `billing_recognized_revenue`, `billing_revenue_schedules`, `payment_records`, `bank_transactions`, `finance_audit_logs`, `finance_retention_policies`, `expense_items`, `expense_reports`, `expense_policies`, `fuel_logs`, `fuel_transactions`, `fuel_readings`, `fuel_vendors`, `fuel_settings`, `fuel_audit_logs`, `fuel_card_transactions`, `fuel_card_reconciliation`, `fuel_efficiency_history`, `org_fuel_settings`, `receipt_ocr_jobs`, `vendor_payments`, `tds_records`, `purchase_orders`, `purchase_order_items`

### 3.5 Compliance & India (10+ tables)
`fastag_transactions`, `toll_transactions`, `gst_compliances`, `gst_invoices`, `local_permits`, `vehicle_compliance`, `authority_delivery_logs`, `india_integration_logs`, `ewaybills`, `vahan_records`, `sarathi_records`

### 3.6 Governance & Audit (15+ tables)
`governance_requests`, `governance_settings`, `governance_exports`, `governance_export_lineage`, `governance_download_tokens`, `governance_audit_logs`, `governance_compliance_reports`, `audit_logs`, `audit_policies`, `audit_settings`, `audit_history`, `evidence_exports`, `legal_holds`, `entity_audit_history`, `security_events`, `security_history`, `auth_audit_logs`, `system_audit_logs`, `system_flags`, `system_limits`

### 3.7 Alerts & Monitoring (15+ tables)
`alerts`, `alert_rules`, `alert_groups`, `alert_thresholds`, `alerts_v2_templates`, `alerts_v2_logs`, `alerts_v2_dlq`, `alerts_v2_user_preferences`, `threat_alerts`, `health_alerts`, `health_history`, `synthetic_checks`, `error_budgets`, `service_slos`, `monitoring_dashboards`

### 3.8 Maintenance & Work Orders (8+ tables)
`work_orders`, `work_order_parts`, `maintenance_schedule`, `maintenance_windows`, `maintenance_technicians`, `maintenance_vendors`, `maintenance_settings`, `technician_skills`, `inventory_parts`, `inspection_templates`, `inspection_documents`

### 3.9 Geofencing & Routing (6+ tables)
`geofences`, `geofence_assignments`, `geofence_events`, `route_optimization_jobs`, `route_optimization_requests`, `route_optimization_results`

### 3.10 Emergency & Safety (10+ tables)
`emergency_incidents`, `emergency_incidents_archive`, `emergency_incident_snapshots`, `emergency_commands`, `emergency_contacts`, `emergency_contact_responses`, `emergency_escalations`, `emergency_response_units`, `emergency_status`, `emergency_vehicle_state`, `sos_incidents`

### 3.11 Video & Diagnostics (8+ tables)
`video_streams`, `video_recordings`, `video_analytics`, `video_retention_policies`, `firmware_packages`, `firmware_rollouts`, `firmware_rollout_events`, `rollout_device_status`

### 3.12 Notifications (6+ tables)
`notification_channels`, `notification_events`, `notification_logs`, `notification_settings`, `notification_preference_audit`, `user_notification_preferences`, `global_notification_defaults`

### 3.13 Integrations & Webhooks (6+ tables)
`integrations`, `integration_configs`, `integration_logs`, `webhooks`, `webhook_subscriptions`, `webhook_event_types`, `erp_configs`

### 3.14 Support & Incidents (10+ tables)
`incidents`, `incident_comments`, `incident_logs`, `incident_media`, `incident_sla_violations`, `incident_timeline`, `incident_timeline_events`, `incident_updates`, `incident_vehicle_snapshot`, `support_tickets`, `support_ticket_messages`, `support_ticket_attachments`, `ticket_events`, `sla_metrics`, `support_audit_log`, `org_support_metrics`

### 3.15 Platform & Settings (10+ tables)
`platform_settings`, `feature_flags`, `settings_audit_logs`, `localization_settings`, `templates`, `ab_test_experiments`, `usage_meters`, `usage_records`, `org_budgets`, `id_mappings`, `sync_jobs`, `schema_migrations`, `schema_migrations_legacy`

### 3.16 Analytics & ML (8+ tables)
`analytics_jobs`, `analytics_quotas`, `analytics_access_logs`, `analytics_query_logs`, `analytics_query_plans`, `prediction_results`, `predictive_models`, `model_deployments`, `model_predictions`, `historical_demand`, `predictive_maintenance_analytics`, `telemetry_analytics`

### 3.17 Disaster Recovery (6+ tables)
`backups`, `backup_verifications`, `restore_jobs`, `admin_search_sync_state`, `admin_operation_approvals`, `admin_audit_logs`

### 3.18 Spatial (PostGIS)
`spatial_ref_sys`, `topology`, `layer`

---

## 12. Frontend Applications

### 4.1 Dashboard Web (React/Vite) — 180+ Pages

| Section | Pages | Route Prefix | Status |
|---------|-------|-------------|--------|
| **Auth** | Login, Register, MFA Setup/Verify, Forgot Password, Reset Password, Verify Email, Account Locked, Session Expired, Auth Error, SSO Callback | `/login`, `/register`, `/mfa-*`, `/forgot-password`, etc. | ✅ Live |
| **Dashboard** | Global stats, KPIs, activity feed | `/dashboard` | ✅ Live |
| **Vehicles** | List, detail, add, edit, bulk upload | `/vehicles`, `/vehicles/:id`, `/admin/vehicles` | ✅ Live |
| **Drivers** | List, detail, manage | `/drivers`, `/admin/drivers` | ✅ Live |
| **Trips** | List, add, edit, detail, telemetry, settings | `/trips`, `/trips/:id`, `/admin/trips` | ✅ Live |
| **Dispatch** | Dispatch center, route optimization | `/dispatch`, `/route-optimization` | ⚠️ Partial |
| **Map** | Live vehicle map | `/map` | ❌ Not wired |
| **Billing** | Dashboard, create, detail, edit, settings, designer | `/admin/billing`, `/admin/billing/new`, `/admin/billing/invoices/:id` | ✅ Live (endpoints) |
| **Fuel** | Log list, add log, settings | `/fuel/add`, `/fuel/settings` | ⚠️ Partial |
| **Maintenance** | Work orders, inventory, inspections, vendors | `/maintenance`, `/maintenance/work-orders`, `/inventory` | ⚠️ Partial |
| **Geofences** | List, create, manage | `/geofences` | ⚠️ Partial |
| **Alerts** | Alert management | `/alerts` | ✅ Live |
| **Analytics** | Fleet, driver behavior, predictive maintenance, models, jobs | `/analytics`, `/analytics/fleet`, `/analytics/drivers`, `/analytics/maintenance`, `/analytics/models`, `/analytics/jobs` | ⚠️ Partial |
| **Reports** | Create, list, history | `/reports`, `/reports/create`, `/reports/history` | ⚠️ Partial |
| **Finance** | Expense approvals, fuel dashboard, policy | `/finance/expenses`, `/finance/fuel`, `/finance/policy` | ⚠️ Partial |
| **Revenue** | Forecasting dashboard | `/revenue/forecasting` | ❌ Stub |
| **Security** | Auth stats, sessions, MFA center, threat alerts, IP controls, login activity, access review, user session admin | `/admin/security/*` | ✅ Live |
| **RBAC** | Roles, user assignments, security logs, role matrix | `/admin/rbac/*` | ✅ Live |
| **Settings** | Feature flags, localization, notification channels, platform config, policies/limits, templates | `/admin/settings/*` | ⚠️ Partial |
| **Notifications** | Global settings, user preferences, logs, detail | `/admin/notifications/*` | ⚠️ Partial |
| **Integrations** | Hub, API client, webhooks, marketplace, logs | `/integrations` | ⚠️ Partial |
| **Governance** | Settings, requests, detail | `/governance`, `/governance/requests/*` | ✅ Live |
| **Audit** | Center, list, policy, entity history, evidence export | `/audit/*` | ✅ Live |
| **Emergency** | Dashboard, incident management | `/emergency`, `/admin/emergency` | ⚠️ Partial |
| **DR** | Backups, restores, tests, settings | `/admin/dr/*` | ⚠️ Partial |
| **Devices** | Dashboard, diagnostics, telemetry, twins, firmware, commands | `/admin/devices/*` | ⚠️ Partial |
| **Video** | Video telematics | `/admin/video` | ❌ Stub |
| **Support** | Dashboard, incidents, org console, user console, SLA reports, troubleshooting | `/support`, `/support/incidents` | ⚠️ Partial |
| **Organizations** | List, detail, settings, users, vehicles, add, edit | `/admin/organizations`, `/admin/organizations/:id` | ⚠️ Partial |
| **Portal (Customer)** | Login, dashboard, shipments, tracking, invoices, bookings | `/portal/*` | ⚠️ Partial |
| **Monitoring (SRE)** | System health, metrics, logs, queue monitor, SLA report | `/admin/monitoring/*` | ⚠️ Partial |
| **Landing** | Marketing page | `/` | ✅ Live |

### 4.2 Landing Page (Next.js)
- Customer-facing marketing site
- Route: Separate app at `apps/landing-page/`

### 4.3 Mobile (React Native)
- Driver/customer mobile app
- Route: Separate app at `apps/mobile/`
- Status: Exists but not actively developed

---

## 13. API Endpoints (Registered Routes)

### 5.1 Auth (`/api/v1/auth/*`)
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/login` | Public | Email/password login |
| POST | `/refresh` | Cookie | Token refresh |
| POST | `/revoke` | Public | Token revocation |
| POST | `/logout` | Cookie | Logout |
| POST | `/register` | Public | User registration |
| POST | `/validate-token` | Public | Token validation |
| POST | `/login/mfa/verify` | Public | MFA verification |
| POST | `/login/mfa/setup` | Public | Admin MFA setup |
| POST | `/login/oidc/{provider}` | Public | OIDC login |
| POST | `/password/reset/request` | Public | Password reset request |
| POST | `/password/reset/confirm` | Public | Password reset confirm |
| POST | `/resolve-orgs` | Public | Resolve user organizations |
| POST | `/invites/validate` | Rate limited | Invite validation |
| POST | `/dev-login` | Dev only | Development auto-login |
| POST | `/mfa/generate-secret` | JWT | Generate MFA secret |
| POST | `/mfa/enable` | JWT | Enable MFA |
| POST | `/mfa/disable` | JWT | Disable MFA |
| POST | `/mfa/verify` | JWT | Verify MFA |
| POST | `/switch-context` | JWT | Switch org context |
| GET | `/users/{id}` | JWT | Get user |
| GET | `/organizations/{id}` | JWT | Get organization |
| GET | `/organizations` | JWT | List organizations |
| GET | `/stats` | Admin | Auth stats |
| GET | `/security-telemetry` | Admin | Security telemetry |
| GET | `/audit-logs` | Admin | Audit logs |
| POST | `/api-keys` | Admin | Create API key |
| GET | `/api-keys` | Admin | List API keys |
| POST | `/api-keys/{id}/revoke` | Admin | Revoke API key |
| POST | `/service-accounts` | Admin | Create service account |
| GET | `/service-accounts` | Admin | List service accounts |
| POST | `/emergency-revoke` | Super admin | Emergency revoke |
| GET | `/.well-known/jwks.json` | Public | JWKS endpoint |
| GET | `/auth/profile` | JWT | **Current user profile** |

### 5.2 Billing (`/api/v1/billing/*`)
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/health` | Public | Health check |
| GET | `/invoices` | JWT | List invoices |
| POST | `/invoices` | JWT | Create invoice |
| GET | `/invoices/{id}` | JWT | Get invoice |
| PUT | `/invoices/{id}` | JWT | Update invoice |
| DELETE | `/invoices/{id}` | JWT | Delete invoice |
| POST | `/invoices/{id}/payments` | JWT | Record payment |
| GET | `/invoices/{id}/pdf` | JWT | Download PDF |
| GET | `/settings` | JWT | Get billing settings |
| PUT | `/settings` | JWT | Update billing settings |
| GET | `/metrics` | JWT | Get billing metrics |

### 5.3 Dashboard (`/api/v1/dashboard/*`)
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/stats` | JWT | Global stats |
| GET | `/overview` | JWT | Admin overview |
| GET | `/data` | JWT | Dashboard data |
| GET | `/admin/dashboard/data` | JWT | Admin dashboard data |

### 5.4 Support (`/api/v1/support/*`)
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/dashboard/trends` | Support | Dashboard trends |
| GET | `/dashboard/top-errors` | Support | Top errors |
| GET | `/dashboard/summary` | Support | Dashboard summary |
| GET | `/dashboard/recent-activity` | Support | Recent activity |
| GET | `/slas` | Support | SLA metrics |
| GET | `/tickets` | Support | Support tickets |
| GET | `/incidents` | Support | Incidents |
| GET | `/incidents/metrics` | Support | Incident metrics |
| GET | `/incidents/trends` | Support | Incident trends |
| GET | `/incidents/feed` | Support | Incident feed |
| GET | `/incidents/detail` | Support | Incident detail |
| POST | `/incidents/comment` | Support | Add comment |
| PATCH | `/incidents/update` | Support | Update incident |
| GET | `/users` | Support | Search users |
| GET | `/users/{id}/profile` | Support | User profile |
| GET | `/users/{id}/security` | Support | User security |
| GET | `/users/{id}/sessions` | Support | User sessions |
| POST | `/users/{id}/sessions/{sessionId}/revoke` | Support | Revoke session |
| GET | `/users/{id}/logins` | Support | Login history |
| GET | `/users/{id}/roles` | Support | User roles |
| GET | `/users/{id}/permissions` | Support | User permissions |
| GET | `/users/{id}/notes` | Support | User notes |
| POST | `/users/{id}/notes` | Support | Add note |
| POST | `/users/{id}/password-reset` | Support | Reset password |
| POST | `/users/{id}/unlock` | Support | Unlock account |
| POST | `/users/{id}/mfa-reset` | Support | Reset MFA |
| POST | `/users/{id}/logout` | Support | Force logout |

### 5.5 Other Domains (Registered)
- `/api/v1/drivers/*` — Driver CRUD
- `/api/v1/vehicles/*` — Vehicle CRUD
- `/api/v1/bulk-jobs/*` — Bulk upload jobs
- `/api/v1/routes/*` — Route optimization
- `/api/v1/reports/*` — Report management
- `/api/v1/diagnostics/*` — Device diagnostics
- `/api/v1/india/*` — India compliance (FastTag, e-way bill, Vahan, Sarathi, GST, permits)
- `/api/v1/settings/*` — User settings (with user context middleware)
- `/api/v1/governance/*` — Governance/DSAR
- `/api/v1/admin-support/*` — Admin support (DR, organizations, emergency, tickets)
- `/api/v1/monitoring/health` — System health
- `/api/v1/monitoring/sla` — SLA metrics
- `/api/v1/ml/*` — ML service proxy
- `/api/v1/ui/*` — UI config, menus, legal docs
- `/api/v1/fuel/*` — Fuel management
- `/api/v1/work-orders/*` — Maintenance work orders
- `/api/v1/inventory/*` — Inventory/parts
- `/api/v1/vendors/*` — Vendor management
- `/api/v1/inspections/*` — Vehicle inspections
- `/api/v1/finance/*` — Finance/expense management
- `/api/v1/notifications/*` — Notification management
- `/api/v1/integrations/*` — Integration management
- `/api/v1/trips/*` — Trip management
- `/api/v1/emergency/*` — Emergency management
- `/api/v1/video/*` — Video telematics
- `/api/v1/analytics/*` — Analytics management
- `/api/v1/geofences/*` — Geofence management

---

## 14. What's Actually Working vs What's Stub

### ✅ Fully Working (Tested & Verified)
- **Auth**: Login, refresh, profile, MFA, RBAC, JWKS, session management
- **Billing**: Invoice CRUD, payments, PDF, settings, metrics (5 sample invoices)
- **Dashboard**: Stats, overview, data endpoints
- **Support**: Incident management, user search, session management
- **Governance**: DSAR requests, exports, compliance reports
- **Security**: Auth stats, threat alerts, IP controls, MFA center
- **RBAC**: Role management, user assignments, security logs
- **UI Config**: Global config, menus, legal docs

### ⚠️ Partially Working (Endpoints exist, UI not fully wired)
- **Vehicles**: CRUD endpoints exist, UI pages exist but may have gaps
- **Drivers**: CRUD endpoints exist
- **Trips**: CRUD + telemetry endpoints exist
- **Fuel**: Endpoints + some UI pages
- **Maintenance**: Work order endpoints + UI pages
- **Reports**: Report management endpoints
- **Analytics**: Jobs, quotas, access logs
- **Notifications**: Settings + preference management
- **Integrations**: Hub + webhook management
- **Organizations**: CRUD + settings
- **Devices**: Diagnostics + firmware management
- **Emergency**: Incident management
- **DR**: Backup/restore operations
- **Finance**: Expense management
- **Geofences**: CRUD + events
- **Route Optimization**: Job management
- **India Compliance**: FastTag, toll, GST tables exist

### ❌ Not Working (Stub or Not Wired)
- **Live Vehicle Map**: No real GPS data ingestion, map UI not wired
- **Dispatch Flow**: No working dispatch workflow
- **Payment Gateway**: No Razorpay/Stripe integration
- **Driver Mobile App**: React Native app exists but not connected
- **Landing Page**: Next.js app exists but not deployed
- **Video Telematics**: Tables exist, no streaming
- **ML Predictions**: Tables exist, no model serving
- **Revenue Forecasting**: UI page exists, no data
- **Real GPS Integration**: No Traccar/Teltonika integration

---

## 15. Archived Services (35 — Consolidated into Monolith)

These were separate microservices that have been consolidated into `fleet-core`:

| Service | Domain | Tables | Status |
|---------|--------|--------|--------|
| `auth-service-legacy` | Auth | Migrated | ✅ In monolith |
| `billing-service-new` | Billing | Migrated | ✅ In monolith |
| `driver-service` / `driver-service-new` | Driver | Migrated | ✅ In monolith |
| `vehicle-service` / `vehicle-service-new` | Vehicle | Migrated | ✅ In monolith |
| `tracking-service-new` | Tracking | Migrated | ✅ In monolith |
| `trip-service` / `trip-service-new` | Trip | Migrated | ✅ In monolith |
| `analytics-service` / `analytics-service-new` | Analytics | Migrated | ⚠️ Partial |
| `alert-service-new` | Alert | Migrated | ✅ In monolith |
| `notification-service` | Notifications | Migrated | ⚠️ Partial |
| `fuel-service` | Fuel | Migrated | ⚠️ Partial |
| `maintenance-service` / `maintenance-service-new` | Maintenance | Migrated | ⚠️ Partial |
| `geofencing-service` | Geofencing | Migrated | ⚠️ Partial |
| `route-optimization-service` | Route | Migrated | ⚠️ Partial |
| `reports-service-new` | Reports | Migrated | ⚠️ Partial |
| `settings-service` / `settings-service-new` | Settings | Migrated | ⚠️ Partial |
| `integrations-service` / `integrations-service-new` | Integrations | Migrated | ⚠️ Partial |
| `governance-service` / `governance-service-new` | Governance | Migrated | ✅ In monolith |
| `india-integration-service-new` | India | Migrated | ⚠️ Stub |
| `ingestion-service-new` | Telemetry Ingestion | Migrated | ⚠️ Partial |
| `optimization-service-new` | Optimization | Migrated | ⚠️ Partial |
| `support-service` | Support | Migrated | ✅ In monolith |
| `dr-service` | Disaster Recovery | Migrated | ⚠️ Partial |
| `admin-support-service-new` | Admin Support | Migrated | ⚠️ Partial |
| `device-diagnostics-service` | Diagnostics | Migrated | ⚠️ Partial |
| `fleet-ml-service-new` | ML | Migrated | ❌ Stub |
| `video-service` | Video | Migrated | ❌ Stub |
| `worker-service-new` | Background Workers | Migrated | ⚠️ Partial |

---

## 16. Technology Stack

| Layer | Technology |
|-------|-----------|
| **Backend** | Go 1.26, Chi router, Zap logging, pgx (Postgres driver) |
| **Frontend** | React 18, TypeScript, Vite 8, Tailwind CSS, Redux Toolkit, React Query |
| **Database** | PostgreSQL 15+ with Row-Level Security, PostGIS, table partitioning |
| **Cache** | Redis 7 (sessions, rate limits, pub/sub, dead letter queue) |
| **Message Bus** | NATS (event streaming, telemetry ingestion) |
| **Auth** | JWKS + RSA-256 JWT signing, bcrypt passwords, TOTP MFA |
| **Maps** | Leaflet / MapLibre (planned, not wired) |
| **Payments** | Razorpay (planned, not integrated) |
| **Email** | SendGrid (planned) |
| **SMS** | MSG91 (planned, India) |
| **Infrastructure** | Docker, Docker Compose, AWS/GCP (planned) |
| **Monitoring** | Custom health endpoints, SLA tracking, synthetic checks |

---

## 17. Multi-Tenancy Model

- **Tenant isolation**: Row-Level Security (RLS) on all org-scoped tables
- **Org context**: `org_id` in JWT claims → extracted by JWKS middleware → enforced by `TenantContextMiddleware`
- **Cross-org**: Super admins can view all orgs
- **User-org relationship**: Many-to-many via `organization_memberships` and `user_organizations`
- **Org switching**: Via `/auth/refresh` with `X-Org-ID` header

---

## 18. Security Model

| Layer | Implementation |
|-------|---------------|
| **Authentication** | JWKS Auth Gate → RSA-256 JWT validation → 15-min access tokens, 7-day refresh tokens |
| **Authorization** | RBAC with 6 roles: `super_admin`, `org_admin`, `admin`, `fleet_manager`, `dispatcher`, `driver` |
| **Mutation protection** | `RequireRoleForMutations` middleware on all POST/PUT/PATCH/DELETE |
| **Tenant isolation** | RLS policies + `TenantContextMiddleware` |
| **Rate limiting** | Redis-based per-endpoint limits |
| **MFA** | TOTP with recovery codes, policy overrides per org |
| **Session management** | Single-use refresh tokens with reuse detection |
| **Audit trail** | All mutations logged to `security_events` and `audit_logs` |
| **PII protection** | Data masking in governance exports, encryption at rest |
| **IP controls** | IP rules, geo-blocking, allow/deny lists |

---

## 19. Current Gaps (What Needs to Be Built)

### 11.1 Critical (Blockers for Beta)
1. **Live Vehicle Map** — No real GPS data ingestion, no map UI wired
2. **Dispatch Workflow** — No working assign → track → complete flow
3. **IAM Routes** — User management endpoints partially missing (just fixed `/admin/users`)
4. **Token Refresh** — Fixed the reuse detection bug, but needs monitoring

### 11.2 Important (Required for Launch)
5. **Payment Gateway** — Razorpay integration for invoice payments
6. **Notification System** — Email/SMS alerts for geofence, delays, due invoices
7. **GPS Telematics** — Traccar or Teltonika device integration
8. **India Compliance** — Real FastTag, e-way bill, Vahan API connections

### 11.3 Nice to Have (Phase 2)
9. **Driver Mobile App** — React Native app for drivers
10. **AI Route Optimization** — ML-based route suggestions
11. **Predictive Maintenance** — ML model for vehicle maintenance prediction
12. **Video Telematics** — Dashcam video streaming and analytics
13. **Revenue Forecasting** — ML-based revenue predictions

---

## 20. Scale & Capacity

| Metric | Current | Target |
|--------|---------|--------|
| Database tables | 392 | 400+ |
| Backend domains | 27 | 30+ |
| Frontend pages | 180+ | 200+ |
| API endpoints | 100+ | 200+ |
| DB connection pools | 4 (Critical/High/Medium/Low) | 4 |
| Max pool sizes | 25/20/10/10 | 50/40/20/20 |
| Telemetry partitioning | Monthly (2026-03 through 2027-01) | Monthly ongoing |

---

## 21. Open Questions

1. **GPS device strategy**: Traccar (open source) vs Teltonika (commercial) vs both?
2. **Driver mobile app**: Build in parallel or after web dispatch is stable?
3. **Multi-language**: Hindi + English for beta, or English only?
4. **Payment gateway**: Razorpay only, or add Stripe for international customers?
5. **Map provider**: Leaflet (free) vs Mapbox (paid, better UX)?
6. **SMS provider**: MSG91 (India) vs Twilio (global)?
7. **Deployment**: Self-hosted for enterprise customers, or SaaS only?
8. **Data retention**: How long to keep telemetry data? (Currently partitioned through 2027-01)

---

## 22. Engineering Quality Metrics

### 11.1 Codebase Size
| Metric | Count |
|--------|-------|
| Go code | 124,401 lines |
| Frontend code (TS/TSX) | 70,974 lines |
| Route registrations | 163 in main.go |
| Handler files | 61 |
| Service files | 147 |
| Repository files | 75 |
| Test functions | 509 across 123 files |
| DB indexes | 1,370 |
| CI/CD workflows | 28 GitHub Actions |
| Docker Compose configs | 3 (dev, prod, full) |

### 11.2 Performance Benchmarks (Live)
| Endpoint | Status | Response Time | Assessment |
|----------|--------|---------------|------------|
| `POST /auth/login` | 200 | 531ms | ⚠️ Slow (bcrypt + DB) |
| `GET /auth/profile` | 200 | 93ms | ✅ Good |
| `GET /billing/health` | 200 | 8ms | ✅ Excellent |
| `GET /dashboard/data` | 200 | 170ms | ✅ Good |
| `GET /billing/invoices` | 200 | 36ms | ✅ Good |

### 11.3 Database State
| Table | Live Rows | Assessment |
|-------|-----------|------------|
| `refresh_tokens` | 439 | Normal (accumulated from testing) |
| `invoice_items` | 15 | Sample data from invoice feature |
| `users` | 11 | Dev seed data |
| `invoices` | 5 | Sample invoices |
| `organizations` | 1 | Single dev org |
| **All other tables** | 0 | Empty — no production data |

**Partitioned tables**: 16 partitions for `telemetry_events` (monthly) and `telemetry_points` (monthly through 2027).

---

## 23. Security Posture

### 12.1 Audit Results
| Check | Result | Severity | Status |
|-------|--------|----------|--------|
| Hardcoded secrets | ✅ None found | — | Pass |
| SQL injection (string concat) | ⚠️ 42 instances of `fmt.Sprintf` in queries | Medium | Needs audit |
| CORS | ✅ Configurable via env var | — | Pass |
| Rate limiting | ⚠️ Only 4/41 endpoint groups (10%) | Medium | Q2 priority |
| TLS/HTTPS | ❌ No enforcement (HTTP only) | High | Q1 priority |
| JWKS auth | ✅ RSA-256, on-demand key refresh | — | Pass |
| RLS | ✅ Row-Level Security on org-scoped tables | — | Pass |
| Token reuse fix | ✅ Fixed (atomic revoke + IP check) | — | Pass |

### 12.2 Production Blockers
1. **No TLS** — All traffic is HTTP. Need HTTPS termination at reverse proxy.
2. **Rate limiting incomplete** — Only login, register, governance, analytics have limits.
3. **42 SQL concat queries** — Need audit: most likely safe but some could be injection vectors.
4. **No API versioning** — All endpoints are `/api/v1/`. No v2 strategy.

---

## 24. Deployment Readiness

### 13.1 Current State
| Aspect | Status | Notes |
|--------|--------|-------|
| `.env` | Dev only | Localhost URLs, dev JWT secret |
| Docker Compose | 3 configs | `docker-compose.yml`, `.prod.yml`, `.full.yml` |
| CI/CD | 28 workflows | Tests, security scanning, deploy |
| Kubernetes | ❌ None | No manifests or Helm charts |
| Secrets management | ❌ None | `.env` file only |
| Monitoring | ⚠️ Partial | Health endpoints exist, no aggregation |
| Backups | ⚠️ Untested | DR tables exist, no restore test |

### 13.2 Pre-Launch Checklist
- [ ] TLS/HTTPS enforcement (Q1)
- [ ] Rate limiting on all endpoints (Q2)
- [ ] SQL injection audit complete (Q1)
- [ ] API versioning strategy defined (Q4)
- [ ] Production secrets management (Vault/AWS Secrets Manager)
- [ ] DR test with restore verification
- [ ] Load testing at 100 concurrent users
- [ ] 80%+ test coverage report
- [ ] OpenAPI/Swagger spec generated
- [ ] Unified health endpoint

---

## 25. Appendix

### 14.1 File Locations
- **Backend**: `/data/data/com.termux/files/home/flyfleet/services/fleet-core/`
- **Frontend**: `/data/data/com.termux/files/home/flyfleet/apps/dashboard-web/`
- **Database**: PostgreSQL on `localhost:5433`, database `fleet_db`
- **Redis**: `localhost:6379`
- **NATS**: `nats://localhost:4222`

### 14.2 Running Services
| Service | Port | Status |
|---------|------|--------|
| Backend (fleet-core) | 8084 | ✅ Running |
| Frontend (Vite) | 3000 | ✅ Running |
| PostgreSQL | 5433 | ✅ Running |
| Redis | 6379 | ✅ Running |
| NATS | 4222 | ✅ Running |

### 14.3 Git Status
- **Branch**: `feature/invoice-generation`
- **Commits**: 10 on feature branch
- **Remote**: `origin/feature/invoice-generation` — pushed
- **Files changed**: 11 files, +1534/-125 lines

---

## 26. Shared Libraries (Packages)

### 26.1 `packages/common/go` — Shared Go Library
| Metric | Count |
|--------|-------|
| Source files | 50 |
| Test files | 2 |
| Test coverage | 4% (critically low) |

**Purpose**: Cross-service shared code — logging adapters, metrics, middleware, platform middleware, error types, context keys, HTTP utilities, observability helpers.

**Key packages**:
- `logadapter/` — Zap logger adapter for consistent logging
- `middleware/` — Shared HTTP middleware (CORS, tracing, auth helpers)
- `platformmw/` — Platform-level middleware (observability, health)
- `metrics/` — Prometheus metrics registration
- `errors/` — Shared error types
- `ctxkeys/` — Context key constants

**Status**: ✅ Live — imported by all domains

### 26.2 `packages/config/` — Branding & Whitelabel
| File | Purpose |
|------|---------|
| `branding/default.json` | Default brand colors, logos, typography, app name |
| `src/index.ts` | TypeScript config loader |
| `package.json` | NPM package definition |

**Status**: ⚠️ Minimal — only default branding, no multi-brand support yet

---

## 27. Proto/gRPC Contracts

### 27.1 Proto Files (13)
| Module | Proto File | Purpose |
|--------|-----------|---------|
| `admin/` | Admin service contracts | Organization, user management |
| `alert/` | Alert service contracts | Alert rules, notifications |
| `analytics/` | Analytics service contracts | Reports, predictions |
| `auth/` | Auth service contracts | Login, tokens, MFA, sessions |
| `billing/` | Billing service contracts | Invoices, payments, subscriptions |
| `compliance/` | Compliance contracts | Regulatory compliance |
| `driver/` | Driver service contracts | Driver CRUD, shifts, ratings |
| `fleet/` | Fleet service contracts | Vehicle management |
| `maintenance/` | Maintenance contracts | Work orders, inspections |
| `reports/` | Report contracts | Report generation |
| `telemetry/` | Telemetry contracts | GPS, OBD, events |
| `tracking/` | Tracking contracts | Real-time tracking |
| `trip/` | Trip contracts | Trip management |

**Status**: ✅ Live — 13 proto files defining service contracts across all domains

### 27.2 Generated Code
- Go protobuf generated code in each domain's `proto/` subdirectory
- `go.mod` + `go.sum` in `proto/` for Go module management

---

## 28. Firmware & Device Layer

### 28.1 Firmware Files (5)
| File | Purpose |
|------|---------|
| `firmware/keys/` | Firmware signing keys |
| `firmware/releases/` | Firmware release binaries |

**Status**: ⚠️ Minimal — signing keys and release binaries exist, no OTA update pipeline wired

### 28.2 Device Backend (Diagnostics Domain)
Already covered in Section 16 — 15 tables, 28 backend files, 20 tests, 12 frontend pages, 15 routes.

---

## 29. Infrastructure & DevOps

### 29.1 Docker Compose (3 Configs — 1,006 lines total)
| File | Lines | Purpose |
|------|-------|---------|
| `docker-compose.yml` | 196 | Development — Postgres, Redis, NATS |
| `docker-compose.prod.yml` | 387 | Production — adds reverse proxy, monitoring |
| `docker-compose.full.yml` | 423 | Full stack — adds Grafana, Prometheus, Loki, ClickHouse, Vault, OTel |

**Status**: ✅ Live — 3 deployment profiles from dev to full observability

### 29.2 Kubernetes/Helm (29 files)
| Directory | Files | Purpose |
|-----------|-------|---------|
| `infrastructure/helm/` | 10 | Helm charts for Kubernetes deployment |
| `infrastructure/k8s/` | 19 | Raw Kubernetes manifests (deployments, services, ingress, configmaps) |

**Status**: ⚠️ Partial — manifests exist but not tested against production cluster

### 29.3 Terraform (31 files)
| Directory | Files | Purpose |
|-----------|-------|---------|
| `infrastructure/terraform/` | 31 | IaC for AWS/GCP — VPC, RDS, ElastiCache, ECS, EKS |

**Status**: ⚠️ Partial — Terraform configs exist, not applied to live infrastructure

### 29.4 Monitoring Stack (29 files)
| Component | Files | Purpose |
|-----------|-------|---------|
| `infrastructure/prometheus/` | 1 | Prometheus config |
| `infrastructure/grafana/` | 3 | Grafana dashboards + provisioning |
| `infrastructure/loki/` | 1 | Loki log aggregation config |
| `infrastructure/clickhouse/` | 1 | ClickHouse analytics DB config |
| `infrastructure/monitoring/` | 24 | Custom monitoring configs, alerts, rules |
| `infrastructure/otel-collector-config.yaml` | 1 | OpenTelemetry collector config |

**Status**: ⚠️ Partial — configs exist, not running in production

### 29.5 Secrets & Vault (7 files)
| Component | Files | Purpose |
|-----------|-------|---------|
| `infrastructure/vault/` | 3 | HashiCorp Vault config for secrets management |
| `infrastructure/secrets/` | 4 | Encrypted secret templates |
| `infrastructure/certs/` | 1 | TLS certificate management |

**Status**: ❌ Stub — configs exist, no live Vault deployment

### 29.6 Incident Response (33 files)
| Directory | Files | Purpose |
|-----------|-------|---------|
| `infrastructure/incident-response/` | 33 | Runbooks, escalation policies, communication templates, post-mortem templates |

**Status**: ✅ Live — comprehensive incident response documentation

### 29.7 Backup & Recovery (3 files)
| Directory | Files | Purpose |
|-----------|-------|---------|
| `infrastructure/backup/` | 3 | Backup scripts, retention policies |

**Status**: ⚠️ Partial — scripts exist, not scheduled/automated

### 29.8 Database Migrations (92 files)
| Directory | Files | Purpose |
|-----------|-------|---------|
| `infrastructure/migrations/` | 92 | SQL migration files for schema evolution |

**Status**: ✅ Live — 92 migrations tracked and versioned

### 29.9 Logging (5 files)
| Component | Files | Purpose |
|-----------|-------|---------|
| `infrastructure/logging/` | 4 | Log rotation, structured logging config |
| `infrastructure/ci-cd/` | 3 | CI/CD pipeline configs |

**Status**: ⚠️ Partial — configs exist, not fully wired

### 29.10 Self-Hosted (4 files)
| Directory | Files | Purpose |
|-----------|-------|---------|
| `infrastructure/self-hosted/` | 4 | Self-hosted deployment guide and scripts |

**Status**: ⚠️ Partial — guide exists, not tested end-to-end

---

## 30. CI/CD Pipelines

### 30.1 GitHub Actions Workflows (26 files)
| Workflow | Purpose |
|----------|---------|
| `active-dashboard-ci.yml` | Dashboard web app CI |
| `all-services-foundation.yml` | Foundation checks for all services |
| `api-contract-clients.yml` | API contract client generation |
| `auth-tests.yml` | Auth service tests |
| `bff-gateway-check.yml` | BFF gateway health check |
| `billing_testing.yml` | Billing service tests |
| `contract-tests.yml` | API contract tests |
| `deploy.yml` | Production deployment pipeline |
| `design-system-foundation.yml` | Design system checks |
| `device-reliability-foundation.yml` | Device reliability tests |
| `diagnostics_tests.yml` | Diagnostics service tests |
| `docs-governance.yml` | Documentation governance |
| `enterprise-foundation.yml` | Enterprise feature checks |
| `firmware-ota-foundation.yml` | Firmware OTA checks |
| `identity-auth-foundation.yml` | Identity/auth checks |
| `mobile-sync-foundation.yml` | Mobile sync checks |
| `observability-foundation.yml` | Observability checks |
| `operations-governance.yml` | Operations governance |
| `ota-simulation.yml` | OTA rollout simulation |
| `performance-gates.yml` | Performance regression gates |
| `runtime-integration.yml` | Runtime integration tests |
| `security-scanning.yml` | Semgrep + dependency scanning |
| `security.yml` | Security audit pipeline |
| `service-integration-matrix.yml` | Service integration tests |
| `trace-propagation.yml` | Distributed trace propagation |
| `vehicle-device-foundation.yml` | Vehicle/device foundation checks |

**Status**: ✅ Live — 26 workflows covering tests, security, deployment, governance

### 30.2 Makefile (420 lines)
**Key targets**: `bootstrap`, `hot-reload`, `full-stack-up`, `full-stack-health`, `full-stack-logs`, `test`, `lint`, `migrate`, `backup`, `restore`, `deploy`

**Status**: ✅ Live — comprehensive build/deploy automation

---

## 31. API Contracts (OpenAPI)

### 31.1 OpenAPI Specifications
| File | Lines | Size | Purpose |
|------|-------|------|---------|
| `openapi.yaml` | 2,286 | 56KB | Human-readable API spec |
| `openapi.json` | 11,149 | 263KB | Machine-readable API spec |
| `openapi.combined.json` | 11,149 | 263KB | Combined spec (all services) |
| `openapi.combined.fixed.json` | 11,149 | 263KB | Fixed combined spec |
| `openapi_settings.yaml` | — | 9.6KB | Settings API spec |

**Status**: ✅ Live — comprehensive API documentation, but not auto-generated from code

---

## 32. Operational Scripts (122 files)

### 32.1 Script Categories
| Category | Count | Examples |
|----------|-------|----------|
| **Database** | 15 | `fix_schema.sql`, `fix_null_columns.sql`, `seed_organizations.sql`, `mock_users.sql` |
| **Deployment** | 10 | `bootstrap.sh`, `deploy_remaining_services.sh`, `start_local.sh`, `start_full_stack.sh` |
| **Testing** | 12 | `billing_test.sh`, `test_driver_service.sh`, `test_tracking_service.sh`, `stress_test.js` |
| **Security** | 8 | `iam_production_readiness.sh`, `check_coverage.sh`, `smoke_bff.sh` |
| **Verification** | 10 | `verify_database_fixes.sh`, `verify_full_flow.sh`, `check_implementation_status.sh` |
| **Backup/Restore** | 4 | `backup.sh`, `restore.sh`, `bulk_migrate.sh`, `bulk_undo.sh` |
| **Diagnostics** | 8 | `pprof.sh`, `detect-n1-queries.py`, `deep_backend_api_scan.sh` |
| **Setup** | 15 | `setup-mcp.sh`, `create_admin.sh`, `setup_admin.sql` |
| **Auth** | 12 | Scripts in `scripts/auth/` for user management |
| **API** | 5 | Scripts in `scripts/api/` for API testing |
| **Security Scanning** | 5 | Scripts in `scripts/security/` |
| **DB Migration** | 8 | Scripts in `scripts/db/` |
| **Diagnostics** | 6 | Scripts in `scripts/diagnostics/` |
| **Setup** | 6 | Scripts in `scripts/setup/` |

**Status**: ✅ Live — comprehensive operational tooling

---

## 33. Documentation Ecosystem (52 files across 15 directories)

### 33.1 Docs by Category
| Directory | Files | Content |
|-----------|-------|---------|
| `docs/api/` | 3 | API documentation |
| `docs/architecture/` | 6 | Architecture diagrams, decisions |
| `docs/backend/` | 3 | Backend documentation |
| `docs/design/` | 4 | Design system, UI specs |
| `docs/device/` | 5 | Device protocol docs |
| `docs/disaster-recovery/` | 1 | DR runbooks |
| `docs/firmware/` | 2 | Firmware documentation |
| `docs/legacy-docs/` | 13 | Legacy documentation |
| `docs/operations/` | 1 | Operations guides |
| `docs/pilot/` | 1 | Pilot program docs |
| `docs/product/` | 8 | Product specs, role definitions |
| `docs/quality/` | 1 | Quality standards |
| `docs/release/` | 1 | Release process |
| `docs/runbooks/` | 2 | Operational runbooks |
| `docs/security/` | 1 | Security documentation |

**Root-level docs**: 80+ additional `.md` files at repo root (architecture assessments, security audits, completion certificates, etc.)

**Status**: ✅ Live — extensive documentation, but scattered and needs consolidation

---

## 34. Tests & Quality

### 34.1 Test Infrastructure
| Directory | Files | Purpose |
|-----------|-------|---------|
| `tests/e2e/` | — | End-to-end tests |
| `tests/infrastructure/` | — | Infrastructure tests |
| `tests/k6/` | — | k6 load tests |
| `tests/load/` | — | Load testing scripts |
| `tests/run-tests.sh` | 1 | Test runner |

### 34.2 Backend Test Coverage (by Domain)
| Domain | Source Files | Test Files | Ratio |
|--------|-------------|------------|-------|
| Auth | 13 | 2 | 15% |
| Vehicle | 35 | 4 | 11% |
| Driver | 31 | 3 | 10% |
| Trip | 12 | 4 | 33% |
| Tracking | 29 | 5 | 17% |
| Billing | 106 | 15 | 14% |
| Fuel | 18 | 3 | 17% |
| Maintenance | 35 | 4 | 11% |
| Geofencing | 9 | 2 | 22% |
| Alert | 13 | 4 | 31% |
| Analytics | 35 | 2 | 6% |
| Governance | 18 | 8 | 44% |
| India | 31 | 1 | 3% |
| Emergency | 22 | 10 | 45% |
| Video | 23 | 0 | 0% |
| Diagnostics | 28 | 20 | 71% |
| Notifications | 14 | 5 | 36% |
| Integrations | 13 | 0 | 0% |
| Support | 20 | 1 | 5% |
| Finance | 34 | 6 | 18% |
| Route Opt | 10 | 0 | 0% |
| Settings | 13 | 6 | 46% |
| Reports | 11 | 0 | 0% |
| Monitoring | 14 | 0 | 0% |
| DR | 6 | 1 | 17% |
| Customer | 26 | 0 | 0% |
| Platform Compat | 7 | 0 | 0% |
| Compliance | 35 | 3 | 9% |
| **packages/common/go** | **50** | **2** | **4%** |

**Total**: ~650 source files, ~110 test files = **~17% overall test coverage by file count**

**Status**: ⚠️ Partial — 8 domains have zero tests, overall coverage is low

### 34.3 Security Scanning
| Tool | Config | Purpose |
|------|--------|---------|
| Semgrep | `.semgrep.yaml` (20 lines) | Static analysis |
| Pre-commit | `.pre-commit-config.yaml` (149 lines) | Git hook checks |
| SECURITY.md | 140 lines | Security policy |

**Status**: ✅ Live — security scanning configured

---

## 35. Monorepo Configuration

### 35.1 Build System
| File | Purpose |
|------|---------|
| `Makefile` (420 lines) | Build, test, deploy automation |
| `go.work` | Go workspace for multi-module development |
| `go.mod` (root) | Root Go module |
| `package.json` | NPM workspace config |
| `turbo.json` | Turborepo monorepo orchestration |
| `.npmrc` | NPM config (offline mode, workspace settings) |
| `bun.lock` | Bun lockfile |
| `package-lock.json` | NPM lockfile |

### 35.2 Git & CI Config
| File | Purpose |
|------|---------|
| `.gitignore` | Git ignore rules |
| `.dockerignore` | Docker ignore rules |
| `.prettierrc.json` | Code formatting config |
| `.changeset/` | Changeset versioning |

### 35.3 Environment Configuration
| File | Purpose |
|------|---------|
| `.env` | Local dev environment variables |
| `.env.example` | Environment template |
| `config.toml` | Application config |

### 35.4 Backups & Uploads
| Directory | Purpose |
|-----------|---------|
| `backups/` | Database backup storage |
| `uploads/` | User file uploads |
| `invoices/` | Generated invoice PDFs |
| `logs/` | Application logs |
| `tmp/` | Temporary files |

---

## 36. Complete System Summary

### 36.1 By the Numbers

| Layer | Metric | Count |
|-------|--------|-------|
| **Database** | Tables | 392 |
| **Database** | Indexes | 1,370 |
| **Database** | Partitioned tables | 16 |
| **Database** | Migrations | 92 |
| **Backend** | Domains | 27 |
| **Backend** | Go source files | ~650 |
| **Backend** | Test files | ~110 |
| **Backend** | Archived services | 35 |
| **Frontend** | Pages/components | ~300 |
| **Frontend** | TypeScript/TSX lines | 70,974 |
| **API** | Route prefixes | 41 |
| **API** | OpenAPI spec lines | 13,435 |
| **Proto** | Proto files | 13 |
| **Infra** | Docker Compose files | 3 (1,006 lines) |
| **Infra** | Kubernetes manifests | 29 |
| **Infra** | Terraform files | 31 |
| **Infra** | Monitoring configs | 29 |
| **Infra** | Incident response docs | 33 |
| **CI/CD** | GitHub Actions workflows | 26 |
| **CI/CD** | Makefile targets | 420 lines |
| **Scripts** | Operational scripts | 122 |
| **Docs** | Documentation files | 130+ |
| **Security** | Scanning configs | 3 |
| **Packages** | Shared Go library | 50 files |
| **Packages** | Branding config | 3 files |
| **Firmware** | Firmware files | 5 |
| **Total** | Go code lines | 124,401 |
| **Total** | Frontend code lines | 70,974 |
| **Total** | Infrastructure configs | ~200 files |

### 36.2 Production Readiness Scorecard

| Area | Score | Notes |
|------|-------|-------|
| **Auth & Security** | 8/10 | JWKS, RBAC, RLS, MFA ✅ — No TLS ❌ |
| **API Coverage** | 7/10 | 41 prefixes, OpenAPI spec ✅ — No versioning ❌ |
| **Database** | 9/10 | 392 tables, RLS, partitioning ✅ — No connection pooling metrics ❌ |
| **Frontend** | 6/10 | 180+ pages ✅ — Map not wired ❌ |
| **Testing** | 3/10 | 17% coverage, 8 domains zero tests ❌ |
| **CI/CD** | 8/10 | 26 workflows, Makefile ✅ — No production deploy tested ❌ |
| **Infrastructure** | 5/10 | Docker, K8s, Terraform exist ✅ — Not applied to prod ❌ |
| **Monitoring** | 4/10 | Prometheus, Grafana, Loki configs ✅ — Not running ❌ |
| **Documentation** | 7/10 | 130+ docs ✅ — Scattered, needs consolidation ❌ |
| **Operations** | 6/10 | 122 scripts ✅ — Not automated/scheduled ❌ |
| **Overall** | **6.3/10** | Strong foundation, needs production hardening |

### 36.3 Production Blockers (Must Fix Before Launch)

| # | Blocker | Severity | Quarter |
|---|---------|----------|---------|
| 1 | No TLS/HTTPS enforcement | Critical | Q1 |
| 2 | Rate limiting on only 10% of endpoints | High | Q2 |
| 3 | 42 SQL string concat queries need audit | High | Q1 |
| 4 | 8 domains with zero test coverage | High | Q3-Q4 |
| 5 | No production secrets management | High | Q2 |
| 6 | No load testing at scale | Medium | Q3 |
| 7 | Kubernetes manifests untested | Medium | Q4 |
| 8 | Monitoring stack not running in prod | Medium | Q4 |

---

*This document is based on a complete audit of the entire FlyFleet monorepo as of April 3, 2026.*
*392 database tables, 27 backend domains, 180+ frontend pages, 35 archived services, 13 proto files, 26 CI/CD workflows, 122 operational scripts, 130+ documentation files, 3 Docker Compose configs, 29 K8s manifests, 31 Terraform files.*
*Next review: After each major feature launch or infrastructure change.*
