import { CheckCircleIcon, Cog6ToothIcon, FolderIcon } from '@heroicons/react/24/outline';
import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import api from '../../../services/api';

interface Settings {
  retention_period: number;
  policies: Record<string, any>;
}

export default function TripSettings() {
  const [settings, setSettings] = useState<Settings>({ retention_period: 36, policies: {} });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    loadSettings();
  }, []);

  const loadSettings = async () => {
    try {
      const data = await api.getTripSettings();
      setSettings(data);
    } catch (error) {
      console.error('Failed to load settings:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    setSaving(true);
    setSaved(false);
    try {
      await api.updateTripSettings(settings);
      setSaved(true);
      setTimeout(() => setSaved(false), 3000);
    } catch (error) {
      console.error('Failed to save settings:', error);
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return <div className="p-6 text-center text-text-muted">Loading...</div>;
  }

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <div className="flex items-center gap-2 text-xs text-text-muted mb-1">
            <Link to="/admin/trips" className="hover:text-white">Trips</Link>
            <span>→</span>
            <span className="text-white">Settings</span>
          </div>
          <h1 className="text-2xl font-bold text-white">Trip Configuration</h1>
        </div>
        <button
          onClick={handleSave}
          disabled={saving}
          className="px-4 py-2 bg-primary hover:bg-primary-dark text-white text-sm font-medium rounded flex items-center gap-2 transition-colors disabled:opacity-50"
        >
          {saved ? <CheckCircleIcon className="w-4 h-4" /> : <Cog6ToothIcon className="w-4 h-4" />}
          {saving ? 'Saving...' : saved ? 'Saved!' : 'Save Changes'}
        </button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-4 gap-4">
        <div className="bg-surface-dark rounded-xl border border-border-color p-4">
          <div className="text-text-muted text-xs uppercase mb-1">Total Active Policies</div>
          <div className="text-2xl font-bold text-white">{Object.keys(settings.policies).length || 1}</div>
        </div>
        <div className="bg-surface-dark rounded-xl border border-border-color p-4">
          <div className="text-text-muted text-xs uppercase mb-1">Data Retention</div>
          <div className="text-2xl font-bold text-white">{settings.retention_period} <span className="text-sm font-normal text-text-muted">months</span></div>
        </div>
        <div className="bg-surface-dark rounded-xl border border-border-color p-4">
          <div className="text-text-muted text-xs uppercase mb-1">Compliance Status</div>
          <div className="text-2xl font-bold text-success">Compliant</div>
        </div>
        <div className="bg-surface-dark rounded-xl border border-border-color p-4">
          <div className="text-text-muted text-xs uppercase mb-1">Last Updated</div>
          <div className="text-lg font-medium text-white">Just now</div>
        </div>
      </div>

      {/* Settings Form */}
      <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
        <div className="p-4 border-b border-border-color bg-surface-darker flex items-center gap-3">
          <FolderIcon className="w-5 h-5 text-text-muted" />
          <h3 className="text-white font-semibold">Default Data Retention</h3>
        </div>
        <div className="p-6">
          <div className="grid grid-cols-2 gap-6">
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-white">Retention Period</label>
              <select
                value={settings.retention_period}
                onChange={(e) => setSettings({ ...settings, retention_period: parseInt(e.target.value) })}
                className="w-full bg-surface-darker text-sm text-white border border-border-color rounded-lg px-3 py-2.5 focus:ring-1 focus:ring-primary outline-none"
              >
                <option value={12}>12 Months (1 Year)</option>
                <option value={24}>24 Months (2 Years)</option>
                <option value={36}>36 Months (3 Years) - Recommended</option>
                <option value={60}>60 Months (5 Years)</option>
                <option value={84}>84 Months (7 Years)</option>
              </select>
              <p className="text-xs text-text-muted">Trip data will be archived after this period</p>
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-white">Auto-Archive</label>
              <select
                className="w-full bg-surface-darker text-sm text-white border border-border-color rounded-lg px-3 py-2.5 focus:ring-1 focus:ring-primary outline-none"
                defaultValue="enabled"
              >
                <option value="enabled">Enabled</option>
                <option value="disabled">Disabled</option>
              </select>
              <p className="text-xs text-text-muted">Automatically archive old trip data</p>
            </div>
          </div>
        </div>
      </div>

      {/* Telemetry Settings */}
      <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
        <div className="p-4 border-b border-border-color bg-surface-darker">
          <h3 className="text-white font-semibold">Telemetry Configuration</h3>
        </div>
        <div className="p-6">
          <div className="grid grid-cols-2 gap-6">
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-white">Update Frequency</label>
              <select
                className="w-full bg-surface-darker text-sm text-white border border-border-color rounded-lg px-3 py-2.5 focus:ring-1 focus:ring-primary outline-none"
                defaultValue="10"
              >
                <option value="5">Every 5 seconds</option>
                <option value="10">Every 10 seconds</option>
                <option value="30">Every 30 seconds</option>
                <option value="60">Every 60 seconds</option>
              </select>
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-white">Max Events per Trip</label>
              <input
                type="number"
                defaultValue={10000}
                className="w-full bg-surface-darker text-sm text-white border border-border-color rounded-lg px-3 py-2.5 focus:ring-1 focus:ring-primary outline-none"
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
