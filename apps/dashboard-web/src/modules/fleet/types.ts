export type VehicleStatus =
  | "ACTIVE"
  | "IDLE"
  | "MAINTENANCE"
  | "PROVISIONING"
  | "SUSPENDED"
  | "DECOMMISSIONED"
  | "RETIRED";

export type VehicleType = "TRUCK" | "VAN" | "HEAVY_HAUL" | "BUS" | "CAR" | "BIKE";
export type FuelType = "DIESEL" | "PETROL" | "CNG" | "EV" | "HYBRID";
export type OwnershipType = "OWNED" | "LEASED" | "PARTNER";

export type GPSStatus = "ONLINE" | "OFFLINE" | "STALE";
export type DeviceHealthStatus = "HEALTHY" | "WARNING" | "ERROR" | "DISCONNECTED";
export type IgnitionStatus = "ON" | "OFF" | "UNKNOWN";

export type VerificationStatus = "PENDING" | "APPROVED" | "REJECTED";
export type DocumentType = "INSURANCE" | "PUC" | "PERMIT" | "FITNESS" | "TAX";

export type MaintenanceStatus = "HEALTHY" | "DUE" | "OVERDUE";
export type DriverLicenseStatus = "VALID" | "EXPIRED" | "MISSING";
export type RiskCategory = "LOW" | "MEDIUM" | "HIGH";
export type NetworkType = "CELLULAR" | "WIFI" | "SATELLITE" | "OFFLINE";

export interface VehicleDocument {
  id: string;
  type: DocumentType;
  documentNumber: string;
  expiryDate: string;
  status: VerificationStatus;
  version: number;
  documentUrl?: string;
  verifiedAt?: string;
}

export interface VehicleSummary {
  complianceStatus: "RED" | "YELLOW" | "GREEN";
  complianceScore: number;
  complianceBreachCount: number;
  gpsStatus: GPSStatus;
  deviceHealthStatus: DeviceHealthStatus;
  ignitionStatus: IgnitionStatus;
  lastSeen?: string;
  alertCount: number;
  criticalAlertCount: number;
  overSpeedCount: number;
  geoFenceViolation: boolean;
  activeTrip: boolean;
  lastTripEndTime?: string;
  dataFreshnessScore: number;
  nextExpiryDate?: string;
  nextExpiryType?: DocumentType;
  packetLossPercentage?: number;
  telemetryReliability?: number;
  maintenanceStatus: MaintenanceStatus;
  utilizationScore: number;
  fuelEfficiency30d: number;
  driverAssigned: boolean;
  driverLicenseStatus: DriverLicenseStatus;
  riskScore: number;
  riskCategory: RiskCategory;
  lastLocationLat?: number;
  lastLocationLng?: number;
  lastPingIP?: string;
  networkType?: NetworkType;
  isHardBlocked: boolean;
}

export interface Vehicle {
  id: string;
  registrationNumber: string;
  maskedRegistration?: string;
  type: VehicleType;
  status: VehicleStatus;
  tenantId: string;
  driverId?: string;
  lastUpdate: string;
  make: string;
  model: string;
  year: number;
  fuelType: FuelType;
  capacityTonnage: number;
  fuelTankCapacity: number;
  ownershipType: OwnershipType;
  isBlacklisted: boolean;
  isHardBlocked: boolean;
  riskScore: number;
  gpsEnabled: boolean;
  gpsDeviceId?: string;
  version: number;
  odometer?: number;
  createdAt?: string;
  summary?: VehicleSummary;
  documents?: VehicleDocument[];
  currentDriver?: {
    id: string;
    name: string;
  };
}

export type ComplianceStatus = "RED" | "YELLOW" | "GREEN" | "ALL";

export interface VehicleSearchParams {
  q?: string;
  status?: VehicleStatus;
  vehicle_type?: VehicleType;
  fuel_type?: FuelType;
  compliance_status?: ComplianceStatus;
  gps_status?: GPSStatus;
  risk_category?: RiskCategory;
  tenant_id?: string;
  page?: number;
  limit?: number;
}

export interface VehicleListResponse {
  vehicles: Vehicle[];
  total: number;
  page: number;
  limit: number;
}
