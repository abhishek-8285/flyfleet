import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import SecurityFooter from '../components/SecurityFooter';
import SecuritySidebar from '../components/SecuritySidebar';
import { useAuth } from '../context/AuthProvider';
import api from '../services/api';

const MFAOverrides: React.FC = () => {
    const navigate = useNavigate();
    const { user } = useAuth();
    const [organizations, setOrganizations] = useState<any[]>([]);
    const [overrides, setOverrides] = useState<Record<string, any>>({});
    const [loading, setLoading] = useState(true);
    const [updating, setUpdating] = useState<string | null>(null);

    useEffect(() => {
        const fetchData = async () => {
            try {
                const orgs = await api.getOrganizations();
                setOrganizations(orgs || []);
                
                // Fetch override status for each org (in parallel)
                const overrideResults = await Promise.all(
                    orgs.map((o: any) => api.getMFAOrgOverride(o.id).catch(() => ({ org_id: o.id, enabled: false })))
                );
                
                const overrideMap: Record<string, any> = {};
                overrideResults.forEach((ov: any) => {
                    overrideMap[ov.org_id] = ov;
                });
                setOverrides(overrideMap);
            } catch (err) {
                 console.error("Failed to load ecosystem data", err);
            } finally {
                setLoading(false);
            }
        };
        fetchData();
    }, []);

    const handleToggleOverride = async (orgId: string, currentEnabled: boolean) => {
        setUpdating(orgId);
        try {
            const newStatus = !currentEnabled;
            await api.updateMFAOrgOverride({
                org_id: orgId,
                enabled: newStatus
            });
            setOverrides(prev => ({
                ...prev,
                [orgId]: { ...prev[orgId], enabled: newStatus }
            }));
        } catch (err) {
            console.error("Failed to propagate override", err);
        } finally {
            setUpdating(null);
        }
    };

    return (
        <div className="flex h-screen w-full bg-background-dark text-white overflow-hidden">
            <SecuritySidebar />
            <main className="flex-1 flex flex-col min-w-0 bg-background-dark overflow-hidden relative">
                <div className="absolute top-0 right-0 w-[600px] h-[600px] bg-primary/5 rounded-full blur-[140px] -z-10"></div>
                
                <header className="h-16 flex items-center justify-between px-6 border-b border-white/5 bg-surface-dark/40 backdrop-blur-md flex-shrink-0 z-20">
                     <div className="flex flex-col justify-center">
                        <nav className="flex items-center gap-2 text-[10px] text-text-muted uppercase tracking-widest mb-0.5 font-bold">
                            <span onClick={() => navigate('/security')} className="hover:text-primary cursor-pointer transition-colors">Security Center</span>
                            <span className="material-symbols-outlined text-[10px]">chevron_right</span>
                            <span className="text-white">Organization architecture</span>
                        </nav>
                        <h2 className="text-white text-lg font-bold tracking-tight">Ecosystem MFA Overrides</h2>
                    </div>
                    <div className="flex items-center gap-2 bg-primary/10 px-3 py-1.5 rounded-lg border border-primary/20">
                        <span className="material-symbols-outlined text-[16px] text-primary">hub</span>
                        <span className="text-[10px] font-black uppercase text-text-muted tracking-tighter">Total Entities: {organizations.length}</span>
                    </div>
                </header>

                <div className="flex-1 overflow-y-auto p-8 custom-scrollbar">
                    {loading ? (
                        <div className="flex flex-col items-center justify-center h-full gap-4">
                            <div className="w-10 h-10 border-2 border-primary/20 border-t-primary rounded-full animate-spin"></div>
                            <p className="text-[10px] text-text-muted font-mono tracking-[0.3em]">SYNCHRONIZING TENANT POLICIES...</p>
                        </div>
                    ) : (
                    <div className="max-w-6xl mx-auto space-y-6">
                        <div className="bg-surface-dark/40 backdrop-blur-xl rounded-3xl border border-white/5 shadow-2xl overflow-hidden min-h-[400px]">
                            <div className="overflow-x-auto">
                                <table className="w-full text-left border-collapse">
                                    <thead>
                                        <tr className="border-b border-white/5 text-[10px] text-text-muted uppercase tracking-[0.2em] bg-white/5 font-black">
                                            <th className="px-8 py-4">Organization Identity</th>
                                            <th className="px-8 py-4">Tenant Domain</th>
                                            <th className="px-8 py-4">Policy Enforcement</th>
                                            <th className="px-8 py-4 text-right">Override Status</th>
                                        </tr>
                                    </thead>
                                    <tbody className="text-sm divide-y divide-white/5">
                                        {organizations.map(org => {
                                            const ov = overrides[org.id];
                                            const isEnabled = ov?.enabled || false;
                                            const isBusy = updating === org.id;
                                            
                                            return (
                                                <tr key={org.id} className="hover:bg-white/[0.02] transition-colors group">
                                                    <td className="px-8 py-5">
                                                        <div className="flex items-center gap-3">
                                                            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-white/5 to-white/10 flex items-center justify-center text-white border border-white/10 font-black text-xs group-hover:border-primary/50 transition-colors">
                                                                {org.name?.substring(0, 2).toUpperCase()}
                                                            </div>
                                                            <div>
                                                                <div className="text-sm font-bold text-white group-hover:text-primary transition-colors">{org.name}</div>
                                                                <div className="text-[9px] font-mono text-text-muted mt-0.5 uppercase tracking-tighter">{org.id}</div>
                                                            </div>
                                                        </div>
                                                    </td>
                                                    <td className="px-8 py-5">
                                                        <span className="text-xs text-text-muted font-medium">{org.domain || 'Internal Fleet'}</span>
                                                    </td>
                                                    <td className="px-8 py-5">
                                                        <div className="flex items-center gap-2">
                                                            <div className={`size-1.5 rounded-full ${isEnabled ? 'bg-primary animate-pulse' : 'bg-success/50'}`}></div>
                                                            <span className={`text-[10px] font-bold uppercase tracking-wider ${isEnabled ? 'text-primary' : 'text-text-muted'}`}>
                                                                {isEnabled ? 'Custom Partition' : 'Global Baseline'}
                                                            </span>
                                                        </div>
                                                    </td>
                                                    <td className="px-8 py-5 text-right">
                                                        <div className="flex items-center justify-end gap-4">
                                                            <span className={`text-[10px] font-black uppercase tracking-widest transition-colors ${isEnabled ? 'text-primary' : 'text-text-muted/30'}`}>
                                                                {isEnabled ? 'Override Active' : 'Override Disabled'}
                                                            </span>
                                                            <button 
                                                                onClick={() => handleToggleOverride(org.id, isEnabled)}
                                                                disabled={isBusy}
                                                                className={`relative inline-flex items-center h-5 w-10 rounded-full transition-all duration-200 outline-none ${isEnabled ? 'bg-primary/40 text-primary border border-primary/50' : 'bg-white/5 border border-white/10'}`}
                                                            >
                                                                <span className={`inline-block h-3 w-3 transform rounded-full bg-white transition-transform duration-200 ${isEnabled ? 'translate-x-6 bg-primary shadow-[0_0_8px_rgba(19,127,236,0.6)]' : 'translate-x-1 bg-text-muted/50'}`} />
                                                                {isBusy && (
                                                                    <div className="absolute inset-0 flex items-center justify-center bg-black/20 rounded-full">
                                                                        <div className="size-2 border border-white/30 border-t-white rounded-full animate-spin"></div>
                                                                    </div>
                                                                )}
                                                            </button>
                                                        </div>
                                                    </td>
                                                </tr>
                                            );
                                        })}
                                        {organizations.length === 0 && (
                                            <tr>
                                                <td colSpan={4} className="px-8 py-20 text-center">
                                                    <span className="material-symbols-outlined text-4xl text-white/5 mb-3 block">domain_disabled</span>
                                                    <p className="text-text-muted text-sm font-medium">No system organizations detected.</p>
                                                </td>
                                            </tr>
                                        )}
                                    </tbody>
                                </table>
                            </div>
                        </div>
                    </div>
                    )}
                </div>
                <SecurityFooter />
            </main>
        </div>
    );
};

export default MFAOverrides;
