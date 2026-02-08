import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { notificationService } from '../../../services/notificationService';
import { NotificationSettings as SettingsType } from '../../../types/notification';

const NotificationSettings: React.FC = () => {
  const navigate = useNavigate();
  const [settings, setSettings] = useState<SettingsType[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    notificationService.getSettings()
      .then(setSettings)
      .catch(err => console.error('Failed to fetch settings', err))
      .finally(() => setLoading(false));
  }, []);

  const handleToggle = (index: number) => {
    const newSettings = [...settings];
    newSettings[index].enabled = !newSettings[index].enabled;
    setSettings(newSettings);
  };

  const handleChange = (index: number, field: keyof SettingsType, value: any) => {
    const newSettings = [...settings];
    (newSettings[index] as any)[field] = value;
    setSettings(newSettings);
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      await notificationService.updateSettings(settings);
      alert('Settings saved successfully!');
    } catch (err) {
      alert('Failed to save settings');
    } finally {
      setSaving(false);
    }
  };

  const getChannelIcon = (channel: string) => {
    switch (channel) {
      case 'EMAIL': return <span className="material-symbols-outlined text-blue-400">mail</span>;
      case 'SMS': return <span className="material-symbols-outlined text-yellow-400">sms</span>;
      case 'PUSH': return <span className="material-symbols-outlined text-red-400">notifications_active</span>;
      case 'IN_APP': return <span className="material-symbols-outlined text-purple-400">app_registration</span>;
      default: return <span className="material-symbols-outlined">notifications_active</span>;
    }
  };

  if (loading) return (
    <div className="p-8 flex items-center justify-center">
      <div className="w-12 h-12 rounded-full border-4 border-slate-700 border-t-blue-500 animate-spin"></div>
    </div>
  );

  return (
    <div className="p-6 space-y-6 animate-in slide-in-from-right-4 duration-500">
      <div className="flex items-center justify-between">
        <button 
          onClick={() => navigate('..')}
          className="flex items-center gap-2 text-slate-400 hover:text-white transition-all group"
        >
          <span className="material-symbols-outlined group-hover:-translate-x-1 transition-transform">arrow_back</span>
          Settings
        </button>
        <button 
          onClick={handleSave}
          disabled={saving}
          className="px-6 py-2.5 rounded-xl bg-blue-600 hover:bg-blue-500 text-white font-medium transition-all shadow-lg shadow-blue-900/40 flex items-center gap-2 disabled:opacity-50"
        >
          <span className="material-symbols-outlined text-[18px]">save</span>
          {saving ? 'Saving...' : 'Save Configuration'}
        </button>
      </div>

      <div className="flex items-center gap-4 mb-2">
        <div className="p-3 rounded-2xl bg-blue-600/20 border border-blue-500/30">
          <span className="material-symbols-outlined text-blue-400">settings</span>
        </div>
        <div>
          <h1 className="text-2xl font-bold text-white">Channel Orchestration</h1>
          <p className="text-slate-400 text-sm">Configure global delivery rules and persistence policies</p>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {settings.map((s, idx) => (
          <div key={s.id || s.channel} className="p-6 rounded-2xl bg-slate-900/50 border border-slate-800 backdrop-blur-xl space-y-6">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-lg bg-slate-800 border border-slate-700">
                  {getChannelIcon(s.channel)}
                </div>
                <div>
                  <h3 className="font-bold text-slate-100">{s.channel} Delivery</h3>
                  <span className="text-[10px] text-slate-500 uppercase font-bold tracking-widest">Global Status</span>
                </div>
              </div>
              <button 
                onClick={() => handleToggle(idx)}
                className={`transition-all ${s.enabled ? 'text-blue-500' : 'text-slate-600'} flex items-center`}
              >
                <span className="material-symbols-outlined text-[48px]">
                  {s.enabled ? 'toggle_on' : 'toggle_off'}
                </span>
              </button>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <label className="text-xs font-bold text-slate-500 uppercase tracking-wider">Max Retries</label>
                <input 
                  type="number"
                  min="0"
                  max="10"
                  value={s.max_retries}
                  onChange={(e) => handleChange(idx, 'max_retries', parseInt(e.target.value))}
                  className="w-full px-4 py-2 rounded-xl bg-slate-950 border border-slate-800 text-slate-200 outline-none focus:border-blue-500 transition-all"
                />
              </div>
              <div className="space-y-2">
                <label className="text-xs font-bold text-slate-500 uppercase tracking-wider">Priority</label>
                <select 
                  value={s.priority}
                  onChange={(e) => handleChange(idx, 'priority', e.target.value)}
                  className="w-full px-4 py-2 rounded-xl bg-slate-950 border border-slate-800 text-slate-200 outline-none focus:border-blue-500 transition-all"
                >
                  <option value="HIGH">High</option>
                  <option value="STANDARD">Standard</option>
                  <option value="LOW">Low</option>
                </select>
              </div>
              <div className="col-span-2 space-y-2">
                <label className="text-xs font-bold text-slate-500 uppercase tracking-wider">Retention Period (Months)</label>
                <div className="flex items-center gap-4">
                  <input 
                    type="range"
                    min="1"
                    max="24"
                    value={s.retention_months}
                    onChange={(e) => handleChange(idx, 'retention_months', parseInt(e.target.value))}
                    className="flex-1 accent-blue-500"
                  />
                  <span className="text-sm font-mono text-blue-400 w-8">{s.retention_months}</span>
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>

      <div className="p-4 rounded-2xl bg-blue-900/10 border border-blue-500/20 flex gap-4 items-start">
        <span className="material-symbols-outlined text-blue-400 mt-0.5">info</span>
        <div>
          <h4 className="text-sm font-bold text-blue-200">Governance Policy</h4>
          <p className="text-xs text-blue-300/70 mt-1 leading-relaxed">
            Changes to delivery settings are logged in the global audit trail. Disabling a channel will prevent all systemic communications including security alerts and password resets for that specific channel.
          </p>
        </div>
      </div>
    </div>
  );
};

export default NotificationSettings;
