import React, { useEffect, useState } from 'react';

const SLAReports: React.FC = () => {
  const [reports, setReports] = useState<any>(null);

  useEffect(() => {
    // For now, using mock data as there is no specific endpoint for SLA reports yet
    // but follows the pattern of other support pages.
    setReports({
        uptime: "99.98%",
        avgResponse: "14m",
        resolutionRate: "92%",
        violations: 0
    });
  }, []);

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
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
            <div className="bg-[#16202a] rounded-xl border border-[#233648] p-5">
                <div className="text-[#92adc9] text-xs font-medium uppercase mb-1">Fleet Uptime</div>
                <div className="text-2xl font-bold text-emerald-500">{reports?.uptime}</div>
                <div className="text-[10px] text-[#92adc9] mt-1">Target: 99.9%</div>
            </div>
            <div className="bg-[#16202a] rounded-xl border border-[#233648] p-5">
                <div className="text-[#92adc9] text-xs font-medium uppercase mb-1">Avg. First Response</div>
                <div className="text-2xl font-bold text-white">{reports?.avgResponse}</div>
                <div className="text-[10px] text-emerald-500 mt-1">Within limit</div>
            </div>
            <div className="bg-[#16202a] rounded-xl border border-[#233648] p-5">
                <div className="text-[#92adc9] text-xs font-medium uppercase mb-1">Resolution Rate</div>
                <div className="text-2xl font-bold text-white">{reports?.resolutionRate}</div>
                <div className="text-[10px] text-[#92adc9] mt-1">Last 30 days</div>
            </div>
            <div className="bg-[#16202a] rounded-xl border border-[#233648] p-5">
                <div className="text-[#92adc9] text-xs font-medium uppercase mb-1">SLA Violations</div>
                <div className="text-2xl font-bold text-emerald-500">{reports?.violations}</div>
                <div className="text-[10px] text-[#92adc9] mt-1">Current period</div>
            </div>
        </div>

        <div className="bg-[#16202a] rounded-xl border border-[#233648] p-6 mb-8">
            <h3 className="text-white font-semibold text-sm mb-6 flex items-center gap-2">
                <span className="material-symbols-outlined text-primary">analytics</span>
                Weekly Compliance Trend
            </h3>
            <div className="h-48 w-full bg-[#111a22] rounded-lg border border-[#233648] flex items-center justify-center">
                <span className="text-[#92adc9] text-sm italic">Compliance Chart Rendering...</span>
            </div>
        </div>

        <div className="bg-[#16202a] rounded-xl border border-[#233648] overflow-hidden">
            <div className="p-4 border-b border-[#233648] bg-[#111a22]">
                <h3 className="text-white font-semibold text-sm">Service Availability by Region</h3>
            </div>
            <table className="w-full text-left border-collapse">
                <thead>
                    <tr className="border-b border-[#233648] text-xs text-[#92adc9] uppercase tracking-wider bg-[#111a22]/50">
                        <th className="px-6 py-3">Region</th>
                        <th className="px-6 py-3">Availability</th>
                        <th className="px-6 py-3">Incidents</th>
                        <th className="px-6 py-3 text-right">Status</th>
                    </tr>
                </thead>
                <tbody className="text-sm">
                    <tr className="border-b border-[#233648] hover:bg-[#111a22]/50 transition-colors">
                        <td className="px-6 py-4 text-white font-medium">US-East (Virginia)</td>
                        <td className="px-6 py-4 text-emerald-500 font-mono">99.99%</td>
                        <td className="px-6 py-4 text-[#92adc9]">2</td>
                        <td className="px-6 py-4 text-right">
                            <span className="px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-500 border border-emerald-500/20 text-[10px] font-bold">OPERATIONAL</span>
                        </td>
                    </tr>
                    <tr className="border-b border-[#233648] hover:bg-[#111a22]/50 transition-colors">
                        <td className="px-6 py-4 text-white font-medium">EU-West (Ireland)</td>
                        <td className="px-6 py-4 text-emerald-500 font-mono">99.98%</td>
                        <td className="px-6 py-4 text-[#92adc9]">1</td>
                        <td className="px-6 py-4 text-right">
                            <span className="px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-500 border border-emerald-500/20 text-[10px] font-bold">OPERATIONAL</span>
                        </td>
                    </tr>
                    <tr className="border-b border-[#233648] hover:bg-[#111a22]/50 transition-colors">
                        <td className="px-6 py-4 text-white font-medium">AP-South (Mumbai)</td>
                        <td className="px-6 py-4 text-amber-500 font-mono">99.92%</td>
                        <td className="px-6 py-4 text-[#92adc9]">5</td>
                        <td className="px-6 py-4 text-right">
                            <span className="px-2 py-0.5 rounded bg-amber-500/10 text-amber-500 border border-amber-500/20 text-[10px] font-bold">DEGRADED</span>
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
