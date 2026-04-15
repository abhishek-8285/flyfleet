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
  blood_group: "",
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

/**
 * AddDriver: Multi-step form for creating a new driver.
 * Follows the AddVehicle pattern.
 */
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
      if (!form.first_name.trim()) newErrors.first_name = "First name is required";
      if (!form.last_name.trim()) newErrors.last_name = "Last name is required";
      if (!form.email.trim()) newErrors.email = "Email is required";
      else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email)) newErrors.email = "Invalid email";
      if (!form.phone.trim()) newErrors.phone = "Phone is required";
      if (!form.date_of_birth) newErrors.date_of_birth = "Date of birth is required";
      if (!form.address_street.trim()) newErrors.address_street = "Street is required";
      if (!form.address_city.trim()) newErrors.address_city = "City is required";
      if (!form.address_state.trim()) newErrors.address_state = "State is required";
      if (!form.address_pincode.trim()) newErrors.address_pincode = "Pincode is required";
      if (!form.emergency_name.trim()) newErrors.emergency_name = "Emergency contact name is required";
      if (!form.emergency_phone.trim()) newErrors.emergency_phone = "Emergency contact phone is required";
    }

    if (step === 2) {
      if (!form.license_number.trim()) newErrors.license_number = "License number is required";
      if (!form.license_expiry) newErrors.license_expiry = "License expiry is required";
      if (!form.license_authority.trim()) newErrors.license_authority = "License authority is required";
      if (form.aadhaar_number && !/^\d{12}$/.test(form.aadhaar_number)) newErrors.aadhaar_number = "Aadhaar must be 12 digits";
      if (form.pan_number && !/^[A-Z]{5}[0-9]{4}[A-Z]{1}$/.test(form.pan_number)) newErrors.pan_number = "Invalid PAN format";
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleNext = () => {
    if (validateStep(currentStep)) {
      setCurrentStep((s) => Math.min(s + 1, 3));
    }
  };

  const handleBack = () => {
    setCurrentStep((s) => Math.max(s - 1, 1));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validateStep(3)) return;

    setSubmitting(true);
    try {
      await api.createDriver({
        first_name: form.first_name,
        last_name: form.last_name,
        email: form.email,
        phone: form.phone.startsWith("+") ? form.phone : `+91${form.phone}`,
        license_number: form.license_number,
        license_type: form.license_type,
        license_expiry: new Date(form.license_expiry).toISOString(),
        license_authority: form.license_authority,
        date_of_birth: new Date(form.date_of_birth).toISOString(),
        employment_type: form.employment_type,
        address: {
          street: form.address_street,
          city: form.address_city,
          state: form.address_state,
          pincode: form.address_pincode,
          country: form.address_country,
        },
        emergency_contact: {
          name: form.emergency_name,
          phone: form.emergency_phone.startsWith("+") ? form.emergency_phone : `+91${form.emergency_phone}`,
          relation: form.emergency_relation,
        },
        employee_code: form.employee_code || undefined,
        aadhaar_number: form.aadhaar_number || undefined,
        pan_number: form.pan_number || undefined,
        license_issue_date: form.license_issue_date ? new Date(form.license_issue_date).toISOString() : undefined,
        experience_years: parseInt(form.experience_years) || 0,
        metadata: {
          blood_group: form.blood_group,
          shift_type: form.shift_type,
          behavior_monitoring: form.behavior_monitoring,
          notes: form.notes,
          joining_date: form.joining_date,
        },
      });
      navigate("/admin/drivers");
    } catch (error: any) {
      const message = error?.response?.data?.error || "Failed to create driver";
      alert(message);
    } finally {
      setSubmitting(false);
    }
  };

  const renderInput = (
    label: string,
    field: keyof FormState,
    type: string = "text",
    placeholder?: string,
    required?: boolean,
  ) => (
    <div>
      <label className="block text-text-muted text-xs mb-1">
        {label} {required && <span className="text-red-400">*</span>}
      </label>
      <input
        type={type}
        value={(form[field] as string) || ""}
        onChange={(e) => updateField(field, e.target.value)}
        placeholder={placeholder}
        className={`w-full bg-surface-dark border text-white text-sm rounded-lg px-3 py-2 focus:ring-1 focus:ring-primary focus:border-primary ${
          errors[field] ? "border-red-400" : "border-border-color"
        }`}
      />
      {errors[field] && <p className="text-red-400 text-xs mt-1">{errors[field]}</p>}
    </div>
  );

  const renderSelect = (
    label: string,
    field: keyof FormState,
    options: { value: string; label: string }[],
    required?: boolean,
  ) => (
    <div>
      <label className="block text-text-muted text-xs mb-1">
        {label} {required && <span className="text-red-400">*</span>}
      </label>
      <select
        value={(form[field] as string) || ""}
        onChange={(e) => updateField(field, e.target.value)}
        className={`w-full bg-surface-dark border text-white text-sm rounded-lg px-3 py-2 focus:ring-1 focus:ring-primary focus:border-primary ${
          errors[field] ? "border-red-400" : "border-border-color"
        }`}
      >
        {options.map((opt) => (
          <option key={opt.value} value={opt.value}>{opt.label}</option>
        ))}
      </select>
      {errors[field] && <p className="text-red-400 text-xs mt-1">{errors[field]}</p>}
    </div>
  );

  const stepTitles = ["Personal Information", "License & Documents", "Assignment & Settings"];

  return (
    <div className="w-full max-w-4xl mx-auto">
      {/* Header */}
      <div className="mb-6">
        <button
          onClick={() => navigate("/admin/drivers")}
          className="text-text-muted hover:text-white text-sm mb-2 flex items-center gap-1"
        >
          <span className="material-symbols-outlined text-[18px]">arrow_back</span>
          Back to Drivers
        </button>
        <h1 className="text-white text-2xl font-bold">Add Driver</h1>
        <p className="text-text-muted text-sm">{stepTitles[currentStep - 1]}</p>
      </div>

      {/* Step Indicators */}
      <div className="flex items-center gap-2 mb-6">
        {[1, 2, 3].map((step) => (
          <React.Fragment key={step}>
            <div className={`flex items-center gap-2 ${step <= currentStep ? "text-primary" : "text-text-muted"}`}>
              <div className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-semibold ${
                step < currentStep ? "bg-primary text-white" :
                step === currentStep ? "bg-primary/20 text-primary border border-primary" :
                "bg-surface-dark text-text-muted border border-border-color"
              }`}>
                {step < currentStep ? "✓" : step}
              </div>
              <span className="text-sm hidden md:inline">{stepTitles[step - 1]}</span>
            </div>
            {step < 3 && <div className={`flex-1 h-px ${step < currentStep ? "bg-primary" : "bg-border-color"}`} />}
          </React.Fragment>
        ))}
      </div>

      {/* Form */}
      <form onSubmit={handleSubmit} className="bg-surface-dark border border-border-color rounded-lg p-6">
        {currentStep === 1 && (
          <div className="space-y-4">
            <h3 className="text-white font-semibold text-lg mb-4">Personal Details</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {renderInput("First Name", "first_name", "text", "John", true)}
              {renderInput("Last Name", "last_name", "text", "Doe", true)}
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {renderInput("Email", "email", "email", "john@example.com", true)}
              {renderInput("Phone", "phone", "tel", "+91 98765 43210", true)}
            </div>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              {renderInput("Employee Code", "employee_code", "text", "EMP-001")}
              {renderInput("Date of Birth", "date_of_birth", "date", "", true)}
              {renderInput("Joining Date", "joining_date", "date")}
            </div>

            <h3 className="text-white font-semibold text-lg mb-4 mt-6">Address</h3>
            <div className="grid grid-cols-1 gap-4">
              {renderInput("Street", "address_street", "text", "123 Main St", true)}
            </div>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              {renderInput("City", "address_city", "text", "Mumbai", true)}
              {renderInput("State", "address_state", "text", "Maharashtra", true)}
              {renderInput("Pincode", "address_pincode", "text", "400001", true)}
            </div>

            <h3 className="text-white font-semibold text-lg mb-4 mt-6">Emergency Contact</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {renderInput("Name", "emergency_name", "text", "", true)}
              {renderInput("Phone", "emergency_phone", "tel", "", true)}
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {renderInput("Relationship", "emergency_relation", "text", "Spouse")}
              {renderSelect("Blood Group", "blood_group", BLOOD_GROUPS.map((bg) => ({ value: bg, label: bg })))}
            </div>
          </div>
        )}

        {currentStep === 2 && (
          <div className="space-y-4">
            <h3 className="text-white font-semibold text-lg mb-4">License Information</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {renderInput("License Number", "license_number", "text", "MH0120230001234", true)}
              {renderSelect("License Type", "license_type", LICENSE_TYPES, true)}
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {renderInput("License Issue Date", "license_issue_date", "date")}
              {renderInput("License Expiry", "license_expiry", "date", "", true)}
            </div>
            {renderInput("License Authority", "license_authority", "text", "RTO Mumbai", true)}

            <h3 className="text-white font-semibold text-lg mb-4 mt-6">Identity Documents</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {renderInput("Aadhaar Number", "aadhaar_number", "text", "1234 5678 9012")}
              {renderInput("PAN Number", "pan_number", "text", "ABCDE1234F")}
            </div>
          </div>
        )}

        {currentStep === 3 && (
          <div className="space-y-4">
            <h3 className="text-white font-semibold text-lg mb-4">Employment Details</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {renderSelect("Employment Type", "employment_type", EMPLOYMENT_TYPES, true)}
              {renderInput("Experience (Years)", "experience_years", "number", "0")}
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {renderSelect("Shift Type", "shift_type", SHIFT_TYPES)}
            </div>

            <h3 className="text-white font-semibold text-lg mb-4 mt-6">Settings</h3>
            <div className="flex items-center gap-3">
              <input
                type="checkbox"
                id="behavior_monitoring"
                checked={form.behavior_monitoring}
                onChange={(e) => updateField("behavior_monitoring", e.target.checked)}
                className="rounded border-border-color"
              />
              <label htmlFor="behavior_monitoring" className="text-white text-sm">
                Enable behavior monitoring
              </label>
            </div>

            <div>
              <label className="block text-text-muted text-xs mb-1">Notes</label>
              <textarea
                value={form.notes}
                onChange={(e) => updateField("notes", e.target.value)}
                placeholder="Any additional notes about this driver..."
                rows={3}
                className="w-full bg-surface-dark border border-border-color text-white text-sm rounded-lg px-3 py-2 focus:ring-1 focus:ring-primary focus:border-primary resize-none"
              />
            </div>
          </div>
        )}

        {/* Buttons */}
        <div className="flex items-center justify-between mt-6 pt-6 border-t border-border-color">
          <button
            type="button"
            onClick={handleBack}
            disabled={currentStep === 1}
            className="px-4 py-2 text-sm font-medium text-text-muted hover:text-white disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
          >
            Back
          </button>
          {currentStep < 3 ? (
            <button
              type="button"
              onClick={handleNext}
              className="px-6 py-2 bg-primary hover:bg-primary-dark text-white text-sm font-semibold rounded-lg transition-colors"
            >
              Next
            </button>
          ) : (
            <button
              type="submit"
              disabled={submitting}
              className="px-6 py-2 bg-green-600 hover:bg-green-700 disabled:opacity-50 text-white text-sm font-semibold rounded-lg transition-colors flex items-center gap-2"
            >
              {submitting && <span className="material-symbols-outlined animate-spin text-[16px]">progress_activity</span>}
              {submitting ? "Creating..." : "Create Driver"}
            </button>
          )}
        </div>
      </form>
    </div>
  );
};
