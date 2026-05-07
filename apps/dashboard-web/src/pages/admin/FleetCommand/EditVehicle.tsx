import React, { useEffect, useState, useCallback, useMemo } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { motion, AnimatePresence } from "framer-motion";
import { 
  Shield, 
  Activity, 
  Truck, 
  Cpu, 
  Globe, 
  Save,
  ArrowLeft,
  Info,
  History,
  Trash2,
  AlertTriangle,
  ClipboardCheck,
  Settings,
  Wrench,
  ShieldCheck,
  Link as LinkIcon,
  MapPin
} from "lucide-react";
import apiClient from "../../../services/apiClient";
import { 
  FUEL_TYPES, VEHICLE_TYPES, OWNERSHIP_TYPES, 
  ODOMETER_UNITS, WEIGHT_UNITS, getFuelUnits,
  STATUS_OPTIONS
} from "./constants";
import { toKm, toTonnes, toLitres } from "./utils";
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
  immobilizer_enabled: boolean;
  ais140_compliant: boolean;
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
  // Maintenance
  last_service_date: string;
  last_service_odometer: string;
  next_service_date: string;
}

type FieldName = keyof FormState;

const STEPS = [
  { id: 1, label: "Identity", icon: Truck, description: "Vehicle classification & status" },
  { id: 2, label: "Hardware", icon: Cpu, description: "Engineering & technical specs" },
  { id: 3, label: "Network", icon: Activity, description: "Telemetry & organizational hooks" },
  { id: 4, label: "Registry", icon: Shield, description: "Legal & compliance validity" },
];

export const EditVehicle: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [form, setForm] = useState<FormState>({
    make: "", model: "", year: "", vin_number: "", fuel_type: "DIESEL",
    vehicle_type: "", ownership_type: "OWNED", capacity_tonnage: "",
    fuel_tank_capacity: "", gps_device_id: "", gps_enabled: false, gps_device_type: "AIS140_HARDWIRED", color: "", chassis_number: "",
    engine_number: "", parent_vehicle_id: "", branch_id: "", cost_center_id: "",
    immobilizer_enabled: false, ais140_compliant: false, status: "ACTIVE", notes: "",
    odometer: "0", odometer_unit: "km", speed_limit: "0", current_driver_id: "",
    insurance: "", insurance_expiry: "", permit_number: "", permit_expiry: "",
    pollution_expiry: "", fitness_expiry: "", tax_expiry: "",
    last_service_date: "", last_service_odometer: "", next_service_date: "",
  });
  const [currentStep, setCurrentStep] = useState(1);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [submitError, setSubmitError] = useState("");
  const [confirmDelete, setConfirmDelete] = useState(false);
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
            apiClient.get("/admin/drivers"),
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
            immobilizer_enabled: v.immobilizer_enabled || false,
            ais140_compliant: v.ais140_compliant || false,
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
            last_service_date: v.last_service_date ? v.last_service_date.split("T")[0] : "",
            last_service_odometer: v.last_service_odometer ? v.last_service_odometer.toString() : "",
            next_service_date: v.next_service_date ? v.next_service_date.split("T")[0] : "",
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

  const updateField = (name: FieldName, value: any) => {
    setForm(prev => ({ ...prev, [name]: value }));
    if (errors[name]) setErrors(prev => {
        const n = { ...prev };
        delete n[name];
        return n;
    });
  };

  const handleSubmit = async () => {
    setSaving(true);
    setSubmitError("");
    try {
      const payload = {
        ...form,
        registration_number: vehicle.registration_number, // Keep original
        year: parseInt(form.year) || 0,
        capacity_tonnage: parseFloat(form.capacity_tonnage) || 0,
        fuel_tank_capacity: parseFloat(form.fuel_tank_capacity) || 0,
        speed_limit: parseFloat(form.speed_limit) || 0,
        odometer: Math.round(toKm(parseFloat(form.odometer) || 0, form.odometer_unit) * 1000),
        // Filter out empty UUID strings to prevent backend decoding errors
        branch_id: form.branch_id || null,
        cost_center_id: form.cost_center_id || null,
        current_driver_id: form.current_driver_id || null,
        parent_vehicle_id: form.parent_vehicle_id || null,
      };
      console.log("[EditVehicle] Sending update payload:", payload);
      await apiClient.put(`/admin/vehicles/${id}`, payload);
      navigate(`/admin/vehicles/${id}`);
    } catch (err: any) {
      setSubmitError(err.response?.data?.error || "Failed to update asset node.");
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
      setSubmitError(err.response?.data?.error || "Purge failed.");
    }
  };

  if (loading) return (
    <div className="flex h-screen bg-[#0a0a0b] items-center justify-center">
       <div className="w-12 h-12 border-4 border-primary/20 border-t-primary rounded-full animate-spin" />
    </div>
  );

  return (
    <div className="flex h-screen bg-[#0a0a0b] overflow-hidden">
      {/* ── Left Navigation Rail ── */}
      <div className="w-24 border-r border-[#1a1a1c] bg-[#0d0d0f] flex flex-col items-center py-10 gap-10">
        <div className="w-12 h-12 bg-primary/20 rounded-2xl flex items-center justify-center border border-primary/20">
          <Truck className="text-primary w-6 h-6" />
        </div>
        
        <div className="flex flex-col gap-6">
          {STEPS.map((s) => {
            const Icon = s.icon;
            const isActive = currentStep === s.id;
            return (
              <button
                key={s.id}
                onClick={() => setCurrentStep(s.id)}
                className={`group relative w-12 h-12 rounded-xl flex items-center justify-center transition-all ${
                  isActive ? "bg-primary shadow-lg shadow-primary/20" : "hover:bg-white/5"
                }`}
              >
                <Icon className={`w-5 h-5 ${isActive ? "text-white" : "text-text-muted group-hover:text-white"}`} />
                {isActive && (
                  <motion.div 
                    layoutId="active-pill"
                    className="absolute -right-[1px] w-1 h-6 bg-primary rounded-l-full"
                  />
                )}
                <div className="absolute left-16 px-3 py-1 bg-surface-darker border border-border-color rounded-md text-[10px] whitespace-nowrap opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-50">
                  {s.label}
                </div>
              </button>
            );
          })}
        </div>

        <div className="mt-auto flex flex-col gap-6 mb-4">
           <button 
             onClick={() => navigate(`/admin/vehicles/${id}/history`)}
             className="w-12 h-12 rounded-xl flex items-center justify-center hover:bg-white/5 transition-all text-text-muted hover:text-white"
           >
              <History className="w-5 h-5" />
           </button>
           <button 
             onClick={handleDelete}
             className={`w-12 h-12 rounded-xl flex items-center justify-center transition-all ${confirmDelete ? "bg-danger text-white scale-110" : "hover:bg-danger/10 text-text-muted hover:text-danger"}`}
           >
              <Trash2 className="w-5 h-5" />
           </button>
        </div>
      </div>

      {/* ── Main Content Area ── */}
      <div className="flex-1 flex overflow-hidden">
        {/* Form Pane */}
        <div className="flex-1 overflow-y-auto custom-scrollbar bg-gradient-to-br from-[#0a0a0b] to-[#0d0d0f]">
          <div className="max-w-3xl mx-auto px-10 py-16">
            <header className="mb-12">
              <div className="flex items-center gap-3 text-[10px] font-bold text-primary uppercase tracking-[0.2em] mb-4">
                <Shield className="w-3 h-3" /> Reconfiguring Asset Cluster
              </div>
              <h1 className="text-4xl font-black text-white tracking-tight mb-4">
                Update <span className="text-primary">Registry</span> Node
              </h1>
              <p className="text-text-muted text-sm max-w-xl leading-relaxed">
                Modifying technical specifications for node <span className="text-white font-mono">{vehicle.registration_number}</span>. Change history is logged for compliance.
              </p>
            </header>

            <AnimatePresence mode="wait">
              <motion.div
                key={currentStep}
                initial={{ opacity: 0, x: 20 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: -20 }}
              >
                {currentStep === 1 && (
                  <div className="space-y-12">
                    <section className="space-y-8">
                       <h3 className="text-xs font-bold text-white uppercase tracking-widest flex items-center gap-3">
                         <div className="w-1.5 h-1.5 rounded-full bg-primary" /> core identity
                       </h3>
                       <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                         <TextInput label="Registration (Read Only)" name="registration_number" value={vehicle.registration_number} onChange={() => {}} disabled className="opacity-50" />
                         <SelectInput label="Registry Classification" name="vehicle_type" value={form.vehicle_type} options={VEHICLE_TYPES} onChange={(e) => updateField("vehicle_type", e.target.value)} />
                       </div>
                       <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                         <SelectInput label="Financial Model" name="ownership_type" value={form.ownership_type} options={OWNERSHIP_TYPES} onChange={(e) => updateField("ownership_type", e.target.value)} />
                         <SelectInput label="Operational Status" name="status" value={form.status} options={STATUS_OPTIONS} onChange={(e) => updateField("status", e.target.value)} />
                       </div>
                    </section>
                  </div>
                )}

                {currentStep === 2 && (
                  <div className="space-y-12">
                    <section className="space-y-8">
                       <h3 className="text-xs font-bold text-white uppercase tracking-widest flex items-center gap-3">
                         <div className="w-1.5 h-1.5 rounded-full bg-primary" /> Engineering Metadata
                       </h3>
                       <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                         <TextInput label="Manufacturer" name="make" value={form.make} onChange={(e) => updateField("make", e.target.value)} />
                         <TextInput label="Variant/Model" name="model" value={form.model} onChange={(e) => updateField("model", e.target.value)} />
                         <TextInput label="Deployment Year" name="year" value={form.year} onChange={(e) => updateField("year", e.target.value)} />
                       </div>
                       <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                         <TextInput label="VIN Signal" name="vin_number" value={form.vin_number} onChange={(e) => updateField("vin_number", e.target.value)} />
                         <TextInput label="Chassis Vector" name="chassis_number" value={form.chassis_number} onChange={(e) => updateField("chassis_number", e.target.value)} />
                       </div>
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                         <TextInput label="Engine Sequence" name="engine_number" value={form.engine_number} onChange={(e) => updateField("engine_number", e.target.value)} />
                         <TextInput label="Registry Color" name="color" value={form.color} onChange={(e) => updateField("color", e.target.value)} />
                       </div>
                    </section>
                    
                    <section className="space-y-8">
                       <h3 className="text-xs font-bold text-white uppercase tracking-widest flex items-center gap-3">
                         <div className="w-1.5 h-1.5 rounded-full bg-primary" /> Operational Metrics
                       </h3>
                       <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                         <TextInput label="Net Capacity (Tonnes)" name="capacity_tonnage" type="number" value={form.capacity_tonnage} onChange={(e) => updateField("capacity_tonnage", e.target.value)} />
                         <TextInput label="Fuel Reservoir (Litres)" name="fuel_tank_capacity" type="number" value={form.fuel_tank_capacity} onChange={(e) => updateField("fuel_tank_capacity", e.target.value)} />
                         <TextInput label="Update Odometer" name="odometer" type="number" value={form.odometer} onChange={(e) => updateField("odometer", e.target.value)} />
                         <TextInput label="Neural Speed Limit" name="speed_limit" type="number" value={form.speed_limit} onChange={(e) => updateField("speed_limit", e.target.value)} />
                       </div>
                    </section>
                  </div>
                )}

                {currentStep === 3 && (
                  <div className="space-y-12">
                    <section className="space-y-8">
                       <h3 className="text-xs font-bold text-white uppercase tracking-widest flex items-center gap-3">
                         <div className="w-1.5 h-1.5 rounded-full bg-primary" /> Telemetry Interface
                       </h3>
                       <div className="p-6 rounded-2xl bg-primary/5 border border-primary/20 space-y-6">
                          <div className="flex items-start justify-between">
                            <div className="space-y-1">
                               <h4 className="text-sm font-bold flex items-center gap-2">
                                 <Globe className="w-4 h-4 text-primary" /> Uplink Connectivity
                               </h4>
                               <p className="text-xs text-text-muted">Broadcast real-time positional data to the control tower.</p>
                            </div>
                            <ToggleInput label="" checked={form.gps_enabled} onChange={(v) => updateField("gps_enabled", v)} />
                          </div>
                          
                          {form.gps_enabled && (
                            <motion.div 
                              initial={{ opacity: 0, height: 0 }} 
                              animate={{ opacity: 1, height: "auto" }}
                              className="grid grid-cols-1 md:grid-cols-2 gap-6 pt-4 border-t border-primary/10"
                            >
                              <TextInput label="Device Serial (IMEI)" name="gps_device_id" value={form.gps_device_id} onChange={(e) => updateField("gps_device_id", e.target.value)} />
                              <SelectInput label="Link Protocol" name="gps_device_type" value={form.gps_device_type} options={gpsDeviceTypes} onChange={(e) => updateField("gps_device_type", e.target.value)} />
                            </motion.div>
                          )}
                       </div>
                    </section>

                    <section className="space-y-8">
                       <h3 className="text-xs font-bold text-white uppercase tracking-widest flex items-center gap-3">
                         <div className="w-1.5 h-1.5 rounded-full bg-primary" /> Org Hierarchy
                       </h3>
                       <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                         <SelectInput label="Operational Branch" name="branch_id" value={form.branch_id} options={dropdowns.branches} onChange={(e) => updateField("branch_id", e.target.value)} placeholder="Ungrouped" />
                         <SelectInput label="Cost Allocation" name="cost_center_id" value={form.cost_center_id} options={dropdowns.costCenters} onChange={(e) => updateField("cost_center_id", e.target.value)} placeholder="Default" />
                         <SelectInput label="Guardian Driver" name="current_driver_id" value={form.current_driver_id} options={dropdowns.drivers} onChange={(e) => updateField("current_driver_id", e.target.value)} placeholder="Unassigned" />
                         <SelectInput label="Parent Unified Node" name="parent_vehicle_id" value={form.parent_vehicle_id} options={dropdowns.vehicles} onChange={(e) => updateField("parent_vehicle_id", e.target.value)} placeholder="Top Level" />
                       </div>
                    </section>
                  </div>
                )}

                {currentStep === 4 && (
                  <div className="space-y-12">
                    <section className="space-y-8">
                       <h3 className="text-xs font-bold text-white uppercase tracking-widest flex items-center gap-3">
                         <div className="w-1.5 h-1.5 rounded-full bg-primary" /> Compliance & Legal
                       </h3>
                       <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                         <TextInput label="Insurance Policy #" name="insurance" value={form.insurance} onChange={(e) => updateField("insurance", e.target.value)} />
                         <TextInput label="Insurance Expiry" name="insurance_expiry" type="date" value={form.insurance_expiry} onChange={(e) => updateField("insurance_expiry", e.target.value)} />
                         <TextInput label="Permit Number" name="permit_number" value={form.permit_number} onChange={(e) => updateField("permit_number", e.target.value)} />
                         <TextInput label="Permit Expiry" name="permit_expiry" type="date" value={form.permit_expiry} onChange={(e) => updateField("permit_expiry", e.target.value)} />
                         <TextInput label="Pollution (PUC) Expiry" name="pollution_expiry" type="date" value={form.pollution_expiry} onChange={(e) => updateField("pollution_expiry", e.target.value)} />
                         <TextInput label="Fitness Expiry" name="fitness_expiry" type="date" value={form.fitness_expiry} onChange={(e) => updateField("fitness_expiry", e.target.value)} />
                         <TextInput label="Tax Validity Expiry" name="tax_expiry" type="date" value={form.tax_expiry} onChange={(e) => updateField("tax_expiry", e.target.value)} />
                       </div>
                    </section>

                    <section className="space-y-8">
                       <h3 className="text-xs font-bold text-white uppercase tracking-widest flex items-center gap-3">
                         <div className="w-1.5 h-1.5 rounded-full bg-primary" /> Maintenance History
                       </h3>
                       <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
                         <TextInput label="Last Service Date" name="last_service_date" type="date" value={form.last_service_date} onChange={(e) => updateField("last_service_date", e.target.value)} />
                         <TextInput label="Last Service Odometer" name="last_service_odometer" type="number" value={form.last_service_odometer} onChange={(e) => updateField("last_service_odometer", e.target.value)} />
                         <TextInput label="Next Service Due" name="next_service_date" type="date" value={form.next_service_date} onChange={(e) => updateField("next_service_date", e.target.value)} />
                       </div>
                    </section>
                  </div>
                )}
              </motion.div>
            </AnimatePresence>

            <div className="mt-20 pt-10 border-t border-[#1a1a1c] flex items-center justify-between">
              <button
                onClick={() => navigate(`/admin/vehicles/${id}`)}
                className="flex items-center gap-2 text-xs font-bold text-text-muted hover:text-white transition-colors"
                disabled={saving}
              >
                <ArrowLeft className="w-4 h-4" /> Discard Changes
              </button>

              <button
                onClick={handleSubmit}
                disabled={saving}
                className="px-10 py-4 bg-primary text-white rounded-2xl text-xs font-black hover:bg-primary/90 transition-all flex items-center gap-4 active:scale-95 shadow-2xl shadow-primary/20 disabled:opacity-50"
              >
                {saving ? <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" /> : <Save className="w-4 h-4" />}
                {saving ? "Committing Node..." : "Save Configuration"}
              </button>
            </div>
            
            {submitError && (
              <div className="mt-6 p-4 rounded-xl bg-danger/10 border border-danger/20 flex items-center gap-3 text-xs text-danger">
                 <AlertTriangle className="w-4 h-4" /> {submitError}
              </div>
            )}
          </div>
        </div>

        {/* ── Asset Preview Pane ── */}
        <div className="w-[480px] border-l border-[#1a1a1c] bg-[#0d0d0f] flex flex-col items-center justify-center p-12 relative">
          <div className="absolute top-12 left-10 flex items-center gap-3">
             <div className="w-2 h-2 rounded-full bg-success animate-pulse" />
             <span className="text-[10px] font-bold text-success uppercase tracking-widest text-white/40">Real-time Visualization Engine</span>
          </div>

          <div className="w-full space-y-12">
            {/* Holographic Asset Card */}
            <div className="relative group">
              <div className="absolute -inset-1 bg-gradient-to-r from-primary/20 to-purple-500/20 rounded-[2.5rem] blur-2xl opacity-50 group-hover:opacity-100 transition duration-1000" />
              <div className="relative bg-[#141416] border border-white/5 rounded-[2.5rem] p-10 overflow-hidden shadow-2xl">
                {/* Number Plate Style Header */}
                <div className="flex flex-col items-center mb-10">
                   <div className="bg-white px-6 py-2 rounded-md shadow-white/5 border border-black/10 flex items-center gap-3 mb-6 transform -rotate-1">
                      <div className="w-2 h-2 rounded-full bg-blue-600" />
                      <span className="text-black font-mono font-black text-xl tracking-tight">
                        {vehicle.registration_number}
                      </span>
                   </div>
                   <div className="text-[10px] font-bold text-text-muted uppercase tracking-[0.3em]">Fleet Registry Node</div>
                </div>

                {/* Chassis Status Info */}
                <div className="space-y-6">
                   <div className="flex items-center justify-between p-4 rounded-2xl bg-white/5 border border-white/5">
                      <div className="flex items-center gap-3">
                         <div className="w-8 h-8 rounded-lg bg-primary/10 flex items-center justify-center">
                            <Truck className="w-4 h-4 text-primary" />
                         </div>
                         <div className="flex flex-col">
                            <span className="text-[10px] text-text-muted font-bold uppercase">Classification</span>
                            <span className="text-sm font-bold text-white">{form.vehicle_type || "Awaiting Class..."}</span>
                         </div>
                      </div>
                      <div className="text-right">
                         <span className="text-[10px] text-text-muted font-bold uppercase">Fuel</span>
                         <div className="text-sm font-bold text-white">{form.fuel_type}</div>
                      </div>
                   </div>

                   <div className="grid grid-cols-2 gap-4">
                      <div className="p-4 rounded-2xl bg-white/5 border border-white/5">
                         <span className="text-[10px] text-text-muted font-bold uppercase block mb-1">Payload Max</span>
                         <span className="text-lg font-black text-white">{form.capacity_tonnage || "0"}<span className="text-xs font-medium ml-1 text-text-muted">TONS</span></span>
                      </div>
                      <div className="p-4 rounded-2xl bg-white/5 border border-white/5">
                         <span className="text-[10px] text-text-muted font-bold uppercase block mb-1">Speed Thresh</span>
                         <span className="text-lg font-black text-white">{form.speed_limit || "0"}<span className="text-xs font-medium ml-1 text-text-muted">KMPH</span></span>
                      </div>
                   </div>

                   <div className="p-4 rounded-2xl bg-primary/5 border border-primary/20 flex items-center justify-between">
                      <div className="flex items-center gap-3">
                         <Activity className={`w-4 h-4 ${form.gps_enabled ? "text-success" : "text-text-muted"}`} />
                         <span className="text-[10px] font-bold text-white uppercase tracking-wider">Telemetry Uplink</span>
                      </div>
                      <span className={`text-[10px] font-black uppercase tracking-tighter ${form.gps_enabled ? "text-success" : "text-text-muted"}`}>
                        {form.gps_enabled ? "Streaming Online" : "Link Terminated"}
                      </span>
                   </div>
                </div>
              </div>
            </div>

            {/* Quick Context Tip */}
            <div className="flex gap-4 p-6 rounded-3xl bg-white/5 border border-white/5">
               <div className="w-10 h-10 rounded-full bg-white/10 flex items-center justify-center shrink-0">
                  <Info className="w-5 h-5 text-white/60" />
               </div>
               <div className="space-y-1">
                  <h4 className="text-[10px] font-bold text-white uppercase tracking-widest">Registry Advisor</h4>
                  <p className="text-xs text-text-muted leading-relaxed">Updating the neural speed limit will trigger a recalibration across all active routes for this node.</p>
               </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default EditVehicle;
