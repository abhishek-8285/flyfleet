import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "../../services/api";

interface FormState {
  first_name: string;
  last_name: string;
  email: string;
  phone: string;
  employee_code: string;
  aadhaar_number: string;
  pan_number: string;
  date_of_birth: string;
  license_number: string;
  license_type: string;
  license_expiry: string;
  license_issue_date: string;
  license_authority: string;
  employment_type: string;
  experience_years: string;
  joining_date: string;
  address_street: string;
  address_city: string;
  address_state: string;
  address_pincode: string;
  address_country: string;
  emergency_name: string;
  emergency_phone: string;
  emergency_relation: string;
  blood_group: string;
  shift_type: string;
  behavior_monitoring: boolean;
  notes: string;
}

const initialForm: FormState = {
  first_name: "",
  last_name: "",
  email: "",
  phone: "",
  employee_code: "",
  aadhaar_number: "",
  pan_number: "",
  date_of_birth: "",
  license_number: "",
  license_type: "LMV",
  license_expiry: "",
  license_issue_date: "",
  license_authority: "",
  employment_type: "FULL_TIME",
  experience_years: "0",
  joining_date: new Date().toISOString().split("T")[0],
  address_street: "",
  address_city: "",
  address_state: "",
  address_pincode: "",
  address_country: "India",
  emergency_name: "",
  emergency_phone: "",
  emergency_relation: "",
  blood_group: "O+",
  shift_type: "MORNING",
  behavior_monitoring: true,
  notes: "",
};

const LICENSE_TYPES = [
  { value: "LMV", label: "LMV (Light Motor Vehicle)" },
  { value: "HMV", label: "HMV (Heavy Motor Vehicle)" },
  { value: "Transport", label: "Transport" },
];

const EMPLOYMENT_TYPES = [
  { value: "FULL_TIME", label: "Full Time" },
  { value: "CONTRACT", label: "Contract" },
];

const SHIFT_TYPES = [
  { value: "MORNING", label: "Morning" },
  { value: "AFTERNOON", label: "Afternoon" },
  { value: "NIGHT", label: "Night" },
  { value: "ROTATING", label: "Rotating" },
];

const BLOOD_GROUPS = ["A+", "A-", "B+", "B-", "AB+", "AB-", "O+", "O-"];

export const AddDriver: React.FC = () => {
  const navigate = useNavigate();
  const [currentStep, setCurrentStep] = useState(1);
  const [form, setForm] = useState<FormState>(initialForm);
  const [errors, setErrors] = useState<Partial<Record<keyof FormState, string>>>({});
  const [submitting, setSubmitting] = useState(false);

  const updateField = (field: keyof FormState, value: string | boolean) => {
    setForm((prev) => ({ ...prev, [field]: value }));
    if (errors[field]) setErrors((prev) => ({ ...prev, [field]: undefined }));
  };

  const validateStep = (step: number): boolean => {
    const newErrors: Partial<Record<keyof FormState, string>> = {};
    if (step === 1) {
      if (!form.first_name.trim()) newErrors.first_name = "Required";
      if (!form.last_name.trim()) newErrors.last_name = "Required";
      if (!form.email.trim()) newErrors.email = "Required";
      if (!form.phone.trim()) newErrors.phone = "Required";
    }
    if (step === 2) {
      if (!form.license_number.trim()) newErrors.license_number = "Required";
      if (!form.license_expiry) newErrors.license_expiry = "Required";
    }
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleNext = () => {
    if (validateStep(currentStep)) setCurrentStep((s) => Math.min(s + 1, 3));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validateStep(3)) return;
    setSubmitting(true);
    try {
      await api.createDriver({
        ...form,
        phone: form.phone.startsWith("+") ? form.phone : `+91${form.phone}`,
        license_expiry: new Date(form.license_expiry).toISOString(),
        date_of_birth: new Date(form.date_of_birth).toISOString(),
        experience_years: parseInt(form.experience_years) || 0,
        address: { street: form.address_street, city: form.address_city, state: form.address_state, pincode: form.address_pincode, country: form.address_country },
        emergency_contact: { name: form.emergency_name, phone: form.emergency_phone, relation: form.emergency_relation },
        metadata: { blood_group: form.blood_group, shift_type: form.shift_type, behavior_monitoring: form.behavior_monitoring, notes: form.notes, joining_date: form.joining_date }
      });
      navigate("/admin/drivers");
    } catch (error: any) {
      alert(error?.response?.data?.error || "Initialization failed.");
    } finally {
      setSubmitting(false);
    }
  };

  const renderInput = (label: string, field: keyof FormState, type: string = "text", placeholder?: string) => (
    <div className="space-y-1.5">
      <label className="text-[10px] font-black text-on-surface-variant uppercase tracking-widest block italic ml-1">
        {label}
      </label>
      <input
        type={type}
        value={(form[field] as string) || ""}
        onChange={(e) => updateField(field, e.target.value)}
        placeholder={placeholder}
        className={`w-full bg-surface-container border ${errors[field] ? 'border-rose-500/50' : 'border-outline'} text-white text-sm rounded-xl px-4 py-2.5 focus:ring-1 focus:ring-primary focus:border-primary outline-none transition-all placeholder:text-on-surface-variant/30 font-body shadow-inner`}
      />
      {errors[field] && <p className="text-[10px] text-rose-400 font-bold uppercase tracking-tighter ml-1">*{errors[field]}</p>}
    </div>
  );

  const renderSelect = (label: string, field: keyof FormState, options: { value: string; label: string }[]) => (
    <div className="space-y-1.5">
      <label className="text-[10px] font-black text-on-surface-variant uppercase tracking-widest block italic ml-1">
        {label}
      </label>
      <select
        value={(form[field] as string) || ""}
        onChange={(e) => updateField(field, e.target.value)}
        className="w-full bg-surface-container border border-outline text-white text-sm rounded-xl px-4 py-2.5 focus:ring-1 focus:ring-primary focus:border-primary outline-none transition-all cursor-pointer font-body shadow-inner appearance-none"
      >
        {options.map((opt) => (
          <option key={opt.value} value={opt.value} className="bg-surface-container-highest">{opt.label}</option>
        ))}
      </select>
    </div>
  );

  const steps = ["Personal Bio", "Operational Credentials", "Protocol Setup"];

  return (
    <div className="w-full max-w-5xl mx-auto py-8 px-4 relative">
       {/* Background Aura */}
      <div className="absolute top-0 right-0 w-[400px] h-[400px] bg-primary/5 blur-[120px] rounded-full pointer-events-none z-0"></div>

      {/* Header */}
      <div className="mb-10 relative z-10">
        <button
          onClick={() => navigate("/admin/drivers")}
          className="flex items-center gap-2 text-[10px] font-black text-on-surface-variant uppercase tracking-[0.2em] mb-4 hover:text-white transition-all italic group"
        >
          <span className="material-symbols-outlined text-[18px] group-hover:-translate-x-1 transition-transform">arrow_back</span>
          Terminate Request
        </button>
        <h1 className="text-4xl font-black text-white tracking-tighter uppercase italic">Initialize Operational Node</h1>
        <p className="text-on-surface-variant text-xs font-medium mt-2">Section {currentStep}: {steps[currentStep-1]}</p>
      </div>

      {/* Step Rail */}
      <div className="flex items-center gap-4 mb-10 relative z-10">
        {steps.map((title, i) => {
          const stepNum = i + 1;
          const isActive = stepNum === currentStep;
          const isDone = stepNum < currentStep;
          return (
            <React.Fragment key={i}>
              <div className="flex items-center gap-3">
                <div className={`size-10 rounded-xl border flex items-center justify-center font-black transition-all duration-500 shadow-xl ${
                  isDone ? 'bg-emerald-500 border-emerald-500 text-white' :
                  isActive ? 'bg-primary border-primary text-white shadow-primary/20' :
                  'bg-surface-container border-outline text-on-surface-variant opacity-40'
                }`}>
                  {isDone ? <span className="material-symbols-outlined text-lg">check</span> : stepNum}
                </div>
                <span className={`text-[10px] font-black uppercase tracking-widest hidden md:block transition-opacity duration-500 ${isActive ? 'opacity-100 text-white' : 'opacity-30 text-on-surface-variant'}`}>
                  {title}
                </span>
              </div>
              {i < 2 && <div className={`flex-1 h-0.5 rounded-full transition-all duration-700 ${isDone ? 'bg-emerald-500' : 'bg-outline'}`} />}
            </React.Fragment>
          );
        })}
      </div>

      {/* Form Terminal */}
      <form onSubmit={(e) => e.preventDefault()} className="bg-surface-container/40 backdrop-blur-md border border-outline rounded-3xl p-10 shadow-2xl relative z-10 animate-entrance">
        {currentStep === 1 && (
          <div className="space-y-8">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
              {renderInput("First Signature", "first_name", "text", "Aiden")}
              {renderInput("Last Signature", "last_name", "text", "Pierce")}
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
              {renderInput("Communication Hub (Email)", "email", "email", "aiden.p@obsidian.fleet")}
              {renderInput("Mobile Frequency", "phone", "tel", "9876543210")}
            </div>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
              {renderInput("Node Sequence ID", "employee_code", "text", "FF-NODE-88")}
              {renderInput("Origin Date (DOB)", "date_of_birth", "date")}
              {renderInput("Activation Date", "joining_date", "date")}
            </div>
            <div className="pt-4 border-t border-outline">
              <h3 className="text-[11px] font-black text-white uppercase tracking-[0.2em] mb-6 italic">Geographic Location</h3>
              <div className="space-y-6">
                {renderInput("Sector Address (Street)", "address_street", "text", "Terminal Alpha, Block 7")}
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                  {renderInput("City Node", "address_city", "text", "Neo-Delhi")}
                  {renderInput("State Cluster", "address_state", "text", "NCR")}
                  {renderInput("Registry Pincode", "address_pincode", "text", "110001")}
                </div>
              </div>
            </div>
          </div>
        )}

        {currentStep === 2 && (
          <div className="space-y-8 text-white">
            <h3 className="text-[11px] font-black text-white uppercase tracking-[0.2em] mb-6 italic">Operational Credentials</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
              {renderInput("Registry License Number", "license_number", "text", "MH01-202X000123")}
              {renderSelect("Asset Handling Grade", "license_type", LICENSE_TYPES)}
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
              {renderInput("License Issuance", "license_issue_date", "date")}
              {renderInput("Registry Expiry", "license_expiry", "date")}
            </div>
            {renderInput("Issuing Terminal (Authority)", "license_authority", "text", "RTO Sector-14")}
            <div className="pt-4 border-t border-outline">
              <h3 className="text-[11px] font-black text-white uppercase tracking-[0.2em] mb-6 italic">Identity Encryption</h3>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                {renderInput("Biometric Registry (Aadhaar)", "aadhaar_number", "text", "4455-8899-2211")}
                {renderInput("Taxation ID (PAN)", "pan_number", "text", "ABCDE1234F")}
              </div>
            </div>
          </div>
        )}

        {currentStep === 3 && (
          <div className="space-y-8">
             <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
              {renderSelect("Deployment Class", "employment_type", EMPLOYMENT_TYPES)}
              {renderInput("Flight Experience (Years)", "experience_years", "number", "5")}
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
              {renderSelect("Pulse Shift Pattern", "shift_type", SHIFT_TYPES)}
              {renderSelect("Biological Group", "blood_group", BLOOD_GROUPS.map(bg => ({value: bg, label: bg})))}
            </div>
            <div className="p-6 bg-primary/5 border border-primary/20 rounded-2xl flex items-center justify-between">
              <div>
                <p className="text-white font-black text-xs uppercase italic tracking-widest">Telemetry Monitoring</p>
                <p className="text-[10px] text-on-surface-variant font-medium mt-1">Enable real-time behavior tracking and incident logging.</p>
              </div>
              <button 
                onClick={() => updateField("behavior_monitoring", !form.behavior_monitoring)}
                className={`w-12 h-6 rounded-full transition-all relative ${form.behavior_monitoring ? 'bg-primary' : 'bg-surface-container-highest'}`}
              >
                <div className={`absolute top-1 size-4 rounded-full bg-white transition-all ${form.behavior_monitoring ? 'left-7' : 'left-1'}`} />
              </button>
            </div>
            <div className="space-y-2">
              <label className="text-[10px] font-black text-on-surface-variant uppercase tracking-widest italic ml-1">Terminal Notes</label>
              <textarea
                value={form.notes}
                onChange={(e) => updateField("notes", e.target.value)}
                rows={4}
                className="w-full bg-surface-container border border-outline text-white text-sm rounded-2xl px-4 py-3 focus:ring-1 focus:ring-primary focus:border-primary outline-none transition-all placeholder:text-on-surface-variant/30 font-body shadow-inner resize-none"
                placeholder="Additional operational instructions..."
              />
            </div>
          </div>
        )}

        {/* Footer Navigation */}
        <div className="flex items-center justify-between mt-12 pt-8 border-t border-outline">
          <button
            type="button"
            onClick={() => currentStep > 1 ? setCurrentStep(s => s - 1) : navigate("/admin/drivers")}
            className="px-6 py-2.5 text-[10px] font-black text-on-surface-variant uppercase tracking-[0.2em] hover:text-white transition-all italic"
          >
            {currentStep === 1 ? 'Abort Session' : 'Previous Logic'}
          </button>
          
          <div className="flex items-center gap-4">
            {currentStep < 3 ? (
              <button
                type="button"
                onClick={handleNext}
                className="px-8 py-2.5 bg-surface-container-highest text-white text-[10px] font-black uppercase tracking-widest rounded-xl hover:bg-primary transition-all shadow-lg active:scale-95 italic"
              >
                Next Sequence
              </button>
            ) : (
              <button
                onClick={handleSubmit}
                disabled={submitting}
                className="px-10 py-2.5 bg-emerald-500 hover:bg-emerald-600 disabled:opacity-50 text-white text-[10px] font-black uppercase tracking-widest rounded-xl transition-all shadow-lg shadow-emerald-500/20 flex items-center gap-3 active:scale-95 italic"
              >
                {submitting && <span className="material-symbols-outlined animate-spin text-[16px]">progress_activity</span>}
                {submitting ? "Compiling..." : "Initialize Node"}
              </button>
            )}
          </div>
        </div>
      </form>
    </div>
  );
};
