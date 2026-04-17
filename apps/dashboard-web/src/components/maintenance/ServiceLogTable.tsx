import React, { useEffect, useState } from "react";

export const ServiceLogTable: React.FC = () => {
  const [logs, setLogs] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch("/api/v1/ui/misc/service_logs")
      .then((res) => res.json())
      .then((data) => {
        setLogs(Array.isArray(data) ? data : data.items || []);
        setLoading(false);
      })
      .catch((err) => {
        console.error(err);
        setLoading(false);
      });
  }, []);

  return (
    <div className="bg-surface-container rounded-[3rem] border border-outline overflow-hidden flex flex-col h-full shadow-2xl relative group">
      <div className="p-10 border-b border-outline flex justify-between items-center bg-surface-container/40 backdrop-blur-xl">
        <div className="space-y-1">
          <nav className="text-[10px] font-black uppercase tracking-[0.3em] text-primary flex items-center gap-2">
            <span className="h-1 w-1 rounded-full bg-primary" />
            Asset Lifecycle Ledger
          </nav>
          <h3 className="text-2xl font-black text-white uppercase tracking-tight">Health Status Registry</h3>
        </div>
        <div className="flex gap-4">
          <button className="h-10 px-6 bg-surface-dim border border-outline text-[10px] font-black uppercase tracking-widest text-white hover:bg-surface-container-highest rounded-xl transition-all active:scale-95 flex items-center gap-3">
            <span className="material-symbols-outlined text-lg">download</span>
            Export Dataset
          </button>
        </div>
      </div>
      <div className="overflow-x-auto custom-scrollbar flex-1">
        <table className="w-full text-left text-sm text-text-muted">
          <thead className="bg-surface-dim border-b border-outline sticky top-0 z-10">
            <tr>
              <th className="px-10 py-6 text-[10px] font-black uppercase tracking-[0.2em] text-white">
                <div className="flex items-center gap-2">
                  Health Identifier
                  <span className="material-symbols-outlined text-sm opacity-30">sort</span>
                </div>
              </th>
              <th className="px-10 py-6 text-[10px] font-black uppercase tracking-[0.2em]">Operational Status</th>
              <th className="px-10 py-6 text-[10px] font-black uppercase tracking-[0.2em]">
                <div className="flex items-center gap-2">
                  Service Window
                  <span className="material-symbols-outlined text-sm opacity-30">clinical_notes</span>
                </div>
              </th>
              <th className="px-10 py-6 text-[10px] font-black uppercase tracking-[0.2em]">Telemetry (KM)</th>
              <th className="px-10 py-6 text-[10px] font-black uppercase tracking-[0.2em]">Active Incidents</th>
              <th className="px-10 py-6 text-right text-[10px] font-black uppercase tracking-[0.2em]">Management</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border-color">
            {loading ? (
              <tr>
                <td colSpan={6} className="px-10 py-24 text-center">
                  <div className="w-10 h-10 rounded-full border-2 border-primary/20 border-t-primary animate-spin mx-auto mb-4" />
                  <p className="text-[10px] font-black uppercase tracking-widest text-on-surface-variant">Syncing Logs...</p>
                </td>
              </tr>
            ) : logs.length === 0 ? (
              <tr>
                <td colSpan={6} className="px-10 py-32 text-center">
                   <div className="flex flex-col items-center gap-4 opacity-20">
                    <span className="material-symbols-outlined text-5xl">inventory</span>
                    <p className="text-[10px] font-black uppercase tracking-[0.3em]">No maintenance vectors found.</p>
                  </div>
                </td>
              </tr>
            ) : (
              logs.map((log: any, index: number) => (
                <tr
                  key={log.id || index}
                  className="hover:bg-surface-container-highest/40 transition-all duration-300 group"
                >
                  <td className="px-10 py-6">
                    <div className="flex items-center gap-4">
                      <div className="h-10 w-10 bg-surface-dim border border-outline rounded-xl flex items-center justify-center text-primary shadow-inner group-hover:scale-110 transition-transform">
                        <span className="material-symbols-outlined text-xl">
                          {log.icon || "local_shipping"}
                        </span>
                      </div>
                      <div>
                        <div className="text-white font-bold tracking-tight text-base mb-0.5 uppercase">
                          {log.vehicle_name}
                        </div>
                        <div className="text-[9px] text-on-surface-variant font-mono font-black tracking-widest uppercase">VIN: {log.vin}</div>
                      </div>
                    </div>
                  </td>
                  <td className="px-10 py-6">
                    <span
                      className={`inline-flex items-center px-4 py-1 rounded-lg text-[9px] font-black uppercase tracking-widest border ${log.status === "Critical" ? "bg-rose-500/10 text-rose-400 border-rose-500/20" : log.status === "Maintenance Due" ? "bg-amber-500/10 text-amber-500 border-amber-500/20" : "bg-emerald-500/10 text-emerald-500 border-emerald-500/20"}`}
                    >
                      <span
                        className={`w-1.5 h-1.5 rounded-full mr-2 ${log.status === "Critical" ? "bg-rose-500 animate-pulse" : log.status === "Maintenance Due" ? "bg-amber-500" : "bg-emerald-500"}`}
                      ></span>
                      {log.status || "Operational"}
                    </span>
                  </td>
                  <td className="px-10 py-6">
                    <div className="text-white font-mono font-bold text-sm tracking-tight mb-0.5">
                      {log.service_date}
                    </div>
                    <div
                      className={`text-[9px] font-black uppercase tracking-widest ${log.status === "Critical" ? "text-rose-400" : log.status === "Maintenance Due" ? "text-amber-500" : "text-on-surface-variant"}`}
                    >
                      {log.service_status}
                    </div>
                  </td>
                  <td className="px-10 py-6">
                    <div className="text-white font-mono font-bold text-sm tracking-tighter uppercase">{log.odometer || '---'}</div>
                  </td>
                  <td className="px-10 py-6">
                    {log.open_tickets > 0 ? (
                      <div className="space-y-1">
                        <div className="flex items-center gap-2 text-primary font-black text-[9px] uppercase tracking-widest">
                          <span className="material-symbols-outlined text-base">confirmation_number</span>
                          {log.open_tickets} Active Issues
                        </div>
                        <div className="text-[9px] text-on-surface-variant opacity-60 font-medium">
                          REF: {log.ticket_details}
                        </div>
                      </div>
                    ) : (
                      <span className="text-[9px] font-black uppercase tracking-widest text-on-surface-variant opacity-20">Clearance Secured</span>
                    )}
                  </td>
                  <td className="px-10 py-6 text-right">
                    <button className="h-10 px-5 bg-surface-dim border border-outline text-[9px] font-black uppercase tracking-widest text-white hover:bg-white hover:text-black rounded-xl transition-all active:scale-95 shadow-md">
                      Audit Assets
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
      <div className="bg-surface-dim px-10 py-6 border-t border-outline flex items-center justify-between">
        <span className="text-[10px] font-black uppercase tracking-widest text-on-surface-variant">
          Inventory Trace:{" "}
          <span className="text-white">
            1-{Math.min(5, logs.length)}
          </span>{" "}
          / <span className="text-white">{logs.length}</span>{" "}
          Active Records
        </span>
        <div className="flex gap-3">
          <button className="h-9 px-4 text-[9px] font-black uppercase tracking-widest text-on-surface-variant bg-surface-container border border-outline rounded-lg hover:text-white transition-all">
            Previous
          </button>
          <button className="h-9 px-4 text-[9px] font-black uppercase tracking-widest text-white bg-primary rounded-lg shadow-md shadow-primary/20">
            01
          </button>
          <button className="h-9 px-4 text-[9px] font-black uppercase tracking-widest text-on-surface-variant bg-surface-container border border-outline rounded-lg hover:text-white transition-all">
            Next
          </button>
        </div>
      </div>
    </div>
  );
};
