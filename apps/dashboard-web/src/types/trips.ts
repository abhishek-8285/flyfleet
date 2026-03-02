// AUTO-GENERATED from OpenAPI spec - DO NOT EDIT MANUALLY
// Source: services/microservices/vehicle-service/api/openapi/trips.yaml

// Trip status values
export type TripStatus = 'scheduled' | 'in_progress' | 'completed' | 'cancelled';

// Core Trip model
export interface Trip {
  id: string;
  trip_id: string; // Pattern: TR-YYYY-NNNN
  vehicle_id: string;
  driver_id?: string | null;
  start_location: string;
  end_location: string;
  start_time: string; // ISO 8601 date-time
  end_time?: string | null;
  distance: number;
  status: TripStatus;
  organization_id: string;
  created_at: string;
  updated_at: string;
}

// Request to create a new trip
export interface CreateTripRequest {
  vehicle_id: string; // UUID, required
  driver_id?: string | null;
  start_location: string; // 1-255 chars, required
  end_location: string; // 1-255 chars, required
  start_time: string; // ISO 8601, required
}

// Request to update an existing trip
export interface UpdateTripRequest {
  driver_id?: string | null;
  start_location?: string;
  end_location?: string;
  distance?: number; // >= 0
  status?: TripStatus;
  end_time?: string;
}

// Response for listing trips
export interface TripListResponse {
  trips: Trip[];
  total: number;
  page: number;
  limit: number;
}

// Telemetry event
export interface Telemetry {
  id: string;
  trip_id: string;
  latitude: number; // -90 to 90
  longitude: number; // -180 to 180
  speed?: number; // >= 0
  heading?: number; // 0-360
  event_type?: string;
  event_time: string;
}

// Response for telemetry data
export interface TelemetryResponse {
  events: Telemetry[];
  count: number;
}

// Organization trip settings
export interface TripSettings {
  id?: string;
  organization_id: string;
  retention_period: number; // 1-120 months
  policies?: Record<string, unknown>;
  created_at?: string;
  updated_at?: string;
}

// Standard error response
export interface ApiError {
  error: string;
  details?: string;
  code?: string;
}

// Validation error response
export interface ValidationError {
  error: string;
  validation_errors: Array<{
    field: string;
    message: string;
  }>;
}

// Query parameters for listing trips
export interface TripListParams {
  page?: number; // >= 1, default 1
  limit?: number; // 1-100, default 20
  status?: TripStatus;
  vehicle_id?: string;
  driver_id?: string;
  start_date?: string;
  end_date?: string;
}

// Query parameters for telemetry
export interface TelemetryParams {
  limit?: number; // 1-500, default 50
}
