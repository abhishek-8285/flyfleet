import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { 
  ChevronLeft, 
  Save, 
  User, 
  IdCard, 
  Briefcase, 
  ShieldCheck,
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

export const AddDriver: React.FC = () => {
  const navigate = useNavigate();
  const [form, setForm] = useState<FormState>(initialForm);
  const [loading, setLoading] = useState(false);
  const [errors, setErrors] = useState<Partial<Record<keyof FormState, string>>>({});

  const updateField = (field: keyof FormState, value: string | boolean) => {
    setForm((prev) => ({ ...prev, [field]: value }));
    if (errors[field]) setErrors((prev) => ({ ...prev, [field]: undefined }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      await api.createDriver({
        ...form,
        phone: form.phone.startsWith("+") ? form.phone : `+91${form.phone}`,
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
          phone: form.emergency_phone.startsWith("+") ? form.emergency_phone : `+91${form.emergency_phone}`, 
          relation: form.emergency_relation 
        },
        metadata: { 
          blood_group: form.blood_group, 
          shift_type: form.shift_type, 
          behavior_monitoring: form.behavior_monitoring, 
          notes: form.notes,
        }
      });
      navigate("/admin/drivers");
    } catch (error) {
      console.error("Submission failed:", error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-4xl mx-auto space-y-8 animate-in fade-in duration-500">
      <div className="flex items-center justify-between">
        <button
          onClick={() => navigate("/admin/drivers")}
          className="flex items-center gap-2 text-text-muted hover:text-white transition-colors group"
        >
          <ChevronLeft className="w-5 h-5 group-hover:-translate-x-1 transition-transform" />
          Back to Personnel Registry
        </button>
        <h1 className="text-2xl font-bold text-white">Initialize Personnel Node</h1>
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

        <div className="flex items-center justify-end gap-4 pt-4">
            <button
              type="button"
              onClick={() => navigate("/admin/drivers")}
              className="px-8 py-3 text-text-muted hover:text-white transition-all font-semibold"
            >
              Abort
            </button>
            <button
              type="submit"
              disabled={loading}
              className="flex items-center gap-2 px-10 py-3 bg-primary text-white rounded-xl hover:bg-primary/90 transition-all font-bold shadow-xl shadow-primary/20 disabled:opacity-50"
            >
              {loading ? (
                <div className="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
              ) : (
                <Save className="w-5 h-5" />
              )}
              Initialize Profile
            </button>
        </div>
      </form>
    </div>
  );
};

export default AddDriver;
