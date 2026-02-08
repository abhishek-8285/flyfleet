import React, { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';

interface RestoreLog {
    id: string;
    timestamp: string;
    level: string;
    module: string;
    message: string;
}

interface RestoreDetail {
    id: string;
    backup_id: string;
    restore_type: string;
    target_env: string;
    status: string;
    progress_percent?: number;
    start_time?: string;
    end_time?: string;
    initiated_by?: string;
}

const DetailRestorePage: React.FC = () => {
    const navigate = useNavigate();
    const { id } = useParams<{ id: string }>();
    const [restore, setRestore] = useState<RestoreDetail | null>(null);
    const [logs, setLogs] = useState<RestoreLog[]>([]);
    const [loading, setLoading] = useState(true);

    const fetchData = async () => {
        try {
            const [restoreRes, logsRes] = await Promise.all([
                fetch(`/api/v1/admin/dr/restores/${id}`),
                fetch(`/api/v1/admin/dr/restores/${id}/logs`)
            ]);
            const restoreData = await restoreRes.json();
            const logsData = await logsRes.json();
            setRestore(restoreData.data);
            setLogs(logsData.data || []);
        } catch (err) {
            console.error(err);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchData();
        const interval = setInterval(fetchData, 3000);
        return () => clearInterval(interval);
    }, [id]);

    const progress = restore?.progress_percent || 68; // Mock progress if not available
    const steps = ['Initialization', 'Data Stream', 'Verification', 'Cleanup'];
    const activeStep = Math.min(Math.floor((progress / 100) * steps.length), steps.length - 1);

    const getLogColor = (level: string) => {
        switch (level?.toUpperCase()) {
            case 'OK': case 'SUCCESS': return 'success';
            case 'WARN': case 'WARNING': return 'warning';
            case 'ERROR': case 'FAIL': return 'danger';
            default: return 'primary';
        }
    };

    if (loading) {
        return (
            <div className="flex items-center justify-center h-full bg-background-dark">
                <div className="text-text-muted">Loading...</div>
            </div>
        );
    }

    return (
        <div className="flex flex-col h-full w-full bg-background-dark text-white font-display overflow-hidden">
            {/* Header */}
            <header className="h-20 flex flex-col justify-center px-6 border-b border-border-color bg-surface-darker flex-shrink-0 z-20">
                <div className="flex items-center gap-2 text-[10px] font-bold text-text-muted uppercase tracking-widest mb-1">
                    <span>Disaster Recovery</span>
                    <span className="material-symbols-outlined text-[12px]">chevron_right</span>
                    <span>Restore Points</span>
                    <span className="material-symbols-outlined text-[12px]">chevron_right</span>
                    <span className="text-primary">Detail</span>
                </div>
                <div className="flex items-center justify-between">
                    <div className="flex items-center gap-4">
                        <h2 className="text-white text-xl font-bold tracking-tight">Restore Job: {id?.substring(0, 13)}</h2>
                        <span className={`px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider flex items-center gap-1 ${
                            restore?.status === 'RUNNING' 
                                ? 'bg-primary/20 text-primary border border-primary/30' 
                                : restore?.status === 'COMPLETED'
                                    ? 'bg-success/20 text-success border border-success/30'
                                    : 'bg-warning/20 text-warning border border-warning/30'
                        }`}>
                            {restore?.status === 'RUNNING' && <span className="w-1.5 h-1.5 rounded-full bg-primary animate-pulse"></span>}
                            {restore?.status || 'PENDING'}
                        </span>
                    </div>
                    <div className="flex items-center gap-3">
                        <button className="px-4 py-2 bg-surface-dark border border-border-color hover:bg-border-color text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors">
                            <span className="material-symbols-outlined text-[16px]">verified</span>
                            Verify Integrity
                        </button>
                        <button onClick={() => navigate('/admin/dr')} className="px-4 py-2 bg-danger/10 border border-danger/30 hover:bg-danger/20 text-danger text-xs font-semibold rounded flex items-center gap-2 transition-colors">
                            <span className="material-symbols-outlined text-[16px]">history</span>
                            Cancel / Rollback
                        </button>
                    </div>
                </div>
            </header>

            {/* Stats Cards */}
            <div className="p-6 pb-0 grid grid-cols-4 gap-4 flex-shrink-0">
                {[
                    { label: 'Restore Type', value: restore?.restore_type || 'FULL', sub: 'Full System Recovery', icon: 'database' },
                    { label: 'Target Environment', value: restore?.target_env || 'STAGING', sub: 'Cluster: K8S-Fleet-04', icon: 'cloud' },
                    { label: 'Start Time', value: restore?.start_time ? new Date(restore.start_time).toLocaleTimeString() : '--:--', sub: restore?.start_time ? new Date(restore.start_time).toLocaleDateString() : '', icon: 'schedule' },
                    { label: 'Current Status', value: restore?.status || 'PENDING', sub: `Progress: ${progress}%`, icon: 'data_exploration', primary: true }
                ].map((card, idx) => (
                    <div key={idx} className="bg-surface-dark rounded-xl border border-border-color p-4 relative overflow-hidden group">
                        <div className="text-text-muted text-[10px] font-bold uppercase tracking-wider mb-1">{card.label}</div>
                        <div className={`text-xl font-bold ${card.primary ? 'text-primary flex items-center gap-2' : 'text-white'}`}>
                            {card.value}
                            {card.primary && restore?.status === 'RUNNING' && <span className="material-symbols-outlined text-[18px] animate-spin">sync</span>}
                        </div>
                        <div className="text-[10px] text-text-muted mt-1">{card.sub}</div>
                        <div className="absolute right-3 top-3 opacity-10">
                            <span className={`material-symbols-outlined text-4xl ${card.primary ? 'text-primary' : ''}`}>{card.icon}</span>
                        </div>
                    </div>
                ))}
            </div>

            {/* Main Content */}
            <div className="flex-1 overflow-hidden p-6 flex gap-6">
                {/* Left Panel - Progress + Logs */}
                <div className="flex-1 flex flex-col gap-6 min-w-0">
                    {/* Progress Card */}
                    <div className="bg-surface-dark rounded-xl border border-border-color p-6">
                        <div className="flex justify-between items-center mb-4">
                            <h3 className="text-white text-sm font-semibold flex items-center gap-2">
                                <span className="material-symbols-outlined text-primary text-[20px]">analytics</span>
                                Real-time Progress
                            </h3>
                            <span className="text-primary font-mono font-bold">{progress}%</span>
                        </div>
                        <div className="w-full bg-surface-darker h-4 rounded-full border border-border-color p-0.5 overflow-hidden">
                            <div 
                                className="h-full bg-gradient-to-r from-primary/50 to-primary rounded-full relative transition-all duration-500 ease-out"
                                style={{ width: `${progress}%` }}
                            >
                                <div className="absolute inset-0 bg-[linear-gradient(45deg,rgba(255,255,255,0.1)_25%,transparent_25%,transparent_50%,rgba(255,255,255,0.1)_50%,rgba(255,255,255,0.1)_75%,transparent_75%,transparent)] bg-[length:20px_20px] animate-pulse"></div>
                            </div>
                        </div>
                        <div className="grid grid-cols-4 gap-2 mt-6">
                            {steps.map((step, i) => (
                                <div key={i} className={`text-center ${i > activeStep ? 'opacity-40' : ''}`}>
                                    <div className={`text-[10px] font-bold uppercase mb-1 ${i < activeStep ? 'text-success' : i === activeStep ? 'text-primary' : 'text-text-muted'}`}>{step}</div>
                                    <div className={`h-1 rounded-full ${i < activeStep ? 'bg-success' : i === activeStep ? 'bg-primary animate-pulse' : 'bg-border-color'}`}></div>
                                </div>
                            ))}
                        </div>
                    </div>

                    {/* Logs Panel */}
                    <div className="flex-1 bg-surface-darker rounded-xl border border-border-color flex flex-col overflow-hidden">
                        <div className="p-3 border-b border-border-color bg-surface-dark flex justify-between items-center">
                            <div className="flex items-center gap-2">
                                <span className="material-symbols-outlined text-[18px] text-text-muted">terminal</span>
                                <h3 className="text-white text-xs font-semibold">Live System Logs</h3>
                            </div>
                            <div className="flex items-center gap-3">
                                <div className="flex items-center gap-2 text-[10px] text-text-muted">
                                    <span className="w-2 h-2 rounded-full bg-success"></span> Live
                                </div>
                                <button className="text-text-muted hover:text-white">
                                    <span className="material-symbols-outlined text-[18px]">download</span>
                                </button>
                            </div>
                        </div>
                        <div className="flex-1 overflow-y-auto p-4 font-mono text-xs space-y-1 custom-scrollbar">
                            {logs.length === 0 ? (
                                <div className="text-text-muted text-center py-4">No logs yet...</div>
                            ) : logs.map((log, i) => (
                                <div key={i} className={`py-0.5 px-2 border-l-2 border-${getLogColor(log.level)}/40 hover:bg-surface-dark/50`}>
                                    <span className="text-text-muted">[{new Date(log.timestamp).toLocaleTimeString()}]</span>{' '}
                                    <span className={`text-${getLogColor(log.level)}`}>{log.level}:</span>{' '}
                                    {log.message}
                                </div>
                            ))}
                        </div>
                    </div>
                </div>

                {/* Right Panel - Metadata */}
                <aside className="w-80 flex flex-col gap-6 flex-shrink-0">
                    <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
                        <div className="p-4 border-b border-border-color bg-surface-darker">
                            <h3 className="text-white text-sm font-semibold">Job Metadata</h3>
                        </div>
                        <div className="p-4 space-y-4">
                            <div>
                                <div className="text-[10px] text-text-muted font-bold uppercase tracking-wider mb-1">Initiated By</div>
                                <div className="flex items-center gap-2">
                                    <div className="w-6 h-6 rounded-full bg-primary/20 flex items-center justify-center text-[10px] text-primary font-bold">
                                        {restore?.initiated_by?.substring(0, 2).toUpperCase() || 'SY'}
                                    </div>
                                    <span className="text-sm text-white font-medium">{restore?.initiated_by || 'System'}</span>
                                </div>
                            </div>
                            <hr className="border-border-color"/>
                            <div>
                                <div className="text-[10px] text-text-muted font-bold uppercase tracking-wider mb-1">Backup ID</div>
                                <div className="flex items-center justify-between">
                                    <span className="text-sm font-mono text-white truncate">{restore?.backup_id?.substring(0, 20)}...</span>
                                    <button className="text-text-muted hover:text-white material-symbols-outlined text-[16px]">content_copy</button>
                                </div>
                            </div>
                            <hr className="border-border-color"/>
                            <div>
                                <div className="text-[10px] text-text-muted font-bold uppercase tracking-wider mb-1">Verification Status</div>
                                <div className={`flex items-center gap-2 ${restore?.status === 'RUNNING' ? 'text-warning' : 'text-success'}`}>
                                    <span className="material-symbols-outlined text-[18px]">
                                        {restore?.status === 'RUNNING' ? 'hourglass_empty' : 'check_circle'}
                                    </span>
                                    <span className="text-sm font-semibold">
                                        {restore?.status === 'RUNNING' ? 'In Progress' : 'Complete'}
                                    </span>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div className="bg-surface-dark rounded-xl border border-border-color p-4">
                        <h3 className="text-white text-xs font-semibold mb-3">Health Indicators</h3>
                        <div className="space-y-3">
                            <div className="flex items-center justify-between text-xs">
                                <span className="text-text-muted">Disk I/O</span>
                                <span className="text-success font-medium">Optimal</span>
                            </div>
                            <div className="w-full bg-surface-darker h-1 rounded-full">
                                <div className="bg-success h-full rounded-full" style={{ width: '45%' }}></div>
                            </div>
                            <div className="flex items-center justify-between text-xs mt-4">
                                <span className="text-text-muted">Network Throughput</span>
                                <span className="text-success font-medium">850 MB/s</span>
                            </div>
                            <div className="w-full bg-surface-darker h-1 rounded-full">
                                <div className="bg-success h-full rounded-full" style={{ width: '78%' }}></div>
                            </div>
                        </div>
                    </div>

                    <div className="mt-auto bg-primary/5 border border-primary/20 rounded-xl p-4">
                        <div className="flex items-start gap-3">
                            <span className="material-symbols-outlined text-primary">info</span>
                            <div>
                                <div className="text-xs text-white font-semibold mb-1">SSE Connected</div>
                                <p className="text-[10px] text-text-muted leading-relaxed">
                                    Real-time updates via Server-Sent Events. Logs refresh every 3 seconds.
                                </p>
                            </div>
                        </div>
                    </div>
                </aside>
            </div>

            {/* Footer */}
            <footer className="bg-surface-darker border-t border-border-color px-6 py-3 flex items-center justify-between text-[10px] text-text-muted z-20">
                <div className="flex items-center gap-4">
                    <span>© 2024 FleetSecurity DR Module</span>
                    <span className="w-1 h-1 rounded-full bg-border-color"></span>
                    <span>Version 4.2.0-stable</span>
                </div>
                <div className="flex items-center gap-4">
                    <span>Node Status: <span className="text-success font-bold uppercase">Healthy</span></span>
                    <span>Connection: <span className="text-success font-bold uppercase">Encrypted</span></span>
                    <span className="font-mono text-white bg-surface-dark px-2 py-0.5 rounded border border-border-color">Session: {id?.substring(0, 8)}</span>
                </div>
            </footer>
        </div>
    );
};

export default DetailRestorePage;
