import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';

const TroubleshootingLogs: React.FC = () => {
  const [logs, setLogs] = useState<any[]>([]);
  const navigate = useNavigate();

  useEffect(() => {
    const token = localStorage.getItem('token');
    fetch('http://localhost:8081/api/support/logs', {
      headers: { 'Authorization': `Bearer ${token}` }
    })
      .then(res => res.ok ? res.json() : [])
      .then(data => Array.isArray(data) ? setLogs(data) : setLogs([]))
      .catch(() => setLogs([]));
  }, []);

  return (
    <main className="flex-1 flex flex-col min-w-0 bg-[#0b1219] overflow-hidden">
      <header className="h-16 flex items-center justify-between px-6 border-b border-[#233648] bg-[#111a22] flex-shrink-0 z-20">
        <div className="flex items-center gap-4">
          <div>
            <h2 className="text-white text-lg font-bold">Consolidated Debugging Logs</h2>
          </div>
        </div>
      </header>
      
      <div className="flex flex-1 overflow-hidden">
        <div className="flex-1 flex flex-col min-w-0 border-r border-[#233648]">
          <div className="flex-1 overflow-y-auto custom-scrollbar bg-[#111a22]">
            <table className="w-full text-left border-collapse table-fixed">
              <thead className="sticky top-0 z-10 bg-[#16202a] border-b border-[#233648]">
                <tr className="text-[11px] text-[#92adc9] uppercase tracking-wider font-semibold">
                  <th className="px-4 py-2 w-48">Timestamp</th>
                  <th className="px-4 py-2 w-24">Severity</th>
                  <th className="px-4 py-2 w-32">Module</th>
                  <th className="px-4 py-2 w-auto">Message</th>
                </tr>
              </thead>
              <tbody className="text-[13px] font-mono divide-y divide-[#233648]/50">
                {logs.map((log, idx) => (
                  <tr 
                    key={idx} 
                    onClick={() => navigate(`/support/logs?t=${log.timestamp}`)}
                    className={`hover:bg-[#16202a] cursor-pointer transition-all border-l-2 active:bg-[#1c2a38] ${
                      log.severity === 'ERROR' ? 'border-l-rose-500 bg-rose-500/5' : log.severity === 'WARN' ? 'border-l-amber-500 bg-amber-500/5' : 'border-l-transparent hover:border-l-primary/30'
                    }`}
                  >
                    <td className="px-4 py-2 text-[#92adc9] whitespace-nowrap">{log.timestamp}</td>
                    <td className="px-4 py-2">
                       <span className={`inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-bold border ${
                         log.severity === 'ERROR' ? 'bg-rose-500/20 text-rose-500 border-rose-500/30 shadow-[0_0_10px_rgba(244,63,94,0.1)]' : 'bg-amber-500/20 text-amber-500 border-amber-500/20'
                       }`}>
                         {log.severity}
                       </span>
                    </td>
                    <td className="px-4 py-2 text-white font-medium">{log.module}</td>
                    <td className="px-4 py-2 text-[#92adc9] group-hover:text-white transition-colors">{log.message}</td>
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

export default TroubleshootingLogs;
