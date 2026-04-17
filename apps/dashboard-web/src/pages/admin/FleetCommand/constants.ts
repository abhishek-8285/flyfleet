export interface DropdownOption {
  value: string;
  label: string;
  color?: string;
}

export const FUEL_TYPES: DropdownOption[] = [
  { value: "DIESEL", label: "Diesel" },
  { value: "PETROL", label: "Petrol" },
  { value: "CNG", label: "CNG" },
  { value: "ELECTRIC", label: "Electric" },
  { value: "HYBRID", label: "Hybrid" },
  { value: "LPG", label: "LPG" },
];

export const VEHICLE_TYPES: DropdownOption[] = [
  { value: "TRUCK", label: "Truck" },
  { value: "BUS", label: "Bus" },
  { value: "VAN", label: "Van" },
  { value: "CAR", label: "Car" },
  { value: "TRAILER", label: "Trailer" },
  { value: "TANKER", label: "Tanker" },
  { value: "TIPPER", label: "Tipper" },
  { value: "CRANE", label: "Crane" },
  { value: "OTHER", label: "Other" },
];

export const OWNERSHIP_TYPES: DropdownOption[] = [
  { value: "OWNED", label: "Owned" },
  { value: "LEASED", label: "Leased" },
  { value: "FINANCED", label: "Financed" },
  { value: "VENDOR", label: "Vendor" },
];

export const STATUS_OPTIONS: DropdownOption[] = [
  { value: "ACTIVE", label: "Active", color: "text-success" },
  { value: "MAINTENANCE", label: "Maintenance", color: "text-warning" },
  { value: "SUSPENDED", label: "Out of Service", color: "text-danger" },
  { value: "DECOMMISSIONED", label: "Decommissioned", color: "text-text-muted" },
];

export const ODOMETER_UNITS: DropdownOption[] = [
  { value: "km", label: "Kilometers (km)" },
  { value: "mi", label: "Miles (mi)" },
];

export const WEIGHT_UNITS: DropdownOption[] = [
  { value: "tonnes", label: "Tonnes (metric)" },
  { value: "kg", label: "Kilograms (kg)" },
  { value: "lbs", label: "Pounds (lbs)" },
];

export const FUEL_UNITS: DropdownOption[] = [
  { value: "litres", label: "Litres (L)" },
  { value: "gallons", label: "Gallons (US)" },
];

export const FUEL_LABELS: Record<string, string> = {
  DIESEL: "Diesel",
  PETROL: "Gasoline (Unleaded)",
  CNG: "Compressed Natural Gas",
  ELECTRIC: "Electric",
  HYBRID: "Hybrid",
  LPG: "LPG",
};

export const TYPE_ICONS: Record<string, string> = {
  TRUCK: "local_shipping",
  VAN: "airport_shuttle",
  CAR: "directions_car",
  BUS: "directions_bus",
  TRAILER: "luggage",
  TANKER: "water_drop",
  TIPPER: "dump_truck",
  CRANE: "construction",
};

export const STATUS_COLORS: Record<string, { bg: string; text: string; label: string }> = {
  ACTIVE: { bg: "bg-success/20", text: "text-success", label: "Active" },
  MAINTENANCE: { bg: "bg-warning/20", text: "text-warning", label: "Maintenance" },
  SUSPENDED: { bg: "bg-danger/20", text: "text-danger", label: "Inactive" },
  DECOMMISSIONED: { bg: "bg-text-muted/20", text: "text-text-muted", label: "Decommissioned" },
};

export const getFuelUnits = (fuelType: string): DropdownOption[] => {
  if (fuelType === "CNG") return [{ value: "kg", label: "Kilograms (kg)" }];
  return FUEL_UNITS;
};
