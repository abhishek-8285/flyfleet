# FlyFleet — Technical Specification Diff

**Version**: 1.0  
**Date**: April 3, 2026  
**Purpose**: Gap analysis between current state and production-ready system  
**Scope**: Every domain, endpoint, table, and frontend page  

---

## How to Read This Document

| Symbol | Meaning |
|--------|---------|
| ✅ | Implemented, tested, working |
| ⚠️ | Partially implemented — endpoints exist but UI not wired or data is stub |
| ❌ | Not implemented — tables may exist but no code, or nothing at all |
| 🔧 | Needs fix — exists but has bugs or security issues |

---

## 1. Auth & Identity

### 1.1 Authentication

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| Login (email/password) | ✅ | Working, returns JWT + refresh token | — | — |
| Refresh token rotation | 🔧 | Fixed atomic revoke + IP check | — | Was nuking sessions on concurrent refresh |
| Logout | ✅ | Cookie-based + token revoke | — | — |
| Registration | ✅ | Working with rate limiting | — | — |
| MFA (TOTP) | ✅ | Setup, verify, recovery codes | — | — |
| Password reset | ✅ | Request + confirm flow | — | — |
| Dev auto-login | ✅ | Development only | Remove before prod | Security risk |
| OIDC login | ⚠️ | Endpoint exists, no providers configured | Google, Microsoft SSO | Need OIDC provider config |
| Session management | ✅ | View sessions, revoke | — | — |
| Invite system | ⚠️ | Validate endpoint exists | Full invite → register flow | Missing invite creation + email |
| Account lockout | ✅ | After failed attempts | — | — |
| Login history | ⚠️ | `login_events` table exists | UI to view history | No frontend page |

### 1.2 Authorization

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| JWKS Auth Gate | ✅ | RSA-256, on-demand key refresh | — | — |
| RBAC (6 roles) | ✅ | super_admin, org_admin, admin, fleet_manager, dispatcher, driver | — | — |
| RequireRole middleware | ✅ | On protected routes | — | — |
| RequireRoleForMutations | ✅ | On POST/PUT/PATCH/DELETE | — | — |
| TenantContextMiddleware | ✅ | org_id from JWT → context | — | — |
| Permission checks | ⚠️ | `CheckPermission` endpoint exists | UI integration | No frontend permission gates |
| Role matrix UI | ⚠️ | `/admin/rbac/*` pages exist | Fully wired | Partially connected |

### 1.3 User Management

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| User CRUD | ⚠️ | GET `/users/{id}` exists | Full CRUD API + UI | POST/PUT/DELETE missing |
| PUT `/admin/users/{id}` | ✅ | Added for preference sync | — | Was missing, caused 403 redirect loop |
| User profile | ✅ | `GET /auth/profile` | — | Was missing JWKS auth gate — fixed |
| User search | ⚠️ | Support endpoint exists | Admin search UI | Not wired to admin UI |
| Bulk user import | ❌ | Nothing | CSV upload | New feature |
| User deactivation | ❌ | Nothing | Soft delete + audit | New feature |

### 1.4 Organization Management

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| List organizations | ✅ | `GET /organizations` | — | — |
| Get organization | ✅ | `GET /organizations/{id}` | — | — |
| Create organization | ❌ | Nothing | API + UI | New feature |
| Update organization | ❌ | Nothing | API + UI | New feature |
| Org switching | ✅ | Via `/auth/refresh` + `X-Org-ID` | — | — |
| Multi-org membership | ⚠️ | `organization_memberships` table exists | UI to manage | No frontend |
| Org settings | ⚠️ | `/admin/organizations/:id/settings` page exists | Wired to API | Partially connected |

---

## 2. Vehicle Management

### 2.1 Vehicle CRUD

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| List vehicles | ✅ | `GET /vehicles` | — | — |
| Get vehicle detail | ✅ | `GET /vehicles/{id}` | — | — |
| Create vehicle | ✅ | `POST /vehicles` | — | — |
| Update vehicle | ✅ | `PUT /vehicles/{id}` | — | — |
| Delete vehicle | ✅ | `DELETE /vehicles/{id}` | — | — |
| Bulk upload | ⚠️ | `/admin/bulk-uploads/*` pages exist | Full flow with validation | UI exists, needs wiring |
| Vehicle specs | ⚠️ | `vehicle_specs` table exists | CRUD API | No endpoints |
| Vehicle assignments | ⚠️ | `vehicle_assignments` table exists | Assign driver ↔ vehicle | No endpoints |
| Vehicle compliance | ⚠️ | `vehicle_compliance` table exists | Track permits, insurance | No endpoints |
| Vehicle financials | ⚠️ | `vehicle_financials` table exists | Cost tracking | No endpoints |
| Vehicle inspections | ⚠️ | `vehicle_inspections` table exists + `/inspections/*` pages | Full flow | Partially wired |
| Vehicle downtime | ⚠️ | `vehicle_downtime` table exists | Track + report | No endpoints |
| Vehicle lifecycle metrics | ⚠️ | `vehicle_lifecycle_metrics` table exists | Dashboard widget | No endpoints |
| Vehicle verifications | ⚠️ | `vehicle_verifications` table exists | KYC flow | No endpoints |
| Vehicle events | ⚠️ | `vehicle_events` table exists | Event timeline | No endpoints |
| Vehicle daily stats | ⚠️ | `vehicle_daily_stats` table exists | Aggregation job | No endpoints |
| Vehicle telemetry | ⚠️ | `vehicle_telemetry` table exists | Real-time stream | No endpoints |

### 2.2 Vehicle UI

| Page | Status | Route | Notes |
|------|--------|-------|-------|
| Vehicle list | ✅ | `/vehicles`, `/admin/vehicles` | Working |
| Vehicle detail | ✅ | `/vehicles/:id`, `/admin/vehicles/:id` | Working |
| Add vehicle | ✅ | `/vehicles/new`, `/admin/vehicles/new` | Working |
| Edit vehicle | ✅ | `/vehicles/:id/edit` | Working |
| Bulk uploads | ⚠️ | `/admin/bulk-uploads/*` | Pages exist, partial wiring |

---

## 3. Driver Management

### 3.1 Driver CRUD

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| List drivers | ✅ | `GET /drivers` | — | — |
| Get driver | ✅ | `GET /driver/{id}` | — | — |
| Create driver | ✅ | `POST /drivers` | — | — |
| Update driver | ✅ | `PUT /drivers/{id}` | — | — |
| Delete driver | ✅ | `DELETE /drivers/{id}` | — | — |
| Driver profiles | ⚠️ | `driver_profiles` table exists | Full CRUD | No dedicated endpoints |
| Driver onboarding | ⚠️ | `driver_onboarding` table exists | Workflow | No endpoints |
| Driver shifts | ⚠️ | `driver_shifts` table exists | Schedule management | No endpoints |
| Driver training | ⚠️ | `driver_training` table exists | Certification tracking | No endpoints |
| Driver payroll | ⚠️ | `driver_payroll` table exists | Pay calculation | No endpoints |
| Driver ratings | ⚠️ | `driver_ratings` table exists | Rating system | No endpoints |
| Driver safety | ⚠️ | `driver_safety` table exists | Safety score | No endpoints |
| Driver mobile sync | ⚠️ | `driver_mobile_sync` table exists | Sync API | No endpoints |

### 3.2 Driver UI

| Page | Status | Route | Notes |
|------|--------|-------|-------|
| Driver list | ✅ | `/drivers`, `/admin/drivers` | Working |
| Driver detail | ⚠️ | Partial | Needs full detail page |

---

## 4. Trip & Dispatch

### 4.1 Trip Management

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| List trips | ✅ | `GET /trips` | — | — |
| Create trip | ✅ | `POST /trips` | — | — |
| Update trip | ✅ | `PUT /trips/{id}` | — | — |
| Get trip detail | ✅ | `GET /trips/{id}` | — | — |
| Delete trip | ✅ | `DELETE /trips/{id}` | — | — |
| Trip settings | ⚠️ | `trip_settings` table exists | Config API | No endpoints |
| Trip profitability | ⚠️ | `trip_profitability` table exists | Calculation engine | No endpoints |
| Trip telemetry | ⚠️ | `/trips/:id/telemetry` page exists | Real-time tracking | Partially wired |

### 4.2 Dispatch Flow

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| Dispatch center | ⚠️ | `/dispatch` page exists | Full workflow | Not wired to API |
| Assign vehicle + driver | ❌ | Nothing | API endpoint | New feature |
| Track delivery status | ❌ | Nothing | pending → in-transit → delivered | New state machine |
| Delivery receipt | ❌ | Nothing | Auto-generate with timestamp + location | New feature |
| Route optimization | ⚠️ | `route_optimization_*` tables exist + `/route-optimization` page | Working optimizer | Partially wired |

### 4.3 Trip UI

| Page | Status | Route | Notes |
|------|--------|-------|-------|
| Trip list | ✅ | `/trips`, `/admin/trips` | Working |
| Trip detail | ✅ | `/trips/:id` | Working |
| Add trip | ✅ | `/trips/add` | Working |
| Edit trip | ✅ | `/trips/:id/edit` | Working |
| Trip telemetry | ⚠️ | `/trips/:id/telemetry` | Page exists, partial |
| Trip settings | ⚠️ | `/trips/settings` | Page exists, partial |
| Dispatch center | ⚠️ | `/dispatch` | Page exists, not wired |
| Route optimization | ⚠️ | `/route-optimization` | Page exists, partial |

---

## 5. Tracking & Telemetry

### 5.1 GPS Ingestion

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| GPS location API | ⚠️ | `gps_locations` table exists | Ingestion endpoint | No POST endpoint |
| OBD readings | ⚠️ | `obd_readings` table exists | Ingestion endpoint | No POST endpoint |
| Telemetry events | ⚠️ | Partitioned table (16 partitions) | Ingestion via NATS | No ingestion pipeline |
| Telemetry points | ⚠️ | Partitioned table (12 monthly partitions) | Real-time stream | No WebSocket feed |
| Telemetry analytics | ⚠️ | `telemetry_analytics` table exists | Aggregation jobs | No endpoints |
| Telemetry anomalies | ⚠️ | `telemetry_anomalies` table exists | Detection engine | No endpoints |
| Telemetry outbox | ⚠️ | `telemetry_outbox` table exists | Event publishing | No publisher |
| Vehicle telemetry | ⚠️ | `vehicle_telemetry` table exists | Per-vehicle stream | No endpoints |
| Fleet analytics | ⚠️ | `fleet_analytics` table exists | Dashboard data | No endpoints |
| Fleet hourly stats | ⚠️ | `fleet_hourly_stats` table exists | Aggregation job | No endpoints |
| Materialized reports | ⚠️ | `materialized_fleet_reports` table exists | Scheduled refresh | No endpoints |
| Tracking config | ⚠️ | `tracking_config` table exists | Config API | No endpoints |

### 5.2 Live Map

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| Map UI | ❌ | `/map` route not registered | Leaflet/MapLibre with live markers | New feature |
| WebSocket feed | ❌ | `/ws` route exists but not wired | Real-time position updates | New feature |
| Vehicle markers | ❌ | Nothing | Clickable markers with info | New feature |
| Geofence overlay | ❌ | Nothing | Draw geofences on map | New feature |
| History trail | ❌ | Nothing | 24-hour playback | New feature |

---

## 6. Billing & Invoicing

### 6.1 Invoice Management

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| List invoices | ✅ | `GET /billing/invoices` | — | 5 sample invoices |
| Create invoice | ✅ | `POST /billing/invoices` | — | — |
| Get invoice | ✅ | `GET /billing/invoices/{id}` | — | — |
| Update invoice | ✅ | `PUT /billing/invoices/{id}` | — | — |
| Delete invoice | ✅ | `DELETE /billing/invoices/{id}` | — | — |
| Record payment | ✅ | `POST /billing/invoices/{id}/payments` | — | — |
| Download PDF | ✅ | `GET /billing/invoices/{id}/pdf` | — | — |
| Email invoice | ❌ | Nothing | Send via email with PDF | New feature |
| Auto-create from delivery | ❌ | Nothing | Trigger on trip completion | New feature |
| Invoice approval workflow | ⚠️ | `invoice_approvals`, `invoice_approval_policies`, `invoice_approval_matrix` tables exist | Multi-level approval | No endpoints |
| Invoice audit log | ⚠️ | `invoice_audit_log` table exists | Audit trail | No endpoints |
| Invoice permissions | ⚠️ | `invoice_permissions` table exists | Access control | No endpoints |
| Invoice sequence | ⚠️ | `invoice_sequence` table exists | Auto-numbering | No endpoints |
| User invoice access | ⚠️ | `user_invoice_access` table exists | Customer portal access | No endpoints |

### 6.2 Billing Settings & Metrics

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| Billing settings | ✅ | `GET/PUT /billing/settings` | — | — |
| Billing metrics | ✅ | `GET /billing/metrics` | — | — |
| Billing health | ✅ | `GET /billing/health` | — | — |
| Billing configurations | ⚠️ | `billing_configurations` table exists | Config API | No endpoints |
| Billing plans | ⚠️ | `billing_plans` table exists | Plan management | No endpoints |
| Billing subscriptions | ⚠️ | `billing_subscriptions` table exists | Subscription management | No endpoints |
| Billing pricing components | ⚠️ | `billing_pricing_components` table exists | Pricing engine | No endpoints |
| Billing tax rates | ⚠️ | `billing_tax_rates` table exists | Tax calculation | No endpoints |
| Billing coupons | ⚠️ | `billing_coupons` table exists | Coupon management | No endpoints |
| Billing credit transactions | ⚠️ | `billing_credit_transactions` table exists | Credit system | No endpoints |
| Billing customer balances | ⚠️ | `billing_customer_balances` table exists | Balance tracking | No endpoints |
| Billing hardware products | ⚠️ | `billing_hardware_products` table exists | Hardware billing | No endpoints |
| Billing invoice documents | ⚠️ | `billing_invoice_documents` table exists | Document storage | No endpoints |
| Billing invoice templates | ⚠️ | `billing_invoice_templates` table exists | Template management | No endpoints |
| Billing invoice sequences | ⚠️ | `billing_invoice_sequences` table exists | Sequence management | No endpoints |
| Billing accounting exports | ⚠️ | `billing_accounting_exports` table exists | Export to accounting | No endpoints |
| Billing recognized revenue | ⚠️ | `billing_recognized_revenue` table exists | Revenue recognition | No endpoints |
| Billing revenue schedules | ⚠️ | `billing_revenue_schedules` table exists | Revenue scheduling | No endpoints |

### 6.3 Payment Gateway

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| Razorpay integration | ❌ | Nothing | Payment links in invoices | New feature |
| Payment records | ⚠️ | `payment_records` table exists | Payment tracking | No endpoints |
| Bank transactions | ⚠️ | `bank_transactions` table exists | Reconciliation | No endpoints |
| Payment reconciliation | ❌ | Nothing | Auto-match payments to invoices | New feature |

### 6.4 Billing UI

| Page | Status | Route | Notes |
|------|--------|-------|-------|
| Billing dashboard | ✅ | `/admin/billing` | Working, real metrics |
| Invoice list | ✅ | Via billing dashboard | Working |
| Create invoice | ✅ | `/admin/billing/new` | Working |
| Invoice detail | ✅ | `/admin/billing/invoices/:id` | Working |
| Edit invoice | ✅ | `/admin/billing/invoices/:id/edit` | Working |
| Billing settings | ✅ | `/admin/billing/settings` | Working |
| Invoice designer | ⚠️ | `/admin/billing/designer` | Page exists, partial |
| Customer portal invoices | ⚠️ | `/portal/invoices` | Page exists, partial |

---

## 7. Fuel Management

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| Fuel logs | ⚠️ | `fuel_logs` table exists + `/fuel/*` pages | Full CRUD | Partially wired |
| Fuel transactions | ⚠️ | `fuel_transactions` table exists | Transaction tracking | No endpoints |
| Fuel readings | ⚠️ | `fuel_readings` table exists | Sensor data | No endpoints |
| Fuel vendors | ⚠️ | `fuel_vendors` table exists | Vendor management | No endpoints |
| Fuel settings | ⚠️ | `fuel_settings` table exists + page | Config API | Partially wired |
| Fuel audit logs | ⚠️ | `fuel_audit_logs` table exists | Audit trail | No endpoints |
| Fuel card reconciliation | ⚠️ | `fuel_card_reconciliation` table exists | Reconciliation flow | No endpoints |
| Fuel efficiency history | ⚠️ | `fuel_efficiency_history` table exists | Efficiency tracking | No endpoints |
| Org fuel settings | ⚠️ | `org_fuel_settings` table exists | Per-org config | No endpoints |
| Receipt OCR | ⚠️ | `receipt_ocr_jobs` table exists | OCR processing | No endpoints |
| Fuel dashboard | ⚠️ | `/finance/fuel` page exists | Dashboard | Partially wired |
| Add fuel log | ⚠️ | `/fuel/add` page exists | Form | Partially wired |
| Fuel policy | ⚠️ | `/finance/policy` page exists | Policy management | Partially wired |

---

## 8. Maintenance & Work Orders

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| Work orders | ⚠️ | `work_orders` table exists + `/maintenance/*` pages | Full CRUD | Partially wired |
| Work order parts | ⚠️ | `work_order_parts` table exists | Parts tracking | No endpoints |
| Maintenance schedule | ⚠️ | `maintenance_schedule` table exists | Scheduling | No endpoints |
| Maintenance windows | ⚠️ | `maintenance_windows` table exists | Window management | No endpoints |
| Maintenance technicians | ⚠️ | `maintenance_technicians` table exists | Technician management | No endpoints |
| Maintenance vendors | ⚠️ | `maintenance_vendors` table exists | Vendor management | No endpoints |
| Maintenance settings | ⚠️ | `maintenance_settings` table exists | Config API | No endpoints |
| Technician skills | ⚠️ | `technician_skills` table exists | Skills matrix | No endpoints |
| Inventory parts | ⚠️ | `inventory_parts` table exists + `/inventory` page | Inventory management | Partially wired |
| Inspection templates | ⚠️ | `inspection_templates` table exists | Template management | No endpoints |
| Inspection documents | ⚠️ | `inspection_documents` table exists | Document storage | No endpoints |
| Vendor payments | ⚠️ | `vendor_payments` table exists | Payment tracking | No endpoints |
| Purchase orders | ⚠️ | `purchase_orders` table exists | PO management | No endpoints |
| Purchase order items | ⚠️ | `purchase_order_items` table exists | Line items | No endpoints |
| Work order list | ⚠️ | `/maintenance/work-orders` page | List view | Partially wired |
| Work order detail | ⚠️ | `/maintenance/work-orders/:id` page | Detail view | Partially wired |
| Add work order | ⚠️ | `/maintenance/work-orders/new` page | Create form | Partially wired |
| Vendor directory | ⚠️ | `/vendors` page | List view | Partially wired |
| Vendor form | ⚠️ | `/vendors/new` page | Create form | Partially wired |
| Inspection form | ⚠️ | `/inspections/new` page | Create form | Partially wired |

---

## 9. Geofencing

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| Geofences | ⚠️ | `geofences` table exists + `/geofences` page | Full CRUD | Partially wired |
| Geofence assignments | ⚠️ | `geofence_assignments` table exists | Assign to vehicles | No endpoints |
| Geofence events | ⚠️ | `geofence_events` table exists | Enter/exit events | No endpoints |
| Geofence alerts | ❌ | Nothing | Trigger notifications on enter/exit | New feature |

---

## 10. Alerts & Monitoring

### 10.1 Alert Management

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| Alert rules | ⚠️ | `alert_rules` table exists | Rule management | No endpoints |
| Alert groups | ⚠️ | `alert_groups` table exists | Grouping | No endpoints |
| Alert thresholds | ⚠️ | `alert_thresholds` table exists | Threshold config | No endpoints |
| Alerts | ⚠️ | `alerts` table exists + `/alerts` page | Alert list | Partially wired |
| Alerts v2 templates | ⚠️ | `alerts_v2_templates` table exists | Template system | No endpoints |
| Alerts v2 logs | ⚠️ | `alerts_v2_logs` table exists | Alert history | No endpoints |
| Alerts v2 DLQ | ⚠️ | `alerts_v2_dlq` table exists | Dead letter queue | No endpoints |
| Alerts v2 user preferences | ⚠️ | `alerts_v2_user_preferences` table exists | Per-user config | No endpoints |
| Threat alerts | ⚠️ | `threat_alerts` table exists + `/admin/security/threat-alerts` page | Threat dashboard | Partially wired |
| Health alerts | ⚠️ | `health_alerts` table exists | System health alerts | No endpoints |

### 10.2 System Monitoring

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| System health | ✅ | `GET /monitoring/health` | — | — |
| SLA metrics | ✅ | `GET /monitoring/sla` | — | — |
| Health history | ⚠️ | `health_history` table exists | Historical health | No endpoints |
| Synthetic checks | ⚠️ | `synthetic_checks` table exists | Synthetic monitoring | No endpoints |
| Error budgets | ⚠️ | `error_budgets` table exists | Budget tracking | No endpoints |
| Service SLOs | ⚠️ | `service_slos` table exists | SLO management | No endpoints |
| Metrics hub | ⚠️ | `/admin/monitoring/metrics` page | Metrics dashboard | Partially wired |
| Log viewer | ⚠️ | `/admin/monitoring/logs` page | Log viewer | Partially wired |
| Queue monitor | ⚠️ | `/admin/monitoring/queue` page | Queue monitoring | Partially wired |
| SLA report | ⚠️ | `/admin/monitoring/sla` page | SLA dashboard | Partially wired |

---

## 11. Notifications

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| Notification channels | ⚠️ | `notification_channels` table exists | Channel config (email, SMS, push) | No endpoints |
| Notification events | ⚠️ | `notification_events` table exists | Event tracking | No endpoints |
| Notification logs | ⚠️ | `notification_logs` table exists | Delivery logs | No endpoints |
| Notification settings | ⚠️ | `notification_settings` table exists + pages | Config API | Partially wired |
| Notification preference audit | ⚠️ | `notification_preference_audit` table exists | Audit trail | No endpoints |
| User notification preferences | ⚠️ | `user_notification_preferences` table exists | Per-user config | No endpoints |
| Global notification defaults | ⚠️ | `global_notification_defaults` table exists | Default config | No endpoints |
| Email sending | ❌ | Nothing | SendGrid integration | New feature |
| SMS sending | ❌ | Nothing | MSG91 integration | New feature |
| Push notifications | ❌ | Nothing | FCM/APNs integration | New feature |
| Notification settings UI | ⚠️ | `/admin/notifications/*` (5 pages) | Settings management | Partially wired |
| User preferences UI | ⚠️ | `/preferences`, `/preferences/:id`, `/preferences/global` | Preference management | Partially wired |

---

## 12. Analytics & Reports

### 12.1 Analytics

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| Analytics jobs | ⚠️ | `analytics_jobs` table exists + `/analytics/jobs` page | Job management | Partially wired |
| Analytics quotas | ⚠️ | `analytics_quotas` table exists | Quota management | No endpoints |
| Analytics access logs | ⚠️ | `analytics_access_logs` table exists | Access tracking | No endpoints |
| Analytics query logs | ⚠️ | `analytics_query_logs` table exists | Query history | No endpoints |
| Analytics query plans | ⚠️ | `analytics_query_plans` table exists | Query optimization | No endpoints |
| Prediction results | ⚠️ | `prediction_results` table exists | ML predictions | No endpoints |
| Predictive models | ⚠️ | `predictive_models` table exists + `/analytics/models` page | Model management | Partially wired |
| Model deployments | ⚠️ | `model_deployments` table exists | Deployment tracking | No endpoints |
| Model predictions | ⚠️ | `model_predictions` table exists | Prediction storage | No endpoints |
| Historical demand | ⚠️ | `historical_demand` table exists | Demand data | No endpoints |
| Predictive maintenance analytics | ⚠️ | `predictive_maintenance_analytics` table exists + `/analytics/maintenance` page | Maintenance predictions | Partially wired |
| Fleet analytics page | ⚠️ | `/analytics/fleet` page | Fleet dashboard | Partially wired |
| Driver behavior page | ⚠️ | `/analytics/drivers` page | Driver analytics | Partially wired |

### 12.2 Reports

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| Report definitions | ⚠️ | `reports` endpoints exist | Report management | Partially wired |
| Report schedules | ⚠️ | `/reports` page exists | Scheduled reports | Partially wired |
| Report execution history | ⚠️ | `/reports/history` page exists | Execution tracking | Partially wired |
| Create report | ⚠️ | `/reports/create` page exists | Report builder | Partially wired |
| Admin reports | ⚠️ | `/admin/reports` redirect exists | Admin report management | Partially wired |

### 12.3 Revenue

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| Revenue forecasting | ❌ | `/revenue/forecasting` page exists but no data | ML-based predictions | New feature |
| Pricing rules | ⚠️ | `pricing_rules` table exists | Pricing engine | No endpoints |

---

## 13. India Compliance

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| FastTag transactions | ⚠️ | `fastag_transactions` table exists | Reconciliation | No endpoints |
| Toll transactions | ⚠️ | `toll_transactions` table exists | Toll tracking | No endpoints |
| GST compliance | ⚠️ | `gst_compliances` table exists | GST filing | No endpoints |
| GST invoices | ⚠️ | `gst_invoices` table exists | GST invoicing | No endpoints |
| Local permits | ⚠️ | `local_permits` table exists | Permit management | No endpoints |
| India integration logs | ⚠️ | `india_integration_logs` table exists | API call logs | No endpoints |
| TDS records | ⚠️ | `tds_records` table exists | TDS tracking | No endpoints |
| FastTag API integration | ❌ | Nothing | Real FastTag data | New feature |
| E-way bill API | ❌ | Nothing | Real e-way bill generation | New feature |
| Vahan API | ❌ | Nothing | Vehicle registration lookup | New feature |
| Sarathi API | ❌ | Nothing | Driver license verification | New feature |

---

## 14. Governance & Audit

### 14.1 Governance

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| DSAR requests | ✅ | `GET/POST /governance` | — | Working |
| Governance settings | ⚠️ | `governance_settings` table exists | Config API | No endpoints |
| Governance exports | ⚠️ | `governance_exports` table exists | Export management | No endpoints |
| Governance export lineage | ⚠️ | `governance_export_lineage` table exists | Lineage tracking | No endpoints |
| Governance download tokens | ⚠️ | `governance_download_tokens` table exists | Secure downloads | No endpoints |
| Governance audit logs | ⚠️ | `governance_audit_logs` table exists | Audit trail | No endpoints |
| Governance compliance reports | ⚠️ | `governance_compliance_reports` table exists | Report generation | No endpoints |
| Governance settings UI | ⚠️ | `/governance/settings` page | Settings management | Partially wired |
| Governance request list | ⚠️ | `/governance` page | Request list | Partially wired |
| Governance request detail | ⚠️ | `/governance/requests/:id` page | Request detail | Partially wired |
| New governance request | ⚠️ | `/governance/requests/new` page | Create form | Partially wired |

### 14.2 Audit

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| Audit logs | ⚠️ | `audit_logs` table exists | Audit trail API | No endpoints |
| Audit policies | ⚠️ | `audit_policies` table exists | Policy management | No endpoints |
| Audit settings | ⚠️ | `audit_settings` table exists | Config API | No endpoints |
| Entity audit history | ⚠️ | `entity_audit_history` table exists | Entity history API | No endpoints |
| Evidence exports | ⚠️ | `evidence_exports` table exists | Export management | No endpoints |
| Legal holds | ⚠️ | `legal_holds` table exists | Legal hold management | No endpoints |
| Audit center UI | ⚠️ | `/audit/*` (5 pages) | Audit management | Partially wired |

### 14.3 Security Audit

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| Security events | ⚠️ | `security_events` table exists | Event API | No endpoints |
| Security history | ⚠️ | `security_history` table exists | History API | No endpoints |
| Auth audit logs | ⚠️ | `auth_audit_logs` table exists | Auth audit API | No endpoints |
| System audit logs | ⚠️ | `system_audit_logs` table exists | System audit API | No endpoints |
| System flags | ⚠️ | `system_flags` table exists | Feature flags API | No endpoints |
| System limits | ⚠️ | `system_limits` table exists | Rate limit config | No endpoints |
| Security logs UI | ⚠️ | `/admin/rbac/logs` page | Security log viewer | Partially wired |

---

## 15. Emergency & Safety

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| Emergency incidents | ⚠️ | `emergency_incidents` table exists + `/emergency` page | Incident management | Partially wired |
| Emergency incidents archive | ⚠️ | `emergency_incidents_archive` table exists | Archive management | No endpoints |
| Emergency incident snapshots | ⚠️ | `emergency_incident_snapshots` table exists | Snapshot storage | No endpoints |
| Emergency commands | ⚠️ | `emergency_commands` table exists | Command API | No endpoints |
| Emergency contacts | ⚠️ | `emergency_contacts` table exists | Contact management | No endpoints |
| Emergency contact responses | ⚠️ | `emergency_contact_responses` table exists | Response tracking | No endpoints |
| Emergency escalations | ⚠️ | `emergency_escalations` table exists | Escalation management | No endpoints |
| Emergency response units | ⚠️ | `emergency_response_units` table exists | Unit management | No endpoints |
| Emergency status | ⚠️ | `emergency_status` table exists | Status tracking | No endpoints |
| Emergency vehicle state | ⚠️ | `emergency_vehicle_state` table exists | Vehicle state tracking | No endpoints |
| SOS incidents | ⚠️ | `sos_incidents` table exists | SOS management | No endpoints |
| Emergency dashboard UI | ⚠️ | `/admin/emergency` page | Emergency dashboard | Partially wired |

---

## 16. Video Telematics

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| Video streams | ⚠️ | `video_streams` table exists | Stream management | No endpoints |
| Video recordings | ⚠️ | `video_recordings` table exists | Recording storage | No endpoints |
| Video analytics | ⚠️ | `video_analytics` table exists | Analytics API | No endpoints |
| Video retention policies | ⚠️ | `video_retention_policies` table exists | Retention config | No endpoints |
| Video telematics UI | ❌ | `/admin/video` route exists | Video dashboard | Stub page |

---

## 17. Device Management & Diagnostics

### 17.1 Devices

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| Device list | ⚠️ | `GET /devices` | Device management | Partially wired |
| Device status | ⚠️ | `GET /devices/status` | Status API | Partially wired |
| Device summary | ⚠️ | `GET /devices/summary` | Summary API | Partially wired |
| Device twins | ⚠️ | `GET /devices/twins` | Device twin API | Partially wired |
| Device dashboard UI | ⚠️ | `/devices` page | Device dashboard | Partially wired |
| Device telemetry UI | ⚠️ | `/devices/:id/telemetry` page | Telemetry view | Partially wired |
| Device diagnostics UI | ⚠️ | `/devices/:id/diagnostics` page | Diagnostics view | Partially wired |
| Device twins UI | ⚠️ | `/devices/:id/twins` page | Twin management | Partially wired |

### 17.2 Firmware

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| Firmware packages | ⚠️ | `firmware_packages` table exists | Package management | No endpoints |
| Firmware rollouts | ⚠️ | `firmware_rollouts` table exists + `/diagnostics/firmware` | Rollout management | Partially wired |
| Firmware rollout events | ⚠️ | `firmware_rollout_events` table exists | Event tracking | No endpoints |
| Rollout device status | ⚠️ | `rollout_device_status` table exists | Per-device status | No endpoints |
| Remote commands | ⚠️ | `/devices/:id/commands` page | Command management | Partially wired |

### 17.3 Diagnostics

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| Device diagnostics | ⚠️ | `/diagnostics/devices` | Diagnostics API | Partially wired |
| Device status | ⚠️ | `/diagnostics/status` | Status API | Partially wired |
| Diagnostics summary | ⚠️ | `/diagnostics/summary` | Summary API | Partially wired |

---

## 18. Integrations & Webhooks

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| Integrations | ⚠️ | `integrations` table exists + `/integrations` page | Integration management | Partially wired |
| Integration configs | ⚠️ | `integration_configs` table exists | Config API | No endpoints |
| Integration logs | ⚠️ | `integration_logs` table exists | Log API | No endpoints |
| Webhooks | ⚠️ | `webhooks` table exists | Webhook management | No endpoints |
| Webhook subscriptions | ⚠️ | `webhook_subscriptions` table exists | Subscription management | No endpoints |
| Webhook event types | ⚠️ | `webhook_event_types` table exists | Event type registry | No endpoints |
| ERP configs | ⚠️ | `erp_configs` table exists | ERP integration config | No endpoints |
| Integration hub UI | ⚠️ | `/integrations` page | Integration hub | Partially wired |
| API client hub UI | ⚠️ | `ApiClientHub` component | API key management | Partially wired |
| Webhook manager UI | ⚠️ | `WebhookManager` component | Webhook management | Partially wired |
| Integration logs UI | ⚠️ | `IntegrationLogs` component | Log viewer | Partially wired |
| Marketplace UI | ⚠️ | `Marketplace` component | Integration marketplace | Partially wired |

---

## 19. Support & Incidents

### 19.1 Incident Management

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| Incidents | ⚠️ | `incidents` table exists + `/support/incidents` | Incident management | Partially wired |
| Incident comments | ⚠️ | `incident_comments` table exists | Comment API | No endpoints |
| Incident logs | ⚠️ | `incident_logs` table exists | Log API | No endpoints |
| Incident media | ⚠️ | `incident_media` table exists | Media storage | No endpoints |
| Incident SLA violations | ⚠️ | `incident_sla_violations` table exists | SLA tracking | No endpoints |
| Incident timeline | ⚠️ | `incident_timeline` table exists | Timeline API | No endpoints |
| Incident timeline events | ⚠️ | `incident_timeline_events` table exists | Event API | No endpoints |
| Incident updates | ⚠️ | `incident_updates` table exists | Update API | No endpoints |
| Incident vehicle snapshot | ⚠️ | `incident_vehicle_snapshot` table exists | Snapshot storage | No endpoints |
| Incident detail UI | ⚠️ | `/support/incidents/:id` page | Incident detail | Partially wired |

### 19.2 Support Tickets

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| Support tickets | ⚠️ | `support_tickets` table exists | Ticket management | No endpoints |
| Support ticket messages | ⚠️ | `support_ticket_messages` table exists | Message API | No endpoints |
| Support ticket attachments | ⚠️ | `support_ticket_attachments` table exists | Attachment API | No endpoints |
| Ticket events | ⚠️ | `ticket_events` table exists | Event tracking | No endpoints |
| SLA metrics | ⚠️ | `sla_metrics` table exists + `/support/sla` page | SLA dashboard | Partially wired |
| Support audit log | ⚠️ | `support_audit_log` table exists | Audit trail | No endpoints |
| Org support metrics | ⚠️ | `org_support_metrics` table exists | Metrics API | No endpoints |
| Support dashboard UI | ⚠️ | `/support` page | Support dashboard | Partially wired |
| Org console UI | ⚠️ | `/support/org-console/:id` page | Org console | Partially wired |
| User console UI | ⚠️ | `/support/user-console` page | User console | Partially wired |
| Troubleshooting logs UI | ⚠️ | `/support/troubleshooting` page | Log viewer | Partially wired |

---

## 20. Finance

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| Expense items | ⚠️ | `expense_items` table exists | Expense management | No endpoints |
| Expense reports | ⚠️ | `expense_reports` table exists | Report generation | No endpoints |
| Expense policies | ⚠️ | `expense_policies` table exists | Policy management | No endpoints |
| Finance audit logs | ⚠️ | `finance_audit_logs` table exists | Audit trail | No endpoints |
| Finance retention policies | ⚠️ | `finance_retention_policies` table exists | Retention config | No endpoints |
| Expense approvals UI | ⚠️ | `/finance/expenses` page | Approval workflow | Partially wired |

---

## 21. Platform & Settings

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| Platform settings | ⚠️ | `platform_settings` table exists | Config API | No endpoints |
| Feature flags | ⚠️ | `feature_flags` table exists + `/admin/settings/flags` page | Flag management | Partially wired |
| Settings audit logs | ⚠️ | `settings_audit_logs` table exists | Audit trail | No endpoints |
| Localization settings | ⚠️ | `localization_settings` table exists + `/admin/settings/localization` page | Localization config | Partially wired |
| Templates | ⚠️ | `templates` table exists + `/admin/settings/templates` page | Template management | Partially wired |
| AB test experiments | ⚠️ | `ab_test_experiments` table exists | Experiment management | No endpoints |
| Usage meters | ⚠️ | `usage_meters` table exists | Usage tracking | No endpoints |
| Usage records | ⚠️ | `usage_records` table exists | Usage storage | No endpoints |
| Org budgets | ⚠️ | `org_budgets` table exists | Budget management | No endpoints |
| ID mappings | ⚠️ | `id_mappings` table exists | ID translation | No endpoints |
| Sync jobs | ⚠️ | `sync_jobs` table exists | Job management | No endpoints |
| Schema migrations | ⚠️ | `schema_migrations` table exists | Migration tracking | — |
| Schema migrations legacy | ⚠️ | `schema_migrations_legacy` table exists | Legacy migration tracking | — |
| Storage pools | ⚠️ | `storage_pools` table exists | Storage management | No endpoints |
| Platform config UI | ⚠️ | `/admin/settings/platform` page | Platform settings | Partially wired |
| Policies/limits UI | ⚠️ | `/admin/settings/policies` page | Policy management | Partially wired |

---

## 22. Disaster Recovery

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| Backups | ⚠️ | `backups` table exists + `/admin/dr` pages | Backup management | Partially wired |
| Backup verifications | ⚠️ | `backup_verifications` table exists | Verification API | No endpoints |
| Restore jobs | ⚠️ | `restore_jobs` table exists | Restore management | No endpoints |
| Admin search sync state | ⚠️ | `admin_search_sync_state` table exists | Sync tracking | No endpoints |
| Admin operation approvals | ⚠️ | `admin_operation_approvals` table exists | Approval workflow | No endpoints |
| Admin audit logs | ⚠️ | `admin_audit_logs` table exists | Audit trail | No endpoints |
| Backup list UI | ⚠️ | `/admin/dr` page | Backup list | Partially wired |
| Create backup UI | ⚠️ | `/admin/dr/create` page | Backup creation | Partially wired |
| Restore operations UI | ⚠️ | `/admin/dr/restores` page | Restore management | Partially wired |
| DR test list UI | ⚠️ | `/admin/dr/tests` page | Test list | Partially wired |
| DR test detail UI | ⚠️ | `/admin/dr/tests/:id` page | Test detail | Partially wired |
| DR test settings UI | ⚠️ | `/admin/dr/tests/settings` page | Test config | Partially wired |

---

## 23. Customer & Shipments

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| Shipments | ⚠️ | `shipments` table exists + `/portal/shipments` page | Shipment management | Partially wired |
| Proof of delivery | ⚠️ | `proof_of_delivery` table exists | POD management | No endpoints |
| Authority delivery logs | ⚠️ | `authority_delivery_logs` table exists | Delivery tracking | No endpoints |
| Customer portal login | ⚠️ | `/portal/login` page | Customer auth | Partially wired |
| Customer portal dashboard | ⚠️ | `/portal/dashboard` page | Customer dashboard | Partially wired |
| Shipment tracking | ⚠️ | `/portal/track/:id` page | Tracking view | Partially wired |
| New booking | ⚠️ | `/portal/bookings` page | Booking form | Partially wired |
| Portal invoices | ⚠️ | `/portal/invoices` page | Invoice list | Partially wired |

---

## 24. ML & Predictions

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| ML service proxy | ⚠️ | `/api/v1/ml/*` routes exist | ML service integration | Proxy exists, no backend |
| ML predictions | ⚠️ | `model_predictions` table exists | Prediction storage | No endpoints |
| Model deployments | ⚠️ | `model_deployments` table exists | Deployment tracking | No endpoints |
| Predictive models | ⚠️ | `predictive_models` table exists | Model management | No endpoints |
| Demand prediction | ❌ | Nothing | ML-based demand forecasting | New feature |
| Pricing suggestion | ❌ | Nothing | ML-based pricing suggestions | New feature |

---

## 25. Cross-Cutting Concerns

### 25.1 API Infrastructure

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| API versioning | ❌ | Only `/api/v1/` | `/api/v2/` strategy | New feature |
| OpenAPI/Swagger spec | ❌ | Nothing | Auto-generated docs | New feature |
| Request ID tracking | ✅ | TraceID middleware | — | — |
| CORS | ✅ | Configurable via env | — | — |
| Timeout (60s) | ✅ | Chi middleware | — | — |
| Recovery | ✅ | Chi recoverer | — | — |
| Real IP | ✅ | Chi RealIP middleware | — | — |
| Request logging | ✅ | Chi Logger middleware | — | — |
| Observability | ✅ | Platform middleware | — | — |
| HTTP metrics | ✅ | Metrics middleware | — | — |
| Circuit breaker | ✅ | SystemGuard middleware | — | — |
| Priority shedding | ✅ | LowPriorityShedder on reports | Apply to more endpoints | Expand coverage |
| Compatibility redirects | ✅ | 28 legacy → canonical redirects | — | — |

### 25.2 Database

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| Connection pools | ✅ | 4 pools (Critical/High/Medium/Low) | — | 25/20/10/10 max |
| Row-Level Security | ✅ | On org-scoped tables | — | — |
| Table partitioning | ✅ | 16 partitions for telemetry | — | Monthly through 2027 |
| PostGIS | ✅ | `spatial_ref_sys`, `topology`, `layer` | Spatial queries | — |
| Indexes | ✅ | 1,370 indexes | — | — |
| Pool saturation monitoring | ✅ | Debug logging every minute | Alerting | Add alerts |
| RLS BeforeAcquire overhead | ⚠️ | 5-13ms overhead | <2ms | Optimize pool acquisition |

### 25.3 Security

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| JWKS auth | ✅ | RSA-256, on-demand refresh | — | — |
| RBAC | ✅ | 6 roles, permission checks | — | — |
| Rate limiting | ⚠️ | 4/41 endpoint groups (10%) | 100% coverage | Q2 priority |
| TLS/HTTPS | ❌ | HTTP only | HTTPS enforced | Q1 priority |
| SQL parameterization | ⚠️ | 42 `fmt.Sprintf` queries | All parameterized | Audit needed |
| Password hashing | ✅ | bcrypt | — | — |
| MFA | ✅ | TOTP + recovery codes | — | — |
| Token reuse detection | ✅ | Fixed (atomic revoke + IP check) | — | — |
| IP controls | ⚠️ | `ip_rules` table exists + UI page | IP allow/deny lists | Partially wired |
| Impersonation sessions | ⚠️ | `impersonation_sessions` table exists | Admin impersonation | No endpoints |
| OAuth clients | ⚠️ | `oauth_clients` table exists | OAuth app management | No endpoints |
| OIDC providers | ⚠️ | `oidc_providers` table exists | OIDC config | No endpoints |
| User OIDC links | ⚠️ | `user_oidc_links` table exists | Social login linking | No endpoints |
| SSO callback | ⚠️ | `/auth/callback` route exists | SSO flow | Partially wired |
| Auth challenge | ⚠️ | `/auth/challenge` route exists | Challenge flow | Partially wired |
| Auth error page | ⚠️ | `/auth/error` page exists | Error handling | Partially wired |

### 25.4 Frontend Infrastructure

| Component | Status | Current | Target | Diff |
|-----------|--------|---------|--------|------|
| API client | ✅ | Axios with auth interceptor | — | — |
| Token refresh interceptor | ✅ | Queue-based, handles 401s | — | Fixed concurrent refresh |
| RBAC provider | ✅ | `can()`, `hasRole()` checks | — | — |
| Auth provider | ✅ | Login, logout, profile, org switch | — | — |
| Redux store | ✅ | Auth slice, user state | — | — |
| React Query | ✅ | API caching | — | — |
| Vite proxy | ✅ | `/api` → `localhost:8084` | — | — |
| Security headers | ✅ | CSP, X-Frame-Options, etc. | — | — |
| Theme (dark/light) | ✅ | Toggle + sync to API | — | — |
| Language (i18n) | ⚠️ | `flyfleet_language` in localStorage | Full i18n | Partial |
| Service files | ⚠️ | 10 service files, most empty | Fully wired | Only billingService + dashboardApi wired |
| Hardcoded URLs | ✅ | 2 instances | 0 | Minor cleanup |

---

## 26. Summary: What's Actually Done vs What's Not

### ✅ Fully Working (Tested & Verified)
- Auth: login, refresh, profile, MFA, RBAC, JWKS, session management
- Billing: invoice CRUD, payments, PDF, settings, metrics (5 sample invoices)
- Dashboard: stats, overview, data endpoints
- Support: incident management, user search, session management
- Governance: DSAR requests, exports, compliance reports
- Security: auth stats, threat alerts, IP controls, MFA center
- RBAC: role management, user assignments, security logs
- UI Config: global config, menus, legal docs
- Infrastructure: 4 connection pools, RLS, partitioning, 1,370 indexes, 28 CI workflows

### ⚠️ Partially Working (Endpoints or UI exist, not fully wired)
- Vehicles: CRUD endpoints + UI pages exist
- Drivers: CRUD endpoints + UI pages exist
- Trips: CRUD + telemetry endpoints + UI pages exist
- Fuel: endpoints + some UI pages
- Maintenance: work order endpoints + UI pages
- Reports: report management endpoints + UI pages
- Analytics: jobs, quotas, access logs + UI pages
- Notifications: settings + preference management + UI pages
- Integrations: hub + webhook management + UI pages
- Organizations: CRUD + settings + UI pages
- Devices: diagnostics + firmware management + UI pages
- Emergency: incident management + UI pages
- DR: backup/restore operations + UI pages
- Finance: expense management + UI pages
- Geofences: CRUD + events + UI pages
- Route optimization: job management + UI pages
- India compliance: tables exist, no real API connections
- Customer portal: pages exist, not fully wired
- Monitoring: SRE pages exist, partially wired

### ❌ Not Working (Stub or Not Wired)
- **Live Vehicle Map** — No real GPS data ingestion, no map UI wired
- **Dispatch Flow** — No working assign → track → complete workflow
- **Payment Gateway** — No Razorpay/Stripe integration
- **Driver Mobile App** — React Native app exists but not connected
- **Landing Page** — Next.js app exists but not deployed
- **Video Telematics** — Tables exist, no streaming
- **ML Predictions** — Tables exist, no model serving
- **Revenue Forecasting** — UI page exists, no data
- **Real GPS Integration** — No Traccar/Teltonika integration
- **Email/SMS Notifications** — No SendGrid/MSG91 integration
- **TLS/HTTPS** — HTTP only, no enforcement
- **API Versioning** — Only v1, no v2 strategy
- **OpenAPI/Swagger** — No API contract documentation

---

## 27. Priority Fix List (Based on RICE + Engineering Audit)

| Priority | Item | Effort | Impact | Quarter |
|----------|------|--------|--------|---------|
| **P0** | TLS/HTTPS enforcement | S | Critical | Q1 |
| **P0** | SQL injection audit (42 queries) | M | Critical | Q1 |
| **P0** | Unified health endpoint | S | High | Q1 |
| **P0** | Rate limiting coverage (4/41 → 41/41) | M | High | Q2 |
| **P0** | Complete IAM / User Management | M | High | Q2 |
| **P1** | Login performance optimization | S | Medium | Q2 |
| **P1** | Live Vehicle Tracking on Map | XL | Massive | Q3 |
| **P1** | Working Dispatch Flow | XL | High | Q6 |
| **P1** | GPS Telematics Integration | XL | Massive | Q5 |
| **P1** | Notification System | M | Medium | Q4 |
| **P2** | Payment Gateway (Razorpay) | L | High | Q7 |
| **P2** | India Compliance APIs | L | High | Q8 |
| **P2** | Driver Mobile App | XL | High | Q9 |
| **P2** | API Versioning Strategy | S | Medium | Q4 |
| **P2** | Test Coverage Report | S | Medium | Q10 |
| **P2** | OpenAPI/Swagger Spec | M | Medium | Q10 |

---

*This document is a living technical diff. Last updated: April 3, 2026.*
*Based on: 392 DB tables, 27 backend domains, 180+ frontend pages, 124K Go lines, 71K TS lines.*
*Next update: After each major feature launch or domain completion.*
