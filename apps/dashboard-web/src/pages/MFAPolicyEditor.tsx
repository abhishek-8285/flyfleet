import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import SecurityFooter from '../components/SecurityFooter';
import SecuritySidebar from '../components/SecuritySidebar';
import { useAuth } from '../context/AuthProvider';
import api from '../services/api';

const MFAPolicyEditor: React.FC = () => {
    const navigate = useNavigate();
    const { user } = useAuth();
    const [policy, setPolicy] = useState<any>({
        require_mfa: false,
        allowed_methods: ["totp"],
        max_attempts: 5,
        lockout_minutes: 30,
        grace_period_days: 7,
        remember_device_days: 30,
        version: 1
    });
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [auditReason, setAuditReason] = useState("");
    const [notification, setNotification] = useState<{msg: string, type: 'success'|'error'} | null>(null);

    const orgId = user?.org_id || localStorage.getItem('org_id') || "";

    useEffect(() => {
        if (!orgId) return;
        const fetchPolicy = async () => {
            try {
                const data = await api.getMFAPolicy(orgId);
                setPolicy(data);
            } catch (err) {
                console.error("Failed to load policy", err);
            } finally {
                setLoading(false);
            }
        };
        fetchPolicy();
    }, [orgId]);

    const handleChange = (field: string, value: any) => {
        setPolicy((prev: any) => ({ ...prev, [field]: value }));
    };

    const handleMethodToggle = (method: string) => {
        setPolicy((prev: any) => {
            const current = prev.allowed_methods || [];
            if (current.includes(method)) {
                return { ...prev, allowed_methods: current.filter((m: string) => m !== method) };
            } else {
                return { ...prev, allowed_methods: [...current, method] };
            }
        });
    };

    const handleSave = async () => {
        if (!auditReason) {
             setNotification({msg: "Policy update requires a valid justification for compliance logging.", type: "error"});
             return;
        }
        setSaving(true);
        try {
            await api.updateMFAPolicy({ policy, reason: auditReason });
            setNotification({msg: "MFA Security Policy propagated successfully across all clusters.", type: "success"});
            setAuditReason("");
            const updated = await api.getMFAPolicy(orgId);
            setPolicy(updated);
        } catch (err) {
            setNotification({msg: "Synchronisation conflict: Another admin modified this policy concurrently.", type: "error"});
        } finally {
            setSaving(false);
        }
    };

    return (
        <div className="flex h-screen w-full bg-background-dark text-white overflow-hidden">
            <SecuritySidebar />
            <main className="flex-1 flex flex-col min-w-0 bg-background-dark overflow-hidden relative">
                <div className="absolute top-0 right-0 w-[500px] h-[500px] bg-primary/5 rounded-full blur-[120px] -z-10"></div>
                
                <header className="h-16 flex items-center justify-between px-6 border-b border-white/5 bg-surface-dark/40 backdrop-blur-md flex-shrink-0 z-20">
                    <div className="flex items-center gap-4">
                        <button onClick={() => navigate('/security')} className="p-2 hover:bg-white/5 rounded-full text-text-muted hover:text-white transition-colors">
                            <span className="material-symbols-outlined">arrow_back</span>
                        </button>
                        <div>
                            <h2 className="text-white text-lg font-bold tracking-tight">Security Infrastructure Control</h2>
                            <p className="text-[11px] text-text-muted mt-0.5">Policy Definition & Cryptographic Standards</p>
                        </div>
                    </div>
                    <div className="flex items-center gap-3">
                        <button 
                            onClick={handleSave} 
                            disabled={saving} 
                            className="px-6 py-2 bg-primary hover:bg-primary-dark text-white text-xs font-bold rounded-lg flex items-center gap-2 transition-all shadow-lg shadow-primary/25 disabled:opacity-50 disabled:cursor-not-allowed border border-white/10"
                        >
                            {saving ? <span className="material-symbols-outlined animate-spin text-[18px]">sync</span> : <span className="material-symbols-outlined text-[18px]">shield_lock</span>}
                            Commit Changes
                        </button>
                    </div>
                </header>
                
                {notification && (
                    <div className={`absolute top-20 right-6 z-50 px-5 py-3.5 rounded-xl shadow-2xl border flex items-center gap-3 animate-in fade-in slide-in-from-top-4 duration-300 ${notification.type === 'success' ? 'bg-success/10 border-success/30 text-success' : 'bg-danger/10 border-danger/30 text-danger'}`}>
                        <span className="material-symbols-outlined">{notification.type === 'success' ? 'verified' : 'gpp_maybe'}</span>
                        <span className="text-xs font-bold tracking-wide uppercase">{notification.msg}</span>
                        <button onClick={() => setNotification(null)} className="ml-4 opacity-50 hover:opacity-100 transition-opacity"><span className="material-symbols-outlined text-sm">close</span></button>
                    </div>
                )}

                <div className="flex-1 overflow-y-auto p-8 custom-scrollbar">
                    {loading ? (
                         <div className="flex flex-col items-center justify-center h-full gap-4">
                            <div className="w-10 h-10 border-2 border-primary/20 border-t-primary rounded-full animate-spin"></div>
                            <p className="text-xs text-text-muted font-mono">RETRIEVING ENCRYPTION SCHEMAS...</p>
                         </div>
                    ) : (
                    <div className="max-w-4xl mx-auto space-y-8">
                        <section className="bg-surface-dark/40 backdrop-blur-xl rounded-3xl border border-white/5 overflow-hidden shadow-2xl">
                            <div className="px-8 py-5 border-b border-white/5 bg-white/5 flex items-center gap-3">
                                <div className="w-8 h-8 rounded-lg bg-primary/20 flex items-center justify-center text-primary">
                                    <span className="material-symbols-outlined text-[20px]">admin_panel_settings</span>
                                </div>
                                <div>
                                    <h3 className="text-white font-bold text-sm tracking-wide">Primary MFA Directives</h3>
                                    <p className="text-[10px] text-text-muted">Configuration Revision: v{policy.version}</p>
                                </div>
                            </div>
                            
                            <div className="p-8 grid grid-cols-1 md:grid-cols-2 gap-12">
                                <div className="space-y-8">
                                    <div className="flex items-center justify-between p-4 bg-white/[0.02] rounded-2xl border border-white/5">
                                        <div>
                                            <label className="text-sm font-bold text-white block">Global Require MFA</label>
                                            <p className="text-[11px] text-text-muted mt-1 leading-relaxed">Force all users to configure a second factor before accessing the command center.</p>
                                        </div>
                                        <label className="relative inline-flex items-center cursor-pointer">
                                            <input type="checkbox" className="sr-only peer" checked={policy.require_mfa} onChange={(e) => handleChange('require_mfa', e.target.checked)}/>
                                            <div className="w-12 h-6 bg-white/5 border border-white/10 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-text-muted after:border-gray-500 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary/40 peer-checked:after:bg-primary peer-checked:border-primary/50"></div>
                                        </label>
                                    </div>

                                    <div>
                                        <label className="text-xs font-bold text-text-muted uppercase tracking-widest block mb-4">Cryptographic Methods</label>
                                        <div className="grid grid-cols-1 gap-3">
                                            {[
                                                {id: 'totp', name: 'Authenticator App', desc: 'Secure time-based codes', reco: true},
                                                {id: 'sms', name: 'SMS Delivery', desc: 'Legacy phone-based codes', reco: false},
                                                {id: 'fido2', name: 'Hardware Security Keys', desc: 'FIDO2 / WebAuthn standard', reco: true}
                                            ].map(method => (
                                                <div 
                                                    key={method.id} 
                                                    onClick={() => handleMethodToggle(method.id)}
                                                    className={`p-4 rounded-2xl border transition-all cursor-pointer flex items-center gap-4 ${policy.allowed_methods?.includes(method.id) ? 'bg-primary/5 border-primary/30' : 'bg-white/[0.02] border-white/5 opacity-60 hover:opacity-100 hover:bg-white/[0.04]'}`}
                                                >
                                                    <div className={`w-5 h-5 rounded border flex items-center justify-center transition-all ${policy.allowed_methods?.includes(method.id) ? 'bg-primary border-primary text-white' : 'bg-transparent border-white/20'}`}>
                                                        {policy.allowed_methods?.includes(method.id) && <span className="material-symbols-outlined text-[14px] font-bold">check</span>}
                                                    </div>
                                                    <div>
                                                        <div className="text-xs font-bold text-white flex items-center gap-2">
                                                            {method.name}
                                                            {method.reco && <span className="text-[9px] bg-success/10 text-success border border-success/20 px-1.5 py-0.5 rounded-full">OPTIMAL</span>}
                                                        </div>
                                                        <p className="text-[10px] text-text-muted mt-0.5">{method.desc}</p>
                                                    </div>
                                                </div>
                                            ))}
                                        </div>
                                    </div>
                                </div>

                                <div className="space-y-8">
                                    <div className="grid grid-cols-1 gap-5">
                                        {[
                                            {id: 'max_attempts', label: 'Maximum Authentication Failures', icon: 'lock_open'},
                                            {id: 'lockout_minutes', label: 'Lockout Duration (Minutes)', icon: 'timer'},
                                            {id: 'grace_period_days', label: 'Enforcement Grace Period (Days)', icon: 'event_repeat'}
                                        ].map(field => (
                                            <div key={field.id} className="space-y-2">
                                                <label className="text-xs font-bold text-text-muted hover:text-white transition-colors flex items-center gap-2 uppercase tracking-wide">
                                                    <span className="material-symbols-outlined text-[16px]">{field.icon}</span>
                                                    {field.label}
                                                </label>
                                                <input 
                                                    className="w-full bg-white/[0.03] border border-white/10 text-white text-sm font-mono rounded-xl focus:ring-2 focus:ring-primary/40 focus:border-primary/50 block p-3 px-4 transition-all outline-none" 
                                                    type="number" 
                                                    value={policy[field.id]} 
                                                    onChange={(e) => handleChange(field.id, parseInt(e.target.value))}
                                                />
                                            </div>
                                        ))}
                                    </div>

                                    <div className="relative group">
                                         <label className="text-xs font-bold text-primary mb-3 flex items-center gap-2 uppercase tracking-widest">
                                            <span className="material-symbols-outlined text-[16px]">history_edu</span>
                                            Modification Narrative <span className="text-danger">*</span>
                                         </label>
                                         <textarea 
                                            value={auditReason} 
                                            onChange={(e) => setAuditReason(e.target.value)}
                                            className="w-full bg-white/[0.03] border border-white/10 text-white text-sm rounded-2xl focus:ring-2 focus:ring-primary/40 focus:border-primary/50 block p-4 min-h-[120px] transition-all outline-none resize-none placeholder:text-text-muted/30"
                                            placeholder="State the technical or business requirement driving this change for the permanent audit registry..."
                                         ></textarea>
                                         <div className="absolute bottom-3 right-3 text-[9px] font-mono text-text-muted group-hover:text-primary transition-colors">COMPLIANCE REQUIREMENT</div>
                                    </div>
                                </div>
                            </div>
                            
                            <div className="px-8 py-5 bg-danger/5 border-t border-white/5 flex items-start gap-4">
                                <span className="material-symbols-outlined text-danger shrink-0 animate-pulse">crisis_alert</span>
                                <div>
                                    <h3 className="text-danger font-bold text-[11px] uppercase tracking-tighter mb-0.5">Critical System Warning</h3>
                                    <p className="text-[10px] text-text-muted leading-relaxed">Propagation of these security directives will affect all live authentication sessions across the fleet infrastructure instantly.</p>
                                </div>
                            </div>
                        </section>
                    </div>
                    )}
                </div>
                <SecurityFooter />
            </main>
        </div>
    );
};

export default MFAPolicyEditor;
