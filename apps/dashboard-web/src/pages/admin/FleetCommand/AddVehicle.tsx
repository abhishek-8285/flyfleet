import { XMarkIcon } from '@heroicons/react/24/outline';
import { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import api from '../../../services/api';
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
        api.get('/admin/drivers?limit=500'),
        api.get('/admin/branches'),
        api.get('/admin/cost-centers'),
        api.get('/ui/settings/gps-types'),
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
        branches: ensureArray(bResp, 'branches').map((b: any) => ({
          value: b.id,
          label: b.name,
        })),
        costCenters: ensureArray(cResp, 'cost_centers').map((c: any) => ({
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

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
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
        // Filter out empty UUID strings to prevent backend decoding errors
        branch_id: formData.branch_id || undefined,
        cost_center_id: formData.cost_center_id || undefined,
        current_driver_id: formData.current_driver_id || undefined,
      };
      await api.createVehicle(payload);
      navigate('/admin/vehicles');
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to initialize vehicle node');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="p-10 max-w-7xl mx-auto space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <div className="flex items-center gap-2 text-xs text-text-muted mb-1">
            <Link to="/admin/vehicles" className="hover:text-white">
              Fleet
            </Link>
            <span>/</span>
            <span className="text-primary font-medium">Add</span>
          </div>
          <h1 className="text-2xl font-bold text-white">Add Vehicle</h1>
        </div>
        <div className="flex items-center gap-3">
          <Link
            to="/admin/vehicles"
            className="px-4 py-2 bg-surface-dark hover:bg-surface-darker border border-border-color text-text-muted hover:text-white text-xs font-semibold rounded transition-colors"
          >
            Cancel
          </Link>
          <button
            onClick={() => handleSubmit(new Event('submit') as unknown as React.FormEvent)}
            disabled={loading}
            className="px-4 py-2 bg-primary hover:bg-primary-dark text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors shadow-lg shadow-primary/20 disabled:opacity-50"
          >
            <span className="material-symbols-outlined text-[16px]">add</span>
            {loading ? 'Initializing...' : 'Create Vehicle'}
          </button>
        </div>
      </div>

      {error && (
        <div className="p-4 bg-danger/10 border border-danger/20 rounded-lg flex items-center gap-2 text-danger text-sm">
          <XMarkIcon className="w-5 h-5" />
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-6 pb-12">
        {/* ── Registry Identity ── */}
        <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
          <div className="px-6 py-4 border-b border-border-color bg-surface-darker flex items-center gap-3">
            <div className="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center text-primary border border-primary/20">
              <span className="material-symbols-outlined text-sm">app_registration</span>
            </div>
            <div>
              <h3 className="text-white text-sm font-semibold">Registry Identity</h3>
              <p className="text-xs text-text-muted mt-0.5">
                Core identification and RTO registration details
              </p>
            </div>
          </div>
          <div className="p-6 grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Registration Number <span className="text-danger">*</span>
              </label>
              <div className="relative">
                <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted">
                  <span className="material-symbols-outlined text-[18px]">badge</span>
                </span>
                <input
                  type="text"
                  name="registration_number"
                  value={formData.registration_number}
                  onChange={handleChange}
                  placeholder="e.g. MH01AB1234"
                  className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder-text-muted/50"
                />
              </div>
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Registry Classification <span className="text-danger">*</span>
              </label>
              <div className="relative">
                <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted">
                  <span className="material-symbols-outlined text-[18px]">category</span>
                </span>
                <select
                  name="vehicle_type"
                  value={formData.vehicle_type}
                  onChange={handleChange}
                  className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-10 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none appearance-none"
                >
                  <option value="">Select Type</option>
                  {VEHICLE_TYPES.map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}
                    </option>
                  ))}
                </select>
                <span className="absolute inset-y-0 right-0 flex items-center pr-3 pointer-events-none text-text-muted">
                  <span className="material-symbols-outlined text-[18px]">expand_more</span>
                </span>
              </div>
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Financial Model
              </label>
              <div className="relative">
                <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted">
                  <span className="material-symbols-outlined text-[18px]">payments</span>
                </span>
                <select
                  name="ownership_type"
                  value={formData.ownership_type}
                  onChange={handleChange}
                  className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-10 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none appearance-none"
                >
                  {OWNERSHIP_TYPES.map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}
                    </option>
                  ))}
                </select>
                <span className="absolute inset-y-0 right-0 flex items-center pr-3 pointer-events-none text-text-muted">
                  <span className="material-symbols-outlined text-[18px]">expand_more</span>
                </span>
              </div>
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Energy Protocol
              </label>
              <div className="relative">
                <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted">
                  <span className="material-symbols-outlined text-[18px]">ev_station</span>
                </span>
                <select
                  name="fuel_type"
                  value={formData.fuel_type}
                  onChange={handleChange}
                  className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-10 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none appearance-none"
                >
                  {FUEL_TYPES.map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}
                    </option>
                  ))}
                </select>
                <span className="absolute inset-y-0 right-0 flex items-center pr-3 pointer-events-none text-text-muted">
                  <span className="material-symbols-outlined text-[18px]">expand_more</span>
                </span>
              </div>
            </div>
          </div>
        </div>

        {/* ── Engineering Metadata ── */}
        <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
          <div className="px-6 py-4 border-b border-border-color bg-surface-darker flex items-center gap-3">
            <div className="w-8 h-8 rounded-full bg-indigo-500/10 flex items-center justify-center text-indigo-400 border border-indigo-500/20">
              <span className="material-symbols-outlined text-sm">precision_manufacturing</span>
            </div>
            <div>
              <h3 className="text-white text-sm font-semibold">Engineering Metadata</h3>
              <p className="text-xs text-text-muted mt-0.5">
                Technical specifications and hardware vectors
              </p>
            </div>
          </div>
          <div className="p-6 grid grid-cols-1 md:grid-cols-3 gap-6">
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Manufacturer
              </label>
              <input
                type="text"
                name="make"
                value={formData.make}
                onChange={handleChange}
                placeholder="e.g. Tata"
                className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none"
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Variant / Model
              </label>
              <input
                type="text"
                name="model"
                value={formData.model}
                onChange={handleChange}
                placeholder="e.g. 407 Gold"
                className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none"
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Deployment Year
              </label>
              <input
                type="number"
                name="year"
                value={formData.year}
                onChange={handleChange}
                className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none"
              />
            </div>
            <div className="space-y-1.5 md:col-span-2">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                VIN Signal (17 Digits)
              </label>
              <input
                type="text"
                name="vin_number"
                value={formData.vin_number}
                onChange={handleChange}
                placeholder="Enter Chassis VIN"
                className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none"
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Engine Number
              </label>
              <input
                type="text"
                name="engine_number"
                value={formData.engine_number}
                onChange={handleChange}
                placeholder="ERX-12345"
                className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none"
              />
            </div>
          </div>
        </div>

        {/* ── Operational Metrics ── */}
        <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
          <div className="px-6 py-4 border-b border-border-color bg-surface-darker flex items-center gap-3">
            <div className="w-8 h-8 rounded-full bg-amber-500/10 flex items-center justify-center text-amber-400 border border-amber-500/20">
              <span className="material-symbols-outlined text-sm">monitoring</span>
            </div>
            <div>
              <h3 className="text-white text-sm font-semibold">Operational Metrics</h3>
              <p className="text-xs text-text-muted mt-0.5">
                Capacity limits and telemetry baselines
              </p>
            </div>
          </div>
          <div className="p-6 grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Net Capacity (Tonnes)
              </label>
              <div className="relative">
                <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted">
                  <span className="material-symbols-outlined text-[18px]">weight</span>
                </span>
                <input
                  type="number"
                  name="capacity_tonnage"
                  value={formData.capacity_tonnage}
                  onChange={handleChange}
                  placeholder="0.0"
                  step="0.1"
                  className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none"
                />
              </div>
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Fuel Reservoir (Litres)
              </label>
              <div className="relative">
                <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted">
                  <span className="material-symbols-outlined text-[18px]">oil_barrel</span>
                </span>
                <input
                  type="number"
                  name="fuel_tank_capacity"
                  value={formData.fuel_tank_capacity}
                  onChange={handleChange}
                  placeholder="0"
                  className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none"
                />
              </div>
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Initial Odometer
              </label>
              <div className="grid grid-cols-3 gap-2">
                <input
                  type="number"
                  name="odometer"
                  value={formData.odometer}
                  onChange={handleChange}
                  placeholder="0"
                  className="col-span-2 bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none"
                />
                <select
                  name="odometer_unit"
                  value={formData.odometer_unit}
                  onChange={handleChange}
                  className="bg-surface-darker border border-border-color rounded-lg py-2.5 px-2 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none"
                >
                  {ODOMETER_UNITS.map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}
                    </option>
                  ))}
                </select>
              </div>
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Neural Speed Limit (km/h)
              </label>
              <div className="relative">
                <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted">
                  <span className="material-symbols-outlined text-[18px]">speed</span>
                </span>
                <input
                  type="number"
                  name="speed_limit"
                  value={formData.speed_limit}
                  onChange={handleChange}
                  placeholder="80"
                  className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none"
                />
              </div>
            </div>
          </div>
        </div>

        {/* ── Telemetry Uplink ── */}
        <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
          <div className="px-6 py-4 border-b border-border-color bg-surface-darker flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-full bg-emerald-500/10 flex items-center justify-center text-emerald-400 border border-emerald-500/20">
                <span className="material-symbols-outlined text-sm">sensors</span>
              </div>
              <div>
                <h3 className="text-white text-sm font-semibold">Telemetry Uplink</h3>
                <p className="text-xs text-text-muted mt-0.5">
                  Broadcast positional data to the control tower
                </p>
              </div>
            </div>
            <label className="relative inline-flex items-center cursor-pointer">
              <input
                type="checkbox"
                name="gps_enabled"
                checked={formData.gps_enabled}
                onChange={handleChange}
                className="sr-only peer"
              />
              <div className="w-11 h-6 bg-surface-darker peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-text-muted after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary peer-checked:after:bg-white"></div>
            </label>
          </div>
          {formData.gps_enabled && (
            <div className="p-6 grid grid-cols-1 md:grid-cols-2 gap-6 animate-in slide-in-from-top-2 duration-300">
              <div className="space-y-1.5">
                <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                  Device Serial (IMEI)
                </label>
                <input
                  type="text"
                  name="gps_device_id"
                  value={formData.gps_device_id}
                  onChange={handleChange}
                  placeholder="15-digit IMEI"
                  className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none"
                />
              </div>
              <div className="space-y-1.5">
                <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                  Link Protocol
                </label>
                <select
                  name="gps_device_type"
                  value={formData.gps_device_type}
                  onChange={handleChange}
                  className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none"
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

        {/* ── Hierarchy & Custody ── */}
        <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
          <div className="px-6 py-4 border-b border-border-color bg-surface-darker flex items-center gap-3">
            <div className="w-8 h-8 rounded-full bg-cyan-500/10 flex items-center justify-center text-cyan-400 border border-cyan-500/20">
              <span className="material-symbols-outlined text-sm">account_tree</span>
            </div>
            <div>
              <h3 className="text-white text-sm font-semibold">Hierarchy & Custody</h3>
              <p className="text-xs text-text-muted mt-0.5">
                Organizational placement and primary custodian
              </p>
            </div>
          </div>
          <div className="p-6 grid grid-cols-1 md:grid-cols-3 gap-6">
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Operational Branch
              </label>
              <select
                name="branch_id"
                value={formData.branch_id}
                onChange={handleChange}
                className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none appearance-none"
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
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Cost Allocation
              </label>
              <select
                name="cost_center_id"
                value={formData.cost_center_id}
                onChange={handleChange}
                className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none appearance-none"
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
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Guardian Driver
              </label>
              <select
                name="current_driver_id"
                value={formData.current_driver_id}
                onChange={handleChange}
                className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none appearance-none"
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
        </div>

        {/* ── Asset Visualization ── (Integrated Preview) */}
        <div className="bg-gradient-to-br from-primary/5 to-purple-500/5 rounded-2xl border border-primary/10 p-8 flex flex-col items-center justify-center relative overflow-hidden group">
          <div className="absolute top-4 left-6 flex items-center gap-2">
            <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" />
            <span className="text-[10px] font-bold text-text-muted uppercase tracking-widest">
              Real-time Visualization engine
            </span>
          </div>

          <div className="w-full max-w-sm space-y-6 relative z-10">
            <div className="flex flex-col items-center">
              <div className="bg-white px-6 py-2 rounded-md shadow-lg border border-black/10 flex items-center gap-3 mb-4 transform -rotate-1">
                <div className="w-2 h-2 rounded-full bg-blue-600" />
                <span className="text-black font-mono font-black text-xl tracking-tight">
                  {formData.registration_number.toUpperCase() || 'MH 01 AB 1234'}
                </span>
              </div>
              <h4 className="text-sm font-bold text-white tracking-wide">
                {formData.make} {formData.model}
              </h4>
              <p className="text-[10px] text-text-muted uppercase font-bold tracking-widest mt-1">
                {formData.vehicle_type || 'Awaiting Classification'}
              </p>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="p-4 rounded-xl bg-surface-dark/50 border border-white/5 backdrop-blur-sm">
                <span className="text-[10px] text-text-muted font-bold uppercase block mb-1">
                  Payload
                </span>
                <span className="text-lg font-black text-white">
                  {formData.capacity_tonnage || '0'}
                  <span className="text-xs font-medium ml-1 text-text-muted">TONS</span>
                </span>
              </div>
              <div className="p-4 rounded-xl bg-surface-dark/50 border border-white/5 backdrop-blur-sm">
                <span className="text-[10px] text-text-muted font-bold uppercase block mb-1">
                  Telemetry
                </span>
                <span
                  className={`text-[10px] font-black uppercase flex items-center gap-1 ${formData.gps_enabled ? 'text-emerald-400' : 'text-amber-400'}`}
                >
                  <span className="material-symbols-outlined text-[14px]">
                    {formData.gps_enabled ? 'wifi' : 'wifi_off'}
                  </span>
                  {formData.gps_enabled ? 'Online' : 'Offline'}
                </span>
              </div>
            </div>
          </div>
        </div>
      </form>
    </div>
  );
}
