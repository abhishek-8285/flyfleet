import React, { useState, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import apiClient from "../../../services/apiClient";

// ── Constants ──
const FUEL_TYPES = [
  { value: "DIESEL", label: "Diesel" },
  { value: "PETROL", label: "Petrol" },
  { value: "CNG", label: "CNG" },
  { value: "ELECTRIC", label: "Electric" },
  { value: "HYBRID", label: "Hybrid" },
  { value: "LPG", label: "LPG" },
];

const VEHICLE_TYPES = [
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

const OWNERSHIP_TYPES = [
  { value: "OWNED", label: "Owned" },
  { value: "LEASED", label: "Leased" },
  { value: "FINANCED", label: "Financed" },
  { value: "VENDOR", label: "Vendor" },
];

// ── Unit Constants ──
const ODOMETER_UNITS = [
  { value: "km", label: "Kilometers (km)" },
  { value: "mi", label: "Miles (mi)" },
];

const WEIGHT_UNITS = [
  { value: "tonnes", label: "Tonnes (metric)" },
  { value: "kg", label: "Kilograms (kg)" },
  { value: "lbs", label: "Pounds (lbs)" },
];

const FUEL_UNITS = [
  { value: "litres", label: "Litres (L)" },
  { value: "gallons", label: "Gallons (US)" },
];

// Get fuel units based on fuel type (CNG uses kg, not litres)
const getFuelUnits = (fuelType: string) => {
  if (fuelType === "CNG") return [{ value: "kg", label: "Kilograms (kg)" }];
  return FUEL_UNITS;
};

// Conversion to metric (what backend stores)
const toKm = (val: number, unit: string) => unit === "mi" ? val * 1.60934 : val;
const toTonnes = (val: number, unit: string) => unit === "lbs" ? val / 2204.62 : unit === "kg" ? val / 1000 : val;
const toLitres = (val: number, unit: string) => unit === "gallons" ? val * 3.78541 : val;

// ── Types ──
interface FormState {
  registration_number: string;
  vehicle_type: string;
  fuel_type: string;
  ownership_type: string;
  make: string;
  model: string;
  year: string;
  vin_number: string;
  chassis_number: string;
  engine_number: string;
  capacity_tonnage: string;
  fuel_tank_capacity: string;
  gps_device_id: string;
  color: string;
  parent_vehicle_id: string;
  branch_id: string;
  cost_center_id: string;
  immobilizer_enabled: boolean;
  ais140_compliant: boolean;
}

type FieldName = keyof FormState;

interface FieldValidation {
  required?: boolean;
  maxLength?: number;
  minLength?: number;
  pattern?: RegExp;
  min?: number;
  max?: number;
  custom?: (value: string) => string | null;
}

interface ValidationRule {
  [key: string]: {
    label: string;
    rules: FieldValidation[];
  };
}

// ── Validation Rules ──
const VALIDATION_RULES: ValidationRule = {
  registration_number: {
    label: "Registration Number",
    rules: [
      { required: true },
      { minLength: 6 },
      { maxLength: 20 },
      {
        pattern: /^[A-Z0-9\s\-\/]+$/i,
      },
    ],
  },
  vehicle_type: {
    label: "Vehicle Type",
    rules: [{ required: true }],
  },
  fuel_type: {
    label: "Fuel Type",
    rules: [{ required: true }],
  },
  ownership_type: {
    label: "Ownership Type",
    rules: [{ required: true }],
  },
  make: {
    label: "Make",
    rules: [{ maxLength: 100 }],
  },
  model: {
    label: "Model",
    rules: [{ maxLength: 100 }],
  },
  year: {
    label: "Year",
    rules: [
      {
        custom: (v) => {
          if (!v) return null; // optional
          const year = parseInt(v);
          const currentYear = new Date().getFullYear();
          if (isNaN(year)) return "Must be a valid year";
          if (year < 1900 || year > currentYear + 1)
            return `Year must be between 1900 and ${currentYear + 1}`;
          return null;
        },
      },
    ],
  },
  vin_number: {
    label: "VIN Number",
    rules: [
      {
        custom: (v) => {
          if (!v) return null; // optional
          const cleaned = v.trim();
          if (cleaned.length !== 17) return "VIN must be exactly 17 characters";
          if (!/^[A-HJ-NPR-Z0-9]{17}$/i.test(cleaned))
            return "VIN contains invalid characters (no I, O, Q)";
          return null;
        },
      },
    ],
  },
  chassis_number: {
    label: "Chassis Number",
    rules: [{ maxLength: 50 }],
  },
  engine_number: {
    label: "Engine Number",
    rules: [{ maxLength: 50 }],
  },
  capacity_tonnage: {
    label: "Capacity",
    rules: [
      {
        custom: (v) => {
          if (!v) return null;
          const num = parseFloat(v);
          if (isNaN(num)) return "Must be a valid number";
          if (num <= 0) return "Must be greater than 0";
          if (num > 1000) return "Must be less than 1000 tonnes";
          return null;
        },
      },
    ],
  },
  fuel_tank_capacity: {
    label: "Fuel Tank Capacity",
    rules: [
      {
        custom: (v) => {
          if (!v) return null;
          const num = parseFloat(v);
          if (isNaN(num)) return "Must be a valid number";
          if (num <= 0) return "Must be greater than 0";
          if (num > 5000) return "Must be less than 5000 litres";
          return null;
        },
      },
    ],
  },
  gps_device_id: {
    label: "GPS Device ID",
    rules: [{ maxLength: 100 }],
  },
  color: {
    label: "Color",
    rules: [{ maxLength: 50 }],
  },
};

// ── Helpers ──
const initialForm: FormState = {
  registration_number: "",
  vehicle_type: "",
  fuel_type: "DIESEL",
  ownership_type: "OWNED",
  make: "",
  model: "",
  year: new Date().getFullYear().toString(),
  vin_number: "",
  chassis_number: "",
  engine_number: "",
  capacity_tonnage: "",
  fuel_tank_capacity: "",
  gps_device_id: "",
  color: "",
  parent_vehicle_id: "",
  branch_id: "",
  cost_center_id: "",
  immobilizer_enabled: false,
  ais140_compliant: false,
};

function validateField(name: FieldName, value: string): string | null {
  const config = VALIDATION_RULES[name];
  if (!config) return null;

  for (const rule of config.rules) {
    if (rule.required && !value.trim()) {
      return `${config.label} is required`;
    }
    if (value.trim() && rule.minLength && value.trim().length < rule.minLength) {
      return `${config.label} must be at least ${rule.minLength} characters`;
    }
    if (value.trim() && rule.maxLength && value.trim().length > rule.maxLength) {
      return `${config.label} must be less than ${rule.maxLength} characters`;
    }
    if (value.trim() && rule.pattern && !rule.pattern.test(value)) {
      return `${config.label} contains invalid characters`;
    }
    if (value.trim() && rule.custom) {
      const error = rule.custom(value);
      if (error) return error;
    }
  }
  return null;
}

function validateStep(step: number, form: FormState): Record<string, string> {
  const errors: Record<string, string> = {};

  const stepFields: Record<number, FieldName[]> = {
    1: ["registration_number", "vehicle_type", "fuel_type", "ownership_type"],
    2: ["make", "model", "year", "vin_number", "chassis_number", "engine_number", "capacity_tonnage", "fuel_tank_capacity"],
    3: ["gps_device_id", "color"],
  };

  for (const field of stepFields[step] || []) {
    const error = validateField(field, form[field]);
    if (error) errors[field] = error;
  }

  return errors;
}

// ── Input Component ──
interface InputProps {
  name: FieldName;
  value: string;
  onChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  onBlur?: () => void;
  label: string;
  placeholder?: string;
  type?: string;
  icon?: string;
  error?: string;
  required?: boolean;
  hint?: string;
  className?: string;
  maxLength?: number;
}

const TextInput: React.FC<InputProps> = ({
  name, value, onChange, onBlur, label, placeholder, type = "text",
  icon, error, required, hint, className, maxLength,
}) => (
  <div className={`space-y-1.5 ${className || ""}`}>
    <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
      {label} {required && <span className="text-danger">*</span>}
    </label>
    <div className="relative">
      {icon && (
        <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted">
          <span className="material-symbols-outlined text-[18px]">{icon}</span>
        </span>
      )}
      <input
        type={type}
        name={name}
        value={value}
        onChange={onChange}
        onBlur={onBlur}
        placeholder={placeholder}
        maxLength={maxLength}
        className={`w-full bg-surface-darker border rounded-lg py-2.5 ${
          icon ? "pl-10" : "px-4"
        } pr-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder-text-muted/50 ${
          error
            ? "border-danger/50 focus:ring-danger focus:border-danger"
            : "border-border-color"
        }`}
      />
    </div>
    {error ? (
      <p className="text-[10px] text-danger flex items-center gap-1">
        <span className="material-symbols-outlined text-[12px]">error</span>
        {error}
      </p>
    ) : hint ? (
      <p className="text-[10px] text-text-muted">{hint}</p>
    ) : null}
  </div>
);

// ── Select Component ──
interface SelectProps {
  name: FieldName;
  value: string;
  onChange: (e: React.ChangeEvent<HTMLSelectElement>) => void;
  onBlur?: () => void;
  label: string;
  options: { value: string; label: string }[];
  icon?: string;
  error?: string;
  required?: boolean;
  placeholder?: string;
}

const SelectInput: React.FC<SelectProps> = ({
  name, value, onChange, onBlur, label, options, icon, error, required, placeholder,
}) => (
  <div className="space-y-1.5">
    <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
      {label} {required && <span className="text-danger">*</span>}
    </label>
    <div className="relative">
      {icon && (
        <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted">
          <span className="material-symbols-outlined text-[18px]">{icon}</span>
        </span>
      )}
      <select
        name={name}
        value={value}
        onChange={onChange}
        onBlur={onBlur}
        className={`w-full bg-surface-darker border rounded-lg py-2.5 ${
          icon ? "pl-10" : "pl-4"
        } pr-10 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none appearance-none ${
          error
            ? "border-danger/50 focus:ring-danger focus:border-danger"
            : "border-border-color"
        } ${!value ? "text-text-muted/50" : ""}`}
      >
        {placeholder && <option value="">{placeholder}</option>}
        {options.map((o) => (
          <option key={o.value} value={o.value}>{o.label}</option>
        ))}
      </select>
      <span className="absolute inset-y-0 right-0 flex items-center pr-3 pointer-events-none text-text-muted">
        <span className="material-symbols-outlined text-[18px]">expand_more</span>
      </span>
    </div>
    {error && (
      <p className="text-[10px] text-danger flex items-center gap-1">
        <span className="material-symbols-outlined text-[12px]">error</span>
        {error}
      </p>
    )}
  </div>
);

// ── Main Component ──
export const AddVehicle: React.FC = () => {
  const navigate = useNavigate();
  const [form, setForm] = useState<FormState>(initialForm);
  const [units, setUnits] = useState({
    odometer: "km",
    capacity: "tonnes",
    fuel: "litres",
  });
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [touched, setTouched] = useState<Record<string, boolean>>({});
  const [loading, setLoading] = useState(false);
  const [submitError, setSubmitError] = useState("");
  const [step, setStep] = useState<1 | 2 | 3>(1);

  const handleChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
      const { name, value } = e.target;
      setForm((prev) => ({ ...prev, [name]: value }));
      // Clear error on change if field is touched
      if (touched[name]) {
        const error = validateField(name as FieldName, value);
        setErrors((prev) => {
          const next = { ...prev };
          if (error) next[name] = error;
          else delete next[name];
          return next;
        });
      }
    },
    [touched]
  );

  const handleBlur = useCallback(
    (fieldName: FieldName) => {
      setTouched((prev) => ({ ...prev, [fieldName]: true }));
      const error = validateField(fieldName, form[fieldName]);
      setErrors((prev) => {
        const next = { ...prev };
        if (error) next[fieldName] = error;
        else delete next[fieldName];
        return next;
      });
    },
    [form]
  );

  const canProceed = useCallback((): boolean => {
    const stepErrors = validateStep(step, form);
    return Object.keys(stepErrors).length === 0;
  }, [step, form]);

  const handleNext = useCallback(() => {
    // Validate current step
    const stepErrors = validateStep(step, form);
    setErrors((prev) => ({ ...prev, ...stepErrors }));

    // Mark all step fields as touched
    const stepFields: Record<number, FieldName[]> = {
      1: ["registration_number", "vehicle_type", "fuel_type", "ownership_type"],
      2: ["make", "model", "year", "vin_number", "chassis_number", "engine_number", "capacity_tonnage", "fuel_tank_capacity"],
      3: ["gps_device_id", "color"],
    };
    const touchedMap: Record<string, boolean> = {};
    for (const f of stepFields[step] || []) touchedMap[f] = true;
    setTouched((prev) => ({ ...prev, ...touchedMap }));

    if (Object.keys(stepErrors).length === 0) {
      setStep((s) => (s + 1) as 1 | 2 | 3);
    }
  }, [step, form]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitError("");

    // Validate ALL steps before submit
    const allErrors: Record<string, string> = {};
    for (const s of [1, 2, 3]) {
      Object.assign(allErrors, validateStep(s, form));
    }

    // Mark all fields as touched
    const allTouched: Record<string, boolean> = {};
    for (const field of Object.keys(VALIDATION_RULES)) allTouched[field] = true;
    setTouched(allTouched);
    setErrors(allErrors);

    if (Object.keys(allErrors).length > 0) {
      setSubmitError("Please fix the errors below before submitting");
      // Go to the first step with errors
      if (allErrors.registration_number || allErrors.vehicle_type ||
          allErrors.fuel_type || allErrors.ownership_type) setStep(1);
      else if (allErrors.make || allErrors.model || allErrors.year ||
               allErrors.vin_number || allErrors.capacity_tonnage) setStep(2);
      else setStep(3);
      return;
    }

    setLoading(true);
    try {
      const payload: any = {
        registration_number: form.registration_number.trim().toUpperCase().replace(/\s+/g, ""),
        vehicle_type: form.vehicle_type,
        fuel_type: form.fuel_type,
        ownership_type: form.ownership_type,
      };

      if (form.make) payload.make = form.make.trim();
      if (form.model) payload.model = form.model.trim();
      if (form.year) payload.year = parseInt(form.year);
      if (form.vin_number) payload.vin_number = form.vin_number.trim().toUpperCase();
      if (form.chassis_number) payload.chassis_number = form.chassis_number.trim();
      if (form.engine_number) payload.engine_number = form.engine_number.trim();
      if (form.capacity_tonnage) payload.capacity_tonnage = toTonnes(parseFloat(form.capacity_tonnage), units.capacity);
      if (form.fuel_tank_capacity) {
        if (form.fuel_type === "CNG") {
          // CNG: kg stays as-is (fuel_tank_capacity stores kg for CNG)
          payload.fuel_tank_capacity = parseFloat(form.fuel_tank_capacity);
        } else {
          payload.fuel_tank_capacity = toLitres(parseFloat(form.fuel_tank_capacity), units.fuel);
        }
      }
      if (form.gps_device_id) payload.gps_device_id = form.gps_device_id.trim();
      if (form.color) payload.color = form.color.trim();
      if (form.parent_vehicle_id) payload.parent_vehicle_id = form.parent_vehicle_id.trim();
      if (form.branch_id) payload.branch_id = form.branch_id.trim();
      if (form.cost_center_id) payload.cost_center_id = form.cost_center_id.trim();
      payload.immobilizer_enabled = form.immobilizer_enabled;
      payload.ais140_compliant = form.ais140_compliant;

      await apiClient.post("/admin/vehicles", payload);
      navigate("/admin/vehicles");
    } catch (err: any) {
      console.error("Failed to create vehicle:", err);
      const msg = err.response?.data?.error || err.response?.data?.message || err.message;

      // Check for duplicate registration number
      if (err.response?.status === 409 || msg?.toLowerCase().includes("already exists") || msg?.toLowerCase().includes("duplicate") || msg?.toLowerCase().includes("unique")) {
        setErrors({ registration_number: "This registration number already exists in your organization" });
        setSubmitError("A vehicle with this registration number already exists");
      } else if (msg) {
        setSubmitError(msg);
      } else {
        setSubmitError("Failed to create vehicle. Please try again.");
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex-1 flex flex-col min-w-0 overflow-y-auto bg-background-dark">
      {/* Header */}
      <header className="h-16 flex items-center justify-between px-6 border-b border-border-color bg-surface-darker sticky top-0 z-10 flex-shrink-0">
        <div className="flex items-center gap-4">
          <button
            onClick={() => navigate("/admin/vehicles")}
            className="text-text-muted hover:text-white transition-colors"
          >
            <span className="material-symbols-outlined">arrow_back</span>
          </button>
          <div>
            <h1 className="text-white text-lg font-bold">Add New Vehicle</h1>
            <p className="text-[11px] text-text-muted mt-0.5">
              Fleet Management <span className="mx-1">&gt;</span> Vehicles <span className="mx-1">&gt;</span> Add
            </p>
          </div>
        </div>
      </header>

      {/* Form */}
      <div className="flex-1 overflow-y-auto custom-scrollbar">
        <form onSubmit={handleSubmit} className="max-w-4xl mx-auto w-full p-6 space-y-6" noValidate>
          {/* Step Indicator */}
          <div className="flex items-center gap-2">
            {[
              { num: 1, label: "Registration" },
              { num: 2, label: "Specifications" },
              { num: 3, label: "Additional" },
            ].map((s, i) => (
              <React.Fragment key={s.num}>
                {i > 0 && (
                  <div className={`flex-1 h-px ${step >= s.num ? "bg-primary" : "bg-border-color"}`} />
                )}
                <button
                  type="button"
                  onClick={() => setStep(s.num as 1 | 2 | 3)}
                  className={`flex items-center gap-2 px-3 py-1.5 rounded-lg text-xs font-medium transition-colors ${
                    step === s.num
                      ? "bg-primary/10 text-primary"
                      : step > s.num
                      ? "bg-success/10 text-success"
                      : "text-text-muted"
                  }`}
                >
                  <span className={`w-5 h-5 rounded-full flex items-center justify-center text-[10px] font-bold ${
                    step > s.num
                      ? "bg-success text-white"
                      : step === s.num
                      ? "bg-primary text-white"
                      : "bg-border-color text-text-muted"
                  }`}>
                    {step > s.num ? "✓" : s.num}
                  </span>
                  <span className="hidden sm:inline">{s.label}</span>
                </button>
              </React.Fragment>
            ))}
          </div>

          {/* Submit Error */}
          {submitError && (
            <div className="rounded-lg border border-danger/30 bg-danger/10 px-4 py-3 text-sm text-danger flex items-center gap-2">
              <span className="material-symbols-outlined text-sm">error</span>
              {submitError}
            </div>
          )}

          {/* ── Step 1: Registration ── */}
          {step === 1 && (
            <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
              <div className="px-6 py-4 border-b border-border-color bg-surface-darker flex items-center gap-3">
                <div className="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center text-primary border border-primary/20">
                  <span className="material-symbols-outlined text-sm">badge</span>
                </div>
                <div>
                  <h3 className="text-white text-sm font-semibold">Registration Details</h3>
                  <p className="text-xs text-text-muted mt-0.5">
                    Government registration and classification
                  </p>
                </div>
              </div>
              <div className="p-6 space-y-6">
                <TextInput
                  name="registration_number"
                  value={form.registration_number}
                  onChange={handleChange}
                  onBlur={() => handleBlur("registration_number")}
                  label="Registration Number"
                  placeholder="e.g. MH01AB1234"
                  icon="confirmation_number"
                  error={errors.registration_number}
                  required
                  maxLength={20}
                  hint="As per RC — unique per organization"
                />

                <SelectInput
                  name="vehicle_type"
                  value={form.vehicle_type}
                  onChange={handleChange}
                  onBlur={() => handleBlur("vehicle_type")}
                  label="Vehicle Type"
                  options={VEHICLE_TYPES}
                  icon="category"
                  error={errors.vehicle_type}
                  required
                  placeholder="Select Type"
                />

                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <SelectInput
                    name="fuel_type"
                    value={form.fuel_type}
                    onChange={handleChange}
                    onBlur={() => handleBlur("fuel_type")}
                    label="Fuel Type"
                    options={FUEL_TYPES}
                    icon="local_gas_station"
                    error={errors.fuel_type}
                    required
                  />

                  <SelectInput
                    name="ownership_type"
                    value={form.ownership_type}
                    onChange={handleChange}
                    onBlur={() => handleBlur("ownership_type")}
                    label="Ownership Type"
                    options={OWNERSHIP_TYPES}
                    icon="home_work"
                    error={errors.ownership_type}
                    required
                  />
                </div>
              </div>
            </div>
          )}

          {/* ── Step 2: Specifications ── */}
          {step === 2 && (
            <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
              <div className="px-6 py-4 border-b border-border-color bg-surface-darker flex items-center gap-3">
                <div className="w-8 h-8 rounded-full bg-purple-500/10 flex items-center justify-center text-purple-400 border border-purple-500/20">
                  <span className="material-symbols-outlined text-sm">directions_car</span>
                </div>
                <div>
                  <h3 className="text-white text-sm font-semibold">Vehicle Specifications</h3>
                  <p className="text-xs text-text-muted mt-0.5">
                    Make, model, capacity, and identification numbers
                  </p>
                </div>
              </div>
              <div className="p-6 space-y-6">
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                  <TextInput
                    name="make"
                    value={form.make}
                    onChange={handleChange}
                    onBlur={() => handleBlur("make")}
                    label="Make"
                    placeholder="e.g. Tata, Ashok Leyland"
                    icon="business"
                    error={errors.make}
                    maxLength={100}
                  />
                  <TextInput
                    name="model"
                    value={form.model}
                    onChange={handleChange}
                    onBlur={() => handleBlur("model")}
                    label="Model"
                    placeholder="e.g. 407, Eicher Pro"
                    error={errors.model}
                    maxLength={100}
                  />
                  <TextInput
                    name="year"
                    value={form.year}
                    onChange={handleChange}
                    onBlur={() => handleBlur("year")}
                    label="Year"
                    placeholder="e.g. 2023"
                    type="number"
                    error={errors.year}
                  />
                </div>

                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                  <TextInput
                    name="vin_number"
                    value={form.vin_number}
                    onChange={handleChange}
                    onBlur={() => handleBlur("vin_number")}
                    label="VIN Number"
                    placeholder="17-character VIN"
                    error={errors.vin_number}
                    maxLength={17}
                    hint="Must be exactly 17 characters"
                  />
                  <TextInput
                    name="chassis_number"
                    value={form.chassis_number}
                    onChange={handleChange}
                    onBlur={() => handleBlur("chassis_number")}
                    label="Chassis Number"
                    placeholder="Chassis number"
                    error={errors.chassis_number}
                    maxLength={50}
                  />
                  <TextInput
                    name="engine_number"
                    value={form.engine_number}
                    onChange={handleChange}
                    onBlur={() => handleBlur("engine_number")}
                    label="Engine Number"
                    placeholder="Engine number"
                    error={errors.engine_number}
                    maxLength={50}
                  />
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div className="space-y-1.5">
                    <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                      Capacity
                    </label>
                    <div className="flex gap-2">
                      <input
                        type="number"
                        name="capacity_tonnage"
                        value={form.capacity_tonnage}
                        onChange={handleChange}
                        onBlur={() => handleBlur("capacity_tonnage")}
                        placeholder="e.g. 10"
                        className="flex-1 bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder-text-muted/50"
                      />
                      <select
                        value={units.capacity}
                        onChange={(e) => setUnits((prev) => ({ ...prev, capacity: e.target.value }))}
                        className="bg-surface-darker border border-border-color rounded-lg px-3 py-2.5 text-sm text-white focus:ring-1 focus:ring-primary outline-none"
                      >
                        {WEIGHT_UNITS.map((u) => (
                          <option key={u.value} value={u.value}>{u.label}</option>
                        ))}
                      </select>
                    </div>
                    {errors.capacity_tonnage && (
                      <p className="text-[10px] text-danger flex items-center gap-1">
                        <span className="material-symbols-outlined text-[12px]">error</span>
                        {errors.capacity_tonnage}
                      </p>
                    )}
                  </div>
                  <div className="space-y-1.5">
                    <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                      Fuel Tank Capacity
                    </label>
                    <div className="flex gap-2">
                      <input
                        type="number"
                        name="fuel_tank_capacity"
                        value={form.fuel_tank_capacity}
                        onChange={handleChange}
                        onBlur={() => handleBlur("fuel_tank_capacity")}
                        placeholder="e.g. 200"
                        className="flex-1 bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder-text-muted/50"
                      />
                      <select
                        value={units.fuel}
                        onChange={(e) => setUnits((prev) => ({ ...prev, fuel: e.target.value }))}
                        className="bg-surface-darker border border-border-color rounded-lg px-3 py-2.5 text-sm text-white focus:ring-1 focus:ring-primary outline-none"
                      >
                        {getFuelUnits(form.fuel_type).map((u) => (
                          <option key={u.value} value={u.value}>{u.label}</option>
                        ))}
                      </select>
                    </div>
                    {errors.fuel_tank_capacity && (
                      <p className="text-[10px] text-danger flex items-center gap-1">
                        <span className="material-symbols-outlined text-[12px]">error</span>
                        {errors.fuel_tank_capacity}
                      </p>
                    )}
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* ── Step 3: Additional ── */}
          {step === 3 && (
            <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
              <div className="px-6 py-4 border-b border-border-color bg-surface-darker flex items-center gap-3">
                <div className="w-8 h-8 rounded-full bg-amber-500/10 flex items-center justify-center text-amber-400 border border-amber-500/20">
                  <span className="material-symbols-outlined text-sm">settings</span>
                </div>
                <div>
                  <h3 className="text-white text-sm font-semibold">Additional Details</h3>
                  <p className="text-xs text-text-muted mt-0.5">
                    GPS device, color, hierarchy, and compliance
                  </p>
                </div>
              </div>
              <div className="p-6 space-y-6">
                {/* GPS Device & Color */}
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <TextInput
                    name="gps_device_id"
                    value={form.gps_device_id}
                    onChange={handleChange}
                    onBlur={() => handleBlur("gps_device_id")}
                    label="GPS Device ID"
                    placeholder="e.g. GPS-IMEI-12345678"
                    icon="gps_fixed"
                    error={errors.gps_device_id}
                    maxLength={100}
                    hint="AIS-140 compliant GPS tracker IMEI or serial number"
                  />
                  <TextInput
                    name="color"
                    value={form.color}
                    onChange={handleChange}
                    onBlur={() => handleBlur("color")}
                    label="Color"
                    placeholder="e.g. White, Blue"
                    error={errors.color}
                    maxLength={50}
                  />
                </div>

                {/* Parent Vehicle & Branch */}
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <TextInput
                    name="parent_vehicle_id"
                    value={form.parent_vehicle_id}
                    onChange={handleChange}
                    onBlur={() => handleBlur("parent_vehicle_id")}
                    label="Parent Vehicle ID"
                    placeholder="UUID of parent vehicle"
                    icon="account_tree"
                    error={errors.parent_vehicle_id}
                    hint="For hierarchical vehicle grouping (e.g., trailer to truck)"
                  />
                  <TextInput
                    name="branch_id"
                    value={form.branch_id}
                    onChange={handleChange}
                    onBlur={() => handleBlur("branch_id")}
                    label="Branch ID"
                    placeholder="UUID of branch"
                    icon="account_balance"
                    error={errors.branch_id}
                    hint="Branch or depot this vehicle is assigned to"
                  />
                </div>

                {/* Cost Center */}
                <div className="space-y-1.5">
                  <TextInput
                    name="cost_center_id"
                    value={form.cost_center_id}
                    onChange={handleChange}
                    onBlur={() => handleBlur("cost_center_id")}
                    label="Cost Center ID"
                    placeholder="UUID of cost center"
                    icon="account_balance_wallet"
                    error={errors.cost_center_id}
                    hint="For financial tracking and cost allocation"
                  />
                </div>

                {/* Toggles: Immobilizer & AIS-140 */}
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div className="flex items-center justify-between p-4 bg-surface-darker rounded-lg border border-border-color">
                    <div>
                      <div className="text-sm font-medium text-white">Immobilizer</div>
                      <div className="text-[10px] text-text-muted mt-0.5">Remote engine immobilization capability</div>
                    </div>
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input
                        type="checkbox"
                        checked={form.immobilizer_enabled}
                        onChange={(e) => setForm((prev) => ({ ...prev, immobilizer_enabled: e.target.checked }))}
                        className="sr-only peer"
                      />
                      <div className="w-9 h-5 bg-border-color peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-primary/30 rounded-full peer peer-checked:after:translate-x-full after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-primary"></div>
                    </label>
                  </div>
                  <div className="flex items-center justify-between p-4 bg-surface-darker rounded-lg border border-border-color">
                    <div>
                      <div className="text-sm font-medium text-white">AIS-140 Compliant</div>
                      <div className="text-[10px] text-text-muted mt-0.5">Indian government telematics standard</div>
                    </div>
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input
                        type="checkbox"
                        checked={form.ais140_compliant}
                        onChange={(e) => setForm((prev) => ({ ...prev, ais140_compliant: e.target.checked }))}
                        className="sr-only peer"
                      />
                      <div className="w-9 h-5 bg-border-color peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-primary/30 rounded-full peer peer-checked:after:translate-x-full after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-primary"></div>
                    </label>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* ── Actions ── */}
          <div className="flex items-center justify-between border-t border-border-color pt-6">
            <button
              type="button"
              onClick={() => navigate("/admin/vehicles")}
              className="px-6 py-2.5 text-sm text-text-muted hover:text-white font-medium transition-colors border border-border-color rounded-lg hover:bg-border-color"
            >
              Cancel
            </button>

            <div className="flex items-center gap-3">
              {step > 1 && (
                <button
                  type="button"
                  onClick={() => setStep((s) => (s - 1) as 1 | 2 | 3)}
                  className="px-6 py-2.5 text-sm text-text-muted hover:text-white font-medium transition-colors"
                >
                  <span className="flex items-center gap-2">
                    <span className="material-symbols-outlined text-sm">chevron_left</span>
                    Previous
                  </span>
                </button>
              )}
              {step < 3 ? (
                <button
                  type="button"
                  onClick={handleNext}
                  className="bg-primary hover:bg-primary-dark text-white px-8 py-2.5 rounded-lg text-sm font-medium transition-colors shadow-lg shadow-primary/20"
                >
                  <span className="flex items-center gap-2">
                    Next
                    <span className="material-symbols-outlined text-sm">chevron_right</span>
                  </span>
                </button>
              ) : (
                <button
                  type="submit"
                  disabled={loading}
                  className="bg-primary hover:bg-primary-dark text-white px-8 py-2.5 rounded-lg text-sm font-medium transition-colors shadow-lg shadow-primary/20 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
                >
                  {loading ? (
                    <span className="animate-spin material-symbols-outlined text-sm">
                      progress_activity
                    </span>
                  ) : (
                    <span className="material-symbols-outlined text-sm">check</span>
                  )}
                  Create Vehicle
                </button>
              )}
            </div>
          </div>
        </form>
      </div>
    </div>
  );
};

export default AddVehicle;
