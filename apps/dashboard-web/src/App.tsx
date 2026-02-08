import React from 'react';
import { Provider } from 'react-redux';
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';
import Layout from './components/common/Layout';
import { AdminLayout } from './components/layout/AdminLayout';
import { AuthProvider } from './context/AuthProvider';
import { RBACProvider } from './context/RBACProvider';
import AddFuelLog from './pages/AddFuelLog';
import { AddOrganization } from './pages/admin/AddOrganization';
import BillingDashboard from './pages/admin/billing/BillingDashboard';
import BillingSettingsPage from './pages/admin/billing/BillingSettings';
import InvoiceDetail from './pages/admin/billing/InvoiceDetail';
import InvoiceForm from './pages/admin/billing/InvoiceForm';
import JobDetail from './pages/admin/BulkUploads/JobDetail';
import JobsList from './pages/admin/BulkUploads/JobsList';
import NewUpload from './pages/admin/BulkUploads/NewUpload';
import { Devices } from './pages/admin/Devices';
import BackupDetailView from './pages/admin/dr/BackupDetailView';
import CreateBackupView from './pages/admin/dr/CreateBackupView';
import CreateDRTestView from './pages/admin/dr/CreateDRTestView';
import DetailRestorePage from './pages/admin/dr/DetailRestorePage';
import DRTestDetailView from './pages/admin/dr/DRTestDetailView';
import DRTestListView from './pages/admin/dr/DRTestListView';
import DRTestSettingsView from './pages/admin/dr/DRTestSettingsView';
import InitiateRestorePage from './pages/admin/dr/InitiateRestorePage';
import RestorePointsPage from './pages/admin/dr/RestorePointsPage';
import { Drivers as AdminDrivers } from './pages/admin/Drivers';
import { EditOrganization } from './pages/admin/EditOrganization';
import { AddVehicle as FleetAddVehicle } from './pages/admin/FleetCommand/AddVehicle';
import { EditVehicle as FleetEditVehicle } from './pages/admin/FleetCommand/EditVehicle';
import { VehicleDetail as FleetVehicleDetail } from './pages/admin/FleetCommand/VehicleDetail';
import { VehicleList as FleetVehicleList } from './pages/admin/FleetCommand/VehicleList';
import GlobalNotificationSettings from './pages/admin/notifications/GlobalNotificationSettings';
import NotificationDetail from './pages/admin/notifications/NotificationDetail';
import NotificationLogs from './pages/admin/notifications/NotificationLogs';
import NotificationSettings from './pages/admin/notifications/NotificationSettings';
import UserPreferenceDetail from './pages/admin/notifications/UserPreferenceDetail';
import UserPreferencesList from './pages/admin/notifications/UserPreferencesList';
import { OrganizationDetail } from './pages/admin/OrganizationDetail';
import { Organizations } from './pages/admin/Organizations';
import { OrganizationSettings } from './pages/admin/OrganizationSettings';
import CreateReport from './pages/admin/reports/CreateReport';
import ExecutionHistory from './pages/admin/reports/ExecutionHistory';
import ScheduledReportsList from './pages/admin/reports/ScheduledReportsList';
import { AddTrip, EditTrip, TripDetail, TripSettings, TripsList, TripTelemetry } from './pages/admin/Trips';
import { Users } from './pages/admin/Users';
import Alerts from './pages/Alerts';
import Analytics from './pages/Analytics';
import AuditLogs from './pages/AuditLogs';
import AuditSettings from './pages/AuditSettings';
import Dashboard from './pages/Dashboard';
import Drivers from './pages/Drivers';
import ErrorScreen from './pages/ErrorScreen';
import ForgotPassword from './pages/ForgotPassword';
import FuelSettingsView from './pages/FuelSettingsView';
import Geofences from './pages/Geofences';
import GovernanceSettingsView from './pages/governance/GovernanceSettingsView';
import NewRequest from './pages/governance/NewRequest';
import RequestDetail from './pages/governance/RequestDetail';
import RequestList from './pages/governance/RequestList';
import History from './pages/History';
import Landing from './pages/Landing';
import LiveMap from './pages/LiveMap';
import Login from './pages/Login';
import Maintenance from './pages/Maintenance';
import {
    AddWorkOrder,
    Inventory,
    PartAssignment,
    WorkOrderDetail,
    WorkOrderList,
} from './pages/maintenance';
import MFACenter from './pages/MFACenter';
import MFAExceptions from './pages/MFAExceptions';
import MFAOverrides from './pages/MFAOverrides';
import MFAPolicyEditor from './pages/MFAPolicyEditor';
import MFAVerify from './pages/MFAVerify';
import OrgSelection from './pages/OrgSelection';
import Reports from './pages/Reports';
import ResetPassword from './pages/ResetPassword';
import SessionExpired from './pages/SessionExpired';
import Signup from './pages/Signup';
import Trips from './pages/Trips';
import TrustedDevices from './pages/TrustedDevices';
import VehicleDetails from './pages/VehicleDetails';
import VerifyEmail from './pages/VerifyEmail';
import { store } from './store';
const PrivateRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const token = localStorage.getItem('token');
  return token ? <>{children}</> : <Navigate to="/login" replace />;
};

function App() {
  return (
    <Provider store={store}>
      <AuthProvider>
        <RBACProvider>
          <BrowserRouter>
            <Routes>
              {/* Public Routes */}
              <Route path="/" element={<Landing />} />
              <Route path="/login" element={<Login />} />
              <Route path="/signup" element={<Signup />} />
              <Route path="/forgot-password" element={<ForgotPassword />} />
              <Route path="/mfa-verify" element={<MFAVerify />} />
              <Route path="/reset-password" element={<ResetPassword />} />
              <Route path="/verify-email" element={<VerifyEmail />} />
              <Route path="/select-org" element={<OrgSelection />} />
              <Route path="/session-expired" element={<SessionExpired />} />
              <Route path="/error" element={<ErrorScreen />} />
              <Route path="/trusted-devices" element={<TrustedDevices />} />
              
              {/* User Dashboard */}
              <Route
                path="/dashboard"
                element={
                  <PrivateRoute>
                    <Layout />
                  </PrivateRoute>
                }
              >
                <Route index element={<Dashboard />} />
                <Route path="map" element={<LiveMap />} />
                <Route path="vehicles" element={<FleetVehicleList />} />
                <Route path="vehicles/:id" element={<VehicleDetails />} />
                <Route path="trips" element={<Trips />} />
                <Route path="geofences" element={<Geofences />} />
                <Route path="alerts" element={<Alerts />} />
                <Route path="analytics" element={<Analytics />} />
                <Route path="maintenance" element={<Maintenance />} />
                <Route path="drivers" element={<Drivers />} />
                <Route path="reports" element={<Reports />} />
                <Route path="history" element={<History />} />
                <Route path="security/mfa" element={<MFACenter />} />
              </Route>

              {/* Security Module */}
              <Route path="/security" element={<PrivateRoute><MFACenter /></PrivateRoute>} />
              <Route path="/security/policy" element={<PrivateRoute><MFAPolicyEditor /></PrivateRoute>} />
              <Route path="/security/exceptions" element={<PrivateRoute><MFAExceptions /></PrivateRoute>} />
              <Route path="/security/overrides" element={<PrivateRoute><MFAOverrides /></PrivateRoute>} />
              <Route path="/security/audit" element={<PrivateRoute><AuditLogs /></PrivateRoute>} />
              <Route path="/security/audit/settings" element={<PrivateRoute><AuditSettings /></PrivateRoute>} />

              {/* Admin Portal */}
              <Route
                path="/admin"
                element={
                  <PrivateRoute>
                    <AdminLayout />
                  </PrivateRoute>
                }
              >
                <Route index element={<Navigate to="organizations" replace />} />
                <Route path="organizations" element={<Organizations />} />
                <Route path="organizations/new" element={<AddOrganization />} />
                <Route path="organizations/:id" element={<OrganizationDetail />} />
                <Route path="organizations/:id/edit" element={<EditOrganization />} />
                <Route path="organizations/:id/settings" element={<OrganizationSettings />} />
                <Route path="users" element={<Users />} />
                <Route path="vehicles" element={<FleetVehicleList />} />
                <Route path="vehicles/new" element={<FleetAddVehicle />} />
                <Route path="vehicles/:id" element={<FleetVehicleDetail />} />
                <Route path="vehicles/:id/edit" element={<FleetEditVehicle />} />
                <Route path="drivers" element={<AdminDrivers />} />
                <Route path="devices" element={<Devices />} />
                <Route path="bulk-uploads" element={<JobsList />} />
                <Route path="bulk-uploads/new" element={<NewUpload />} />
                <Route path="bulk-uploads/:id" element={<JobDetail />} />
                <Route path="trips" element={<TripsList />} />
                <Route path="trips/add" element={<AddTrip />} />
                <Route path="trips/settings" element={<TripSettings />} />
                <Route path="trips/:id" element={<TripDetail />} />
                <Route path="trips/:id/edit" element={<EditTrip />} />
                <Route path="trips/:id/telemetry" element={<TripTelemetry />} />
                
                {/* Maintenance Module */}
                <Route path="work-orders" element={<WorkOrderList />} />
                <Route path="work-orders/new" element={<AddWorkOrder />} />
                <Route path="work-orders/:id" element={<WorkOrderDetail />} />
                <Route path="work-orders/:id/parts" element={<PartAssignment />} />
                <Route path="inventory" element={<Inventory />} />
                
                {/* Fuel Module */}
                <Route path="fuel/add" element={<AddFuelLog />} />
                <Route path="fuel/settings" element={<FuelSettingsView />} />
                
                {/* Governance Module */}
                <Route path="governance" element={<RequestList />} />
                <Route path="governance/requests/new" element={<NewRequest />} />
                <Route path="governance/requests/:id" element={<RequestDetail />} />
                <Route path="governance/settings" element={<GovernanceSettingsView />} />
                
                {/* Billing Module */}
                <Route path="billing" element={<BillingDashboard />} />
                <Route path="billing/new" element={<InvoiceForm />} />
                <Route path="billing/invoices/:id" element={<InvoiceDetail />} />
                <Route path="billing/invoices/:id/edit" element={<InvoiceForm />} />
                <Route path="billing/settings" element={<BillingSettingsPage />} />
                
                {/* Notification Module */}
                <Route path="notifications" element={<NotificationLogs />} />
                <Route path="notifications/:id" element={<NotificationDetail />} />
                <Route path="notifications/settings" element={<NotificationSettings />} />
                <Route path="preferences" element={<UserPreferencesList />} />
                <Route path="preferences/global" element={<GlobalNotificationSettings />} />
                <Route path="preferences/:id" element={<UserPreferenceDetail />} />

                {/* Reports Module */}
                <Route path="reports" element={<ScheduledReportsList />} />
                <Route path="reports/create" element={<CreateReport />} />
                <Route path="reports/history" element={<ExecutionHistory />} />

                <Route path="dr" element={<RestorePointsPage />} />
                <Route path="dr/create" element={<CreateBackupView />} />
                <Route path="dr/:id" element={<BackupDetailView />} />
                <Route path="dr/restores" element={<RestorePointsPage />} />
                <Route path="dr/restores/initiate" element={<InitiateRestorePage />} />
                <Route path="dr/restores/:id" element={<DetailRestorePage />} />

                {/* DR Testing Module */}
                <Route path="dr/tests" element={<DRTestListView />} />
                <Route path="dr/tests/create" element={<CreateDRTestView />} />
                <Route path="dr/tests/settings" element={<DRTestSettingsView />} />
                <Route path="dr/tests/:id" element={<DRTestDetailView />} />
                
                {/* Future: Roles, etc */}
              </Route>
            </Routes>
          </BrowserRouter>
        </RBACProvider>
      </AuthProvider>
    </Provider>
  );
}

export default App;
