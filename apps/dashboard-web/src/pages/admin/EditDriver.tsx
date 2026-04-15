import React, { useEffect, useState, useCallback } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { api } from "../../services/api";
import { Driver } from "../../modules/drivers/types";

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

const STATUS_OPTIONS = [
  { value: "active", label: "Active", color: "text-green-400" },
  { value: "inactive", label: "Inactive", color: "text-gray-400" },
  { value: "suspended", label: "Suspended", color: "text-red-400" },
  { value: "resigned", label: "Resigned", color: "text-yellow-400" },
];

interface FormState {
  first_name: string;
  last_name: string;
  email: string;
  phone: string;
  employee_code: string;
  license_number: string;
  license_type: string;
  license_expiry: string;
  license_issue_date: string;
  license_authority: string;
  employment_type: string;
  experience_years: string;
  status: string;
  address_street: string;
  address_city: string;
  address_state: string;
  address_pincode: string;
  address_country: string;
  emergency_name: string;
  emergency_phone: string;
  emergency_relation: string;
  notes: string;
  shift_type: string;
  behavior_monitoring: boolean;
}

/**
 * EditDriver: Single-page form for editing an existing driver.
 * Follows the EditVehicle pattern.
 */
export const EditDriver: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [driver, setDriver] = useState<Driver | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [submitError, setSubmitError] = useState("");
  const [confirmDelete, setConfirmDelete] = useState(false);
  const [form, setForm] = useState<FormState>({
    first_name: "",
    last_name: "",
    email: "",
    phone: "",
    employee_code: "",
    license_number: "",
    license_type: "LMV",
    license_expiry: "",
    license_issue_date: "",
    license_authority: "",
    employment_type: "FULL_TIME",
    experience_years: "0",
    status: "active",
    address_street: "",
    address_city: "",
    address_state: "",
    address_pincode: "",
    address_country: "India",
    emergency_name: "",
    emergency_phone: "",
    emergency_relation: "",
    notes: "",
    shift_type: "MORNING",
    behavior_monitoring: true,
  });

  const loadDriver = useCallback(async () => {
    if (!id) return;
    setLoading(true);
    try {
      const data = await api.getDriver(id);
      setDriver(data);
      const metadata = data.metadata || {};
      setForm({
        first_name: data.first_name || "",
        last_name: data.last_name || "",
        email: data.email || "",
        phone: data.phone || "",
        employee_code: data.employee_code || "",
        license_number: data.license_number || "",
        license_type: data.license_type || "LMV",
        license_expiry: data.license_expiry ? data.license_expiry.split("T")[0] : "",
        license_issue_date: data.license_issue_date ? data.license_issue_date.split("T")[0] : "",
        license_authority: data.license_authority || "",
        employment_type: data.employment_type || "FULL_TIME",
        experience_years: String(data.experience_years || 0),
        status: data.status || "active",
        address_street: data.address?.street || "",
        address_city: data.address?.city || "",
        address_state: data.address?.state || "",
        address_pincode: data.address?.pincode || "",
        address_country: data.address?.country || "India",
        emergency_name: data.emergency_contact?.name || "",
        emergency_phone: data.emergency_contact?.phone || "",
        emergency_relation: data.emergency_contact?.relation || "",
        notes: metadata.notes || "",
        shift_type: metadata.shift_type || "MORNING",
        behavior_monitoring: metadata.behavior_monitoring ?? true,
      });
    } catch (error) {
      console.error("Failed to load driver:", error);
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    loadDriver();
  }, [loadDriver]);

  const updateField = (field: keyof FormState, value: string | boolean) => {
    setForm((prev) => ({ ...prev, [field]: value }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitError("");
    setSaving(true);
    try {
      await api.updateDriver(id!, {
        first_name: form.first_name,
        last_name: form.last_name,
        email: form.email,
        phone: form.phone,
        employee_code: form.employee_code,
        license_number: form.license_number,
        license_type: form.license_type,
        license_expiry: form.license_expiry ? new Date(form.license_expiry).toISOString() : undefined,
        license_authority: form.license_authority,
        employment_type: form.employment_type,
        experience_years: parseInt(form.experience_years) || 0,
        status: form.status,
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
          notes: form.notes,
          shift_type: form.shift_type,
          behavior_monitoring: form.behavior_monitoring,
        },
      });
      navigate(`/admin/drivers/${id}`);
    } catch (error: any) {
      setSubmitError(error?.response?.data?.error || "Failed to update driver");
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!confirm("Are you sure? This will permanently delete this driver.")) return;
    try {
      await api.deleteDriver(id!);
      navigate("/admin/drivers");
    } catch (error) {
      alert("Failed to delete driver. They may have an active assignment or trip.");
    }
  };

  const renderInput = (label: string, field: keyof FormState, type: string = "text", required?: boolean) => (
    <div>
      <label className="block text-text-muted text-xs mb-1">
        {label} {required && <span className="text-red-400">*</span>}
      </label>
      <input
        type={type}
        value={form[field] as string}
        onChange={(e) => updateField(field, e.target.value)}
        className="w-full bg-surface-dark border border-border-color text-white text-sm rounded-lg px-3 py-2 focus:ring-1 focus:ring-primary focus:border-primary"
      />
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
        value={form[field] as string}
        onChange={(e) => updateField(field, e.target.value)}
        className="w-full bg-surface-dark border border-border-color text-white text-sm rounded-lg px-3 py-2 focus:ring-1 focus:ring-primary focus:border-primary"
      >
        {options.map((opt) => (
          <option key={opt.value} value={opt.value}>{opt.label}</option>
        ))}
      </select>
    </div>
  );

  if (loading) {
    return (
      <div className="w-full flex items-center justify-center py-20">
        <span className="material-symbols-outlined animate-spin text-primary text-4xl">progress_activity</span>
      </div>
    );
  }

  if (!driver) {
    return (
      <div className="w-full text-center py-20 text-text-muted">
        <p>Driver not found.</p>
        <button onClick={() => navigate("/admin/drivers")} className="text-primary hover:underline mt-2">
          Back to Drivers
        </button>
      </div>
    );
  }

  const statusOption = STATUS_OPTIONS.find((s) => s.value === driver.status);

  return (
    <div className="w-full max-w-4xl mx-auto">
      {/* Header */}
      <div className="mb-6">
        <button
          onClick={() => navigate(`/admin/drivers/${id}`)}
          className="text-text-muted hover:text-white text-sm mb-2 flex items-center gap-1"
        >
          <span className="material-symbols-outlined text-[18px]">arrow_back</span>
          Back to Driver
        </button>
        <div className="flex items-center gap-3">
          <h1 className="text-white text-2xl font-bold">{driver.first_name} {driver.last_name}</h1>
          {statusOption && (
            <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${
              driver.status === "active" ? "bg-green-500/20 text-green-400" :
              driver.status === "inactive" ? "bg-gray-500/20 text-gray-400" :
              driver.status === "suspended" ? "bg-red-500/20 text-red-400" :
              "bg-yellow-500/20 text-yellow-400"
            }`}>
              {statusOption.label}
            </span>
          )}
        </div>
        <p className="text-text-muted text-sm">{driver.license_number} · {driver.license_type}</p>
      </div>

      {/* Form */}
      <form onSubmit={handleSubmit} className="bg-surface-dark border border-border-color rounded-lg p-6">
        {submitError && (
          <div className="bg-red-500/10 border border-red-500/30 rounded-lg p-3 mb-4 text-red-400 text-sm">
            {submitError}
          </div>
        )}

        <div className="space-y-6">
          <div>
            <h3 className="text-white font-semibold text-lg mb-4">Personal Information</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {renderInput("First Name", "first_name", "text", true)}
              {renderInput("Last Name", "last_name", "text", true)}
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-4">
              {renderInput("Email", "email", "email", true)}
              {renderInput("Phone", "phone", "tel", true)}
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-4">
              {renderInput("Employee Code", "employee_code")}
              {renderInput("Experience (Years)", "experience_years", "number")}
            </div>
          </div>

          <div>
            <h3 className="text-white font-semibold text-lg mb-4">License</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {renderInput("License Number", "license_number", "text", true)}
              {renderSelect("License Type", "license_type", LICENSE_TYPES, true)}
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-4">
              {renderInput("License Issue Date", "license_issue_date", "date")}
              {renderInput("License Expiry", "license_expiry", "date", true)}
            </div>
            {renderInput("License Authority", "license_authority", "text", true)}
          </div>

          <div>
            <h3 className="text-white font-semibold text-lg mb-4">Address</h3>
            {renderInput("Street", "address_street")}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mt-4">
              {renderInput("City", "address_city")}
              {renderInput("State", "address_state")}
              {renderInput("Pincode", "address_pincode")}
            </div>
          </div>

          <div>
            <h3 className="text-white font-semibold text-lg mb-4">Emergency Contact</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {renderInput("Name", "emergency_name")}
              {renderInput("Phone", "emergency_phone", "tel")}
            </div>
            {renderInput("Relationship", "emergency_relation")}
          </div>

          <div>
            <h3 className="text-white font-semibold text-lg mb-4">Settings</h3>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              {renderSelect("Status", "status", STATUS_OPTIONS)}
              {renderSelect("Employment Type", "employment_type", EMPLOYMENT_TYPES)}
              {renderSelect("Shift Type", "shift_type", SHIFT_TYPES)}
            </div>
            <div className="flex items-center gap-3 mt-4">
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
            <div className="mt-4">
              <label className="block text-text-muted text-xs mb-1">Notes</label>
              <textarea
                value={form.notes}
                onChange={(e) => updateField("notes", e.target.value)}
                rows={3}
                className="w-full bg-surface-dark border border-border-color text-white text-sm rounded-lg px-3 py-2 focus:ring-1 focus:ring-primary focus:border-primary resize-none"
              />
            </div>
          </div>
        </div>

        {/* Danger Zone */}
        <div className="mt-8 pt-6 border-t border-border-color">
          {!confirmDelete ? (
            <button
              type="button"
              onClick={() => setConfirmDelete(true)}
              className="px-4 py-2 text-sm text-red-400 hover:bg-red-500/10 rounded-lg transition-colors"
            >
              Delete Driver
            </button>
          ) : (
            <div className="bg-red-500/10 border border-red-500/30 rounded-lg p-4 flex items-center gap-3">
              <span className="text-red-400 text-sm">Permanently delete this driver?</span>
              <button
                type="button"
                onClick={handleDelete}
                className="px-3 py-1.5 bg-red-600 hover:bg-red-700 text-white text-sm rounded transition-colors"
              >
                Confirm Delete
              </button>
              <button
                type="button"
                onClick={() => setConfirmDelete(false)}
                className="px-3 py-1.5 text-text-muted hover:text-white text-sm transition-colors"
              >
                Cancel
              </button>
            </div>
          )}
        </div>

        {/* Buttons */}
        <div className="flex items-center justify-end gap-3 mt-6 pt-6 border-t border-border-color">
          <button
            type="button"
            onClick={() => navigate(`/admin/drivers/${id}`)}
            className="px-4 py-2 text-sm font-medium text-text-muted hover:text-white transition-colors"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={saving}
            className="px-6 py-2 bg-primary hover:bg-primary-dark disabled:opacity-50 text-white text-sm font-semibold rounded-lg transition-colors flex items-center gap-2"
          >
            {saving && <span className="material-symbols-outlined animate-spin text-[16px]">progress_activity</span>}
            {saving ? "Saving..." : "Save Changes"}
          </button>
        </div>
      </form>
    </div>
  );
};
