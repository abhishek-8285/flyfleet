import React, { useEffect, useState } from "react";
import { useUIStrings } from "../../../hooks/useUIStrings";
import apiClient from "../../../services/apiClient";

interface SystemLimits {
  id: string;
  api_rate_limit: number;
  burst_limit: number;
  audit_retention: string;
  activity_retention: string;
}

const PoliciesLimits: React.FC = () => {
  const [limits, setLimits] = useState<SystemLimits>({
    id: "",
    api_rate_limit: 10000,
    burst_limit: 500,
    audit_retention: "1y",
    activity_retention: "180d",
  });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { uiStrings } = useUIStrings("admin_settings");

  useEffect(() => {
    fetchLimits();
  }, []);

  const fetchLimits = async () => {
    setError(null);
    try {
      const response = (await apiClient.get("/settings/limits")) as any;
      const data = response?.data ?? response;
      if (data) {
        setLimits((prev) => ({
          id: data.id || prev.id,
          api_rate_limit: data.api_rate_limit ?? prev.api_rate_limit,
          burst_limit: data.burst_limit ?? prev.burst_limit,
          audit_retention: data.audit_retention || prev.audit_retention,
          activity_retention:
            data.activity_retention || prev.activity_retention,
        }));
      }
    } catch (error) {
      console.error("Failed to fetch system limits", error);
      setError(uiStrings?.error_load_limits || "Failed to load system limits");
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      await apiClient.put("/settings/limits", limits);
      alert(
        uiStrings?.msg_save_success ||
          "Policies and limits saved successfully!",
      );
    } catch (error) {
      console.error("Failed to save limits", error);
      alert(uiStrings?.msg_error_save || "Failed to save limits");
    } finally {
      setSaving(false);
    }
  };

  // Parse retention string to number/unit for UI if needed, but simple select for now

  if (loading)
    return (
      <div className="p-6 text-white">{uiStrings?.loading || "Loading..."}</div>
    );

  return (
    <main className="flex-1 flex flex-col min-w-0 bg-[#0b1219] overflow-hidden">
      <header className="h-16 flex items-center justify-between px-6 border-b border-[#233648] bg-[#111a22] flex-shrink-0 z-20">
        <div className="flex items-center gap-4">
          <div>
            <h2 className="text-white text-lg font-bold">
              {uiStrings?.title_policies || "System Policies & Limits"}
            </h2>
            <div className="flex items-center gap-2 text-[11px] text-[#92adc9] mt-0.5 uppercase tracking-wide font-medium">
              <span>{uiStrings?.breadcrumb_system || "System Settings"}</span>
              <span className="material-symbols-outlined text-[10px]">
                chevron_right
              </span>
              <span className="text-white">
                {uiStrings?.breadcrumb_policies || "Policies"}
              </span>
            </div>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <button className="px-4 py-2 bg-[#16202a] hover:bg-[#233648] text-[#92adc9] hover:text-white border border-[#233648] text-xs font-semibold rounded transition-colors">
            {uiStrings?.btn_cancel || "Cancel"}
          </button>
          <button
            onClick={handleSave}
            disabled={saving}
            className="px-4 py-2 bg-[#137fec] hover:bg-[#0f6bd0] text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors shadow-lg shadow-[#137fec]/20 disabled:opacity-50"
          >
            <span className="material-symbols-outlined text-[16px]">save</span>
            {saving
              ? uiStrings?.btn_saving || "Saving..."
              : uiStrings?.btn_save_policies || "Save Policies"}
          </button>
        </div>
      </header>
      <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
        {error && (
          <div className="mb-4 rounded-lg border border-red-500/40 bg-red-500/10 px-4 py-3 text-sm text-red-200">
            {error}
          </div>
        )}
        <div className="max-w-5xl mx-auto space-y-6">
          <div className="bg-[#16202a] rounded-xl border border-[#233648] overflow-hidden">
            <div className="p-4 border-b border-[#233648] bg-[#111a22]/50 flex justify-between items-center">
              <h3 className="text-white font-semibold text-sm flex items-center gap-2">
                <span className="material-symbols-outlined text-[#137fec] text-[20px]">
                  speed
                </span>
                {uiStrings?.card_api_limits || "API Rate Limiting"}
              </h3>
              <span className="text-[10px] text-[#92adc9] uppercase tracking-wide">
                {uiStrings?.label_performance || "Performance"}
              </span>
            </div>
            <div className="p-6 grid grid-cols-1 md:grid-cols-2 gap-8">
              <div>
                <label className="block text-xs font-medium text-[#92adc9] mb-1.5">
                  {uiStrings?.label_rate_limit ||
                    "Global Rate Limit (Requests/Hour)"}
                </label>
                <input
                  className="w-full bg-[#111a22] border border-[#233648] rounded px-3 py-2 text-sm text-white focus:outline-none focus:border-[#137fec] focus:ring-1 focus:ring-[#137fec] placeholder-gray-600 font-mono"
                  type="number"
                  value={limits.api_rate_limit}
                  onChange={(e) =>
                    setLimits({
                      ...limits,
                      api_rate_limit: parseInt(e.target.value),
                    })
                  }
                />
                <p className="text-[10px] text-[#92adc9] mt-1">
                  {uiStrings?.help_rate_limit ||
                    "Maximum number of API requests allowed per tenant per hour."}
                </p>
              </div>
              <div>
                <label className="block text-xs font-medium text-[#92adc9] mb-1.5">
                  {uiStrings?.label_burst || "Burst Capability (Requests/Sec)"}
                </label>
                <input
                  className="w-full bg-[#111a22] border border-[#233648] rounded px-3 py-2 text-sm text-white focus:outline-none focus:border-[#137fec] focus:ring-1 focus:ring-[#137fec] placeholder-gray-600 font-mono"
                  type="number"
                  value={limits.burst_limit}
                  onChange={(e) =>
                    setLimits({
                      ...limits,
                      burst_limit: parseInt(e.target.value),
                    })
                  }
                />
                <p className="text-[10px] text-[#92adc9] mt-1">
                  {uiStrings?.help_burst ||
                    "Short-term spike allowance to prevent throttling during high load."}
                </p>
              </div>
            </div>
          </div>

          <div className="bg-[#16202a] rounded-xl border border-[#233648] overflow-hidden">
            <div className="p-4 border-b border-[#233648] bg-[#111a22]/50 flex justify-between items-center">
              <h3 className="text-white font-semibold text-sm flex items-center gap-2">
                <span className="material-symbols-outlined text-[#137fec] text-[20px]">
                  inventory_2
                </span>
                {uiStrings?.card_retention || "Data Retention"}
              </h3>
              <span className="text-[10px] text-[#92adc9] uppercase tracking-wide">
                {uiStrings?.label_storage_compliance || "Storage & Compliance"}
              </span>
            </div>
            <div className="p-6 grid grid-cols-1 md:grid-cols-2 gap-8">
              <div>
                <label className="block text-xs font-medium text-[#92adc9] mb-1.5">
                  {uiStrings?.label_audit_retention || "Audit Log Retention"}
                </label>
                <div className="relative">
                  <select
                    className="w-full bg-[#111a22] border border-[#233648] rounded px-3 py-2 text-sm text-white focus:outline-none focus:border-[#137fec] appearance-none"
                    value={limits.audit_retention}
                    onChange={(e) =>
                      setLimits({ ...limits, audit_retention: e.target.value })
                    }
                  >
                    <option value="90d">90 Days</option>
                    <option value="180d">180 Days (6 Months)</option>
                    <option value="1y">1 Year</option>
                    <option value="3y">3 Years</option>
                    <option value="forever">Indefinite (Forever)</option>
                  </select>
                  <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center px-2 text-[#92adc9]">
                    <span className="material-symbols-outlined text-[16px]">
                      expand_more
                    </span>
                  </div>
                </div>
                <p className="text-[10px] text-[#92adc9] mt-1">
                  {uiStrings?.help_audit_retention ||
                    "Period to keep security and compliance audit logs."}
                </p>
              </div>
              <div>
                <label className="block text-xs font-medium text-[#92adc9] mb-1.5">
                  {uiStrings?.label_activity_retention ||
                    "Activity Log Retention"}
                </label>
                <div className="relative">
                  <select
                    className="w-full bg-[#111a22] border border-[#233648] rounded px-3 py-2 text-sm text-white focus:outline-none focus:border-[#137fec] appearance-none"
                    value={limits.activity_retention}
                    onChange={(e) =>
                      setLimits({
                        ...limits,
                        activity_retention: e.target.value,
                      })
                    }
                  >
                    <option value="30d">30 Days</option>
                    <option value="90d">90 Days</option>
                    <option value="180d">180 Days</option>
                    <option value="1y">1 Year</option>
                  </select>
                  <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center px-2 text-[#92adc9]">
                    <span className="material-symbols-outlined text-[16px]">
                      expand_more
                    </span>
                  </div>
                </div>
                <p className="text-[10px] text-[#92adc9] mt-1">
                  {uiStrings?.help_activity_retention ||
                    "Period to keep general user activity and debug logs."}
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>
  );
};

export default PoliciesLimits;
