import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import apiClient from '../../../services/apiClient';

interface Restore {
    id: string;
    backup_id: string;
    restore_type: string;
    start_time: string;
    status: string;
    duration_seconds?: number;
}

const RestoreOperations: React.FC = () => {
    const navigate = useNavigate();
    const [restores, setRestores] = useState<Restore[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const fetchRestores = async () => {
        try {
            setLoading(true);
            setError(null);
            const data = await apiClient.get('/admin/dr/restore-jobs') as any;
            const items = data?.data || data?.items || data?.restores || data || [];
            setRestores(Array.isArray(items) ? items : []);
        } catch (err) {
            console.error(err);
            setError('Failed to load restore operations.');
            setRestores([]);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchRestores();
        const interval = setInterval(fetchRestores, 5000);
        return () => clearInterval(interval);
    }, []);

    const formatDuration = (seconds?: number) => {
        if (!seconds) return '--:--:--';
        const h = Math.floor(seconds / 3600);
        const m = Math.floor((seconds % 3600) / 60);
        const s = seconds % 60;
        return `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`;
    };

    return (
        <div className="flex flex-col h-full w-full bg-background-dark text-white font-display overflow-hidden">
            <header className="h-16 flex items-center justify-between px-6 border-b border-border-color bg-surface-darker flex-shrink-0 z-20">
                <div className="flex flex-col">
                    <div className="flex items-center gap-2 text-[10px] text-text-muted uppercase tracking-wider mb-0.5">
                        <span>Infrastructure</span>
                        <span className="material-symbols-outlined text-[10px]">chevron_right</span>
                        <span>Disaster Recovery</span>
                        <span className="material-symbols-outlined text-[10px]">chevron_right</span>
                        <span className="text-primary font-bold">Restore Operations</span>
                    </div>
                    <h2 className="text-white text-lg font-bold leading-tight">Restore Operations Screen</h2>
                </div>
                <div className="flex items-center gap-4">
                    <div className="relative group">
                        <span className="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-text-muted text-[18px]">search</span>
                        <input className="bg-surface-dark border border-border-color text-xs text-white rounded-lg pl-9 pr-3 py-2 w-72 focus:ring-1 focus:ring-primary focus:border-primary outline-none transition-all" placeholder="Global Search..." type="text"/>
                    </div>
                    <button 
                        onClick={() => navigate('/admin/dr/restores/initiate')}
                        className="px-4 py-2 bg-primary hover:bg-primary-dark text-white text-xs font-semibold rounded-lg flex items-center gap-2 transition-colors shadow-lg shadow-primary/20"
                    >
                        <span className="material-symbols-outlined text-[16px]">play_circle</span>
                        Initiate Restore
                    </button>
                </div>
            </header>

            <div className="p-6 pb-0">
                <div className="grid grid-cols-4 gap-4 mb-6">
                    <div className="bg-surface-dark rounded-xl border border-border-color p-4 flex items-center gap-4">
                        <div className="w-10 h-10 rounded bg-primary/10 flex items-center justify-center text-primary">
                            <span className="material-symbols-outlined">sync</span>
                        </div>
                        <div>
                            <div className="text-[10px] font-bold text-text-muted uppercase">Active Restores</div>
                            <div className="text-xl font-bold text-white">{restores.filter(r => r.status === 'RUNNING').length}</div>
                        </div>
                    </div>
                    {/* ... other stats ... */}
                </div>
            </div>

            <div className="flex-1 overflow-hidden px-6 pb-6">
                <div className="bg-surface-dark rounded-xl border border-border-color flex flex-col h-full overflow-hidden">
                    <div className="p-4 border-b border-border-color flex justify-between items-center bg-surface-darker/30">
                        <h3 className="text-white font-semibold text-sm flex items-center gap-2">
                            <span className="material-symbols-outlined text-text-muted text-[18px]">list_alt</span>
                            Active and Past Operations
                        </h3>
                        <div className="flex items-center gap-2">
                            <button onClick={fetchRestores} className="p-1.5 hover:bg-border-color rounded text-text-muted hover:text-white transition-colors">
                                <span className="material-symbols-outlined text-[18px]">refresh</span>
                            </button>
                        </div>
                    </div>
                    <div className="overflow-y-auto custom-scrollbar flex-1">
                        <table className="w-full text-left border-collapse">
                            <thead className="sticky top-0 z-10">
                                <tr className="border-b border-border-color text-xs text-text-muted uppercase tracking-wider bg-surface-darker">
                                    <th className="px-6 py-4 font-semibold">Restore ID</th>
                                    <th className="px-6 py-4 font-semibold">Backup ID</th>
                                    <th className="px-6 py-4 font-semibold">Type</th>
                                    <th className="px-6 py-4 font-semibold">Start Time</th>
                                    <th className="px-6 py-4 font-semibold">Status</th>
                                    <th className="px-6 py-4 font-semibold text-right">Duration</th>
                                </tr>
                            </thead>
                            <tbody className="text-sm">
                                {loading ? (
                                    <tr><td colSpan={6} className="px-6 py-4 text-center text-text-muted">Loading...</td></tr>
                                ) : error ? (
                                    <tr><td colSpan={6} className="px-6 py-4 text-center text-red-400">{error}</td></tr>
                                ) : restores.length === 0 ? (
                                    <tr><td colSpan={6} className="px-6 py-4 text-center text-text-muted">No restore operations found.</td></tr>
                                ) : restores.map((row) => (
                                    <tr key={row.id} 
                                        className="border-b border-border-color hover:bg-surface-darker/50 transition-colors cursor-pointer"
                                        onClick={() => navigate(`/admin/dr/restores/${row.id}`)}
                                    >
                                        <td className="px-6 py-4 font-mono text-xs text-primary">{row.id.substring(0, 13)}...</td>
                                        <td className="px-6 py-4 font-mono text-xs text-text-muted">{row.backup_id.substring(0, 13)}...</td>
                                        <td className="px-6 py-4">
                                            <span className={`px-2 py-0.5 rounded text-[10px] font-bold ${row.restore_type === 'FULL' ? 'bg-border-color text-white' : 'bg-border-color/50 text-text-muted'}`}>{row.restore_type}</span>
                                        </td>
                                        <td className="px-6 py-4 text-text-muted whitespace-nowrap">{new Date(row.start_time).toLocaleString()}</td>
                                        <td className="px-6 py-4">
                                            <span className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded text-[10px] font-bold ${row.status === 'RUNNING' ? 'bg-primary/10 text-primary border border-primary/20' : 'bg-surface-darker text-text-muted border border-border-color'}`}>
                                                {row.status === 'RUNNING' && <span className="size-1.5 rounded-full bg-primary animate-pulse"></span>}
                                                {row.status}
                                            </span>
                                        </td>
                                        <td className="px-6 py-4 text-right text-text-muted font-mono">{formatDuration(row.duration_seconds)}</td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>
             <footer className="bg-surface-darker border-t border-border-color px-6 py-2 flex items-center justify-between text-[10px] text-text-muted z-20 flex-shrink-0">
                <span>© 2023 Enterprise Fleet Manager. DR Module V2.4</span>
                <div className="flex items-center gap-4">
                    <span>API Endpoint: <span className="text-primary">/admin/dr/restore-jobs</span></span>
                    <span>System Health: <span className="text-success">Stable</span></span>
                </div>
            </footer>
        </div>
    );
};

export default RestoreOperations;
