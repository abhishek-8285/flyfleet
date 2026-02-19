import React, { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import apiClient from '../../../services/apiClient';

interface Log {
    id: string;
    level: string;
    module: string;
    message: string;
    timestamp: string;
}

interface RestoreDetail {
    id: string;
    backup_id: string;
    restore_type: string;
    target_env: string;
    status: string;
    start_time: string;
    duration_seconds?: number;
    initiated_by: string;
}

const OperationDetail: React.FC = () => {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const [restore, setRestore] = useState<RestoreDetail | null>(null);
    const [logs, setLogs] = useState<Log[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const fetchData = async () => {
        if (!id) return;
        try {
            setLoading(true);
            setError(null);
            const data = await apiClient.get(`/admin/dr/restore-jobs/${id}`) as any;
            setRestore(data.data || data.restore || data);
            setLogs(data.logs || []);
        } catch (err) {
            console.error(err);
            setError('Failed to load restore operation.');
            setRestore(null);
            setLogs([]);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchData();
        const interval = setInterval(fetchData, 3000);
        return () => clearInterval(interval);
    }, [id]);

    const formatDuration = (seconds?: number) => {
        if (!seconds) return '--:--:--';
        const h = Math.floor(seconds / 3600);
        const m = Math.floor((seconds % 3600) / 60);
        const s = seconds % 60;
        return `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`;
    };

    if (loading) return <div className="p-8 text-white">Loading restore details...</div>;
    if (error) return <div className="p-8 text-red-400">{error}</div>;
    if (!restore) return <div className="p-8 text-white">Restore operation not found</div>;

    const steps = [
        { name: "Initialization", status: "COMPLETED" },
        { name: "Extraction", status: restore.status === 'RUNNING' || restore.status === 'COMPLETED' ? "COMPLETED" : "PENDING" },
        { name: "Validation", status: restore.status === 'RUNNING' ? "RUNNING" : restore.status === 'COMPLETED' ? "COMPLETED" : "PENDING" },
        { name: "Completion", status: restore.status === 'COMPLETED' ? "COMPLETED" : "PENDING" }
    ];

    return (
        <div className="flex flex-col h-full w-full bg-background-dark text-white font-display overflow-hidden">
             <header className="h-16 flex items-center justify-between px-6 border-b border-border-color bg-surface-darker flex-shrink-0 z-20">
                <div className="flex flex-col">
                    <nav className="flex items-center text-[10px] text-text-muted uppercase tracking-widest font-semibold mb-0.5">
                        <button onClick={() => navigate('/admin/dr/restores')} className="hover:text-primary transition-colors">Restores</button>
                        <span className="mx-2 text-border-color">/</span>
                        <span className="text-white">{restore.id}</span>
                    </nav>
                    <div className="flex items-center gap-4">
                        <h2 className="text-white text-lg font-bold">Operation Detail View</h2>
                        <span className={`px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider bg-${restore.status === 'RUNNING' ? 'primary' : restore.status === 'COMPLETED' ? 'success' : 'danger'}/20 text-${restore.status === 'RUNNING' ? 'primary' : restore.status === 'COMPLETED' ? 'success' : 'danger'} border border-${restore.status === 'RUNNING' ? 'primary' : restore.status === 'COMPLETED' ? 'success' : 'danger'}/30`}>
                            {restore.status}
                        </span>
                    </div>
                </div>
                {/* ... Actions ... */}
            </header>

            <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
                <div className="grid grid-cols-4 gap-4 mb-6">
                    <div className="bg-surface-dark rounded-xl border border-border-color p-4">
                        <div className="text-text-muted text-xs font-medium uppercase mb-1">Restore Type</div>
                        <div className="text-xl font-bold text-white uppercase">{restore.restore_type}</div>
                        <div className="text-[10px] text-primary flex items-center gap-1 mt-1">
                            <span className="material-symbols-outlined text-[12px]">database</span>
                            Target: {restore.target_env}
                        </div>
                    </div>
                    {/* ... other stats ... */}
                </div>

                <div className="grid grid-cols-12 gap-6">
                    <div className="col-span-12">
                        <div className="bg-surface-dark rounded-xl border border-border-color p-8 mb-6">
                            <h3 className="text-white text-sm font-semibold mb-8 flex items-center gap-2">
                                <span className="material-symbols-outlined text-primary text-[18px]">account_tree</span>
                                Progress Timeline
                            </h3>
                            <div className="relative flex items-center justify-between px-4">
                                {steps.map((step, i) => (
                                    <div key={i} className={`flex flex-col items-center z-10 w-1/4 relative timeline-step ${step.status === 'COMPLETED' ? 'completed' : ''}`}>
                                        <div className={`w-10 h-10 rounded-full flex items-center justify-center border-4 border-background-dark mb-3 ${step.status === 'COMPLETED' ? 'bg-success text-white' : step.status === 'RUNNING' ? 'bg-primary text-white ring-2 ring-primary/30' : 'bg-border-color text-text-muted'}`}>
                                            <span className={`material-symbols-outlined ${step.status === 'RUNNING' ? 'animate-pulse' : ''}`}>{step.status === 'COMPLETED' ? 'check' : step.status === 'RUNNING' ? 'settings_ethernet' : 'circle'}</span>
                                        </div>
                                        <span className={`text-xs font-bold ${step.status === 'PENDING' ? 'text-text-muted' : 'text-white'}`}>{step.name}</span>
                                        <span className={`text-[10px] mt-1 ${step.status === 'RUNNING' ? 'text-primary' : 'text-text-muted'}`}>{step.status}</span>
                                    </div>
                                ))}
                            </div>
                        </div>
                    </div>

                    <div className="col-span-12">
                         <div className="bg-surface-dark rounded-xl border border-border-color flex flex-col overflow-hidden">
                             {/* ... Logs Table ... */}
                             <div className="p-4 border-b border-border-color flex justify-between items-center bg-surface-darker">
                                 <h3 className="text-white font-semibold text-sm flex items-center gap-2">Operation Logs</h3>
                             </div>
                             <div className="max-h-[400px] overflow-y-auto custom-scrollbar bg-surface-darker/30">
                                 <table className="w-full text-left border-collapse">
                                     <thead>
                                         <tr className="sticky top-0 z-10 border-b border-border-color text-xs text-text-muted uppercase tracking-wider bg-surface-darker/80 backdrop-blur-sm">
                                             <th className="px-4 py-3 font-semibold">Timestamp</th>
                                             <th className="px-4 py-3 font-semibold">Level</th>
                                             <th className="px-4 py-3 font-semibold">Message</th>
                                         </tr>
                                     </thead>
                                     <tbody className="text-xs font-mono">
                                         {logs.map((log) => (
                                             <tr key={log.id} className="border-b border-border-color/50 hover:bg-surface-darker transition-colors">
                                                 <td className="px-4 py-2.5 text-text-muted">{new Date(log.timestamp).toLocaleTimeString()}</td>
                                                 <td className="px-4 py-2.5"><span className={log.level === 'ERROR' ? 'text-danger' : log.level === 'WARN' ? 'text-warning' : 'text-primary'}>{log.level}</span></td>
                                                 <td className="px-4 py-2.5 text-white">{log.message}</td>
                                             </tr>
                                         ))}
                                     </tbody>
                                 </table>
                             </div>
                         </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default OperationDetail;
