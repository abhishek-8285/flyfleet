import React, { useEffect, useState } from "react";
import { useUIStrings } from "../../hooks/useUIStrings";
import { LogEntry, monitoringService } from "../../services/monitoringService";
import PageHeader from "./components/PageHeader";

const LogViewer: React.FC = () => {
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const { uiStrings } = useUIStrings("monitoring");
  const [selectedLog, setSelectedLog] = useState<LogEntry | null>(null);

  const fetchData = async () => {
    setLoading(true);
    try {
      const data = await monitoringService.getLogs();
      setLogs(data.logs);
    } catch (error) {
      console.error("Failed to fetch logs", error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  return (
    <main className="flex-1 flex flex-col min-w-0 bg-monitor-bg-dark overflow-hidden relative font-public-sans text-monitor-text-muted">
      <PageHeader
        title={uiStrings?.logs_title || "Error & Exception Log Viewer"}
        subtitle={
          <>
            <span>{uiStrings?.metrics_subtitle || "Monitoring"}</span>
            <span className="material-symbols-outlined text-[10px]">
              chevron_right
            </span>
            <span className="text-monitor-primary">
              {uiStrings?.logs_primary || "Error Logs"}
            </span>
          </>
        }
        actions={
          <>
            <div className="relative">
              <span className="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-monitor-text-muted text-[18px]">
                search
              </span>
              <input
                className="bg-monitor-surface border border-monitor-border text-xs text-white rounded-md pl-9 pr-4 py-2 w-64 focus:ring-1 focus:ring-monitor-primary focus:border-monitor-primary outline-none placeholder-monitor-text-muted"
                placeholder={
                  uiStrings?.logs_search_placeholder ||
                  "Global Search (ID, Msg, Trace)..."
                }
                type="text"
              />
            </div>
            <button className="px-3 py-2 bg-monitor-surface hover:bg-monitor-border border border-monitor-border text-monitor-text-muted hover:text-white text-xs font-medium rounded flex items-center gap-2 transition-colors">
              <span className="material-symbols-outlined text-[16px]">
                download
              </span>
              {uiStrings?.logs_export || "Export Logs"}
            </button>
            <button
              onClick={fetchData}
              className="p-2 bg-monitor-primary hover:bg-monitor-primary-dark text-white rounded-md transition-colors shadow-lg shadow-monitor-primary/20"
            >
              <span
                className={`material-symbols-outlined text-[20px] ${loading ? "animate-spin" : ""}`}
              >
                refresh
              </span>
            </button>
          </>
        }
      />
      <div className="flex flex-1 overflow-hidden">
        <div className="w-64 bg-monitor-surface border-r border-monitor-border flex flex-col flex-shrink-0 z-10">
          <div className="p-4 border-b border-monitor-border flex justify-between items-center">
            <h3 className="text-xs font-bold text-white uppercase tracking-wider">
              {uiStrings?.logs_filters || "Filters"}
            </h3>
            <button
              onClick={() => setSelectedLog(null)}
              className="text-[10px] text-monitor-primary hover:underline"
            >
              {uiStrings?.logs_clear_selection || "Clear Selection"}
            </button>
          </div>
          <div className="flex-1 overflow-y-auto custom-scrollbar p-4 space-y-6">
            <div className="space-y-2">
              <label className="text-xs font-medium text-monitor-text-muted">
                {uiStrings?.logs_time_range || "Time Range"}
              </label>
              <select className="w-full bg-monitor-surface-darker border border-monitor-border text-xs text-white rounded px-2 py-2 focus:ring-1 focus:ring-monitor-primary focus:border-monitor-primary outline-none">
                <option>Last 15 minutes</option>
                <option defaultValue="Last 1 Hour">Last 1 Hour</option>
                <option>Last 24 Hours</option>
              </select>
            </div>
            <div className="space-y-2">
              <label className="text-xs font-medium text-monitor-text-muted">
                {uiStrings?.logs_severity || "Severity"}
              </label>
              <div className="space-y-1">
                {["CRITICAL", "ERROR", "WARNING", "INFO"].map((sev) => (
                  <label
                    key={sev}
                    className="flex items-center gap-2 cursor-pointer hover:bg-monitor-surface-darker p-1.5 rounded transition-colors"
                  >
                    <input
                      defaultChecked
                      className={`rounded border-monitor-border bg-monitor-surface-darker text-monitor-${sev.toLowerCase()} focus:ring-0 focus:ring-offset-0`}
                      type="checkbox"
                    />
                    <span
                      className={`text-xs text-monitor-${sev.toLowerCase()} font-bold flex items-center gap-1`}
                    >
                      <span className="material-symbols-outlined text-[14px]">
                        {sev === "CRITICAL"
                          ? "dangerous"
                          : sev === "WARNING"
                            ? "warning"
                            : sev === "ERROR"
                              ? "error"
                              : "info"}
                      </span>{" "}
                      {sev}
                    </span>
                  </label>
                ))}
              </div>
            </div>
          </div>
        </div>
        <div className="flex-1 overflow-hidden flex flex-col">
          <div className="flex-1 overflow-y-auto custom-scrollbar bg-monitor-bg-dark p-0">
            <table className="w-full text-left border-collapse">
              <thead className="sticky top-0 z-10 bg-monitor-surface-darker border-b border-monitor-border shadow-sm">
                <tr>
                  <th className="px-4 py-3 text-xs font-semibold text-monitor-text-muted uppercase tracking-wider w-40">
                    {uiStrings?.logs_th_timestamp || "Timestamp"}
                  </th>
                  <th className="px-4 py-3 text-xs font-semibold text-monitor-text-muted uppercase tracking-wider w-32">
                    {uiStrings?.logs_th_service || "Service"}
                  </th>
                  <th className="px-4 py-3 text-xs font-semibold text-monitor-text-muted uppercase tracking-wider w-24">
                    {uiStrings?.logs_th_severity || "Severity"}
                  </th>
                  <th className="px-4 py-3 text-xs font-semibold text-monitor-text-muted uppercase tracking-wider w-48">
                    Error Type
                  </th>
                  <th className="px-4 py-3 text-xs font-semibold text-monitor-text-muted uppercase tracking-wider">
                    {uiStrings?.logs_th_message || "Message"}
                  </th>
                  <th className="px-4 py-3 text-xs font-semibold text-monitor-text-muted uppercase tracking-wider text-right w-20">
                    Count
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-monitor-border">
                {loading ? (
                  <tr>
                    <td
                      colSpan={6}
                      className="p-4 text-center text-monitor-text-muted"
                    >
                      {uiStrings?.logs_loading || "Loading logs..."}
                    </td>
                  </tr>
                ) : logs.length === 0 ? (
                  <tr>
                    <td
                      colSpan={6}
                      className="p-4 text-center text-monitor-text-muted"
                    >
                      {uiStrings?.logs_not_found || "No logs found."}
                    </td>
                  </tr>
                ) : (
                  logs.map((log, idx) => (
                    <tr
                      key={idx}
                      onClick={() => setSelectedLog(log)}
                      className={`group hover:bg-monitor-surface cursor-pointer transition-colors bg-monitor-surface-darker/30 border-l-2 ${selectedLog === log ? "border-l-monitor-primary bg-monitor-surface" : "border-l-transparent"}`}
                    >
                      <td className="px-4 py-2.5">
                        <div className="flex flex-col">
                          <span className="text-xs text-white font-mono">
                            {log.timestamp}
                          </span>
                        </div>
                      </td>
                      <td className="px-4 py-2.5">
                        <span className="inline-flex items-center px-2 py-0.5 rounded text-[10px] font-medium bg-monitor-surface-darker border border-monitor-border text-monitor-text-muted">
                          {log.service}
                        </span>
                      </td>
                      <td className="px-4 py-2.5">
                        <span
                          className={`inline-flex items-center gap-1 px-2 py-0.5 rounded text-[10px] font-bold ${log.severity === "CRITICAL" ? "bg-monitor-danger/10 text-monitor-danger border-monitor-danger/20" : "bg-monitor-warning/10 text-monitor-warning border-monitor-warning/20"} border`}
                        >
                          {log.severity}
                        </span>
                      </td>
                      <td
                        className={`px-4 py-2.5 text-xs font-mono truncate max-w-[12rem] ${log.severity === "CRITICAL" ? "text-monitor-danger" : "text-monitor-warning"}`}
                      >
                        {log.type}
                      </td>
                      <td className="px-4 py-2.5">
                        <div className="text-xs text-white truncate max-w-md">
                          {log.message}
                        </div>
                      </td>
                      <td className="px-4 py-2.5 text-right text-xs font-bold text-white">
                        {log.count || 1}
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
        {selectedLog && (
          <div className="w-[450px] bg-monitor-surface border-l border-monitor-border flex flex-col flex-shrink-0 z-20 shadow-xl">
            <div className="p-4 border-b border-monitor-border flex justify-between items-start">
              <div>
                <h3 className="text-sm font-bold text-white font-mono">
                  {selectedLog.type}
                </h3>
                <p className="text-xs text-monitor-text-muted mt-1">
                  {selectedLog.service} • {selectedLog.timestamp}
                </p>
              </div>
              <div className="flex gap-2">
                <span
                  className={`inline-flex items-center gap-1 px-2 py-0.5 rounded text-[10px] font-bold ${selectedLog.severity === "CRITICAL" ? "bg-monitor-danger/10 text-monitor-danger border-monitor-danger/20" : "bg-monitor-warning/10 text-monitor-warning border-monitor-warning/20"} border h-fit`}
                >
                  {selectedLog.severity}
                </span>
                <button
                  onClick={() => setSelectedLog(null)}
                  className="text-monitor-text-muted hover:text-white"
                >
                  <span className="material-symbols-outlined text-[20px]">
                    close
                  </span>
                </button>
              </div>
            </div>
            <div className="flex-1 overflow-y-auto custom-scrollbar p-5 space-y-6">
              <div className="space-y-2">
                <h4 className="text-xs font-semibold text-monitor-text-muted uppercase tracking-wider">
                  {uiStrings?.logs_th_message || "Message"}
                </h4>
                <div className="bg-monitor-bg-dark border border-monitor-border rounded p-2 font-mono text-xs text-gray-300 whitespace-pre-wrap">
                  {selectedLog.message}
                </div>
              </div>
              <div className="space-y-2">
                <h4 className="text-xs font-semibold text-monitor-text-muted uppercase tracking-wider">
                  Stacktrace
                </h4>
                <div className="bg-monitor-bg-dark border border-monitor-border rounded p-2 font-mono text-xs text-gray-300 whitespace-pre-wrap overflow-x-auto max-h-60 custom-scrollbar">
                  Stacktrace unavailable in current log level.
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </main>
  );
};

export default LogViewer;
