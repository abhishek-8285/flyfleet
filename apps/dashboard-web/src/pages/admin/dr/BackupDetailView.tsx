import React, { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { useUIStrings } from "../../../hooks/useUIStrings";
import apiClient from "../../../services/apiClient";

interface Backup {
  id: string;
  backup_type: string;
  status: string;
  start_time?: string;
  end_time?: string;
  duration_seconds?: number;
  size_bytes?: number;
  location?: string;
  retention_days?: number;
  checksum?: string;
  initiated_by?: string;
  remarks?: string;
}

const BackupDetailView: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const [backup, setBackup] = useState<Backup | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [systemMeta, setSystemMeta] = useState<any>(null);
  const { uiStrings, loading: uiLoading } = useUIStrings("dr");

  useEffect(() => {
    if (!id) return;
    const fetchBackup = async () => {
      try {
        setError(null);
        const data = (await apiClient.get(`/admin/dr/backups/${id}`)) as any;
        setBackup(data.data || data.backup || data);
      } catch (err) {
        console.error(err);
        setError(
          uiStrings?.msg_load_backup_detail_failed ||
            "Failed to load backup details.",
        );
      } finally {
        setLoading(false);
      }
    };
    const fetchMetadata = async () => {
      try {
        const data = (await apiClient.get("/ui/global_config")) as any;
        setSystemMeta(data.system_metadata);
      } catch (err) {
        console.error("Global config fetch failed", err);
      }
    };
    fetchBackup();
    fetchMetadata();
  }, [id]);

  const handleVerify = async () => {
    if (!id) return;
    try {
      await apiClient.post(`/admin/dr/backups/${id}/verify`);
      alert(uiStrings?.msg_verif_scheduled || "Verification scheduled");
    } catch (err) {
      console.error(err);
      alert(uiStrings?.msg_verif_failed || "Failed to schedule verification");
    }
  };

  if (loading || uiLoading)
    return (
      <div className="p-8 text-white">
        {uiStrings?.msg_loading || "Loading..."}
      </div>
    );
  if (error) return <div className="p-8 text-red-400">{error}</div>;
  if (!backup)
    return (
      <div className="p-8 text-white">
        {uiStrings?.msg_backup_not_found || "Backup not found"}
      </div>
    );

  const formatSize = (bytes?: number) =>
    bytes ? `${(bytes / 1024 / 1024 / 1024).toFixed(2)} GB` : "---";

  return (
    <div className="flex flex-col h-full w-full bg-background-dark text-white font-display overflow-hidden">
      <header className="h-16 flex items-center justify-between px-6 border-b border-border-color bg-surface-darker flex-shrink-0 z-20">
        <div className="flex flex-col">
          <div className="flex items-center gap-2 text-[10px] text-text-muted uppercase tracking-widest font-semibold">
            <Link
              to="/admin/dr"
              className="hover:text-primary transition-colors"
            >
              {uiStrings?.nav_backups || "Backups"}
            </Link>
            <span className="material-symbols-outlined text-[10px]">
              chevron_right
            </span>
            <span className="text-primary">{backup.id}</span>
          </div>
          <h2 className="text-white text-lg font-bold">
            {uiStrings?.title_detail_view}
          </h2>
        </div>
        <div className="flex items-center gap-3">
          <button
            onClick={handleVerify}
            className="px-4 py-2 bg-surface-dark hover:bg-border-color border border-border-color text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors"
          >
            <span className="material-symbols-outlined text-[16px]">
              verified
            </span>
            {uiStrings?.btn_verify_integrity}
          </button>
          <button className="px-4 py-2 bg-danger/10 hover:bg-danger/20 border border-danger/30 text-danger text-xs font-semibold rounded flex items-center gap-2 transition-colors">
            <span className="material-symbols-outlined text-[16px]">
              delete
            </span>
            {uiStrings?.btn_delete_backup}
          </button>
        </div>
      </header>

      <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
          <div className="bg-surface-dark rounded-xl border border-border-color p-4">
            <div className="text-text-muted text-xs font-medium uppercase mb-1">
              {uiStrings?.label_backup_type}
            </div>
            <div className="flex items-center gap-2">
              <span className="material-symbols-outlined text-primary">
                database
              </span>
              <div className="text-xl font-bold text-white">
                {backup.backup_type}
              </div>
            </div>
          </div>
          <div className="bg-surface-dark rounded-xl border border-border-color p-4">
            <div className="text-text-muted text-xs font-medium uppercase mb-1">
              {uiStrings?.label_total_size}
            </div>
            <div className="flex items-center gap-2">
              <span className="material-symbols-outlined text-indigo-400">
                hard_drive
              </span>
              <div className="text-xl font-bold text-white">
                {formatSize(backup.size_bytes)}
              </div>
            </div>
          </div>
          <div className="bg-surface-dark rounded-xl border border-border-color p-4">
            <div className="text-text-muted text-xs font-medium uppercase mb-1">
              {uiStrings?.th_status || "Status"}
            </div>
            <div className="flex items-center gap-2">
              <span
                className={`flex h-2 w-2 rounded-full ${backup.status === "VERIFIED" ? "bg-success" : "bg-warning"}`}
              ></span>
              <div
                className={`text-xl font-bold uppercase ${backup.status === "VERIFIED" ? "text-success" : "text-warning"}`}
              >
                {backup.status}
              </div>
            </div>
          </div>
          <div className="bg-surface-dark rounded-xl border border-border-color p-4">
            <div className="text-text-muted text-xs font-medium uppercase mb-1">
              {uiStrings?.label_retention_policy}
            </div>
            <div className="flex items-center gap-2">
              <span className="material-symbols-outlined text-warning">
                history
              </span>
              <div className="text-xl font-bold text-white">
                {backup.retention_days} {uiStrings?.label_days || "Days"}
              </div>
            </div>
          </div>
        </div>

        <div className="grid grid-cols-12 gap-6">
          <div className="col-span-12 lg:col-span-7">
            <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
              <div className="px-5 py-4 border-b border-border-color bg-surface-darker flex items-center justify-between">
                <h3 className="text-white text-sm font-semibold flex items-center gap-2">
                  <span className="material-symbols-outlined text-primary text-[18px]">
                    info
                  </span>
                  {uiStrings?.title_job_metadata}
                </h3>
                <span className="text-[10px] text-text-muted font-mono bg-border-color/30 px-2 py-0.5 rounded">
                  {uiStrings?.label_id_prefix || "ID:"} {backup.id}
                </span>
              </div>
              <div className="p-5">
                <div className="space-y-4">
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                    <div className="p-3 bg-surface-darker/50 rounded border border-border-color/50">
                      <div className="text-[10px] text-text-muted uppercase font-bold mb-1">
                        {uiStrings?.label_start_time}
                      </div>
                      <div className="text-sm text-white font-medium">
                        {backup.start_time
                          ? new Date(backup.start_time).toLocaleString()
                          : "-"}
                      </div>
                    </div>
                    <div className="p-3 bg-surface-darker/50 rounded border border-border-color/50">
                      <div className="text-[10px] text-text-muted uppercase font-bold mb-1">
                        {uiStrings?.label_end_time}
                      </div>
                      <div className="text-sm text-white font-medium">
                        {backup.end_time
                          ? new Date(backup.end_time).toLocaleString()
                          : "-"}
                      </div>
                    </div>
                  </div>
                  <div className="p-3 bg-surface-darker/50 rounded border border-border-color/50">
                    <div className="text-[10px] text-text-muted uppercase font-bold mb-1">
                      {uiStrings?.label_initiated_by}
                    </div>
                    <div className="flex items-center gap-2 mt-1">
                      <div className="w-6 h-6 rounded-full bg-primary/20 flex items-center justify-center">
                        <span className="material-symbols-outlined text-[14px] text-primary">
                          schedule
                        </span>
                      </div>
                      <span className="text-sm text-white">
                        {backup.initiated_by ||
                          uiStrings?.val_system_scheduler ||
                          "System Scheduler"}
                      </span>
                    </div>
                  </div>
                  <div className="p-3 bg-surface-darker/50 rounded border border-border-color/50">
                    <div className="flex justify-between items-center mb-1">
                      <div className="text-[10px] text-text-muted uppercase font-bold">
                        {uiStrings?.label_checksum}
                      </div>
                      <button className="text-primary text-[10px] hover:underline">
                        {uiStrings?.btn_copy_hash}
                      </button>
                    </div>
                    <div className="text-xs text-white font-mono break-all bg-background-dark p-2 rounded">
                      {backup.checksum || "Pending verification..."}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <footer className="bg-surface-darker border-t border-border-color px-6 py-2 flex items-center justify-between text-[10px] text-text-muted z-20">
        <span>
          © {new Date().getFullYear()} Enterprise Fleet Manager. DR Module{" "}
          {systemMeta?.version || "V--"}
        </span>
      </footer>
    </div>
  );
};

export default BackupDetailView;
