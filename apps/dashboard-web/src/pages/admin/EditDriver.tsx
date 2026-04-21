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
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [errors, setErrors] = useState<Partial<Record<keyof FormState, string>>>({});

  useEffect(() => {
    const fetchDriver = async () => {
      if (!id) return;
      try {
        const res = await api.getDriver(id);
        const data = res.data || res;
        setForm({
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
        });
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
    if (!id || !form) return;
    setSaving(true);
    try {
      await api.updateDriver(id, {
        ...form,
        experience_years: parseInt(form.experience_years) || 0,
        address: { 
          street: form.address_street, 
          city: form.address_city, 
          state: form.address_state, 
          pincode: form.address_pincode, 
          country: form.address_country 
        },
        emergency_contact: { 
          name: form.emergency_name, 
          phone: form.emergency_phone, 
          relation: form.emergency_relation 
        },
        metadata: { 
          blood_group: form.blood_group, 
          shift_type: form.shift_type, 
          behavior_monitoring: form.behavior_monitoring, 
          notes: form.notes,
          joining_date: form.joining_date
        }
      });
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
