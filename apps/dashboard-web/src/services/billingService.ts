import apiClient from './apiClient';
import { BillingSettings, BillingMetrics, Invoice } from '../types/billing';

// Note: apiClient interceptors unwrap .data automatically,
// so all methods return the response body directly (not AxiosResponse).

const billingService = {
  getInvoices: async (page = 1, size = 10) => {
    return apiClient.get<any, { data: any[]; total: number; page: number; size: number }>('/admin/billing/invoices', {
      params: { page, size },
    });
  },

  getInvoice: async (id: string) => {
    return apiClient.get<any, Invoice>(`/admin/billing/invoices/${id}`);
  },

  createInvoice: async (invoice: Partial<Invoice>) => {
    return apiClient.post<any, Invoice>('/admin/billing/invoices', invoice);
  },

  updateInvoice: async (id: string, invoice: Partial<Invoice>) => {
    return apiClient.put<any, Invoice>(`/admin/billing/invoices/${id}`, invoice);
  },

  deleteInvoice: async (id: string) => {
    return apiClient.delete(`/admin/billing/invoices/${id}`);
  },

  recordPayment: async (id: string, amount: number) => {
    return apiClient.post(`/admin/billing/invoices/${id}/payments`, { amount });
  },

  downloadPDF: async (id: string): Promise<Blob> => {
    const response = await apiClient.get(`/admin/billing/invoices/${id}/pdf`, {
      responseType: 'blob',
    });
    return response as unknown as Blob;
  },

  emailInvoice: async (id: string, email: string, customerName: string) => {
    return apiClient.post(`/admin/billing/invoices/${id}/email`, {
      email,
      customer_name: customerName,
    });
  },

  getSettings: async () => {
    return apiClient.get<any, BillingSettings>('/admin/billing/settings');
  },

  updateSettings: async (settings: Partial<BillingSettings>) => {
    return apiClient.put('/admin/billing/settings', settings);
  },

  getMetrics: async () => {
    return apiClient.get<any, BillingMetrics>('/admin/billing/metrics');
  },

  getHealth: async () => {
    return apiClient.get('/admin/billing/health');
  },

  getStats: async () => {
    return apiClient.get('/admin/billing/stats');
  },

  generateDraftInvoice: async () => {
    return apiClient.post<any, Invoice>('/admin/billing/invoices/generate-draft');
  },

  getUsageSummary: async () => {
    return apiClient.get<any, any[]>('/admin/billing/usage/summary');
  },

  getPlans: async () => {
    return apiClient.get<any, any[]>('/admin/billing/plans');
  },

  getSubscription: async () => {
    return apiClient.get<any, any>('/admin/billing/subscription');
  },

  getGateLogs: async (limit = 50, offset = 0) => {
    return apiClient.get<any, any[]>('/admin/saas/gate-logs', {
      params: { limit, offset },
    });
  },

  getFintechTransactions: async (limit = 50, offset = 0) => {
    return apiClient.get<any, any[]>('/admin/saas/fintech-transactions', {
      params: { limit, offset },
    });
  },
};

export default billingService;
