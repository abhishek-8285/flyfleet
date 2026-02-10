import axios, { AxiosError, AxiosInstance } from 'axios';

const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:8081/api/v1';
const BFF_URL = process.env.REACT_APP_BFF_URL || 'http://localhost:8081';

class ApiService {
  private client: AxiosInstance;
  private bffClient: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: API_URL,
      timeout: 10000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    this.bffClient = axios.create({
      baseURL: BFF_URL,
      timeout: 10000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    const setupInterceptors = (instance: AxiosInstance) => {
      instance.interceptors.request.use(
        (config) => {
          const token = localStorage.getItem('access_token');
          if (token) {
            config.headers.Authorization = `Bearer ${token}`;
          }
          return config;
        },
        (error) => Promise.reject(error)
      );

      instance.interceptors.response.use(
        (response) => response,
        (error: AxiosError) => {
          if (error.response?.status === 401) {
            localStorage.removeItem('token');
            window.location.href = '/login';
          }
          return Promise.reject(error);
        }
      );
    };

    setupInterceptors(this.client);
    setupInterceptors(this.bffClient);
  }

  // Auth
  async login(email: string, password: string) {
    const response = await this.client.post('/auth/login', { email, password });
    return response.data;
  }

  async refreshToken() {
    const response = await this.client.post('/auth/refresh');
    return response.data;
  }

  async logout() {
    await this.client.post('/auth/logout');
    localStorage.removeItem('token');
  }

  // Vehicles
  async getVehicles(params?: any) {
    const response = await this.client.get('/admin/vehicles', { params });
    return response.data;
  }

  async getVehicle(id: string) {
    const response = await this.client.get(`/admin/vehicles/${id}`);
    return response.data;
  }

  async createVehicle(data: any) {
    const response = await this.client.post('/admin/vehicles', data);
    return response.data;
  }

  async updateVehicle(id: string, data: any) {
    const response = await this.client.put(`/admin/vehicles/${id}`, data);
    return response.data;
  }

  async deleteVehicle(id: string) {
    await this.client.delete(`/admin/vehicles/${id}`);
  }

  // Trips
  async getTrips(params?: any) {
    const response = await this.client.get('/admin/trips', { params });
    return response.data;
  }

  async getTrip(id: string) {
    const response = await this.client.get(`/admin/trips/${id}`);
    return response.data;
  }

  async createTrip(data: any) {
    const response = await this.client.post('/admin/trips', data);
    return response.data;
  }

  async updateTrip(id: string, data: any) {
    const response = await this.client.put(`/admin/trips/${id}`, data);
    return response.data;
  }

  async deleteTrip(id: string) {
    const response = await this.client.delete(`/admin/trips/${id}`);
    return response.data;
  }

  async getTripTelemetry(id: string, limit: number = 50) {
    const response = await this.client.get(`/admin/trips/${id}/telemetry`, { params: { limit } });
    return response.data;
  }

  async getTripSettings() {
    const response = await this.client.get('/admin/trips/settings');
    return response.data;
  }

  async updateTripSettings(data: any) {
    const response = await this.client.put('/admin/trips/settings', data);
    return response.data;
  }

  // Users
  async getUsers(orgId: string) {
    const response = await this.client.get('/admin/users', { params: { org_id: orgId } });
    return response.data;
  }

  // Organizations
  async getOrganizations() {
    const response = await this.client.get('/admin/organizations');
    return response.data;
  }

  async getOrganizationStats() {
    const response = await this.bffClient.get('/api/v1/admin/organizations/stats');
    return response.data.data;
  }

  // Drivers
  async getDrivers(params?: any) {
    const response = await this.client.get('/admin/drivers', { params });
    return response.data;
  }

  async getDriver(id: string) {
    const response = await this.client.get(`/admin/drivers/${id}`);
    return response.data;
  }

  // Alerts
  async getAlerts(params?: any) {
    const response = await this.client.get('/admin/alerts', { params });
    return response.data;
  }

  async acknowledgeAlert(id: string) {
    const response = await this.client.post(`/admin/alerts/${id}/acknowledge`);
    return response.data;
  }

  // Geofences
  async getGeofences() {
    const response = await this.client.get('/admin/geofences');
    return response.data;
  }

  async createGeofence(data: any) {
    const response = await this.client.post('/admin/geofences', data);
    return response.data;
  }

  async deleteGeofence(id: string) {
    await this.client.delete(`/admin/geofences/${id}`);
  }

  // Statistics
  async getStatistics() {
    const response = await this.bffClient.get('/api/v1/dashboard/stats');
    return response.data;
  }

  // Documents
  async getDocuments(vehicleId: string) {
    const response = await this.client.get('/admin/documents', { params: { vehicle_id: vehicleId } });
    return response.data;
  }

  async createDocument(vehicleId: string, data: any) {
    const response = await this.client.post('/admin/documents', data, { params: { vehicle_id: vehicleId } });
    return response.data;
  }

  async updateDocument(id: string, data: any) {
    const response = await this.client.put(`/admin/documents/${id}`, data);
    return response.data;
  }

  async deleteDocument(id: string) {
    await this.client.delete(`/admin/documents/${id}`);
  }

  async getDriverStats() {
    const response = await this.client.get('/admin/analytics/drivers');
    return response.data;
  }

  async getTripStats() {
    const response = await this.client.get('/admin/analytics/trips');
    return response.data;
  }

  // Bulk Uploads
  async getBulkJobs(params?: any) {
    const response = await this.client.get('/admin/bulk-jobs', { params });
    return response.data;
  }

  async createBulkJob(file: File) {
    const formData = new FormData();
    formData.append('file', file);
    const response = await this.client.post('/admin/bulk-jobs', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  }

  async getBulkJob(id: string) {
    const response = await this.client.get(`/admin/bulk-jobs/${id}`);
    return response.data;
  }

  async getBulkJobRecords(id: string, params?: any) {
    const response = await this.client.get(`/admin/bulk-jobs/${id}/records`, { params });
    return response.data;
  }

  // MFA Security (Routed through BFF)
  async getMFADashboard(orgId: string) {
    const response = await this.bffClient.get('/api/mfa/dashboard', { params: { org_id: orgId } });
    return response.data;
  }

  async getMFAAdoptionTrend(orgId: string, days: number = 30) {
    const response = await this.bffClient.get('/api/mfa/adoption-trend', { params: { org_id: orgId, days } });
    return response.data;
  }

  async getMFAMethodDistribution(orgId: string) {
    const response = await this.bffClient.get('/api/mfa/method-distribution', { params: { org_id: orgId } });
    return response.data;
  }

  async getMFAEvents(orgId: string, page: number = 1, pageSize: number = 20) {
    const response = await this.bffClient.get('/api/mfa/events', { params: { org_id: orgId, page, page_size: pageSize } });
    return response.data;
  }

  async getMFAFailureLocations(orgId: string) {
    const response = await this.bffClient.get('/api/mfa/failures/locations', { params: { org_id: orgId } });
    return response.data;
  }

  async getMFAFailureUsers(orgId: string) {
    const response = await this.bffClient.get('/api/mfa/failures/users', { params: { org_id: orgId } });
    return response.data;
  }

  async getLockedUsers(orgId: string) {
    const response = await this.bffClient.get('/api/mfa/locked-users', { params: { org_id: orgId } });
    return response.data;
  }

  async unlockUser(userId: string) {
    const response = await this.bffClient.post('/api/mfa/unlock', { userId });
    return response.data;
  }

  async resetMFA(userId: string) {
    const response = await this.bffClient.post('/api/mfa/reset', { userId });
    return response.data;
  }

  async getMFAPolicy(orgId: string) {
    const response = await this.bffClient.get('/api/mfa/policy', { params: { org_id: orgId } });
    return response.data;
  }

  async updateMFAPolicy(data: { policy: any, reason: string }) {
    const response = await this.bffClient.put('/api/mfa/policy', data);
    return response.data;
  }

  async getMFAPolicyHistory(orgId: string) {
    const response = await this.bffClient.get('/api/mfa/policy/history', { params: { org_id: orgId } });
    return response.data;
  }

  async rollbackMFAPolicy(historyId: string) {
    const response = await this.bffClient.post('/api/mfa/policy/rollback', { historyId });
    return response.data;
  }

  async getMFAOrgOverride(orgId: string) {
    const response = await this.bffClient.get('/api/mfa/overrides', { params: { org_id: orgId } });
    return response.data;
  }

  async updateMFAOrgOverride(data: any) {
    const response = await this.bffClient.put('/api/mfa/overrides', data);
    return response.data;
  }

  // Audit Module
  async getAuditLogs(params: { org_id: string; page?: number; action?: string; severity?: string; user_id?: string }) {
    const response = await this.bffClient.get('/api/audit/logs', { params });
    return response.data;
  }

  async getAuditLogDetails(id: string, org_id: string) {
    const response = await this.bffClient.get(`/api/audit/logs/${id}`, { params: { org_id } });
    return response.data;
  }

  async getAuditSettings(orgId: string) {
    const response = await this.bffClient.get('/api/audit/settings', { params: { org_id: orgId } });
    return response.data;
  }

  async updateAuditSettings(data: any) {
    const response = await this.bffClient.put('/api/audit/settings', data);
    return response.data;
  }

  async getMFAExceptions(orgId: string) {
    const response = await this.bffClient.get('/api/mfa/exceptions', { params: { org_id: orgId } });
    return response.data;
  }

  async addMFAException(data: { userId: string, orgId: string, reason: string, days: number }) {
    const response = await this.bffClient.post('/api/mfa/exceptions', data);
    return response.data;
  }

  async deleteMFAException(id: string) {
    await this.bffClient.delete(`/api/mfa/exceptions/${id}`);
  }

  // Security Operations Expansion
  async getLoginActivityLogs(params: { page?: number; size?: number; status?: string }) {
    const response = await this.bffClient.get('/api/security/logins', { params });
    return response.data;
  }

  async getThreatAlerts(params: { page?: number; size?: number; severity?: string; status?: string }) {
    const response = await this.bffClient.get('/api/security/alerts', { params });
    return response.data;
  }

  async updateThreatAlertStatus(id: string, status: string) {
    const response = await this.bffClient.patch(`/api/security/alerts/${id}/status`, { status });
    return response.data;
  }

  async getSecuritySessions(orgId: string) {
    const response = await this.bffClient.get('/api/security/sessions', { params: { org_id: orgId } });
    return response.data;
  }

  async revokeSecuritySession(sessionId: string) {
    const response = await this.bffClient.delete(`/api/security/sessions/${sessionId}`);
    return response.data;
  }

  async getIPRules(orgId: string) {
    const response = await this.bffClient.get('/api/security/ip-rules', { params: { org_id: orgId } });
    return response.data;
  }

  async createIPRule(data: any) {
    const response = await this.bffClient.post('/api/security/ip-rules', data);
    return response.data;
  }

  // Audit & Compliance Extensions
  async getEntityHistory(type: string, id: string) {
    const response = await this.bffClient.get('/api/audit/entity-history', { params: { type, id } });
    return response.data;
  }

  async createEvidenceExport(data: { org_id: string; module: string; date_range_start: string; date_range_end: string; format: string }) {
    const response = await this.bffClient.post('/api/audit/evidence-exports', data);
    return response.data;
  }

  async getEvidenceExports(orgId: string) {
    const response = await this.bffClient.get('/api/audit/evidence-exports', { params: { org_id: orgId } });
    return response.data;
  }

  async getEvidenceExportDetails(id: string) {
    const response = await this.bffClient.get(`/api/audit/evidence-exports/${id}`);
    return response.data;
  }

  async getAuditPolicy(orgId: string) {
    const response = await this.bffClient.get('/api/audit/policy', { params: { org_id: orgId } });
    return response.data;
  }

  async updateAuditPolicy(data: any) {
    const response = await this.bffClient.put('/api/audit/policy', data);
    return response.data;
  }

  // Integrations Module
  async getIntegrations(orgId: string) {
    const response = await this.bffClient.get('/api/integrations/', { params: { org_id: orgId } });
    return response.data;
  }

  async createIntegration(data: any) {
    const response = await this.bffClient.post('/api/integrations/', data);
    return response.data;
  }

  async updateIntegration(id: string, data: any) {
    const response = await this.bffClient.put(`/api/integrations/${id}`, data);
    return response.data;
  }

  async testIntegration(id: string) {
    const response = await this.bffClient.post(`/api/integrations/${id}/test`);
    return response.data;
  }

  async getApiClients(orgId: string) {
    const response = await this.bffClient.get('/api/api-clients/', { params: { org_id: orgId } });
    return response.data;
  }

  async createApiClient(data: any) {
    const response = await this.bffClient.post('/api/api-clients/', data);
    return response.data;
  }

  async getWebhooks(orgId: string) {
    const response = await this.bffClient.get('/api/webhooks/', { params: { org_id: orgId } });
    return response.data;
  }

  async getIntegrationLogs(params: { org_id: string; integration_id?: string; limit?: number; offset?: number }) {
    const response = await this.bffClient.get('/api/integrations/logs', { params });
    return response.data;
  }
}

export const api = new ApiService();
export default api;
