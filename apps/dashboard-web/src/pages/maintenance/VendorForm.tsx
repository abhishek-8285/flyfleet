import React, { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useUIStrings } from "../../hooks/useUIStrings";

const VendorForm: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const isEditing = !!id;
  const { uiStrings: rawUiStrings } = useUIStrings("maintenance");
  const { uiStrings: commonStrings } = useUIStrings("common");
  const uiStrings = rawUiStrings || {};

  const [formData, setFormData] = useState({
    name: "",
    contact_person: "",
    email: "",
    phone: "",
    address: "",
    category: "PARTS",
    is_active: true,
  });
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (isEditing) {
      // Fetch vendor details
      fetch(`/api/v1/admin/vendors/${id}`)
        .then((res) => res.json())
        .then((data) => {
          if (data) {
            setFormData({
              name: data.name || "",
              contact_person: data.contact_person || "",
              email: data.email || "",
              phone: data.phone || "",
              address: data.address || "",
              category: data.category || "PARTS",
              is_active: data.is_active ?? true,
            });
          }
        })
        .catch((err) => console.error("Failed to fetch vendor:", err));
    }
  }, [id, isEditing]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);

    try {
      const url = isEditing
        ? `/api/v1/admin/vendors/${id}`
        : "/api/v1/admin/vendors";
      const method = isEditing ? "PUT" : "POST";

      const response = await fetch(url, {
        method,
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(formData),
      });

      if (response.ok) {
        navigate("/maintenance/vendors");
      } else {
        alert(uiStrings.error_save_vendor || "Failed to save vendor");
      }
    } catch (error) {
      console.error("Failed to save vendor:", error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex-1 flex flex-col min-w-0 bg-background-dark overflow-hidden h-full">
      <div className="w-full bg-surface-darker border-b border-border-color px-6 py-4 flex-shrink-0">
        <div className="flex items-center gap-4">
          <button
            onClick={() => navigate("/maintenance/vendors")}
            className="p-2 hover:bg-surface-dark rounded-lg transition-colors"
          >
            <span className="material-symbols-outlined text-text-muted">
              arrow_back
            </span>
          </button>
          <h1 className="text-2xl font-bold text-white">
            {isEditing
              ? uiStrings.title_edit_vendor || "Edit Vendor"
              : uiStrings.title_new_vendor || "New Vendor"}
          </h1>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto p-6">
        <div className="max-w-2xl mx-auto bg-surface-dark border border-border-color rounded-xl p-6">
          <form onSubmit={handleSubmit} className="space-y-6">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <label className="block text-sm font-medium text-text-muted mb-2">
                  {uiStrings.label_vendor_name || "Vendor Name"}
                </label>
                <input
                  type="text"
                  required
                  value={formData.name}
                  onChange={(e) =>
                    setFormData({ ...formData, name: e.target.value })
                  }
                  className="w-full bg-background-dark border border-border-color rounded-lg px-4 py-2 text-white focus:ring-1 focus:ring-primary focus:border-primary"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-text-muted mb-2">
                  {uiStrings.label_category || "Category"}
                </label>
                <select
                  value={formData.category}
                  onChange={(e) =>
                    setFormData({ ...formData, category: e.target.value })
                  }
                  className="w-full bg-background-dark border border-border-color rounded-lg px-4 py-2 text-white focus:ring-1 focus:ring-primary focus:border-primary"
                >
                  <option value="PARTS">Parts Supplier</option>
                  <option value="SERVICE">Service Provider</option>
                  <option value="DEALERSHIP">Dealership</option>
                  <option value="INSURANCE">Insurance</option>
                  <option value="OTHER">Other</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-text-muted mb-2">
                  {uiStrings.label_contact_person || "Contact Person"}
                </label>
                <input
                  type="text"
                  required
                  value={formData.contact_person}
                  onChange={(e) =>
                    setFormData({ ...formData, contact_person: e.target.value })
                  }
                  className="w-full bg-background-dark border border-border-color rounded-lg px-4 py-2 text-white focus:ring-1 focus:ring-primary focus:border-primary"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-text-muted mb-2">
                  {uiStrings.label_email || "Email"}
                </label>
                <input
                  type="email"
                  required
                  value={formData.email}
                  onChange={(e) =>
                    setFormData({ ...formData, email: e.target.value })
                  }
                  className="w-full bg-background-dark border border-border-color rounded-lg px-4 py-2 text-white focus:ring-1 focus:ring-primary focus:border-primary"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-text-muted mb-2">
                  {uiStrings.label_phone || "Phone"}
                </label>
                <input
                  type="tel"
                  required
                  value={formData.phone}
                  onChange={(e) =>
                    setFormData({ ...formData, phone: e.target.value })
                  }
                  className="w-full bg-background-dark border border-border-color rounded-lg px-4 py-2 text-white focus:ring-1 focus:ring-primary focus:border-primary"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-text-muted mb-2">
                  {commonStrings?.label_status ||
                    uiStrings?.th_status ||
                    "Status"}
                </label>
                <div className="flex items-center gap-4 mt-2">
                  <label className="flex items-center gap-2 cursor-pointer">
                    <input
                      type="radio"
                      checked={formData.is_active}
                      onChange={() =>
                        setFormData({ ...formData, is_active: true })
                      }
                      className="text-primary focus:ring-primary"
                    />
                    <span className="text-white">
                      {uiStrings.label_active || "Active"}
                    </span>
                  </label>
                  <label className="flex items-center gap-2 cursor-pointer">
                    <input
                      type="radio"
                      checked={!formData.is_active}
                      onChange={() =>
                        setFormData({ ...formData, is_active: false })
                      }
                      className="text-primary focus:ring-primary"
                    />
                    <span className="text-white">
                      {uiStrings.label_inactive || "Inactive"}
                    </span>
                  </label>
                </div>
              </div>
              <div className="md:col-span-2">
                <label className="block text-sm font-medium text-text-muted mb-2">
                  {uiStrings.label_address || "Address"}
                </label>
                <textarea
                  value={formData.address}
                  onChange={(e) =>
                    setFormData({ ...formData, address: e.target.value })
                  }
                  rows={3}
                  className="w-full bg-background-dark border border-border-color rounded-lg px-4 py-2 text-white focus:ring-1 focus:ring-primary focus:border-primary"
                />
              </div>
            </div>

            <div className="flex justify-end pt-4 border-t border-border-color">
              <button
                type="button"
                onClick={() => navigate("/maintenance/vendors")}
                className="px-4 py-2 mr-3 text-sm font-medium text-text-muted hover:text-white transition-colors"
              >
                {uiStrings.btn_cancel || commonStrings?.cancel || "Cancel"}
              </button>
              <button
                type="submit"
                disabled={loading}
                className="px-6 py-2 text-sm font-medium bg-primary hover:bg-primary/90 text-white rounded-lg shadow-[0_0_15px_rgba(19,127,236,0.3)] transition-all disabled:opacity-50 flex items-center gap-2"
              >
                {loading
                  ? uiStrings.label_saving || "Saving..."
                  : uiStrings.btn_save_vendor || "Save Vendor"}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
};

export default VendorForm;
