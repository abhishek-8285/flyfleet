import { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { 
  ArrowLeft, 
  ChevronRight, 
  Plus, 
  Trash2, 
  ShieldCheck, 
  Cpu, 
  Activity, 
  Globe, 
  Zap,
  Truck,
  Settings,
  Info,
  Badge,
  AlertTriangle
} from 'lucide-react';
import apiClient from '../../../services/apiClient';
import { FUEL_TYPES, ODOMETER_UNITS, OWNERSHIP_TYPES, VEHICLE_TYPES } from './constants';
import { toKm } from './utils';

interface DropdownOption {
  value: string;
  label: string;
}

export default function AddVehicle() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [gpsDeviceTypes, setGpsDeviceTypes] = useState<DropdownOption[]>([]);
  const [dropdowns, setDropdowns] = useState({
    drivers: [] as DropdownOption[],
    branches: [] as DropdownOption[],
    costCenters: [] as DropdownOption[],
  });

  const [formData, setFormData] = useState({
    registration_number: '',
    vehicle_type: '',
    fuel_type: 'DIESEL',
    ownership_type: 'OWNED',
    make: '',
    model: '',
    year: new Date().getFullYear().toString(),
    vin_number: '',
    chassis_number: '',
    engine_number: '',
    capacity_tonnage: '',
    fuel_tank_capacity: '',
    gps_device_id: '',
    gps_enabled: false,
    gps_device_type: 'AIS140_HARDWIRED',
    color: '',
    odometer: '',
    odometer_unit: 'km',
    branch_id: '',
    cost_center_id: '',
    current_driver_id: '',
    speed_limit: '0',
  });

  useEffect(() => {
    fetchMetadata();
  }, []);

  const fetchMetadata = async () => {
    try {
      const [dResp, bResp, cResp, tResp] = await Promise.all([
        apiClient.get('/admin/drivers?limit=500'),
        apiClient.get('/admin/branches'),
        apiClient.get('/admin/cost-centers'),
        apiClient.get('/ui/settings/gps-types'),
      ]);

      const ensureArray = (resp: any, key?: string) => {
        if (!resp) return [];
        if (Array.isArray(resp)) return resp;
        if (key && Array.isArray(resp[key])) return resp[key];
        if (Array.isArray(resp.data)) return resp.data;
        if (resp.data && key && Array.isArray(resp.data[key])) return resp.data[key];
        return [];
      };

      setDropdowns({
        drivers: ensureArray(dResp, 'drivers').map((d: any) => ({
          value: d.id,
          label: `${d.first_name || ''} ${d.last_name || ''}`.trim() || d.name || d.id,
        })),
        branches: ensureArray(bResp, 'data').map((b: any) => ({
          value: b.id,
          label: b.name,
        })),
        costCenters: ensureArray(cResp, 'data').map((c: any) => ({
          value: c.id,
          label: c.name,
        })),
      });
      setGpsDeviceTypes(ensureArray(tResp));
    } catch (err) {
      console.error('Failed to load metadata:', err);
    }
  };

  const handleChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>
  ) => {
    const { name, value, type } = e.target;
    const val = type === 'checkbox' ? (e.target as HTMLInputElement).checked : value;
    setFormData((prev) => ({ ...prev, [name]: val }));
  };

  const handleSubmit = async (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    setError('');

    if (!formData.registration_number) {
      setError('Registration number is required');
      return;
    }
    if (!formData.vehicle_type) {
      setError('Vehicle type is required');
      return;
    }

    setLoading(true);
    try {
      const payload = {
        ...formData,
        registration_number: formData.registration_number.toUpperCase().replace(/\s/g, ''),
        year: parseInt(formData.year) || 0,
        capacity_tonnage: parseFloat(formData.capacity_tonnage) || 0,
        fuel_tank_capacity: parseFloat(formData.fuel_tank_capacity) || 0,
        speed_limit: parseFloat(formData.speed_limit) || 0,
        odometer: Math.round(
          toKm(parseFloat(formData.odometer) || 0, formData.odometer_unit) * 1000
        ),
        branch_id: formData.branch_id || undefined,
        cost_center_id: formData.cost_center_id || undefined,
        current_driver_id: formData.current_driver_id || undefined,
      };
      await apiClient.post('/admin/vehicles', payload);
      navigate('/admin/vehicles');
    } catch (err: any) {
      setError(err.response?.data?.error || err.message || 'Failed to initialize vehicle node');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex-1 flex flex-col bg-background-dark overflow-hidden font-sans relative">
      {/* --- Standardized Header --- */}
      <header className="h-16 flex items-center justify-between px-6 border-b border-border-color bg-surface-darker flex-shrink-0 z-30">
        <div className="flex items-center gap-4">
          <Link to="/admin/vehicles" className="p-2 hover:bg-white/5 rounded-full transition-colors text-text-muted">
            <ArrowLeft className="w-5 h-5" />
          </Link>
          <div>
            <div className="flex items-center gap-2 text-[10px] text-text-muted uppercase tracking-[0.2em] font-black mb-0.5">
              <span>Fleet Management</span>
              <ChevronRight className="w-3 h-3 opacity-30" />
              <span className="text-white">Add New Vehicle</span>
            </div>
            <h2 className="text-white text-lg font-bold flex items-center gap-2 uppercase tracking-tight">
              Register New Vehicle
              <span className="px-2 py-0.5 rounded text-[10px] font-black uppercase tracking-wider border bg-primary/10 text-primary border-primary/20">
                New
              </span>
            </h2>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <Link
            to="/admin/vehicles"
            className="px-4 py-2 bg-surface-dark border border-border-color text-text-muted hover:text-white text-xs font-semibold rounded-lg transition-all"
          >
            Cancel
          </Link>
          <button
            onClick={() => handleSubmit()}
            disabled={loading}
            className="px-6 py-2 bg-primary hover:bg-primary-dark text-white text-xs font-semibold rounded-lg flex items-center gap-2 shadow-lg shadow-primary/20 transition-all disabled:opacity-50"
          >
            {loading ? <div className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin" /> : <Plus className="w-3.5 h-3.5" />}
            {loading ? 'Processing...' : 'Save Vehicle'}
          </button>
        </div>
      </header>

      {/* --- Main Area --- */}
      <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
        <div className="max-w-6xl mx-auto space-y-6 pb-32">
          
          {/* Signal Summary Cards (Parity with VehicleDetail) */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-2">
            <div className="bg-surface-dark border border-border-color p-5 rounded-xl flex items-center gap-4">
              <div className="p-3 rounded-xl bg-primary/10 text-primary">
                <Truck className="w-6 h-6" />
              </div>
              <div>
                <div className="text-[10px] text-text-muted uppercase font-bold tracking-widest mb-1">Registration</div>
                <div className="text-xl font-bold text-white uppercase">{formData.registration_number || 'AWAITING'}</div>
              </div>
            </div>

            <div className="bg-surface-dark border border-border-color p-5 rounded-xl flex items-center gap-4">
              <div className={`p-3 rounded-xl ${formData.gps_enabled ? 'bg-emerald-500/10 text-emerald-500' : 'bg-rose-500/10 text-rose-500'}`}>
                <Zap className="w-6 h-6" />
              </div>
              <div>
                <div className="text-[10px] text-text-muted uppercase font-bold tracking-widest mb-1">Tracking</div>
                <div className={`text-xl font-bold ${formData.gps_enabled ? 'text-emerald-500' : 'text-rose-500'}`}>{formData.gps_enabled ? 'ENABLED' : 'OFF'}</div>
              </div>
            </div>

            <div className="bg-surface-dark border border-border-color p-5 rounded-xl flex items-center gap-4">
              <div className="p-3 rounded-xl bg-amber-500/10 text-amber-500">
                <Fuel className="w-6 h-6" />
              </div>
              <div>
                <div className="text-[10px] text-text-muted uppercase font-bold tracking-widest mb-1">Fuel Type</div>
                <div className="text-xl font-bold text-white uppercase">{formData.fuel_type || 'DIESEL'}</div>
              </div>
            </div>

            <div className="bg-surface-dark border border-border-color p-5 rounded-xl flex items-center gap-4">
              <div className="p-3 rounded-xl bg-white/5 text-white">
                <Scale className="w-6 h-6" />
              </div>
              <div>
                <div className="text-[10px] text-text-muted uppercase font-bold tracking-widest mb-1">Capacity</div>
                <div className="text-xl font-bold text-white uppercase">{formData.capacity_tonnage || '0'} <span className="text-sm font-normal text-text-muted uppercase tracking-widest">TON</span></div>
              </div>
            </div>
          </div>

          {error && (
            <div className="p-4 bg-danger/10 border border-danger/20 rounded-xl flex items-center gap-3 text-danger text-[10px] font-black uppercase tracking-widest animate-in fade-in slide-in-from-top-2">
              <AlertTriangle className="w-5 h-5" />
              {error}
            </div>
          )}

          <div className="grid grid-cols-12 gap-6">
            {/* Form Side */}
            <div className="col-span-12 lg:col-span-8 flex flex-col gap-6">
        {/* 1. Vehicle Details */}
        <section className="bg-surface-dark border border-border-color rounded-xl overflow-hidden shadow-xl">
           <div className="px-6 py-4 border-b border-border-color bg-white/[0.02] flex items-center gap-3">
              <Truck className="w-5 h-5 text-primary" />
              <h3 className="text-xs font-black uppercase tracking-[0.2em] text-white">Vehicle Details</h3>
           </div>
           <div className="p-8 grid grid-cols-1 md:grid-cols-2 gap-x-8 gap-y-6">
              <div className="space-y-1.5">
                <label className="text-[10px] font-black text-gray-600 uppercase tracking-widest block">
                  Registration Number <span className="text-danger">*</span>
                </label>
                <div className="relative">
                  <Badge className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
                  <input
                    type="text"
                    name="registration_number"
                    value={formData.registration_number}
                    onChange={handleChange}
                    placeholder="e.g. MH01AB1234"
                    className="w-full bg-black/40 border border-white/5 rounded-xl py-3 pl-10 pr-4 text-sm text-white focus:outline-none focus:border-primary transition-all placeholder-gray-800"
                  />
                </div>
              </div>
              <div className="space-y-1.5">
                <label className="text-[10px] font-black text-gray-600 uppercase tracking-widest block">
                  Vehicle Type <span className="text-danger">*</span>
                </label>
                <select
                  name="vehicle_type"
                  value={formData.vehicle_type}
                  onChange={handleChange}
                  className="w-full bg-black/40 border border-white/5 rounded-xl py-3 px-4 text-sm text-white focus:outline-none focus:border-primary transition-all appearance-none cursor-pointer"
                >
                  <option value="">Select Type</option>
                  {VEHICLE_TYPES.map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}
                    </option>
                  ))}
                </select>
              </div>
              <div className="space-y-1.5">
                <label className="text-[10px] font-black text-gray-600 uppercase tracking-widest block">
                  Financial Model
                </label>
                <select
                  name="ownership_type"
                  value={formData.ownership_type}
                  onChange={handleChange}
                  className="w-full bg-black/40 border border-white/5 rounded-xl py-3 px-4 text-sm text-white focus:outline-none focus:border-primary transition-all appearance-none cursor-pointer"
                >
                  {OWNERSHIP_TYPES.map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}
                    </option>
                  ))}
                </select>
              </div>
              <div className="space-y-1.5">
                <label className="text-[10px] font-black text-gray-600 uppercase tracking-widest block">
                  Fuel Type
                </label>
                <select
                  name="fuel_type"
                  value={formData.fuel_type}
                  onChange={handleChange}
                  className="w-full bg-black/40 border border-white/5 rounded-xl py-3 px-4 text-sm text-white focus:outline-none focus:border-primary transition-all appearance-none cursor-pointer"
                >
                  {FUEL_TYPES.map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}
                    </option>
                  ))}
                </select>
              </div>
           </div>
        </section>

        {/* 2. Performance Details */}
        <section className="bg-surface-dark border border-border-color rounded-xl overflow-hidden shadow-xl">
           <div className="px-6 py-4 border-b border-border-color bg-white/[0.02] flex items-center gap-3">
              <Cpu className="w-5 h-5 text-amber-500" />
              <h3 className="text-xs font-black uppercase tracking-[0.2em] text-white">Performance Details</h3>
           </div>
           <div className="p-8">
              <div className="grid grid-cols-1 md:grid-cols-3 gap-8 mb-10">
                 <div className="space-y-1.5">
                   <label className="text-[10px] font-black text-gray-600 uppercase tracking-widest block">Manufacturer</label>
                   <input
                     type="text"
                     name="make"
                     value={formData.make}
                     onChange={handleChange}
                     placeholder="e.g. Tata"
                     className="w-full bg-black/40 border border-white/5 rounded-xl py-3 px-4 text-sm text-white focus:outline-none focus:border-primary transition-all"
                   />
                 </div>
                 <div className="space-y-1.5">
                   <label className="text-[10px] font-black text-gray-600 uppercase tracking-widest block">Variant / Model</label>
                   <input
                     type="text"
                     name="model"
                     value={formData.model}
                     onChange={handleChange}
                     placeholder="e.g. 407 Gold"
                     className="w-full bg-black/40 border border-white/5 rounded-xl py-3 px-4 text-sm text-white focus:outline-none focus:border-primary transition-all"
                   />
                 </div>
                 <div className="space-y-1.5">
                   <label className="text-[10px] font-black text-gray-600 uppercase tracking-widest block">Deployment Year</label>
                   <input
                     type="number"
                     name="year"
                     value={formData.year}
                     onChange={handleChange}
                     className="w-full bg-black/40 border border-white/5 rounded-xl py-3 px-4 text-sm text-white focus:outline-none focus:border-primary transition-all"
                   />
                 </div>
              </div>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-8 mb-10">
                 <div className="space-y-1.5">
                   <label className="text-[10px] font-black text-gray-600 uppercase tracking-widest block">VIN Number (17 Digits)</label>
                   <input
                     type="text"
                     name="vin_number"
                     value={formData.vin_number}
                     onChange={handleChange}
                     placeholder="Enter Chassis VIN"
                     className="w-full bg-black/40 border border-white/5 rounded-xl py-3 px-4 text-sm text-white focus:outline-none focus:border-primary transition-all"
                   />
                 </div>
                 <div className="space-y-1.5">
                   <label className="text-[10px] font-black text-gray-600 uppercase tracking-widest block">Chassis Number</label>
                   <input
                     type="text"
                     name="engine_number"
                     value={formData.engine_number}
                     onChange={handleChange}
                     placeholder="ERX-12345"
                     className="w-full bg-black/40 border border-white/5 rounded-xl py-3 px-4 text-sm text-white focus:outline-none focus:border-primary transition-all"
                   />
                 </div>
              </div>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-6 pt-10 border-t border-white/5">
                 <div className="space-y-1.5">
                   <label className="text-[10px] font-black text-gray-600 uppercase tracking-widest block">Payload (Tons)</label>
                   <input
                     type="number"
                     name="capacity_tonnage"
                     value={formData.capacity_tonnage}
                     onChange={handleChange}
                     placeholder="0.0"
                     step="0.1"
                     className="w-full bg-black/40 border border-white/5 rounded-xl py-3 px-4 text-sm text-white focus:outline-none focus:border-primary transition-all"
                   />
                 </div>
                 <div className="space-y-1.5">
                   <label className="text-[10px] font-black text-gray-600 uppercase tracking-widest block">Fuel Res (L)</label>
                   <input
                     type="number"
                     name="fuel_tank_capacity"
                     value={formData.fuel_tank_capacity}
                     onChange={handleChange}
                     placeholder="0"
                     className="w-full bg-black/40 border border-white/5 rounded-xl py-3 px-4 text-sm text-white focus:outline-none focus:border-primary transition-all"
                   />
                 </div>
                 <div className="space-y-1.5">
                   <label className="text-[10px] font-black text-gray-600 uppercase tracking-widest block">Initial Odometer</label>
                   <div className="grid grid-cols-3 gap-2">
                     <input
                       type="number"
                       name="odometer"
                       value={formData.odometer}
                       onChange={handleChange}
                       placeholder="0"
                       className="col-span-2 bg-black/40 border border-white/5 rounded-xl py-3 px-4 text-sm text-white focus:outline-none focus:border-primary transition-all"
                     />
                     <select
                       name="odometer_unit"
                       value={formData.odometer_unit}
                       onChange={handleChange}
                       className="bg-black/40 border border-white/5 rounded-xl py-3 px-2 text-[10px] font-black text-white focus:outline-none focus:border-primary transition-all appearance-none text-center cursor-pointer"
                     >
                       {ODOMETER_UNITS.map((opt) => (
                         <option key={opt.value} value={opt.value}>{opt.label}</option>
                       ))}
                     </select>
                   </div>
                 </div>
                 <div className="space-y-1.5">
                   <label className="text-[10px] font-black text-gray-600 uppercase tracking-widest block">Speed Limit (km/h)</label>
                   <input
                     type="number"
                     name="speed_limit"
                     value={formData.speed_limit}
                     onChange={handleChange}
                     placeholder="80"
                     className="w-full bg-black/40 border border-white/5 rounded-xl py-3 px-4 text-sm text-white focus:outline-none focus:border-primary transition-all"
                   />
                 </div>
              </div>
           </div>
        </section>

        {/* 3. Telemetry Uplink */}
        <section className="bg-surface-dark border border-border-color rounded-xl overflow-hidden shadow-xl">
           <div className="px-6 py-4 border-b border-border-color bg-white/[0.02] flex items-center justify-between">
              <div className="flex items-center gap-3">
                 <Activity className="w-5 h-5 text-emerald-500" />
                 <h3 className="text-xs font-black uppercase tracking-[0.2em] text-white">Tracking Configuration</h3>
              </div>
           </div>
           <div className="p-8 space-y-8">
              <div className="p-8 rounded-2xl bg-primary/5 border border-primary/20 flex items-center justify-between">
                 <div className="space-y-1">
                    <h4 className="text-xs font-black uppercase tracking-widest flex items-center gap-2">
                       <Globe className="w-4 h-4 text-primary" /> Tracking Gateway
                    </h4>
                    <p className="text-[10px] text-gray-600 uppercase font-bold tracking-widest leading-none">Enable real-time positional broadcast</p>
                 </div>
                 <label className="relative inline-flex items-center cursor-pointer">
                   <input
                     type="checkbox"
                     name="gps_enabled"
                     checked={formData.gps_enabled}
                     onChange={handleChange}
                     className="sr-only peer"
                   />
                   <div className="w-11 h-6 bg-black/40 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-gray-700 after:border-gray-600 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary peer-checked:after:bg-white shadow-inner transition-all"></div>
                 </label>
              </div>

              {formData.gps_enabled && (
                <div className="grid grid-cols-1 md:grid-cols-2 gap-8 animate-in fade-in slide-in-from-top-2 duration-300">
                  <div className="space-y-1.5">
                    <label className="text-[10px] font-black text-gray-600 uppercase tracking-widest block">Device IMEI</label>
                    <input
                      type="text"
                      name="gps_device_id"
                      value={formData.gps_device_id}
                      onChange={handleChange}
                      placeholder="15-digit IMEI"
                      className="w-full bg-black/40 border border-white/5 rounded-xl py-3 px-4 text-sm text-white focus:outline-none focus:border-primary transition-all placeholder-gray-800"
                    />
                  </div>
                  <div className="space-y-1.5">
                    <label className="text-[10px] font-black text-gray-600 uppercase tracking-widest block">Device Type</label>
                    <select
                      name="gps_device_type"
                      value={formData.gps_device_type}
                      onChange={handleChange}
                      className="w-full bg-black/40 border border-white/5 rounded-xl py-3 px-4 text-sm text-white focus:outline-none focus:border-primary transition-all appearance-none cursor-pointer"
                    >
                      {gpsDeviceTypes.map((opt) => (
                        <option key={opt.value} value={opt.value}>
                          {opt.label}
                        </option>
                      ))}
                    </select>
                  </div>
                </div>
              )}
           </div>
        </section>

        {/* 4. Hierarchy & Custody */}
        <section className="bg-surface-dark border border-border-color rounded-xl overflow-hidden shadow-xl">
           <div className="px-6 py-4 border-b border-border-color bg-white/[0.02] flex items-center gap-3">
              <Settings className="w-5 h-5 text-indigo-400" />
              <h3 className="text-xs font-black uppercase tracking-[0.2em] text-white">Assignment & Branch</h3>
           </div>
           <div className="p-8 grid grid-cols-1 md:grid-cols-3 gap-8">
              <div className="space-y-1.5">
                <label className="text-[10px] font-black text-gray-600 uppercase tracking-widest block">Operational Branch</label>
                <select
                  name="branch_id"
                  value={formData.branch_id}
                  onChange={handleChange}
                  className="w-full bg-black/40 border border-white/5 rounded-xl py-3 px-4 text-sm text-white focus:outline-none focus:border-primary transition-all appearance-none cursor-pointer"
                >
                  <option value="">Ungrouped</option>
                  {dropdowns.branches.map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}
                    </option>
                  ))}
                </select>
              </div>
              <div className="space-y-1.5">
                <label className="text-[10px] font-black text-gray-600 uppercase tracking-widest block">Cost Allocation</label>
                <select
                  name="cost_center_id"
                  value={formData.cost_center_id}
                  onChange={handleChange}
                  className="w-full bg-black/40 border border-white/5 rounded-xl py-3 px-4 text-sm text-white focus:outline-none focus:border-primary transition-all appearance-none cursor-pointer"
                >
                  <option value="">Default</option>
                  {dropdowns.costCenters.map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}
                    </option>
                  ))}
                </select>
              </div>
              <div className="space-y-1.5">
                <label className="text-[10px] font-black text-gray-600 uppercase tracking-widest block">Assigned Driver</label>
                <select
                  name="current_driver_id"
                  value={formData.current_driver_id}
                  onChange={handleChange}
                  className="w-full bg-black/40 border border-white/5 rounded-xl py-3 px-4 text-sm text-white focus:outline-none focus:border-primary transition-all appearance-none cursor-pointer"
                >
                  <option value="">Unassigned</option>
                  {dropdowns.drivers.map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}
                    </option>
                  ))}
                </select>
              </div>
           </div>
        </section>

            </div>

            {/* Preview Side */}
            <div className="col-span-12 lg:col-span-4 space-y-6">
              <section className="bg-surface-dark border border-border-color rounded-xl overflow-hidden shadow-xl p-8 relative overflow-hidden group min-h-[400px] flex flex-col justify-center items-center text-center">
                <div className="absolute top-6 left-8 flex items-center gap-3">
                  <div className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
                  <span className="text-[10px] font-black text-white uppercase tracking-[0.3em]">
                    Live Vehicle Preview
                  </span>
                </div>

                <div className="w-full space-y-8 relative z-10">
                  <div className="space-y-4">
                    <div className="bg-white px-8 py-3 rounded-lg shadow-2xl border-4 border-black/80 flex items-center justify-center gap-4 mx-auto w-fit transform -rotate-1 group-hover:rotate-0 transition-transform duration-500">
                      <div className="w-3 h-3 rounded-full bg-blue-600" />
                      <span className="text-black font-mono font-black text-3xl tracking-tight leading-none">
                        {formData.registration_number.toUpperCase() || 'MH 01 AB 1234'}
                      </span>
                    </div>
                    <div>
                      <h4 className="text-xl font-black text-white uppercase tracking-tight">
                        {formData.make || 'MANUFACTURER'} {formData.model || 'MODEL'}
                      </h4>
                      <div className="inline-block px-3 py-1 bg-white/5 rounded text-[10px] font-black text-gray-400 uppercase tracking-widest mt-2">
                        {formData.vehicle_type || 'Awaiting Classification'}
                      </div>
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div className="p-4 rounded-xl bg-black/40 border border-white/5 backdrop-blur-md">
                      <span className="text-[9px] text-gray-600 font-black uppercase tracking-widest block mb-1">
                        Capacity
                      </span>
                      <span className="text-lg font-black text-white">
                        {formData.capacity_tonnage || '0.0'}
                        <span className="text-[8px] font-bold ml-1 text-gray-700">TON</span>
                      </span>
                    </div>
                    <div className="p-4 rounded-xl bg-black/40 border border-white/5 backdrop-blur-md">
                      <span className="text-[9px] text-gray-600 font-black uppercase tracking-widest block mb-1">
                        Status
                      </span>
                      <span className="text-lg font-black text-emerald-500">ACTIVE</span>
                    </div>
                  </div>
                </div>

                <div className="absolute inset-0 opacity-[0.03] pointer-events-none" style={{ backgroundImage: 'radial-gradient(circle, white 1px, transparent 1px)', backgroundSize: '24px 24px' }} />
              </section>

              <section className="bg-surface-dark border border-border-color rounded-xl p-6 shadow-xl">
                 <h3 className="text-[10px] font-black text-text-muted uppercase tracking-[0.2em] flex items-center gap-2 mb-6">
                    <Settings className="w-4 h-4" />
                    Verification Status
                 </h3>
                 <div className="space-y-4">
                    {[
                      { label: 'Details Verified', val: formData.registration_number ? 'YES' : 'PENDING', s: formData.registration_number ? 'ok' : 'warn' },
                      { label: 'Telemetry Hook', val: formData.gps_enabled ? 'ARMED' : 'DISABLED', s: formData.gps_enabled ? 'ok' : 'warn' },
                      { label: 'Branch Assigned', val: formData.branch_id ? 'YES' : 'UNGROUPED', s: formData.branch_id ? 'ok' : 'warn' },
                    ].map((h, i) => (
                      <div key={i} className="flex justify-between items-center">
                         <span className="text-[10px] text-text-muted font-bold uppercase tracking-tight">{h.label}</span>
                         <span className={`text-[10px] font-black uppercase tracking-tighter ${h.s === 'ok' ? 'text-emerald-500' : 'text-amber-500'}`}>{h.val}</span>
                      </div>
                    ))}
                 </div>
              </section>
            </div>
          </div>
        </div>
      </div>
    </div>
);
}
