import {
    ArchiveBoxIcon,
    ArrowPathIcon,
    CheckCircleIcon,
    CloudArrowDownIcon,
    CloudArrowUpIcon,
    XCircleIcon,
} from "@heroicons/react/24/outline";
import { format } from "date-fns";
import React, { useEffect, useState } from "react";
import { useUIStrings } from "../../hooks/useUIStrings";
import { api } from "../../services/api";

interface Backup {
  id: string;
  name: string;
  size_bytes: number;
  created_at: string;
  status:
    | "completed"
    | "failed"
    | "in_progress"
    | "COMPLETED"
    | "FAILED"
    | "IN_PROGRESS";
  type: "full" | "incremental";
}

interface RestoreJob {
  id: string;
  backup_id: string;
  status: "pending" | "in_progress" | "completed" | "failed";
  started_at: string;
  progress: number;
}

const DisasterRecovery: React.FC = () => {
  const { uiStrings, loading: uiLoading } = useUIStrings("dr");
  const [backups, setBackups] = useState<Backup[]>([]);
  const [restoreJobs, setRestoreJobs] = useState<RestoreJob[]>([]);
  const [loading, setLoading] = useState(true);
  const [creatingBackup, setCreatingBackup] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [sizes, setSizes] = useState<string[]>([
    "Bytes",
    "KB",
    "MB",
    "GB",
    "TB",
    "PB",
  ]); // Fallback until fetched

  useEffect(() => {
    fetch("/api/v1/ui/misc/sizes")
      .then((res) => res.json())
      .then((data) => setSizes(data))
      .catch((err) => console.error("Sizes fetch failed", err));
  }, []);

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 10000); // Poll status
    return () => clearInterval(interval);
  }, []);

  const fetchData = async () => {
    try {
      setLoading(true);
      setError(null);
      const [backupsRes, jobsRes] = await Promise.all([
        api.getBackups(),
        api.getRestoreJobs(),
      ]);
      const backupsData = (backupsRes as any)?.data ?? backupsRes;
      const jobsData = (jobsRes as any)?.data ?? jobsRes;
      const backupItems =
        backupsData?.backups ||
        backupsData?.items ||
        backupsData?.data ||
        backupsData ||
        [];
      const jobItems =
        jobsData?.jobs || jobsData?.items || jobsData?.data || jobsData || [];
      setBackups(Array.isArray(backupItems) ? backupItems : []);
      setRestoreJobs(Array.isArray(jobItems) ? jobItems : []);
    } catch (error) {
      console.error("Failed to fetch DR data", error);
      setError(
        uiStrings?.msg_load_backups_failed || "Failed to load backup jobs.",
      );
      setBackups([]);
      setRestoreJobs([]);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateBackup = async () => {
    setCreatingBackup(true);
    try {
      await api.initiateBackup("FULL");
      fetchData(); // Refresh list
    } catch (error) {
      console.error("Failed to initiate backup", error);
      setError(uiStrings?.msg_init_failed || "Backup initiation failed");
    } finally {
      setCreatingBackup(false);
    }
  };

  const handleRestore = async (backupId: string) => {
    if (
      !window.confirm(
        uiStrings?.confirm_restore ||
          "WARNING: This will overwrite current data. Are you sure?",
      )
    )
      return;
    try {
      await api.initiateRestore(backupId, "production");
      fetchData();
    } catch (error) {
      console.error("Failed to initiate restore", error);
    }
  };

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    // Use generic fallback if API takes too long
    const resolvedSize = sizes[i] || ["Bytes", "KB", "MB", "GB", "TB", "PB"][i];
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + resolvedSize;
  };

  if (uiLoading)
    return (
      <div className="p-8 text-white">
        {uiStrings?.msg_loading || "Loading..."}
      </div>
    );
  if (error) return <div className="p-8 text-red-500">{error}</div>;

  return (
    <div className="h-full flex flex-col bg-background-dark text-white space-y-6 p-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">
            {uiStrings?.title_main || "Disaster Recovery"}
          </h1>
          <p className="text-text-muted text-sm border-l-2 border-primary/30 pl-3 mt-1">
            {uiStrings?.subtitle_main ||
              "System resilience and data continuity management"}
          </p>
        </div>
        <button
          onClick={handleCreateBackup}
          disabled={creatingBackup}
          className={`flex items-center gap-2 px-4 py-2 rounded-lg font-medium transition-all shadow-lg ${
            creatingBackup
              ? "bg-primary/50 cursor-not-allowed"
              : "bg-primary hover:bg-primary/90 hover:shadow-primary/20"
          }`}
        >
          {creatingBackup ? (
            <ArrowPathIcon className="w-5 h-5 animate-spin" />
          ) : (
            <CloudArrowUpIcon className="w-5 h-5" />
          )}
          <span>
            {creatingBackup
              ? uiStrings?.btn_backing_up || "Backing Up..."
              : uiStrings?.btn_create_backup || "Create Backup"}
          </span>
        </button>
      </div>

      {/* Active Jobs Section */}
      {restoreJobs.some(
        (j) => j.status === "in_progress" || j.status === "pending",
      ) && (
        <div className="bg-surface-dark border border-amber-500/30 rounded-xl p-4 flex items-center gap-4 animate-pulse">
          <div className="p-3 rounded-full bg-amber-500/20 text-amber-500">
            <ArrowPathIcon className="w-6 h-6 animate-spin" />
          </div>
          <div className="flex-1">
            <h3 className="font-bold text-amber-500 flex items-center gap-2">
              <span className="material-symbols-outlined text-sm animate-spin">
                sync
              </span>
              {uiStrings?.title_restoration || "System Restoration in Progress"}
            </h3>
            <p className="text-sm text-text-muted">
              {uiStrings?.msg_restoration_warning ||
                "Do not close this window. System may be unstable during this process."}
            </p>
          </div>
          <div className="text-right">
            <span className="text-2xl font-mono text-white">
              {restoreJobs.find((j) => j.status === "in_progress")?.progress}%
            </span>
          </div>
        </div>
      )}

      <div className="flex-1 bg-surface-dark rounded-xl border border-border-color overflow-hidden flex flex-col">
        <div className="p-4 border-b border-border-color bg-surface-darker/50 flex justify-between items-center">
          <h3 className="font-semibold text-sm uppercase tracking-wide text-text-muted">
            {uiStrings?.title_available_backups || "Available Backups"}
          </h3>
          <span className="bg-surface-darker text-text-muted text-xs font-bold px-2 py-1 rounded border border-border-color">
            {uiStrings?.label_total_size || "Total Size"}:{" "}
            {formatBytes(backups.reduce((acc, b) => acc + b.size_bytes, 0))}
          </span>
        </div>

        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-border-color">
            <thead className="bg-surface-darker">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">
                  {uiStrings?.th_name || "Backup Name"}
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">
                  {uiStrings?.th_date || "Date"}
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">
                  {uiStrings?.th_size || "Size"}
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">
                  {uiStrings?.th_type || "Type"}
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">
                  {uiStrings?.th_status || "Status"}
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-text-muted uppercase tracking-wider">
                  {uiStrings?.th_actions || "Actions"}
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border-color">
              {loading ? (
                <tr>
                  <td
                    colSpan={6}
                    className="px-6 py-12 text-center text-text-muted"
                  >
                    {uiStrings?.msg_loading_backups || "Loading backups..."}
                  </td>
                </tr>
              ) : error ? (
                <tr>
                  <td
                    colSpan={6}
                    className="px-6 py-12 text-center text-red-400"
                  >
                    {error}
                  </td>
                </tr>
              ) : (
                backups.map((backup) => (
                  <tr
                    key={backup.id}
                    className="hover:bg-surface-darker/50 transition-colors"
                  >
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center gap-3">
                        <ArchiveBoxIcon className="w-5 h-5 text-primary" />
                        <span className="text-sm font-medium text-white">
                          {backup.name}
                        </span>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-text-muted">
                      {format(new Date(backup.created_at), "MMM d, yyyy HH:mm")}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-text-muted font-mono">
                      {formatBytes(backup.size_bytes)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className="px-2 py-0.5 rounded text-xs uppercase font-bold bg-surface-darker border border-border-color text-text-muted">
                        {backup.type}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      {backup.status === "completed" ||
                      backup.status === "COMPLETED" ? (
                        <span className="flex items-center gap-1 text-emerald-500 text-xs font-bold uppercase">
                          <CheckCircleIcon className="w-4 h-4" />{" "}
                          {uiStrings?.badge_completed || "Completed"}
                        </span>
                      ) : (
                        <span className="flex items-center gap-1 text-red-500 text-xs font-bold uppercase">
                          <XCircleIcon className="w-4 h-4" />{" "}
                          {uiStrings?.badge_failed || "Failed"}
                        </span>
                      )}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                      <button
                        onClick={() => handleRestore(backup.id)}
                        className="text-primary hover:text-primary/80 transition-colors flex items-center gap-1 justify-end ml-auto"
                        title={
                          uiStrings?.btn_restore_title ||
                          "Restore from this backup"
                        }
                      >
                        <CloudArrowDownIcon className="w-4 h-4" />{" "}
                        {uiStrings?.btn_restore || "Restore"}
                      </button>
                    </td>
                  </tr>
                ))
              )}
              {backups.length === 0 && !loading && !error && (
                <tr>
                  <td
                    colSpan={6}
                    className="px-6 py-12 text-center text-text-muted"
                  >
                    <ArchiveBoxIcon className="w-12 h-12 mx-auto mb-3 opacity-20" />
                    {uiStrings?.msg_no_backups_main ||
                      "No backups available. Create one to get started."}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};

export default DisasterRecovery;
