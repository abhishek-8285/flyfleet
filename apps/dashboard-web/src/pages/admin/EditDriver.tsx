import React, { useState, useEffect } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { 
  ChevronLeft, 
  Save, 
  User, 
  IdCard, 
  Phone,
  Mail,
  MapPin,
  Info 
} from "lucide-react";
import { api } from "../../services/api";
import { TextInput, SelectInput, ToggleInput } from "../../components/common/FormFields";

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

export const EditDriver: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [form, setForm] = useState<FormState | null>(null);
  const [initialForm, setInitialForm] = useState<FormState | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [errors, setErrors] = useState<Partial<Record<keyof FormState, string>>>({});

  useEffect(() => {
    const fetchDriver = async () => {
      if (!id) return;
      try {
        const res = await api.getDriver(id);
        const data = res.data || res;
        const formData = {
          first_name: data.first_name || "",
          last_name: data.last_name || "",
          email: data.email || "",
          phone: data.phone || "",
          employee_code: data.employee_code || "",
          aadhaar_number: data.aadhaar_number || "",
          pan_number: data.pan_number || "",
          date_of_birth: data.date_of_birth ? data.date_of_birth.split("T")[0] : "",
          license_number: data.license_number || "",
          license_type: data.license_type || "LMV",
          license_expiry: data.license_expiry ? data.license_expiry.split("T")[0] : "",
          license_issue_date: data.license_issue_date ? data.license_issue_date.split("T")[0] : "",
          license_authority: data.license_authority || "",
          employment_type: data.employment_type || "FULL_TIME",
          experience_years: (data.experience_years || 0).toString(),
          joining_date: data.metadata?.joining_date ? data.metadata.joining_date.split("T")[0] : "",
          address_street: data.address?.street || "",
          address_city: data.address?.city || "",
          address_state: data.address?.state || "",
          address_pincode: data.address?.pincode || "",
          address_country: data.address?.country || "India",
          emergency_name: data.emergency_contact?.name || "",
          emergency_phone: data.emergency_contact?.phone || "",
          emergency_relation: data.emergency_contact?.relation || "",
          blood_group: data.metadata?.blood_group || "O+",
          shift_type: data.metadata?.shift_type || "MORNING",
          behavior_monitoring: data.metadata?.behavior_monitoring ?? true,
          notes: data.metadata?.notes || "",
        };
        setForm(formData);
        setInitialForm(formData);
      } catch (err) {
        console.error("Failed to fetch driver", err);
      } finally {
        setLoading(false);
      }
    };
    fetchDriver();
  }, [id]);

  const updateField = (field: keyof FormState, value: string | boolean) => {
    if (!form) return;
    setForm((prev) => prev ? ({ ...prev, [field]: value }) : null);
    if (errors[field]) setErrors((prev) => ({ ...prev, [field]: undefined }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!id || !form || !initialForm) return;
    setSaving(true);

    const payload: any = {};

    if (form.first_name !== initialForm.first_name) payload.first_name = form.first_name;
    if (form.last_name !== initialForm.last_name) payload.last_name = form.last_name;
    if (form.email !== initialForm.email) payload.email = form.email;
    if (form.phone !== initialForm.phone) payload.phone = form.phone;
    if (form.employee_code !== initialForm.employee_code) payload.employee_code = form.employee_code;
    if (form.aadhaar_number !== initialForm.aadhaar_number) payload.aadhaar_number = form.aadhaar_number;
    if (form.pan_number !== initialForm.pan_number) payload.pan_number = form.pan_number;
    if (form.license_number !== initialForm.license_number) payload.license_number = form.license_number;
    if (form.license_type !== initialForm.license_type) payload.license_type = form.license_type;
    if (form.license_authority !== initialForm.license_authority) payload.license_authority = form.license_authority;
    if (form.employment_type !== initialForm.employment_type) payload.employment_type = form.employment_type;

    if (form.experience_years !== initialForm.experience_years) {
      payload.experience_years = parseInt(form.experience_years) || 0;
    }

    if (form.date_of_birth !== initialForm.date_of_birth) {
      payload.date_of_birth = form.date_of_birth ? `${form.date_of_birth}T00:00:00Z` : null;
    }
    if (form.license_expiry !== initialForm.license_expiry) {
      payload.license_expiry = form.license_expiry ? `${form.license_expiry}T00:00:00Z` : null;
    }
    if (form.license_issue_date !== initialForm.license_issue_date) {
      payload.license_issue_date = form.license_issue_date ? `${form.license_issue_date}T00:00:00Z` : null;
    }

    if (
      form.address_street !== initialForm.address_street ||
      form.address_city !== initialForm.address_city ||
      form.address_state !== initialForm.address_state ||
      form.address_pincode !== initialForm.address_pincode ||
      form.address_country !== initialForm.address_country
    ) {
      payload.address = {
        street: form.address_street,
        city: form.address_city,
        state: form.address_state,
        pincode: form.address_pincode,
        country: form.address_country,
      };
    }

    if (
      form.emergency_name !== initialForm.emergency_name ||
      form.emergency_phone !== initialForm.emergency_phone ||
      form.emergency_relation !== initialForm.emergency_relation
    ) {
      payload.emergency_contact = {
        name: form.emergency_name,
        phone: form.emergency_phone,
        relation: form.emergency_relation,
      };
    }

    if (
      form.blood_group !== initialForm.blood_group ||
      form.shift_type !== initialForm.shift_type ||
      form.behavior_monitoring !== initialForm.behavior_monitoring ||
      form.notes !== initialForm.notes ||
      form.joining_date !== initialForm.joining_date
    ) {
      payload.metadata = {
        blood_group: form.blood_group,
        shift_type: form.shift_type,
        behavior_monitoring: form.behavior_monitoring,
        notes: form.notes,
        joining_date: form.joining_date,
      };
    }

    if (Object.keys(payload).length === 0) {
      setSaving(false);
      navigate(`/admin/drivers/${id}`);
      return;
    }

    try {
      await api.updateDriver(id, payload);
      navigate(`/admin/drivers/${id}`);
    } catch (error) {
      console.error("Update failed:", error);
    } finally {
      setSaving(false);
    }
  };

  if (loading) return <div className="flex h-screen items-center justify-center text-white">Loading...</div>;
  if (!form) return null;

  return (
    <div className="max-w-4xl mx-auto space-y-8 animate-in fade-in duration-500">
      <div className="flex items-center justify-between">
        <button
          onClick={() => navigate(-1)}
          className="flex items-center gap-2 text-text-muted hover:text-white transition-colors group"
        >
          <ChevronLeft className="w-5 h-5 group-hover:-translate-x-1 transition-transform" />
          Back to Profile
        </button>
        <h1 className="text-2xl font-bold text-white">Modify Personnel Profile</h1>
        <div className="w-20" />
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        <div className="bg-surface-dark border border-border-color p-8 rounded-3xl space-y-8 shadow-2xl">
          <div className="flex items-center gap-3">
             <div className="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center text-primary">
                <User className="w-5 h-5" />
             </div>
             <h2 className="text-lg font-semibold text-white">Identity Signature</h2>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <TextInput label="First Name" name="first_name" value={form.first_name} onChange={(e) => updateField("first_name", e.target.value)} required />
            <TextInput label="Last Name" name="last_name" value={form.last_name} onChange={(e) => updateField("last_name", e.target.value)} required />
            <TextInput label="Email" name="email" type="email" value={form.email} onChange={(e) => updateField("email", e.target.value)} />
            <TextInput label="Phone" name="phone" type="tel" value={form.phone} onChange={(e) => updateField("phone", e.target.value)} required />
          </div>
        </div>

        <div className="bg-surface-dark border border-border-color p-8 rounded-3xl space-y-8 shadow-2xl">
          <div className="flex items-center gap-3">
             <div className="w-10 h-10 rounded-xl bg-indigo-500/10 flex items-center justify-center text-indigo-400">
                <IdCard className="w-5 h-5" />
             </div>
             <h2 className="text-lg font-semibold text-white">Compliance & Credentials</h2>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <TextInput label="License Number" name="license_number" value={form.license_number} onChange={(e) => updateField("license_number", e.target.value)} required />
            <SelectInput 
              label="License Type" 
              name="license_type"
              value={form.license_type} 
              options={[
                { value: "LMV", label: "LMV" },
                { value: "HMV", label: "HMV" }
              ]} 
              onChange={(e) => updateField("license_type", e.target.value)} 
            />
            <TextInput label="Aadhaar ID" name="aadhaar_number" value={form.aadhaar_number} onChange={(e) => updateField("aadhaar_number", e.target.value)} />
            <TextInput label="PAN Signal" name="pan_number" value={form.pan_number} onChange={(e) => updateField("pan_number", e.target.value)} />
          </div>
        </div>

        <div className="bg-surface-dark border border-border-color p-8 rounded-3xl space-y-8 shadow-2xl">
          <div className="flex items-center gap-3">
             <div className="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center text-primary">
                <Info className="w-5 h-5" />
             </div>
             <h2 className="text-lg font-semibold text-white">Additional Metadata</h2>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <SelectInput 
              label="Employment Type" 
              name="employment_type"
              value={form.employment_type} 
              options={[
                { value: "FULL_TIME", label: "Full Time" },
                { value: "CONTRACT", label: "Contract" }
              ]} 
              onChange={(e) => updateField("employment_type", e.target.value)} 
            />
            <TextInput label="Experience (Years)" name="experience_years" type="number" value={form.experience_years} onChange={(e) => updateField("experience_years", e.target.value)} />
            <SelectInput 
              label="Shift Pattern" 
              name="shift_type"
              value={form.shift_type} 
              options={[
                { value: "MORNING", label: "Morning" },
                { value: "EVENING", label: "Evening" },
                { value: "NIGHT", label: "Night" }
              ]} 
              onChange={(e) => updateField("shift_type", e.target.value)} 
            />
            <ToggleInput label="Behavior Monitoring" checked={form.behavior_monitoring} onChange={(v) => updateField("behavior_monitoring", v)} />
          </div>
        </div>

        <div className="flex items-center justify-end gap-4 pt-4">
            <button
              type="button"
              onClick={() => navigate(-1)}
              className="px-8 py-3 text-text-muted hover:text-white transition-all font-semibold"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={saving}
              className="flex items-center gap-2 px-10 py-3 bg-primary text-white rounded-xl hover:bg-primary/90 transition-all font-bold shadow-xl shadow-primary/20 disabled:opacity-50"
            >
              {saving ? (
                <div className="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
              ) : (
                <Save className="w-5 h-5" />
              )}
              Save Changes
            </button>
        </div>
      </form>
    </div>
  );
};
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
