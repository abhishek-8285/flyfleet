import axios from 'axios';
import {
    CreateGovernanceRequest,
    GovernanceRequest,
    GovernanceRequestFilter,
    GovernanceRequestListResponse,
    GovernanceSettings
} from '../types/governance';

class GovernanceService {
  private readonly baseUrl = '/api/v1/admin/governance';

  async listRequests(filter: GovernanceRequestFilter): Promise<GovernanceRequestListResponse> {
    const params = new URLSearchParams();
    if (filter.type) params.append('type', filter.type);
    if (filter.status) filter.status.forEach(s => params.append('status', s));
    if (filter.search) params.append('search', filter.search);
    if (filter.page) params.append('page', filter.page.toString());
    if (filter.page_size) params.append('page_size', filter.page_size.toString());

    const response = await axios.get(`${this.baseUrl}/requests`, { params });
    return response.data;
  }

  async getRequest(id: string): Promise<GovernanceRequest> {
    const response = await axios.get(`${this.baseUrl}/requests/${id}`);
    return response.data;
  }

  async createRequest(request: CreateGovernanceRequest): Promise<GovernanceRequest> {
    const response = await axios.post(`${this.baseUrl}/requests`, request);
    return response.data;
  }

  async approveRequest(id: string, notes?: string): Promise<void> {
    await axios.post(`${this.baseUrl}/requests/${id}/approve`, { notes });
  }

  async rejectRequest(id: string, notes?: string): Promise<void> {
    await axios.post(`${this.baseUrl}/requests/${id}/reject`, { notes });
  }

  async getSettings(): Promise<GovernanceSettings> {
    const response = await axios.get(`${this.baseUrl}/settings`);
    return response.data;
  }

  async updateSettings(settings: Partial<GovernanceSettings>): Promise<GovernanceSettings> {
    const response = await axios.put(`${this.baseUrl}/settings`, settings);
    return response.data;
  }
}

const governanceService = new GovernanceService();
export default governanceService;
