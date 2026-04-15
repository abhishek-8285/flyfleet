export type DriverStatus = "active" | "inactive" | "suspended" | "resigned";
export type LicenseType = "LMV" | "HMV" | "Transport";
export type EmploymentType = "FULL_TIME" | "CONTRACT";
export type DocumentType = "license" | "aadhaar" | "pan" | "medical" | "training" | "other";
export type AttendanceStatus = "present" | "absent" | "leave" | "holiday";

export interface Address {
  street: string;
  city: string;
  state: string;
  pincode: string;
  country: string;
  latitude?: number;
  longitude?: number;
}

export interface EmergencyContact {
  name: string;
  phone: string;
  relation: string;
}

export interface DriverDocument {
  id: string;
  document_type: string;
  document_number: string;
  issuing_authority: string;
  issue_date: string;
  expiry_date: string;
  file_url: string;
  verified: boolean;
  verified_by?: string;
  verified_at?: string;
  created_at: string;
}

export interface AttendanceRecord {
  id: string;
  date: string;
  status: AttendanceStatus;
  check_in_time?: string;
  check_out_time?: string;
  remarks?: string;
  created_at: string;
}

export interface WorkSchedule {
  id: string;
  day_of_week: number;
  start_time: string;
  end_time: string;
  is_working: boolean;
  created_at: string;
  updated_at: string;
}

export interface Driver {
  id: string;
  org_id: string;
  first_name: string;
  last_name: string;
  name: string;
  email: string;
  phone: string;
  employee_code?: string;
  aadhaar_number?: string;
  pan_number?: string;
  license_number: string;
  license_type?: string;
  license_expiry: string;
  license_issue_date?: string;
  license_authority: string;
  date_of_birth: string;
  address: Address;
  emergency_contact: EmergencyContact;
  status: DriverStatus;
  current_vehicle_id?: string;
  current_trip_id?: string;
  joining_date: string;
  resignation_date?: string;
  experience_years: number;
  employment_type: EmploymentType;
  rating: number;
  total_trips: number;
  total_distance_km: number;
  total_hours_driven: number;
  behavior_score: number;
  documents: DriverDocument[];
  attendance_records: AttendanceRecord[];
  schedules: WorkSchedule[];
  metadata?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
  deleted_at?: string;
}

export interface DriverListResponse {
  data: Driver[];
  total: number;
  page: number;
  limit: number;
}

export interface DriverSession {
  id: string;
  driverId: string;
  vehicleId: string;
  startTime: string;
  endTime?: string;
  startLocation: { lat: number; lng: number };
  endLocation?: { lat: number; lng: number };
}

export interface SafetyScore {
  driverId: string;
  score: number;
  events: {
    harshBraking: number;
    harshTurning: number;
    overspeeding: number;
  };
  lastUpdated: string;
}

export interface BehaviorEvent {
  id: string;
  driver_id: string;
  type: string;
  severity: "low" | "medium" | "high" | "critical";
  timestamp: string;
  score_change: number;
  reason: string;
  location?: { lat: number; lng: number };
}

export interface DriverAnalytics {
  driver_id: string;
  total_trips: number;
  total_distance_km: number;
  total_hours_driven: number;
  avg_behavior_score: number;
  avg_safety_score: number;
  behavior_score: number;
  incidents_count: number;
  recent_incidents: BehaviorEvent[];
  trend: "improving" | "stable" | "declining";
}
