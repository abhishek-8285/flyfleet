import {
    ArrowPathIcon,
    ChevronRightIcon,
    CircleStackIcon,
    ClockIcon,
    ExclamationCircleIcon,
    EyeSlashIcon,
    LockClosedIcon,
    ShieldCheckIcon,
} from "@heroicons/react/24/outline";
import React, { useEffect, useState } from "react";
import { useSelector } from "react-redux";
import type { RootState } from "../store";
import { useUIStrings } from "../hooks/useUIStrings";
import { api } from "../services/api";

const AuditSettings: React.FC = () => {
  const user = useSelector((state: RootState) => state.auth.user);
  const [settings, setSettings] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [success, setSuccess] = useState(false);
  const { uiStrings } = useUIStrings("audit_settings");

  useEffect(() => {
    const fetchSettings = async () => {
      if (!user?.org_id) return;
      try {
        const data = await api.getAuditSettings(user.org_id);
        setSettings(data);
      } catch (error) {
        console.error("Failed to fetch audit settings", error);
      } finally {
        setLoading(false);
      }
    };
    fetchSettings();

    fetchSettings();
  }, [user?.org_id]);

  const handleSave = async () => {
    if (!settings) return;
    setSaving(true);
    try {
      await api.updateAuditSettings(settings);
      setSuccess(true);
      setTimeout(() => setSuccess(false), 3000);
    } catch (error) {
      console.error("Failed to update settings", error);
    } finally {
      setSaving(false);
    }
  };

  if (loading)
    return (
      <div className="p-8 text-slate-400">
        {uiStrings?.loading || "Loading governance protocol..."}
      </div>
    );

  return (
    <div className="p-8 max-w-5xl mx-auto space-y-10 animate-in fade-in slide-in-from-bottom-4 duration-700">
      {/* Header */}
      <div className="flex items-center justify-between border-b border-white/5 pb-8">
        <div>
          <h1 className="text-4xl font-light tracking-tight text-white mb-2">
            {uiStrings?.title_prefix || "Audit"}{" "}
            <span className="font-semibold text-blue-500">
              {uiStrings?.title_highlight || "Governance"}
            </span>
          </h1>
          <p className="text-slate-400 text-lg font-light">
            {uiStrings?.subtitle ||
              "Configure data retention, masking, and automated security alerts"}
          </p>
        </div>
        <div className="flex items-center gap-4">
          <button
            onClick={() => setSettings({ ...settings })} // Reset logic placeholder
            className="p-3 bg-slate-800/50 border border-slate-700/50 rounded-xl text-slate-400 hover:text-white transition-all shadow-xl"
          >
            <ArrowPathIcon className="w-5 h-5" />
          </button>
          <button
            onClick={handleSave}
            disabled={saving}
            className="flex items-center gap-2 px-6 py-3 bg-blue-600 hover:bg-blue-500 text-white rounded-xl font-semibold transition-all shadow-lg shadow-blue-500/20 disabled:opacity-50"
          >
            {saving ? (
              <div className="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
            ) : (
              <ArrowPathIcon className="w-5 h-5" />
            )}
            {success
              ? uiStrings?.msg_success || "Settings Perfected"
              : uiStrings?.btn_apply || "Apply Changes"}
          </button>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-8 text-slate-300">
        {/* Retention Policy */}
        <div className="md:col-span-2 space-y-8">
          <section className="bg-slate-900/40 backdrop-blur-xl border border-white/5 rounded-3xl p-8 shadow-2xl relative overflow-hidden group">
            <div className="absolute top-0 right-0 p-8 opacity-5 group-hover:opacity-10 transition-opacity">
              <ClockIcon className="w-32 h-32 text-blue-400" />
            </div>
            <div className="flex items-center gap-4 mb-6">
              <div className="p-3 bg-blue-500/20 rounded-2xl">
                <ClockIcon className="w-6 h-6 text-blue-400" />
              </div>
              <h2 className="text-2xl font-light text-white">
                {uiStrings?.section_retention || "Retention Protocol"}
              </h2>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
              <div className="space-y-3">
                <label className="text-sm font-semibold text-slate-500 uppercase tracking-widest flex items-center gap-2">
                  {uiStrings?.label_online || "Online Access"}
                  <ExclamationCircleIcon className="w-3.5 h-3.5" />
                </label>
                <div className="relative group/input">
                  <input
                    type="number"
                    value={settings.retention_days_online}
                    onChange={(e) =>
                      setSettings({
                        ...settings,
                        retention_days_online: parseInt(e.target.value),
                      })
                    }
                    className="w-full bg-slate-950/50 border border-slate-800 rounded-2xl px-6 py-4 text-white text-lg font-light focus:outline-none focus:ring-2 focus:ring-blue-500/50 transition-all font-mono"
                  />
                  <span className="absolute right-6 top-1/2 -translate-y-1/2 text-slate-500 font-light">
                    Days
                  </span>
                </div>
                <p className="text-xs text-slate-500 italic">
                  {uiStrings?.hint_online ||
                    "Logs will be instantly searchable in the explorer during this period."}
                </p>
              </div>
              <div className="space-y-3">
                <label className="text-sm font-semibold text-slate-500 uppercase tracking-widest flex items-center gap-2">
                  {uiStrings?.label_archival || "Archival Storage"}
                  <CircleStackIcon className="w-3.5 h-3.5" />
                </label>
                <div className="relative">
                  <input
                    type="number"
                    value={settings.retention_days_archival}
                    onChange={(e) =>
                      setSettings({
                        ...settings,
                        retention_days_archival: parseInt(e.target.value),
                      })
                    }
                    className="w-full bg-slate-950/50 border border-slate-800 rounded-2xl px-6 py-4 text-white text-lg font-light focus:outline-none focus:ring-2 focus:ring-blue-500/50 transition-all font-mono"
                  />
                  <span className="absolute right-6 top-1/2 -translate-y-1/2 text-slate-500 font-light">
                    Days
                  </span>
                </div>
                <p className="text-xs text-slate-500 italic">
                  {uiStrings?.hint_archival ||
                    "Compressed cold storage for long-term compliance (7 years recommended)."}
                </p>
              </div>
            </div>
          </section>

          {/* Data Masking */}
          <section className="bg-slate-900/40 backdrop-blur-xl border border-white/5 rounded-3xl p-8 shadow-2xl relative overflow-hidden group">
            <div className="flex items-center gap-4 mb-6">
              <div className="p-3 bg-purple-500/20 rounded-2xl">
                <EyeSlashIcon className="w-6 h-6 text-purple-400" />
              </div>
              <h2 className="text-2xl font-light text-white">
                {uiStrings?.section_privacy || "Privacy Cloaking"}
              </h2>
            </div>
            <p className="text-slate-400 mb-8 font-light">
              {uiStrings?.hint_privacy ||
                "Define sensitive keys that should be masked with ******* in log details to ensure PII protection."}
            </p>

            <div className="flex flex-wrap gap-2">
              {settings.masking_rules.map((rule: string, i: number) => (
                <div
                  key={i}
                  className="flex items-center gap-3 pl-4 pr-2 py-2 bg-slate-950/50 border border-slate-800 rounded-xl group/tag hover:border-purple-500/50 transition-all"
                >
                  <span className="text-sm font-mono text-purple-300 tracking-wider">
                    {rule}
                  </span>
                  <button className="p-1 hover:bg-red-500/20 rounded-md text-slate-600 hover:text-red-400 transition-all">
                    <ChevronRightIcon className="w-4 h-4 rotate-90" />
                  </button>
                </div>
              ))}
              <button className="px-4 py-2 border border-dashed border-slate-700 rounded-xl text-slate-500 hover:text-white hover:border-white transition-all text-sm font-medium">
                {uiStrings?.btn_add_rule || "+ Add Rule"}
              </button>
            </div>
          </section>
        </div>

        {/* Automation & Security */}
        <div className="space-y-8">
          <section className="bg-slate-900/40 border border-white/5 rounded-3xl p-6 space-y-6">
            <div className="flex items-center gap-3">
              <ShieldCheckIcon className="w-5 h-5 text-green-400" />
              <h3 className="font-semibold text-white">
                {uiStrings?.section_guard || "System Guard"}
              </h3>
            </div>
            <div className="flex items-center justify-between p-4 bg-slate-950/30 rounded-2xl border border-slate-800/50">
              <div className="space-y-1">
                <p className="text-sm font-medium text-white">
                  {uiStrings?.label_critical || "Critical Alerts"}
                </p>
                <p className="text-[10px] text-slate-500">
                  {uiStrings?.hint_critical ||
                    "Notify SEC-OPS on severity: CRITICAL"}
                </p>
              </div>
              <div
                onClick={() =>
                  setSettings({
                    ...settings,
                    alert_on_critical: !settings.alert_on_critical,
                  })
                }
                className={`w-12 h-6 rounded-full p-1 cursor-pointer transition-colors ${settings.alert_on_critical ? "bg-blue-600" : "bg-slate-800"}`}
              >
                <div
                  className={`w-4 h-4 rounded-full bg-white transition-transform ${settings.alert_on_critical ? "translate-x-6" : "translate-x-0"}`}
                />
              </div>
            </div>
            <div className="flex items-center justify-between p-4 bg-slate-950/30 rounded-2xl border border-slate-800/50">
              <div className="space-y-1">
                <p className="text-sm font-medium text-white">
                  {uiStrings?.label_diffing || "Policy Diffing"}
                </p>
                <p className="text-[10px] text-slate-500">
                  {uiStrings?.hint_diffing || "Auto-calculate JSON deltas"}
                </p>
              </div>
              <div className="w-12 h-6 rounded-full bg-blue-600 p-1">
                <div className="w-4 h-4 rounded-full bg-white translate-x-6" />
              </div>
            </div>
          </section>

          <div className="bg-gradient-to-br from-blue-600/20 to-purple-600/20 p-8 rounded-3xl border border-blue-500/20 space-y-4">
            <LockClosedIcon className="w-10 h-10 text-white opacity-40" />
            <h4 className="text-xl font-bold text-white">
              {uiStrings?.compliance_title || "Compliance Standard"}
            </h4>
            <p className="text-sm text-slate-400 leading-relaxed font-light">
              {uiStrings?.compliance_text ||
                "These settings ensure alignment with global data protection and security logging requirements."}
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default AuditSettings;
