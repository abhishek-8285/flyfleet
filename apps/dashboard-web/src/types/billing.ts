export type InvoiceStatus = "PAID" | "PENDING" | "OVERDUE" | "PARTIAL" | "DRAFT" | "SENT" | "VIEWED";

export interface InvoiceItem {
  id: string;
  invoice_id: string;
  description: string;
  quantity: number;
  unit_price: number;
  tax_percent: number;
  total: number;
  hsn_sac?: string;
  unit?: string;
  list_price?: number;
  discount_amount?: number;
  created_at?: string;
  updated_at?: string;
}

export interface Invoice {
  id: string;
  org_id: string;
  customer_name: string;
  customer_email?: string;
  customer_phone?: string;
  billing_address?: string;
  invoice_number?: string;
  invoice_date: string;
  due_date: string;
  status: InvoiceStatus;
  currency: string;
  total_amount: number;
  amount_paid: number;
  balance_due: number;
  subtotal?: number;
  tax_amount?: number;
  notes?: string;
  category?: string;
  template_type?: "TAX_INVOICE" | "TRIP_NOTE";
  
  // Logistics
  trip_id?: string;
  vehicle_no?: string;
  driver_name?: string;
  driver_mobile?: string;
  pickup_location?: string;
  drop_location?: string;
  pickup_time?: string;
  drop_time?: string;
  map_url?: string;
  billed_distance?: number;

  // Compliance
  place_of_supply?: string;
  reverse_charge?: boolean;
  gstin_seller?: string;
  pan_seller?: string;
  eway_bill_no?: string;
  irn?: string;

  // Bank
  bank_name?: string;
  account_no?: string;
  ifsc_code?: string;
  branch_name?: string;
  beneficiary_name?: string;

  items?: InvoiceItem[];
  created_at: string;
  updated_at: string;
}

export interface BillingSettings {
  org_id: string;
  default_tax_rate: number;
  base_currency: string;
  grace_period_days: number;
  auto_invoicing_enabled: boolean;
  created_at?: string;
  updated_at?: string;
}

export interface BillingMetrics {
  total_invoiced_mtd: number;
  total_invoiced_trend?: string;
  total_invoiced_trend_up?: boolean;
  pending_payment: number;
  overdue_amount: number;
  paid_count: number;
  open_count: number;
  subtitle_overdue?: string;
  subtitle_paid?: string;
}
