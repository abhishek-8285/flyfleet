import React, { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useUIStrings } from "../../../hooks/useUIStrings";
import { notificationService } from "../../../services/notificationService";
import {
    PreferenceAudit,
    UserNotificationPreference,
} from "../../../types/notification";

const UserPreferenceDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [pref, setPref] = useState<UserNotificationPreference | null>(null);
  const [audits, setAudits] = useState<PreferenceAudit[]>([]);
  const [loading, setLoading] = useState(true);
  const [editing, setEditing] = useState(false);
  const [saving, setSaving] = useState(false);
  const { uiStrings: rawUiStrings } = useUIStrings("notifications");
  const uiStrings = rawUiStrings || {};
  const [form, setForm] = useState({
    email_enabled: true,
    sms_enabled: false,
    push_enabled: true,
    in_app_enabled: true,
    reason: "",
  });

  const fetchData = async () => {
    if (!id) return;
    setLoading(true);
    try {
      const [prefData, auditData] = await Promise.all([
        notificationService.getUserPreference(id),
        notificationService.getPreferenceAudit(id),
      ]);
      setPref(prefData);
      setAudits(auditData);
      if (prefData) {
        setForm({
          email_enabled: prefData.email_enabled,
          sms_enabled: prefData.sms_enabled,
          push_enabled: prefData.push_enabled,
          in_app_enabled: prefData.in_app_enabled,
          reason: "",
        });
      }
    } catch (err) {
      console.error("Failed to fetch preference detail", err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [id]);

  const handleUpdate = async () => {
    if (!id || !form.reason) return;
    setSaving(true);
    try {
      await notificationService.updateUserPreference(id, form);
      setEditing(false);
      fetchData();
    } catch (err) {
      console.error("Failed to update preference", err);
    } finally {
      setSaving(false);
    }
  };

  const Toggle = ({
    enabled,
    label,
    onChange,
  }: {
    enabled: boolean;
    label: string;
    onChange?: (val: boolean) => void;
  }) => (
    <div
      className={`flex items-center justify-between p-4 rounded-2xl border transition-all ${
        enabled
          ? "bg-blue-500/5 border-blue-500/30"
          : "bg-slate-900/50 border-slate-800 opacity-60"
      }`}
    >
      <div className="flex items-center gap-3">
        <div
          className={`w-10 h-10 rounded-xl flex items-center justify-center ${enabled ? "bg-blue-500/20 text-blue-400" : "bg-slate-800 text-slate-500"}`}
        >
          <span className="material-symbols-outlined">
            notifications_active
          </span>
        </div>
        <span
          className={`font-semibold ${enabled ? "text-slate-100" : "text-slate-500"}`}
        >
          {label}
        </span>
      </div>
      <button
        disabled={!editing}
        onClick={() => onChange?.(!enabled)}
        className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors outline-none cursor-pointer ${
          enabled ? "bg-blue-600" : "bg-slate-700"
        } ${!editing && "cursor-not-allowed"}`}
      >
        <span
          className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${enabled ? "translate-x-6" : "translate-x-1"}`}
        />
      </button>
    </div>
  );

  if (loading)
    return (
      <div className="p-8 text-slate-500">
        {uiStrings?.msg_loading_context ||
          "Loading preference security context..."}
      </div>
    );
  if (!pref)
    return (
      <div className="p-8 text-red-500">
        {uiStrings?.msg_user_not_found || "User not found in registry."}
      </div>
    );

  return (
    <div className="p-6 space-y-6 max-w-5xl mx-auto animate-in fade-in zoom-in-95 duration-500">
      <div className="flex items-center gap-4">
        <button
          onClick={() => navigate(-1)}
          className="p-2 rounded-xl bg-slate-800/50 hover:bg-slate-700 border border-slate-700 transition-all text-slate-300 flex items-center justify-center"
        >
          <span className="material-symbols-outlined">arrow_back</span>
        </button>
        <div>
          <h2 className="text-2xl font-bold text-slate-100">
            {pref.user_name || uiStrings?.default_user_name || "System User"}
          </h2>
          <p className="text-sm text-slate-500 font-mono">ID: {pref.user_id}</p>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2 space-y-6">
          {/* Main Controls */}
          <section className="p-6 rounded-3xl bg-slate-900/40 border border-slate-800 backdrop-blur-xl relative overflow-hidden group">
            <div className="absolute top-0 right-0 p-8 opacity-5 group-hover:opacity-10 transition-opacity">
              <span className="material-symbols-outlined text-[120px]">
                security
              </span>
            </div>

            <div className="flex items-center justify-between mb-8">
              <h3 className="text-lg font-bold text-slate-200 flex items-center gap-2">
                <span className="material-symbols-outlined text-blue-400 text-[20px]">
                  edit
                </span>
                {uiStrings?.title_channel_governance || "Channel Governance"}
              </h3>
              {!editing ? (
                <button
                  onClick={() => setEditing(true)}
                  className="px-4 py-2 rounded-xl bg-blue-600 hover:bg-blue-500 text-white text-sm font-bold transition-all uppercase"
                >
                  {uiStrings?.btn_modify_rules || "MODIFY RULES"}
                </button>
              ) : (
                <div className="flex gap-2">
                  <button
                    onClick={() => setEditing(false)}
                    className="px-4 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 text-sm font-bold transition-all uppercase"
                  >
                    {uiStrings?.btn_cancel || "CANCEL"}
                  </button>
                  <button
                    disabled={saving || !form.reason}
                    onClick={handleUpdate}
                    className="px-4 py-2 rounded-xl bg-green-600 hover:bg-green-500 disabled:opacity-30 text-white text-sm font-bold transition-all flex items-center gap-2 uppercase"
                  >
                    <span className="material-symbols-outlined text-[18px]">
                      save
                    </span>{" "}
                    {saving
                      ? uiStrings?.btn_saving || "SAVING..."
                      : uiStrings?.btn_save_protect || "PROTECT & SAVE"}
                  </button>
                </div>
              )}
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <Toggle
                enabled={form.email_enabled}
                label={uiStrings?.label_email || "Email Delivery"}
                onChange={(val) => setForm({ ...form, email_enabled: val })}
              />
              <Toggle
                enabled={form.sms_enabled}
                label={uiStrings?.label_sms || "SMS / Mobile"}
                onChange={(val) => setForm({ ...form, sms_enabled: val })}
              />
              <Toggle
                enabled={form.push_enabled}
                label={uiStrings?.label_push || "Push Notifications"}
                onChange={(val) => setForm({ ...form, push_enabled: val })}
              />
              <Toggle
                enabled={form.in_app_enabled}
                label={uiStrings?.label_in_app || "In-App Feed"}
                onChange={(val) => setForm({ ...form, in_app_enabled: val })}
              />
            </div>

            {editing && (
              <div className="mt-8 p-4 rounded-2xl bg-blue-500/5 border border-blue-500/20 animate-in slide-in-from-top-2 duration-300">
                <label className="block text-xs font-bold text-blue-400 uppercase tracking-widest mb-2">
                  {uiStrings?.label_justification ||
                    "Change Justification (Required)"}
                </label>
                <textarea
                  className="w-full bg-slate-950/50 border border-slate-700 rounded-xl p-3 text-slate-200 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 outline-none transition-all h-24"
                  placeholder={
                    uiStrings?.placeholder_justification ||
                    "Explain why these communication rules are being modified..."
                  }
                  value={form.reason}
                  onChange={(e) => setForm({ ...form, reason: e.target.value })}
                />
                <div className="flex items-center gap-2 mt-2 text-[10px] text-slate-500 uppercase tracking-tighter">
                  <span className="material-symbols-outlined text-[14px]">
                    info
                  </span>{" "}
                  {uiStrings?.msg_audit_trail ||
                    "Audit trail will record this justification for compliance."}
                </div>
              </div>
            )}
          </section>

          {/* System Alerts Restriction */}
          <div className="p-4 rounded-2xl bg-amber-900/20 border border-amber-500/30 flex items-start gap-3">
            <span className="material-symbols-outlined text-amber-500 mt-0.5">
              security
            </span>
            <div>
              <h4 className="text-sm font-bold text-amber-200 uppercase tracking-tight">
                {uiStrings?.title_immutable_alerts ||
                  "System Alerts are Immutable"}
              </h4>
              <p className="text-xs text-amber-400/80 mt-1 leading-relaxed">
                {uiStrings?.desc_immutable_alerts ||
                  "Critical system alerts (security breaches, mfa failures) ignore these preferences and are always delivered via all available channels to ensure platform integrity."}
              </p>
            </div>
          </div>
        </div>

        {/* Audit Timeline */}
        <div className="space-y-6">
          <section className="p-6 rounded-3xl bg-slate-900/20 border border-slate-800 h-full overflow-hidden flex flex-col">
            <h3 className="text-lg font-bold text-slate-200 flex items-center gap-2 mb-6">
              <span className="material-symbols-outlined text-purple-400 text-[22px]">
                history
              </span>
              {uiStrings?.title_audit_timeline || "Audit Timeline"}
            </h3>

            <div className="space-y-6 overflow-y-auto pr-2 custom-scrollbar flex-1">
              {audits.length === 0 ? (
                <div className="py-12 text-center text-slate-600 text-sm italic">
                  {uiStrings?.msg_no_history || "No history recorded yet."}
                </div>
              ) : (
                audits.map((audit) => (
                  <div
                    key={audit.id}
                    className="relative pl-6 border-l border-slate-800 pb-2 last:pb-0 group"
                  >
                    <div className="absolute -left-1.5 top-0 w-3 h-3 rounded-full bg-slate-800 border border-slate-700 group-hover:bg-blue-500 transition-colors" />
                    <div className="flex flex-col">
                      <span className="text-xs font-bold text-slate-400 uppercase tracking-tighter">
                        {new Date(audit.created_at).toLocaleString()}
                      </span>
                      <p className="text-sm text-slate-200 mt-2 font-medium bg-slate-800/50 p-3 rounded-xl border border-slate-800 italic">
                        "{audit.change_reason}"
                      </p>
                      <span className="text-[10px] text-slate-600 mt-2 flex items-center gap-1 font-mono uppercase">
                        Admin: {audit.changed_by.slice(0, 8)}
                      </span>
                    </div>
                  </div>
                ))
              )}
            </div>
          </section>
        </div>
      </div>
    </div>
  );
};

export default UserPreferenceDetail;
