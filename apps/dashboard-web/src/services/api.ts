import axios, { AxiosError, AxiosInstance } from "axios";
import {
    createTraceId,
    REQUEST_ID_HEADER,
    TRACE_ID_HEADER,
} from "../utils/trace";

// Get store from global reference (set in store/index.ts after creation) to avoid circular dependency
const getStore = () => (window as any).__REDUX_STORE__;

let API_URL = import.meta.env.VITE_API_URL || "/api/v1";
if (/^https?:\/\/[^\/]+\/?$/.test(API_URL)) {
  API_URL = API_URL.replace(/\/$/, '') + '/api/v1';
}
const BFF_URL = import.meta.env.VITE_BFF_URL || "/";

class ApiService {
  private client: AxiosInstance;
  private bffClient: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: API_URL,
      timeout: 10000,
      headers: {
        "Content-Type": "application/json",
      },
      withCredentials: true, // Send refresh cookie automatically
    });

    this.bffClient = axios.create({
      baseURL: API_URL, // Use API_URL (/api/v1) instead of BFF_URL (/) for consistency with monolith
      timeout: 10000,
      headers: {
        "Content-Type": "application/json",
      },
      withCredentials: true,
    });

    const setupInterceptors = (instance: AxiosInstance) => {
      instance.interceptors.request.use(
        (config) => {
          // SECURITY: Get token from Redux store (memory only), not localStorage
          const state = getStore().getState();
          const token = state.auth.token || localStorage.getItem('access_token');
          if (token) {
            config.headers.Authorization = `Bearer ${token}`;
          }
          // Attach tenant context — required by TenantMiddleware on all non-auth routes
          const tenantId = localStorage.getItem('tenant_id');
          if (tenantId && token) {
            config.headers['X-Tenant-ID'] = tenantId;
          }
          const traceId = createTraceId();
          config.headers[TRACE_ID_HEADER] = traceId;
          config.headers[REQUEST_ID_HEADER] = traceId;
          return config;
        },
        (error) => Promise.reject(error),
      );

      instance.interceptors.response.use(
        (response) => response.data,
        (error: AxiosError) => {
          const config = error.config;
          const isAuthEndpoint = config?.url?.includes("/auth/") || config?.url?.includes("/login");

          if (error.response?.status === 401 && !isAuthEndpoint) {
            // Token invalid, clear state and redirect
            getStore().dispatch({ type: 'auth/setToken', payload: null });
            getStore().dispatch({ type: 'auth/setUser', payload: null });
            if (window.location.pathname !== "/session-expired") {
              window.location.href = `/session-expired?redirect=${encodeURIComponent(window.location.pathname)}`;
            }
          }
          return Promise.reject(error);
        },
      );
    };

    setupInterceptors(this.client);
    setupInterceptors(this.bffClient);
  }

  // Generic Methods
  async get(url: string, config?: any): Promise<any> {
    return this.client.get(url, config);
  }

  async post(url: string, data?: any, config?: any): Promise<any> {
    return this.client.post(url, data, config);
  }

  async put(url: string, data?: any, config?: any): Promise<any> {
    return this.client.put(url, data, config);
  }

  async delete(url: string, config?: any): Promise<any> {
    return this.client.delete(url, config);
  }

  // Auth
  async login(email: string, password: string): Promise<any> {
    return this.client.post("/auth/login", { email, password });
  }

  async refreshToken(): Promise<any> {
    return this.client.post("/auth/refresh");
  }

  async logout() {
    await this.client.post("/auth/logout", {}, { withCredentials: true });
    // No localStorage cleanup needed
  }

  // Auth Extensions
  async signup(data: any): Promise<any> {
    return this.client.post("/auth/register", data);
  }

  async verifyEmail(token: string): Promise<any> {
    return this.client.post("/auth/verify-email", { token });
  }

  async resetPassword(data: any): Promise<any> {
    return this.client.post("/auth/reset-password", data);
  }

  async changePassword(data: any): Promise<any> {
    return this.client.post("/auth/change-password", data);
  }

  async resetPasswordForced(data: any): Promise<any> {
    return this.client.post("/auth/reset-password-forced", data);
  }

  async approveDevice(fingerprint: string, approve: boolean): Promise<any> {
    return this.client.post("/auth/device/approve", { fingerprint, approve });
  }

  async ssoCallback(data: any): Promise<any> {
    return this.client.post("/auth/sso/callback", data);
  }

  // MFA Enrollment
  async setupMFA(): Promise<any> {
    return this.client.post("/auth/mfa/setup");
  }

  async verifyMFA(code: string): Promise<any> {
    return this.client.post("/auth/mfa/verify", { code });
  }

  // Trusted Devices
  async getDevices(): Promise<any> {
    return this.client.get("/auth/devices");
  }

  async revokeDevice(id: string): Promise<any> {
    return this.client.delete(`/auth/devices/${id}`);
  }

  // Security Keys (FIDO2)
  async registerSecurityKey(data: any): Promise<any> {
    return this.client.post("/auth/security-keys/register", data);
  }

  async deleteSecurityKey(id: string): Promise<any> {
    return this.client.delete(`/auth/security-keys/${id}`);
  }

  async getSecurityKeys(): Promise<any> {
    return this.client.get("/auth/security-keys");
  }

  // Password Policy
  async getPasswordPolicy(): Promise<any> {
    return this.client.get("/auth/password-policy");
  }

  // Account Lock
  async getLockStatus(): Promise<any> {
    return this.client.get("/auth/lock-status");
  }

  // Risk Challenge
  async resolveRiskChallenge(challengeId: string, data: any): Promise<any> {
    return this.client.post("/auth/challenge", { challengeId, ...data });
  }

  // Phase 3: Advanced Security Management
  async getAPIKeys(): Promise<any> {
    return this.client.get("/auth/api-keys");
  }

  async createAPIKey(name: string): Promise<any> {
    return this.client.post("/auth/api-keys", { name });
  }

  async revokeAPIKey(id: string): Promise<any> {
    return this.client.delete(`/auth/api-keys/${id}/revoke`);
  }

  async getSecurityTelemetry(): Promise<any> {
    return this.client.get("/auth/security-telemetry");
  }

  async getAuthStats(): Promise<any> {
    return this.client.get("/auth/stats");
  }

  async getUserSessions(userId?: string): Promise<any> {
    const url = userId ? `/auth/users/${userId}/sessions` : "/auth/sessions";
    return this.client.get(url);
  }

  async revokeUserSession(userId: string, sessionId: string): Promise<any> {
    return this.client.delete(
      `/auth/users/${userId}/sessions/${sessionId}/revoke`,
    );
  }

  // Vehicles
  async getVehicles(params?: any): Promise<any> {
    return this.client.get("/admin/vehicles", { params });
  }

  async searchVehicles(q: string, params?: any): Promise<any> {
    return this.client.get("/admin/vehicles/search", { params: { ...params, q } });
  }

  async getVehicle(id: string): Promise<any> {
    return this.client.get(`/admin/vehicles/${id}`);
  }

  async getTrips(params?: any): Promise<any> {
    return this.client.get("/admin/trips", { params });
  }

  async createVehicle(data: any): Promise<any> {
    return this.client.post("/admin/vehicles", data);
  }

  async updateVehicle(id: string, data: any): Promise<any> {
    return this.client.put(`/admin/vehicles/${id}`, data);
  }

  async deleteVehicle(id: string) {
    await this.client.delete(`/admin/vehicles/${id}`);
  }

  // Dispatch Trips
  async getDispatchTrips(params?: any): Promise<any> {
    return this.client.get("/admin/trips", { params });
  }

  async getTripCustomers(): Promise<any> {
    return this.client.get("/admin/trips/customers");
  }

  async getTrip(id: string): Promise<any> {
    return this.client.get(`/admin/trips/${id}`);
  }

  async createTrip(data: any): Promise<any> {
    return this.client.post("/admin/trips", data);
  }

  async updateTrip(id: string, data: any): Promise<any> {
    return this.client.put(`/admin/trips/${id}`, data);
  }

  async deleteTrip(id: string): Promise<any> {
    return this.client.delete(`/admin/trips/${id}`);
  }

  async getTripTelemetry(id: string, limit: number = 50): Promise<any> {
    return this.client.get(`/admin/trips/${id}/telemetry`, {
      params: { limit },
    });
  }

  async getTripSettings(): Promise<any> {
    return this.client.get("/admin/trips/settings");
  }

  async updateTripSettings(data: any): Promise<any> {
    return this.client.put("/admin/trips/settings", data);
  }

  // Users
  async getUsers(tenantId: string): Promise<any> {
    return this.client.get("/admin/users", { params: { tenant_id: tenantId } });
  }

  async adminCreateUser(data: any): Promise<any> {
    return this.client.post("/admin/users/create", data);
  }

  // Tenants
  async getTenants(params?: any): Promise<any> {
    return this.client.get("/admin/tenants", { params });
  }

  async getOrganizations(params?: any): Promise<any> {
    return this.getTenants(params);
  }

  async getTenant(id: string): Promise<any> {
    return this.client.get(`/admin/tenants/${id}`);
  }

  async createTenant(data: any): Promise<any> {
    return this.client.post("/admin/tenants", data);
  }

  async updateTenant(id: string, data: any): Promise<any> {
    return this.client.put(`/admin/tenants/${id}`, data);
  }

  async deleteTenant(id: string) {
    await this.client.delete(`/admin/tenants/${id}`);
  }

  async getTenantStats(): Promise<any> {
    // Try to get stats, fallback to empty/default if backend doesn't support it yet
    try {
      return await this.client.get("/admin/tenants/stats");
    } catch (error) {
      console.warn("Tenant stats API not available, using fallback");
      return {
        data: {
          total_tenants: 0,
          active_tenants: 0,
          pending_tenants: 0,
          inactive_tenants: 0,
          total_users: 0,
          total_vehicles: 0,
          total_devices: 0,
          active_alerts: 0,
        },
      };
    }
  }

  // Drivers
  async getDrivers(params?: any): Promise<any> {
    return this.client.get("/admin/drivers", { params });
  }

  async getDriver(id: string): Promise<any> {
    return this.client.get(`/admin/drivers/${id}`);
  }

  async createDriver(data: any): Promise<any> {
    return this.client.post("/admin/drivers", data);
  }

  async updateDriver(id: string, data: any): Promise<any> {
    return this.client.put(`/admin/drivers/${id}`, data);
  }

  async deleteDriver(id: string): Promise<any> {
    return this.client.delete(`/admin/drivers/${id}`);
  }

  async updateDriverStatus(id: string, status: string): Promise<any> {
    return this.client.post(`/admin/drivers/${id}/update-status`, { status });
  }

  async assignVehicle(driverId: string, vehicleId: string): Promise<any> {
    return this.client.post(`/admin/drivers/${driverId}/assign-vehicle`, { vehicle_id: vehicleId });
  }

  async unassignVehicle(driverId: string): Promise<any> {
    return this.client.post(`/admin/drivers/${driverId}/unassign-vehicle`);
  }

  async getDriverAnalytics(id: string): Promise<any> {
    return this.client.get(`/admin/drivers/${id}/analytics`);
  }

  async getDriverBehaviorScore(id: string): Promise<any> {
    return this.client.get(`/admin/drivers/${id}/behavior-score`);
  }

  async getDriverDocuments(id: string): Promise<any> {
    return this.client.get(`/admin/drivers/${id}/documents`);
  }

  async uploadDriverDocument(id: string, data: any): Promise<any> {
    return this.client.post(`/admin/drivers/${id}/documents`, data);
  }

  async deleteDriverDocument(id: string, docId: string): Promise<any> {
    return this.client.delete(`/admin/drivers/${id}/documents/${docId}`);
  }

  // Alerts
  async getAlerts(params?: any): Promise<any> {
    return this.client.get("/alerts", { params });
  }

  async acknowledgeAlert(id: string): Promise<any> {
    return this.client.post(`/admin/alerts/${id}/acknowledge`);
  }

  // Geofences
  async getGeofences(): Promise<any> {
    return this.client.get("/admin/geofences");
  }

  async createGeofence(data: any): Promise<any> {
    return this.client.post("/admin/geofences", data);
  }

  async deleteGeofence(id: string): Promise<any> {
    await this.client.delete(`/admin/geofences/${id}`);
  }

  // Statistics
  async getStatistics(): Promise<any> {
    return this.bffClient.get("/dashboard/stats");
  }

  async getAdminDashboardData(timeRange?: string): Promise<any> {
    return this.bffClient.get("/dashboard/data", {
      params: { time_range: timeRange },
    });
  }

  async getTimeline(): Promise<any> {
    return this.bffClient.get("/timeline");
  }

  // Documents
  async getDocuments(vehicleId: string): Promise<any> {
    return this.client.get("/admin/documents", {
      params: { vehicle_id: vehicleId },
    });
  }

  async createDocument(vehicleId: string, data: any): Promise<any> {
    return this.client.post("/admin/documents", data, {
      params: { vehicle_id: vehicleId },
    });
  }

  async updateDocument(id: string, data: any): Promise<any> {
    return this.client.put(`/admin/documents/${id}`, data);
  }

  async deleteDocument(id: string): Promise<any> {
    return this.client.delete(`/admin/documents/${id}`);
  }

  async getDriverStats(): Promise<any> {
    return this.client.get("/admin/analytics/drivers");
  }

  async getTripStats(): Promise<any> {
    return this.client.get("/admin/analytics/trips");
  }

  // Bulk Uploads
  async getBulkJobs(params?: any): Promise<any> {
    return this.client.get("/admin/bulk-jobs", { params });
  }

  async createBulkJob(file: File): Promise<any> {
    const formData = new FormData();
    formData.append("file", file);
    return this.client.post("/admin/bulk-jobs", formData, {
      headers: {
        "Content-Type": "multipart/form-data",
      },
    });
  }

  async getBulkJob(id: string): Promise<any> {
    return this.client.get(`/admin/bulk-jobs/${id}`);
  }

  async getBulkJobRecords(id: string, params?: any): Promise<any> {
    return this.client.get(`/admin/bulk-jobs/${id}/records`, { params });
  }

  // MFA Security (Routed through BFF)
  async getMFADashboard(tenantId: string): Promise<any> {
    return this.bffClient.get("/security/mfa/dashboard", {
      params: { tenant_id: tenantId },
    });
  }

  async getMFAAdoptionTrend(tenantId: string, days: number = 30): Promise<any> {
    return this.bffClient.get("/security/mfa/adoption-trend", {
      params: { tenant_id: tenantId, days },
    });
  }

  async getMFAMethodDistribution(tenantId: string): Promise<any> {
    return this.bffClient.get("/security/mfa/method-distribution", {
      params: { tenant_id: tenantId },
    });
  }

  async getMFAEvents(
    tenantId: string,
    page: number = 1,
    pageSize: number = 20,
  ): Promise<any> {
    return this.bffClient.get("/security/mfa/events", {
      params: { tenant_id: tenantId, page, page_size: pageSize },
    });
  }

  async getMFAFailureLocations(tenantId: string): Promise<any> {
    return this.bffClient.get("/security/mfa/failures/locations", {
      params: { tenant_id: tenantId },
    });
  }

  async getMFAFailureUsers(tenantId: string): Promise<any> {
    return this.bffClient.get("/security/mfa/failures/users", {
      params: { tenant_id: tenantId },
    });
  }

  async getLockedUsers(tenantId: string): Promise<any> {
    return this.bffClient.get("/security/mfa/locked-users", {
      params: { tenant_id: tenantId },
    });
  }

  async unlockUser(userId: string): Promise<any> {
    return this.bffClient.post("/security/mfa/unlock", { userId });
  }

  async resetMFA(userId: string): Promise<any> {
    return this.bffClient.post("/security/mfa/reset", { userId });
  }

  async getMFAPolicy(tenantId: string): Promise<any> {
    return this.bffClient.get("/security/mfa/policy", {
      params: { tenant_id: tenantId },
    });
  }

  async updateMFAPolicy(data: { policy: any; reason: string }): Promise<any> {
    return this.bffClient.put("/security/mfa/policy", data);
  }

  async getMFAPolicyHistory(tenantId: string): Promise<any> {
    return this.bffClient.get("/security/mfa/policy/history", {
      params: { tenant_id: tenantId },
    });
  }

  async rollbackMFAPolicy(historyId: string): Promise<any> {
    return this.bffClient.post("/security/mfa/policy/rollback", {
      historyId,
    });
  }

  async getMFATenantOverride(tenantId: string): Promise<any> {
    return this.bffClient.get("/security/mfa/overrides", {
      params: { tenant_id: tenantId },
    });
  }

  async getMFAOrgOverride(tenantId: string): Promise<any> {
    return this.getMFATenantOverride(tenantId);
  }

  async updateMFATenantOverride(data: any): Promise<any> {
    return this.bffClient.put("/security/mfa/overrides", data);
  }

  async updateMFAOrgOverride(data: any): Promise<any> {
    return this.updateMFATenantOverride(data);
  }

  // Audit Module
  async getAuditLogs(params: {
    tenant_id: string;
    page?: number;
    action?: string;
    severity?: string;
    user_id?: string;
  }): Promise<any> {
    return this.bffClient.get("/audit/logs", { params });
  }

  async getAuditLogDetails(id: string, tenant_id: string): Promise<any> {
    return this.bffClient.get(`/audit/logs/${id}`, {
      params: { tenant_id },
    });
  }

  async getAuditSettings(tenantId: string): Promise<any> {
    return this.bffClient.get("/audit/settings", {
      params: { tenant_id: tenantId },
    });
  }

  async updateAuditSettings(data: any): Promise<any> {
    return this.bffClient.put("/audit/settings", data);
  }

  async getMFAExceptions(tenantId: string): Promise<any> {
    return this.bffClient.get("/security/mfa/exceptions", {
      params: { tenant_id: tenantId },
    });
  }

  async addMFAException(data: {
    userId: string;
    tenantId: string;
    reason: string;
    days: number;
  }): Promise<any> {
    return this.bffClient.post("/security/mfa/exceptions", {
      ...data,
      tenant_id: data.tenantId,
    });
  }

  async deleteMFAException(id: string): Promise<any> {
    return this.bffClient.delete(`/security/mfa/exceptions/${id}`);
  }

  // Security Operations Expansion
  async getLoginActivityLogs(params: {
    page?: number;
    size?: number;
    status?: string;
  }): Promise<any> {
    return this.bffClient.get("/security/logins", { params });
  }

  async getThreatAlerts(params: {
    page?: number;
    size?: number;
    severity?: string;
    status?: string;
  }): Promise<any> {
    return this.bffClient.get("/security/alerts", { params });
  }

  async updateThreatAlertStatus(id: string, status: string): Promise<any> {
    return this.bffClient.patch(`/security/alerts/${id}/status`, {
      status,
    });
  }

  async getSecuritySessions(tenantId: string): Promise<any> {
    return this.bffClient.get("/security/sessions", {
      params: { tenant_id: tenantId },
    });
  }

  async revokeSecuritySession(sessionId: string): Promise<any> {
    return this.bffClient.delete(`/security/sessions/${sessionId}`);
  }

  async getIPRules(tenantId: string): Promise<any> {
    return this.bffClient.get("/security/ip-rules", {
      params: { tenant_id: tenantId },
    });
  }

  async createIPRule(data: any): Promise<any> {
    return this.bffClient.post("/security/ip-rules", data);
  }

  // Audit & Compliance Extensions
  async getEntityHistory(type: string, id: string): Promise<any> {
    return this.bffClient.get("/audit/entity-history", {
      params: { type, id },
    });
  }

  async createEvidenceExport(data: {
    tenant_id: string;
    module: string;
    date_range_start: string;
    date_range_end: string;
    format: string;
  }): Promise<any> {
    return this.bffClient.post("/audit/evidence-exports", data);
  }

  async getEvidenceExports(tenantId: string): Promise<any> {
    return this.bffClient.get("/audit/evidence-exports", {
      params: { tenant_id: tenantId },
    });
  }

  async getEvidenceExportDetails(id: string): Promise<any> {
    return this.bffClient.get(`/audit/evidence-exports/${id}`);
  }

  async getAuditPolicy(tenantId: string): Promise<any> {
    return this.bffClient.get("/audit/policy", {
      params: { tenant_id: tenantId },
    });
  }

  async updateAuditPolicy(data: any): Promise<any> {
    return this.bffClient.put("/audit/policy", data);
  }

  // Integrations Module
  async getIntegrations(tenantId: string): Promise<any> {
    return this.bffClient.get("/integrations/", {
      params: { tenant_id: tenantId },
    });
  }

  async createIntegration(data: any): Promise<any> {
    return this.bffClient.post("/integrations/", data);
  }

  async updateIntegration(id: string, data: any): Promise<any> {
    return this.bffClient.put(`/integrations/${id}`, data);
  }

  async testIntegration(id: string): Promise<any> {
    return this.bffClient.post(`/integrations/${id}/test`);
  }

  async getApiClients(tenantId: string): Promise<any> {
    return this.bffClient.get("/api-clients/", {
      params: { tenant_id: tenantId },
    });
  }

  async createApiClient(data: any): Promise<any> {
    return this.bffClient.post("/api-clients/", data);
  }

  async getWebhooks(tenantId: string): Promise<any> {
    return this.bffClient.get("/webhooks/", {
      params: { tenant_id: tenantId },
    });
  }

  async getIntegrationLogs(params: {
    tenant_id: string;
    integration_id?: string;
    limit?: number;
    offset?: number;
  }): Promise<any> {
    return this.bffClient.get("/integrations/logs", { params });
  }
  // Route Optimization
  async createOptimizationRequest(data: any): Promise<any> {
    return this.client.post("/routes/requests", data);
  }

  async getOptimizationRequest(id: string): Promise<any> {
    return this.client.get(`/routes/requests/${id}`);
  }

  async listOptimizationRequests(params?: any): Promise<any> {
    return this.client.get("/routes/requests", { params });
  }

  async getOptimizationResult(id: string): Promise<any> {
    return this.client.get(`/routes/results/${id}`);
  }

  async listOptimizationResults(params?: any): Promise<any> {
    return this.client.get("/routes/results", { params });
  }

  async listOptimizationJobs(params?: any): Promise<any> {
    return this.client.get("/routes/jobs", { params });
  }

  // Video Telematics
  // Video Telematics
  async createVideoStream(data: any): Promise<any> {
    return this.client.post("/admin/video/streams", data);
  }

  async getVideoStream(id: string): Promise<any> {
    return this.client.get(`/admin/video/streams/${id}`);
  }

  async listVideoStreams(params?: any): Promise<any> {
    return this.client.get("/admin/video/streams", { params });
  }

  async listVideoRecordings(params?: any): Promise<any> {
    return this.client.get("/admin/video/recordings", { params });
  }

  async getVideoAnalytics(params?: any): Promise<any> {
    return this.client.get("/admin/video/analytics", { params });
  }

  // Advanced Analytics
  async getPredictiveMaintenance(params?: any): Promise<any> {
    return this.client.get("/admin/analytics/maintenance", { params });
  }

  async getDriverBehaviorAnalytics(params?: any): Promise<any> {
    return this.client.get("/admin/analytics/driver-behavior", { params });
  }

  async getFleetUtilization(params?: any): Promise<any> {
    return this.client.get("/admin/analytics/utilization", { params });
  }

  async listAnalyticsJobs(params?: any): Promise<any> {
    return this.client.get("/admin/analytics/jobs", { params });
  }

  async createAnalyticsJob(data: any): Promise<any> {
    return this.client.post("/admin/analytics/jobs", data);
  }

  // Disaster Recovery
  async initiateBackup(type: string): Promise<any> {
    return this.client.post("/admin/dr/backups", { type });
  }

  async getBackups(params?: any): Promise<any> {
    return this.client.get("/admin/dr/backups", { params });
  }

  async initiateRestore(backupId: string, targetEnv: string): Promise<any> {
    return this.client.post("/admin/dr/restore-jobs", {
      backup_id: backupId,
      target_environment: targetEnv,
    });
  }

  async getRestoreJobs(params?: any): Promise<any> {
    return this.client.get("/admin/dr/restore-jobs", { params });
  }
}

export const api = new ApiService();
export default api;
