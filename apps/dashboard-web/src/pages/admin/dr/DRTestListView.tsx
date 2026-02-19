import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import apiClient from '../../../services/apiClient';

interface DRTest {
    id: string;
    test_id: string;
    name: string;
    test_type: string;
    target_env: string;
    status: string;
    progress_percent: number;
    initiated_by: string;
    scheduled_at: string;
    started_at: string;
    completed_at: string;
}

interface Metrics {
    total_tests: number;
    success_rate: number;
    avg_recovery_time: string;
    next_scheduled: string;
}

const DRTestListView: React.FC = () => {
    const navigate = useNavigate();
    const [tests, setTests] = useState<DRTest[]>([]);
    const [metrics, setMetrics] = useState<Metrics | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const fetchTests = async () => {
        try {
            setLoading(true);
            setError(null);
            const [testsRes, metricsRes] = await Promise.all([
                apiClient.get('/admin/dr/tests') as any,
                apiClient.get('/admin/dr/tests/metrics') as any
            ]);
            const testsData = testsRes;
            const metricsData = metricsRes;
            const items = testsData?.data || testsData?.items || testsData?.tests || testsData || [];
            setTests(Array.isArray(items) ? items : []);
            setMetrics(metricsData?.data || metricsData?.metrics || null);
        } catch (error) {
            console.error('Failed to fetch DR tests:', error);
            setError('Failed to load DR tests.');
            setTests([]);
            setMetrics(null);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchTests();
        const interval = setInterval(fetchTests, 10000);
        return () => clearInterval(interval);
    }, []);

    const getStatusColor = (status: string) => {
        switch (status) {
            case 'PASSED': return 'text-green-400 bg-green-400/10';
            case 'FAILED': return 'text-red-400 bg-red-400/10';
            case 'RUNNING': return 'text-blue-400 bg-blue-400/10';
            case 'SCHEDULED': return 'text-yellow-400 bg-yellow-400/10';
            case 'ABORTED': return 'text-gray-400 bg-gray-400/10';
            default: return 'text-gray-400 bg-gray-400/10';
        }
    };

    const getTypeIcon = (type: string) => {
        switch (type) {
            case 'FAILOVER': return '🔄';
            case 'BACKUP': return '💾';
            case 'RESTORE': return '📥';
            default: return '🧪';
        }
    };

    return (
        <div className="flex flex-col h-full w-full bg-[#0a0a0f] text-white font-sans overflow-hidden">
            {/* Header */}
            <header className="h-16 flex items-center justify-between px-6 border-b border-white/10 bg-[#12121a] flex-shrink-0 z-20">
                <div className="flex items-center gap-4">
                    <button onClick={() => navigate('/admin/dr')} className="text-white/60 hover:text-white">
                        ← Back
                    </button>
                    <h1 className="text-lg font-semibold">DR Testing</h1>
                </div>
                <div className="flex gap-3">
                    <button
                        onClick={() => navigate('/admin/dr/tests/settings')}
                        className="px-4 py-2 bg-white/5 hover:bg-white/10 border border-white/10 rounded-lg text-sm"
                    >
                        ⚙️ Settings
                    </button>
                    <button
                        onClick={() => navigate('/admin/dr/tests/create')}
                        className="px-4 py-2 bg-gradient-to-r from-emerald-500 to-cyan-500 hover:opacity-90 rounded-lg text-sm font-medium"
                    >
                        + New Test
                    </button>
                </div>
            </header>

            {/* Stats Grid */}
            <div className="p-6 pb-4">
                <div className="grid grid-cols-4 gap-4">
                    <div className="bg-[#12121a] border border-white/10 rounded-xl p-4">
                        <p className="text-white/50 text-xs mb-1">Total Tests</p>
                        <p className="text-2xl font-bold">{metrics?.total_tests || 0}</p>
                    </div>
                    <div className="bg-[#12121a] border border-white/10 rounded-xl p-4">
                        <p className="text-white/50 text-xs mb-1">Success Rate</p>
                        <p className="text-2xl font-bold text-green-400">{metrics?.success_rate?.toFixed(1) || 0}%</p>
                    </div>
                    <div className="bg-[#12121a] border border-white/10 rounded-xl p-4">
                        <p className="text-white/50 text-xs mb-1">Avg Recovery Time</p>
                        <p className="text-2xl font-bold">{metrics?.avg_recovery_time || '—'}</p>
                    </div>
                    <div className="bg-[#12121a] border border-white/10 rounded-xl p-4">
                        <p className="text-white/50 text-xs mb-1">Next Scheduled</p>
                        <p className="text-lg font-medium text-cyan-400">{metrics?.next_scheduled ? new Date(metrics.next_scheduled).toLocaleDateString() : '—'}</p>
                    </div>
                </div>
            </div>

            {/* Table */}
            <div className="flex-1 overflow-hidden px-6 pb-6">
                <div className="bg-[#12121a] rounded-xl border border-white/10 flex flex-col h-full overflow-hidden">
                    <div className="grid grid-cols-7 gap-4 px-4 py-3 border-b border-white/10 text-xs text-white/50 font-medium">
                        <span>Test ID</span>
                        <span>Name</span>
                        <span>Type</span>
                        <span>Target</span>
                        <span>Status</span>
                        <span>Progress</span>
                        <span>Actions</span>
                    </div>
                    <div className="flex-1 overflow-auto">
                        {loading ? (
                            <div className="flex items-center justify-center h-32 text-white/50">Loading...</div>
                        ) : error ? (
                            <div className="flex items-center justify-center h-32 text-red-400">{error}</div>
                        ) : tests.length === 0 ? (
                            <div className="flex items-center justify-center h-32 text-white/50">
                                No DR tests found. Create your first test to get started.
                            </div>
                        ) : (
                            tests.map((test) => (
                                <div
                                    key={test.id}
                                    className="grid grid-cols-7 gap-4 px-4 py-3 border-b border-white/5 hover:bg-white/5 cursor-pointer items-center"
                                    onClick={() => navigate(`/admin/dr/tests/${test.id}`)}
                                >
                                    <span className="text-sm font-mono text-cyan-400">{test.test_id}</span>
                                    <span className="text-sm truncate">{test.name}</span>
                                    <span className="text-sm">{getTypeIcon(test.test_type)} {test.test_type}</span>
                                    <span className="text-sm text-white/70">{test.target_env}</span>
                                    <span className={`text-xs px-2 py-1 rounded-full inline-flex items-center gap-1 w-fit ${getStatusColor(test.status)}`}>
                                        {test.status === 'RUNNING' && <span className="animate-pulse">●</span>}
                                        {test.status}
                                    </span>
                                    <div className="flex items-center gap-2">
                                        <div className="flex-1 bg-white/10 rounded-full h-2 overflow-hidden">
                                            <div
                                                className="h-full bg-gradient-to-r from-emerald-500 to-cyan-500 transition-all"
                                                style={{ width: `${test.progress_percent}%` }}
                                            />
                                        </div>
                                        <span className="text-xs text-white/50">{test.progress_percent}%</span>
                                    </div>
                                    <div className="flex gap-2">
                                        <button
                                            onClick={(e) => {
                                                e.stopPropagation();
                                                navigate(`/admin/dr/tests/${test.id}`);
                                            }}
                                            className="text-xs px-2 py-1 bg-white/5 hover:bg-white/10 rounded"
                                        >
                                            View
                                        </button>
                                    </div>
                                </div>
                            ))
                        )}
                    </div>
                </div>
            </div>

            {/* Footer */}
            <footer className="bg-[#12121a] border-t border-white/10 px-6 py-2 flex items-center justify-between text-[10px] text-white/40 flex-shrink-0">
                <span>DR Testing Module</span>
                <span>Last refresh: {new Date().toLocaleTimeString()}</span>
            </footer>
        </div>
    );
};

export default DRTestListView;
