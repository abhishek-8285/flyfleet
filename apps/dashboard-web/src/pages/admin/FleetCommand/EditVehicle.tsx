import React, { useEffect, useState } from "react";
import { useNavigate, useParams, Link } from "react-router-dom";
import { motion } from "framer-motion";
import { 
  Shield, 
  Activity, 
  Truck, 
  Cpu, 
  Globe, 
  Save,
  ArrowLeft,
  ChevronRight,
  Settings,
  ShieldCheck,
  History,
  Trash2,
  AlertTriangle,
  Zap,
  Info
} from "lucide-react";
import apiClient from "../../../services/apiClient";
import { 
  FUEL_TYPES, VEHICLE_TYPES, OWNERSHIP_TYPES, 
  STATUS_OPTIONS
} from "./constants";
import { toKm } from "./utils";
import { TextInput, SelectInput, ToggleInput } from "../../../components/common/FormFields";
import { DropdownOption } from "../../../types/common";

// ── Types ──
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
  gps_enabled: boolean;
  gps_device_type: string;
  color: string;
  chassis_number: string;
  engine_number: string;
  parent_vehicle_id: string;
  branch_id: string;
  cost_center_id: string;
  current_driver_id: string;
  status: string;
  notes: string;
  odometer: string;
  odometer_unit: string;
  speed_limit: string;
  // Compliance
  insurance: string;
  insurance_expiry: string;
  permit_number: string;
  permit_expiry: string;
  pollution_expiry: string;
  fitness_expiry: string;
  tax_expiry: string;
}

export const EditVehicle: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [form, setForm] = useState<FormState>({
    make: "", model: "", year: "", vin_number: "", fuel_type: "DIESEL",
    vehicle_type: "", ownership_type: "OWNED", capacity_tonnage: "",
    fuel_tank_capacity: "", gps_device_id: "", gps_enabled: false, gps_device_type: "AIS140_HARDWIRED", color: "", chassis_number: "",
    engine_number: "", parent_vehicle_id: "", branch_id: "", cost_center_id: "",
    status: "ACTIVE", notes: "", odometer: "0", odometer_unit: "km", speed_limit: "0", current_driver_id: "",
    insurance: "", insurance_expiry: "", permit_number: "", permit_expiry: "",
    pollution_expiry: "", fitness_expiry: "", tax_expiry: "",
  });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [submitError, setSubmitError] = useState("");
  const [vehicle, setVehicle] = useState<any>(null);
  
  const [gpsDeviceTypes, setGpsDeviceTypes] = useState<DropdownOption[]>([]);
  const [dropdowns, setDropdowns] = useState({
    vehicles: [] as DropdownOption[],
    drivers: [] as DropdownOption[],
    branches: [] as DropdownOption[],
    costCenters: [] as DropdownOption[],
  });

  useEffect(() => { 
    const fetchData = async () => {
        try {
          const [vDetails, vList, dResp, bResp, cResp, tResp] = await Promise.all([
            apiClient.get(`/admin/vehicles/${id}`),
            apiClient.get("/admin/vehicles"),
            apiClient.get("/admin/drivers?limit=500"),
            apiClient.get("/admin/branches"),
            apiClient.get("/admin/cost-centers"),
            apiClient.get("/ui/settings/gps-types")
          ]);
          
          const v = vDetails.data;
          setVehicle(v);
          setForm({
            make: v.make || "", model: v.model || "", year: v.year ? v.year.toString() : "",
            vin_number: v.vin_number || "", fuel_type: v.fuel_type || "DIESEL",
            vehicle_type: v.vehicle_type || "", ownership_type: v.ownership_type || "OWNED",
            capacity_tonnage: v.capacity_tonnage ? v.capacity_tonnage.toString() : "",
            fuel_tank_capacity: v.fuel_tank_capacity ? v.fuel_tank_capacity.toString() : "",
            gps_device_id: v.gps_device_id || "",
            gps_enabled: v.gps_enabled || false,
            gps_device_type: v.gps_device_type || "AIS140_HARDWIRED",
            color: v.color || "",
            chassis_number: v.chassis_number || "", engine_number: v.engine_number || "",
            parent_vehicle_id: v.parent_vehicle_id || "", branch_id: v.branch_id || "",
            cost_center_id: v.cost_center_id || "",
            current_driver_id: v.current_driver_id || "",
            status: v.status || "ACTIVE", notes: v.notes || "",
            odometer: v.odometer ? (v.odometer / 1000).toString() : "0",
            odometer_unit: "km",
            speed_limit: v.speed_limit ? v.speed_limit.toString() : "0",
            insurance: v.insurance || "",
            insurance_expiry: v.insurance_expiry ? v.insurance_expiry.split("T")[0] : "",
            permit_number: v.permit_number || "",
            permit_expiry: v.permit_expiry ? v.permit_expiry.split("T")[0] : "",
            pollution_expiry: v.pollution_expiry ? v.pollution_expiry.split("T")[0] : "",
            fitness_expiry: v.fitness_expiry ? v.fitness_expiry.split("T")[0] : "",
            tax_expiry: v.tax_expiry ? v.tax_expiry.split("T")[0] : "",
          });

          const ensureArray = (resp: any, key?: string) => {
            if (!resp) return [];
            if (Array.isArray(resp)) return resp;
            if (key && Array.isArray(resp[key])) return resp[key];
            if (Array.isArray(resp.data)) return resp.data;
            if (resp.data && key && Array.isArray(resp.data[key])) return resp.data[key];
            return [];
          };

          setDropdowns({
            vehicles: ensureArray(vList, "vehicles").map((v: any) => ({ value: v.id, label: v.registration_number })),
            drivers: ensureArray(dResp, "drivers").map((d: any) => ({ 
              value: d.id, 
              label: `${d.first_name || ""} ${d.last_name || ""}`.trim() || d.name || d.id 
            })),
            branches: ensureArray(bResp, "data").map((b: any) => ({ value: b.id, label: b.name })),
            costCenters: ensureArray(cResp, "data").map((c: any) => ({ value: c.id, label: c.name })),
          });
          setGpsDeviceTypes(ensureArray(tResp));
        } catch (err) {
          console.error("Failed to fetch vehicle data:", err);
        } finally {
          setLoading(false);
        }
    };
    fetchData();
  }, [id]);

  const updateField = (name: keyof FormState, value: any) => {
    setForm(prev => ({ ...prev, [name]: value }));
  };

  const handleSubmit = async () => {
    setSaving(true);
    setSubmitError("");
    try {
      const payload = {
        ...form,
        registration_number: vehicle.registration_number,
        year: parseInt(form.year) || 0,
        capacity_tonnage: parseFloat(form.capacity_tonnage) || 0,
        fuel_tank_capacity: parseFloat(form.fuel_tank_capacity) || 0,
        speed_limit: parseFloat(form.speed_limit) || 0,
        odometer: Math.round(toKm(parseFloat(form.odometer) || 0, form.odometer_unit) * 1000),
        branch_id: form.branch_id || null,
        cost_center_id: form.cost_center_id || null,
        current_driver_id: form.current_driver_id || null,
        parent_vehicle_id: form.parent_vehicle_id || null,
      };
      await apiClient.put(`/admin/vehicles/${id}`, payload);
      navigate(`/admin/vehicles/${id}`);
    } catch (err: any) {
      setSubmitError(err.response?.data?.error || "Failed to update asset node.");
    } finally {
      setSaving(false);
    }
  };

  if (loading || !vehicle) return (
    <div className="flex h-screen bg-background-dark items-center justify-center">
       <div className="w-12 h-12 border-4 border-primary/20 border-t-primary rounded-full animate-spin" />
    </div>
  );

  return (
    <div className="flex-1 flex flex-col bg-background-dark overflow-hidden font-sans relative">
      {/* --- Standardized Header --- */}
      <header className="h-16 flex items-center justify-between px-6 border-b border-border-color bg-surface-darker flex-shrink-0 z-30">
        <div className="flex items-center gap-4">
          <button onClick={() => navigate(`/admin/vehicles/${id}`)} className="p-2 hover:bg-white/5 rounded-full transition-colors text-text-muted">
            <ArrowLeft className="w-5 h-5" />
          </button>
          <div>
            <div className="flex items-center gap-2 text-[10px] text-text-muted uppercase tracking-[0.2em] font-black mb-0.5">
              <span>Fleet Management</span>
              <ChevronRight className="w-3 h-3 opacity-30" />
              <span className="text-white">Edit Vehicle</span>
            </div>
            <h2 className="text-white text-lg font-bold flex items-center gap-2 uppercase tracking-tight">
              {vehicle.registration_number}
              <span className="px-2 py-0.5 rounded text-[10px] font-black uppercase tracking-wider border bg-primary/10 text-primary border-primary/20">
                Editing
              </span>
            </h2>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <Link
            to={`/admin/vehicles/${id}`}
            className="px-4 py-2 bg-surface-dark border border-border-color text-text-muted hover:text-white text-[10px] font-black uppercase tracking-widest rounded-lg transition-all"
          >
            Cancel
          </Link>
          <button 
            onClick={handleSubmit} 
            disabled={saving}
            className="px-6 py-2 bg-primary hover:bg-primary-dark text-white text-[10px] font-black uppercase tracking-widest rounded-lg flex items-center gap-2 shadow-lg shadow-primary/20 transition-all disabled:opacity-50"
          >
            {saving ? <div className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin" /> : <Save className="w-3.5 h-3.5" />}
            {saving ? 'Processing...' : 'Save Changes'}
          </button>
        </div>
      </header>

      <div className="flex-1 overflow-y-auto custom-scrollbar bg-background-dark p-6">
        <div className="max-w-5xl mx-auto space-y-8 pb-32">
          
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            <div className="bg-surface-dark border border-border-color rounded-xl p-5 flex flex-col gap-1 relative overflow-hidden group">
              <span className="text-[10px] font-black text-gray-500 uppercase tracking-widest">Fuel Type</span>
              <span className="text-white font-black text-xl uppercase tracking-tight">{vehicle.fuel_type}</span>
              <Zap className="absolute -bottom-2 -right-2 w-12 h-12 text-white opacity-[0.02]" />
            </div>
            <div className="bg-surface-dark border border-border-color rounded-xl p-5 flex flex-col gap-1 relative overflow-hidden group">
              <span className="text-[10px] font-black text-gray-500 uppercase tracking-widest">Load Capacity</span>
              <span className="text-white font-black text-xl uppercase tracking-tight">{vehicle.capacity_tonnage || 0} <span className="text-xs text-gray-600 font-bold">TONS</span></span>
              <Truck className="absolute -bottom-2 -right-2 w-12 h-12 text-white opacity-[0.02]" />
            </div>
            <div className="bg-surface-dark border border-border-color rounded-xl p-5 flex flex-col gap-1 relative overflow-hidden group">
              <span className="text-[10px] font-black text-gray-500 uppercase tracking-widest">Speed Limit</span>
              <span className="text-white font-black text-xl uppercase tracking-tight">{vehicle.speed_limit || 0} <span className="text-xs text-gray-600 font-bold">KMPH</span></span>
              <Activity className="absolute -bottom-2 -right-2 w-12 h-12 text-white opacity-[0.02]" />
            </div>
            <div className="bg-surface-dark border border-border-color rounded-xl p-5 flex flex-col gap-1 relative overflow-hidden group">
              <span className="text-[10px] font-black text-gray-500 uppercase tracking-widest">Vehicle Status</span>
              <span className={`font-black text-xl uppercase tracking-tight ${vehicle.status === 'ACTIVE' ? 'text-emerald-500' : 'text-amber-500'}`}>{vehicle.status}</span>
              <Globe className="absolute -bottom-2 -right-2 w-12 h-12 text-white opacity-[0.02]" />
            </div>
          </div>

          <section className="bg-surface-dark border border-border-color rounded-xl overflow-hidden shadow-xl">
             <div className="px-6 py-4 border-b border-border-color bg-white/[0.02] flex items-center gap-3">
                <Truck className="w-5 h-5 text-primary" />
                <h3 className="text-xs font-black uppercase tracking-[0.2em] text-white">Vehicle Details</h3>
             </div>
              <div className="p-8 grid grid-cols-1 md:grid-cols-2 gap-x-8 gap-y-6">
                <TextInput label="Registration Number (Read Only)" value={vehicle.registration_number} disabled className="opacity-50" />
                <SelectInput label="Vehicle Type" value={form.vehicle_type} options={VEHICLE_TYPES} onChange={(e) => updateField("vehicle_type", e.target.value)} />
                <SelectInput label="Ownership Type" value={form.ownership_type} options={OWNERSHIP_TYPES} onChange={(e) => updateField("ownership_type", e.target.value)} />
                <SelectInput label="Vehicle Status" value={form.status} options={STATUS_OPTIONS} onChange={(e) => updateField("status", e.target.value)} />
              </div>
          </section>

          {/* 2. Technical Specs */}
          <section className="bg-surface-dark border border-border-color rounded-xl overflow-hidden shadow-xl">
            <div className="px-6 py-4 border-b border-border-color bg-white/[0.02] flex items-center gap-3">
              <Cpu className="w-5 h-5 text-amber-500" />
              <h3 className="text-xs font-black uppercase tracking-[0.2em] text-white">Technical Specs</h3>
            </div>
             <div className="p-8">
                <div className="grid grid-cols-1 md:grid-cols-3 gap-8 mb-10">
                   <TextInput label="Manufacturer" value={form.make} onChange={(e) => updateField("make", e.target.value)} />
                   <TextInput label="Variant/Model" value={form.model} onChange={(e) => updateField("model", e.target.value)} />
                   <TextInput label="Deployment Year" value={form.year} onChange={(e) => updateField("year", e.target.value)} />
                </div>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-8 mb-10">
                   <TextInput label="VIN Number" value={form.vin_number} onChange={(e) => updateField("vin_number", e.target.value)} />
                   <TextInput label="Chassis Number" value={form.chassis_number} onChange={(e) => updateField("chassis_number", e.target.value)} />
                </div>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-8 mb-10">
                   <TextInput label="Engine Number" value={form.engine_number} onChange={(e) => updateField("engine_number", e.target.value)} />
                   <TextInput label="Color" value={form.color} onChange={(e) => updateField("color", e.target.value)} />
                </div>
                <div className="grid grid-cols-2 md:grid-cols-4 gap-6 pt-10 border-t border-white/5">
                   <TextInput label="Payload (Tons)" type="number" value={form.capacity_tonnage} onChange={(e) => updateField("capacity_tonnage", e.target.value)} />
                   <TextInput label="Fuel Res (L)" type="number" value={form.fuel_tank_capacity} onChange={(e) => updateField("fuel_tank_capacity", e.target.value)} />
                   <TextInput label="Odometer (KM)" type="number" value={form.odometer} onChange={(e) => updateField("odometer", e.target.value)} />
                   <TextInput label="Speed Limit" type="number" value={form.speed_limit} onChange={(e) => updateField("speed_limit", e.target.value)} />
                </div>
             </div>
          </section>

          <section className="bg-surface-dark border border-border-color rounded-xl overflow-hidden shadow-xl">
             <div className="px-6 py-4 border-b border-border-color bg-white/[0.02] flex items-center gap-3">
                <Activity className="w-5 h-5 text-emerald-500" />
                <h3 className="text-xs font-black uppercase tracking-[0.2em] text-white">Vehicle Tracking</h3>
             </div>
             <div className="p-8 space-y-8">
                <div className="p-8 rounded-2xl bg-primary/5 border border-primary/20 flex items-center justify-between">
                   <div className="space-y-1">
                      <h4 className="text-xs font-black uppercase tracking-widest flex items-center gap-2">
                         <Globe className="w-4 h-4 text-primary" /> Tracking Status
                      </h4>
                      <p className="text-[10px] text-gray-600 uppercase font-bold tracking-widest leading-none">Enable real-time positional broadcast</p>
                   </div>
                   <ToggleInput label="" checked={form.gps_enabled} onChange={(v) => updateField("gps_enabled", v)} />
                </div>

                {form.gps_enabled && (
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-8 animate-in fade-in slide-in-from-top-2 duration-300">
                    <TextInput label="Device IMEI" value={form.gps_device_id} onChange={(e) => updateField("gps_device_id", e.target.value)} />
                    <SelectInput label="Link Protocol" value={form.gps_device_type} options={gpsDeviceTypes} onChange={(e) => updateField("gps_device_type", e.target.value)} />
                  </div>
                )}

                <div className="grid grid-cols-1 md:grid-cols-2 gap-8 pt-10 border-t border-white/5">
                   <SelectInput label="Operational Branch" value={form.branch_id} options={dropdowns.branches} onChange={(e) => updateField("branch_id", e.target.value)} placeholder="Ungrouped" />
                   <SelectInput label="Cost Allocation" value={form.cost_center_id} options={dropdowns.costCenters} onChange={(e) => updateField("cost_center_id", e.target.value)} placeholder="Default" />
                   <SelectInput label="Primary Driver" value={form.current_driver_id} options={dropdowns.drivers} onChange={(e) => updateField("current_driver_id", e.target.value)} placeholder="Unassigned" />
                   <SelectInput label="Parent Asset Node" value={form.parent_vehicle_id} options={dropdowns.vehicles} onChange={(e) => updateField("parent_vehicle_id", e.target.value)} placeholder="Standalone" />
                </div>
             </div>
          </section>

          <section className="bg-surface-dark border border-border-color rounded-xl overflow-hidden shadow-xl">
             <div className="px-6 py-4 border-b border-border-color bg-white/[0.02] flex items-center gap-3">
                <ShieldCheck className="w-5 h-5 text-indigo-400" />
                <h3 className="text-xs font-black uppercase tracking-[0.2em] text-white">Documents & Compliance</h3>
             </div>
             <div className="p-8 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                <TextInput label="Insurance Node #" value={form.insurance} onChange={(e) => updateField("insurance", e.target.value)} />
                <TextInput label="Insurance Expiry" type="date" value={form.insurance_expiry} onChange={(e) => updateField("insurance_expiry", e.target.value)} />
                <TextInput label="Commercial Permit" value={form.permit_number} onChange={(e) => updateField("permit_number", e.target.value)} />
                <TextInput label="Permit Expiry" type="date" value={form.permit_expiry} onChange={(e) => updateField("permit_expiry", e.target.value)} />
                <TextInput label="Pollution Expiry" type="date" value={form.pollution_expiry} onChange={(e) => updateField("pollution_expiry", e.target.value)} />
                <TextInput label="Fitness Expiry" type="date" value={form.fitness_expiry} onChange={(e) => updateField("fitness_expiry", e.target.value)} />
                <TextInput label="Tax Validity" type="date" value={form.tax_expiry} onChange={(e) => updateField("tax_expiry", e.target.value)} />
             </div>
          </section>

          {/* Danger Zone */}
          <section className="p-8 rounded-xl border border-danger/20 bg-danger/5 flex items-center justify-between">
             <div className="flex items-center gap-6">
                <div className="w-12 h-12 rounded-2xl bg-danger/10 flex items-center justify-center text-danger">
                   <Trash2 className="w-6 h-6" />
                </div>
                <div>
                   <h4 className="text-sm font-black uppercase text-white tracking-widest mb-1">Delete Vehicle</h4>
                   <p className="text-[10px] text-gray-500 uppercase font-black leading-none">Irreversible removal of this vehicle record</p>
                </div>
             </div>
             <button className="px-6 py-3 bg-danger text-white rounded-xl text-[10px] font-black uppercase tracking-[0.2em] hover:bg-danger-dark transition-all">
                CONFIRM DELETE
             </button>
          </section>

        </div>
      </div>
    </div>
  );
};

export default EditVehicle;
