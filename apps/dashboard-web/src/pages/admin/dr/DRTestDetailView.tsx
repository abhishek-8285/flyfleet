import React, { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';

interface DRTest {
    id: string;
    test_id: string;
    name: string;
    description: string;
    test_type: string;
    target_env: string;
    status: string;
    progress_percent: number;
    initiated_by: string;
    success_criteria: string;
    scheduled_at: string;
    started_at: string;
    completed_at: string;
}

interface Step {
    id: string;
    step_order: number;
    step_name: string;
    status: string;
    started_at: string;
    completed_at: string;
}

interface Log {
    id: string;
    timestamp: string;
    level: string;
    message: string;
}

const DRTestDetailView: React.FC = () => {
    const navigate = useNavigate();
    const { id } = useParams();
    const [test, setTest] = useState<DRTest | null>(null);
    const [steps, setSteps] = useState<Step[]>([]);
    const [logs, setLogs] = useState<Log[]>([]);
    const [loading, setLoading] = useState(true);

    const fetchData = async () => {
        try {
            const res = await fetch(`/api/v1/admin/dr/tests/${id}`);
            const data = await res.json();
            setTest(data.data);
            setSteps(data.steps || []);
            setLogs(data.logs || []);
        } catch (error) {
            console.error('Failed to fetch DR test:', error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchData();
        const interval = setInterval(fetchData, 3000);
        return () => clearInterval(interval);
    }, [id]);

    const handleAbort = async () => {
        if (!window.confirm('Are you sure you want to abort this test?')) return;
        try {
            await fetch(`/api/v1/admin/dr/tests/${id}/abort`, { method: 'POST' });
            fetchData();
        } catch (error) {
            console.error('Failed to abort test:', error);
        }
    };

    const getStatusColor = (status: string) => {
        switch (status) {
            case 'PASSED': return 'text-green-400 bg-green-400/10 border-green-400/30';
            case 'FAILED': return 'text-red-400 bg-red-400/10 border-red-400/30';
            case 'RUNNING': return 'text-blue-400 bg-blue-400/10 border-blue-400/30';
            case 'SCHEDULED': case 'PENDING': return 'text-yellow-400 bg-yellow-400/10 border-yellow-400/30';
            case 'ABORTED': case 'SKIPPED': return 'text-gray-400 bg-gray-400/10 border-gray-400/30';
            default: return 'text-gray-400 bg-gray-400/10 border-gray-400/30';
        }
    };

    const getLevelColor = (level: string) => {
        switch (level) {
            case 'ERROR': return 'text-red-400';
            case 'WARN': return 'text-yellow-400';
            case 'RUN': return 'text-cyan-400';
            default: return 'text-white/70';
        }
    };

    if (loading) {
        return <div className="flex items-center justify-center h-full bg-[#0a0a0f] text-white">Loading...</div>;
    }

    if (!test) {
        return <div className="flex items-center justify-center h-full bg-[#0a0a0f] text-white">Test not found</div>;
    }

    return (
        <div className="flex flex-col h-full w-full bg-[#0a0a0f] text-white font-sans overflow-hidden">
            {/* Header */}
            <header className="h-16 flex items-center justify-between px-6 border-b border-white/10 bg-[#12121a] flex-shrink-0">
                <div className="flex items-center gap-4">
                    <button onClick={() => navigate('/admin/dr/tests')} className="text-white/60 hover:text-white">
                        ← Back
                    </button>
                    <div>
                        <h1 className="text-lg font-semibold">{test.name}</h1>
                        <p className="text-xs text-cyan-400">{test.test_id}</p>
                    </div>
                </div>
                <div className="flex items-center gap-3">
                    <span className={`text-xs px-3 py-1 rounded-full border ${getStatusColor(test.status)}`}>
                        {test.status === 'RUNNING' && <span className="animate-pulse mr-1">●</span>}
                        {test.status}
                    </span>
                    {test.status === 'RUNNING' && (
                        <button
                            onClick={handleAbort}
                            className="px-4 py-2 bg-red-500/20 hover:bg-red-500/30 border border-red-500/30 text-red-400 rounded-lg text-sm"
                        >
                            ⏹ Abort
                        </button>
                    )}
                </div>
            </header>

            <div className="flex-1 overflow-hidden flex gap-4 p-6">
                {/* Left - Progress & Steps */}
                <div className="flex-1 flex flex-col gap-4 overflow-auto">
                    {/* Progress */}
                    <div className="bg-[#12121a] border border-white/10 rounded-xl p-6">
                        <div className="flex justify-between items-center mb-4">
                            <span className="text-sm text-white/70">Overall Progress</span>
                            <span className="text-2xl font-bold">{test.progress_percent}%</span>
                        </div>
                        <div className="bg-white/10 rounded-full h-3 overflow-hidden">
                            <div
                                className="h-full bg-gradient-to-r from-emerald-500 to-cyan-500 transition-all duration-500"
                                style={{ width: `${test.progress_percent}%` }}
                            />
                        </div>
                    </div>

                    {/* Steps */}
                    <div className="bg-[#12121a] border border-white/10 rounded-xl p-6 flex-1">
                        <h3 className="text-sm font-medium text-white/70 mb-4">Execution Steps</h3>
                        <div className="space-y-3">
                            {steps.map((step) => (
                                <div key={step.id} className="flex items-center gap-3">
                                    <span className={`w-8 h-8 rounded-full flex items-center justify-center text-sm border ${getStatusColor(step.status)}`}>
                                        {step.status === 'PASSED' ? '✓' : step.status === 'FAILED' ? '✗' : step.status === 'RUNNING' ? '⟳' : step.step_order}
                                    </span>
                                    <div className="flex-1">
                                        <span className="text-sm">{step.step_name}</span>
                                    </div>
                                    <span className={`text-xs ${getStatusColor(step.status)} px-2 py-1 rounded`}>{step.status}</span>
                                </div>
                            ))}
                        </div>
                    </div>

                    {/* Details */}
                    <div className="bg-[#12121a] border border-white/10 rounded-xl p-6">
                        <h3 className="text-sm font-medium text-white/70 mb-4">Test Details</h3>
                        <div className="grid grid-cols-2 gap-4 text-sm">
                            <div><span className="text-white/50">Type:</span> <span>{test.test_type}</span></div>
                            <div><span className="text-white/50">Target:</span> <span>{test.target_env}</span></div>
                            <div><span className="text-white/50">Initiated By:</span> <span>{test.initiated_by || '—'}</span></div>
                            <div><span className="text-white/50">Started:</span> <span>{test.started_at ? new Date(test.started_at).toLocaleString() : '—'}</span></div>
                            <div className="col-span-2"><span className="text-white/50">Success Criteria:</span> <span>{test.success_criteria || '—'}</span></div>
                        </div>
                    </div>
                </div>

                {/* Right - Logs */}
                <div className="w-[400px] bg-[#12121a] border border-white/10 rounded-xl flex flex-col overflow-hidden">
                    <div className="px-4 py-3 border-b border-white/10 flex items-center justify-between">
                        <h3 className="text-sm font-medium">Live Logs</h3>
                        <span className="text-xs text-white/40">{logs.length} entries</span>
                    </div>
                    <div className="flex-1 overflow-auto p-4 font-mono text-xs space-y-1 bg-black/30">
                        {logs.length === 0 ? (
                            <div className="text-white/40">No logs yet...</div>
                        ) : (
                            logs.map((log) => (
                                <div key={log.id} className="flex gap-2">
                                    <span className="text-white/30">{new Date(log.timestamp).toLocaleTimeString()}</span>
                                    <span className={`${getLevelColor(log.level)} w-12`}>[{log.level}]</span>
                                    <span className="text-white/80">{log.message}</span>
                                </div>
                            ))
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
};

export default DRTestDetailView;
