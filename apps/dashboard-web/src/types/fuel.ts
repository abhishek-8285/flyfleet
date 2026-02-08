export type FuelLogStatus = 'PENDING' | 'APPROVED' | 'REJECTED';

export interface FuelLog {
  id: string;
  organization_id: string;
  vehicle_id: string;
  driver_id: string;
  station_name: string;
  fuel_type: string;
  quantity: number;
  unit_price: number;
  total_cost: number;
  odometer: number;
  transaction_date: string;
  status: FuelLogStatus;
  receipt_url?: string;
  notes?: string;
  created_by: string;
  created_at: string;
  updated_at: string;
  vehicle?: {
    id: string;
    name: string;
    plate: string;
    fleet_code: string;
  };
  driver?: {
    id: string;
    name: string;
  };
}

export interface CreateFuelLogRequest {
  vehicle_id: string;
  driver_id: string;
  station_name: string;
  fuel_type: string;
  quantity: number;
  unit_price: number;
  total_cost: number;
  odometer: number;
  transaction_date: string;
  receipt_url?: string;
  notes?: string;
}

export interface UpdateFuelLogRequest {
  station_name?: string;
  fuel_type?: string;
  quantity?: number;
  unit_price?: number;
  total_cost?: number;
  odometer?: number;
  transaction_date?: string;
  receipt_url?: string;
  notes?: string;
}

export interface FuelLogFilter {
  vehicle_id?: string;
  driver_id?: string;
  status?: FuelLogStatus[];
  search?: string;
  page?: number;
  page_size?: number;
}

export interface FuelLogStats {
  total_consumption: number;
  total_cost: number;
  avg_cost_per_km: number;
  avg_consumption: number;
  pending_approvals: number;
  by_status: Record<string, number>;
  by_fuel_type: Record<string, number>;
}

export interface FuelSettings {
  organization_id: string;
  default_currency: string;
  allowed_fuel_types: string[];
  validation_rules: {
    odometer_check: boolean;
    cost_calculation_verify: boolean;
  };
  auto_sync_enabled: boolean;
  updated_at: string;
}
