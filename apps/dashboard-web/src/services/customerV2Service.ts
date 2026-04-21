import apiClient from './apiClient';

export interface Customer {
  id: string;
  name: string;
  email?: string;
  phone?: string;
  address?: string;
  gstin?: string;
  status: string;
  created_at?: string;
  updated_at?: string;
}

export interface CustomerSummary {
  active_shipments: number;
  total_shipments: number;
  pending_bookings: number;
  total_revenue: number;
  unpaid_invoices: number;
}

export const customerV2Service = {
  listCustomers: async (page: number = 1, size: number = 20, filters: any = {}): Promise<{ data: Customer[], meta: { total: number, total_page: number } }> => {
    try {
      const response = await apiClient.get('/admin/customers', {
        params: { page, size, ...filters }
      });
      
      const res = response?.data || response;
      const rawData = res.data || [];
      
      const mappedData = rawData.map((item: any) => {
        let address = "";
        if (item.billing_address) {
          if (typeof item.billing_address === 'string') {
            try {
              // Try to parse if it's JSON-quoted (common for JSONB results from some scanners)
              if (item.billing_address.startsWith('"') && item.billing_address.endsWith('"')) {
                address = JSON.parse(item.billing_address);
              } else if (/^[A-Za-z0-9+/]+={0,2}$/.test(item.billing_address) && item.billing_address.length % 4 === 0 && item.billing_address.length > 20) {
                // Only try atob if it looks like serious base64 (e.g. longer than a typical short word)
                address = atob(item.billing_address);
              } else {
                address = item.billing_address;
              }
            } catch (e) {
              address = item.billing_address;
            }
          } else {
            address = typeof item.billing_address === 'object' ? JSON.stringify(item.billing_address) : String(item.billing_address);
          }
        }

        return {
          id: item.id,
          name: item.name,
          email: item.contact_email || item.email,
          phone: item.contact_phone || item.phone,
          address: address,
          gstin: item.gst_number || item.gstin,
          status: item.status
        };
      });

      return {
        data: mappedData,
        meta: res.meta || { total: 0, total_page: 0 }
      };
    } catch (error) {
      console.error('[customerV2Service] listCustomers failed:', error);
      throw error;
    }
  },

  getCustomerById: async (id: string): Promise<Customer> => {
    const response = await apiClient.get(`/admin/customers/${id}`);
    const item = response?.data || response;
    
    let address = "";
    if (item.billing_address) {
      if (typeof item.billing_address === 'string') {
        try {
          if (item.billing_address.startsWith('"') && item.billing_address.endsWith('"')) {
            address = JSON.parse(item.billing_address);
          } else if (/^[A-Za-z0-9+/]+={0,2}$/.test(item.billing_address) && item.billing_address.length % 4 === 0 && item.billing_address.length > 20) {
            address = atob(item.billing_address);
          } else {
            address = item.billing_address;
          }
        } catch (e) {
          address = item.billing_address;
        }
      } else {
        address = typeof item.billing_address === 'object' ? JSON.stringify(item.billing_address) : String(item.billing_address);
      }
    }
    
    return {
      id: item.id,
      name: item.name,
      email: item.contact_email || item.email,
      phone: item.contact_phone || item.phone,
      address: address,
      gstin: item.gst_number || item.gstin,
      status: item.status,
      created_at: item.created_at,
      updated_at: item.updated_at
    };
  },

  getCustomerSummary: async (id: string): Promise<CustomerSummary> => {
    const response = await apiClient.get(`/admin/customers/${id}/summary`);
    return response?.data || response;
  },

  createCustomer: async (data: Partial<Customer>): Promise<Customer> => {
    try {
      const payload = {
        company_name: data.name,
        gst_number: data.gstin,
        billing_address: data.address,
        contact_email: data.email,
        contact_phone: data.phone,
        status: data.status || 'ACTIVE'
      };

      const response = await apiClient.post('/admin/customers', payload);
      return response?.data || response;
    } catch (error) {
      console.error('[customerV2Service] createCustomer failed:', error);
      throw error;
    }
  },

  updateCustomer: async (id: string, data: Partial<Customer>): Promise<Customer> => {
    try {
      const payload = {
        company_name: data.name,
        gst_number: data.gstin,
        billing_address: data.address,
        contact_email: data.email,
        contact_phone: data.phone,
        status: data.status
      };

      const response = await apiClient.put(`/admin/customers/${id}`, payload);
      return response?.data || response;
    } catch (error) {
      console.error('[customerV2Service] updateCustomer failed:', error);
      throw error;
    }
  }
};
