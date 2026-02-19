import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import apiClient from '../../../services/apiClient';

interface ScheduledReport {
  id: string;
  name: string;
  report_type: string;
  frequency: string;
  next_run?: string;
  status: string;
  format: string;
}

const ScheduledReportsList: React.FC = () => {
    const navigate = useNavigate();
    const [reports, setReports] = useState<ScheduledReport[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const loadReports = async () => {
        setLoading(true);
        setError(null);
        try {
            const data = await apiClient.get('/admin/reports/schedules', {
                params: { page: 1, limit: 100 },
            }) as any;
            setReports(data?.data || []);
        } catch (err: any) {
            console.error('Failed to fetch reports', err);
            setError(err?.message || 'Failed to load schedules');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        void loadReports();
    }, []);

    const toggleReportStatus = async (report: ScheduledReport) => {
        const action = report.status === 'ACTIVE' ? 'pause' : 'resume';
        try {
            await apiClient.post(`/admin/reports/schedules/${report.id}/${action}`);
            await loadReports();
        } catch (err: any) {
            console.error(`Failed to ${action} report`, err);
            setError(err?.message || `Failed to ${action} schedule`);
        }
    };

    const getStatusClass = (status: string) => {
        if (status === 'ACTIVE') return 'bg-success/10 text-success border-success/20';
        if (status === 'PAUSED') return 'bg-warning/10 text-warning border-warning/20';
        return 'bg-surface-darker text-text-muted border-border-color';
    };

    return (
        <div className="flex h-screen w-full bg-background-dark text-white">
            {/* Sidebar would be here, assuming it's in a layout wrapper */}
            <main className="flex-1 flex flex-col min-w-0 overflow-hidden">
                <header className="h-16 flex items-center justify-between px-6 border-b border-border-color bg-surface-darker flex-shrink-0 z-20">
                    <div className="flex flex-col">
                        <nav className="flex text-[10px] font-medium text-text-muted mb-0.5 gap-1 items-center">
                            <span>Reports</span>
                            <span className="material-symbols-outlined text-[10px]">chevron_right</span>
                            <span className="text-primary">Scheduled Reports</span>
                        </nav>
                        <h2 className="text-white text-lg font-bold">Scheduled Reports</h2>
                    </div>
                    <div className="flex items-center gap-4">
                        <div className="relative">
                            <span className="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-text-muted text-[18px]">search</span>
                            <input className="bg-surface-dark border border-border-color text-xs text-white rounded-lg pl-10 pr-4 py-2 w-64 focus:ring-1 focus:ring-primary focus:border-primary outline-none" placeholder="Global Search..." type="text"/>
                        </div>
                        <button onClick={() => navigate('/admin/reports/create')} className="px-4 py-2 bg-primary hover:bg-primary-dark text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors shadow-lg shadow-primary/20">
                            <span className="material-symbols-outlined text-[16px]">add</span>
                            Add Scheduled Report
                        </button>
                    </div>
                </header>
                <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
                    <div className="bg-surface-dark rounded-xl border border-border-color flex flex-col overflow-hidden">
                        <div className="p-4 border-b border-border-color flex justify-between items-center bg-surface-darker/30">
                            <div className="flex items-center gap-3">
                                <span className="text-xs font-medium text-text-muted">Showing {reports.length} schedules</span>
                            </div>
                            <div className="flex items-center gap-2">
                                <button onClick={() => void loadReports()} className="p-1.5 hover:bg-border-color rounded text-text-muted hover:text-white">
                                    <span className="material-symbols-outlined text-[18px]">refresh</span>
                                </button>
                            </div>
                        </div>
                        <div className="overflow-x-auto">
                            <table className="w-full text-left border-collapse">
                                <thead>
                                    <tr className="border-b border-border-color text-[11px] text-text-muted uppercase tracking-wider bg-surface-darker/50">
                                        <th className="px-6 py-4 font-semibold">Report Name</th>
                                        <th className="px-6 py-4 font-semibold">Type</th>
                                        <th className="px-6 py-4 font-semibold">Frequency</th>
                                        <th className="px-6 py-4 font-semibold">Format</th>
                                        <th className="px-6 py-4 font-semibold">Status</th>
                                        <th className="px-6 py-4 font-semibold text-right">Actions</th>
                                    </tr>
                                </thead>
                                <tbody className="text-sm">
                                    {loading ? (
                                        <tr><td colSpan={6} className="px-6 py-8 text-center text-text-muted">Loading...</td></tr>
                                    ) : error ? (
                                        <tr><td colSpan={6} className="px-6 py-8 text-center text-danger">{error}</td></tr>
                                    ) : reports.map((row) => (
                                        <tr key={row.id} className="border-b border-border-color hover:bg-surface-darker/50 transition-colors group cursor-pointer" onClick={() => navigate(`/admin/reports/${row.id}`)}>
                                            <td className="px-6 py-4 font-medium text-white">{row.name}</td>
                                            <td className="px-6 py-4 text-text-muted text-xs">{row.report_type}</td>
                                            <td className="px-6 py-4">
                                                <div className="flex items-center gap-2">
                                                    <span className="material-symbols-outlined text-[14px] text-primary">event_repeat</span>
                                                    <span>{row.frequency}</span>
                                                </div>
                                            </td>
                                            <td className="px-6 py-4 text-text-muted text-xs uppercase">{row.format}</td>
                                            <td className="px-6 py-4">
                                                <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded text-[10px] font-bold border ${getStatusClass(row.status)}`}>
                                                    {row.status}
                                                </span>
                                            </td>
                                            <td className="px-6 py-4 text-right">
                                                <div className="flex justify-end gap-2">
                                                    <button onClick={(e) => { e.stopPropagation(); navigate('/admin/reports/history'); }} className="text-text-muted hover:text-primary transition-colors">
                                                        <span className="material-symbols-outlined text-[18px]">history</span>
                                                    </button>
                                                    <button onClick={(e) => { e.stopPropagation(); void toggleReportStatus(row); }} className="text-text-muted hover:text-white transition-colors text-xs px-2 py-1 border border-border-color rounded">
                                                        {row.status === 'ACTIVE' ? 'Pause' : 'Resume'}
                                                    </button>
                                                </div>
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

export default ScheduledReportsList;
