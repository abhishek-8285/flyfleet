import {
    ArrowPathIcon,
    CheckCircleIcon,
    ChevronLeftIcon,
    CloudArrowUpIcon,
    LinkIcon,
    ShieldCheckIcon
} from '@heroicons/react/24/outline';
import { motion } from 'framer-motion';
import React, { useState } from 'react';

interface ConfigFormProps {
  id: string;
  onBack: () => void;
}

const ConfigForm: React.FC<ConfigFormProps> = ({ id, onBack }) => {
  const [isTesting, setIsTesting] = useState(false);
  const [testResult, setTestResult] = useState<{ success: boolean; msg: string } | null>(null);

  const handleTest = async () => {
    setIsTesting(true);
    // Mock API call
    setTimeout(() => {
      setIsTesting(false);
      setTestResult({ success: true, msg: 'Successfully connected to LogisticsPartner API' });
    }, 1500);
  };

  return (
    <div className="max-w-4xl mx-auto">
      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <button 
          onClick={onBack}
          className="flex items-center gap-2 text-slate-400 hover:text-white transition-colors"
        >
          <ChevronLeftIcon className="w-5 h-5" />
          Back to Marketplace
        </button>
        
        <div className="flex gap-3">
          <button 
            onClick={handleTest}
            disabled={isTesting}
            className="flex items-center gap-2 bg-slate-800 hover:bg-slate-700 text-white px-4 py-2 rounded-xl text-sm font-semibold border border-white/5 transition-all disabled:opacity-50"
          >
            {isTesting ? <ArrowPathIcon className="w-5 h-5 animate-spin" /> : <LinkIcon className="w-5 h-5" />}
            Test Connection
          </button>
          <button className="flex items-center gap-2 bg-primary hover:bg-primary-dark text-white px-6 py-2 rounded-xl text-sm font-semibold transition-all shadow-lg shadow-primary/20">
            <CloudArrowUpIcon className="w-5 h-5" />
            Save Configuration
          </button>
        </div>
      </div>

      <div className="space-y-6">
        {/* Info Card */}
        <div className="bg-slate-900/40 border border-slate-800 rounded-2xl p-6 flex items-start gap-5 backdrop-blur-sm">
          <div className="w-16 h-16 rounded-2xl bg-slate-800 border border-slate-700 flex items-center justify-center flex-shrink-0">
             <span className="text-3xl">🚚</span>
          </div>
          <div>
            <h3 className="text-xl font-bold text-white mb-1">LogisticsPartner Fulfillment Sync</h3>
            <p className="text-sm text-slate-400">Configure the connection details for the LogisticsPartner REST API. Sync order statuses and tracking numbers automatically.</p>
            <div className="flex gap-3 mt-4">
              <span className="flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-emerald-500/10 text-emerald-500 border border-emerald-500/20 text-[11px] font-bold uppercase tracking-wider">
                <span className="w-1.5 h-1.5 rounded-full bg-emerald-500" />
                Active
              </span>
              <span className="px-2.5 py-1 rounded-full bg-slate-800 text-slate-400 border border-white/5 text-[11px] font-bold">
                v2.4.0
              </span>
            </div>
          </div>
        </div>

        {/* Form Sections */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <section className="col-span-1 md:col-span-2 space-y-4">
            <div className="flex items-center gap-2 pb-2 border-b border-white/5">
              <LinkIcon className="w-5 h-5 text-primary" />
              <h4 className="text-sm font-bold text-white uppercase tracking-wider">Endpoint Settings</h4>
            </div>
            <div className="bg-slate-900/40 border border-slate-800 rounded-2xl p-6 space-y-4">
              <div>
                <label className="block text-xs font-bold text-slate-500 uppercase tracking-widest mb-2">Base URL</label>
                <input 
                  type="url" 
                  defaultValue="https://api.logisticspartner.io/v2/fleet-sync"
                  className="w-full bg-slate-950/50 border border-slate-800 rounded-xl px-4 py-2.5 text-sm font-mono text-primary focus:ring-2 focus:ring-primary/50 outline-none transition-all"
                />
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs font-bold text-slate-500 uppercase tracking-widest mb-2">Environment</label>
                  <select className="w-full bg-slate-950/50 border border-slate-800 rounded-xl px-4 py-2.5 text-sm text-white focus:ring-2 focus:ring-primary/50 outline-none appearance-none">
                    <option>Production</option>
                    <option>Staging</option>
                    <option>Development</option>
                  </select>
                </div>
                <div>
                  <label className="block text-xs font-bold text-slate-500 uppercase tracking-widest mb-2">Timeout (Sec)</label>
                  <input type="number" defaultValue="30" className="w-full bg-slate-950/50 border border-slate-800 rounded-xl px-4 py-2.5 text-sm text-white focus:ring-2 focus:ring-primary/50 outline-none" />
                </div>
              </div>
            </div>
          </section>

          <section className="col-span-1 md:col-span-2 space-y-4">
            <div className="flex items-center gap-2 pb-2 border-b border-white/5">
              <ShieldCheckIcon className="w-5 h-5 text-primary" />
              <h4 className="text-sm font-bold text-white uppercase tracking-wider">Authentication</h4>
            </div>
            <div className="bg-slate-900/40 border border-slate-800 rounded-2xl p-6 space-y-6">
              <div className="flex gap-6">
                <label className="flex items-center gap-2 cursor-pointer group">
                  <input type="radio" name="auth" defaultChecked className="sr-only peer" />
                  <div className="w-5 h-5 rounded-full border-2 border-slate-700 peer-checked:border-primary peer-checked:bg-primary/20 flex items-center justify-center transition-all">
                    <div className="w-2 h-2 rounded-full bg-primary opacity-0 peer-checked:opacity-100 transition-all" />
                  </div>
                  <span className="text-sm text-slate-400 group-hover:text-white transition-colors font-medium">OAuth 2.0</span>
                </label>
                <label className="flex items-center gap-2 cursor-pointer group">
                  <input type="radio" name="auth" className="sr-only peer" />
                  <div className="w-5 h-5 rounded-full border-2 border-slate-700 peer-checked:border-primary peer-checked:bg-primary/20 flex items-center justify-center transition-all">
                    <div className="w-2 h-2 rounded-full bg-primary opacity-0 peer-checked:opacity-100 transition-all" />
                  </div>
                  <span className="text-sm text-slate-400 group-hover:text-white transition-colors font-medium">API Key</span>
                </label>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs font-bold text-slate-500 uppercase tracking-widest mb-2">Client ID</label>
                  <input type="text" defaultValue="fp_live_839201938" className="w-full bg-slate-950/50 border border-slate-800 rounded-xl px-4 py-2.5 text-sm font-mono text-white" />
                </div>
                <div>
                  <label className="block text-xs font-bold text-slate-500 uppercase tracking-widest mb-2">Client Secret</label>
                  <input type="password" defaultValue="********" className="w-full bg-slate-950/50 border border-slate-800 rounded-xl px-4 py-2.5 text-sm font-mono text-white" />
                </div>
              </div>
            </div>
          </section>
        </div>

        {testResult && (
          <motion.div 
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: 'auto', opacity: 1 }}
            className={`p-4 rounded-xl border flex items-center gap-3 ${
              testResult.success ? 'bg-emerald-500/10 border-emerald-500/20 text-emerald-400' : 'bg-rose-500/10 border-rose-500/20 text-rose-400'
            }`}
          >
            <CheckCircleIcon className="w-5 h-5" />
            <span className="text-sm font-medium">{testResult.msg}</span>
          </motion.div>
        )}
      </div>
    </div>
  );
};

export default ConfigForm;
