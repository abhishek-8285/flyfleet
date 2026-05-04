import React, { Suspense, lazy } from "react";
import { Provider } from "react-redux";
import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import Layout from "./components/common/Layout";
import { SOSOverlay } from "./components/common/SOSOverlay";
import { CustomerPortalLayout } from "./components/layout/CustomerPortalLayout";
import { GlobalAdminLayout } from "./components/layout/GlobalAdminLayout";
import SupportLayout from "./components/support/SupportLayout";
import { AuthProvider } from "./context/AuthProvider";
import { RBACProvider } from "./context/RBACProvider";
import AddFuelLog from "./pages/AddFuelLog";
import { AddTenant } from "./pages/admin/AddTenant";
import BillingDashboard from "./pages/admin/billing/BillingDashboard";
import BillingSettingsPage from "./pages/admin/billing/BillingSettings";
import InvoiceDetail from "./pages/admin/billing/InvoiceDetail";
import InvoiceForm from "./pages/admin/billing/InvoiceForm";
import JobDetail from "./pages/admin/BulkUploads/JobDetail";
import JobsList from "./pages/admin/BulkUploads/JobsList";
import NewUpload from "./pages/admin/BulkUploads/NewUpload";
import { DeviceDetail } from "./pages/admin/DeviceDetail";
import { Devices } from "./pages/admin/Devices";
import BackupDetailView from "./pages/admin/dr/BackupDetailView";
import CreateBackupView from "./pages/admin/dr/CreateBackupView";
import CreateDRTestView from "./pages/admin/dr/CreateDRTestView";
import DetailRestorePage from "./pages/admin/dr/DetailRestorePage";
import DRTestDetailView from "./pages/admin/dr/DRTestDetailView";
import DRTestListView from "./pages/admin/dr/DRTestListView";
import DRTestSettingsView from "./pages/admin/dr/DRTestSettingsView";
import InitiateRestorePage from "./pages/admin/dr/InitiateRestorePage";
import RestorePointsPage from "./pages/admin/dr/RestorePointsPage";
import { DriverList } from "./pages/admin/DriverList";
import { DriverDetail } from "./pages/admin/DriverDetail";
import AddDriver from "./pages/admin/Drivers/AddDriver";
import { EditDriver } from "./pages/admin/EditDriver";
import { Drivers as AdminDrivers } from "./pages/admin/Drivers";
import { EditTenant } from "./pages/admin/EditTenant";
import FleetAddVehicle from "./pages/admin/FleetCommand/AddVehicle";
import FleetEditVehicle from "./pages/admin/FleetCommand/EditVehicle";
import FleetVehicleDetail from "./pages/admin/FleetCommand/VehicleDetail";
import { VehicleList as FleetVehicleList } from "./pages/admin/VehicleList";
import { GlobalAdminDashboard } from "./pages/admin/GlobalAdminDashboard";
import GlobalNotificationSettings from "./pages/admin/notifications/GlobalNotificationSettings";
import NotificationDetail from "./pages/admin/notifications/NotificationDetail";
import NotificationLogs from "./pages/admin/notifications/NotificationLogs";
import NotificationSettings from "./pages/admin/notifications/NotificationSettings";
import UserPreferenceDetail from "./pages/admin/notifications/UserPreferenceDetail";
import UserPreferencesList from "./pages/admin/notifications/UserPreferencesList";
import { TenantDetail } from "./pages/admin/TenantDetail";
import { Tenants } from "./pages/admin/Tenants";
import { TenantSettings } from "./pages/admin/TenantSettings";
import { TenantUsers } from "./pages/admin/TenantUsers";
import { TenantVehicles } from "./pages/admin/TenantVehicles";
import { RoleMatrix } from "./pages/admin/rbac/RoleMatrix";
import { Roles } from "./pages/admin/rbac/Roles";
import { SecurityLogs } from "./pages/admin/rbac/SecurityLogs";
import TemplateDesignerPage from "./pages/TemplateDesigner";
import { UserAssignments } from "./pages/admin/rbac/UserAssignments";
import CreateReport from "./pages/admin/reports/CreateReport";
import ExecutionHistory from "./pages/admin/reports/ExecutionHistory";
import ScheduledReportsList from "./pages/admin/reports/ScheduledReportsList";
import CustomerLogin from "./pages/portal/CustomerLogin";
import InvoicesList from "./pages/portal/InvoicesList";
import NewBooking from "./pages/portal/NewBooking";
import PortalDashboard from "./pages/portal/PortalDashboard";
import ShipmentsList from "./pages/portal/ShipmentsList";
import ShipmentTracking from "./pages/portal/ShipmentTracking";
import SessionTimeoutWarning from "./components/auth/SessionTimeoutWarning";
import LiveMap from "./components/LiveMap";
import ScenarioSwitcher from "./components/ScenarioSwitcher";
import ControlTower from "./modules/control-tower/ControlTower";
import Alerts from "./modules/dashboard/pages/Alerts";
import AccessDenied from "./pages/AccessDenied";
import AccountLocked from "./pages/AccountLocked";
import AccessReview from "./pages/admin/security/AccessReview";
import ActiveSessions from "./pages/admin/security/ActiveSessions";
import AuthStats from "./pages/admin/security/AuthStats";
import IPControls from "./pages/admin/security/IPControls";
import LoginActivityLogs from "./pages/admin/security/LoginActivityLogs";
import MFACenter from "./pages/admin/security/MFACenter";
import SecurityTelemetry from "./pages/admin/security/SecurityTelemetry";
import ThreatAlerts from "./pages/admin/security/ThreatAlerts";
import UserSessionAdmin from "./pages/admin/security/UserSessionAdmin";
import FeatureFlags from "./pages/admin/settings/FeatureFlags";
import Localization from "./pages/admin/settings/Localization";
import NotificationChannels from "./pages/admin/settings/NotificationChannels";
import PlatformConfig from "./pages/admin/settings/PlatformConfig";
import PoliciesLimits from "./pages/admin/settings/PoliciesLimits";
import Templates from "./pages/admin/settings/Templates";
import {
    AddTrip,
    EditTrip,
    TripDetail,
    TripSettings,
    TripsList,
    TripTelemetry,
} from "./pages/admin/Trips";
import { Users } from "./pages/admin/Users";
import Drivers from "./pages/Drivers";
import ErrorScreen from "./pages/ErrorScreen";
import ForgotPassword from "./pages/ForgotPassword";
import FuelSettingsView from "./pages/FuelSettingsView";
import GovernanceSettingsView from "./pages/governance/GovernanceSettingsView";
import NewRequest from "./pages/governance/NewRequest";
import RequestDetail from "./pages/governance/RequestDetail";
import RequestList from "./pages/governance/RequestList";
import History from "./pages/History";
import Insights from "./pages/Insights";
import {
    AddWorkOrder,
    InspectionForm,
    Inventory,
    PartAssignment,
    VendorDirectory,
    VendorForm,
    WorkOrderDetail,
    WorkOrderList,
} from "./pages/maintenance";
import OrgSelection from "./pages/OrgSelection";
import Reports from "./pages/Reports";
import ResetPassword from "./pages/ResetPassword";
import RiskChallenge from "./pages/RiskChallenge";
import SearchPage from "./pages/SearchPage";
import APIKeyManagement from "./pages/security/APIKeys";
import ChangePassword from "./pages/security/ChangePassword";
import SecurityDevices from "./pages/security/Devices";
import PasswordPolicy from "./pages/security/PasswordPolicy";
import SecurityKeys from "./pages/security/SecurityKeys";
import SessionExpired from "./pages/SessionExpired";
import Signup from "./pages/Signup";
import SSOCallback from "./pages/SSOCallback";
import IncidentDetail from "./pages/support/IncidentDetail";
import IncidentManagement from "./pages/support/IncidentManagement";
import TenantHub from "./pages/support/TenantHub";
import SLAReports from "./pages/support/SLAReports";
import SupportDashboard from "./pages/support/SupportDashboard";
import TroubleshootingLogs from "./pages/support/TroubleshootingLogs";
import UserConsole from "./pages/support/UserConsole";
import Trips from "./pages/Trips";
import TrustedDevices from "./pages/TrustedDevices";
import VehicleDetails from "./pages/VehicleDetails";
import VerifyEmail from "./pages/VerifyEmail";
import ResetPasswordForced from "./pages/ResetPasswordForced";
import { CustomerList } from "./pages/admin/CustomerList";
import { AddCustomer } from "./pages/admin/AddCustomer";
import { CustomerDetail } from "./pages/admin/CustomerDetail";
import { store } from "./store";
import { setToken, setUser as setReduxUser } from "./store/authSlice";
import { LoadingSpinner } from "./components/common/LoadingSpinner";
import { extractUserRole, isRoleAllowed } from "./utils/rbac";
import { useSelector } from "react-redux";
import type { RootState } from "./store";

// Lazy loaded routes for code splitting
const Analytics = lazy(() => import("./modules/dashboard/pages/Analytics"));
const Dashboard = lazy(() => import("./modules/dashboard/pages/Dashboard"));
const TimelinePage = lazy(() => import("./modules/dashboard/pages/TimelinePage"));
const AnalyticsJobsPage = lazy(() => import("./pages/admin/analytics/AnalyticsJobsPage"));
const DriverBehaviorPage = lazy(() => import("./pages/admin/analytics/DriverBehaviorPage"));
const FleetAnalyticsPage = lazy(() => import("./pages/admin/analytics/FleetAnalyticsPage"));
const PredictiveMaintenancePage = lazy(() => import("./pages/admin/analytics/PredictiveMaintenancePage"));
const PredictiveModelsPage = lazy(() => import("./pages/admin/analytics/PredictiveModelsPage"));
const DispatchCenter = lazy(() => import("./pages/admin/DispatchCenter"));
const EmergencyDashboard = lazy(() => import("./pages/admin/emergency/EmergencyDashboard"));
const ExpenseApprovals = lazy(() => import("./pages/admin/finance/ExpenseApprovals"));
const FuelDashboard = lazy(() => import("./pages/admin/finance/FuelDashboard"));
const PolicySettings = lazy(() => import("./pages/admin/finance/PolicySettings"));
const IntegrationHub = lazy(() => import("./pages/admin/integrations/IntegrationHub"));
const WorkOrderManagement = lazy(() => import("./pages/admin/maintenance/WorkOrderManagement"));
const ForecastingDashboard = lazy(() => import("./pages/admin/revenue/ForecastingDashboard"));
const RouteOptimization = lazy(() => import("./pages/admin/RouteOptimization"));
const VideoTelematics = lazy(() => import("./pages/admin/VideoTelematics"));
const AuditCenter = lazy(() => import("./pages/audit/AuditCenter"));
const DeviceApproval = lazy(() => import("./pages/DeviceApproval"));
const DeviceDashboard = lazy(() => import("./pages/devices/DeviceDashboard"));
const DeviceDiagnostics = lazy(() => import("./pages/devices/DeviceDiagnostics"));
const DeviceTelemetry = lazy(() => import("./pages/devices/DeviceTelemetry"));
const DeviceTwins = lazy(() => import("./pages/devices/DeviceTwins"));
const FirmwareManager = lazy(() => import("./pages/devices/FirmwareManager"));
const RemoteCommands = lazy(() => import("./pages/devices/RemoteCommands"));
// IntegrationsCenter removed — IntegrationHub is the canonical integrations page
const Landing = lazy(() => import("./pages/Landing"));
const Login = lazy(() => import("./pages/LoginPage"));
const Register = lazy(() => import("./pages/RegisterPage"));
const Maintenance = lazy(() => import("./pages/Maintenance"));
const SystemMaintenance = lazy(() => import("./pages/SystemMaintenance"));
const MFAExceptions = lazy(() => import("./pages/MFAExceptions"));
const MFAOverrides = lazy(() => import("./pages/MFAOverrides"));
const MFAPolicyEditor = lazy(() => import("./pages/MFAPolicyEditor"));
const MFASetup = lazy(() => import("./pages/MFASetup"));
const MFAVerify = lazy(() => import("./pages/MFAVerify"));
const MonitoringLayout = lazy(() => import("./pages/monitoring/layout/MonitoringLayout"));
const LogViewer = lazy(() => import("./pages/monitoring/LogViewer"));
const MetricsHub = lazy(() => import("./pages/monitoring/MetricsHub"));
const QueueMonitor = lazy(() => import("./pages/monitoring/QueueMonitor"));
const SLAReport = lazy(() => import("./pages/monitoring/SLAReport"));
const SystemHealth = lazy(() => import("./pages/monitoring/SystemHealth"));
const AuthError = lazy(() => import("./pages/AuthError"));

// Redux is the single source of truth for auth state.
// Fallback to localStorage during page refresh recovery.
const PrivateRoute: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const token = useSelector((state: RootState) => state.auth.token);
  const isLoading = useSelector((state: RootState) => state.auth.loading);
  if (isLoading) return <LoadingSpinner />;
  const hasToken = !!token || !!localStorage.getItem('access_token');
  return hasToken ? <>{children}</> : <Navigate to="/login" replace />;
};

/**
 * Role-based route guard.
 * Redirects to /unauthorized if the user's role is not in the allowed list.
 */
const RoleGuard: React.FC<{
  allowed: string[];
  children: React.ReactNode;
}> = ({ allowed, children }) => {
  const user = useSelector((state: RootState) => state.auth.user);
  const effectiveUser = user || JSON.parse(localStorage.getItem('user_data') || 'null');
  const role = extractUserRole(effectiveUser);

  if (isRoleAllowed(role, allowed)) {
    return <>{children}</>;
  }

  return <Navigate to="/unauthorized" replace />;
};

/** Convenience: combines auth check + role check */
const PermissionRoute: React.FC<{
  allowed: string[];
  children: React.ReactNode;
}> = ({ allowed, children }) => (
  <PrivateRoute>
    <RoleGuard allowed={allowed}>{children}</RoleGuard>
  </PrivateRoute>
);

const DynamicAdminLayout: React.FC = () => {
  const user = useSelector((state: RootState) => state.auth.user);
  const isLoading = useSelector((state: RootState) => state.auth.loading);
  const token = useSelector((state: RootState) => state.auth.token);

  // Always wait for auth hydration to complete before making routing decisions
  if (isLoading) return <LoadingSpinner />;

  const effectiveUser = user || JSON.parse(localStorage.getItem('user_data') || 'null');
  const hasToken = !!token || !!localStorage.getItem('access_token');

  // No session at all — go to login
  if (!hasToken) {
    return <Navigate to="/login" replace />;
  }

  // Token exists but user not yet in state — may be a render race during org switch.
  // If user_data is also missing from localStorage, then it truly is a bad session.
  if (!effectiveUser) {
    const savedUser = localStorage.getItem('user_data');
    if (!savedUser) {
      // Genuinely corrupt session — clear and login
      localStorage.removeItem('access_token');
      localStorage.removeItem('refresh_token');
      store.dispatch(setToken(null));
      store.dispatch(setReduxUser(null));
    }
    // In both cases redirect to login — but only clear storage if truly corrupt
    return <Navigate to="/login" replace />;
  }

  const role = extractUserRole(effectiveUser).toLowerCase();

  // Allow all admin-level roles (centralized via isRoleAllowed)
  if (isRoleAllowed(role, ['admin', 'super_admin', 'support', 'org_admin', 'fleet_manager', 'global_admin', 'system_admin'])) {
    return <GlobalAdminLayout />;
  }

  // Standard Users (driver, staff) MUST NOT access /admin
  return <Navigate to="/dashboard" replace />;
};

function App() {
  return (
    <Provider store={store}>
        <AuthProvider>
          <RBACProvider>
            <SessionTimeoutWarning
              onLogout={() => (window.location.href = "/login")}
            />
            <BrowserRouter future={{ v7_startTransition: true, v7_relativeSplatPath: true }}>
            <Routes>
              {/* Public Routes */}
              <Route path="/" element={
                <Suspense fallback={<LoadingSpinner />}>
                  <Landing />
                </Suspense>
              } />
              <Route path="/login" element={
                <Suspense fallback={<LoadingSpinner />}>
                  <Login />
                </Suspense>
              } />
              <Route path="/register" element={
                <Suspense fallback={<LoadingSpinner />}>
                  <Register />
                </Suspense>
              } />
              <Route path="/signup" element={<Signup />} />
              <Route path="/public/designer" element={
                <Suspense fallback={<LoadingSpinner />}>
                  <TemplateDesignerPage />
                </Suspense>
              } />
              <Route path="/forgot-password" element={<ForgotPassword />} />
              <Route path="/mfa-verify" element={<MFAVerify />} />
              <Route path="/reset-password" element={<ResetPassword />} />
              <Route path="/reset-password-forced" element={<ResetPasswordForced />} />
              {/* Fix 7: Deduplicated /verify-email — was declared twice */}
              <Route path="/verify-email" element={<VerifyEmail />} />
              <Route path="/account-locked" element={<AccountLocked />} />
              <Route path="/auth/challenge" element={<RiskChallenge />} />
              <Route path="/auth/callback" element={<SSOCallback />} />
              <Route path="/auth/error" element={
                <Suspense fallback={<LoadingSpinner />}>
                  <AuthError />
                </Suspense>
              } />
              <Route path="/403" element={<AccessDenied />} />

              {/* Customer Portal Public Routes */}
              <Route path="/portal/login" element={<CustomerLogin />} />
              <Route path="/portal/track/:id" element={<ShipmentTracking />} />

              <Route path="/session-expired" element={<SessionExpired />} />
              <Route path="/error" element={<ErrorScreen />} />
              <Route path="/system-maintenance" element={<SystemMaintenance />} />

              {/* Fix 8: Security routes must be protected — unauthenticated access was possible */}
              <Route
                path="/security"
                element={
                  <PrivateRoute>
                    <Layout />
                  </PrivateRoute>
                }
              >
                <Route path="change-password" element={<ChangePassword />} />
                <Route path="devices" element={<SecurityDevices />} />
                <Route path="api-keys" element={<APIKeyManagement />} />
                <Route path="security-keys" element={<SecurityKeys />} />
                <Route path="password-policy" element={<PasswordPolicy />} />
              </Route>

              {/* MFA setup requires auth context */}
              <Route path="/mfa-setup" element={<PrivateRoute><MFASetup /></PrivateRoute>} />
              <Route path="/trusted-devices" element={<PrivateRoute><TrustedDevices /></PrivateRoute>} />
              <Route path="/device-approval" element={<PrivateRoute><DeviceApproval /></PrivateRoute>} />

              {/* User Dashboard */}
              <Route
                path="/dashboard"
                element={
                  <PrivateRoute>
                    <Layout />
                  </PrivateRoute>
                }
              >
                <Route index element={
                  <Suspense fallback={<LoadingSpinner />}>
                    <Dashboard />
                  </Suspense>
                } />
                <Route path="kpi" element={
                  <Suspense fallback={<LoadingSpinner />}>
                    <Analytics />
                  </Suspense>
                } />
                <Route path="activity" element={<Alerts />} />
                <Route path="alerts" element={<Alerts />} />
                <Route path="analytics" element={
                  <Suspense fallback={<LoadingSpinner />}>
                    <Analytics />
                  </Suspense>
                } />
                <Route path="map" element={
                  <Suspense fallback={<LoadingSpinner />}>
                    <LiveMap />
                  </Suspense>
                } />
                <Route path="vehicles" element={<FleetVehicleList />} />
                <Route path="vehicles/:id" element={<VehicleDetails />} />
                <Route path="trips" element={<Trips />} />
                {/* Fix 5: Removed duplicate drivers and reports routes */}
                <Route path="drivers" element={<Drivers />} />
                <Route path="insights" element={<Insights />} />
                <Route path="reports" element={<Reports />} />
                <Route path="search" element={<SearchPage />} />
                <Route path="timeline" element={
                  <Suspense fallback={<LoadingSpinner />}>
                    <TimelinePage />
                  </Suspense>
                } />
                <Route
                  path="analytics/models"
                  element={
                    <Suspense fallback={<LoadingSpinner />}>
                      <PredictiveModelsPage />
                    </Suspense>
                  }
                />
                <Route path="analytics/jobs" element={
                  <Suspense fallback={<LoadingSpinner />}>
                    <AnalyticsJobsPage />
                  </Suspense>
                } />
                <Route
                  path="analytics/fleet"
                  element={
                    <Suspense fallback={<LoadingSpinner />}>
                      <FleetAnalyticsPage />
                    </Suspense>
                  }
                />
                <Route
                  path="analytics/drivers"
                  element={
                    <Suspense fallback={<LoadingSpinner />}>
                      <DriverBehaviorPage />
                    </Suspense>
                  }
                />
                <Route
                  path="analytics/maintenance"
                  element={
                    <Suspense fallback={<LoadingSpinner />}>
                      <PredictiveMaintenancePage />
                    </Suspense>
                  }
                />
                <Route path="maintenance" element={<Maintenance />} />
                <Route path="history" element={<History />} />
                <Route path="security/mfa" element={<MFACenter />} />
              </Route>

              {/* Security Management Module (Expansion) — already wrapped inside /security PrivateRoute above */}

              {/* Admin Portal — layout is open to all authenticated users, but child routes are role-gated */}
              <Route
                path="/admin"
                element={
                  <PrivateRoute>
                    <SOSOverlay />
                    <DynamicAdminLayout />
                  </PrivateRoute>
                }
              >
                <Route index element={<GlobalAdminDashboard />} />
                <Route path="dashboard" element={<GlobalAdminDashboard />} />
                {/* Admin-only: org management, user management */}
                <Route path="tenants" element={
                  <RoleGuard allowed={["super_admin"]}><Tenants /></RoleGuard>
                } />
                <Route path="tenants/new" element={
                  <RoleGuard allowed={["super_admin"]}><AddTenant /></RoleGuard>
                } />
                <Route
                  path="tenants/:id"
                  element={<RoleGuard allowed={["super_admin"]}><TenantDetail /></RoleGuard>}
                />
                <Route
                  path="tenants/:id/edit"
                  element={<RoleGuard allowed={["super_admin"]}><EditTenant /></RoleGuard>}
                />
                <Route
                  path="tenants/:id/settings"
                  element={<RoleGuard allowed={["super_admin"]}><TenantSettings /></RoleGuard>}
                />
                <Route
                  path="tenants/:id/users"
                  element={<RoleGuard allowed={["super_admin"]}><TenantUsers /></RoleGuard>}
                />
                <Route
                  path="tenants/:id/vehicles"
                  element={<RoleGuard allowed={["super_admin"]}><TenantVehicles /></RoleGuard>}
                />
                <Route path="users" element={
                  <RoleGuard allowed={["admin", "super_admin", "org_admin", "fleet_manager"]}><TenantUsers /></RoleGuard>
                } />
                <Route path="vehicles" element={<FleetVehicleList />} />
                <Route path="vehicles/new" element={<FleetAddVehicle />} />
                <Route path="vehicles/:id" element={<FleetVehicleDetail />} />
                <Route
                  path="vehicles/:id/edit"
                  element={<FleetEditVehicle />} 
                />

                <Route path="drivers" element={<DriverList />} />
                <Route path="drivers/new" element={<AddDriver />} />
                <Route path="drivers/:id" element={<DriverDetail />} />
                <Route path="drivers/:id/edit" element={<EditDriver />} />
                <Route path="devices" element={<Devices />} />
                <Route path="devices/:id" element={<DeviceDetail />} />
                <Route path="bulk-uploads" element={<JobsList />} />
                <Route path="bulk-uploads/new" element={<NewUpload />} />
                <Route path="bulk-uploads/:id" element={<JobDetail />} />
                <Route path="trips" element={<TripsList />} />
                <Route
                  path="route-optimization"
                  element={<RouteOptimization />}
                />
                <Route path="video" element={<VideoTelematics />} />
                <Route path="billing/designer" element={<TemplateDesignerPage />} />
                <Route path="trips/add" element={<AddTrip />} />
                <Route path="trips/settings" element={<TripSettings />} />
                <Route path="trips/:id" element={<TripDetail />} />
                <Route path="trips/:id/edit" element={<EditTrip />} />
                <Route path="trips/:id/telemetry" element={<TripTelemetry />} />

                {/* Customer Management */}
                <Route path="customers" element={<CustomerList />} />
                <Route path="customers/new" element={<AddCustomer />} />
                <Route path="customers/:id" element={<CustomerDetail />} />

                {/* Maintenance Module */}
                <Route
                  path="maintenance/work-orders"
                  element={<WorkOrderManagement />}
                />
                <Route path="work-orders" element={<WorkOrderList />} />
                <Route path="work-orders/new" element={<AddWorkOrder />} />
                <Route path="work-orders/:id" element={<WorkOrderDetail />} />
                <Route
                  path="work-orders/:id/parts"
                  element={<PartAssignment />}
                />
                <Route
                  path="work-orders/:id/parts"
                  element={<PartAssignment />}
                />
                <Route path="inventory" element={<Inventory />} />
                <Route path="vendors" element={<VendorDirectory />} />
                <Route path="vendors/new" element={<VendorForm />} />
                <Route path="vendors/:id" element={<VendorForm />} />
                <Route path="inspections/new" element={<InspectionForm />} />
                <Route path="inspections/:id" element={<InspectionForm />} />

                {/* Finance Module */}
                <Route path="finance/fuel" element={<FuelDashboard />} />
                <Route path="finance/policy" element={<PolicySettings />} />
                <Route path="finance/expenses" element={<ExpenseApprovals />} />

                {/* Emergency Module */}
                <Route path="emergency" element={<EmergencyDashboard />} />

                {/* Control Tower Module */}
                <Route path="control-tower" element={<ControlTower />} />
                <Route path="dispatch" element={<DispatchCenter />} />

                {/* Revenue Module */}
                <Route
                  path="revenue/forecasting"
                  element={<ForecastingDashboard />}
                />

                {/* Integrations Module */}
                <Route path="integrations" element={<IntegrationHub />} />

                {/* Fuel Module */}
                <Route path="fuel/add" element={<AddFuelLog />} />
                <Route path="fuel/settings" element={<FuelSettingsView />} />

                {/* Governance Module */}
                <Route path="governance" element={<RequestList />} />
                <Route
                  path="governance/requests/new"
                  element={<NewRequest />}
                />
                <Route
                  path="governance/requests/:id"
                  element={<RequestDetail />}
                />
                <Route
                  path="governance/settings"
                  element={<RoleGuard allowed={["admin", "super_admin"]}><GovernanceSettingsView /></RoleGuard>}
                />

                {/* Billing Module — admin only */}
                <Route path="billing" element={<RoleGuard allowed={["admin", "super_admin"]}><BillingDashboard /></RoleGuard>} />
                <Route path="billing/new" element={<RoleGuard allowed={["admin", "super_admin"]}><InvoiceForm /></RoleGuard>} />
                <Route
                  path="billing/invoices/:id"
                  element={<RoleGuard allowed={["admin", "super_admin"]}><InvoiceDetail /></RoleGuard>}
                />
                <Route
                  path="billing/invoices/:id/edit"
                  element={<RoleGuard allowed={["admin", "super_admin"]}><InvoiceForm /></RoleGuard>}
                />
                <Route
                  path="billing/settings"
                  element={<RoleGuard allowed={["admin", "super_admin"]}><BillingSettingsPage /></RoleGuard>}
                />

                {/* Notification Module */}
                <Route path="notifications" element={<NotificationLogs />} />
                <Route
                  path="notifications/:id"
                  element={<NotificationDetail />}
                />
                <Route
                  path="notifications/settings"
                  element={<NotificationSettings />}
                />
                <Route path="preferences" element={<UserPreferencesList />} />
                <Route
                  path="preferences/global"
                  element={<GlobalNotificationSettings />}
                />
                <Route
                  path="preferences/:id"
                  element={<UserPreferenceDetail />}
                />

                {/* Reports Module — admin + fleet_manager */}
                <Route path="reports" element={<RoleGuard allowed={["admin", "super_admin", "fleet_manager"]}><ScheduledReportsList /></RoleGuard>} />
                <Route path="reports/create" element={<RoleGuard allowed={["admin", "super_admin", "fleet_manager"]}><CreateReport /></RoleGuard>} />
                <Route path="reports/history" element={<RoleGuard allowed={["admin", "super_admin", "fleet_manager"]}><ExecutionHistory /></RoleGuard>} />

                {/* Disaster Recovery — admin only */}
                <Route path="dr" element={<RoleGuard allowed={["admin", "super_admin"]}><RestorePointsPage /></RoleGuard>} />
                <Route path="dr/create" element={<RoleGuard allowed={["admin", "super_admin"]}><CreateBackupView /></RoleGuard>} />
                <Route path="dr/:id" element={<RoleGuard allowed={["admin", "super_admin"]}><BackupDetailView /></RoleGuard>} />
                <Route path="dr/restores" element={<RoleGuard allowed={["admin", "super_admin"]}><RestorePointsPage /></RoleGuard>} />
                <Route
                  path="dr/restores/initiate"
                  element={<RoleGuard allowed={["admin", "super_admin"]}><InitiateRestorePage /></RoleGuard>}
                />
                <Route path="dr/restores/:id" element={<RoleGuard allowed={["admin", "super_admin"]}><DetailRestorePage /></RoleGuard>} />

                {/* DR Testing — admin only */}
                <Route path="dr/tests" element={<RoleGuard allowed={["admin", "super_admin"]}><DRTestListView /></RoleGuard>} />
                <Route path="dr/tests/create" element={<RoleGuard allowed={["admin", "super_admin"]}><CreateDRTestView /></RoleGuard>} />
                <Route
                  path="dr/tests/settings"
                  element={<RoleGuard allowed={["admin", "super_admin"]}><DRTestSettingsView /></RoleGuard>}
                />
                <Route path="dr/tests/:id" element={<RoleGuard allowed={["admin", "super_admin"]}><DRTestDetailView /></RoleGuard>} />

                {/* RBAC — admin + super_admin only */}
                <Route path="rbac/roles" element={<RoleGuard allowed={["admin", "super_admin"]}><Roles /></RoleGuard>} />
                <Route
                  path="rbac/roles/:id/permissions"
                  element={<RoleGuard allowed={["admin", "super_admin"]}><RoleMatrix /></RoleGuard>}
                />
                <Route path="rbac/assignments" element={<RoleGuard allowed={["admin", "super_admin"]}><UserAssignments /></RoleGuard>} />
                <Route path="rbac/logs" element={<RoleGuard allowed={["admin", "super_admin"]}><SecurityLogs /></RoleGuard>} />

                {/* Global Settings — admin only */}
                <Route path="settings/platform" element={<RoleGuard allowed={["admin", "super_admin"]}><PlatformConfig /></RoleGuard>} />
                <Route path="settings/flags" element={<RoleGuard allowed={["admin", "super_admin"]}><FeatureFlags /></RoleGuard>} />
                <Route
                  path="settings/localization"
                  element={<RoleGuard allowed={["admin", "super_admin"]}><Localization /></RoleGuard>}
                />
                <Route path="settings/templates" element={<RoleGuard allowed={["admin", "super_admin"]}><Templates /></RoleGuard>} />
                <Route
                  path="settings/channels"
                  element={<RoleGuard allowed={["admin", "super_admin"]}><NotificationChannels /></RoleGuard>}
                />
                <Route path="settings/policies" element={<RoleGuard allowed={["admin", "super_admin"]}><PoliciesLimits /></RoleGuard>} />

                {/* Device Diagnostics — admin + fleet_manager */}
                <Route
                  path="devices/diagnostics"
                  element={<RoleGuard allowed={["admin", "super_admin", "fleet_manager"]}><DeviceDashboard /></RoleGuard>}
                />
                <Route path="devices/twins" element={<RoleGuard allowed={["admin", "super_admin", "fleet_manager"]}><DeviceTwins /></RoleGuard>} />
                <Route path="devices/telemetry" element={<RoleGuard allowed={["admin", "super_admin", "fleet_manager"]}><DeviceTelemetry /></RoleGuard>} />
                <Route path="devices/commands" element={<RoleGuard allowed={["admin", "super_admin"]}><RemoteCommands /></RoleGuard>} />
                <Route
                  path="devices/:id/diagnostics"
                  element={<RoleGuard allowed={["admin", "super_admin", "fleet_manager"]}><DeviceDiagnostics /></RoleGuard>}
                />
                <Route
                  path="devices/:id/telemetry"
                  element={<DeviceTelemetry />}
                />
                <Route
                  path="devices/:id/commands"
                  element={<RemoteCommands />}
                />
                <Route
                  path="devices/:id/firmware"
                  element={<FirmwareManager />}
                />

                {/* Future: etc */}
                <Route path="devices/firmware" element={<FirmwareManager />} />

                {/* Future: etc */}
              </Route>

              {/* Monitoring Module — admin + fleet_manager */}
              <Route
                path="/admin/monitoring"
                element={
                  <RoleGuard allowed={["super_admin"]}>
                    <MonitoringLayout />
                  </RoleGuard>
                }
              >
                <Route index element={<SystemHealth />} />
                <Route path="metrics" element={<MetricsHub />} />
                <Route path="logs" element={<LogViewer />} />
                <Route path="jobs" element={<QueueMonitor />} />
                <Route path="sla" element={<SLAReport />} />
              </Route>

              {/* Support module — admin only */}
              <Route
                path="/support"
                element={
                  <PrivateRoute>
                    <RoleGuard allowed={["super_admin"]}>
                      <SupportLayout>
                        <SupportDashboard />
                      </SupportLayout>
                    </RoleGuard>
                  </PrivateRoute>
                }
              />
              <Route
                path="/support"
                element={
                  <PrivateRoute>
                    <RoleGuard allowed={["super_admin"]}>
                      <SupportLayout>{null}</SupportLayout>
                    </RoleGuard>
                  </PrivateRoute>
                }
              >
                <Route index element={<SupportDashboard />} />
                <Route path="dashboard" element={<SupportDashboard />} />
                <Route path="incidents" element={<IncidentManagement />} />
                <Route path="incidents/:id" element={<IncidentDetail />} />
                <Route path="logs" element={<TroubleshootingLogs />} />
                <Route path="user-console" element={<UserConsole />} />
                <Route path="tenant-console/:id" element={<TenantHub />} />
                <Route path="reports" element={<SLAReports />} />
              </Route>

              {/* Customer Portal - Protected */}
              <Route path="/portal" element={<CustomerPortalLayout />}>
                <Route
                  index
                  element={<Navigate to="/portal/dashboard" replace />}
                />
                <Route path="dashboard" element={<PortalDashboard />} />
                <Route path="shipments" element={<ShipmentsList />} />
                <Route path="bookings" element={<NewBooking />} />
                <Route path="invoices" element={<InvoicesList />} />
                <Route
                  path="settings"
                  element={
                    <div className="p-6 text-white">Settings (Placeholder)</div>
                  }
                />
              </Route>

              {/* 404 Catch-all */}
              <Route path="*" element={<Navigate to="/error" replace />} />
              {/* Unauthorized page */}
              <Route path="/unauthorized" element={
                <div className="min-h-screen flex items-center justify-center bg-[#0a1118]">
                  <div className="text-center p-8">
                    <span className="material-symbols-outlined text-6xl text-rose-500 mb-4">block</span>
                    <h1 className="text-2xl font-bold text-white mb-2">Access Denied</h1>
                    <p className="text-text-muted mb-6">You don't have permission to view this page.</p>
                    <a href="/admin" className="px-6 py-2 bg-primary text-white rounded-lg hover:bg-primary/80 transition-colors">
                      Go to Dashboard
                    </a>
                  </div>
                </div>
              } />
            </Routes>
            {/* <ScenarioSwitcher /> Disabled: dev-only mock widget */}
          </BrowserRouter>
        </RBACProvider>
      </AuthProvider>
    </Provider>
  );
}

export default App;
