import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { notificationService } from '../../../services/notificationService';
import { GlobalNotificationDefault } from '../../../types/notification';

const GlobalNotificationSettings: React.FC = () => {
  const navigate = useNavigate();
  const [defaults, setDefaults] = useState<GlobalNotificationDefault | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [form, setForm] = useState({
    email_enabled: true,
    sms_enabled: false,
    push_enabled: true,
    in_app_enabled: true
  });

  const fetchData = async () => {
    setLoading(true);
    try {
      const data = await notificationService.getGlobalDefaults();
      setDefaults(data);
      if (data) {
        setForm({
          email_enabled: data.email_enabled,
          sms_enabled: data.sms_enabled,
          push_enabled: data.push_enabled,
          in_app_enabled: data.in_app_enabled
        });
      }
    } catch (err) {
      console.error('Failed to fetch global defaults', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  const handleSave = async () => {
    setSaving(true);
    try {
      await notificationService.updateGlobalDefaults(form);
      fetchData();
    } catch (err) {
      console.error('Failed to update global defaults', err);
    } finally {
      setSaving(false);
    }
  };

  const Toggle = ({ enabled, label, sublabel, onChange }: { enabled: boolean, label: string, sublabel: string, onChange: (val: boolean) => void }) => (
    <div className={`p-6 rounded-3xl border transition-all ${
      enabled ? 'bg-indigo-500/5 border-indigo-500/20 shadow-lg shadow-indigo-900/5' : 'bg-slate-900/30 border-slate-800 opacity-60'
    }`}>
      <div className="flex items-start justify-between">
        <div className="flex gap-4">
          <div className={`w-12 h-12 rounded-2xl flex items-center justify-center transition-all ${
            enabled ? 'bg-indigo-500/20 text-indigo-400 shadow-inner shadow-white/5' : 'bg-slate-800 text-slate-500'
          }`}>
            <span className="material-symbols-outlined">notifications_active</span>
          </div>
          <div>
            <span className="block font-bold text-slate-100 text-lg uppercase tracking-tight">{label}</span>
            <span className="block text-sm text-slate-500 mt-1 max-w-xs">{sublabel}</span>
          </div>
        </div>
        <button 
          onClick={() => onChange(!enabled)}
          className={`relative inline-flex h-7 w-12 items-center rounded-full transition-all outline-none ${
            enabled ? 'bg-indigo-500 shadow-xl shadow-indigo-500/20' : 'bg-slate-700'
          }`}
        >
          <span className={`inline-block h-5 w-5 transform rounded-full bg-white transition-transform ${enabled ? 'translate-x-6' : 'translate-x-1'}`} />
        </button>
      </div>
    </div>
  );

  if (loading) return <div className="p-8 text-slate-500">Decrypting system defaults...</div>;

  return (
    <div className="p-6 space-y-8 max-w-4xl mx-auto animate-in fade-in slide-in-from-top-4 duration-700">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <button onClick={() => navigate(-1)} className="p-2.5 rounded-xl bg-slate-800/50 hover:bg-slate-700 border border-slate-700 transition-all text-slate-300 flex items-center justify-center">
            <span className="material-symbols-outlined">arrow_back</span>
          </button>
          <div>
            <h1 className="text-3xl font-black bg-gradient-to-br from-indigo-400 to-purple-400 bg-clip-text text-transparent flex items-center gap-3">
              <span className="material-symbols-outlined text-[32px]">display_settings</span>
              Global Archetypes
            </h1>
            <p className="text-slate-500 font-medium">Define default notification behavior for all new platform users</p>
          </div>
        </div>
        <button 
          disabled={saving}
          onClick={handleSave}
          className="px-6 py-3 rounded-2xl bg-indigo-600 hover:bg-indigo-500 disabled:opacity-30 text-white font-bold transition-all flex items-center gap-2 shadow-2xl shadow-indigo-900/40 border border-indigo-400/20 uppercase"
        >
          <span className="material-symbols-outlined text-[20px]">save</span> {saving ? 'SYNCING...' : 'ENFORCE DEFAULTS'}
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <Toggle 
          enabled={form.email_enabled} 
          label="Default Email" 
          sublabel="Primary delivery channel for system reports and security updates."
          onChange={(val) => setForm({...form, email_enabled: val})} 
        />
        <Toggle 
          enabled={form.sms_enabled} 
          label="Default SMS" 
          sublabel="Used for high-priority alerts. Disabled by default due to high latency/cost."
          onChange={(val) => setForm({...form, sms_enabled: val})} 
        />
        <Toggle 
          enabled={form.push_enabled} 
          label="Default Push" 
          sublabel="Real-time mobile engagement for fleet operations and tracking."
          onChange={(val) => setForm({...form, push_enabled: val})} 
        />
        <Toggle 
          enabled={form.in_app_enabled} 
          label="Default Feed" 
          sublabel="The in-app notification center (System UI Feed)."
          onChange={(val) => setForm({...form, in_app_enabled: val})} 
        />
      </div>

      <div className="p-6 rounded-3xl bg-indigo-500/10 border border-indigo-500/20 flex items-start gap-4">
        <div className="p-2 rounded-xl bg-indigo-500/20 text-indigo-400">
          <span className="material-symbols-outlined">info</span>
        </div>
        <div className="space-y-1">
          <h4 className="font-bold text-indigo-200 uppercase tracking-widest text-xs">Propagate to All Users</h4>
          <p className="text-sm text-indigo-400/80 leading-relaxed">
            Changing these defaults only affects **future** users. Existing users retain their custom preferences. To perform a mass reset of existing users, contact the infrastructure guardian.
          </p>
        </div>
      </div>

      <div className="pt-8 border-t border-slate-800">
        <div className="flex items-center justify-between text-[10px] text-slate-600 font-mono uppercase tracking-[0.2em]">
          <span>System Archetype Revision: 1.0.4</span>
          <span>Last Enforced: {defaults?.updated_at ? new Date(defaults.updated_at).toLocaleString() : 'NEVER'}</span>
        </div>
      </div>
    </div>
  );
};

export default GlobalNotificationSettings;
