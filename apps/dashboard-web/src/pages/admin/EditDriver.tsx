import React, { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { 
  ArrowLeft, 
  Save, 
  User, 
  ShieldCheck, 
  Briefcase, 
  HeartPulse, 
  AlertTriangle,
  IdCard,
  MapPin,
  Calendar,
  Activity,
  ShieldAlert,
  Stethoscope,
  Fingerprint,
  Droplets,
  FileText
} from "lucide-react";
import apiClient from "../../services/apiClient";
import { TextInput, SelectInput, ToggleInput } from "../../components/common/FormFields";
import { DocumentVault } from "../../components/common/DocumentVault";

export const EditDriver: React.FC = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [driver, setDriver] = useState<any>(null);
  const [error, setError] = useState<string | null>(null);
  const [submitError, setSubmitError] = useState("");

  const [form, setForm] = useState({
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
    medical_checkup_date: "",
    police_verification_date: "",
    police_verification_status: "PENDING",
    blood_group: "",
    address_street: "",
    address_city: "",
    address_state: "",
    address_pincode: "",
    address_country: "India",
    emergency_name: "",
    emergency_phone: "",
    emergency_relation: "",
    shift_type: "MORNING",
    behavior_monitoring: true,
    notes: "",
  });

  useEffect(() => {
    const fetchData = async () => {
      setLoading(true);
      try {
        const res = await apiClient.get(`/admin/drivers/${id}`);
        const data = res.data;
        setDriver(data);
        
        const formatDate = (dateStr: string) => dateStr ? dateStr.split("T")[0] : "";

        setForm({
          first_name: data.first_name || "",
          last_name: data.last_name || "",
          email: data.email || "",
          phone: data.phone || "",
          employee_code: data.employee_code || "",
          aadhaar_number: data.aadhaar_number || "",
          pan_number: data.pan_number || "",
          date_of_birth: formatDate(data.date_of_birth),
          license_number: data.license_number || "",
          license_type: data.license_type || "LMV",
          license_expiry: formatDate(data.license_expiry),
          license_issue_date: formatDate(data.license_issue_date),
          license_authority: data.license_authority || "",
          employment_type: data.employment_type || "FULL_TIME",
          experience_years: (data.experience_years || 0).toString(),
          medical_checkup_date: formatDate(data.medical_checkup_date),
          police_verification_date: formatDate(data.police_verification_date),
          police_verification_status: data.police_verification_status || "PENDING",
          blood_group: data.blood_group || "",
          address_street: data.address?.street || "",
          address_city: data.address?.city || "",
          address_state: data.address?.state || "",
          address_pincode: data.address?.pincode || "",
          address_country: data.address?.country || "India",
          emergency_name: data.emergency_contact?.name || "",
          emergency_phone: data.emergency_contact?.phone || "",
          emergency_relation: data.emergency_contact?.relation || "",
          shift_type: data.metadata?.shift_type || "MORNING",
          behavior_monitoring: data.metadata?.behavior_monitoring ?? true,
          notes: data.metadata?.notes || "",
        });
      } catch (err) {
        setError("Failed to audit personnel record.");
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
        experience_years: parseInt(form.experience_years) || 0,
        date_of_birth: form.date_of_birth ? `${form.date_of_birth}T00:00:00Z` : null,
        license_expiry: form.license_expiry ? `${form.license_expiry}T00:00:00Z` : null,
        license_issue_date: form.license_issue_date ? `${form.license_issue_date}T00:00:00Z` : null,
        medical_checkup_date: form.medical_checkup_date ? `${form.medical_checkup_date}T00:00:00Z` : null,
        police_verification_date: form.police_verification_date ? `${form.police_verification_date}T00:00:00Z` : null,
        address: {
          street: form.address_street,
          city: form.address_city,
          state: form.address_state,
          pincode: form.address_pincode,
          country: form.address_country,
        },
        emergency_contact: {
          name: form.emergency_name,
          phone: form.emergency_phone,
          relation: form.emergency_relation,
        },
        metadata: {
          shift_type: form.shift_type,
          behavior_monitoring: form.behavior_monitoring,
          notes: form.notes,
        }
      };
      await apiClient.put(`/admin/drivers/${id}`, payload);
      navigate(`/admin/drivers/${id}`);
    } catch (err: any) {
      setSubmitError(err.response?.data?.error || "Backend update rejected.");
    } finally {
      setSaving(false);
    }
  };

  if (loading) return <div className="p-10 text-white flex items-center gap-4"><div className="w-4 h-4 border-2 border-primary border-t-transparent rounded-full animate-spin" /> Verifying identity node...</div>;
  if (error || !driver) return (
    <div className="min-h-screen bg-[#0a0a0b] flex items-center justify-center p-6 text-center">
      <div className="bg-[#111114] border border-white/5 p-8 rounded-3xl space-y-4 max-w-sm shadow-2xl">
        <AlertTriangle className="text-red-500 w-10 h-10 mx-auto" />
        <h2 className="text-white font-bold text-lg">Personnel Access Fault</h2>
        <p className="text-text-muted text-sm">{error}</p>
        <button onClick={() => navigate(-1)} className="w-full py-3 bg-white/5 text-white rounded-xl text-xs font-bold mt-4">Return to Registry</button>
      </div>
    </div>
  );

  return (
    <div className="min-h-screen bg-[#0a0a0b] pb-20">
      <div className="max-w-4xl mx-auto px-6 py-10">
        <header className="flex items-center justify-between mb-12">
          <div className="space-y-1">
             <button onClick={() => navigate(-1)} className="flex items-center gap-2 text-xs text-text-muted hover:text-white mb-4 transition-colors">
                <ArrowLeft className="w-3 h-3" /> Back to Profile
             </button>
             <h1 className="text-3xl font-bold text-white">Edit <span className="text-primary">Personnel</span> Profile</h1>
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
          {/* Section: Identity */}
          <section className="bg-[#111114] border border-white/5 rounded-3xl p-8 space-y-8 shadow-sm">
            <div className="flex items-center gap-3">
              <User className="text-primary w-5 h-5" />
              <h2 className="text-white font-bold tracking-tight">Personal Identity</h2>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
              <TextInput label="First Name" name="first_name" value={form.first_name} onChange={(e) => updateField("first_name", e.target.value)} />
              <TextInput label="Last Name" name="last_name" value={form.last_name} onChange={(e) => updateField("last_name", e.target.value)} />
              <TextInput label="Email" name="email" type="email" value={form.email} onChange={(e) => updateField("email", e.target.value)} />
              <TextInput label="Phone" name="phone" type="tel" value={form.phone} onChange={(e) => updateField("phone", e.target.value)} />
              <TextInput label="Employee Code" name="employee_code" value={form.employee_code} onChange={(e) => updateField("employee_code", e.target.value)} />
              <TextInput label="Date of Birth" name="date_of_birth" type="date" value={form.date_of_birth} onChange={(e) => updateField("date_of_birth", e.target.value)} />
            </div>
          </section>

          {/* Section: Health & Verification */}
          <section className="bg-[#111114] border border-white/5 rounded-3xl p-8 space-y-8 shadow-sm">
            <div className="flex items-center gap-3">
              <ShieldCheck className="text-primary w-5 h-5" />
              <h2 className="text-white font-bold tracking-tight">Health & Background Verification</h2>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
              <div className="space-y-2">
                <label className="text-[10px] font-bold text-text-muted uppercase tracking-widest flex items-center gap-2">
                  <Droplets className="w-3 h-3 text-red-500" /> Blood Group
                </label>
                <SelectInput label="" name="blood_group" value={form.blood_group} options={["A+", "A-", "B+", "B-", "O+", "O-", "AB+", "AB-"].map(g => ({value: g, label: g}))} onChange={(e) => updateField("blood_group", e.target.value)} />
              </div>
              <div className="space-y-2">
                <label className="text-[10px] font-bold text-text-muted uppercase tracking-widest flex items-center gap-2">
                  <Stethoscope className="w-3 h-3 text-primary" /> Last Medical Checkup
                </label>
                <TextInput label="" name="medical_checkup_date" type="date" value={form.medical_checkup_date} onChange={(e) => updateField("medical_checkup_date", e.target.value)} />
              </div>
              <div className="space-y-2">
                <label className="text-[10px] font-bold text-text-muted uppercase tracking-widest flex items-center gap-2">
                  <Fingerprint className="w-3 h-3 text-primary" /> Police Verification Status
                </label>
                <SelectInput label="" name="police_verification_status" value={form.police_verification_status} options={[{value: "PENDING", label: "Pending"}, {value: "VERIFIED", label: "Verified"}, {value: "REJECTED", label: "Rejected"}]} onChange={(e) => updateField("police_verification_status", e.target.value)} />
              </div>
              <div className="space-y-2">
                <label className="text-[10px] font-bold text-text-muted uppercase tracking-widest flex items-center gap-2">
                  <Calendar className="w-3 h-3 text-primary" /> Verification Date
                </label>
                <TextInput label="" name="police_verification_date" type="date" value={form.police_verification_date} onChange={(e) => updateField("police_verification_date", e.target.value)} />
              </div>
            </div>
          </section>

          {/* Section: Credentials */}
          <section className="bg-[#111114] border border-white/5 rounded-3xl p-8 space-y-8 shadow-sm">
            <div className="flex items-center gap-3">
              <IdCard className="text-primary w-5 h-5" />
              <h2 className="text-white font-bold tracking-tight">Compliance & Licensing</h2>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
              <TextInput label="License Number" name="license_number" value={form.license_number} onChange={(e) => updateField("license_number", e.target.value)} />
              <SelectInput label="License Type" name="license_type" value={form.license_type} options={[{value: "LMV", label: "LMV"}, {value: "HMV", label: "HMV"}, {value: "Transport", label: "Transport"}]} onChange={(e) => updateField("license_type", e.target.value)} />
              <TextInput label="Issue Date" name="license_issue_date" type="date" value={form.license_issue_date} onChange={(e) => updateField("license_issue_date", e.target.value)} />
              <TextInput label="Expiry Date" name="license_expiry" type="date" value={form.license_expiry} onChange={(e) => updateField("license_expiry", e.target.value)} />
              <TextInput label="Issuing Authority" name="license_authority" value={form.license_authority} onChange={(e) => updateField("license_authority", e.target.value)} />
              <div className="grid grid-cols-2 gap-4">
                <TextInput label="Aadhaar ID" name="aadhaar_number" value={form.aadhaar_number} onChange={(e) => updateField("aadhaar_number", e.target.value)} />
                <TextInput label="PAN Signal" name="pan_number" value={form.pan_number} onChange={(e) => updateField("pan_number", e.target.value)} />
              </div>
            </div>
          </section>

          {/* Section: Residence */}
          <section className="bg-[#111114] border border-white/5 rounded-3xl p-8 space-y-8 shadow-sm">
            <div className="flex items-center gap-3">
              <MapPin className="text-primary w-5 h-5" />
              <h2 className="text-white font-bold tracking-tight">Geographic Vector (Address)</h2>
            </div>
            <div className="space-y-6">
              <TextInput label="Street Address" name="address_street" value={form.address_street} onChange={(e) => updateField("address_street", e.target.value)} />
              <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
                <TextInput label="City" name="address_city" value={form.address_city} onChange={(e) => updateField("address_city", e.target.value)} />
                <TextInput label="State" name="address_state" value={form.address_state} onChange={(e) => updateField("address_state", e.target.value)} />
                <TextInput label="Pincode" name="address_pincode" value={form.address_pincode} onChange={(e) => updateField("address_pincode", e.target.value)} />
                <TextInput label="Country" name="address_country" value={form.address_country} onChange={(e) => updateField("address_country", e.target.value)} />
              </div>
            </div>
          </section>

          {/* Section: Logistics */}
          <section className="bg-[#111114] border border-white/5 rounded-3xl p-8 space-y-8 shadow-sm">
            <div className="flex items-center gap-3">
              <Briefcase className="text-primary w-5 h-5" />
              <h2 className="text-white font-bold tracking-tight">Logistics & Employment</h2>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
              <SelectInput label="Employment Type" name="employment_type" value={form.employment_type} options={[{value: "FULL_TIME", label: "Full Time"}, {value: "CONTRACT", label: "Contract"}]} onChange={(e) => updateField("employment_type", e.target.value)} />
              <TextInput label="Experience (Years)" name="experience_years" type="number" value={form.experience_years} onChange={(e) => updateField("experience_years", e.target.value)} />
              <SelectInput label="Shift Pattern" name="shift_type" value={form.shift_type} options={[{value: "MORNING", label: "Morning"}, {value: "EVENING", label: "Evening"}, {value: "NIGHT", label: "Night"}]} onChange={(e) => updateField("shift_type", e.target.value)} />
              <div className="flex flex-col gap-2">
                <p className="text-[10px] text-text-muted font-bold uppercase tracking-widest">Joining Date (Locked)</p>
                <div className="p-4 bg-white/5 rounded-2xl border border-white/5 text-sm text-white flex items-center gap-2">
                  <Calendar className="w-4 h-4 text-primary" /> {driver?.joining_date ? new Date(driver.joining_date).toLocaleDateString() : "N/A"}
                </div>
              </div>
            </div>
            <div className="p-6 bg-white/5 rounded-2xl border border-white/5 flex items-center justify-between transition-all hover:bg-white/10">
              <div className="space-y-1">
                <p className="text-sm font-bold text-white">Behavior Monitoring</p>
                <p className="text-[10px] text-text-muted">Neural driving pattern analysis</p>
              </div>
              <ToggleInput label="" checked={form.behavior_monitoring} onChange={(v) => updateField("behavior_monitoring", v)} />
            </div>
          </section>

          {/* Section: Emergency & Notes */}
          <section className="bg-[#111114] border border-white/5 rounded-3xl p-8 space-y-8 shadow-sm">
            <div className="flex items-center gap-3">
              <HeartPulse className="text-primary w-5 h-5" />
              <h2 className="text-white font-bold tracking-tight">Emergency & Notes</h2>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <TextInput label="Emergency Name" name="emergency_name" value={form.emergency_name} onChange={(e) => updateField("emergency_name", e.target.value)} />
              <TextInput label="Emergency Phone" name="emergency_phone" value={form.emergency_phone} onChange={(e) => updateField("emergency_phone", e.target.value)} />
              <TextInput label="Kinship Relation" name="emergency_relation" value={form.emergency_relation} onChange={(e) => updateField("emergency_relation", e.target.value)} />
            </div>
            <div className="space-y-2">
              <label className="text-[10px] font-bold text-text-muted uppercase tracking-widest">Administrative Notes</label>
              <textarea 
                value={form.notes} 
                onChange={(e) => updateField("notes", e.target.value)}
                className="w-full bg-[#0a0a0b] border border-white/5 rounded-2xl p-4 text-sm text-white outline-none focus:border-primary/50 transition-all h-24 resize-none"
                placeholder="Record operative notes..."
              />
            </div>
            <div className="pt-6 border-t border-white/5 flex items-center justify-between text-text-muted">
               <div className="flex items-center gap-2">
                  <Activity className="w-4 h-4" />
                  <p className="text-xs font-bold uppercase tracking-widest">Profile Status:</p>
                  <span className="text-xs text-white uppercase">{driver?.status || "Unknown"}</span>
               </div>
               <p className="text-[10px] italic">Status changes require separate administrative authorization</p>
            </div>
          </section>
        </div>
      </div>
    </div>
  );
}
