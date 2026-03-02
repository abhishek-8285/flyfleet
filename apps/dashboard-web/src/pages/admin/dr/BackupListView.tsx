import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useUIStrings } from "../../../hooks/useUIStrings";
import apiClient from "../../../services/apiClient";

interface Backup {
  id: string;
  backup_type: string;
  start_time: string;
  duration_seconds?: string;
  size_bytes?: number;
  location?: string;
  status: string;
}

const BackupListView: React.FC = () => {
  const navigate = useNavigate();
  const [jobs, setJobs] = useState<Backup[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [infraMeta, setInfraMeta] = useState<any>(null);
  const { uiStrings, loading: uiLoading } = useUIStrings("dr");

  const fetchBackups = async () => {
    setLoading(true);
    try {
      setError(null);
      const data = (await apiClient.get("/admin/dr/backups")) as any;
      const items = data?.data || data?.items || data?.backups || data || [];
      setJobs(Array.isArray(items) ? items : []);
    } catch (err) {
      console.error("Failed to fetch backups", err);
      setError(
        uiStrings?.msg_load_backups_failed || "Failed to load backup jobs.",
      );
      setJobs([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchBackups();

    apiClient
      .get("/ui/global_config")
      .then((res: any) => {
        setInfraMeta(res?.infrastructure_metadata?.dr_metrics || null);
      })
      .catch((err) => console.error("Failed to fetch infra config", err));

    // Poll every 5s for updates on running jobs
    const interval = setInterval(fetchBackups, 5000);
    return () => clearInterval(interval);
  }, []);

  const getStatusBadge = (status: string) => {
    const base =
      "inline-flex items-center gap-1 px-2 py-0.5 rounded text-[10px] font-bold border";
    switch (status) {
      case "VERIFIED":
        return `${base} bg-success/10 text-success border-success/20`;
      case "COMPLETED":
        return `${base} bg-emerald-500/10 text-emerald-500 border-emerald-500/20`;
      case "RUNNING":
        return `${base} bg-primary/10 text-primary border-primary/20`;
      case "SCHEDULED":
        return `${base} bg-gray-500/10 text-text-muted border-border-color`;
      case "FAILED":
        return `${base} bg-danger/10 text-danger border-danger/20`;
      default:
        return base;
    }
  };

  const formatDuration = (seconds?: number) => {
    if (!seconds) return "---";
    if (seconds < 60) return `${seconds}s`;
    const mins = Math.floor(seconds / 60);
    return `${mins}m ${seconds % 60}s`;
  };

  const formatSize = (bytes?: number) => {
    if (!bytes) return "---";
    const gb = bytes / (1024 * 1024 * 1024);
    return `${gb.toFixed(2)} GB`;
  };

  return (
    <div className="flex flex-col h-full w-full bg-background-dark text-white font-display overflow-hidden">
      <header className="h-16 flex items-center justify-between px-6 border-b border-border-color bg-surface-darker flex-shrink-0 z-20">
        <div className="flex flex-col">
          <nav className="flex items-center gap-2 text-[10px] text-text-muted font-medium mb-1">
            <span>{uiStrings?.nav_infrastructure || "Infrastructure"}</span>
            <span className="material-symbols-outlined text-[10px]">
              chevron_right
            </span>
            <span>{uiStrings?.title_main || "Disaster Recovery"}</span>
            <span className="material-symbols-outlined text-[10px]">
              chevron_right
            </span>
            <span className="text-white">
              {uiStrings?.nav_backups || "Backups"}
            </span>
          </nav>
          <div className="flex items-center gap-3">
            <h2 className="text-white text-lg font-bold">
              {uiStrings?.title_list_view}
            </h2>
            <span className="px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider bg-primary/20 text-primary border border-primary/30">
              {uiStrings?.badge_dr_module}
            </span>
          </div>
        </div>
        <div className="flex items-center gap-4">
          <div className="relative">
            <span className="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-text-muted text-[18px]">
              search
            </span>
            <input
              className="bg-surface-dark border border-border-color text-xs text-white rounded-lg pl-10 pr-4 py-2 w-72 focus:ring-1 focus:ring-primary focus:border-primary outline-none"
              placeholder={
                uiStrings?.placeholder_search || "Search Global Backups..."
              }
              type="text"
            />
          </div>
          <button
            onClick={() => navigate("/admin/dr/create")}
            className="px-4 py-2 bg-primary hover:bg-primary-dark text-white text-xs font-semibold rounded-lg flex items-center gap-2 transition-colors shadow-lg shadow-primary/20"
          >
            <span className="material-symbols-outlined text-[16px]">
              add_box
            </span>
            {uiStrings?.btn_create_backup}
          </button>
        </div>
      </header>

      <div className="px-6 pt-6 pb-2">
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <div className="bg-surface-dark rounded-xl border border-border-color p-4">
            <div className="text-[10px] text-text-muted font-bold uppercase mb-1">
              {uiStrings?.stat_success_rate}
            </div>
            <div className="text-xl font-bold text-success">
              {infraMeta?.success_rate || "---"}
            </div>
          </div>
          <div className="bg-surface-dark rounded-xl border border-border-color p-4">
            <div className="text-[10px] text-text-muted font-bold uppercase mb-1">
              {uiStrings?.stat_active_transfers}
            </div>
            <div className="text-xl font-bold text-primary">
              {jobs.filter((j) => j.status === "RUNNING").length}
            </div>
          </div>
          <div className="bg-surface-dark rounded-xl border border-border-color p-4">
            <div className="text-[10px] text-text-muted font-bold uppercase mb-1">
              {uiStrings?.stat_total_storage}
            </div>
            <div className="text-xl font-bold text-white">
              {infraMeta?.total_storage || "---"}
            </div>
          </div>
          <div className="bg-surface-dark rounded-xl border border-border-color p-4">
            <div className="text-[10px] text-text-muted font-bold uppercase mb-1">
              {uiStrings?.stat_pending_verification}
            </div>
            <div className="text-xl font-bold text-warning">
              {jobs.filter((j) => j.status === "COMPLETED").length}
            </div>
          </div>
        </div>
      </div>

      <div className="flex-1 overflow-hidden p-6">
        <div className="bg-surface-dark rounded-xl border border-border-color flex flex-col h-full overflow-hidden">
          <div className="p-4 border-b border-border-color flex justify-between items-center bg-surface-darker/50">
            <h3 className="text-white font-semibold text-sm flex items-center gap-2">
              <span className="material-symbols-outlined text-text-muted text-[18px]">
                list_alt
              </span>
              {uiStrings?.title_ops_log}
            </h3>
            <div className="flex items-center gap-2">
              <button
                onClick={fetchBackups}
                className="flex items-center gap-2 px-3 py-1.5 hover:bg-border-color rounded text-xs text-text-muted hover:text-white transition-colors"
              >
                <span className="material-symbols-outlined text-[16px]">
                  refresh
                </span>
                {uiStrings?.btn_sync_bff}
              </button>
              <button className="p-1.5 hover:bg-border-color rounded text-text-muted hover:text-white">
                <span className="material-symbols-outlined text-[18px]">
                  download
                </span>
              </button>
            </div>
          </div>
          <div className="flex-1 overflow-auto custom-scrollbar">
            <table className="w-full text-left border-collapse min-w-[1000px]">
              <thead>
                <tr className="sticky top-0 z-10 border-b border-border-color text-[11px] text-text-muted uppercase tracking-wider bg-surface-darker shadow-sm">
                  <th className="px-6 py-4 font-semibold">
                    {uiStrings?.th_id || "Backup ID"}
                  </th>
                  <th className="px-6 py-4 font-semibold">
                    {uiStrings?.th_type || "Type"}
                  </th>
                  <th className="px-6 py-4 font-semibold">
                    {uiStrings?.th_date || "Start Time"}
                  </th>
                  <th className="px-6 py-4 font-semibold">
                    {uiStrings?.th_duration || "Duration"}
                  </th>
                  <th className="px-6 py-4 font-semibold">
                    {uiStrings?.th_size || "Size"}
                  </th>
                  <th className="px-6 py-4 font-semibold">
                    {uiStrings?.th_location || "Storage Location"}
                  </th>
                  <th className="px-6 py-4 font-semibold text-right">
                    {uiStrings?.th_status || "Status"}
                  </th>
                </tr>
              </thead>
              <tbody className="text-sm divide-y divide-border-color/30">
                {loading ? (
                  <tr>
                    <td
                      colSpan={7}
                      className="px-6 py-8 text-center text-text-muted"
                    >
                      {uiStrings?.msg_loading || "Loading..."}
                    </td>
                  </tr>
                ) : error ? (
                  <tr>
                    <td
                      colSpan={7}
                      className="px-6 py-8 text-center text-red-400"
                    >
                      {error}
                    </td>
                  </tr>
                ) : jobs.length === 0 ? (
                  <tr>
                    <td
                      colSpan={7}
                      className="px-6 py-8 text-center text-text-muted"
                    >
                      {uiStrings?.msg_no_backups_found || "No backups found."}
                    </td>
                  </tr>
                ) : (
                  jobs.map((job) => (
                    <tr
                      key={job.id}
                      onClick={() => navigate(`/admin/dr/${job.id}`)}
                      className={`hover:bg-primary/5 cursor-pointer transition-colors group ${job.status === "RUNNING" ? "bg-primary/5" : ""}`}
                    >
                      <td className="px-6 py-4 font-mono text-xs text-primary font-medium">
                        {job.id.substring(0, 8)}...
                      </td>
                      <td className="px-6 py-4">
                        <span
                          className={`px-2 py-0.5 rounded text-[10px] font-bold ${job.backup_type === "FULL" ? "bg-indigo-500/10 text-indigo-400 border border-indigo-500/20" : "bg-surface-darker text-text-muted border border-border-color"}`}
                        >
                          {job.backup_type}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-text-muted text-xs">
                        {job.start_time
                          ? new Date(job.start_time).toLocaleString()
                          : "-"}
                      </td>
                      <td
                        className={`px-6 py-4 ${job.status === "RUNNING" ? "text-primary animate-pulse" : "text-text-muted"}`}
                      >
                        {job.status === "RUNNING"
                          ? uiStrings?.msg_in_progress || "In Progress..."
                          : formatDuration(Number(job.duration_seconds))}
                      </td>
                      <td className="px-6 py-4 text-white font-medium">
                        {formatSize(job.size_bytes)}
                      </td>
                      <td className="px-6 py-4 flex items-center gap-2 text-text-muted">
                        <span className="material-symbols-outlined text-[16px]">
                          {(job.location || "").includes("AWS") ||
                          (job.location || "").includes("Azure")
                            ? "cloud"
                            : "dns"}
                        </span>
                        {job.location}
                      </td>
                      <td className="px-6 py-4 text-right">
                        <span className={getStatusBadge(job.status)}>
                          {job.status === "VERIFIED" && (
                            <span className="w-1 h-1 rounded-full bg-success"></span>
                          )}
                          {job.status}
                        </span>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <footer className="bg-surface-darker border-t border-border-color px-6 py-2 flex items-center justify-between text-[10px] text-text-muted z-20">
        <span>
          © {new Date().getFullYear()} Enterprise Fleet Manager. DR Module
        </span>
        <div className="flex items-center gap-4">
          <span>
            {uiStrings?.label_data_source || "Data Source"}:{" "}
            <span className="text-primary">BFF /dr/backups</span>
          </span>
          <span>
            {uiStrings?.label_system_status || "System Status"}:{" "}
            <span className="text-success font-bold">
              {infraMeta?.system_status || "--"}
            </span>
          </span>
          <span>
            {uiStrings?.label_server_time || "Server Time"}:{" "}
            <span className="font-mono text-white">
              {infraMeta?.server_time || "---"}
            </span>
          </span>
        </div>
      </footer>
    </div>
  );
};

export default BackupListView;
