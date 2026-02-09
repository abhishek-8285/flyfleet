import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';

const IncidentManagement: React.FC = () => {
  const [incidents, setIncidents] = useState<any[]>([]);
  const navigate = useNavigate();

  useEffect(() => {
    const token = localStorage.getItem('token');
    fetch('http://localhost:8081/api/support/incidents', {
      headers: { 'Authorization': `Bearer ${token}` }
    })
      .then(res => res.ok ? res.json() : [])
      .then(data => Array.isArray(data) ? setIncidents(data) : setIncidents([]))
      .catch(() => setIncidents([]));
  }, []);

  return (
    <main className="flex-1 flex flex-col min-w-0 bg-[#0b1219] overflow-hidden">
      <header className="h-16 flex items-center justify-between px-6 border-b border-[#233648] bg-[#111a22] flex-shrink-0 z-20">
        <div className="flex items-center gap-4">
          <div>
            <div className="flex items-center gap-2 text-[10px] text-[#92adc9] uppercase font-bold tracking-wider mb-0.5">
              <span>Support</span>
              <span className="material-symbols-outlined text-[10px]">chevron_right</span>
              <span className="text-white">Incidents</span>
            </div>
            <h2 className="text-white text-lg font-bold">Incident Management</h2>
          </div>
        </div>
      </header>
      
      <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
        <div className="bg-[#16202a] rounded-xl border border-[#233648] flex flex-col overflow-hidden">
          <div className="p-4 border-b border-[#233648] flex justify-between items-center bg-[#111a22]">
            <h3 className="text-white font-semibold text-sm">Active Incidents</h3>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse">
              <thead>
                <tr className="border-b border-[#233648] text-xs text-[#92adc9] uppercase tracking-wider bg-[#111a22]/50">
                  <th className="px-4 py-3 font-semibold">Incident ID</th>
                  <th className="px-4 py-3 font-semibold">Priority</th>
                  <th className="px-4 py-3 font-semibold">Status</th>
                  <th className="px-4 py-3 font-semibold">Subject</th>
                  <th className="px-4 py-3 font-semibold">Organization</th>
                </tr>
              </thead>
              <tbody className="text-sm">
                {incidents.map((incident) => (
                  <tr 
                    key={incident.id} 
                    onClick={() => navigate(`/support/incidents/${incident.id}`)}
                    className="border-b border-[#233648] hover:bg-[#111a22] transition-all cursor-pointer group hover:shadow-[inset_4px_0_0_0_#137fec]"
                  >
                    <td className="px-4 py-3 font-mono text-xs text-[#137fec] font-medium group-hover:pl-6 transition-all">{incident.id}</td>
                    <td className="px-4 py-3">
                      <span className={`inline-flex items-center px-2 py-0.5 rounded text-[10px] font-bold border ${
                        incident.priority === 'CRITICAL' ? 'bg-rose-500/10 text-rose-500 border-rose-500/20 shadow-[0_0_10px_rgba(244,63,94,0.1)]' : 'bg-amber-500/10 text-amber-500 border-amber-500/20 shadow-[0_0_10px_rgba(245,158,11,0.1)]'
                      }`}>
                        {incident.priority}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <span className={`inline-flex items-center px-2 py-0.5 rounded text-[10px] font-bold border ${
                        incident.status === 'Open' || incident.status === 'OPEN' 
                        ? 'bg-blue-500/10 text-blue-400 border-blue-500/20' 
                        : incident.status === 'RESOLVED' 
                        ? 'bg-emerald-500/10 text-emerald-500 border-emerald-500/20'
                        : 'bg-zinc-500/10 text-zinc-400 border-zinc-500/20'
                      }`}>
                        {incident.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-white font-medium group-hover:text-[#137fec] transition-colors">{incident.subject}</td>
                    <td className="px-4 py-3 text-[#92adc9] text-xs">{incident.organization}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </main>
  );
};

export default IncidentManagement;
