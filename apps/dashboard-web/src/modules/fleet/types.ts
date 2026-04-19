export type VehicleStatus =
  | "ACTIVE"
  | "IDLE"
  | "MAINTENANCE"
  | "PROVISIONING"
  | "STALLED"
  | "RETIRED";

export interface Vehicle {
  id: string; // VIN or Internal UUID
  registrationNumber: string;
  name: string;
  type: string; // 'TRUCK', 'VAN', 'HEAVY_HAUL'
  status: VehicleStatus;
  orgId: string;
  driverId?: string;
  lastUpdate: string;
  metadata: {
    make: string;
    model: string;
    year: number;
    fuelType: "DIESEL" | "EV" | "CNG";
    capacity: number; // in tons or cubic meters
    licensePlate: string;
  };
  gps_enabled?: boolean;
  gps_device_id?: string;
  gps_device_type?: string;
  telemetry?: {
    latitude: number;
    longitude: number;
    speed: number;
    heading: number;
    lastHeartbeat: string;
  };
}

export interface FleetSegment {
  id: string;
  name: string;
  orgId: string;
  h3HomeCell: string; // The "base" location H3 cell
  vehicleIds: string[];
}
