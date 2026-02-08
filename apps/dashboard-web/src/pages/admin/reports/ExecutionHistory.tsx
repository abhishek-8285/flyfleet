import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';

interface Execution {
    id: string;
    report_id: string;
    started_at: string;
    duration_seconds?: number;
    status: string;
    file_url?: string;
    error_message?: string;
}

const ExecutionHistory: React.FC = () => {
    const navigate = useNavigate();
    const [executions, setExecutions] = useState<Execution[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        fetch('/api/v1/admin/reports/executions')
            .then(res => res.json())
            .then(data => {
                setExecutions(data.executions || []);
                setLoading(false);
            })
            .catch(err => {
                console.error("Failed to fetch executions", err);
                setLoading(false);
            });
    }, []);

    const formatDuration = (seconds?: number) => {
        if (!seconds) return '-';
        if (seconds < 60) return `${seconds}s`;
        const mins = Math.floor(seconds / 60);
        return `${mins}m ${seconds % 60}s`;
    };

    return (
        <div className="flex h-screen w-full bg-background-dark text-white">
            <main className="flex-1 flex flex-col min-w-0 overflow-hidden">
                <header className="h-16 flex items-center justify-between px-6 border-b border-border-color bg-surface-darker flex-shrink-0 z-20">
                    <div className="flex flex-col">
                        <nav className="flex items-center gap-2 text-[10px] text-text-muted uppercase tracking-widest mb-1">
                            <span onClick={() => navigate('/admin/reports')} className="hover:text-primary transition-colors cursor-pointer">Schedules</span>
                            <span className="material-symbols-outlined text-[12px]">chevron_right</span>
                            <span className="text-white">Execution History</span>
                        </nav>
                        <h2 className="text-white text-lg font-bold leading-tight">Report Execution History</h2>
                    </div>
                </header>
                <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
                    <div className="bg-surface-dark rounded-xl border border-border-color flex flex-col overflow-hidden">
                        <div className="p-4 border-b border-border-color flex justify-between items-center bg-surface-darker">
                            <h3 className="text-white font-semibold text-sm flex items-center gap-2">
                                <span className="material-symbols-outlined text-text-muted text-[18px]">format_list_bulleted</span> Execution Logs
                            </h3>
                        </div>
                        <div className="overflow-x-auto">
                            <table className="w-full text-left border-collapse">
                                <thead>
                                    <tr className="border-b border-border-color text-[11px] text-text-muted uppercase tracking-wider bg-surface-darker/50">
                                        <th className="px-4 py-3 font-semibold">Run ID</th>
                                        <th className="px-4 py-3 font-semibold">Start Time</th>
                                        <th className="px-4 py-3 font-semibold">Duration</th>
                                        <th className="px-4 py-3 font-semibold">Status</th>
                                        <th className="px-4 py-3 font-semibold">Download</th>
                                    </tr>
                                </thead>
                                <tbody className="text-sm">
                                    {loading ? (
                                        <tr><td colSpan={5} className="px-4 py-4 text-center text-text-muted">Loading...</td></tr>
                                    ) : executions.map((exec) => (
                                        <tr key={exec.id} className="border-b border-border-color hover:bg-surface-darker/50">
                                            <td className="px-4 py-4 text-white font-mono text-xs">{exec.id.substring(0, 8)}...</td>
                                            <td className="px-4 py-4 text-text-muted text-xs">{new Date(exec.started_at).toLocaleString()}</td>
                                            <td className="px-4 py-4 text-text-muted text-xs">{formatDuration(exec.duration_seconds)}</td>
                                            <td className="px-4 py-4">
                                                <span className={`px-2 py-0.5 rounded text-[10px] font-bold border ${exec.status === 'SUCCESS' ? 'bg-success/10 text-success border-success/20' : exec.status === 'FAILED' ? 'bg-danger/10 text-danger border-danger/20' : 'bg-primary/10 text-primary border-primary/20'}`}>
                                                    {exec.status}
                                                </span>
                                            </td>
                                            <td className="px-4 py-4">
                                                {exec.file_url ? (
                                                    <a className="text-primary text-xs hover:underline" href={exec.file_url} target="_blank" rel="noreferrer">Download</a>
                                                ) : (
                                                    <span className="text-text-muted italic text-xs">Unavailable</span>
                                                )}
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    </div>
                </div>
            </main>
        </div>
    );
};

export default ExecutionHistory;
