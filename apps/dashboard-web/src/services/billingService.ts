import apiClient from './apiClient';
import { BillingSettings, Invoice } from '../types/billing';

const billingService = {
  getInvoices: async (page = 1, size = 10) => {
    const response = await apiClient.get('/billing/invoices', {
      params: { page, size },
    });
    return response;
  },

  getInvoice: async (id: string) => {
    const response = await apiClient.get(`/billing/invoices/${id}`);
    return response;
  },

  createInvoice: async (invoice: Partial<Invoice>) => {
    const response = await apiClient.post('/billing/invoices', invoice);
    return response;
  },

  updateInvoice: async (id: string, invoice: Partial<Invoice>) => {
    const response = await apiClient.put(`/billing/invoices/${id}`, invoice);
    return response;
  },

  deleteInvoice: async (id: string) => {
    await apiClient.delete(`/billing/invoices/${id}`);
  },

  recordPayment: async (id: string, amount: number) => {
    const response = await apiClient.post(`/billing/invoices/${id}/payments`, { amount });
    return response;
  },

  downloadPDF: async (id: string): Promise<Blob> => {
    const response = await apiClient.get(`/billing/invoices/${id}/pdf`, {
      responseType: 'blob',
    });
    return response as unknown as Blob;
  },

  emailInvoice: async (id: string, email: string, customerName: string) => {
    const response = await apiClient.post(`/billing/invoices/${id}/email`, {
      email,
      customer_name: customerName,
    });
    return response;
  },

  getSettings: async () => {
    const response = await apiClient.get('/billing/settings');
    return response;
  },

  updateSettings: async (settings: Partial<BillingSettings>) => {
    const response = await apiClient.put('/billing/settings', settings);
    return response;
  },

  getMetrics: async () => {
    const response = await apiClient.get('/billing/metrics');
    return response;
  },

  getHealth: async () => {
    const response = await apiClient.get('/billing/health');
    return response;
  },

  getStats: async () => {
    const response = await apiClient.get('/billing/stats');
    return response;
  },
};

export default billingService;
