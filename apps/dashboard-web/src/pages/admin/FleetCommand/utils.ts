/**
 * Unit conversion and formatting utilities for FleetCommand
 */

// ── Conversion to metric (what backend stores) ──

export const toKm = (val: number, unit: string) => 
  unit === "mi" ? val * 1.60934 : val;

export const toTonnes = (val: number, unit: string) => 
  unit === "lbs" ? val / 2204.62 : unit === "kg" ? val / 1000 : val;

export const toLitres = (val: number, unit: string) => 
  unit === "gallons" ? val * 3.78541 : val;


// ── Conversion from metric (backend) to display units ──

export const fromKm = (val: number, unit: string) => 
  unit === "mi" ? (val / 1.60934).toFixed(1) : val.toLocaleString();

export const fromTonnes = (val: number, unit: string) => 
  unit === "lbs" ? (val * 2204.62).toFixed(0) : unit === "kg" ? (val * 1000).toFixed(0) : val.toString();

export const fromLitres = (val: number, unit: string) => 
  unit === "gallons" ? (val / 3.78541).toFixed(1) : val.toString();

/**
 * Returns the abbreviated unit label for display
 */
export const getUnitLabel = (type: "odometer" | "capacity" | "fuel", unit: string) => {
  if (type === "odometer") return unit === "mi" ? "mi" : "km";
  if (type === "capacity") {
    if (unit === "lbs") return "lbs";
    if (unit === "kg") return "kg";
    return "tonnes";
  }
  return unit === "gallons" ? "gal" : "L";
};
