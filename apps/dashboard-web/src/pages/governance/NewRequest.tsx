import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useUIStrings } from "../../hooks/useUIStrings";
import GovernanceService from "../../services/GovernanceService";
import {
    CreateGovernanceRequest,
    GovernanceRequestType,
} from "../../types/governance";

const NewRequest: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const { uiStrings } = useUIStrings("governance");

  const [formData, setFormData] = useState<CreateGovernanceRequest>({
    request_type: "EXPORT",
    data_scope: "",
    purpose: "",
    format: "CSV",
    delivery_method: "DIRECT",
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      await GovernanceService.createRequest(formData);
      navigate("/admin/governance");
    } catch (error) {
      console.error("Failed to create request:", error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex-1 flex flex-col min-w-0 bg-background-dark overflow-hidden h-full">
      <div className="w-full bg-surface-darker border-b border-border-color px-6 py-3 flex items-center justify-between gap-4 flex-shrink-0 z-10">
        <div className="flex items-center gap-3">
          <button
            onClick={() => navigate("/admin/governance")}
            className="p-1.5 text-text-muted hover:text-white hover:bg-surface-dark rounded transition-colors"
          >
            <span className="material-symbols-outlined text-[20px]">
              arrow_back
            </span>
          </button>
          <div className="text-lg font-semibold text-white tracking-tight">
            {uiStrings?.new_request_title || "New Governance Request"}
          </div>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto p-6 bg-[radial-gradient(circle_at_top_right,rgba(19,127,236,0.05),transparent)]">
        <div className="max-w-2xl mx-auto">
          <form
            onSubmit={handleSubmit}
            className="space-y-6 bg-surface-dark/50 backdrop-blur-md border border-border-color rounded-2xl p-8 shadow-xl"
          >
            <div className="space-y-2">
              <label className="text-xs font-semibold text-text-muted uppercase tracking-wider">
                {uiStrings?.request_type_label || "Request Type"}
              </label>
              <div className="grid grid-cols-3 gap-3">
                {(
                  [
                    "EXPORT",
                    "DELETION",
                    "ACCESS_REVIEW",
                  ] as GovernanceRequestType[]
                ).map((type) => (
                  <button
                    key={type}
                    type="button"
                    onClick={() =>
                      setFormData({ ...formData, request_type: type })
                    }
                    className={`flex flex-col items-center justify-center p-4 rounded-xl border transition-all ${
                      formData.request_type === type
                        ? "border-primary bg-primary/10 text-primary ring-1 ring-primary"
                        : "border-border-color bg-surface-darker/50 text-text-muted hover:border-text-muted"
                    }`}
                  >
                    <span className="material-symbols-outlined mb-2 text-[24px]">
                      {type === "EXPORT"
                        ? "download"
                        : type === "DELETION"
                          ? "delete"
                          : "policy"}
                    </span>
                    <span className="text-[10px] font-bold uppercase">
                      {type === "EXPORT"
                        ? uiStrings?.export_label || "Export"
                        : type === "DELETION"
                          ? uiStrings?.deletion_label || "Deletion"
                          : uiStrings?.review_label || "Review"}
                    </span>
                  </button>
                ))}
              </div>
            </div>

            <div className="space-y-2">
              <label className="text-xs font-semibold text-text-muted uppercase tracking-wider">
                {uiStrings?.scope_label || "Data Scope"}
              </label>
              <textarea
                required
                value={formData.data_scope}
                onChange={(e) =>
                  setFormData({ ...formData, data_scope: e.target.value })
                }
                placeholder={
                  uiStrings?.scope_placeholder ||
                  "Describe the data you need (e.g., Driver logs, Vehicle telemetry...)"
                }
                className="w-full bg-background-dark border border-border-color text-white rounded-xl px-4 py-3 min-h-[100px] outline-none focus:border-primary transition-colors text-sm"
              />
            </div>

            <div className="space-y-2">
              <label className="text-xs font-semibold text-text-muted uppercase tracking-wider">
                {uiStrings?.purpose_label || "Purpose of Request"}
              </label>
              <textarea
                required
                value={formData.purpose}
                onChange={(e) =>
                  setFormData({ ...formData, purpose: e.target.value })
                }
                placeholder={
                  uiStrings?.purpose_placeholder_form ||
                  "What is the business or legal justification for this request?"
                }
                className="w-full bg-background-dark border border-border-color text-white rounded-xl px-4 py-3 min-h-[100px] outline-none focus:border-primary transition-colors text-sm"
              />
            </div>

            <div className="grid grid-cols-2 gap-6">
              <div className="space-y-2">
                <label className="text-xs font-semibold text-text-muted uppercase tracking-wider">
                  {uiStrings?.format_label || "Format"}
                </label>
                <select
                  value={formData.format}
                  onChange={(e) =>
                    setFormData({ ...formData, format: e.target.value })
                  }
                  className="w-full bg-background-dark border border-border-color text-white rounded-xl px-4 py-3 outline-none focus:border-primary transition-colors text-sm appearance-none"
                >
                  <option value="CSV">
                    {uiStrings?.csv_label || "CSV Spreadsheet"}
                  </option>
                  <option value="JSON">
                    {uiStrings?.json_label || "JSON Data"}
                  </option>
                  <option value="PDF">
                    {uiStrings?.pdf_label || "PDF Report"}
                  </option>
                </select>
              </div>

              <div className="space-y-2">
                <label className="text-xs font-semibold text-text-muted uppercase tracking-wider">
                  {uiStrings?.delivery_method_label || "Delivery Method"}
                </label>
                <select
                  value={formData.delivery_method}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      delivery_method: e.target.value,
                    })
                  }
                  className="w-full bg-background-dark border border-border-color text-white rounded-xl px-4 py-3 outline-none focus:border-primary transition-colors text-sm appearance-none"
                >
                  <option value="DIRECT">
                    {uiStrings?.direct_label || "Direct Download"}
                  </option>
                  <option value="SECURE_EMAIL">
                    {uiStrings?.secure_email_label || "Secure Email"}
                  </option>
                  <option value="SFTP">
                    {uiStrings?.sftp_label || "SFTP Upload"}
                  </option>
                </select>
              </div>
            </div>

            <div className="pt-4 flex gap-4">
              <button
                type="button"
                onClick={() => navigate("/admin/governance")}
                className="flex-1 py-3 px-6 rounded-xl border border-border-color text-white font-medium hover:bg-surface-dark transition-colors"
              >
                {uiStrings?.cancel_btn || "Cancel"}
              </button>
              <button
                type="submit"
                disabled={loading}
                className="flex-[2] py-3 px-6 rounded-xl bg-primary text-white font-semibold hover:bg-primary/90 shadow-[0_4px_14px_rgba(19,127,236,0.39)] transition-all disabled:opacity-50"
              >
                {loading
                  ? uiStrings?.submitting_btn || "Submitting..."
                  : uiStrings?.submit_btn || "Submit Request"}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
};

export default NewRequest;
