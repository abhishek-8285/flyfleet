import axios, { AxiosError, AxiosInstance } from 'axios';

const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';

class ApiService {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: API_URL,
      timeout: 10000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Request interceptor
    this.client.interceptors.request.use(
      (config) => {
        const token = localStorage.getItem('token');
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => Promise.reject(error)
    );

    // Response interceptor
    this.client.interceptors.response.use(
      (response) => response,
      (error: AxiosError) => {
        if (error.response?.status === 401) {
          localStorage.removeItem('token');
          window.location.href = '/login';
        }
        return Promise.reject(error);
      }
    );
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
    const response = await this.client.get('/admin/alerts/stats/summary');
    return response.data;
  }

  async getDriverStats() {
    const response = await this.client.get('/admin/analytics/drivers');
    return response.data;
  }

  async getTripStats() {
    const response = await this.client.get('/admin/analytics/trips');
    return response.data;
  }

  // MFA Security
  async getMFADashboard(orgId: string) {
    const response = await this.client.get('/security/mfa/dashboard', { params: { org_id: orgId } });
    return response.data;
  }

  async getMFAAdoptionTrend(orgId: string, days: number = 30) {
    const response = await this.client.get('/security/mfa/adoption-trend', { params: { org_id: orgId, days } });
    return response.data;
  }

  async getMFAMethodDistribution(orgId: string) {
    const response = await this.client.get('/security/mfa/method-distribution', { params: { org_id: orgId } });
    return response.data;
  }

  async getMFAEvents(orgId: string, page: number = 1, pageSize: number = 20) {
    const response = await this.client.get('/security/mfa/events', { params: { org_id: orgId, page, page_size: pageSize } });
    return response.data;
  }

  async getMFAFailureLocations(orgId: string) {
    const response = await this.client.get('/security/mfa/failures/locations', { params: { org_id: orgId } });
    return response.data;
  }

  async getMFAFailureUsers(orgId: string) {
    const response = await this.client.get('/security/mfa/failures/users', { params: { org_id: orgId } });
    return response.data;
  }

  async getLockedUsers(orgId: string) {
    const response = await this.client.get('/security/mfa/locked-users', { params: { org_id: orgId } });
    return response.data;
  }

  async unlockUser(userId: string) {
    const response = await this.client.post('/security/mfa/unlock', { userId });
    return response.data;
  }

  async resetMFA(userId: string) {
    const response = await this.client.post('/security/mfa/reset', { userId });
    return response.data;
  }

  async getMFAPolicy(orgId: string) {
    const response = await this.client.get('/security/mfa/policy', { params: { org_id: orgId } });
    return response.data;
  }

  async updateMFAPolicy(policy: any) {
    const response = await this.client.put('/security/mfa/policy', policy);
    return response.data;
  }

  async getMFAExceptions(orgId: string) {
    const response = await this.client.get('/security/mfa/exceptions', { params: { org_id: orgId } });
    return response.data;
  }

  async addMFAException(data: { userId: string, orgId: string, reason: string, days: number }) {
    const response = await this.client.post('/security/mfa/exceptions', data);
    return response.data;
  }

  async deleteMFAException(id: string) {
    await this.client.delete(`/security/mfa/exceptions/${id}`);
  }
}

export const api = new ApiService();
export default api;
