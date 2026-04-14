import React, { useEffect, useState, useCallback } from "react";
import { useNavigate, useParams } from "react-router-dom";
import apiClient from "../../../services/apiClient";

const FUEL_TYPES = [
  { value: "DIESEL", label: "Diesel" },
  { value: "PETROL", label: "Gasoline" },
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

const STATUS_OPTIONS = [
  { value: "ACTIVE", label: "Active", color: "text-success" },
  { value: "MAINTENANCE", label: "Maintenance", color: "text-warning" },
  { value: "SUSPENDED", label: "Out of Service", color: "text-danger" },
  { value: "DECOMMISSIONED", label: "Decommissioned", color: "text-text-muted" },
];

interface FormState {
  make: string;
  model: string;
  year: string;
  vin_number: string;
  fuel_type: string;
  vehicle_type: string;
  ownership_type: string;
  capacity_tonnage: string;
  fuel_tank_capacity: string;
  gps_device_id: string;
  color: string;
  chassis_number: string;
  engine_number: string;
  parent_vehicle_id: string;
  branch_id: string;
  cost_center_id: string;
  immobilizer_enabled: boolean;
  ais140_compliant: boolean;
  status: string;
  notes: string;
}

const toTonnes = (val: number, unit: string) => unit === "lbs" ? val / 2204.62 : unit === "kg" ? val / 1000 : val;
const toLitres = (val: number, unit: string) => unit === "gallons" ? val * 3.78541 : val;

export const EditVehicle: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [form, setForm] = useState<FormState>({
    make: "", model: "", year: "", vin_number: "", fuel_type: "DIESEL",
    vehicle_type: "", ownership_type: "OWNED", capacity_tonnage: "",
    fuel_tank_capacity: "", gps_device_id: "", color: "", chassis_number: "",
    engine_number: "", parent_vehicle_id: "", branch_id: "", cost_center_id: "",
    immobilizer_enabled: false, ais140_compliant: false, status: "ACTIVE", notes: "",
  });
  const [units, setUnits] = useState({ capacity: "tonnes", fuel: "litres" });
  const [saving, setSaving] = useState(false);
  const [submitError, setSubmitError] = useState("");
  const [loading, setLoading] = useState(true);
  const [confirmDelete, setConfirmDelete] = useState(false);
  const [vehicle, setVehicle] = useState<any>(null);
  const [charCount, setCharCount] = useState(0);

  useEffect(() => { fetchVehicle(); }, [id]);

  const fetchVehicle = async () => {
    try {
      const res = await apiClient.get(`/admin/vehicles/${id}`);
      const v = res.data;
      setVehicle(v);
      setForm({
        make: v.make || "", model: v.model || "", year: v.year ? v.year.toString() : "",
        vin_number: v.vin_number || "", fuel_type: v.fuel_type || "DIESEL",
        vehicle_type: v.vehicle_type || "", ownership_type: v.ownership_type || "OWNED",
        capacity_tonnage: v.capacity_tonnage ? v.capacity_tonnage.toString() : "",
        fuel_tank_capacity: v.fuel_tank_capacity ? v.fuel_tank_capacity.toString() : "",
        gps_device_id: v.gps_device_id || "", color: v.color || "",
        chassis_number: v.chassis_number || "", engine_number: v.engine_number || "",
        parent_vehicle_id: v.parent_vehicle_id || "", branch_id: v.branch_id || "",
        cost_center_id: v.cost_center_id || "",
        immobilizer_enabled: v.immobilizer_enabled || false,
        ais140_compliant: v.ais140_compliant || false,
        status: v.status || "ACTIVE", notes: v.notes || "",
      });
      setCharCount((v.notes || "").length);
      if (v.fuel_type === "CNG") setUnits((p) => ({ ...p, fuel: "kg" }));
    } catch (err) {
      console.error("Failed to fetch vehicle:", err);
    } finally {
      setLoading(false);
    }
  };

  const handleChange = useCallback((e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
    const { name, value } = e.target;
    setForm((prev) => ({ ...prev, [name]: value }));
    if (name === "notes") setCharCount(value.length);
  }, []);

  const handleToggle = useCallback((field: keyof FormState) => {
    setForm((prev) => ({ ...prev, [field]: !prev[field] }));
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitError("");
    setSaving(true);

    const payload: any = {
      vehicle_type: form.vehicle_type, fuel_type: form.fuel_type,
      ownership_type: form.ownership_type, status: form.status,
    };
    if (form.make) payload.make = form.make.trim();
    if (form.model) payload.model = form.model.trim();
    if (form.year) payload.year = parseInt(form.year);
    if (form.vin_number) payload.vin_number = form.vin_number.trim().toUpperCase();
    if (form.chassis_number) payload.chassis_number = form.chassis_number.trim();
    if (form.engine_number) payload.engine_number = form.engine_number.trim();
    if (form.capacity_tonnage) payload.capacity_tonnage = toTonnes(parseFloat(form.capacity_tonnage), units.capacity);
    if (form.fuel_tank_capacity) {
      payload.fuel_tank_capacity = form.fuel_type === "CNG"
        ? parseFloat(form.fuel_tank_capacity)
        : toLitres(parseFloat(form.fuel_tank_capacity), units.fuel);
    }
    if (form.gps_device_id) payload.gps_device_id = form.gps_device_id.trim();
    if (form.color) payload.color = form.color.trim();
    if (form.branch_id) payload.branch_id = form.branch_id.trim();
    payload.immobilizer_enabled = form.immobilizer_enabled;
    payload.ais140_compliant = form.ais140_compliant;

    try {
      await apiClient.put(`/admin/vehicles/${id}`, payload);
      navigate(`/admin/vehicles/${id}`);
    } catch (err: any) {
      const msg = err.response?.data?.error || err.response?.data?.message || err.message;
      setSubmitError(msg || "Failed to update vehicle.");
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!confirmDelete) { setConfirmDelete(true); return; }
    try {
      await apiClient.delete(`/admin/vehicles/${id}`);
      navigate("/admin/vehicles");
    } catch (err: any) {
      setSubmitError(err.response?.data?.error || "Failed to delete vehicle");
    }
  };

  if (loading) return (
    <div className="flex-1 flex items-center justify-center bg-background-dark">
      <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
    </div>
  );

  return (
    <div className="flex-1 flex flex-col min-w-0 bg-background-dark overflow-hidden">
      {/* Header */}
      <header className="h-16 flex items-center justify-between px-6 border-b border-border-color bg-surface-darker flex-shrink-0 z-20">
        <div className="flex flex-col">
          <nav className="flex items-center text-xs text-text-muted mb-0.5">
            <span className="hover:text-white cursor-pointer transition-colors" onClick={() => navigate("/admin/vehicles")}>Vehicles</span>
            <span className="material-symbols-outlined text-[14px] mx-0.5">chevron_right</span>
            <span className="hover:text-white cursor-pointer transition-colors" onClick={() => navigate("/admin/vehicles")}>Fleet List</span>
            <span className="material-symbols-outlined text-[14px] mx-0.5">chevron_right</span>
            <span className="text-white font-medium">Edit Vehicle</span>
          </nav>
          <h2 className="text-white text-lg font-bold">
            Edit Vehicle: {vehicle?.make} {vehicle?.model} #{vehicle?.registration_number?.slice(-4) || vehicle?.id?.slice(-4)}
          </h2>
        </div>
        <div className="flex items-center gap-4">
          <button onClick={() => navigate(`/admin/vehicles/${id}`)} className="px-4 py-2 border border-border-color hover:bg-surface-dark text-text-muted hover:text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors">
            Cancel
          </button>
          <button type="submit" form="edit-vehicle-form" disabled={saving} className="px-4 py-2 bg-primary hover:bg-primary-dark text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors shadow-lg shadow-primary/20 disabled:opacity-50">
            <span className="material-symbols-outlined text-[16px]">save</span>
            Save Changes
          </button>
        </div>
      </header>

      {/* Form */}
      <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
        <form id="edit-vehicle-form" onSubmit={handleSubmit} className="max-w-5xl mx-auto space-y-6">
          {submitError && (
            <div className="rounded-lg border border-danger/30 bg-danger/10 px-4 py-3 text-sm text-danger flex items-center gap-2">
              <span className="material-symbols-outlined text-sm">error</span>{submitError}
            </div>
          )}

          {/* Vehicle Information */}
          <div className="bg-surface-dark rounded-xl border border-border-color p-6">
            <div className="flex items-center gap-3 mb-6 pb-4 border-b border-border-color">
              <div className="w-8 h-8 rounded bg-primary/10 flex items-center justify-center text-primary">
                <span className="material-symbols-outlined">directions_car</span>
              </div>
              <h3 className="text-white font-semibold text-sm">Vehicle Information</h3>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              <Field label="Make">
                <input type="text" name="make" value={form.make} onChange={handleChange} className="w-full bg-surface-darker border border-border-color rounded px-3 py-2 text-sm text-white focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary placeholder-gray-600 transition-shadow" />
              </Field>
              <Field label="Model">
                <input type="text" name="model" value={form.model} onChange={handleChange} className="w-full bg-surface-darker border border-border-color rounded px-3 py-2 text-sm text-white focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary placeholder-gray-600 transition-shadow" />
              </Field>
              <Field label="Year">
                <input type="number" name="year" value={form.year} onChange={handleChange} className="w-full bg-surface-darker border border-border-color rounded px-3 py-2 text-sm text-white focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary transition-shadow" />
              </Field>
              <Field label="VIN">
                <input type="text" name="vin_number" value={form.vin_number} onChange={handleChange} maxLength={17} className="w-full bg-surface-darker border border-border-color rounded px-3 py-2 text-sm text-white focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary placeholder-gray-600 font-mono transition-shadow" />
              </Field>
              <Field label="Registration Number">
                <input type="text" value={vehicle?.registration_number || ""} disabled className="w-full bg-surface-darker/50 border border-border-color rounded px-3 py-2 text-sm text-text-muted focus:outline-none cursor-not-allowed uppercase transition-shadow" />
              </Field>
              <Field label="Status">
                <select name="status" value={form.status} onChange={handleChange} className="w-full bg-surface-darker border border-border-color rounded px-3 py-2 text-sm text-white focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary transition-shadow">
                  {STATUS_OPTIONS.map((s) => (<option key={s.value} value={s.value} className={s.color}>{s.label}</option>))}
                </select>
              </Field>
            </div>
          </div>

          {/* Operational Details */}
          <div className="bg-surface-dark rounded-xl border border-border-color p-6">
            <div className="flex items-center gap-3 mb-6 pb-4 border-b border-border-color">
              <div className="w-8 h-8 rounded bg-primary/10 flex items-center justify-center text-primary">
                <span className="material-symbols-outlined">settings_suggest</span>
              </div>
              <h3 className="text-white font-semibold text-sm">Operational Details</h3>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {/* Vehicle Type */}
              <Field label="Vehicle Type">
                <select name="vehicle_type" value={form.vehicle_type} onChange={handleChange} className="w-full bg-surface-darker border border-border-color rounded px-3 py-2 text-sm text-white focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary transition-shadow">
                  <option value="">Select Type</option>
                  {VEHICLE_TYPES.map((t) => (<option key={t.value} value={t.value}>{t.label}</option>))}
                </select>
              </Field>
              {/* Ownership */}
              <Field label="Ownership Type">
                <select name="ownership_type" value={form.ownership_type} onChange={handleChange} className="w-full bg-surface-darker border border-border-color rounded px-3 py-2 text-sm text-white focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary transition-shadow">
                  {OWNERSHIP_TYPES.map((o) => (<option key={o.value} value={o.value}>{o.label}</option>))}
                </select>
              </Field>
              {/* Capacity */}
              <div className="space-y-1.5">
                <label className="text-xs font-medium text-text-muted">Capacity</label>
                <div className="flex gap-2">
                  <input type="number" name="capacity_tonnage" value={form.capacity_tonnage} onChange={handleChange} className="flex-1 bg-surface-darker border border-border-color rounded px-3 py-2 text-sm text-white focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary placeholder-gray-600 transition-shadow" />
                  <select value={units.capacity} onChange={(e) => setUnits((p) => ({ ...p, capacity: e.target.value }))} className="bg-surface-darker border border-border-color rounded px-3 py-2 text-sm text-white focus:outline-none focus:ring-1 focus:ring-primary">
                    <option value="tonnes">Tonnes</option>
                    <option value="kg">kg</option>
                    <option value="lbs">lbs</option>
                  </select>
                </div>
              </div>
              {/* Fuel Type */}
              <Field label="Fuel Type">
                <div className="flex gap-4 pt-1">
                  {FUEL_TYPES.filter((f) => f.value !== "LPG").map((f) => (
                    <label key={f.value} className="flex items-center gap-2 cursor-pointer">
                      <input type="radio" name="fuel_type" value={f.value} checked={form.fuel_type === f.value} onChange={handleChange} className="text-primary bg-surface-darker border-border-color focus:ring-primary focus:ring-offset-surface-dark" />
                      <span className="text-sm text-white">{f.label}</span>
                    </label>
                  ))}
                </div>
              </Field>
              {/* Fuel Tank / Cylinder */}
              <div className="space-y-1.5">
                <label className="text-xs font-medium text-text-muted">{form.fuel_type === "CNG" ? "CNG Cylinder (kg)" : "Fuel Tank Capacity"}</label>
                <div className="flex gap-2">
                  <input type="number" name="fuel_tank_capacity" value={form.fuel_tank_capacity} onChange={handleChange} placeholder={form.fuel_type === "CNG" ? "e.g. 12" : "e.g. 200"} className="flex-1 bg-surface-darker border border-border-color rounded px-3 py-2 text-sm text-white focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary placeholder-gray-600 transition-shadow" />
                  {form.fuel_type !== "CNG" && (
                    <select value={units.fuel} onChange={(e) => setUnits((p) => ({ ...p, fuel: e.target.value }))} className="bg-surface-darker border border-border-color rounded px-3 py-2 text-sm text-white focus:outline-none focus:ring-1 focus:ring-primary">
                      <option value="litres">Litres</option>
                      <option value="gallons">Gallons</option>
                    </select>
                  )}
                </div>
              </div>
              {/* GPS Device */}
              <Field label="GPS Device ID">
                <input type="text" name="gps_device_id" value={form.gps_device_id} onChange={handleChange} placeholder="e.g. GPS-IMEI-12345678" className="w-full bg-surface-darker border border-border-color rounded px-3 py-2 text-sm text-white focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary placeholder-gray-600 font-mono transition-shadow" />
              </Field>
            </div>
          </div>

          {/* Notes + Metadata */}
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            {/* Notes */}
            <div className="lg:col-span-2 bg-surface-dark rounded-xl border border-border-color p-6">
              <div className="flex items-center gap-3 mb-4">
                <div className="w-8 h-8 rounded bg-primary/10 flex items-center justify-center text-primary">
                  <span className="material-symbols-outlined">description</span>
                </div>
                <h3 className="text-white font-semibold text-sm">Notes</h3>
              </div>
              <div className="space-y-1.5">
                <textarea name="notes" value={form.notes} onChange={handleChange} rows={4} maxLength={500} className="w-full bg-surface-darker border border-border-color rounded px-3 py-2 text-sm text-white focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary placeholder-gray-600 transition-shadow resize-none" />
                <p className="text-[10px] text-text-muted text-right">{charCount}/500 characters</p>
              </div>
            </div>

            {/* System Metadata */}
            <div className="lg:col-span-1 bg-surface-dark rounded-xl border border-border-color p-6">
              <div className="flex items-center gap-3 mb-4">
                <div className="w-8 h-8 rounded bg-primary/10 flex items-center justify-center text-primary">
                  <span className="material-symbols-outlined">info</span>
                </div>
                <h3 className="text-white font-semibold text-sm">System Metadata</h3>
              </div>
              <div className="space-y-4">
                <div className="flex justify-between items-center border-b border-border-color pb-2">
                  <span className="text-xs text-text-muted">Vehicle ID</span>
                  <span className="text-xs font-mono text-white">#{vehicle?.id?.slice(-6).toUpperCase()}</span>
                </div>
                <div className="flex justify-between items-center border-b border-border-color pb-2">
                  <span className="text-xs text-text-muted">Registration</span>
                  <span className="text-xs font-mono text-white">{vehicle?.registration_number || "—"}</span>
                </div>
                <div className="flex justify-between items-center border-b border-border-color pb-2">
                  <span className="text-xs text-text-muted">Created Date</span>
                  <span className="text-xs text-white">{vehicle?.created_at ? new Date(vehicle.created_at).toLocaleDateString() : "—"}</span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-xs text-text-muted">Last Updated</span>
                  <span className="text-xs text-white">{vehicle?.updated_at ? new Date(vehicle.updated_at).toLocaleString() : "—"}</span>
                </div>
                <div className="pt-4 mt-2">
                  <button type="button" onClick={handleDelete} className={`w-full py-2 rounded text-xs font-semibold transition-colors flex items-center justify-center gap-2 ${
                    confirmDelete
                      ? "bg-danger hover:bg-red-600 text-white"
                      : "border border-danger/30 text-danger hover:bg-danger/10"
                  }`}>
                    <span className="material-symbols-outlined text-[16px]">{confirmDelete ? "delete_forever" : "delete"}</span>
                    {confirmDelete ? "Confirm Delete" : "Delete Vehicle"}
                  </button>
                </div>
              </div>
            </div>
          </div>
        </form>
      </div>

      {/* Footer */}
      <footer className="bg-surface-darker border-t border-border-color px-6 py-2 flex items-center justify-between text-[10px] text-text-muted z-20 flex-shrink-0">
        <span>© {new Date().getFullYear()} Enterprise Fleet Manager. Vehicle Module V2.1</span>
        <div className="flex items-center gap-4">
          <span>Data Sync: <span className="text-success">Active</span></span>
          <span>Server Time: <span className="font-mono text-white">{new Date().toUTCString().slice(17, 25)} UTC</span></span>
        </div>
      </footer>
    </div>
  );
};

const Field: React.FC<{ label: string; children: React.ReactNode }> = ({ label, children }) => (
  <div className="space-y-1.5">
    <label className="text-xs font-medium text-text-muted">{label}</label>
    {children}
  </div>
);

export default EditVehicle;
