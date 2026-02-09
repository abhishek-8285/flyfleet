import {
    AdjustmentsHorizontalIcon,
    ArrowPathIcon,
    BellAlertIcon,
    CheckBadgeIcon,
    LifebuoyIcon,
    ShieldCheckIcon
} from '@heroicons/react/24/outline';
import React, { useEffect, useState } from 'react';
import { useAuth } from '../../context/AuthProvider';
import { api } from '../../services/api';

const AuditPolicy: React.FC = () => {
    const { user } = useAuth();
    const [policy, setPolicy] = useState<any>(null);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);

    const fetchPolicy = async () => {
        if (!user?.org_id) return;
        setLoading(true);
        try {
            const data = await api.getAuditPolicy(user.org_id);
            setPolicy(data || {
                retention_days: 365,
                log_level: 'info',
                auto_export_enabled: false,
                auto_export_frequency: 'monthly',
                encryption_enabled: true,
                pii_masking_enabled: true
            });
        } catch (error) {
            console.error("Failed to fetch policy", error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchPolicy();
    }, [user?.org_id]);

    const handleSave = async () => {
        if (!user?.org_id) return;
        setSaving(true);
        try {
            await api.updateAuditPolicy({
                org_id: user.org_id,
                ...policy
            });
            // Show success toast here if available
        } catch (error) {
            console.error("Failed to save policy", error);
        } finally {
            setSaving(false);
        }
    };

    if (loading) return (
        <div className="flex items-center justify-center py-20">
            <ArrowPathIcon className="w-8 h-8 text-blue-500 animate-spin" />
        </div>
    );

    return (
        <div className="max-w-4xl mx-auto space-y-8 animate-in fade-in duration-700">
            {/* High Level Metrics */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div className="bg-slate-900/40 p-6 rounded-3xl border border-white/5 space-y-4 shadow-xl">
                    <div className="flex items-center justify-between">
                        <ShieldCheckIcon className="w-8 h-8 text-green-400" />
                        <span className="px-2 py-0.5 bg-green-400/10 text-green-400 text-[10px] font-bold rounded uppercase">Active</span>
                    </div>
                    <div>
                        <p className="text-2xl font-semibold text-white">Immutable Hub</p>
                        <p className="text-sm text-slate-500">Every mutation is cryptographically signed and hashed.</p>
                    </div>
                </div>
                <div className="bg-slate-900/40 p-6 rounded-3xl border border-white/5 space-y-4 shadow-xl">
                    <div className="flex items-center justify-between">
                        <LifebuoyIcon className="w-8 h-8 text-blue-400" />
                        <span className="px-2 py-0.5 bg-blue-400/10 text-blue-400 text-[10px] font-bold rounded uppercase">Compliance</span>
                    </div>
                    <div>
                        <p className="text-2xl font-semibold text-white">SOC2 Type II</p>
                        <p className="text-sm text-slate-500">Alignment with AICPA Trust Services Criteria.</p>
                    </div>
                </div>
            </div>

            {/* Policy Controls */}
            <div className="bg-slate-900/60 backdrop-blur-xl rounded-3xl border border-white/5 shadow-2xl overflow-hidden">
                <div className="p-8 border-b border-white/5 flex items-center justify-between bg-slate-950/20">
                    <div className="flex items-center gap-3">
                        <div className="p-2 bg-purple-500/20 rounded-xl">
                            <AdjustmentsHorizontalIcon className="w-6 h-6 text-purple-400" />
                        </div>
                        <h2 className="text-xl font-semibold text-white font-display">Governance Parameters</h2>
                    </div>
                    <button 
                        onClick={handleSave}
                        disabled={saving}
                        className="px-6 py-2.5 bg-blue-600 hover:bg-blue-500 text-white rounded-xl text-sm font-semibold transition-all shadow-lg shadow-blue-500/20 disabled:opacity-50 flex items-center gap-2"
                    >
                        {saving && <ArrowPathIcon className="w-4 h-4 animate-spin" />}
                        Commit Changes
                    </button>
                </div>

                <div className="p-8 space-y-10">
                    {/* Retention Section */}
                    <section className="space-y-6">
                        <div className="flex items-center gap-2">
                            <h3 className="text-xs font-bold text-slate-500 uppercase tracking-widest italic">Data Lifecycle</h3>
                            <div className="h-[1px] flex-1 bg-slate-800" />
                        </div>
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                            <div className="space-y-4">
                                <div className="flex items-center justify-between">
                                    <label className="text-sm text-slate-300">Retention window (Days)</label>
                                    <span className="text-xs font-mono text-blue-400">{policy.retention_days} days</span>
                                </div>
                                <input 
                                    type="range" 
                                    min="30" 
                                    max="3650" 
                                    step="30"
                                    value={policy.retention_days}
                                    onChange={(e) => setPolicy({...policy, retention_days: parseInt(e.target.value)})}
                                    className="w-full h-1.5 bg-slate-800 rounded-lg appearance-none cursor-pointer accent-blue-500"
                                />
                                <div className="flex justify-between text-[10px] text-slate-600">
                                    <span>30d (Dev)</span>
                                    <span>3650d (HCP)</span>
                                </div>
                            </div>
                            <div className="space-y-2">
                                <label className="text-sm text-slate-300">System Verbosity</label>
                                <select 
                                    value={policy.log_level}
                                    onChange={(e) => setPolicy({...policy, log_level: e.target.value})}
                                    className="w-full px-4 py-3 bg-slate-950 border border-slate-800 rounded-2xl text-white text-sm focus:ring-2 focus:ring-blue-500/50 transition-all font-light"
                                >
                                    <option value="minimal">Minimal (Core Security Only)</option>
                                    <option value="info">Info (Default Operational)</option>
                                    <option value="verbose">Verbose (Deep Telemetry)</option>
                                    <option value="debug">Debug (Total Transparency)</option>
                                </select>
                            </div>
                        </div>
                    </section>

                    {/* Security Switches */}
                    <section className="space-y-6">
                        <div className="flex items-center gap-2">
                            <h3 className="text-xs font-bold text-slate-500 uppercase tracking-widest italic">Anonymization & Hardening</h3>
                            <div className="h-[1px] flex-1 bg-slate-800" />
                        </div>
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            <button 
                                onClick={() => setPolicy({...policy, encryption_enabled: !policy.encryption_enabled})}
                                className={`flex items-center justify-between p-4 rounded-2xl border transition-all ${policy.encryption_enabled ? 'bg-blue-600/10 border-blue-500/30' : 'bg-slate-950/50 border-slate-800 opacity-60'}`}
                            >
                                <div className="flex items-center gap-4">
                                    <ShieldCheckIcon className={`w-6 h-6 ${policy.encryption_enabled ? 'text-blue-400' : 'text-slate-500'}`} />
                                    <div className="text-left">
                                        <p className="text-sm font-medium text-white">Full-Disk Encryption</p>
                                        <p className="text-[10px] text-slate-500">AES-256 for all stored logs</p>
                                    </div>
                                </div>
                                {policy.encryption_enabled && <CheckBadgeIcon className="w-5 h-5 text-blue-400" />}
                            </button>

                            <button 
                                onClick={() => setPolicy({...policy, pii_masking_enabled: !policy.pii_masking_enabled})}
                                className={`flex items-center justify-between p-4 rounded-2xl border transition-all ${policy.pii_masking_enabled ? 'bg-blue-600/10 border-blue-500/30' : 'bg-slate-950/50 border-slate-800 opacity-60'}`}
                            >
                                <div className="flex items-center gap-4">
                                    <EyeIcon className={`w-6 h-6 ${policy.pii_masking_enabled ? 'text-blue-400' : 'text-slate-500'}`} />
                                    <div className="text-left">
                                        <p className="text-sm font-medium text-white">Dynamic PII Masking</p>
                                        <p className="text-[10px] text-slate-500">Protect user privacy automatically</p>
                                    </div>
                                </div>
                                {policy.pii_masking_enabled && <CheckBadgeIcon className="w-5 h-5 text-blue-400" />}
                            </button>
                        </div>
                    </section>

                    {/* Automation */}
                    <section className="space-y-6">
                        <div className="flex items-center gap-2">
                            <h3 className="text-xs font-bold text-slate-500 uppercase tracking-widest italic">Automated Evidence Pipeline</h3>
                            <div className="h-[1px] flex-1 bg-slate-800" />
                        </div>
                        <div className="p-6 bg-slate-950/50 border border-white/5 rounded-3xl space-y-6">
                            <div className="flex items-center justify-between">
                                <div className="flex items-center gap-4">
                                    <div className="p-3 bg-amber-500/20 rounded-2xl">
                                        <BellAlertIcon className="w-6 h-6 text-amber-400" />
                                    </div>
                                    <div>
                                        <p className="text-sm font-medium text-white">Periodic Archive Export</p>
                                        <p className="text-xs text-slate-500">Automatically deliver signed archives to your compliance bucket.</p>
                                    </div>
                                </div>
                                <div 
                                    onClick={() => setPolicy({...policy, auto_export_enabled: !policy.auto_export_enabled})}
                                    className={`w-12 h-6 rounded-full cursor-pointer transition-all relative ${policy.auto_export_enabled ? 'bg-blue-600' : 'bg-slate-800'}`}
                                >
                                    <div className={`absolute top-1 w-4 h-4 rounded-full bg-white transition-all ${policy.auto_export_enabled ? 'left-7' : 'left-1'}`} />
                                </div>
                            </div>
                            
                            {policy.auto_export_enabled && (
                                <div className="grid grid-cols-2 gap-4 animate-in slide-in-from-top-2">
                                    <div className="space-y-2">
                                        <label className="text-xs text-slate-500">Frequency</label>
                                        <select 
                                            value={policy.auto_export_frequency}
                                            onChange={(e) => setPolicy({...policy, auto_export_frequency: e.target.value})}
                                            className="w-full px-4 py-2.5 bg-slate-950 border border-slate-800 rounded-xl text-white text-sm"
                                        >
                                            <option value="weekly">Every 7 Days</option>
                                            <option value="monthly">Every 30 Days</option>
                                            <option value="quarterly">Every 90 Days</option>
                                        </select>
                                    </div>
                                    <div className="space-y-2">
                                        <label className="text-xs text-slate-500">Destination Engine</label>
                                        <input type="text" placeholder="s3://archives-v1-..." className="w-full px-4 py-2.5 bg-slate-950 border border-slate-800 rounded-xl text-white text-sm placeholder:text-slate-700 font-mono" />
                                    </div>
                                </div>
                            )}
                        </div>
                    </section>
                </div>
            </div>
        </div>
    );
};

// Helper components missing from Heroicons import in this snippet
const EyeIcon = ({ className }: { className: string }) => (
    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className={className}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M2.036 12.322a1.012 1.012 0 010-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178z" />
        <path strokeLinecap="round" strokeLinejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
    </svg>
);

export default AuditPolicy;
