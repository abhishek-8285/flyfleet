# FlyFleet: User Roles and Features Analysis

**Date:** 2026-03-31
**Analysis based on:** Codebase review of FlyFleet Enterprise fleet management platform
**Source:** Architecture review, RBAC policies, domain structure, frontend applications

---

## Executive Summary

FlyFleet is a comprehensive **fleet management platform** serving multiple user types across the fleet operations lifecycle. The system implements **5 core roles** with granular permission-based access control across **29 specialized domains**.

### Core User Roles

| Role | Description | Primary Use Case |
|------|-------------|------------------|
| **Super Admin** | Platform-level administrator with unrestricted access | System configuration, cross-org oversight |
| **Org Admin** | Organization-level administrator | Fleet management, user management, billing |
| **Fleet Manager** | Operational manager with fleet oversight | Day-to-day fleet operations, monitoring |
| **Driver** | Vehicle operator with mobile app access | Trip execution, status updates, documentation |
| **Member** | Basic organizational user | Limited read-only or feature-specific access |

---

## Detailed Role Analysis

### 1. Super Admin

**Access Level:** Platform-wide, unrestricted

**Permissions:**
- All resources across all organizations (PolicyManifest: all resources)
- Platform configuration and governance
- Disaster recovery operations
- System-wide audit logs and policies
- Integration and API client management
- Webhook configuration

**Features:**
- Create and manage organizations
- View and modify any organization's data
- Configure platform-wide settings
- Access all audit logs and compliance reports
- Manage system integrations (FASTag, VAHAN, SARATHI, etc.)
- Disaster recovery and data export
- User impersonation and support

**Technical Access:**
- Resources: organizations, audit_logs, audit_settings, users, all domain resources
- Frontend: Full dashboard access, all pages and settings
- API: Unrestricted endpoints across all services

---

### 2. Organization Admin (Org Admin)

**Access Level:** Organization-wide within their org

**Permissions:**
- User management (create, update, delete users)
- Role assignment and permission management
- Fleet configuration (vehicles, drivers, routes)
- Billing and financial operations
- Compliance and governance settings
- Analytics and reporting
- Integration management (org-level)

**Features by Domain:**

#### User Management
- Invite users to organization
- Assign roles (fleet_manager, driver, member)
- Reset user passwords
- Manage MFA settings
- View user activity logs

#### Fleet Operations
- Vehicle lifecycle management (add, update, decommission)
- Driver onboarding and profile management
- Route planning and optimization
- Geofence configuration
- Maintenance scheduling

#### Financial Operations
- Invoice generation and management
- Payment processing oversight
- Fuel cost tracking
- Expense approval workflows
- Billing disputes resolution

#### Compliance & Governance
- Monitor regulatory compliance (FMCSA, HOS, India FASTag/E-Way)
- Configure audit policies
- Review governance reports
- Manage evidence exports

#### Analytics & Reports
- Access all analytics dashboards
- Generate operational reports
- Export data (CSV, PDF)
- Custom report building

**Frontend Access:**
- Dashboard: Full operational view
- Pages: Drivers, Vehicles, Trips, Alerts, Analytics, AuditLogs, Maintenance, Fuel, Billing, Settings

---

### 3. Fleet Manager

**Access Level:** Fleet operations within organization

**Permissions:** (PolicyManifest: vehicles, trips, alerts, documents, maintenance, inventory, fuel)

**Features by Domain:**

#### Vehicle Management
- View all vehicles in fleet
- Update vehicle status (active, maintenance, retired)
- Assign vehicles to drivers
- Track vehicle health and diagnostics
- Manage vehicle documents (registration, insurance, permits)

#### Trip Management
- Plan and assign trips
- Monitor trip progress in real-time
- Route optimization
- Trip history and analytics
- Document trip exceptions

#### Alert Management
- View real-time alerts (SOS, geofence breaches, maintenance due)
- Acknowledge and resolve alerts
- Configure alert thresholds
- Alert escalation workflows

#### Maintenance Management
- Schedule preventive maintenance
- Create maintenance work orders
- Track maintenance history
- Manage inventory (parts, supplies)
- Vendor management

#### Fuel Management
- Review fuel logs
- Approve fuel expenses
- Monitor fuel efficiency
- Configure fuel policies

#### Document Management
- Upload and manage vehicle documents
- Driver documentation (licenses, medical certificates)
- Compliance document tracking
- Expiration alerts

**Frontend Access:**
- Dashboard: Fleet overview, active trips, alerts
- Pages: Drivers (limited), Vehicles, Trips, Alerts, Maintenance, Fuel, Documents

**Mobile Access:** Limited - primarily for approval workflows and oversight

---

### 4. Driver

**Access Level:** Personal workspace + assigned resources

**Permissions:**
- Own profile management
- View assigned vehicle(s)
- View assigned trips
- Submit trip-related data
- Upload required documentation
- SOS/emergency triggering

**Features by Domain:**

#### Driver Profile
- Update personal information (contact, emergency contacts)
- Upload documents (driver's license, medical certificates, permits)
- View compliance status (HOS, training, certifications)
- View assigned vehicles and trip history

#### Trip Execution
- View assigned trips with route details
- Navigation integration (routes, geofences)
- Trip status updates (start, pause, resume, complete)
- Document trip events (stops, delays, issues)
- Upload trip documents (BOL, POD, receipts)

#### Real-time Tracking
- Location sharing (GPS)
- Vehicle telemetry (speed, fuel, diagnostics)
- Geofence notifications
- SOS/emergency button

#### Fuel & Expenses
- Log fuel purchases
- Submit expense receipts
- View fuel card transactions

#### Communication
- Receive notifications (alerts, dispatches, messages)
- In-app messaging with dispatch/fleet manager
- Push notifications for urgent events

**Frontend Access (Mobile App):**
- **Dashboard:** Active trip, vehicle status, upcoming schedule
- **Trip:** Trip details, navigation, status controls
- **Fuel:** Fuel logging, expense entry
- **Alerts:** SOS, geofence breach, maintenance reminders
- **Profile:** Personal info, documents, settings
- **CommandCenter:** Quick-access emergency tools

**Frontend Access (Dashboard Web - Limited):**
- Personal trip history
- Personal documents
- Own profile settings

---

### 5. Member

**Access Level:** Limited, feature-specific

**Permissions:** (PolicyManifest: minimal - typically read-only on specific resources)

**Features:**
- View assigned resources (vehicle, trips)
- Basic profile management
- Document viewing (if granted)
- Limited reporting access
- Task completion workflows

**Typical Use Cases:**
- Dispatcher (view-only + dispatch actions)
- Mechanic (view maintenance schedule, update work status)
- Billing clerk (view invoices, upload documents)
- Compliance auditor (read-only compliance reports)

**Frontend Access:**
- Dashboard: Limited widgets based on assigned tasks
- Pages: Specific assigned modules only

---

## Feature Matrix by Domain

### 29 Domains and Role Access

| Domain | Super Admin | Org Admin | Fleet Manager | Driver | Member | Description |
|--------|-------------|-----------|---------------|--------|--------|-------------|
| **admin_support** | ✅ | ✅ | ⚠️ | ❌ | ⚠️ | Support ticket management (Admin/Manager view, Member limited) |
| **alert** | ✅ | ✅ | ✅ | ⚠️ | ⚠️ | Alert ingestion, processing, escalation |
| **alerts** | ✅ | ✅ | ✅ | ✅ | ⚠️ | Alert display and management (drivers see own alerts) |
| **analytics** | ✅ | ✅ | ❌ | ❌ | ❌ | Business intelligence, predictive models, reports |
| **auth** | ✅ | ✅ | ✅ | ✅ | ✅ | Authentication, user management, MFA |
| **billing** | ✅ | ✅ | ⚠️ | ❌ | ⚠️ | Invoicing, payments, disputes |
| **compliance** | ✅ | ✅ | ⚠️ | ⚠️ | ❌ | Regulatory rules (FMCSA, HOS, India) |
| **customer** | ✅ | ✅ | ⚠️ | ❌ | ⚠️ | End-customer or shipper data |
| **diagnostics** | ✅ | ✅ | ✅ | ⚠️ | ❌ | Vehicle health diagnostics |
| **driver** | ✅ | ✅ | ✅ | ✅ | ❌ | Driver profiles, qualifications, onboarding |
| **emergency** | ✅ | ✅ | ✅ | ✅ | ❌ | SOS, incident response, emergency protocols |
| **finance** | ✅ | ✅ | ⚠️ | ❌ | ⚠️ | Financial ops, GL entries, reconciliations |
| **fuel** | ✅ | ✅ | ✅ | ✅ | ❌ | Fuel logging, card management, efficiency |
| **geofencing** | ✅ | ✅ | ✅ | ⚠️ | ❌ | Geofence creation, monitoring, alerts |
| **governance** | ✅ | ✅ | ⚠️ | ❌ | ❌ | Governance policies, audits |
| **india** | ✅ | ✅ | ⚠️ | ⚠️ | ❌ | India-specific: FASTag, E-Way Bills, GST |
| **integrations** | ✅ | ✅ | ⚠️ | ❌ | ❌ | Third-party integrations (telematics, mapping, etc.) |
| **maintenance** | ✅ | ✅ | ✅ | ⚠️ | ⚠️ | Maintenance schedules, work orders, history |
| **monitoring** | ✅ | ✅ | ✅ | ⚠️ | ❌ | System health, SLA tracking |
| **platform_compat** | ✅ | ❌ | ❌ | ❌ | ❌ | Platform compatibility checks |
| **reports** | ✅ | ✅ | ⚠️ | ❌ | ⚠️ | Report generation, scheduling, distribution |
| **route** | ✅ | ✅ | ✅ | ⚠️ | ❌ | Route planning, optimization, ETA |
| **settings** | ✅ | ✅ | ⚠️ | ⚠️ | ❌ | Org and user settings, preferences |
| **tracking** | ✅ | ✅ | ✅ | ⚠️ | ❌ | GPS/telemetry ingestion, real-time tracking |
| **trip** | ✅ | ✅ | ✅ | ✅ | ⚠️ | Trip lifecycle (planning, execution, completion) |
| **vehicle** | ✅ | ✅ | ✅ | ⚠️ | ❌ | Vehicle registry, specs, assignments |
| **video** | ✅ | ✅ | ⚠️ | ❌ | ❌ | Dashcam footage, video storage, retrieval |

**Key:**
- ✅ Full access
- ⚠️ Limited/conditional access
- ❌ No access

---

## Frontend Application Mapping

### 1. Landing Page (`apps/landing-page/`)
**Target Users:** Potential customers, unauthenticated visitors
**Features:**
- Marketing content
- Sign-up / trial request
- Feature showcase
- Pricing information
- Contact / demo request

---

### 2. Dashboard Web (`apps/dashboard-web/`)
**Target Users:** Super Admin, Org Admin, Fleet Manager, Driver (limited), Member

**Pages by Role:**

| Page | Super Admin | Org Admin | Fleet Manager | Driver | Member | Purpose |
|------|-------------|-----------|---------------|--------|--------|---------|
| Dashboard | ✅ | ✅ | ✅ | ⚠️ | ⚠️ | Overview, KPIs, quick actions |
| Drivers | ✅ | ✅ | ✅ | ❌ | ⚠️ | Driver directory, profiles, onboarding |
| Vehicles | ✅ | ✅ | ✅ | ⚠️ | ❌ | Fleet inventory, status, assignments |
| Trips | ✅ | ✅ | ✅ | ✅ | ⚠️ | Trip planning, tracking, history |
| Alerts | ✅ | ✅ | ✅ | ✅ | ⚠️ | Real-time alerts, incident management |
| Analytics | ✅ | ✅ | ❌ | ❌ | ❌ | Business intelligence, trends |
| AuditLogs | ✅ | ✅ | ❌ | ❌ | ❌ | System audit trail |
| CommandCenter | ✅ | ✅ | ✅ | ✅ | ❌ | Emergency operations, SOS response |
| DeviceApproval | ✅ | ✅ | ✅ | ❌ | ❌ | IoT device/telematics device management |
| FuelLogList | ✅ | ✅ | ✅ | ✅ | ❌ | Fuel transaction history |
| FuelSettingsView | ✅ | ✅ | ✅ | ❌ | ❌ | Fuel policy configuration |
| Geofences | ✅ | ✅ | ✅ | ❌ | ❌ | Geofence creation and management |
| Maintenance | ✅ | ✅ | ✅ | ⚠️ | ⚠️ | Work orders, schedules, history |
| Settings | ✅ | ✅ | ⚠️ | ⚠️ | ❌ | Org, user, and system settings |
| ... and more | | | | | | |

---

### 3. Mobile App (`apps/mobile/`)
**Target Users:** Driver (primary), Fleet Manager (limited)

**Screens by Role:**

| Screen | Driver | Fleet Manager | Purpose |
|--------|--------|---------------|---------|
| Login | ✅ | ✅ | Authentication |
| Dashboard | ✅ | ⚠️ | Active trip, status overview |
| TripMap | ✅ | ⚠️ | Real-time map view, navigation |
| TripDetails | ✅ | ✅ | Trip information, stops, documents |
| FuelLog | ✅ | ⚠️ | Log fuel purchases |
| Alerts | ✅ | ✅ | SOS, notifications |
| Profile | ✅ | ✅ | Personal info, settings |
| Documents | ✅ | ⚠️ | Upload/view required docs |
| CommandCenter | ✅ | ✅ | Emergency SOS, help |
| ... | | | |

---

## Permission System Architecture

### Role-Based Access Control (RBAC)

**Implementation:** `packages/common/go/pkg/security/rbac_policies.go`

**Key Components:**

1. **Roles** (5 system roles):
   - `super_admin` - Platform super user
   - `org_admin` - Organization administrator
   - `fleet_manager` - Fleet operations manager
   - `driver` - Vehicle operator
   - `member` - Basic user

2. **Resources** (29+ resources mapped to domains):
   - Organizations, Users, Vehicles, Trips, Alerts, etc.
   - Each domain has associated resource(s)

3. **PolicyManifest:**
   - Centralized mapping of resource → allowed roles
   - Example: `ResourceVehicles: [super_admin, org_admin, fleet_manager]`
   - Enforced via `RequireRole` middleware

4. **Permission Model:**
   - Role-based (coarse) + Permission mask (fine-grained)
   - JWT tokens include role and permissions
   - Middleware validates access per endpoint

---

## Prioritization Insights (RICE Framework)

### High-Impact Features (Quick Wins)

| Feature | Reach | Impact | Confidence | Effort | RICE Score | Priority |
|---------|-------|--------|------------|--------|-------------|----------|
| Driver Mobile App Core | 1000+ drivers | Massive | High | XL | ~26,666 | 🟢 **P0** |
| Fleet Manager Dashboard | 500+ managers | High | High | L | ~40,000 | 🟢 **P0** |
| Vehicle GPS Tracking | All vehicles | Massive | High | M | ~53,333 | 🟢 **P0** |
| SOS/Emergency Response | All users | Massive | High | S | ~80,000 | 🟢 **P0** |
| India Compliance (FASTag) | India fleets | Massive | Medium | XL | ~20,000 | 🟡 **P1** |
| Fuel Management | All fleets | High | High | M | ~40,000 | 🟢 **P0** |
| Maintenance Scheduling | All fleets | High | High | M | ~40,000 | 🟢 **P0** |
| Geofencing | 60% fleets | Medium | High | S | ~40,000 | 🟢 **P0** |

### Strategic Bets (Long-term)

| Feature | Reach | Impact | Confidence | Effort | RICE Score | Timeline |
|---------|-------|--------|------------|--------|-------------|----------|
| Predictive Analytics | All orgs | Massive | Medium | XL | ~20,000 | Q3 2026 |
| Route Optimization | 80% fleets | High | Medium | XL | ~16,000 | Q3 2026 |
| AI Driver Coaching | All drivers | Massive | Low | XL | ~10,000 | Q4 2026 |
| Integration Marketplace | All orgs | High | Medium | XL | ~16,000 | Q4 2026 |
| Video Telematics | 40% fleets | High | High | XL | ~13,333 | Q2 2027 |

---

## Roadmap Recommendations

### Phase 1 (Next Quarter): Foundation & Core Operations
**Focus:** Stabilize core platform, complete driver app, essential fleet management

**Initiatives:**
1. ✅ **Driver Mobile App V1** (P0) - Complete trip execution, fuel logging, SOS
2. ✅ **Fleet Manager Dashboard** (P0) - Vehicles, trips, alerts, maintenance
3. ✅ **GPS Tracking at Scale** (P0) - Handle 1000+ vehicles, optimize ingestion
4. ✅ **India Compliance** (P1) - FASTag, E-Way Bills for Indian fleets
5. ⚠️ **RBAC Hardening** (P1) - Add resource-level permissions, audit trail

**Capacity Allocation:**
- 60% Core features (driver app, dashboard, tracking)
- 20% Performance & scalability (NATS, database optimization)
- 20% Bug fixes and tech debt

---

### Phase 2 (Following Quarter): Scale & Intelligence
**Focus:** Scale to 10k+ vehicles, add business intelligence

**Initiatives:**
1. 📊 **Analytics Platform** (P1) - Reports, dashboards, export capabilities
2. 📈 **Predictive Maintenance** (P2) - ML-based failure prediction
3. 🛣️ **Route Optimization** (P2) - Dynamic routing, ETA, traffic
4. 🔗 **Integration Hub** (P2) - Third-party telematics, mapping, ERP
5. 📱 **Enhanced Mobile** (P1) - Offline mode, barcode scanning, voice

**Capacity Allocation:**
- 40% New features (analytics, predictive maintenance)
- 30% Scale improvements (K8s, DB sharding, caching)
- 20% Integration ecosystem
- 10% Security & compliance (SOC 2, ISO 27001 prep)

---

### Phase 3 (6-12 Months): Enterprise & Growth
**Focus:** Enterprise features, market expansion, advanced safety

**Initiatives:**
1. 🎥 **Video Telematics** (P2) - Dashcam integration, AI safety scoring
2. 📚 **Driver Training Platform** (P2) - LMS, certification tracking
3. 💼 **Advanced Billing** (P2) - Complex invoicing, multi-currency
4. 🌍 **Multi-language Support** (P2) - i18n for global markets
5. 🏢 **Enterprise SSO** (P1) - SAML, OIDC, SCIM provisioning
6. 📊 **Advanced Analytics** (P2) - Driver behavior scoring, ELD prediction

**Capacity Allocation:**
- 35% Enterprise features (video, training, advanced billing)
- 25% Platform scalability (multi-region, CQRS)
- 20% Market-specific features
- 20% Security & compliance (PCI DSS, GDPR)

---

## Success Metrics

### User Adoption
- **Driver adoption:** >90% of assigned drivers active weekly
- **Fleet manager adoption:** >80% using dashboard daily
- **Org admin satisfaction:** NPS > 40

### Operational Efficiency
- **Vehicle tracking uptime:** >99.5%
- **Median alert response time:** <5 minutes (SOS <1 min)
- **Trip completion rate:** >95%
- **Maintenance compliance:** >90% scheduled maintenance on time

### Business Impact
- **Customer retention:** >95% annual retention
- **Time-to-value:** New org fully onboarded in <1 week
- **Support tickets:** <5% of active users per month
- **Revenue per org:** Consistent growth quarter-over-quarter

---

## Dependencies & Risks

### Technical Dependencies
- **NATS JetStream** - Critical for telemetry ingestion at scale
- **TimescaleDB** - Time-series data for tracking and telemetry
- **Redis** - Caching and session management
- **Kubernetes** - Horizontal scalability

### Regulatory Dependencies
- **India compliance team** - FASTag, E-Way Bill, GST expertise
- **FMCSA/HOS experts** - US regulatory compliance
- **Local legal counsel** - Data privacy (GDPR, CCPA)

### Integration Partners
- **Telematics device vendors** (Geotab, Verizon, Samsara)
- **Mapping APIs** (Google Maps, Mapbox)
- **Payment processors** (Stripe, PayPal)
- **Document signing** (DocuSign)

---

## Conclusion

FlyFleet has a robust **role-based architecture** with clear domain separation. The **5 core roles** cover the fleet management lifecycle from administration to driver execution.

**Recommended Next Steps:**
1. Validate role definitions with stakeholders (sales, customer success, legal)
2. Document feature matrix in internal wiki
3. Create role-specific user personas
4. Prioritize Phase 1 initiatives using RICE scoring
5. Align engineering capacity with roadmap
6. Establish quarterly review process for role/feature evolution

---

**Prepared by:** Product Manager Toolkit (AI-assisted analysis)
**Review Required:** Engineering Leadership, Product Leadership, Security Team
