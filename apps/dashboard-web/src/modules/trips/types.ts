export type TripStatus = 
  | "CREATED" 
  | "SCHEDULED" 
  | "STARTED" 
  | "PAUSED" 
  | "RESUMED" 
  | "COMPLETED" 
  | "CANCELLED" 
  | "FAILED";

export interface Location {
  latitude: number;
  longitude: number;
  address: string;
}

export interface Waypoint {
  id: string;
  name: string;
  type: "pickup" | "dropoff" | "checkpoint";
  address: string;
  location: Location;
  sequence: number;
  status: "pending" | "arrived" | "departed" | "skipped";
}

export interface TripEvent {
  id: string;
  type: "arrival" | "departure" | "delivery_attempt" | "delivery_success" | "delivery_failed" | "delay" | "breakdown" | "accident" | "speeding" | "braking";
  timestamp: string;
  location: Location;
  description: string;
  payload?: Record<string, any>;
}

export interface Trip {
  id: string;
  org_id: string;
  vehicle_id: string;
  driver_id?: string;
  status: TripStatus;
  trip_type: string;
  title: string;
  description: string;
  pickup_geo?: Location;
  drop_geo?: Location;
  planned_route?: Waypoint[];
  actual_route?: Waypoint[];
  waypoints?: Waypoint[];
  events?: TripEvent[];
  planned_distance: number;
  total_distance?: number;
  estimated_duration_min: number;
  actual_duration_min?: number;
  scheduled_start: string;
  start_time?: string;
  end_time?: string;
  customer_name: string;
  eway_bill_no?: string;
  created_at: string;
  updated_at: string;
}
