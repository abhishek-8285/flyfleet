# 🔍 Dashboard-Web — Complete Functionality Audit (207/207 Files)

> **Total Files Scanned**: 207 `.tsx` + `.ts` files
> **Date**: 2026-02-11

---

## Legend

| Symbol | Meaning |
|--------|---------|
| 🔴 | **Fully Static / Broken** — Hardcoded data, no API calls |
| 🟠 | **Partially Broken** — Some parts dynamic, some static or broken |
| ✅ | **Working / Dynamic** — Fetches data from API/Redux |
| 🔵 | **Stub** — Placeholder, feature not built |
| ⚠️ | **Bug / Issue** — Code exists but has a problem |

---

## A. Routing & Architecture Issues

| Issue | File | Detail |
|---|---|---|
| ⚠️ Duplicate route | [App.tsx:274](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/App.tsx#L274) | `rbac/assignments` defined twice |
| ⚠️ Security routes lack Layout | [App.tsx:175-185](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/App.tsx#L175-L185) | 10 `/security/*` routes render without sidebar/header |
| ⚠️ Missing 404 route | App.tsx | No catch-all `*` route — unknown URLs show blank page |
| ⚠️ 4 layout patterns | App.tsx | `/dashboard` → `Layout`, `/admin` → `AdminLayout`, `/security` → no layout, `/support` → `SupportLayout` |
| ⚠️ Dangling route comment | [App.tsx:293-296](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/App.tsx#L293-L296) | `{/* Future: */}` comments and dangling firmware route |

---

## B. Module-by-Module File Classification

### B1. Dashboard Module (7 files)

| File | Status | Notes |
|---|---|---|
| [Dashboard.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/Dashboard.tsx) | 🟠 | Fetches stats/trips/alerts ✅, but mock fallbacks ("842", "125"), non-functional division/timezone/auto-refresh dropdowns |
| [StatsGrid.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/components/dashboard/StatsGrid.tsx) | 🟠 | Dynamic props ✅, but hardcoded source names ("FlyFleet IoT", "Logistics Engine") |
| [QuickActions.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/components/dashboard/QuickActions.tsx) | ✅ | Navigation works via `useNavigate` |
| [CriticalVehiclesTable.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/components/dashboard/CriticalVehiclesTable.tsx) | 🔴 | 3 hardcoded rows, "Showing 1-50 of 842" |
| [ActivityFeed.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/components/dashboard/ActivityFeed.tsx) | 🔴 | 4 static entries, dead "View All" link (`href="#"`) |
| [LiveOperationsMap.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/components/dashboard/LiveOperationsMap.tsx) | 🔴 | Fake map, 2 pins, non-functional zoom, hardcoded geofence |

---

### B2. Vehicle Module (6 files)

| File | Status | Notes |
|---|---|---|
| [Vehicles.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/Vehicles.tsx) | ✅ | Dynamic via vehiclesSlice |
| [VehicleDetails.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/VehicleDetails.tsx) | 🔴 | Ignores `:id` param. "Vehicle #301", "2022 Freightliner Cascadia" |
| [VehicleSpecsPanel.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/components/vehicles/VehicleSpecsPanel.tsx) | 🔴 | 8 specs, driver "Michael Johnson", 3 dead tabs, dead "Download Report" |
| [VehicleMapCard.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/components/vehicles/VehicleMapCard.tsx) | 🔴 | Fake map, "65 mph", "382 miles" |
| [VehicleStatsGrid.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/components/vehicles/VehicleStatsGrid.tsx) | 🔴 | "48,291 mi", "72%", "2,410 hrs" |

---

### B3. Driver Module (4 files)

| File | Status | Notes |
|---|---|---|
| [Drivers.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/Drivers.tsx) | 🟠 | Header exists, non-functional status dropdown, hardcoded footer |
| [DriverTable.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/components/drivers/DriverTable.tsx) | 🔴 | 3 hardcoded rows (John Doe, Sarah Miller, Robert Johnson), "1-8 of 124" |
| [DriverFilters.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/components/drivers/DriverFilters.tsx) | 🔴 | No `onChange`, dead "Add Driver"/"Download Report" buttons |

---

### B4. Trips & History (3 files)

| File | Status | Notes |
|---|---|---|
| [Trips.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/Trips.tsx) | ⚠️ | Dynamic via Redux ✅, but **light-mode styling** (`bg-white`, `text-gray-900`) |
| [History.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/History.tsx) | 🟠 | Fetches trips ✅, but "4h 20m" mocked, hardcoded footer |
| [LiveMap.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/LiveMap.tsx) | ✅ | Dynamic via react-leaflet |

---

### B5. Maintenance Module (5 files)

| File | Status | Notes |
|---|---|---|
| [Maintenance.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/Maintenance.tsx) | 🟠 | Structure ✅, dead dropdowns, "Log New Service" noop, "Last sync: Just now" |
| [MaintenanceStats.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/components/maintenance/MaintenanceStats.tsx) | 🔴 | "Due: 14", "Critical: 3", "Healthy: 96%" |
| [ServiceLogTable.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/components/maintenance/ServiceLogTable.tsx) | 🔴 | 5 rows, dead ticket links, fake pagination "1-5 of 124", dead Export CSV |
| [MaintenanceFilters.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/components/maintenance/MaintenanceFilters.tsx) | 🔴 | Fake search, hardcoded counts (12,8,3,5,2), dead Reset button |

---

### B6. Integrations Module (6 files)

| File | Status | Notes |
|---|---|---|
| [IntegrationsCenter.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/integrations/IntegrationsCenter.tsx) | 🟠 | Tab container works, but child components are all static |
| [Marketplace.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/integrations/components/Marketplace.tsx) | 🔴 | 4 hardcoded cards (Twilio, AWS, SAP, Google Maps) |
| [WebhookManager.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/integrations/components/WebhookManager.tsx) | 🔴 | 1 endpoint, "2.4M deliveries", "99.98% health" |
| [ApiClientHub.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/integrations/components/ApiClientHub.tsx) | 🔴 | "24 clients", "1.2M requests", 2 rows |
| [ConfigForm.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/integrations/components/ConfigForm.tsx) | 🔴 | `setTimeout` mock test, hardcoded client ID/secret |
| [IntegrationLogs.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/integrations/components/IntegrationLogs.tsx) | 🔴 | 2 hardcoded logs, non-functional search/filter/sort, dead "Load More" |

---

### B7. Reports Page (1 file)

| File | Status | Notes |
|---|---|---|
| [Reports.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/Reports.tsx) | 🔴 | 6 static report cards, dead "Generate Report" buttons ×6 |

---

### B8. Fuel Module (5 files)

| File | Status | Notes |
|---|---|---|
| [FuelLogList.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/FuelLogList.tsx) | 🟠 | Dynamic via FuelService ✅, but date dropdown has no `onChange` |
| [AddFuelLog.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/AddFuelLog.tsx) | ✅ | Form submits via FuelService |
| [FuelSettingsView.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/FuelSettingsView.tsx) | ✅ | Fetches/saves via FuelService |
| [FuelStats.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/components/fuel/FuelStats.tsx) | 🟠 | Props dynamic ✅, but "+2.4% from last month" and "Fleet target: 12.5 km/L" hardcoded |
| [FuelLogTable.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/components/fuel/FuelLogTable.tsx) | ✅ | Renders dynamic data via props |

---

### B9. Auth Pages (10 files) — All ✅

| File | Status | Notes |
|---|---|---|
| [LoginPage.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/LoginPage.tsx) | ✅ | Auth via `useAuth()` |
| [Signup.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/Signup.tsx) | ✅ | Dynamic fetch |
| [ForgotPassword.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/ForgotPassword.tsx) | ✅ | Dynamic fetch |
| [ResetPassword.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/ResetPassword.tsx) | ✅ | Dynamic fetch |
| [VerifyEmail.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/VerifyEmail.tsx) | ✅ | Dynamic fetch |
| [MFAVerify.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/MFAVerify.tsx) | 🟠 | OTP verify ✅, but 3 dead "Alternative Methods" buttons, missing resend API (`// Call resend API here`) |
| [OrgSelection.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/OrgSelection.tsx) | 🔴 | 4 hardcoded orgs ("Acme Logistics", "Premium Freight", etc.), `setTimeout` fallback |
| [TrustedDevices.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/TrustedDevices.tsx) | 🔴 | 3 hardcoded devices, `setTimeout` mock delete instead of API |
| [ErrorScreen.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/ErrorScreen.tsx) | ✅ | Static error page — acceptable |
| [SessionExpired.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/SessionExpired.tsx) | ✅ | Static expired page — acceptable |

---

### B10. MFA/Security Module (8 files)

| File | Status | Notes |
|---|---|---|
| [MFACenter.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/MFACenter.tsx) | ✅ | Fetches stats via `api`, uses recharts |
| [MFAExceptions.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/MFAExceptions.tsx) | ✅ | Full CRUD via `api` |
| [MFAOverrides.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/MFAOverrides.tsx) | ✅ | Fetch + toggle via `api` |
| [MFAPolicyEditor.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/MFAPolicyEditor.tsx) | ✅ | Fetch + save via `api` |
| [MFAUserSecurity.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/MFAUserSecurity.tsx) | ✅ | Full CRUD via `api` |
| [Sessions.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/security/Sessions.tsx) | ⚠️ | Dynamic ✅, uses `bg-gray-900` inline instead of design tokens |
| [AccessReview.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/security/AccessReview.tsx) | ✅ | Dynamic via `api` |
| [ThreatAlerts.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/security/ThreatAlerts.tsx) | ✅ | Dynamic via `api` |
| [IPAccess.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/security/IPAccess.tsx) | ✅ | Dynamic via `api` |

---

### B11. Monitoring Module (7 files) — All ✅

| File | Status |
|---|---|
| [SystemHealth.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/monitoring/SystemHealth.tsx) | ✅ |
| [MetricsHub.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/monitoring/MetricsHub.tsx) | ✅ |
| [LogViewer.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/monitoring/LogViewer.tsx) | ✅ |
| [QueueMonitor.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/monitoring/QueueMonitor.tsx) | ✅ |
| [SLAReport.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/monitoring/SLAReport.tsx) | ✅ |
| [MonitoringLayout.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/monitoring/layout/MonitoringLayout.tsx) | ✅ |
| [PageHeader.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/monitoring/components/PageHeader.tsx) | ✅ |

---

### B12. Support Module (8 files)

| File | Status | Notes |
|---|---|---|
| [SupportDashboard.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/support/SupportDashboard.tsx) | ✅ | Dynamic via fetch |
| [IncidentManagement.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/support/IncidentManagement.tsx) | ✅ | Dynamic |
| [IncidentDetail.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/support/IncidentDetail.tsx) | ✅ | Dynamic |
| [TroubleshootingLogs.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/support/TroubleshootingLogs.tsx) | ✅ | Dynamic |
| [UserConsole.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/support/UserConsole.tsx) | 🟠 | Fetches 7 APIs ✅, but hardcoded default `searchId='USR-882910'` |
| [OrganizationHub.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/support/OrganizationHub.tsx) | ✅ | Fetches 7 APIs |
| [OrgConsole.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/support/OrgConsole.tsx) | 🟠 | Dynamic fetch ✅, but "Interactive Map Load Pending..." placeholder stub |
| [SLAReports.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/support/SLAReports.tsx) | 🔴 | Inline mock: `{uptime: "99.98%", avgResponse: "14m", resolutionRate: "92%"}` |

---

### B13. Admin Module (50+ files) — Almost All ✅

All use `apiClient`, `DataGrid`, or direct `fetch` calls. Key issues only:

| File | Status | Notes |
|---|---|---|
| All CRUD pages (Organizations, Users, Drivers, Devices, Vehicles, FleetCommand) | ✅ | Dynamic via `apiClient`/`DataGrid` |
| All Trips pages (List, Detail, Add, Edit, Settings, Telemetry) | ✅ | Dynamic via `apiClient` |
| All BulkUpload pages (JobsList, JobDetail, NewUpload) | ✅ | Dynamic via `apiClient` |
| All Reports pages (Scheduled, Create, ExecutionHistory) | ✅ | Dynamic via `apiClient` |
| All Billing pages (Dashboard, Detail, Form, Settings) | ✅ | Dynamic via `billingService` |
| All Notification pages (Logs, Settings, Detail, Global, UserPrefs) | ✅ | Dynamic via `notificationService` |
| All DR pages (Backups, Restore, Tests) | ✅ | Dynamic via `fetch` |
| All RBAC pages (Roles, Matrix, Assignments, SecurityLogs) | ⚠️ | Dynamic ✅, but use **light-mode** colors (`text-slate-800`, `bg-white`) |
| All Settings pages (Platform, Flags, Localization, Templates, Channels, Policies) | ✅ | Dynamic via `apiClient` |
| All admin/security pages (MFA, Sessions, IP, Threats, Login Logs, Access Review) | ✅ | Dynamic via `apiClient` |
| [OrganizationSettings.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/admin/OrganizationSettings.tsx) | 🟠 | Static nav hub, 2 dead `() => {}` handlers (RBAC + Notifications), uses light-mode colors |
| [Localization.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/admin/settings/Localization.tsx) | 🟠 | API fetch ✅, but falls back to 4 hardcoded mock languages when empty |
| [SecurityLogs.tsx (RBAC)](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/admin/rbac/SecurityLogs.tsx) | 🟠 | Dynamic ✅, dead "Export CSV" button `{ /* Export logic */ }` |
| [Roles.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/admin/rbac/Roles.tsx) | 🟠 | Dynamic ✅, "Create New Role" button opens nothing `{ /* Open create modal */ }` |

---

### B14. Audit Module (4 files) + 2 standalone pages — All ✅

| File | Status |
|---|---|
| [AuditCenter.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/audit/AuditCenter.tsx) | ✅ |
| [AuditList.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/audit/AuditList.tsx) | ✅ |
| [AuditPolicy.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/audit/AuditPolicy.tsx) | ✅ |
| [EntityHistory.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/audit/EntityHistory.tsx) | ✅ |
| [EvidenceExport.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/audit/EvidenceExport.tsx) | ✅ |
| [AuditLogs.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/AuditLogs.tsx) | ✅ |
| [AuditSettings.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/AuditSettings.tsx) | ✅ |

---

### B15. Governance Module (4 files) — All ✅

| File | Status |
|---|---|
| [RequestList.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/governance/RequestList.tsx) | ✅ |
| [NewRequest.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/governance/NewRequest.tsx) | ✅ |
| [RequestDetail.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/governance/RequestDetail.tsx) | ✅ |
| [GovernanceSettingsView.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/governance/GovernanceSettingsView.tsx) | ✅ |

---

### B16. Devices Module (5 files) — All ✅

| File | Status |
|---|---|
| [DeviceDashboard.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/devices/DeviceDashboard.tsx) | ✅ |
| [DeviceDiagnostics.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/devices/DeviceDiagnostics.tsx) | ✅ |
| [DeviceTelemetry.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/devices/DeviceTelemetry.tsx) | ✅ |
| [FirmwareManager.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/devices/FirmwareManager.tsx) | ✅ |
| [RemoteCommands.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/devices/RemoteCommands.tsx) | ✅ |

---

### B17. Documents Module (2 files) — All ✅

| File | Status |
|---|---|
| [DocumentList.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/components/documents/DocumentList.tsx) | ✅ |
| [AddDocumentModal.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/components/documents/AddDocumentModal.tsx) | ✅ |

---

### B18. Core Infrastructure Files

| File | Status | Notes |
|---|---|---|
| [DataGrid.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/components/common/DataGrid.tsx) | ✅ | Reusable dynamic grid with pagination via `apiClient` |
| [api.ts](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/services/api.ts) | ✅ | 85+ methods |
| [apiClient.ts](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/services/apiClient.ts) | ✅ | Axios with interceptors |
| [billingService.ts](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/services/billingService.ts) | ✅ | Dedicated billing API |
| [FuelService.ts](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/services/FuelService.ts) | ✅ | Dedicated fuel API |
| [GovernanceService.ts](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/services/GovernanceService.ts) | ✅ | Dedicated governance API |
| [monitoringService.ts](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/services/monitoringService.ts) | ✅ | Dedicated monitoring API |
| [notificationService.ts](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/services/notificationService.ts) | ✅ | Dedicated notifications API |
| [AuthProvider.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/context/AuthProvider.tsx) | ✅ | Auth context |
| [RBACProvider.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/context/RBACProvider.tsx) | ✅ | RBAC context |
| [Sidebar.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/components/layout/Sidebar.tsx) | ✅ | Navigation sidebar |
| [AdminLayout.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/components/layout/AdminLayout.tsx) | ✅ | Admin layout wrapper |
| [SupportLayout.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/components/support/SupportLayout.tsx) | ✅ | Support layout wrapper |
| [SecuritySidebar.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/components/SecuritySidebar.tsx) | ✅ | Navigation |
| [SecurityFooter.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/components/SecurityFooter.tsx) | ✅ | Footer |
| [SearchPage.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/SearchPage.tsx) | ✅ | Dynamic search |
| [TimelinePage.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/TimelinePage.tsx) | 🟠 | Fetches ✅, but hardcoded `localhost:8081`, fake filters, "Loading older events..." always vis |
| [Geofences.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/Geofences.tsx) | 🔵 | "Geofence management will be displayed here" — despite Redux store existing |
| [Alerts.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/Alerts.tsx) | ✅ | Dynamic via alertsSlice |
| [Analytics.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/Analytics.tsx) | ✅ | Dynamic |

---

## C. Non-Functional UI Elements — Complete List

| Location | Element | Issue |
|---|---|---|
| Dashboard | Auto-refresh dropdown | No timer, visual only |
| Dashboard | Division dropdown | No state binding |
| Dashboard | Timezone dropdown | No state binding |
| ActivityFeed | "View All Activity" | Dead `href="#"` |
| LiveOperationsMap | Zoom +/− buttons | No map library |
| DriverFilters | License/Safety dropdowns | No `onChange` |
| DriverFilters | "Add Driver" button | No handler |
| DriverFilters | "Download Report" button | No handler |
| Drivers page | Status dropdown | No state binding |
| ServiceLogTable | Sort headers | Non-functional |
| ServiceLogTable | "Export CSV" | No handler |
| ServiceLogTable | Pagination buttons | No state |
| ServiceLogTable | Ticket links ×2 | Dead `href="#"` |
| MaintenanceFilters | Search input | No `onChange` |
| MaintenanceFilters | All checkboxes × 5 | No state |
| MaintenanceFilters | "Reset Filters" | No handler |
| Maintenance page | "This Week"/"All Divisions" | No state |
| Maintenance page | "Log New Service" | No handler |
| Reports page | "Generate Report" ×6 | No handler |
| VehicleDetails | "Edit Vehicle" | No handler |
| VehicleDetails | "Decommission" | No handler |
| VehicleSpecsPanel | 3 tabs | No tab state |
| VehicleSpecsPanel | "View Driver Profile" | No navigation |
| VehicleSpecsPanel | "Download Full Report" | No handler |
| ConfigForm | "Save Configuration" | No handler |
| ConfigForm | "Test Connection" | `setTimeout` mock |
| CriticalVehiclesTable | "Track"/"Actions" | No handler |
| TimelinePage | Filter checkboxes | "Mock Filters" |
| IntegrationLogs | Search input | No filtering |
| IntegrationLogs | "Source Integration" filter | No handler |
| IntegrationLogs | "Newest First" sort | No handler |
| IntegrationLogs | "Load More" | No pagination |
| MFAVerify | 3 Alternative Method icons | No `onClick` |
| MFAVerify | Resend code button | `// Call resend API here` |
| OrgSelection | Org list | 4 hardcoded orgs |
| TrustedDevices | Device remove | `setTimeout` mock, not API |
| OrganizationSettings | "RBAC" card | `() => {}` |
| OrganizationSettings | "Notifications" card | `() => {}` |
| Roles.tsx | "Create New Role" | `{ /* Open create modal */ }` |
| SecurityLogs.tsx | "Export CSV" | `{ /* Export logic */ }` |
| FuelLogList | Date range dropdown | No `onChange` |
| OrgConsole | Map section | "Interactive Map Load Pending..." |
| UserConsole | Search ID | Hardcoded `'USR-882910'` |

---

## D. Theme Inconsistencies

| File | Issue |
|---|---|
| [Trips.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/Trips.tsx) | Light-mode: `bg-white`, `text-gray-900`, `bg-gray-50` |
| [Sessions.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/security/Sessions.tsx) | `bg-gray-900` inline instead of tokens |
| [DeviceDashboard.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/devices/DeviceDashboard.tsx) | `bg-slate-950/900/800` instead of tokens |
| RBAC pages (4 files) | `text-slate-800`, `bg-white`, `bg-slate-50` — entire light-mode palette |
| [OrganizationSettings.tsx](file:///home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/pages/admin/OrganizationSettings.tsx) | `text-slate-900`, `bg-white`, `bg-blue-50` — light-mode |

---

## E. Missing Redux Stores

Only 5: `auth`, `vehicles`, `trips`, `alerts`, `geofences`

| Module Missing Store | Current Approach |
|---|---|
| Drivers | DriverTable is hardcoded |
| Maintenance | Components hardcoded |
| Reports | Page is hardcoded |
| Integrations | Components hardcoded |
| Fuel | Uses direct `FuelService` calls (OK) |
| Billing | Uses direct `billingService` calls (OK) |
| Governance | Uses direct `GovernanceService` calls (OK) |

---

## F. Inconsistent Footers (7 variants)

| Page | Copyright | Version |
|---|---|---|
| Dashboard | "© 2026 FlyFleet Global Operations" | "v2.5.0-PRO" |
| Vehicles | "© 2026 Enterprise Fleet Manager" | "v2.4.1" |
| VehicleDetails | "© 2026 Enterprise Fleet Manager" | — |
| History | "© 2026 FlyFleet Enterprise" | "v2.4.1" |
| Drivers | "© 2026 FleetCommand" | "v2.5.0" |
| TimelinePage | "© 2023 Enterprise Fleet Manager" | "V2.4" |
| Reports | "© 2026 FlyFleet Enterprise" | "v2.4.1" |

---

## G. API Methods Without Consumer

| Method in `api.ts` | Frontend Status |
|---|---|
| `getDrivers()`, `getDriver()` | 🔴 DriverTable is hardcoded |
| `getIntegrations()` | 🔴 Marketplace hardcoded |
| `getApiClients()`, `createApiClient()` | 🔴 ApiClientHub hardcoded |
| `getWebhooks()` | 🔴 WebhookManager hardcoded |
| `testIntegration()` | 🔴 ConfigForm uses `setTimeout` |

---

## H. Final Summary Statistics

| Category | Count |
|---|---|
| **Total Files** | **207** |
| ✅ Fully Working / Dynamic | **~140** |
| 🔴 Fully Static Components | **20** |
| 🟠 Partially Broken | **12** |
| 🔵 Stub Pages | **1** (Geofences) |
| ⚠️ Routing Bugs | **3** |
| ⚠️ Theme Inconsistencies | **8** files |
| Dead Buttons / Links | **42** |
| Non-functional Dropdowns/Filters | **14** |
| API Methods Without Consumer | **~8** |
| Missing Redux Stores | **4** (Drivers, Maintenance, Reports, Integrations) |
| Inconsistent Footers | **7** variants |
| Hardcoded Mock Fallbacks | **3** (Localization, OrgSelection, TrustedDevices) |
