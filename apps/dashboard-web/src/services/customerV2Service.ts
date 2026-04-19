import apiClient from './apiClient';

export interface Customer {
  id: string;
  name: string;
  email?: string;
  phone?: string;
  address?: string;
  gstin?: string;
}

export const customerV2Service = {
  listCustomers: async (limit: number = 50, offset: number = 0): Promise<{ customers: Customer[], total: number }> => {
    try {
      const response = await apiClient.get('/admin/customers', {
        params: { limit, offset }
      });
      return {
        customers: response.data?.customers || [],
        total: response.data?.total || 0
      };
    } catch (error) {
      console.error('[customerV2Service] listCustomers failed:', error);
      throw error;
    }
  },

  getCustomerById: async (id: string): Promise<Customer> => {
    const response = await apiClient.get(`/admin/customers/${id}`);
    return response.data?.customer;
  },

  createCustomer: async (data: Partial<Customer>): Promise<Customer> => {
    try {
      // Map frontend model to backend DTO if needed
      // Backend expects: company_name, gst_number, billing_address, contact_email, contact_phone
      const payload = {
        company_name: data.name,
        gst_number: data.gstin,
        billing_address: data.address,
        contact_email: data.email,
        contact_phone: data.phone,
        status: 'ACTIVE'
      };

      const response = await apiClient.post('/admin/customers', payload);
      return response.data;
    } catch (error) {
      console.error('[customerV2Service] createCustomer failed:', error);
      throw error;
    }
  }
};
