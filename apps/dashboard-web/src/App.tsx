import React from 'react';
import { Provider } from 'react-redux';
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';
import Layout from './components/common/Layout';
import { AdminLayout } from './components/layout/AdminLayout';
import { AuthProvider } from './context/AuthProvider';
import { RBACProvider } from './context/RBACProvider';
import { Devices } from './pages/admin/Devices';
import { Drivers as AdminDrivers } from './pages/admin/Drivers';
import { Organizations } from './pages/admin/Organizations';
import { Users } from './pages/admin/Users';
import { Vehicles } from './pages/admin/Vehicles';
import Alerts from './pages/Alerts';
import Analytics from './pages/Analytics';
import Dashboard from './pages/Dashboard';
import Drivers from './pages/Drivers';
import ErrorScreen from './pages/ErrorScreen';
import ForgotPassword from './pages/ForgotPassword';
import Geofences from './pages/Geofences';
import History from './pages/History';
import Landing from './pages/Landing';
import LiveMap from './pages/LiveMap';
import Login from './pages/Login';
import Maintenance from './pages/Maintenance';
import MFACenter from './pages/MFACenter';
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

// ... existing imports ...

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
                <Route path="vehicles" element={<Vehicles />} />
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
                <Route path="users" element={<Users />} />
                <Route path="vehicles" element={<Vehicles />} />
                <Route path="drivers" element={<AdminDrivers />} />
                <Route path="devices" element={<Devices />} />
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
