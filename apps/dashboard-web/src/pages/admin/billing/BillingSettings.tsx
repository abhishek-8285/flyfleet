import {
    Calendar,
    CheckCircle2,
    Globe,
    IndianRupee,
    Save,
    ShieldCheck,
    Zap
} from 'lucide-react';
import React, { useEffect, useState } from 'react';
import billingService from '../../../services/billingService';
import { BillingSettings } from '../../../types/billing';

const BillingSettingsPage: React.FC = () => {
  const [settings, setSettings] = useState<BillingSettings | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [showToast, setShowToast] = useState(false);

  useEffect(() => {
    fetchSettings();
  }, []);

  const fetchSettings = async () => {
    try {
      const data = await billingService.getSettings();
      setSettings(data);
    } catch (error) {
      console.error('Failed to fetch settings', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!settings) return;
    setSaving(true);
    try {
      await billingService.updateSettings(settings);
      setShowToast(true);
      setTimeout(() => setShowToast(false), 3000);
    } catch (error) {
      console.error('Failed to save settings', error);
    } finally {
      setSaving(false);
    }
  };

  if (loading || !settings) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      <div className="flex items-center justify-between">
        <div className="space-y-1">
          <h1 className="text-2xl font-bold text-white">Billing Settings</h1>
          <p className="text-sm text-gray-400">Configure your organization's financial defaults.</p>
        </div>
        <button 
          onClick={handleSave}
          disabled={saving}
          className="flex items-center gap-2 px-6 py-2.5 bg-primary text-white rounded-xl hover:bg-primary/90 transition-all font-bold shadow-lg shadow-primary/20 disabled:opacity-50"
        >
          {saving ? <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div> : <Save className="w-4 h-4" />}
          Save Changes
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="md:col-span-1">
          <h2 className="text-lg font-semibold text-white">General Configuration</h2>
          <p className="text-sm text-gray-500 mt-1">Set your base currency and tax rates applied to new invoices.</p>
        </div>
        <div className="md:col-span-2 bg-gray-900/50 backdrop-blur-xl border border-white/10 rounded-2xl p-6 space-y-6 shadow-xl">
           <div className="grid grid-cols-2 gap-6">
              <div className="space-y-2">
                <label className="text-sm font-medium text-gray-400 flex items-center gap-2">
                  <Globe className="w-4 h-4" /> Base Currency
                </label>
                <select 
                  value={settings.base_currency}
                  onChange={(e) => setSettings({ ...settings, base_currency: e.target.value })}
                  className="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-primary appearance-none cursor-pointer hover:bg-white/10 transition-colors"
                >
                  <option value="INR" className="bg-gray-900">INR - Indian Rupee</option>
                  <option value="USD" className="bg-gray-900">USD - US Dollar</option>
                  <option value="EUR" className="bg-gray-900">EUR - Euro</option>
                </select>
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium text-gray-400 flex items-center gap-2">
                  <IndianRupee className="w-4 h-4" /> Default Tax Rate (%)
                </label>
                <input 
                  type="number"
                  value={settings.default_tax_rate}
                  onChange={(e) => setSettings({ ...settings, default_tax_rate: parseFloat(e.target.value) || 0 })}
                  className="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-primary"
                />
              </div>
           </div>
           <div className="space-y-2">
              <label className="text-sm font-medium text-gray-400 flex items-center gap-2">
                <Calendar className="w-4 h-4" /> Default Grace Period (Days)
              </label>
              <input 
                type="number"
                value={settings.grace_period_days}
                onChange={(e) => setSettings({ ...settings, grace_period_days: parseInt(e.target.value) || 0 })}
                className="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-primary max-w-[200px]"
              />
              <p className="text-xs text-gray-500 italic">This sets the default due date for new invoices.</p>
           </div>
        </div>

        <div className="md:col-span-1 pt-6">
          <h2 className="text-lg font-semibold text-white">Automation</h2>
          <p className="text-sm text-gray-500 mt-1">Enable AI-powered auto invoicing and smart reminders.</p>
        </div>
        <div className="md:col-span-2 bg-gray-900/50 backdrop-blur-xl border border-white/10 rounded-2xl p-6 space-y-6 shadow-xl">
           <div className="flex items-center justify-between p-4 bg-white/5 rounded-xl border border-white/5 group hover:border-primary/30 transition-all cursor-pointer" onClick={() => setSettings({ ...settings, auto_invoicing_enabled: !settings.auto_invoicing_enabled })}>
              <div className="flex items-center gap-4">
                <div className={`p-3 rounded-xl transition-colors ${settings.auto_invoicing_enabled ? 'bg-primary/20 text-primary' : 'bg-white/5 text-gray-500'}`}>
                  <Zap className="w-5 h-5" />
                </div>
                <div>
                  <h4 className="font-bold text-white uppercase tracking-wider text-xs">Auto-Issue Invoices</h4>
                  <p className="text-xs text-gray-500 mt-0.5">Automatically generate invoices based on completed trips.</p>
                </div>
              </div>
              <div className={`w-12 h-6 rounded-full relative transition-colors ${settings.auto_invoicing_enabled ? 'bg-primary' : 'bg-gray-700'}`}>
                <div className={`absolute top-1 w-4 h-4 bg-white rounded-full transition-all ${settings.auto_invoicing_enabled ? 'left-7' : 'left-1'}`}></div>
              </div>
           </div>
           
           <div className="flex items-center justify-between p-4 bg-white/5 rounded-xl border border-white/5 group opacity-50">
              <div className="flex items-center gap-4">
                <div className="p-3 bg-white/5 text-gray-500 rounded-xl">
                  <ShieldCheck className="w-5 h-5" />
                </div>
                <div>
                  <h4 className="font-bold text-white uppercase tracking-wider text-xs">Fraud Prevention</h4>
                  <p className="text-xs text-gray-500 mt-0.5">AI verification for all issued invoices. (Premium Only)</p>
                </div>
              </div>
              <div className="px-2 py-1 bg-yellow-500/20 text-yellow-500 text-[10px] font-black uppercase rounded">Locked</div>
           </div>
        </div>
      </div>

      {/* Success Toast */}
      {showToast && (
        <div className="fixed bottom-8 right-8 z-50 animate-in slide-in-from-right-full duration-300">
          <div className="bg-primary text-white border border-white/20 px-6 py-4 rounded-2xl shadow-2xl flex items-center gap-3">
             <div className="p-1 bg-white/20 rounded-full"><CheckCircle2 className="w-5 h-5" /></div>
             <div>
               <p className="font-bold text-sm tracking-wide">Settings Saved!</p>
               <p className="text-white/70 text-xs">Your financial configurations are now up to date.</p>
             </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default BillingSettingsPage;
