export interface User {
  id: string;
  user_id?: string;       // Alternate user identifier (backwards compatibility)
  email: string;
  name: string;
  role?: string;           // Optional for backwards compatibility
  role_id?: string;       // Primary role identifier used in RBAC
  org_id?: string;        // Organization ID
  permissions: string[];  // Array of permission strings like "vehicle:read"
  theme?: string;
  language?: string;
  sidebar_collapsed?: boolean;
}

export interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  loading: boolean;
  error: string | null;
}

export interface Vehicle {
  id: string;
  org_id: string;
  registration_number: string;
  type: string;
  make: string;
  model: string;
  year: number;
  status: 'active' | 'inactive' | 'maintenance';
  current_driver?: Driver;
  current_location?: Location;
  last_seen?: string;
  gps_device_id?: string;
  odometer?: number;
  fuel_level?: number;
  engine_hours?: number;
  metadata: Record<string, any>;
}

export interface Driver {
  id: string;
  name: string;
  phone: string;
  license_number: string;
  status: 'available' | 'on_trip' | 'off_duty';
  rating: number;
  behavior_score: number;
}

export interface Location {
  latitude: number;
  longitude: number;
  speed: number;
  heading: number;
  accuracy: number;
  timestamp: string;
  ignition?: boolean;
}

export interface Trip {
  id: string;
  vehicle_id: string;
  driver_id: string;
  status: 'planned' | 'ongoing' | 'paused' | 'completed' | 'cancelled';
  start_time: string;
  end_time?: string;
  start_location: GeoPoint;
  end_location: GeoPoint;
  distance_km?: number;
  fuel_consumed_liters?: number;
  route?: GeoPoint[];
  waypoints: Waypoint[];
}

export interface GeoPoint {
  latitude: number;
  longitude: number;
  name?: string;
  address?: string;
}

export interface Waypoint extends GeoPoint {
  type: string;
  eta?: string;
  arrival_time?: string;
  departure_time?: string;
}

export interface Alert {
  id: string;
  type: string;
  severity: 'info' | 'warning' | 'critical';
  vehicle_id?: string;
  driver_id?: string;
  message: string;
  details: Record<string, any>;
  timestamp: string;
  acknowledged: boolean;
}

export interface Geofence {
  id: string;
  name: string;
  type: 'circle' | 'polygon';
  geometry: any;
  radius?: number;
  entry_alert: boolean;
  exit_alert: boolean;
  active: boolean;
}

export interface Statistics {
  total_vehicles: number;
  active_vehicles: number;
  total_drivers: number;
  active_trips: number;
  total_distance_today: number;
  fuel_consumed_today: number;
  alerts_count: number;
}

export interface TimelineEvent {
  id: string;
  type: string;
  title: string;
  description: string;
  severity: 'info' | 'warning' | 'critical';
  timestamp: string;
  actor: string;
  actor_type: string;
  metadata: Record<string, any>;
}

export interface Organization {
  id: string;
  name: string;
  domain: string;
  plan: string;
  status: 'active' | 'inactive' | 'suspended' | 'pending';
  sla_tier: string;
  users_count: number;
  vehicles_count: number;
  devices_count: number;
  billing_status: string;
  created_at: string;
  last_activity: string;
  risk_level: 'low' | 'medium' | 'high' | 'critical';
  default_theme: string;
}
