import React, { useEffect, useState } from 'react';
import { monitoringService, SLAData } from '../../services/monitoringService';

const SLAReports: React.FC = () => {
  const [reports, setReports] = useState<SLAData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadReports = async () => {
      setLoading(true);
      setError(null);
      try {
        const data = await monitoringService.getSLA();
        setReports(data);
      } catch (err: any) {
        setError(err?.message || 'Failed to load SLA reports');
      } finally {
        setLoading(false);
      }
    };

    loadReports();
  }, []);

  const heatmap = Array.isArray(reports?.heatmap) ? reports!.heatmap : [];
  const maxHeatValue = heatmap.length > 0 ? Math.max(...heatmap, 1) : 1;
  const errorBudget = reports
    ? Number.parseFloat((reports.error_budget_remaining || '0').replace('%', '').trim()) || 0
    : 0;
  const resolutionRate = Math.max(0, Math.min(100, Math.round(100 - errorBudget)));

  return (
    <main className="flex-1 flex flex-col min-w-0 bg-[#0b1219] overflow-hidden">
      <header className="h-16 flex items-center justify-between px-6 border-b border-[#233648] bg-[#111a22] flex-shrink-0 z-20">
        <div className="flex items-center gap-4">
          <div>
            <div className="flex items-center gap-2 text-[10px] text-[#92adc9] uppercase font-bold tracking-wider mb-0.5">
              <span>Support</span>
              <span className="material-symbols-outlined text-[10px]">chevron_right</span>
              <span className="text-white">SLA Reports</span>
            </div>
            <h2 className="text-white text-lg font-bold">Service Level Agreement (SLA) Performance</h2>
          </div>
        </div>
        <div className="flex items-center gap-3">
            <button className="px-4 py-2 bg-[#16202a] border border-[#233648] text-white text-xs font-semibold rounded flex items-center gap-2 hover:bg-[#1c2a38] transition-colors">
                <span className="material-symbols-outlined text-[18px]">download</span>
                Export PDF
            </button>
        </div>
      </header>
      
      <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
        {error && (
          <div className="mb-4 rounded-lg border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-300">
            {error}
          </div>
        )}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
            <div className="bg-[#16202a] rounded-xl border border-[#233648] p-5">
                <div className="text-[#92adc9] text-xs font-medium uppercase mb-1">Fleet Uptime</div>
                <div className="text-2xl font-bold text-emerald-500">{loading ? '--' : `${reports?.global_uptime_90d ?? 0}%`}</div>
                <div className="text-[10px] text-[#92adc9] mt-1">Target: 99.9%</div>
            </div>
            <div className="bg-[#16202a] rounded-xl border border-[#233648] p-5">
                <div className="text-[#92adc9] text-xs font-medium uppercase mb-1">Avg. First Response</div>
                <div className="text-2xl font-bold text-white">{loading ? '--' : `${reports?.mttr_minutes ?? 0}m`}</div>
                <div className="text-[10px] text-emerald-500 mt-1">Within limit</div>
            </div>
            <div className="bg-[#16202a] rounded-xl border border-[#233648] p-5">
                <div className="text-[#92adc9] text-xs font-medium uppercase mb-1">Resolution Rate</div>
                <div className="text-2xl font-bold text-white">{loading ? '--' : `${resolutionRate}%`}</div>
                <div className="text-[10px] text-[#92adc9] mt-1">Last 30 days</div>
            </div>
            <div className="bg-[#16202a] rounded-xl border border-[#233648] p-5">
                <div className="text-[#92adc9] text-xs font-medium uppercase mb-1">SLA Violations</div>
                <div className="text-2xl font-bold text-emerald-500">{loading ? '--' : reports?.active_incidents ?? 0}</div>
                <div className="text-[10px] text-[#92adc9] mt-1">Current period</div>
            </div>
        </div>

        <div className="bg-[#16202a] rounded-xl border border-[#233648] p-6 mb-8">
            <h3 className="text-white font-semibold text-sm mb-6 flex items-center gap-2">
                <span className="material-symbols-outlined text-primary">analytics</span>
                Weekly Compliance Trend
            </h3>
            <div className="h-48 w-full bg-[#111a22] rounded-lg border border-[#233648] p-4">
                {loading ? (
                    <div className="h-full flex items-center justify-center text-[#92adc9] text-sm italic">Loading trend...</div>
                ) : heatmap.length === 0 ? (
                    <div className="h-full flex items-center justify-center text-[#92adc9] text-sm italic">No trend points available.</div>
                ) : (
                    <div className="h-full flex items-end gap-2">
                        {heatmap.map((value, idx) => (
                            <div key={idx} className="flex-1 bg-[#233648] rounded-sm overflow-hidden">
                                <div
                                  className="w-full bg-primary/80 rounded-sm transition-all"
                                  style={{ height: `${Math.max(8, Math.round((value / maxHeatValue) * 100))}%` }}
                                />
                            </div>
                        ))}
                    </div>
                )}
            </div>
        </div>

        <div className="bg-[#16202a] rounded-xl border border-[#233648] overflow-hidden">
            <div className="p-4 border-b border-[#233648] bg-[#111a22]">
                <h3 className="text-white font-semibold text-sm">SLA Snapshot</h3>
            </div>
            <table className="w-full text-left border-collapse">
                <thead>
                    <tr className="border-b border-[#233648] text-xs text-[#92adc9] uppercase tracking-wider bg-[#111a22]/50">
                        <th className="px-6 py-3">Metric</th>
                        <th className="px-6 py-3">Value</th>
                        <th className="px-6 py-3">Threshold</th>
                        <th className="px-6 py-3 text-right">Status</th>
                    </tr>
                </thead>
                <tbody className="text-sm">
                    <tr className="border-b border-[#233648] hover:bg-[#111a22]/50 transition-colors">
                        <td className="px-6 py-4 text-white font-medium">Global Uptime (90d)</td>
                        <td className="px-6 py-4 text-emerald-500 font-mono">{reports?.global_uptime_90d ?? 0}%</td>
                        <td className="px-6 py-4 text-[#92adc9]">99.9%</td>
                        <td className="px-6 py-4 text-right">
                            <span className="px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-500 border border-emerald-500/20 text-[10px] font-bold">{(reports?.global_uptime_90d ?? 0) >= 99.9 ? 'ON TARGET' : 'DEGRADED'}</span>
                        </td>
                    </tr>
                    <tr className="border-b border-[#233648] hover:bg-[#111a22]/50 transition-colors">
                        <td className="px-6 py-4 text-white font-medium">MTTR</td>
                        <td className="px-6 py-4 text-emerald-500 font-mono">{reports?.mttr_minutes ?? 0}m</td>
                        <td className="px-6 py-4 text-[#92adc9]">&lt; 15m</td>
                        <td className="px-6 py-4 text-right">
                            <span className="px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-500 border border-emerald-500/20 text-[10px] font-bold">{(reports?.mttr_minutes ?? 0) <= 15 ? 'ON TARGET' : 'AT RISK'}</span>
                        </td>
                    </tr>
                    <tr className="border-b border-[#233648] hover:bg-[#111a22]/50 transition-colors">
                        <td className="px-6 py-4 text-white font-medium">Error Budget Remaining</td>
                        <td className="px-6 py-4 text-amber-500 font-mono">{reports?.error_budget_remaining ?? '0%'}</td>
                        <td className="px-6 py-4 text-[#92adc9]">&gt; 20%</td>
                        <td className="px-6 py-4 text-right">
                            <span className="px-2 py-0.5 rounded bg-amber-500/10 text-amber-500 border border-amber-500/20 text-[10px] font-bold">
                              {Number((reports?.error_budget_remaining || '0').replace('%', '')) > 20 ? 'HEALTHY' : 'WATCH'}
                            </span>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>
      </div>
    </main>
  );
};

export default SLAReports;
