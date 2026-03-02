import {
    ArrowLeftIcon,
    BuildingOffice2Icon,
    GlobeAltIcon,
    PaintBrushIcon,
} from "@heroicons/react/24/outline";
import React, { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useUIStrings } from "../../hooks/useUIStrings";
import apiClient from "../../services/apiClient";

export const EditOrganization: React.FC = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const { uiStrings } = useUIStrings("admin_organizations");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [formData, setFormData] = useState({
    name: "",
    domain: "",
    default_theme: "dark",
    status: "active",
  });

  useEffect(() => {
    const fetchOrg = async () => {
      try {
        const res = (await apiClient.get(
          `/admin/organizations?id=${id}`,
        )) as any;
        const org = res.data;
        setFormData({
          name: org.name,
          domain: org.domain,
          default_theme: org.default_theme,
          status: org.status,
        });
      } catch (error) {
        console.error("Failed to fetch org:", error);
      } finally {
        setLoading(false);
      }
    };
    fetchOrg();
  }, [id]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    try {
      await apiClient.put(`/admin/organizations?id=${id}`, formData);
      alert(uiStrings?.msg_success || "Organization updated successfully");
      navigate(`/admin/organizations/${id}`);
    } catch (error) {
      alert(uiStrings?.msg_error || "Failed to update organization");
    } finally {
      setSaving(false);
    }
  };

  if (loading)
    return (
      <div className="p-12 text-center text-slate-500">
        {uiStrings?.loading || "Retrieving profile..."}
      </div>
    );

  return (
    <div className="p-8 max-w-4xl mx-auto">
      <button
        onClick={() => navigate(`/admin/organizations/${id}`)}
        className="flex items-center space-x-2 text-slate-500 hover:text-blue-600 transition-colors mb-8 group"
      >
        <ArrowLeftIcon className="h-4 w-4 group-hover:-translate-x-1 transition-transform" />
        <span>{uiStrings?.btn_back || "Back to Profile"}</span>
      </button>

      <div className="bg-white rounded-3xl shadow-xl border border-slate-100 overflow-hidden">
        <div className="px-10 py-8 border-b border-slate-100 bg-slate-50/50">
          <h1 className="text-2xl font-bold text-slate-800">
            {uiStrings?.title_edit || "Edit Organization Profile"}
          </h1>
          <p className="text-slate-500 text-sm mt-1">
            {uiStrings?.subtitle_edit ||
              "Update identity and workspace defaults"}
          </p>
        </div>

        <form onSubmit={handleSubmit} className="p-10 space-y-8">
          <div className="grid grid-cols-2 gap-8">
            <div className="space-y-4">
              <label className="text-sm font-semibold text-slate-700 flex items-center">
                <BuildingOffice2Icon className="h-4 w-4 mr-2 text-blue-500" />
                {uiStrings?.label_name || "Organization Name"}
              </label>
              <input
                type="text"
                className="w-full bg-white border border-slate-200 rounded-xl px-4 py-3 focus:ring-2 focus:ring-blue-500/20 outline-none"
                value={formData.name}
                onChange={(e) =>
                  setFormData({ ...formData, name: e.target.value })
                }
                required
              />
            </div>
            <div className="space-y-4">
              <label className="text-sm font-semibold text-slate-700 flex items-center">
                <GlobeAltIcon className="h-4 w-4 mr-2 text-blue-500" />
                {uiStrings?.label_domain || "Domain"}
              </label>
              <input
                type="text"
                className="w-full bg-white border border-slate-200 rounded-xl px-4 py-3 focus:ring-2 focus:ring-blue-500/20 outline-none"
                value={formData.domain}
                onChange={(e) =>
                  setFormData({ ...formData, domain: e.target.value })
                }
              />
            </div>
          </div>

          <div className="space-y-4">
            <label className="text-sm font-semibold text-slate-700 flex items-center">
              <PaintBrushIcon className="h-4 w-4 mr-2 text-blue-500" />
              {uiStrings?.label_status || "Status"}
            </label>
            <select
              className="w-full bg-white border border-slate-200 rounded-xl px-4 py-3 focus:ring-2 focus:ring-blue-500/20 outline-none"
              value={formData.status}
              onChange={(e) =>
                setFormData({ ...formData, status: e.target.value })
              }
            >
              <option value="active">
                {uiStrings?.opt_active || "Active"}
              </option>
              <option value="suspended">
                {uiStrings?.opt_suspended || "Suspended"}
              </option>
              <option value="pending">
                {uiStrings?.opt_pending || "Pending"}
              </option>
            </select>
          </div>

          <div className="pt-8 border-t border-slate-100 flex justify-end space-x-4">
            <button
              type="button"
              onClick={() => navigate(`/admin/organizations/${id}`)}
              className="px-6 py-3 text-slate-500 font-medium hover:text-slate-800"
            >
              {uiStrings?.btn_cancel || "Cancel"}
            </button>
            <button
              type="submit"
              disabled={saving}
              className="bg-blue-600 hover:bg-blue-700 text-white px-10 py-3 rounded-xl font-semibold shadow-lg shadow-blue-900/20 transition-all"
            >
              {saving
                ? uiStrings?.btn_saving || "Saving..."
                : uiStrings?.btn_update || "Update Workspace"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
