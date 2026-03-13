export type DriverStatus = "ON_DUTY" | "OFF_DUTY" | "BREAK" | "SUSPENDED";

export interface Driver {
  id: string;
  name: string;
  orgId: string;
  status: DriverStatus;
  licenseNumber: string;
  contactNumber: string;
  rating: number; // 0.0 - 5.0
  activeVehicleId?: string;
  activeSessionId?: string;
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
