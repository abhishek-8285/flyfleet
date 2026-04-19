import apiClient from './apiClient';
import { Invoice } from '../types/billing';

export interface CreateInvoiceItem {
  description: string;
  quantity: number;
  unit_price: number;
  tax_percent?: number;
}

export interface CreateInvoiceRequest {
  customer_id?: string;
  customer_name: string;
  customer_email?: string;
  customer_phone: string; // Mandatory as per requirement
  billing_address: string;
  due_date: string;
  currency: string;
  items: CreateInvoiceItem[];
  notes?: string;
}

export const invoicingV2Service = {
  createInvoice: async (data: CreateInvoiceRequest): Promise<Invoice> => {
    try {
      const response = await apiClient.post('/admin/invoicing', data);
      return response.data?.invoice;
    } catch (error) {
      console.error('[invoicingV2Service] createInvoice failed:', error);
      throw error;
    }
  },

  listInvoices: async (limit: number = 20, offset: number = 0): Promise<{ invoices: Invoice[], total: number }> => {
    const response = await apiClient.get('/admin/invoicing', {
      params: { limit, offset }
    });
    return {
      invoices: response.data?.invoices || [],
      total: response.data?.total || 0
    };
  },

  getInvoice: async (id: string): Promise<Invoice> => {
    const response = await apiClient.get(`/admin/invoicing/${id}`);
    return response.data?.invoice;
  },

  voidInvoice: async (id: string): Promise<void> => {
    await apiClient.post(`/admin/invoicing/${id}/void`);
  }
};
