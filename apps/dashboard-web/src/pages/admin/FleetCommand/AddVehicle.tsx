import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import apiClient from "../../../services/apiClient";

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
  branch_id: string;
}

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
  branch_id: "",
};

export const AddVehicle: React.FC = () => {
  const navigate = useNavigate();
  const [form, setForm] = useState<FormState>(initialForm);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [step, setStep] = useState<1 | 2 | 3>(1);

  const handleChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>
  ) => {
    const { name, value } = e.target;
    setForm((prev) => ({ ...prev, [name]: value }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      const payload: any = {
        registration_number: form.registration_number.trim().toUpperCase(),
        vehicle_type: form.vehicle_type,
        fuel_type: form.fuel_type,
        ownership_type: form.ownership_type,
      };

      if (form.make) payload.make = form.make;
      if (form.model) payload.model = form.model;
      if (form.year) payload.year = parseInt(form.year);
      if (form.vin_number) payload.vin_number = form.vin_number.trim().toUpperCase();
      if (form.chassis_number) payload.chassis_number = form.chassis_number;
      if (form.engine_number) payload.engine_number = form.engine_number;
      if (form.capacity_tonnage) payload.capacity_tonnage = parseFloat(form.capacity_tonnage);
      if (form.fuel_tank_capacity) payload.fuel_tank_capacity = parseFloat(form.fuel_tank_capacity);
      if (form.gps_device_id) payload.gps_device_id = form.gps_device_id;
      if (form.color) payload.color = form.color;

      await apiClient.post("/admin/vehicles", payload);
      navigate("/admin/vehicles");
    } catch (err: any) {
      console.error("Failed to create vehicle:", err);
      setError(
        err.response?.data?.error ||
        err.message ||
        "Failed to create vehicle. Please try again."
      );
    } finally {
      setLoading(false);
    }
  };

  const isValidStep1 = form.registration_number && form.vehicle_type;
  const isComplete = isValidStep1;

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
        <div className="flex items-center gap-2">
          <span className={`text-xs px-2 py-1 rounded-full ${
            step === 1 ? "bg-primary/10 text-primary" : "bg-success/10 text-success"
          }`}>
            Step {step} of 3
          </span>
        </div>
      </header>

      {/* Form */}
      <div className="flex-1 overflow-y-auto custom-scrollbar">
        <form onSubmit={handleSubmit} className="max-w-4xl mx-auto w-full p-6 space-y-6">
          {error && (
            <div className="rounded-lg border border-danger/30 bg-danger/10 px-4 py-3 text-sm text-danger flex items-center gap-2">
              <span className="material-symbols-outlined text-sm">error</span>
              {error}
            </div>
          )}

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

          {/* ── Step 1: Registration (Required) ── */}
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
                {/* Registration Number */}
                <div className="space-y-1.5">
                  <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                    Registration Number <span className="text-danger">*</span>
                  </label>
                  <div className="relative">
                    <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted">
                      <span className="material-symbols-outlined text-[18px]">confirmation_number</span>
                    </span>
                    <input
                      type="text"
                      name="registration_number"
                      value={form.registration_number}
                      onChange={handleChange}
                      required
                      placeholder="e.g. MH01AB1234"
                      className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-4 text-sm text-white uppercase focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder-text-muted/50 font-mono"
                    />
                  </div>
                  <p className="text-[10px] text-text-muted">
                    As per RC — unique per organization
                  </p>
                </div>

                {/* Vehicle Type */}
                <div className="space-y-1.5">
                  <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                    Vehicle Type <span className="text-danger">*</span>
                  </label>
                  <div className="relative">
                    <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted">
                      <span className="material-symbols-outlined text-[18px]">category</span>
                    </span>
                    <select
                      name="vehicle_type"
                      value={form.vehicle_type}
                      onChange={handleChange}
                      required
                      className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-10 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none appearance-none"
                    >
                      <option value="">Select Type</option>
                      {VEHICLE_TYPES.map((t) => (
                        <option key={t.value} value={t.value}>{t.label}</option>
                      ))}
                    </select>
                    <span className="absolute inset-y-0 right-0 flex items-center pr-3 pointer-events-none text-text-muted">
                      <span className="material-symbols-outlined text-[18px]">expand_more</span>
                    </span>
                  </div>
                </div>

                {/* Fuel Type & Ownership - Side by side */}
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div className="space-y-1.5">
                    <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                      Fuel Type <span className="text-danger">*</span>
                    </label>
                    <div className="relative">
                      <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted">
                        <span className="material-symbols-outlined text-[18px]">local_gas_station</span>
                      </span>
                      <select
                        name="fuel_type"
                        value={form.fuel_type}
                        onChange={handleChange}
                        required
                        className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-10 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none appearance-none"
                      >
                        {FUEL_TYPES.map((f) => (
                          <option key={f.value} value={f.value}>{f.label}</option>
                        ))}
                      </select>
                      <span className="absolute inset-y-0 right-0 flex items-center pr-3 pointer-events-none text-text-muted">
                        <span className="material-symbols-outlined text-[18px]">expand_more</span>
                      </span>
                    </div>
                  </div>

                  <div className="space-y-1.5">
                    <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                      Ownership Type <span className="text-danger">*</span>
                    </label>
                    <div className="relative">
                      <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted">
                        <span className="material-symbols-outlined text-[18px]">home_work</span>
                      </span>
                      <select
                        name="ownership_type"
                        value={form.ownership_type}
                        onChange={handleChange}
                        required
                        className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-10 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none appearance-none"
                      >
                        {OWNERSHIP_TYPES.map((o) => (
                          <option key={o.value} value={o.value}>{o.label}</option>
                        ))}
                      </select>
                      <span className="absolute inset-y-0 right-0 flex items-center pr-3 pointer-events-none text-text-muted">
                        <span className="material-symbols-outlined text-[18px]">expand_more</span>
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* ── Step 2: Specifications (Optional) ── */}
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
                {/* Make, Model, Year */}
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                  <div className="space-y-1.5">
                    <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                      Make
                    </label>
                    <div className="relative">
                      <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted">
                        <span className="material-symbols-outlined text-[18px]">business</span>
                      </span>
                      <input
                        type="text"
                        name="make"
                        value={form.make}
                        onChange={handleChange}
                        placeholder="e.g. Tata, Ashok Leyland"
                        className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder-text-muted/50"
                      />
                    </div>
                  </div>
                  <div className="space-y-1.5">
                    <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                      Model
                    </label>
                    <input
                      type="text"
                      name="model"
                      value={form.model}
                      onChange={handleChange}
                      placeholder="e.g. 407, Eicher Pro"
                      className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder-text-muted/50"
                    />
                  </div>
                  <div className="space-y-1.5">
                    <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                      Year
                    </label>
                    <input
                      type="number"
                      name="year"
                      value={form.year}
                      onChange={handleChange}
                      placeholder="e.g. 2023"
                      className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder-text-muted/50"
                    />
                  </div>
                </div>

                {/* VIN, Chassis, Engine */}
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                  <div className="space-y-1.5">
                    <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                      VIN Number
                    </label>
                    <input
                      type="text"
                      name="vin_number"
                      value={form.vin_number}
                      onChange={handleChange}
                      placeholder="17-character VIN"
                      className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white uppercase font-mono focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder-text-muted/50"
                    />
                  </div>
                  <div className="space-y-1.5">
                    <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                      Chassis Number
                    </label>
                    <input
                      type="text"
                      name="chassis_number"
                      value={form.chassis_number}
                      onChange={handleChange}
                      placeholder="Chassis number"
                      className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white font-mono focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder-text-muted/50"
                    />
                  </div>
                  <div className="space-y-1.5">
                    <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                      Engine Number
                    </label>
                    <input
                      type="text"
                      name="engine_number"
                      value={form.engine_number}
                      onChange={handleChange}
                      placeholder="Engine number"
                      className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white font-mono focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder-text-muted/50"
                    />
                  </div>
                </div>

                {/* Capacity & Fuel Tank */}
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div className="space-y-1.5">
                    <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                      Capacity (Tonnes)
                    </label>
                    <input
                      type="number"
                      name="capacity_tonnage"
                      value={form.capacity_tonnage}
                      onChange={handleChange}
                      placeholder="e.g. 10"
                      step="0.1"
                      className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder-text-muted/50"
                    />
                  </div>
                  <div className="space-y-1.5">
                    <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                      Fuel Tank Capacity (Litres)
                    </label>
                    <input
                      type="number"
                      name="fuel_tank_capacity"
                      value={form.fuel_tank_capacity}
                      onChange={handleChange}
                      placeholder="e.g. 200"
                      step="0.1"
                      className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder-text-muted/50"
                    />
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* ── Step 3: Additional (Optional) ── */}
          {step === 3 && (
            <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
              <div className="px-6 py-4 border-b border-border-color bg-surface-darker flex items-center gap-3">
                <div className="w-8 h-8 rounded-full bg-amber-500/10 flex items-center justify-center text-amber-400 border border-amber-500/20">
                  <span className="material-symbols-outlined text-sm">settings</span>
                </div>
                <div>
                  <h3 className="text-white text-sm font-semibold">Additional Details</h3>
                  <p className="text-xs text-text-muted mt-0.5">
                    GPS device, color, and branch assignment
                  </p>
                </div>
              </div>
              <div className="p-6 space-y-6">
                {/* GPS Device */}
                <div className="space-y-1.5">
                  <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                    GPS Device ID
                  </label>
                  <div className="relative">
                    <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted">
                      <span className="material-symbols-outlined text-[18px]">gps_fixed</span>
                    </span>
                    <input
                      type="text"
                      name="gps_device_id"
                      value={form.gps_device_id}
                      onChange={handleChange}
                      placeholder="e.g. GPS-IMEI-12345678"
                      className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-4 text-sm text-white font-mono focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder-text-muted/50"
                    />
                  </div>
                  <p className="text-[10px] text-text-muted">
                    AIS-140 compliant GPS tracker IMEI or serial number
                  </p>
                </div>

                {/* Color */}
                <div className="space-y-1.5">
                  <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                    Color
                  </label>
                  <input
                    type="text"
                    name="color"
                    value={form.color}
                    onChange={handleChange}
                    placeholder="e.g. White, Blue"
                    className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder-text-muted/50"
                  />
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
                  onClick={() => setStep((s) => (s + 1) as 1 | 2 | 3)}
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
                  disabled={loading || !isComplete}
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
