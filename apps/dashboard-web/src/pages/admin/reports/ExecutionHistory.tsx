import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useUIStrings } from "../../../hooks/useUIStrings";
import apiClient from "../../../services/apiClient";

interface Execution {
  id: string;
  report_id: string;
  started_at: string;
  duration_seconds?: number;
  status: string;
  file_url?: string;
  error_message?: string;
}

const ExecutionHistory: React.FC = () => {
  const navigate = useNavigate();
  const [executions, setExecutions] = useState<Execution[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const { uiStrings } = useUIStrings("admin_reports");

  const loadExecutions = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = (await apiClient.get("/admin/reports/executions", {
        params: { page: 1, limit: 100 },
      })) as any;
      setExecutions(data?.executions || []);
    } catch (err: any) {
      console.error("Failed to fetch executions", err);
      setError(
        err?.message ||
          uiStrings?.msg_no_executions ||
          "Failed to load execution history",
      );
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadExecutions();
  }, []);

  const formatDuration = (seconds?: number) => {
    if (!seconds) return "-";
    if (seconds < 60) return `${seconds}s`;
    const mins = Math.floor(seconds / 60);
    return `${mins}m ${seconds % 60}s`;
  };

  return (
    <div className="flex h-screen w-full bg-background-dark text-white">
      <main className="flex-1 flex flex-col min-w-0 overflow-hidden">
        <header className="h-16 flex items-center justify-between px-6 border-b border-border-color bg-surface-darker flex-shrink-0 z-20">
          <div className="flex flex-col">
            <nav className="flex items-center gap-2 text-[10px] text-text-muted uppercase tracking-widest mb-1">
              <span
                onClick={() => navigate("/admin/reports")}
                className="hover:text-primary transition-colors cursor-pointer"
              >
                {uiStrings?.breadcrumb_scheduled || "Schedules"}
              </span>
              <span className="material-symbols-outlined text-[12px]">
                chevron_right
              </span>
              <span className="text-white">
                {uiStrings?.breadcrumb_history || "Execution History"}
              </span>
            </nav>
            <h2 className="text-white text-lg font-bold leading-tight">
              {uiStrings?.title_history || "Report Execution History"}
            </h2>
          </div>
          <button
            onClick={() => void loadExecutions()}
            className="text-text-muted hover:text-white transition-colors"
          >
            <span className="material-symbols-outlined text-[18px]">
              refresh
            </span>
          </button>
        </header>
        <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
          <div className="bg-surface-dark rounded-xl border border-border-color flex flex-col overflow-hidden">
            <div className="p-4 border-b border-border-color flex justify-between items-center bg-surface-darker">
              <h3 className="text-white font-semibold text-sm flex items-center gap-2">
                <span className="material-symbols-outlined text-text-muted text-[18px]">
                  format_list_bulleted
                </span>{" "}
                {uiStrings?.label_execution_logs || "Execution Logs"}
              </h3>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full text-left border-collapse">
                <thead>
                  <tr className="border-b border-border-color text-[11px] text-text-muted uppercase tracking-wider bg-surface-darker/50">
                    <th className="px-4 py-3 font-semibold">
                      {uiStrings?.th_run_id || "Run ID"}
                    </th>
                    <th className="px-4 py-3 font-semibold">
                      {uiStrings?.th_start_time || "Start Time"}
                    </th>
                    <th className="px-4 py-3 font-semibold">
                      {uiStrings?.th_duration || "Duration"}
                    </th>
                    <th className="px-4 py-3 font-semibold">
                      {uiStrings?.th_status || "Status"}
                    </th>
                    <th className="px-4 py-3 font-semibold">
                      {uiStrings?.th_download || "Download"}
                    </th>
                  </tr>
                </thead>
                <tbody className="text-sm">
                  {loading ? (
                    <tr>
                      <td
                        colSpan={5}
                        className="px-4 py-4 text-center text-text-muted"
                      >
                        {uiStrings?.loading || "Loading..."}
                      </td>
                    </tr>
                  ) : error ? (
                    <tr>
                      <td
                        colSpan={5}
                        className="px-4 py-4 text-center text-danger"
                      >
                        {error}
                      </td>
                    </tr>
                  ) : (
                    executions.map((exec) => (
                      <tr
                        key={exec.id}
                        className="border-b border-border-color hover:bg-surface-darker/50"
                      >
                        <td className="px-4 py-4 text-white font-mono text-xs">
                          {exec.id.substring(0, 8)}...
                        </td>
                        <td className="px-4 py-4 text-text-muted text-xs">
                          {new Date(exec.started_at).toLocaleString()}
                        </td>
                        <td className="px-4 py-4 text-text-muted text-xs">
                          {formatDuration(exec.duration_seconds)}
                        </td>
                        <td className="px-4 py-4">
                          <span
                            className={`px-2 py-0.5 rounded text-[10px] font-bold border ${exec.status === "SUCCESS" ? "bg-success/10 text-success border-success/20" : exec.status === "FAILED" ? "bg-danger/10 text-danger border-danger/20" : "bg-primary/10 text-primary border-primary/20"}`}
                          >
                            {exec.status}
                          </span>
                        </td>
                        <td className="px-4 py-4">
                          {exec.file_url ? (
                            <a
                              className="text-primary text-xs hover:underline"
                              href={exec.file_url}
                              target="_blank"
                              rel="noreferrer"
                            >
                              {uiStrings?.fmt_download || "Download"}
                            </a>
                          ) : (
                            <span className="text-text-muted italic text-xs">
                              {uiStrings?.fmt_unavailable || "Unavailable"}
                            </span>
                          )}
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
};

export default ExecutionHistory;
