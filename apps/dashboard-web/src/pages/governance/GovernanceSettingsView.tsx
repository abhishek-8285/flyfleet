import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import GovernanceService from '../../services/GovernanceService';
import { GovernanceSettings } from '../../types/governance';

const GovernanceSettingsView: React.FC = () => {
  const navigate = useNavigate();
  const [settings, setSettings] = useState<GovernanceSettings | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    const fetchSettings = async () => {
      try {
        const resp = await GovernanceService.getSettings();
        setSettings(resp);
      } catch (error) {
        console.error('Failed to fetch settings:', error);
      } finally {
        setLoading(false);
      }
    };
    fetchSettings();
  }, []);

  const handleSave = async () => {
    if (!settings) return;
    setSaving(true);
    try {
      await GovernanceService.updateSettings(settings);
      // Optional: show snackbar
    } catch (error) {
      console.error('Failed to save settings:', error);
    } finally {
      setSaving(false);
    }
  };

  if (loading) return <div className="p-8 text-white">Loading settings...</div>;

  return (
    <div className="flex-1 flex flex-col min-w-0 bg-background-dark overflow-hidden h-full">
      <div className="w-full bg-surface-darker border-b border-border-color px-6 py-3 flex items-center justify-between gap-4 flex-shrink-0 z-10">
        <div className="flex items-center gap-3">
          <button 
            onClick={() => navigate('/admin/governance')}
            className="p-1.5 text-text-muted hover:text-white hover:bg-surface-dark rounded transition-colors"
          >
            <span className="material-symbols-outlined text-[20px]">arrow_back</span>
          </button>
          <div className="text-lg font-semibold text-white tracking-tight">Compliance & Governance Settings</div>
        </div>
        <button 
          onClick={handleSave}
          disabled={saving}
          className="bg-primary hover:bg-primary/90 text-white text-xs font-semibold py-1.5 px-6 rounded-full shadow-[0_4px_14px_rgba(19,127,236,0.39)] transition-all disabled:opacity-50"
        >
          {saving ? 'Saving...' : 'Save Settings'}
        </button>
      </div>

      <div className="flex-1 overflow-y-auto p-6 bg-[radial-gradient(circle_at_top_right,rgba(19,127,236,0.05),transparent)]">
        <div className="max-w-3xl mx-auto space-y-6">
          <div className="bg-surface-dark/50 backdrop-blur-md border border-border-color rounded-2xl p-8 shadow-xl space-y-8">
            <div className="space-y-6">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-sm font-bold text-white uppercase tracking-tight">Auto-Expiry of Downloads</h3>
                  <p className="text-xs text-text-muted mt-1">Automatically revoke access to exported data after the TTL expires.</p>
                </div>
                <button 
                  onClick={() => setSettings(s => s ? { ...s, auto_expiry_enabled: !s.auto_expiry_enabled } : null)}
                  className={`w-12 h-6 rounded-full p-1 transition-colors ${settings?.auto_expiry_enabled ? 'bg-primary' : 'bg-surface-darker'}`}
                >
                  <div className={`w-4 h-4 bg-white rounded-full transition-transform ${settings?.auto_expiry_enabled ? 'translate-x-6' : 'translate-x-0'}`}></div>
                </button>
              </div>

              <div className="space-y-2">
                <label className="text-[10px] font-bold text-text-muted uppercase tracking-widest">Download Link TTL (Hours)</label>
                <input 
                  type="number" 
                  value={settings?.download_ttl_hours || 48}
                  onChange={(e) => setSettings(s => s ? { ...s, download_ttl_hours: parseInt(e.target.value) } : null)}
                  className="w-full bg-background-dark border border-border-color text-white rounded-xl px-4 py-3 outline-none focus:border-primary transition-colors text-sm"
                />
              </div>
            </div>

            <div className="border-t border-border-color pt-8 flex items-start gap-4">
              <span className="material-symbols-outlined text-yellow-500">warning</span>
              <div className="flex-1">
                <h4 className="text-sm font-bold text-white uppercase tracking-tight">Retention Policies</h4>
                <p className="text-xs text-text-muted mt-1">Data retention policies are currently managed via the Global Policy Engine. Click the link below to configure advanced rules.</p>
                <button className="mt-4 text-primary text-xs font-semibold hover:underline">Open Policy Engine</button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default GovernanceSettingsView;
