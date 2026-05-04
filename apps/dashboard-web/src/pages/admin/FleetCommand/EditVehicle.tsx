import React, { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { 
  ArrowLeft, 
  Save, 
  Truck, 
  Settings, 
  MapPin, 
  AlertTriangle,
  ShieldCheck,
  Fuel,
  Link as LinkIcon,
  ShieldAlert,
  ClipboardCheck,
  Wrench,
  CalendarDays,
  Clock,
  FileText
} from "lucide-react";
import apiClient from "../../../services/apiClient";
import { TextInput, SelectInput, ToggleInput } from "../../../components/common/FormFields";
import { DocumentVault } from "../../../components/common/DocumentVault";

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

const FUEL_TYPES = [
  { value: "DIESEL", label: "Diesel" },
  { value: "PETROL", label: "Petrol" },
  { value: "CNG", label: "CNG" },
  { value: "ELECTRIC", label: "Electric" },
  { value: "HYBRID", label: "Hybrid" },
  { value: "LPG", label: "LPG" },
];

const OWNERSHIP_TYPES = [
  { value: "OWNED", label: "Owned" },
  { value: "LEASED", label: "Leased" },
  { value: "FINANCED", label: "Financed" },
  { value: "VENDOR", label: "Vendor" },
];

const STATUS_OPTIONS = [
  { value: "ACTIVE", label: "Active" },
  { value: "MAINTENANCE", label: "Maintenance" },
  { value: "SUSPENDED", label: "Suspended" },
  { value: "DECOMMISSIONED", label: "Decommissioned" },
];

export default function EditVehicle() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [vehicle, setVehicle] = useState<any>(null);
  const [error, setError] = useState<string | null>(null);
  const [submitError, setSubmitError] = useState("");
  
  const [dropdowns, setDropdowns] = useState({
    drivers: [] as any[],
    branches: [] as any[],
    costCenters: [] as any[],
  });

  const [form, setForm] = useState({
    vehicle_type: "",
    fuel_type: "",
    ownership_type: "",
    status: "",
    make: "",
    model: "",
    year: "",
    vin_number: "",
    chassis_number: "",
    engine_number: "",
    color: "",
    capacity_tonnage: "",
    fuel_tank_capacity: "",
    odometer: "",
    speed_limit: "",
    insurance: "",
    insurance_expiry: "",
    permit_number: "",
    permit_expiry: "",
    pollution_expiry: "",
    fitness_expiry: "",
    tax_expiry: "",
    last_service_date: "",
    next_service_date: "",
    last_service_odometer: "",
    gps_enabled: false,
    gps_device_id: "",
    gps_device_type: "",
    immobilizer_enabled: false,
    ais140_compliant: false,
    branch_id: "",
    cost_center_id: "",
    current_driver_id: "",
  });

  useEffect(() => {
    const fetchData = async () => {
      setLoading(true);
      try {
        const [vRes, dRes, bRes, cRes] = await Promise.allSettled([
          apiClient.get(`/admin/vehicles/${id}`),
          apiClient.get("/admin/drivers"),
          apiClient.get("/admin/branches"),
          apiClient.get("/admin/cost-centers"),
        ]);

        if (vRes.status === "fulfilled" && vRes.value?.data) {
          const data = vRes.value.data;
          setVehicle(data);
          
          const formatDate = (dateStr: string) => dateStr ? dateStr.split("T")[0] : "";

          setForm({
            vehicle_type: data.vehicle_type || "",
            fuel_type: data.fuel_type || "",
            ownership_type: data.ownership_type || "",
            status: data.status || "",
            make: data.make || "",
            model: data.model || "",
            year: data.year?.toString() || "",
            vin_number: data.vin_number || "",
            chassis_number: data.chassis_number || "",
            engine_number: data.engine_number || "",
            color: data.color || "",
            capacity_tonnage: data.capacity_tonnage?.toString() || "",
            fuel_tank_capacity: data.fuel_tank_capacity?.toString() || "",
            odometer: data.odometer ? (data.odometer / 1000).toString() : "0",
            speed_limit: data.speed_limit?.toString() || "",
            insurance: data.insurance || "",
            insurance_expiry: formatDate(data.insurance_expiry),
            permit_number: data.permit_number || "",
            permit_expiry: formatDate(data.permit_expiry),
            pollution_expiry: formatDate(data.pollution_expiry),
            fitness_expiry: formatDate(data.fitness_expiry),
            tax_expiry: formatDate(data.tax_expiry),
            last_service_date: formatDate(data.last_service_date),
            next_service_date: formatDate(data.next_service_date),
            last_service_odometer: data.last_service_odometer?.toString() || "0",
            gps_enabled: data.gps_enabled || false,
            gps_device_id: data.gps_device_id || "",
            gps_device_type: data.gps_device_type || "",
            immobilizer_enabled: data.immobilizer_enabled || false,
            ais140_compliant: data.ais140_compliant || false,
            branch_id: data.branch_id || "",
            cost_center_id: data.cost_center_id || "",
            current_driver_id: data.current_driver_id || "",
          });
        } else {
          setError("Failed to load vehicle data from backend.");
        }

        setDropdowns({
          drivers: (dRes.status === "fulfilled" && dRes.value?.data?.drivers) ? dRes.value.data.drivers.map((d: any) => ({ value: d.id, label: `${d.first_name} ${d.last_name}` })) : [],
          branches: (bRes.status === "fulfilled" && bRes.value?.data?.data) ? bRes.value.data.data.map((b: any) => ({ value: b.id, label: b.name })) : [],
          costCenters: (cRes.status === "fulfilled" && cRes.value?.data?.data) ? cRes.value.data.data.map((c: any) => ({ value: c.id, label: c.name })) : [],
        });
      } catch (err) {
        setError("A rendering mismatch occurred.");
      } finally {
        setLoading(false);
      }
    };
    fetchData();
  }, [id]);

  const updateField = (name: string, value: any) => {
    setForm(prev => ({ ...prev, [name]: value }));
  };

  const handleSubmit = async () => {
    setSaving(true);
    setSubmitError("");
    try {
      const payload = {
        ...form,
        year: parseInt(form.year) || 0,
        capacity_tonnage: parseFloat(form.capacity_tonnage) || 0,
        fuel_tank_capacity: parseFloat(form.fuel_tank_capacity) || 0,
        speed_limit: parseFloat(form.speed_limit) || 0,
        odometer: Math.round(parseFloat(form.odometer || "0") * 1000),
        last_service_odometer: parseInt(form.last_service_odometer) || 0,
        current_driver_id: form.current_driver_id || "00000000-0000-0000-0000-000000000000",
        // Format dates back to ISO
        insurance_expiry: form.insurance_expiry ? `${form.insurance_expiry}T00:00:00Z` : null,
        permit_expiry: form.permit_expiry ? `${form.permit_expiry}T00:00:00Z` : null,
        pollution_expiry: form.pollution_expiry ? `${form.pollution_expiry}T00:00:00Z` : null,
        fitness_expiry: form.fitness_expiry ? `${form.fitness_expiry}T00:00:00Z` : null,
        tax_expiry: form.tax_expiry ? `${form.tax_expiry}T00:00:00Z` : null,
        last_service_date: form.last_service_date ? `${form.last_service_date}T00:00:00Z` : null,
        next_service_date: form.next_service_date ? `${form.next_service_date}T00:00:00Z` : null,
      };
      await apiClient.put(`/admin/vehicles/${id}`, payload);
      navigate(`/admin/vehicles/${id}`);
    } catch (err: any) {
      setSubmitError(err.response?.data?.error || "Backend update failed.");
    } finally {
      setSaving(false);
    }
  };

  if (loading) return <div className="p-10 text-white flex items-center gap-4"><div className="w-4 h-4 border-2 border-primary border-t-transparent rounded-full animate-spin" /> Auditing asset state...</div>;
  if (error || !vehicle) return (
    <div className="min-h-screen bg-[#0a0a0b] flex items-center justify-center p-6 text-center">
      <div className="bg-[#111114] border border-white/5 p-8 rounded-3xl space-y-4 max-w-sm shadow-2xl">
        <AlertTriangle className="text-red-500 w-10 h-10 mx-auto" />
        <h2 className="text-white font-bold text-lg">Registry Fault</h2>
        <p className="text-text-muted text-sm">{error}</p>
        <button onClick={() => navigate(-1)} className="w-full py-3 bg-white/5 text-white rounded-xl text-xs font-bold mt-4">Return to Fleet</button>
      </div>
    </div>
  );

  return (
    <div className="min-h-screen bg-[#0a0a0b] pb-20">
      <div className="max-w-4xl mx-auto px-6 py-10">
        <header className="flex items-center justify-between mb-12">
          <div className="space-y-1">
             <button onClick={() => navigate(-1)} className="flex items-center gap-2 text-xs text-text-muted hover:text-white mb-4 transition-colors">
                <ArrowLeft className="w-3 h-3" /> Registry Overview
             </button>
             <h1 className="text-3xl font-bold text-white">Modify <span className="text-primary">{vehicle?.registration_number}</span></h1>
          </div>
          <button 
            onClick={handleSubmit}
            disabled={saving}
            className="px-8 py-3 bg-primary text-white rounded-xl text-sm font-bold hover:bg-primary/90 transition-all flex items-center gap-2 shadow-xl shadow-primary/10"
          >
            {saving ? "Saving..." : <><Save className="w-4 h-4" /> Save Configuration</>}
          </button>
        </header>

        {submitError && (
          <div className="mb-8 p-4 bg-red-500/10 border border-red-500/20 rounded-xl text-red-500 text-xs flex items-center gap-2">
            <ShieldAlert className="w-4 h-4" /> {submitError}
          </div>
        )}

        <div className="space-y-10">
          {/* Section: Core Specs */}
          <section className="bg-[#111114] border border-white/5 rounded-3xl p-8 space-y-8 shadow-sm">
            <div className="flex items-center gap-3">
              <Truck className="text-primary w-5 h-5" />
              <h2 className="text-white font-bold tracking-tight">Core Classification</h2>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
              <SelectInput label="Vehicle Type" name="vehicle_type" value={form.vehicle_type} options={VEHICLE_TYPES} onChange={(e) => updateField("vehicle_type", e.target.value)} />
              <SelectInput label="Fuel Profile" name="fuel_type" value={form.fuel_type} options={FUEL_TYPES} onChange={(e) => updateField("fuel_type", e.target.value)} />
              <SelectInput label="Ownership Model" name="ownership_type" value={form.ownership_type} options={OWNERSHIP_TYPES} onChange={(e) => updateField("ownership_type", e.target.value)} />
              <SelectInput label="Operational Status" name="status" value={form.status} options={STATUS_OPTIONS} onChange={(e) => updateField("status", e.target.value)} />
            </div>
          </section>

          {/* Section: Compliance & Validity */}
          <section className="bg-[#111114] border border-white/5 rounded-3xl p-8 space-y-8 shadow-sm">
            <div className="flex items-center gap-3">
              <ClipboardCheck className="text-primary w-5 h-5" />
              <h2 className="text-white font-bold tracking-tight">Compliance & Validity</h2>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
              <TextInput label="Insurance Policy #" name="insurance" value={form.insurance} onChange={(e) => updateField("insurance", e.target.value)} />
              <TextInput label="Insurance Expiry" name="insurance_expiry" type="date" value={form.insurance_expiry} onChange={(e) => updateField("insurance_expiry", e.target.value)} />
              <TextInput label="Permit Number" name="permit_number" value={form.permit_number} onChange={(e) => updateField("permit_number", e.target.value)} />
              <TextInput label="Permit Expiry" name="permit_expiry" type="date" value={form.permit_expiry} onChange={(e) => updateField("permit_expiry", e.target.value)} />
              <TextInput label="Pollution (PUC) Expiry" name="pollution_expiry" type="date" value={form.pollution_expiry} onChange={(e) => updateField("pollution_expiry", e.target.value)} />
              <TextInput label="Fitness Expiry" name="fitness_expiry" type="date" value={form.fitness_expiry} onChange={(e) => updateField("fitness_expiry", e.target.value)} />
              <TextInput label="Tax Validity Expiry" name="tax_expiry" type="date" value={form.tax_expiry} onChange={(e) => updateField("tax_expiry", e.target.value)} />
              <div className="flex flex-col gap-2">
                 <p className="text-[10px] text-text-muted font-bold uppercase tracking-widest">Registry ID:</p>
                 <span className="text-xs text-white">{vehicle?.id}</span>
              </div>
            </div>
          </section>

          {/* Section: Engineering */}
          <section className="bg-[#111114] border border-white/5 rounded-3xl p-8 space-y-8 shadow-sm">
            <div className="flex items-center gap-3">
              <Settings className="text-primary w-5 h-5" />
              <h2 className="text-white font-bold tracking-tight">Engineering Metadata</h2>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <TextInput label="Manufacturer" name="make" value={form.make} onChange={(e) => updateField("make", e.target.value)} />
              <TextInput label="Variant/Model" name="model" value={form.model} onChange={(e) => updateField("model", e.target.value)} />
              <TextInput label="Registry Year" name="year" value={form.year} onChange={(e) => updateField("year", e.target.value)} />
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
              <TextInput label="VIN Number" name="vin_number" value={form.vin_number} onChange={(e) => updateField("vin_number", e.target.value)} />
              <TextInput label="Chassis Number" name="chassis_number" value={form.chassis_number} onChange={(e) => updateField("chassis_number", e.target.value)} />
              <TextInput label="Engine Sequence" name="engine_number" value={form.engine_number} onChange={(e) => updateField("engine_number", e.target.value)} />
              <TextInput label="Registry Color" name="color" value={form.color} onChange={(e) => updateField("color", e.target.value)} />
            </div>
            <div className="grid grid-cols-1 md:grid-cols-4 gap-6 pt-4 border-t border-white/5">
              <TextInput label="Tonnage" name="capacity_tonnage" type="number" value={form.capacity_tonnage} onChange={(e) => updateField("capacity_tonnage", e.target.value)} />
              <TextInput label="Fuel Cap (L)" name="fuel_tank_capacity" type="number" value={form.fuel_tank_capacity} onChange={(e) => updateField("fuel_tank_capacity", e.target.value)} />
              <TextInput label="Odometer (Km)" name="odometer" type="number" value={form.odometer} onChange={(e) => updateField("odometer", e.target.value)} />
              <TextInput label="Speed Limit" name="speed_limit" type="number" value={form.speed_limit} onChange={(e) => updateField("speed_limit", e.target.value)} />
            </div>
          </section>

          {/* Section: Maintenance History */}
          <section className="bg-[#111114] border border-white/5 rounded-3xl p-8 space-y-8 shadow-sm">
            <div className="flex items-center gap-3">
              <Wrench className="text-primary w-5 h-5" />
              <h2 className="text-white font-bold tracking-tight">Maintenance History</h2>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
              <TextInput label="Last Service Date" name="last_service_date" type="date" value={form.last_service_date} onChange={(e) => updateField("last_service_date", e.target.value)} />
              <TextInput label="Last Service Odometer" name="last_service_odometer" type="number" value={form.last_service_odometer} onChange={(e) => updateField("last_service_odometer", e.target.value)} />
              <TextInput label="Next Service Due" name="next_service_date" type="date" value={form.next_service_date} onChange={(e) => updateField("next_service_date", e.target.value)} />
            </div>
          </section>

          {/* Section: Personnel & Assignment */}
          <section className="bg-[#111114] border border-white/5 rounded-3xl p-8 space-y-8 shadow-sm">
            <div className="flex items-center gap-3">
              <ShieldCheck className="text-primary w-5 h-5" />
              <h2 className="text-white font-bold tracking-tight">Personnel & Assignment</h2>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
              <SelectInput label="Assigned Guardian" name="current_driver_id" value={form.current_driver_id} options={dropdowns.drivers} onChange={(e) => updateField("current_driver_id", e.target.value)} placeholder="Unassigned" />
              <SelectInput label="Operational Branch" name="branch_id" value={form.branch_id} options={dropdowns.branches} onChange={(e) => updateField("branch_id", e.target.value)} />
              <SelectInput label="Cost Allocation" name="cost_center_id" value={form.cost_center_id} options={dropdowns.costCenters} onChange={(e) => updateField("cost_center_id", e.target.value)} />
              <div className="pt-6 border-t border-white/5 flex items-center gap-2 text-text-muted">
                 <LinkIcon className="w-4 h-4" />
                 <p className="text-xs font-bold uppercase tracking-widest">Parent Asset Node:</p>
                 <span className="text-xs text-white">{vehicle?.parent_vehicle_id || "None (Top Level)"}</span>
              </div>
            </div>
          </section>

          {/* Section: Telematics */}
          <section className="bg-[#111114] border border-white/5 rounded-3xl p-8 space-y-8 shadow-sm">
            <div className="flex items-center justify-between">
               <div className="flex items-center gap-3">
                <MapPin className="text-primary w-5 h-5" />
                <h2 className="text-white font-bold tracking-tight">Telematics Configuration</h2>
              </div>
              <ToggleInput label="" checked={form.gps_enabled} onChange={(v) => updateField("gps_enabled", v)} />
            </div>
            {form.gps_enabled && (
               <div className="grid grid-cols-1 md:grid-cols-2 gap-8 pt-6 border-t border-white/5">
                  <TextInput label="Device Serial (IMEI)" name="gps_device_id" value={form.gps_device_id} onChange={(e) => updateField("gps_device_id", e.target.value)} />
                  <TextInput label="Link Protocol" name="gps_device_type" value={form.gps_device_type} onChange={(e) => updateField("gps_device_type", e.target.value)} />
               </div>
            )}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-12 pt-4">
               <div className="flex items-center justify-between p-4 bg-white/5 rounded-2xl border border-white/5 transition-all hover:bg-white/10">
                  <div className="space-y-1">
                    <p className="text-sm font-bold text-white">Engine Immobilizer</p>
                    <p className="text-[10px] text-text-muted">Remote kill-switch capability</p>
                  </div>
                  <ToggleInput label="" checked={form.immobilizer_enabled} onChange={(v) => updateField("immobilizer_enabled", v)} />
               </div>
               <div className="flex items-center justify-between p-4 bg-white/5 rounded-2xl border border-white/5 transition-all hover:bg-white/10">
                  <div className="space-y-1">
                    <p className="text-sm font-bold text-white">AIS 140 Standard</p>
                    <p className="text-[10px] text-text-muted">Government safety compliance</p>
                  </div>
                  <ToggleInput label="" checked={form.ais140_compliant} onChange={(v) => updateField("ais140_compliant", v)} />
               </div>
            </div>
          </section>
        </div>
      </div>
    </div>
  );
}
