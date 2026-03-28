import axios, { AxiosError, AxiosInstance } from "axios";
import {
    createTraceId,
    REQUEST_ID_HEADER,
    TRACE_ID_HEADER,
} from "../utils/trace";
import { store } from "../store";

const API_URL = import.meta.env.VITE_API_URL || "/api/v1";
const BFF_URL = process.env.REACT_APP_BFF_URL || "/";

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
      baseURL: BFF_URL,
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
          const state = store.getState();
          const token = state.auth.token;
          if (token) {
            config.headers.Authorization = `Bearer ${token}`;
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
          if (error.response?.status === 401) {
            // Token invalid, clear state and redirect
            store.dispatch({ type: 'auth/setToken', payload: null });
            store.dispatch({ type: 'auth/setUser', payload: null });
            window.location.href = "/login";
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
    return this.client.get("/vehicles", { params });
  }

  async getVehicle(id: string): Promise<any> {
    return this.client.get(`/admin/vehicles/${id}`);
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

  // Trips
  async getTrips(params?: any): Promise<any> {
    return this.client.get("/dispatch/trips", { params });
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
  async getUsers(orgId: string): Promise<any> {
    return this.client.get("/admin/users", { params: { org_id: orgId } });
  }

  // Organizations
  async getOrganizations(params?: any): Promise<any> {
    return this.client.get("/admin/organizations", { params });
  }

  async getOrganization(id: string): Promise<any> {
    return this.client.get(`/admin/organizations/${id}`);
  }

  async createOrganization(data: any): Promise<any> {
    return this.client.post("/admin/organizations", data);
  }

  async updateOrganization(id: string, data: any): Promise<any> {
    return this.client.put(`/admin/organizations/${id}`, data);
  }

  async deleteOrganization(id: string) {
    await this.client.delete(`/admin/organizations/${id}`);
  }

  async getOrganizationStats(): Promise<any> {
    // Try to get stats, fallback to empty/default if backend doesn't support it yet
    try {
      return await this.client.get("/admin/organizations/stats");
    } catch (error) {
      console.warn("Organization stats API not available, using fallback");
      return {
        data: {
          total_orgs: 0,
          active_orgs: 0,
          pending_orgs: 0,
          inactive_orgs: 0,
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
    return this.client.get("/drivers", { params });
  }

  async getDriver(id: string): Promise<any> {
    return this.client.get(`/admin/drivers/${id}`);
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
    return this.bffClient.get("/api/v1/dashboard/stats");
  }

  async getAdminDashboardData(timeRange?: string): Promise<any> {
    return this.bffClient.get("/api/v1/dashboard/data", {
      params: { time_range: timeRange },
    });
  }

  async getTimeline(): Promise<any> {
    return this.bffClient.get("/api/v1/timeline");
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
  async getMFADashboard(orgId: string): Promise<any> {
    return this.bffClient.get("/api/v1/security/mfa/dashboard", {
      params: { org_id: orgId },
    });
  }

  async getMFAAdoptionTrend(orgId: string, days: number = 30): Promise<any> {
    return this.bffClient.get("/api/v1/security/mfa/adoption-trend", {
      params: { org_id: orgId, days },
    });
  }

  async getMFAMethodDistribution(orgId: string): Promise<any> {
    return this.bffClient.get("/api/v1/security/mfa/method-distribution", {
      params: { org_id: orgId },
    });
  }

  async getMFAEvents(
    orgId: string,
    page: number = 1,
    pageSize: number = 20,
  ): Promise<any> {
    return this.bffClient.get("/api/v1/security/mfa/events", {
      params: { org_id: orgId, page, page_size: pageSize },
    });
  }

  async getMFAFailureLocations(orgId: string): Promise<any> {
    return this.bffClient.get("/api/v1/security/mfa/failures/locations", {
      params: { org_id: orgId },
    });
  }

  async getMFAFailureUsers(orgId: string): Promise<any> {
    return this.bffClient.get("/api/v1/security/mfa/failures/users", {
      params: { org_id: orgId },
    });
  }

  async getLockedUsers(orgId: string): Promise<any> {
    return this.bffClient.get("/api/v1/security/mfa/locked-users", {
      params: { org_id: orgId },
    });
  }

  async unlockUser(userId: string): Promise<any> {
    return this.bffClient.post("/api/v1/security/mfa/unlock", { userId });
  }

  async resetMFA(userId: string): Promise<any> {
    return this.bffClient.post("/api/v1/security/mfa/reset", { userId });
  }

  async getMFAPolicy(orgId: string): Promise<any> {
    return this.bffClient.get("/api/v1/security/mfa/policy", {
      params: { org_id: orgId },
    });
  }

  async updateMFAPolicy(data: { policy: any; reason: string }): Promise<any> {
    return this.bffClient.put("/api/v1/security/mfa/policy", data);
  }

  async getMFAPolicyHistory(orgId: string): Promise<any> {
    return this.bffClient.get("/api/v1/security/mfa/policy/history", {
      params: { org_id: orgId },
    });
  }

  async rollbackMFAPolicy(historyId: string): Promise<any> {
    return this.bffClient.post("/api/v1/security/mfa/policy/rollback", {
      historyId,
    });
  }

  async getMFAOrgOverride(orgId: string): Promise<any> {
    return this.bffClient.get("/api/v1/security/mfa/overrides", {
      params: { org_id: orgId },
    });
  }

  async updateMFAOrgOverride(data: any): Promise<any> {
    return this.bffClient.put("/api/v1/security/mfa/overrides", data);
  }

  // Audit Module
  async getAuditLogs(params: {
    org_id: string;
    page?: number;
    action?: string;
    severity?: string;
    user_id?: string;
  }): Promise<any> {
    return this.bffClient.get("/api/v1/audit/logs", { params });
  }

  async getAuditLogDetails(id: string, org_id: string): Promise<any> {
    return this.bffClient.get(`/api/v1/audit/logs/${id}`, {
      params: { org_id },
    });
  }

  async getAuditSettings(orgId: string): Promise<any> {
    return this.bffClient.get("/api/v1/audit/settings", {
      params: { org_id: orgId },
    });
  }

  async updateAuditSettings(data: any): Promise<any> {
    return this.bffClient.put("/api/v1/audit/settings", data);
  }

  async getMFAExceptions(orgId: string): Promise<any> {
    return this.bffClient.get("/api/v1/security/mfa/exceptions", {
      params: { org_id: orgId },
    });
  }

  async addMFAException(data: {
    userId: string;
    orgId: string;
    reason: string;
    days: number;
  }): Promise<any> {
    return this.bffClient.post("/api/v1/security/mfa/exceptions", data);
  }

  async deleteMFAException(id: string): Promise<any> {
    return this.bffClient.delete(`/api/v1/security/mfa/exceptions/${id}`);
  }

  // Security Operations Expansion
  async getLoginActivityLogs(params: {
    page?: number;
    size?: number;
    status?: string;
  }): Promise<any> {
    return this.bffClient.get("/api/v1/security/logins", { params });
  }

  async getThreatAlerts(params: {
    page?: number;
    size?: number;
    severity?: string;
    status?: string;
  }): Promise<any> {
    return this.bffClient.get("/api/v1/security/alerts", { params });
  }

  async updateThreatAlertStatus(id: string, status: string): Promise<any> {
    return this.bffClient.patch(`/api/v1/security/alerts/${id}/status`, {
      status,
    });
  }

  async getSecuritySessions(orgId: string): Promise<any> {
    return this.bffClient.get("/api/v1/security/sessions", {
      params: { org_id: orgId },
    });
  }

  async revokeSecuritySession(sessionId: string): Promise<any> {
    return this.bffClient.delete(`/api/v1/security/sessions/${sessionId}`);
  }

  async getIPRules(orgId: string): Promise<any> {
    return this.bffClient.get("/api/v1/security/ip-rules", {
      params: { org_id: orgId },
    });
  }

  async createIPRule(data: any): Promise<any> {
    return this.bffClient.post("/api/v1/security/ip-rules", data);
  }

  // Audit & Compliance Extensions
  async getEntityHistory(type: string, id: string): Promise<any> {
    return this.bffClient.get("/api/v1/audit/entity-history", {
      params: { type, id },
    });
  }

  async createEvidenceExport(data: {
    org_id: string;
    module: string;
    date_range_start: string;
    date_range_end: string;
    format: string;
  }): Promise<any> {
    return this.bffClient.post("/api/v1/audit/evidence-exports", data);
  }

  async getEvidenceExports(orgId: string): Promise<any> {
    return this.bffClient.get("/api/v1/audit/evidence-exports", {
      params: { org_id: orgId },
    });
  }

  async getEvidenceExportDetails(id: string): Promise<any> {
    return this.bffClient.get(`/api/v1/audit/evidence-exports/${id}`);
  }

  async getAuditPolicy(orgId: string): Promise<any> {
    return this.bffClient.get("/api/v1/audit/policy", {
      params: { org_id: orgId },
    });
  }

  async updateAuditPolicy(data: any): Promise<any> {
    return this.bffClient.put("/api/v1/audit/policy", data);
  }

  // Integrations Module
  async getIntegrations(orgId: string): Promise<any> {
    return this.bffClient.get("/api/v1/integrations/", {
      params: { org_id: orgId },
    });
  }

  async createIntegration(data: any): Promise<any> {
    return this.bffClient.post("/api/v1/integrations/", data);
  }

  async updateIntegration(id: string, data: any): Promise<any> {
    return this.bffClient.put(`/api/v1/integrations/${id}`, data);
  }

  async testIntegration(id: string): Promise<any> {
    return this.bffClient.post(`/api/v1/integrations/${id}/test`);
  }

  async getApiClients(orgId: string): Promise<any> {
    return this.bffClient.get("/api/v1/api-clients/", {
      params: { org_id: orgId },
    });
  }

  async createApiClient(data: any): Promise<any> {
    return this.bffClient.post("/api/v1/api-clients/", data);
  }

  async getWebhooks(orgId: string): Promise<any> {
    return this.bffClient.get("/api/v1/webhooks/", {
      params: { org_id: orgId },
    });
  }

  async getIntegrationLogs(params: {
    org_id: string;
    integration_id?: string;
    limit?: number;
    offset?: number;
  }): Promise<any> {
    return this.bffClient.get("/api/v1/integrations/logs", { params });
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
