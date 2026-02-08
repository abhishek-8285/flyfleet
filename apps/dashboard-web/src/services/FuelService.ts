import {
    CreateFuelLogRequest,
    FuelLog,
    FuelLogFilter,
    FuelLogStats,
    FuelSettings,
    UpdateFuelLogRequest
} from '../types/fuel';
import apiClient from './apiClient';

class FuelService {
  async listLogs(filter: FuelLogFilter = {}) {
    const params = new URLSearchParams();
    if (filter.vehicle_id) params.append('vehicle_id', filter.vehicle_id);
    if (filter.driver_id) params.append('driver_id', filter.driver_id);
    if (filter.search) params.append('search', filter.search);
    if (filter.page) params.append('page', filter.page.toString());
    if (filter.page_size) params.append('page_size', filter.page_size.toString());
    filter.status?.forEach(s => params.append('status', s));

    const response = await apiClient.get<{ logs: FuelLog[], total: number }>(`/api/v1/admin/fuel-logs?${params.toString()}`);
    return response.data;
  }

  async getLog(id: string) {
    const response = await apiClient.get<FuelLog>(`/api/v1/admin/fuel-logs/${id}`);
    return response.data;
  }

  async createLog(data: CreateFuelLogRequest) {
    const response = await apiClient.post<FuelLog>('/api/v1/admin/fuel-logs', data);
    return response.data;
  }

  async updateLog(id: string, data: UpdateFuelLogRequest) {
    const response = await apiClient.put<FuelLog>(`/api/v1/admin/fuel-logs/${id}`, data);
    return response.data;
  }

  async deleteLog(id: string) {
    await apiClient.delete(`/api/v1/admin/fuel-logs/${id}`);
  }

  async changeStatus(id: string, status: 'APPROVED' | 'REJECTED', notes?: string) {
    await apiClient.patch(`/api/v1/admin/fuel-logs/${id}/status`, { status, notes });
  }

  async getStats() {
    const response = await apiClient.get<FuelLogStats>('/api/v1/admin/fuel/stats');
    return response.data;
  }

  async getSettings() {
    const response = await apiClient.get<FuelSettings>('/api/v1/admin/fuel');
    return response.data;
  }

  async updateSettings(data: Partial<FuelSettings>) {
    const response = await apiClient.put<FuelSettings>('/api/v1/admin/fuel', data);
    return response.data;
  }
}

const fuelService = new FuelService();
export default fuelService;
