import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';

interface Backup {
    id: string;
    backup_type: string;
    status: string;
    size_bytes?: number;
    created_at: string;
    location?: string;
    checksum?: string;
}

const RestorePointsPage: React.FC = () => {
    const navigate = useNavigate();
    const [backups, setBackups] = useState<Backup[]>([]);
    const [loading, setLoading] = useState(true);

    const fetchBackups = async () => {
        try {
            const res = await fetch('/api/v1/admin/dr/backups');
            const data = await res.json();
            setBackups(data.data || []);
        } catch (err) {
            console.error(err);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchBackups();
    }, []);

    const formatSize = (bytes?: number) => {
        if (!bytes) return '--';
        if (bytes >= 1e12) return (bytes / 1e12).toFixed(2) + ' TB';
        if (bytes >= 1e9) return (bytes / 1e9).toFixed(2) + ' GB';
        return (bytes / 1e6).toFixed(2) + ' MB';
    };

    const getStatusConfig = (status: string) => {
        switch (status) {
            case 'COMPLETED': return { color: 'success', icon: 'cloud_done', label: 'Verified' };
            case 'RUNNING': return { color: 'primary', icon: 'cloud_sync', label: 'Available', pulse: true };
            case 'FAILED': return { color: 'danger', icon: 'cloud_off', label: 'Corrupted' };
            default: return { color: 'warning', icon: 'hourglass_empty', label: status };
        }
    };

    const stats = [
        { label: 'Total Storage', value: formatSize(backups.reduce((acc, b) => acc + (b.size_bytes || 0), 0)), icon: 'storage', color: 'primary' },
        { label: 'Verified Points', value: backups.filter(b => b.status === 'COMPLETED').length.toString(), icon: 'verified', color: 'success' },
        { label: 'Retention Policy', value: '90 Days', icon: 'history', color: 'warning' },
        { label: 'Failed Backups', value: backups.filter(b => b.status === 'FAILED').length.toString(), icon: 'report', color: 'danger' }
    ];

    return (
        <div className="flex flex-col h-full w-full bg-background-dark text-white font-display overflow-hidden">
            <header className="h-16 flex items-center justify-between px-6 border-b border-border-color bg-surface-darker flex-shrink-0 z-20">
                <div className="flex flex-col">
                    <nav className="flex items-center gap-2 text-[10px] font-bold uppercase tracking-widest text-text-muted mb-0.5">
                        <span>Disaster Recovery</span>
                        <span className="material-symbols-outlined text-[12px]">chevron_right</span>
                        <span className="text-primary">Restore Points</span>
                    </nav>
                    <h2 className="text-white text-lg font-bold">Manage Restore Points</h2>
                </div>
                <div className="flex items-center gap-4">
                    <div className="relative group">
                        <span className="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-text-muted text-[18px]">search</span>
                        <input className="bg-surface-dark border border-border-color text-xs text-white rounded-lg pl-10 pr-4 py-2 w-64 focus:ring-1 focus:ring-primary focus:border-primary outline-none transition-all" placeholder="Global search backups..." type="text"/>
                    </div>
                    <button onClick={() => navigate('/admin/dr/restores/initiate')} className="px-4 py-2 bg-primary hover:bg-primary-dark text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors shadow-lg shadow-primary/20">
                        <span className="material-symbols-outlined text-[18px]">backup</span>
                        Initiate New Restore
                    </button>
                </div>
            </header>

            <div className="flex-1 overflow-hidden p-6 flex flex-col gap-6">
                {/* Stats Grid */}
                <div className="grid grid-cols-4 gap-4">
                    {stats.map((stat, i) => (
                        <div key={i} className="bg-surface-dark border border-border-color p-4 rounded-xl flex items-center gap-4">
                            <div className={`w-10 h-10 rounded-lg bg-${stat.color}/10 text-${stat.color} flex items-center justify-center`}>
                                <span className="material-symbols-outlined">{stat.icon}</span>
                            </div>
                            <div>
                                <div className="text-[10px] text-text-muted font-bold uppercase">{stat.label}</div>
                                <div className="text-lg font-bold text-white">{stat.value}</div>
                            </div>
                        </div>
                    ))}
                </div>

                {/* Table */}
                <div className="flex-1 bg-surface-dark rounded-xl border border-border-color flex flex-col overflow-hidden">
                    <div className="p-4 border-b border-border-color flex justify-between items-center bg-surface-darker/50">
                        <div className="flex items-center gap-4">
                            <h3 className="text-white font-semibold text-sm">Restore Points Inventory</h3>
                            <div className="h-4 w-[1px] bg-border-color"></div>
                            <div className="flex items-center gap-1 text-[11px] text-text-muted">
                                <span className="w-2 h-2 rounded-full bg-success"></span>
                                System Health: Optimal
                            </div>
                        </div>
                        <div className="flex items-center gap-2">
                            <button onClick={fetchBackups} className="flex items-center gap-2 px-3 py-1.5 bg-surface-dark border border-border-color rounded text-xs text-text-muted hover:text-white hover:border-text-muted transition-all">
                                <span className="material-symbols-outlined text-[16px]">refresh</span>
                                Refresh
                            </button>
                            <button className="flex items-center gap-2 px-3 py-1.5 bg-surface-dark border border-border-color rounded text-xs text-text-muted hover:text-white hover:border-text-muted transition-all">
                                <span className="material-symbols-outlined text-[16px]">filter_alt</span>
                                Filter
                            </button>
                            <button className="flex items-center gap-2 px-3 py-1.5 bg-surface-dark border border-border-color rounded text-xs text-text-muted hover:text-white hover:border-text-muted transition-all">
                                <span className="material-symbols-outlined text-[16px]">download</span>
                                Export
                            </button>
                        </div>
                    </div>
                    <div className="flex-1 overflow-auto custom-scrollbar">
                        <table className="w-full text-left border-collapse min-w-[1000px]">
                            <thead>
                                <tr className="border-b border-border-color text-[11px] text-text-muted uppercase tracking-wider bg-surface-darker/30 sticky top-0 z-10 backdrop-blur-md">
                                    <th className="px-6 py-4 font-semibold">Backup ID <span className="material-symbols-outlined text-[14px]">arrow_downward</span></th>
                                    <th className="px-6 py-4 font-semibold">Type</th>
                                    <th className="px-6 py-4 font-semibold">Created At</th>
                                    <th className="px-6 py-4 font-semibold text-right">Size</th>
                                    <th className="px-6 py-4 font-semibold text-center">Status</th>
                                    <th className="px-6 py-4 font-semibold text-right">Actions</th>
                                </tr>
                            </thead>
                            <tbody className="text-sm divide-y divide-border-color">
                                {loading ? (
                                    <tr><td colSpan={6} className="px-6 py-8 text-center text-text-muted">Loading...</td></tr>
                                ) : backups.length === 0 ? (
                                    <tr><td colSpan={6} className="px-6 py-8 text-center text-text-muted">No restore points found</td></tr>
                                ) : backups.map((item) => {
                                    const statusConfig = getStatusConfig(item.status);
                                    return (
                                        <tr key={item.id} className="hover:bg-surface-darker/50 transition-colors group">
                                            <td className="px-6 py-4 whitespace-nowrap">
                                                <div className="flex items-center gap-3">
                                                    <span className={`material-symbols-outlined text-[18px] text-${statusConfig.color}`}>{statusConfig.icon}</span>
                                                    <span className="font-mono text-xs text-white">{item.id.substring(0, 18)}...</span>
                                                </div>
                                            </td>
                                            <td className="px-6 py-4">
                                                <span className={`px-2 py-1 rounded text-[10px] font-bold uppercase border ${item.backup_type === 'FULL' ? 'bg-indigo-500/10 text-indigo-400 border-indigo-500/20' : 'bg-primary/10 text-primary border-primary/20'}`}>{item.backup_type}</span>
                                            </td>
                                            <td className="px-6 py-4 whitespace-nowrap">
                                                <div className="flex flex-col">
                                                    <span className="text-white">{new Date(item.created_at).toLocaleString()}</span>
                                                    <span className="text-[10px] text-text-muted">{item.location || 'System Backup'}</span>
                                                </div>
                                            </td>
                                            <td className="px-6 py-4 text-right font-mono text-text-muted">{formatSize(item.size_bytes)}</td>
                                            <td className="px-6 py-4 text-center">
                                                <span className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-[10px] font-bold bg-${statusConfig.color}/10 text-${statusConfig.color} border border-${statusConfig.color}/20 uppercase`}>
                                                    <span className={`w-1.5 h-1.5 rounded-full bg-${statusConfig.color} ${statusConfig.pulse ? 'animate-pulse' : ''}`}></span>
                                                    {statusConfig.label}
                                                </span>
                                            </td>
                                            <td className="px-6 py-4 text-right">
                                                <div className="flex items-center justify-end gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                                                    <button onClick={() => navigate(`/admin/dr/restores/initiate?backup=${item.id}`)} className="p-1.5 hover:bg-primary/20 text-text-muted hover:text-primary rounded transition-colors" title="Restore">
                                                        <span className="material-symbols-outlined text-[20px]">settings_backup_restore</span>
                                                    </button>
                                                    <button className="p-1.5 hover:bg-white/10 text-text-muted hover:text-white rounded transition-colors" title="Download">
                                                        <span className="material-symbols-outlined text-[20px]">download</span>
                                                    </button>
                                                    <button className="p-1.5 hover:bg-danger/20 text-text-muted hover:text-danger rounded transition-colors" title="Delete">
                                                        <span className="material-symbols-outlined text-[20px]">delete</span>
                                                    </button>
                                                </div>
                                            </td>
                                        </tr>
                                    );
                                })}
                            </tbody>
                        </table>
                    </div>
                    <div className="p-4 border-t border-border-color bg-surface-darker flex items-center justify-between text-xs text-text-muted">
                        <div className="flex items-center gap-4">
                            <span>Showing <span className="text-white font-medium">1-{backups.length}</span> of <span className="text-white font-medium">{backups.length}</span> points</span>
                        </div>
                        <div className="flex items-center gap-1">
                            <button className="w-7 h-7 flex items-center justify-center rounded bg-primary text-white font-medium">1</button>
                        </div>
                    </div>
                </div>
            </div>

            <footer className="bg-surface-darker border-t border-border-color px-6 py-2 flex items-center justify-between text-[10px] text-text-muted z-20">
                <span>© 2024 FleetSecurity Management Platform. DR Module V2.4</span>
                <div className="flex items-center gap-4">
                    <span>API Endpoint: <span className="font-mono">GET /dr/backups</span></span>
                    <span>BFF Connectivity: <span className="text-success font-bold uppercase">Active</span></span>
                </div>
            </footer>
        </div>
    );
};

export default RestorePointsPage;
