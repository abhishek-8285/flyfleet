import React, { useEffect, useState, useCallback } from "react";
import { useNavigate, useParams } from "react-router-dom";
import apiClient from "../../../services/apiClient";
import { 
  FUEL_TYPES, VEHICLE_TYPES, OWNERSHIP_TYPES, 
  STATUS_OPTIONS, WEIGHT_UNITS, getFuelUnits 
} from "./constants";
import { toKm, toTonnes, toLitres } from "./utils";
import { TextInput, SelectInput, ToggleInput } from "../../../components/common/FormFields";

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
  current_driver_id: string;
  immobilizer_enabled: boolean;
  ais140_compliant: boolean;
  status: string;
  notes: string;
  odometer: string;
  odometer_unit: string;
  speed_limit: string;
}

export const EditVehicle: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [form, setForm] = useState<FormState>({
    make: "", model: "", year: "", vin_number: "", fuel_type: "DIESEL",
    vehicle_type: "", ownership_type: "OWNED", capacity_tonnage: "",
    fuel_tank_capacity: "", gps_device_id: "", color: "", chassis_number: "",
    engine_number: "", parent_vehicle_id: "", branch_id: "", cost_center_id: "",
    immobilizer_enabled: false, ais140_compliant: false, status: "ACTIVE", notes: "",
    odometer: "0", odometer_unit: "km", speed_limit: "0", current_driver_id: "",
  });
  const [units, setUnits] = useState({ capacity: "tonnes", fuel: "litres" });
  const [saving, setSaving] = useState(false);
  const [submitError, setSubmitError] = useState("");
  const [loading, setLoading] = useState(true);
  const [confirmDelete, setConfirmDelete] = useState(false);
  const [vehicle, setVehicle] = useState<any>(null);
  const [charCount, setCharCount] = useState(0);

  const [dropdowns, setDropdowns] = useState({
    vehicles: [] as { value: string; label: string }[],
    drivers: [] as { value: string; label: string }[],
    branches: [] as { value: string; label: string }[],
    costCenters: [] as { value: string; label: string }[],
  });

  useEffect(() => { 
    fetchVehicle(); 
    fetchDropdowns();
  }, [id]);

  const fetchDropdowns = async () => {
    try {
      const [vResp, dResp, bResp, cResp] = await Promise.all([
        apiClient.get("/admin/vehicles"),
        apiClient.get("/admin/drivers"),
        apiClient.get("/admin/branches"),
        apiClient.get("/admin/cost-centers")
      ]);
      setDropdowns(prev => ({
        ...prev,
        vehicles: (vResp.data?.vehicles || vResp.data || []).map((v: any) => ({
          value: v.id,
          label: `${v.registration_number || v.license_plate} (${v.make} ${v.model})`
        })),
        drivers: (dResp.data?.drivers || dResp.data || []).map((d: any) => ({
          value: d.id,
          label: `${d.first_name} ${d.last_name} (${d.employee_id})`
        })),
        branches: bResp.data?.data || bResp.data || [],
        costCenters: cResp.data?.data || cResp.data || []
      }));
    } catch (err) {
      console.error("Failed to fetch dropdown data:", err);
    }
  };

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
        current_driver_id: v.current_driver_id || "",
        immobilizer_enabled: v.immobilizer_enabled || false,
        ais140_compliant: v.ais140_compliant || false,
        status: v.status || "ACTIVE", notes: v.notes || "",
        odometer: v.odometer ? (v.odometer / 1000).toString() : "0",
        odometer_unit: "km",
        speed_limit: v.speed_limit ? v.speed_limit.toString() : "0",
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
    if (form.cost_center_id) payload.cost_center_id = form.cost_center_id.trim();
    if (form.current_driver_id) payload.current_driver_id = form.current_driver_id.trim();
    if (form.parent_vehicle_id) payload.parent_vehicle_id = form.parent_vehicle_id.trim();
    if (form.odometer) payload.odometer = Math.round(toKm(parseFloat(form.odometer), form.odometer_unit) * 1000);
    if (form.speed_limit) payload.speed_limit = parseFloat(form.speed_limit);
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
        <form id="edit-vehicle-form" onSubmit={handleSubmit} className="max-w-5xl mx-auto space-y-8">
          {submitError && (
            <div className="rounded-lg border border-danger/30 bg-danger/10 px-4 py-3 text-sm text-danger flex items-center gap-2 animate-shake">
              <span className="material-symbols-outlined text-sm">error</span>{submitError}
            </div>
          )}

          {/* Section 1: Basic Information */}
          <div className="bg-surface-dark rounded-xl border border-border-color p-6 shadow-xl relative overflow-hidden group">
            <div className="absolute top-0 right-0 p-4 opacity-5 group-hover:opacity-10 transition-opacity">
              <span className="material-symbols-outlined text-6xl text-white">badge</span>
            </div>
            <div className="flex items-center gap-3 mb-6 pb-4 border-b border-border-color">
              <div className="w-8 h-8 rounded bg-primary/10 flex items-center justify-center text-primary">
                <span className="material-symbols-outlined text-[18px]">directions_car</span>
              </div>
              <h3 className="text-white font-semibold text-sm">Vehicle Information</h3>
            </div>
            
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              <TextInput
                label="Make" name="make" value={form.make} onChange={handleChange}
                placeholder="e.g. Tata, Mahindra" icon="factory"
              />
              <TextInput
                label="Model" name="model" value={form.model} onChange={handleChange}
                placeholder="e.g. Prima, Furio" icon="model_training"
              />
              <TextInput
                label="Year" name="year" value={form.year} onChange={handleChange}
                placeholder="e.g. 2024" icon="calendar_today"
              />
              <TextInput
                label="VIN Number" name="vin_number" value={form.vin_number} onChange={handleChange}
                placeholder="17-digit chassis number" icon="fingerprint" maxLength={17}
              />
              <TextInput
                label="Registration" name="registration_number" value={vehicle?.registration_number || ""}
                onChange={() => {}} icon="license" className="opacity-70 pointer-events-none"
              />
              <SelectInput
                label="Operational Status" name="status" value={form.status} 
                onChange={handleChange} options={STATUS_OPTIONS} icon="traffic"
              />
            </div>
          </div>

          {/* Section 2: Hierarchy & Assignment */}
          <div className="bg-surface-dark rounded-xl border border-border-color p-6 shadow-xl relative overflow-hidden group">
            <div className="absolute top-0 right-0 p-4 opacity-5 group-hover:opacity-10 transition-opacity">
              <span className="material-symbols-outlined text-6xl text-white">account_tree</span>
            </div>
            <div className="flex items-center gap-3 mb-6 pb-4 border-b border-border-color">
              <div className="w-8 h-8 rounded bg-primary/10 flex items-center justify-center text-primary">
                <span className="material-symbols-outlined text-[18px]">hub</span>
              </div>
              <h3 className="text-white font-semibold text-sm">Hierarchy & Assignment</h3>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <SelectInput
                label="Branch / Depot" name="branch_id" value={form.branch_id}
                onChange={handleChange} options={dropdowns.branches} icon="location_on"
                placeholder="Select assigned branch"
              />
              <SelectInput
                label="Cost Center" name="cost_center_id" value={form.cost_center_id}
                onChange={handleChange} options={dropdowns.costCenters} icon="payments"
                placeholder="Select billing group"
              />
              <SelectInput
                label="Parent Asset (Trailer/Convoy)" name="parent_vehicle_id" value={form.parent_vehicle_id}
                onChange={handleChange} options={dropdowns.vehicles} icon="link"
                placeholder="Search vehicles..."
              />
              <SelectInput
                label="Primary Driver" name="current_driver_id" value={form.current_driver_id}
                onChange={handleChange} options={dropdowns.drivers} icon="person"
                placeholder="Assign a driver"
              />
              <TextInput
                label="GPS Device ID" name="gps_device_id" value={form.gps_device_id}
                onChange={handleChange} placeholder="IMEI or Serial Number" icon="gps_fixed"
              />
            </div>
          </div>

          <div className="bg-surface-dark rounded-xl border border-border-color p-6 shadow-xl space-y-6">
             <div className="flex items-center gap-3 mb-6 pb-4 border-b border-border-color">
              <div className="w-8 h-8 rounded bg-primary/10 flex items-center justify-center text-primary">
                <span className="material-symbols-outlined text-[18px]">speed</span>
              </div>
              <h3 className="text-white font-semibold text-sm">Operational Metrics</h3>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
               <div className="space-y-1.5 lg:col-span-1">
                <label className="text-xs font-medium text-text-muted uppercase tracking-wide">Current Odometer Reading</label>
                <div className="flex gap-2">
                  <TextInput
                    label="" name="odometer" value={form.odometer}
                    onChange={handleChange} type="number" placeholder="0"
                    className="flex-1"
                  />
                  <select 
                    name="odometer_unit"
                    value={form.odometer_unit} 
                    onChange={handleChange}
                    className="bg-surface-darker border border-border-color rounded-lg px-2 text-xs text-white focus:ring-1 focus:ring-primary outline-none"
                  >
                    <option value="km">km</option>
                    <option value="mi">mi</option>
                  </select>
                </div>
              </div>
              <TextInput
                label="Safety Speed Limit (km/h)" name="speed_limit" value={form.speed_limit}
                onChange={handleChange} type="number" placeholder="80" icon="speed"
              />
            </div>
          </div>

          {/* Section 3: Technical Specifications */}
          <div className="bg-surface-dark rounded-xl border border-border-color p-6 shadow-xl">
            <div className="flex items-center gap-3 mb-6 pb-4 border-b border-border-color">
              <div className="w-8 h-8 rounded bg-primary/10 flex items-center justify-center text-primary">
                <span className="material-symbols-outlined text-[18px]">settings_suggest</span>
              </div>
              <h3 className="text-white font-semibold text-sm">Technical Specifications</h3>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              <SelectInput
                label="Vehicle Body Type" name="vehicle_type" value={form.vehicle_type}
                onChange={handleChange} options={VEHICLE_TYPES} icon="local_shipping"
                placeholder="Select type"
              />
              <SelectInput
                label="Fuel Engine Type" name="fuel_type" value={form.fuel_type}
                onChange={handleChange} options={FUEL_TYPES} icon="local_gas_station"
              />
              <SelectInput
                label="Ownership Model" name="ownership_type" value={form.ownership_type}
                onChange={handleChange} options={OWNERSHIP_TYPES} icon="key"
              />

              <div className="space-y-1.5 lg:col-span-1">
                <label className="text-xs font-medium text-text-muted uppercase tracking-wide">Payload Capacity</label>
                <div className="flex gap-2">
                  <TextInput
                    label="" name="capacity_tonnage" value={form.capacity_tonnage}
                    onChange={handleChange} type="number" placeholder="0.00"
                    className="flex-1"
                  />
                  <select 
                    value={units.capacity} 
                    onChange={(e) => setUnits(p => ({ ...p, capacity: e.target.value }))}
                    className="bg-surface-darker border border-border-color rounded-lg px-2 text-xs text-white focus:ring-1 focus:ring-primary outline-none"
                  >
                    {WEIGHT_UNITS.map(u => <option key={u.value} value={u.value}>{u.value}</option>)}
                  </select>
                </div>
              </div>

              <div className="space-y-1.5 lg:col-span-1">
                <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                  {form.fuel_type === "CNG" ? "Cylinder Capacity (kg)" : "Fuel Tank Capacity"}
                </label>
                <div className="flex gap-2">
                  <TextInput
                    label="" name="fuel_tank_capacity" value={form.fuel_tank_capacity}
                    onChange={handleChange} type="number" placeholder="0.00"
                    className="flex-1"
                  />
                  {form.fuel_type !== "CNG" && (
                    <select 
                      value={units.fuel} 
                      onChange={(e) => setUnits(p => ({ ...p, fuel: e.target.value }))}
                      className="bg-surface-darker border border-border-color rounded-lg px-2 text-xs text-white focus:ring-1 focus:ring-primary outline-none"
                    >
                      <option value="litres">L</option>
                      <option value="gallons">gal</option>
                    </select>
                  )}
                </div>
              </div>

              <TextInput
                label="GPS Device ID" name="gps_device_id" value={form.gps_device_id}
                onChange={handleChange} placeholder="e.g. GPS-12345" icon="router"
              />
            </div>
          </div>

          {/* Section 4: Notes & Metadata */}
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            <div className="lg:col-span-2 space-y-6">
              {/* Telemetry & Compliance */}
              <div className="bg-surface-dark rounded-xl border border-border-color p-6 shadow-xl">
                 <div className="flex items-center gap-3 mb-6 pb-4 border-b border-border-color">
                  <div className="w-8 h-8 rounded bg-primary/10 flex items-center justify-center text-primary">
                    <span className="material-symbols-outlined text-[18px]">verified</span>
                  </div>
                  <h3 className="text-white font-semibold text-sm">Telemetry & Compliance</h3>
                </div>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <ToggleInput
                    label="Immobilizer" description="Remote engine lock capability"
                    checked={form.immobilizer_enabled}
                    onChange={(checked) => setForm(p => ({ ...p, immobilizer_enabled: checked }))}
                    icon="lock"
                  />
                  <ToggleInput
                    label="AIS-140 Compliant" description="Govt tracking standards"
                    checked={form.ais140_compliant}
                    onChange={(checked) => setForm(p => ({ ...p, ais140_compliant: checked }))}
                    icon="gavel"
                  />
                </div>
              </div>

              {/* Notes */}
              <div className="bg-surface-dark rounded-xl border border-border-color p-6 shadow-xl">
                <div className="flex items-center gap-3 mb-4">
                  <div className="w-8 h-8 rounded bg-primary/10 flex items-center justify-center text-primary">
                    <span className="material-symbols-outlined text-[18px]">description</span>
                  </div>
                  <h3 className="text-white font-semibold text-sm">Operational Notes</h3>
                </div>
                <textarea 
                  name="notes" value={form.notes} onChange={handleChange} rows={5} maxLength={500}
                  placeholder="Internal notes about vehicle history or specs..."
                  className="w-full bg-surface-darker border border-border-color rounded-lg px-4 py-3 text-sm text-white focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary placeholder-text-muted/30 transition-all resize-none"
                />
                <div className="flex justify-between items-center mt-2 px-1">
                  <span className="text-[10px] text-text-muted">Avoid sensitive data in notes</span>
                  <span className={`text-[10px] font-medium ${charCount > 450 ? "text-warning" : "text-text-muted"}`}>{charCount}/500</span>
                </div>
              </div>
            </div>

            {/* Side Info Panel */}
            <div className="lg:col-span-1 bg-surface-dark rounded-xl border border-border-color p-6 shadow-xl flex flex-col">
              <div className="flex items-center gap-3 mb-6">
                <div className="w-8 h-8 rounded bg-primary/10 flex items-center justify-center text-primary">
                  <span className="material-symbols-outlined text-[18px]">info</span>
                </div>
                <h3 className="text-white font-semibold text-sm">System Metadata</h3>
              </div>
              <div className="flex-1 space-y-4">
                <div className="flex justify-between items-center bg-surface-darker p-3 rounded-lg border border-border-color">
                  <span className="text-xs text-text-muted">Vehicle ID</span>
                  <span className="text-xs font-mono text-white">#{vehicle?.id?.slice(-8).toUpperCase()}</span>
                </div>
                <div className="flex justify-between items-center bg-surface-darker p-3 rounded-lg border border-border-color">
                  <span className="text-xs text-text-muted">Last Active</span>
                  <span className="text-xs text-white">2 hrs ago</span>
                </div>
                <div className="flex justify-between items-center bg-surface-darker p-3 rounded-lg border border-border-color">
                  <span className="text-xs text-text-muted">Created</span>
                  <span className="text-xs text-white">{vehicle?.created_at ? new Date(vehicle.created_at).toLocaleDateString() : "—"}</span>
                </div>
              </div>
              
              <div className="mt-8">
                <button 
                  type="button" onClick={handleDelete}
                  className={`w-full py-3 rounded-lg text-xs font-bold transition-all flex items-center justify-center gap-2 group ${
                    confirmDelete
                      ? "bg-danger hover:bg-red-600 text-white shadow-lg shadow-danger/20"
                      : "border border-danger/30 text-danger hover:bg-danger/10"
                  }`}
                >
                  <span className="material-symbols-outlined text-[18px] group-hover:animate-pulse">
                    {confirmDelete ? "delete_forever" : "delete"}
                  </span>
                  {confirmDelete ? "CONFIRM PERMANENT DELETE" : "DELETE ASSET"}
                </button>
                {confirmDelete && (
                  <p className="text-[10px] text-danger text-center mt-2 uppercase tracking-tighter">This action cannot be undone</p>
                )}
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


export default EditVehicle;
