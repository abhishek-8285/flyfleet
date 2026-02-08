import axios from 'axios';
import { BillingSettings, Invoice } from '../types/billing';

const API_BASE = '/api/v1/admin';

const billingService = {
  getInvoices: async (page = 1, size = 10) => {
    const response = await axios.get(`${API_BASE}/billing/invoices`, {
      params: { page, size },
    });
    return response.data;
  },

  getInvoice: async (id: string) => {
    const response = await axios.get(`${API_BASE}/billing/invoices/${id}`);
    return response.data;
  },

  createInvoice: async (invoice: Partial<Invoice>) => {
    const response = await axios.post(`${API_BASE}/billing/invoices`, invoice);
    return response.data;
  },

  updateInvoice: async (id: string, invoice: Partial<Invoice>) => {
    const response = await axios.put(`${API_BASE}/billing/invoices/${id}`, invoice);
    return response.data;
  },

  deleteInvoice: async (id: string) => {
    await axios.delete(`${API_BASE}/billing/invoices/${id}`);
  },

  recordPayment: async (id: string, amount: number) => {
    const response = await axios.post(`${API_BASE}/billing/invoices/${id}/payments`, { amount });
    return response.data;
  },

  getSettings: async () => {
    const response = await axios.get(`${API_BASE}/billing-settings`);
    return response.data;
  },

  updateSettings: async (settings: Partial<BillingSettings>) => {
    const response = await axios.put(`${API_BASE}/billing-settings`, settings);
    return response.data;
  },

  getMetrics: async () => {
    const response = await axios.get(`${API_BASE}/billing-metrics`);
    return response.data;
  },
};

export default billingService;
