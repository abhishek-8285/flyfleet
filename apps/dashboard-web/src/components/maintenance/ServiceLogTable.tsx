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
    <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden flex flex-col h-full">
      <div className="p-4 border-b border-border-color flex justify-between items-center bg-surface-darker">
        <h3 className="text-white font-semibold">Service Log Entries</h3>
        <div className="flex gap-3">
          <button className="text-xs font-medium text-text-muted hover:text-white flex items-center gap-1 bg-border-color/30 px-3 py-1.5 rounded transition-colors border border-border-color/50">
            <span className="material-symbols-outlined text-[16px]">
              download
            </span>
            Export CSV
          </button>
        </div>
      </div>
      <div className="overflow-x-auto custom-scrollbar flex-1">
        <table className="w-full text-left text-sm text-text-muted">
          <thead className="bg-[#1d2936] text-xs uppercase font-medium sticky top-0 z-10 shadow-sm">
            <tr>
              <th className="px-6 py-4 cursor-pointer hover:text-white group">
                <div className="flex items-center gap-1">
                  Vehicle{" "}
                  <span className="material-symbols-outlined text-[14px] text-text-muted/50">
                    sort
                  </span>
                </div>
              </th>
              <th className="px-6 py-4">Health Status</th>
              <th className="px-6 py-4 cursor-pointer hover:text-white group">
                <div className="flex items-center gap-1">
                  Upcoming Service{" "}
                  <span className="material-symbols-outlined text-[14px] text-text-muted/50">
                    arrow_upward
                  </span>
                </div>
              </th>
              <th className="px-6 py-4">Odometer</th>
              <th className="px-6 py-4">Open Tickets</th>
              <th className="px-6 py-4 text-right">Action</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border-color">
            {loading ? (
              <tr>
                <td colSpan={6} className="text-center py-8">
                  Loading logs...
                </td>
              </tr>
            ) : logs.length === 0 ? (
              <tr>
                <td colSpan={6} className="text-center py-8">
                  No service logs available.
                </td>
              </tr>
            ) : (
              logs.map((log: any, index: number) => (
                <tr
                  key={log.id || index}
                  className="hover:bg-[#1d2936] transition-colors group"
                >
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-3">
                      <div className="w-8 h-8 rounded bg-border-color flex items-center justify-center text-text-muted">
                        <span className="material-symbols-outlined text-[18px]">
                          {log.icon || "local_shipping"}
                        </span>
                      </div>
                      <div>
                        <div className="text-white font-medium">
                          {log.vehicle_name}
                        </div>
                        <div className="text-xs text-text-muted">{log.vin}</div>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <span
                      className={`inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium ${log.status === "Critical" ? "bg-red-500/10 text-red-400 border border-red-500/20" : log.status === "Maintenance Due" ? "bg-yellow-500/10 text-yellow-500 border border-yellow-500/20" : "bg-emerald-500/10 text-emerald-500 border border-emerald-500/20"}`}
                    >
                      <span
                        className={`w-1.5 h-1.5 rounded-full mr-1.5 ${log.status === "Critical" ? "bg-red-500 animate-pulse" : log.status === "Maintenance Due" ? "bg-yellow-500" : "bg-emerald-500"}`}
                      ></span>
                      {log.status || "Healthy"}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    <div className="text-white font-medium">
                      {log.service_date}
                    </div>
                    <div
                      className={`text-xs ${log.status === "Critical" ? "text-red-400" : log.status === "Maintenance Due" ? "text-yellow-500" : "text-text-muted"}`}
                    >
                      {log.service_status}
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="text-white">{log.odometer}</div>
                  </td>
                  <td className="px-6 py-4">
                    {log.open_tickets > 0 ? (
                      <>
                        <a
                          className="text-primary hover:underline hover:text-white flex items-center gap-1"
                          href="#"
                        >
                          <span className="material-symbols-outlined text-[16px]">
                            confirmation_number
                          </span>
                          {log.open_tickets} Tickets
                        </a>
                        <div className="text-[10px] text-text-muted mt-0.5">
                          {log.ticket_details}
                        </div>
                      </>
                    ) : (
                      <span className="text-text-muted/50 text-xs italic">
                        No open tickets
                      </span>
                    )}
                  </td>
                  <td className="px-6 py-4 text-right">
                    <button className="bg-border-color hover:bg-white text-text-muted hover:text-surface-darker px-3 py-1.5 rounded text-xs font-medium transition-colors">
                      View Details
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
      <div className="bg-surface-darker px-4 py-3 border-t border-border-color flex items-center justify-between">
        <span className="text-xs text-text-muted">
          Showing{" "}
          <span className="text-white font-medium">
            1-{Math.min(5, logs.length)}
          </span>{" "}
          of <span className="text-white font-medium">{logs.length}</span>{" "}
          results
        </span>
        <div className="flex gap-2">
          <button className="px-3 py-1 text-xs font-medium text-text-muted bg-surface-dark border border-border-color rounded hover:text-white hover:bg-border-color transition-colors">
            Previous
          </button>
          <button className="px-3 py-1 text-xs font-medium text-white bg-primary rounded shadow-sm">
            1
          </button>
          <button className="px-3 py-1 text-xs font-medium text-text-muted bg-surface-dark border border-border-color rounded hover:text-white hover:bg-border-color transition-colors">
            Next
          </button>
        </div>
      </div>
    </div>
  );
};
