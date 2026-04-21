import { useState, useEffect } from "react";
import { Link, useNavigate } from "react-router-dom";
import {
  ArrowLeftIcon,
  CheckCircleIcon,
  IdentificationIcon,
  ShieldCheckIcon,
  UserIcon,
  XMarkIcon,
} from "@heroicons/react/24/outline";
import api from "../../../services/api";

export default function AddDriver() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const [formData, setFormData] = useState({
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
    branch_id: "",
    cost_center_id: "",
  });

  const [dropdowns, setDropdowns] = useState({
    branches: [] as any[],
    costCenters: [] as any[],
  });

  useEffect(() => {
    fetchMetadata();
  }, []);

  const fetchMetadata = async () => {
    try {
      const [bResp, cResp] = await Promise.all([
        api.get("/admin/branches"),
        api.get("/admin/cost-centers"),
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
        branches: ensureArray(bResp, "branches").map((b: any) => ({
          value: b.id,
          label: b.name,
        })),
        costCenters: ensureArray(cResp, "cost_centers").map((c: any) => ({
          value: c.id,
          label: c.name,
        })),
      });
    } catch (err) {
      console.error("Failed to load metadata:", err);
    }
  };

  const handleChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>
  ) => {
    const { name, value, type } = e.target;
    const val = type === "checkbox" ? (e.target as HTMLInputElement).checked : value;
    setFormData((prev) => ({ ...prev, [name]: val }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (!formData.first_name || !formData.last_name || !formData.phone) {
      setError("First name, last name, and phone are required");
      return;
    }
    if (!formData.license_number) {
      setError("License number is required");
      return;
    }

    setLoading(true);
    try {
      const payload = {
        ...formData,
        phone: formData.phone.startsWith("+") ? formData.phone : `+91${formData.phone}`,
        experience_years: parseInt(formData.experience_years) || 0,
        address: {
          street: formData.address_street,
          city: formData.address_city,
          state: formData.address_state,
          pincode: formData.address_pincode,
          country: formData.address_country,
        },
        emergency_contact: {
          name: formData.emergency_name,
          phone: formData.emergency_phone.startsWith("+") ? formData.emergency_phone : `+91${formData.emergency_phone}`,
          relation: formData.emergency_relation,
        },
        metadata: {
          blood_group: formData.blood_group,
          shift_type: formData.shift_type,
          behavior_monitoring: formData.behavior_monitoring,
          notes: formData.notes,
        },
      };
      await api.createDriver(payload);
      navigate("/admin/drivers");
    } catch (err: any) {
      setError(err.response?.data?.error || "Failed to initialize personnel node");
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
            <Link to="/admin/drivers" className="hover:text-white">
              Personnel
            </Link>
            <span>/</span>
            <span className="text-primary font-medium">Add</span>
          </div>
          <h1 className="text-2xl font-bold text-white">Initialize Personnel Node</h1>
        </div>
        <div className="flex items-center gap-3">
          <Link
            to="/admin/drivers"
            className="px-4 py-2 bg-surface-dark hover:bg-surface-darker border border-border-color text-text-muted hover:text-white text-xs font-semibold rounded transition-colors"
          >
            Cancel
          </Link>
          <button
            onClick={() => handleSubmit(new Event("submit") as unknown as React.FormEvent)}
            disabled={loading}
            className="px-4 py-2 bg-primary hover:bg-primary-dark text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors shadow-lg shadow-primary/20 disabled:opacity-50"
          >
            <span className="material-symbols-outlined text-[16px]">add</span>
            {loading ? "Initializing..." : "Create Profile"}
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
        {/* ── Identity Signature ── */}
        <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
          <div className="px-6 py-4 border-b border-border-color bg-surface-darker flex items-center gap-3">
            <div className="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center text-primary border border-primary/20">
              <span className="material-symbols-outlined text-sm">person</span>
            </div>
            <div>
              <h3 className="text-white text-sm font-semibold">Identity Signature</h3>
              <p className="text-xs text-text-muted mt-0.5">
                Primary biometric and contact identification
              </p>
            </div>
          </div>
          <div className="p-6 grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                First Name <span className="text-danger">*</span>
              </label>
              <div className="relative">
                <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted">
                  <span className="material-symbols-outlined text-[18px]">badge</span>
                </span>
                <input
                  type="text"
                  name="first_name"
                  value={formData.first_name}
                  onChange={handleChange}
                  placeholder="e.g. John"
                  className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder-text-muted/50"
                />
              </div>
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Last Name <span className="text-danger">*</span>
              </label>
              <input
                type="text"
                name="last_name"
                value={formData.last_name}
                onChange={handleChange}
                placeholder="e.g. Doe"
                className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder-text-muted/50"
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Phone Number <span className="text-danger">*</span>
              </label>
              <div className="relative">
                <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted">
                  <span className="material-symbols-outlined text-[18px]">call</span>
                </span>
                <input
                  type="tel"
                  name="phone"
                  value={formData.phone}
                  onChange={handleChange}
                  placeholder="+91 98765 43210"
                  className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder-text-muted/50"
                />
              </div>
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Email Address
              </label>
              <div className="relative">
                <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted">
                  <span className="material-symbols-outlined text-[18px]">mail</span>
                </span>
                <input
                  type="email"
                  name="email"
                  value={formData.email}
                  onChange={handleChange}
                  placeholder="john.doe@example.com"
                  className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 pl-10 pr-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder-text-muted/50"
                />
              </div>
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Date of Birth
              </label>
              <input
                type="date"
                name="date_of_birth"
                value={formData.date_of_birth}
                onChange={handleChange}
                className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none [color-scheme:dark]"
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Blood Group
              </label>
              <select
                name="blood_group"
                value={formData.blood_group}
                onChange={handleChange}
                className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none appearance-none"
              >
                {["O+", "O-", "A+", "A-", "B+", "B-", "AB+", "AB-"].map(bg => (
                  <option key={bg} value={bg}>{bg}</option>
                ))}
              </select>
            </div>
          </div>
        </div>

        {/* ── Credentials & Compliance ── */}
        <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
          <div className="px-6 py-4 border-b border-border-color bg-surface-darker flex items-center gap-3">
            <div className="w-8 h-8 rounded-full bg-indigo-500/10 flex items-center justify-center text-indigo-400 border border-indigo-500/20">
              <span className="material-symbols-outlined text-sm">assignment_ind</span>
            </div>
            <div>
              <h3 className="text-white text-sm font-semibold">Credentials & Compliance</h3>
              <p className="text-xs text-text-muted mt-0.5">
                License details and regulatory identification
              </p>
            </div>
          </div>
          <div className="p-6 grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                License Number <span className="text-danger">*</span>
              </label>
              <input
                type="text"
                name="license_number"
                value={formData.license_number}
                onChange={handleChange}
                placeholder="DL-1234567890"
                className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none"
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                License Type
              </label>
              <select
                name="license_type"
                value={formData.license_type}
                onChange={handleChange}
                className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none appearance-none"
              >
                <option value="LMV">LMV (Light Motor Vehicle)</option>
                <option value="HMV">HMV (Heavy Motor Vehicle)</option>
                <option value="HGV">HGV (Heavy Goods Vehicle)</option>
                <option value="MCWG">MCWG (Motorcycle with Gear)</option>
              </select>
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                License Expiry
              </label>
              <input
                type="date"
                name="license_expiry"
                value={formData.license_expiry}
                onChange={handleChange}
                className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none [color-scheme:dark]"
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Aadhaar Number
              </label>
              <input
                type="text"
                name="aadhaar_number"
                value={formData.aadhaar_number}
                onChange={handleChange}
                placeholder="1234 5678 9012"
                className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none"
              />
            </div>
          </div>
        </div>

        {/* ── Employment & Context ── */}
        <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
          <div className="px-6 py-4 border-b border-border-color bg-surface-darker flex items-center gap-3">
            <div className="w-8 h-8 rounded-full bg-amber-500/10 flex items-center justify-center text-amber-400 border border-amber-500/20">
              <span className="material-symbols-outlined text-sm">work</span>
            </div>
            <div>
              <h3 className="text-white text-sm font-semibold">Employment Context</h3>
              <p className="text-xs text-text-muted mt-0.5">
                Organization details and deployment parameters
              </p>
            </div>
          </div>
          <div className="p-6 grid grid-cols-1 md:grid-cols-3 gap-6">
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Employee Code
              </label>
              <input
                type="text"
                name="employee_code"
                value={formData.employee_code}
                onChange={handleChange}
                placeholder="EMP-101"
                className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none"
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Employment Type
              </label>
              <select
                name="employment_type"
                value={formData.employment_type}
                onChange={handleChange}
                className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none appearance-none"
              >
                <option value="FULL_TIME">Full Time</option>
                <option value="CONTRACT">Contract</option>
                <option value="PART_TIME">Part Time</option>
              </select>
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Experience (Years)
              </label>
              <input
                type="number"
                name="experience_years"
                value={formData.experience_years}
                onChange={handleChange}
                min="0"
                className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none"
              />
            </div>
          </div>
        </div>

        {/* ── Emergency Contact ── */}
        <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
          <div className="px-6 py-4 border-b border-border-color bg-surface-darker flex items-center gap-3">
            <div className="w-8 h-8 rounded-full bg-rose-500/10 flex items-center justify-center text-rose-400 border border-rose-500/20">
              <span className="material-symbols-outlined text-sm">emergency</span>
            </div>
            <div>
              <h3 className="text-white text-sm font-semibold">Emergency Contact</h3>
              <p className="text-xs text-text-muted mt-0.5">
                Primary contact for critical contingencies
              </p>
            </div>
          </div>
          <div className="p-6 grid grid-cols-1 md:grid-cols-3 gap-6">
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Contact Name
              </label>
              <input
                type="text"
                name="emergency_name"
                value={formData.emergency_name}
                onChange={handleChange}
                placeholder="Name"
                className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none"
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Relation
              </label>
              <input
                type="text"
                name="emergency_relation"
                value={formData.emergency_relation}
                onChange={handleChange}
                placeholder="e.g. Spouse"
                className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none"
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Phone Number
              </label>
              <input
                type="tel"
                name="emergency_phone"
                value={formData.emergency_phone}
                onChange={handleChange}
                placeholder="+91..."
                className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none"
              />
            </div>
          </div>
        </div>

        {/* ── Organizational Hierarchy ── */}
        <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
          <div className="px-6 py-4 border-b border-border-color bg-surface-darker flex items-center gap-3">
            <div className="w-8 h-8 rounded-full bg-emerald-500/10 flex items-center justify-center text-emerald-400 border border-emerald-500/20">
              <span className="material-symbols-outlined text-sm">account_tree</span>
            </div>
            <div>
              <h3 className="text-white text-sm font-semibold">Organizational Hierarchy</h3>
              <p className="text-xs text-text-muted mt-0.5">
                Mapping personnel to organizational nodes
              </p>
            </div>
          </div>
          <div className="p-6 grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Branch Assignment
              </label>
              <select
                name="branch_id"
                value={formData.branch_id}
                onChange={handleChange}
                className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none appearance-none"
              >
                <option value="">Select Branch</option>
                {dropdowns.branches.map((b) => (
                  <option key={b.value} value={b.value}>{b.label}</option>
                ))}
              </select>
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Cost Center
              </label>
              <select
                name="cost_center_id"
                value={formData.cost_center_id}
                onChange={handleChange}
                className="w-full bg-surface-darker border border-border-color rounded-lg py-2.5 px-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none appearance-none"
              >
                <option value="">Select Cost Center</option>
                {dropdowns.costCenters.map((c) => (
                  <option key={c.value} value={c.value}>{c.label}</option>
                ))}
              </select>
            </div>
          </div>
        </div>
      </form>
    </div>
  );
}
