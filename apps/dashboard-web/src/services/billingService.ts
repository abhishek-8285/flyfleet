import apiClient from './apiClient';
import { BillingSettings, BillingMetrics, Invoice } from '../types/billing';

// Note: apiClient interceptors unwrap .data automatically,
// so all methods return the response body directly (not AxiosResponse).

const billingService = {
  getInvoices: async (page = 1, size = 10) => {
    return apiClient.get<any, { data: any[]; total: number; page: number; size: number }>('/billing/invoices', {
      params: { page, size },
    });
  },

  getInvoice: async (id: string) => {
    return apiClient.get<any, Invoice>(`/billing/invoices/${id}`);
  },

  createInvoice: async (invoice: Partial<Invoice>) => {
    return apiClient.post<any, Invoice>('/billing/invoices', invoice);
  },

  updateInvoice: async (id: string, invoice: Partial<Invoice>) => {
    return apiClient.put<any, Invoice>(`/billing/invoices/${id}`, invoice);
  },

  deleteInvoice: async (id: string) => {
    return apiClient.delete(`/billing/invoices/${id}`);
  },

  recordPayment: async (id: string, amount: number) => {
    return apiClient.post(`/billing/invoices/${id}/payments`, { amount });
  },

  downloadPDF: async (id: string): Promise<Blob> => {
    const response = await apiClient.get(`/billing/invoices/${id}/pdf`, {
      responseType: 'blob',
    });
    return response as unknown as Blob;
  },

  emailInvoice: async (id: string, email: string, customerName: string) => {
    return apiClient.post(`/billing/invoices/${id}/email`, {
      email,
      customer_name: customerName,
    });
  },

  getSettings: async () => {
    return apiClient.get<any, BillingSettings>('/billing/settings');
  },

  updateSettings: async (settings: Partial<BillingSettings>) => {
    return apiClient.put('/billing/settings', settings);
  },

  getMetrics: async () => {
    return apiClient.get<any, BillingMetrics>('/billing/metrics');
  },

  getHealth: async () => {
    return apiClient.get('/billing/health');
  },

  getStats: async () => {
    return apiClient.get('/billing/stats');
  },
};

export default billingService;
