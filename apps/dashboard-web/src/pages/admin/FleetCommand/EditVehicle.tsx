import React, { useEffect, useState, useCallback } from "react";
import { useNavigate, useParams } from "react-router-dom";
import apiClient from "../../../services/apiClient";

// ── Reuse AddVehicle's constants ──
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

const getFuelUnits = (fuelType: string) => {
  if (fuelType === "CNG") return [{ value: "kg", label: "Kilograms (kg)" }];
  return FUEL_UNITS;
};

// Conversions
const toKm = (val: number, unit: string) => unit === "mi" ? val * 1.60934 : val;
const toTonnes = (val: number, unit: string) => unit === "lbs" ? val / 2204.62 : unit === "kg" ? val / 1000 : val;
const toLitres = (val: number, unit: string) => unit === "gallons" ? val * 3.78541 : val;

// Reverse conversions (metric → display)
const fromKm = (val: number, unit: string) => unit === "mi" ? val / 1.60934 : val;
const fromTonnes = (val: number, unit: string) => unit === "lbs" ? val * 2204.62 : unit === "kg" ? val * 1000 : val;
const fromLitres = (val: number, unit: string) => unit === "gallons" ? val / 3.78541 : val;

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
  status: string;
}

const initialForm: FormState = {
  registration_number: "",
  vehicle_type: "",
  fuel_type: "DIESEL",
  ownership_type: "OWNED",
  make: "",
  model: "",
  year: "",
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
  status: "ACTIVE",
};

export const EditVehicle: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [form, setForm] = useState<FormState>(initialForm);
  const [units, setUnits] = useState({
    odometer: "km",
    capacity: "tonnes",
    fuel: "litres",
  });
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [touched, setTouched] = useState<Record<string, boolean>>({});
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [submitError, setSubmitError] = useState("");
  const [confirmDelete, setConfirmDelete] = useState(false);

  useEffect(() => {
    fetchVehicle();
  }, [id]);

  const fetchVehicle = async () => {
    try {
      const res = await apiClient.get(`/admin/vehicles/${id}`);
      const v = res.data;
      setForm({
        registration_number: v.registration_number || "",
        vehicle_type: v.vehicle_type || "",
        fuel_type: v.fuel_type || "DIESEL",
        ownership_type: v.ownership_type || "OWNED",
        make: v.make || "",
        model: v.model || "",
        year: v.year ? v.year.toString() : "",
        vin_number: v.vin_number || "",
        chassis_number: v.chassis_number || "",
        engine_number: v.engine_number || "",
        capacity_tonnage: v.capacity_tonnage ? v.capacity_tonnage.toString() : "",
        fuel_tank_capacity: v.fuel_tank_capacity ? v.fuel_tank_capacity.toString() : "",
        gps_device_id: v.gps_device_id || "",
        color: v.color || "",
        parent_vehicle_id: v.parent_vehicle_id || "",
        branch_id: v.branch_id || "",
        cost_center_id: v.cost_center_id || "",
        immobilizer_enabled: v.immobilizer_enabled || false,
        ais140_compliant: v.ais140_compliant || false,
        status: v.status || "ACTIVE",
      });

      // Set unit defaults based on fuel type
      if (v.fuel_type === "CNG") {
        setUnits((prev) => ({ ...prev, fuel: "kg" }));
      }
    } catch (error) {
      console.error("Failed to fetch vehicle:", error);
    } finally {
      setLoading(false);
    }
  };

  const handleChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
      const { name, value } = e.target;
      setForm((prev) => ({ ...prev, [name]: value }));
      if (touched[name]) {
        const error = validateField(name as keyof FormState, value);
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

  const handleToggle = useCallback((field: keyof FormState) => {
    setForm((prev) => ({ ...prev, [field]: !prev[field] }));
  }, []);

  const handleBlur = useCallback(
    (fieldName: keyof FormState) => {
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

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitError("");

    const payload: any = {
      vehicle_type: form.vehicle_type,
      fuel_type: form.fuel_type,
      ownership_type: form.ownership_type,
      status: form.status,
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

    setSaving(true);
    try {
      await apiClient.put(`/admin/vehicles/${id}`, payload);
      navigate(`/admin/vehicles/${id}`);
    } catch (err: any) {
      console.error("Failed to update vehicle:", err);
      const msg = err.response?.data?.error || err.response?.data?.message || err.message;
      setSubmitError(msg || "Failed to update vehicle. Please try again.");
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!confirmDelete) {
      setConfirmDelete(true);
      return;
    }
    try {
      await apiClient.delete(`/admin/vehicles/${id}`);
      navigate("/admin/vehicles");
    } catch (err: any) {
      console.error("Failed to delete vehicle:", err);
      setSubmitError(err.response?.data?.error || "Failed to delete vehicle");
    }
  };

  if (loading) {
    return (
      <div className="flex-1 flex items-center justify-center bg-background-dark">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
      </div>
    );
  }

  return (
    <div className="flex-1 flex flex-col min-w-0 overflow-y-auto bg-background-dark">
      {/* Header */}
      <header className="h-16 flex items-center justify-between px-6 border-b border-border-color bg-surface-darker sticky top-0 z-10 flex-shrink-0">
        <div className="flex items-center gap-4">
          <button
            onClick={() => navigate(`/admin/vehicles/${id}`)}
            className="text-text-muted hover:text-white transition-colors"
          >
            <span className="material-symbols-outlined">arrow_back</span>
          </button>
          <div>
            <h1 className="text-white text-lg font-bold">Edit Vehicle</h1>
            <p className="text-[11px] text-text-muted mt-0.5">
              {form.registration_number || form.make} {form.model || ""}
            </p>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <button
            type="button"
            onClick={handleDelete}
            className={`px-4 py-2 text-xs font-semibold rounded flex items-center gap-2 transition-colors ${
              confirmDelete
                ? "bg-danger hover:bg-red-600 text-white"
                : "text-text-muted hover:text-danger hover:bg-danger/10"
            }`}
          >
            <span className="material-symbols-outlined text-[16px]">{confirmDelete ? "delete_forever" : "delete"}</span>
            {confirmDelete ? "Confirm Delete" : "Delete"}
          </button>
          <button
            type="submit"
            form="edit-vehicle-form"
            disabled={saving}
            className="px-4 py-2 bg-primary hover:bg-primary-dark text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors shadow-lg shadow-primary/20 disabled:opacity-50"
          >
            {saving ? (
              <span className="animate-spin material-symbols-outlined text-sm">progress_activity</span>
            ) : (
              <span className="material-symbols-outlined text-[16px]">save</span>
            )}
            Save Changes
          </button>
        </div>
      </header>

      {/* Form */}
      <div className="flex-1 overflow-y-auto custom-scrollbar">
        <form id="edit-vehicle-form" onSubmit={handleSubmit} className="max-w-4xl mx-auto w-full p-6 space-y-6" noValidate>
          {submitError && (
            <div className="rounded-lg border border-danger/30 bg-danger/10 px-4 py-3 text-sm text-danger flex items-center gap-2">
              <span className="material-symbols-outlined text-sm">error</span>
              {submitError}
            </div>
          )}

          {/* Registration & Status */}
          <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
            <div className="px-6 py-4 border-b border-border-color bg-surface-darker flex items-center gap-3">
              <div className="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center text-primary border border-primary/20">
                <span className="material-symbols-outlined text-sm">badge</span>
              </div>
              <div>
                <h3 className="text-white text-sm font-semibold">Registration & Status</h3>
                <p className="text-xs text-text-muted mt-0.5">Vehicle identification and lifecycle status</p>
              </div>
            </div>
            <div className="p-6 space-y-6">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <InputField
                  name="registration_number"
                  value={form.registration_number}
                  onChange={handleChange}
                  label="Registration Number"
                  disabled
                  hint="Cannot be changed after creation"
                />
                <SelectField
                  name="status"
                  value={form.status}
                  onChange={handleChange}
                  label="Status"
                  options={[
                    { value: "ACTIVE", label: "Active" },
                    { value: "MAINTENANCE", label: "Maintenance" },
                    { value: "SUSPENDED", label: "Suspended" },
                    { value: "DECOMMISSIONED", label: "Decommissioned" },
                  ]}
                  icon="toggle_on"
                />
              </div>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <SelectField
                  name="fuel_type"
                  value={form.fuel_type}
                  onChange={handleChange}
                  label="Fuel Type"
                  options={FUEL_TYPES}
                  icon="local_gas_station"
                />
                <SelectField
                  name="ownership_type"
                  value={form.ownership_type}
                  onChange={handleChange}
                  label="Ownership Type"
                  options={OWNERSHIP_TYPES}
                  icon="home_work"
                />
              </div>
            </div>
          </div>

          {/* Specifications */}
          <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
            <div className="px-6 py-4 border-b border-border-color bg-surface-darker flex items-center gap-3">
              <div className="w-8 h-8 rounded-full bg-purple-500/10 flex items-center justify-center text-purple-400 border border-purple-500/20">
                <span className="material-symbols-outlined text-sm">directions_car</span>
              </div>
              <div>
                <h3 className="text-white text-sm font-semibold">Specifications</h3>
                <p className="text-xs text-text-muted mt-0.5">Make, model, capacity, and identification numbers</p>
              </div>
            </div>
            <div className="p-6 space-y-6">
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <InputField name="make" value={form.make} onChange={handleChange} label="Make" placeholder="e.g. Tata" icon="business" />
                <InputField name="model" value={form.model} onChange={handleChange} label="Model" placeholder="e.g. 407" />
                <InputField name="year" value={form.year} onChange={handleChange} label="Year" type="number" />
              </div>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <InputField name="vin_number" value={form.vin_number} onChange={handleChange} label="VIN Number" placeholder="17-character VIN" maxLength={17} hint="Must be exactly 17 characters" />
                <InputField name="chassis_number" value={form.chassis_number} onChange={handleChange} label="Chassis Number" maxLength={50} />
                <InputField name="engine_number" value={form.engine_number} onChange={handleChange} label="Engine Number" maxLength={50} />
              </div>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div className="space-y-1.5">
                  <label className="text-xs font-medium text-text-muted uppercase tracking-wide">Capacity</label>
                  <div className="flex gap-2">
                    <input type="number" name="capacity_tonnage" value={form.capacity_tonnage} onChange={handleChange} placeholder="e.g. 10" className="flex-1 bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder-text-muted/50" />
                    <select value={units.capacity} onChange={(e) => setUnits((prev) => ({ ...prev, capacity: e.target.value }))} className="bg-surface-darker border border-border-color rounded-lg px-3 py-2.5 text-sm text-white focus:ring-1 focus:ring-primary outline-none">
                      {WEIGHT_UNITS.map((u) => (<option key={u.value} value={u.value}>{u.label}</option>))}
                    </select>
                  </div>
                </div>
                <div className="space-y-1.5">
                  <label className="text-xs font-medium text-text-muted uppercase tracking-wide">{form.fuel_type === "CNG" ? "CNG Cylinder" : "Fuel Tank"}</label>
                  <div className="flex gap-2">
                    <input type="number" name="fuel_tank_capacity" value={form.fuel_tank_capacity} onChange={handleChange} placeholder={form.fuel_type === "CNG" ? "e.g. 12" : "e.g. 200"} className="flex-1 bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder-text-muted/50" />
                    <select value={units.fuel} onChange={(e) => setUnits((prev) => ({ ...prev, fuel: e.target.value }))} className="bg-surface-darker border border-border-color rounded-lg px-3 py-2.5 text-sm text-white focus:ring-1 focus:ring-primary outline-none">
                      {getFuelUnits(form.fuel_type).map((u) => (<option key={u.value} value={u.value}>{u.label}</option>))}
                    </select>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Additional */}
          <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
            <div className="px-6 py-4 border-b border-border-color bg-surface-darker flex items-center gap-3">
              <div className="w-8 h-8 rounded-full bg-amber-500/10 flex items-center justify-center text-amber-400 border border-amber-500/20">
                <span className="material-symbols-outlined text-sm">settings</span>
              </div>
              <div>
                <h3 className="text-white text-sm font-semibold">Additional Details</h3>
                <p className="text-xs text-text-muted mt-0.5">GPS device, hierarchy, and compliance</p>
              </div>
            </div>
            <div className="p-6 space-y-6">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <InputField name="gps_device_id" value={form.gps_device_id} onChange={handleChange} label="GPS Device ID" placeholder="e.g. GPS-IMEI-12345678" icon="gps_fixed" hint="AIS-140 compliant GPS tracker IMEI" />
                <InputField name="color" value={form.color} onChange={handleChange} label="Color" placeholder="e.g. White" />
              </div>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <InputField name="parent_vehicle_id" value={form.parent_vehicle_id} onChange={handleChange} label="Parent Vehicle ID" placeholder="UUID" icon="account_tree" hint="For hierarchical grouping" />
                <InputField name="branch_id" value={form.branch_id} onChange={handleChange} label="Branch ID" placeholder="UUID" icon="account_balance" hint="Depot/branch assignment" />
              </div>
              <InputField name="cost_center_id" value={form.cost_center_id} onChange={handleChange} label="Cost Center ID" placeholder="UUID" icon="account_balance_wallet" hint="Financial tracking" />
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <ToggleField label="Immobilizer" description="Remote engine cut-off" checked={form.immobilizer_enabled} onChange={() => handleToggle("immobilizer_enabled")} />
                <ToggleField label="AIS-140 Compliant" description="Indian govt telematics standard" checked={form.ais140_compliant} onChange={() => handleToggle("ais140_compliant")} />
              </div>
            </div>
          </div>
        </form>
      </div>
    </div>
  );
};

// ── Sub-components ──
const InputField: React.FC<{
  name: string; value: string; onChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  label: string; placeholder?: string; type?: string; icon?: string; hint?: string;
  disabled?: boolean; maxLength?: number;
}> = ({ name, value, onChange, label, placeholder, type = "text", icon, hint, disabled, maxLength }) => (
  <div className="space-y-1.5">
    <label className="text-xs font-medium text-text-muted uppercase tracking-wide">{label}</label>
    <div className="relative">
      {icon && <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted"><span className="material-symbols-outlined text-[18px]">{icon}</span></span>}
      <input
        type={type} name={name} value={value} onChange={onChange} placeholder={placeholder}
        maxLength={maxLength} disabled={disabled}
        className={`w-full bg-surface-darker border border-border-color rounded-lg py-2.5 ${icon ? "pl-10" : "px-4"} pr-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder-text-muted/50 ${disabled ? "opacity-50 cursor-not-allowed" : ""}`}
      />
    </div>
    {hint && <p className="text-[10px] text-text-muted">{hint}</p>}
  </div>
);

const SelectField: React.FC<{
  name: string; value: string; onChange: (e: React.ChangeEvent<HTMLSelectElement>) => void;
  label: string; options: { value: string; label: string }[]; icon?: string;
}> = ({ name, value, onChange, label, options, icon }) => (
  <div className="space-y-1.5">
    <label className="text-xs font-medium text-text-muted uppercase tracking-wide">{label}</label>
    <div className="relative">
      {icon && <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted"><span className="material-symbols-outlined text-[18px]">{icon}</span></span>}
      <select
        name={name} value={value} onChange={onChange}
        className={`w-full bg-surface-darker border border-border-color rounded-lg py-2.5 ${icon ? "pl-10" : "pl-4"} pr-10 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none appearance-none`}
      >
        {options.map((o) => (<option key={o.value} value={o.value}>{o.label}</option>))}
      </select>
      <span className="absolute inset-y-0 right-0 flex items-center pr-3 pointer-events-none text-text-muted"><span className="material-symbols-outlined text-[18px]">expand_more</span></span>
    </div>
  </div>
);

const ToggleField: React.FC<{ label: string; description: string; checked: boolean; onChange: () => void }> = ({ label, description, checked, onChange }) => (
  <div className="flex items-center justify-between p-4 bg-surface-darker rounded-lg border border-border-color">
    <div>
      <div className="text-sm font-medium text-white">{label}</div>
      <div className="text-[10px] text-text-muted mt-0.5">{description}</div>
    </div>
    <label className="relative inline-flex items-center cursor-pointer">
      <input type="checkbox" checked={checked} onChange={onChange} className="sr-only peer" />
      <div className="w-9 h-5 bg-border-color peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-primary/30 rounded-full peer peer-checked:after:translate-x-full after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-primary"></div>
    </label>
  </div>
);

export default EditVehicle;
