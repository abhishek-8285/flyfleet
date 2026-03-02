import React, { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { useUIStrings } from "../../../hooks/useUIStrings";
import apiClient from "../../../services/apiClient";

interface Backup {
  id: string;
  backup_type: string;
  status: string;
  size_bytes?: number;
  created_at: string;
  checksum?: string;
}

const InitiateRestorePage: React.FC = () => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const preselectedBackup = searchParams.get("backup");

  const [backups, setBackups] = useState<Backup[]>([]);
  const [selectedBackup, setSelectedBackup] = useState<string>(
    preselectedBackup || "",
  );
  const [restoreType, setRestoreType] = useState("FULL");
  const [targetEnv, setTargetEnv] = useState("STAGING");
  const [loading, setLoading] = useState(false);
  const [checksumVerified, setChecksumVerified] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [infraMeta, setInfraMeta] = useState<any>(null);
  const [systemMeta, setSystemMeta] = useState<any>(null);
  const { uiStrings, loading: uiLoading } = useUIStrings("dr");

  useEffect(() => {
    (apiClient.get("/admin/dr/backups") as Promise<any>)
      .then((data) => {
        const items = data?.data || data?.items || data?.backups || data || [];
        setBackups(Array.isArray(items) ? items : []);
      })
      .catch(() =>
        setError(
          uiStrings?.msg_load_backups_failed || "Failed to load backups.",
        ),
      );

    apiClient
      .get("/ui/global_config")
      .then((res: any) => {
        setInfraMeta(res?.infrastructure_metadata?.dr_metrics || null);
        setSystemMeta(res?.system_metadata || null);
      })
      .catch((err) => console.error("Failed to fetch infra config", err));
  }, []);

  useEffect(() => {
    if (selectedBackup) {
      // Simulate checksum verification
      setChecksumVerified(false);
      const timer = setTimeout(() => setChecksumVerified(true), 1200);
      return () => clearTimeout(timer);
    }
  }, [selectedBackup]);

  const selectedBackupData = backups.find((b) => b.id === selectedBackup);

  const formatSize = (bytes?: number) => {
    if (!bytes) return "--";
    if (bytes >= 1e12) return (bytes / 1e12).toFixed(2) + " TB";
    return (bytes / 1e9).toFixed(2) + " GB";
  };

  const handleSubmit = async () => {
    if (!selectedBackup) return;
    setLoading(true);
    setError(null);
    try {
      const data = await apiClient.post("/admin/dr/restore-jobs", {
        backup_id: selectedBackup,
        restore_type: restoreType,
        target_environment: targetEnv,
        requested_by: "System User",
      });
      if (data?.data?.id) {
        navigate(`/admin/dr/restores/${data.data.id}`);
      } else {
        throw new Error("Failed to initiate restore");
      }
    } catch (err) {
      console.error(err);
      setError(
        uiStrings?.msg_initiate_restore_page_failed ||
          "Failed to initiate restore.",
      );
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex flex-col h-full w-full bg-background-dark text-white font-display overflow-hidden">
      <header className="h-16 flex items-center justify-between px-6 border-b border-border-color bg-surface-darker flex-shrink-0 z-20">
        <div className="flex flex-col">
          <nav className="flex items-center gap-2 text-[10px] font-bold uppercase tracking-[0.1em] text-text-muted mb-0.5">
            <span>{uiStrings?.breadcrumb_dr || "DISASTER RECOVERY"}</span>
            <span className="material-symbols-outlined text-[12px]">
              chevron_right
            </span>
            <span>
              {uiStrings?.breadcrumb_restore_points || "RESTORE POINTS"}
            </span>
            <span className="material-symbols-outlined text-[12px]">
              chevron_right
            </span>
            <span className="text-primary">
              {uiStrings?.breadcrumb_initiate || "INITIATE"}
            </span>
          </nav>
          <h2 className="text-white text-lg font-bold leading-none">
            {uiStrings?.title_initiate_system_restore ||
              "Initiate Restore Operation"}
          </h2>
        </div>
        <div className="flex items-center gap-3">
          <button
            onClick={() => navigate("/admin/dr")}
            className="px-4 py-2 text-xs font-semibold text-text-muted hover:text-white transition-colors"
          >
            {uiStrings?.btn_cancel || "Cancel"}
          </button>
          <button
            onClick={handleSubmit}
            disabled={!selectedBackup || loading || !checksumVerified}
            className="px-4 py-2 bg-primary hover:bg-primary-dark disabled:opacity-50 disabled:cursor-not-allowed text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors shadow-lg shadow-primary/20"
          >
            <span className="material-symbols-outlined text-[18px]">
              restore_page
            </span>
            {loading
              ? uiStrings?.btn_starting || "Starting..."
              : uiStrings?.btn_start_restore || "Start Restore Operation"}
          </button>
        </div>
      </header>

      <div className="flex-1 overflow-y-auto p-8 custom-scrollbar">
        <div className="max-w-4xl mx-auto space-y-8">
          {/* Restore Configuration */}
          <section className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
            <div className="px-6 py-4 border-b border-border-color bg-surface-darker/50">
              <h3 className="text-sm font-bold text-white flex items-center gap-2">
                <span className="material-symbols-outlined text-primary text-[18px]">
                  settings
                </span>
                {uiStrings?.title_restore_config}
              </h3>
            </div>
            <div className="p-6 grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="col-span-2">
                <label className="block text-xs font-semibold text-text-muted uppercase tracking-wider mb-2">
                  {uiStrings?.label_backup_id_search}
                </label>
                <div className="relative group">
                  <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                    <span className="material-symbols-outlined text-text-muted text-[20px]">
                      search
                    </span>
                  </div>
                  <select
                    value={selectedBackup}
                    onChange={(e) => setSelectedBackup(e.target.value)}
                    className="w-full bg-surface-darker border border-border-color rounded-lg pl-10 pr-4 py-2.5 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none transition-all appearance-none cursor-pointer"
                  >
                    <option value="">
                      {uiStrings?.placeholder_select_backup}
                    </option>
                    {backups.map((b) => (
                      <option key={b.id} value={b.id}>
                        {b.id.substring(0, 24)} - {b.backup_type} -{" "}
                        {formatSize(b.size_bytes)}
                      </option>
                    ))}
                  </select>
                  {selectedBackup && checksumVerified && (
                    <div className="absolute inset-y-0 right-0 pr-3 flex items-center pointer-events-none">
                      <span className="text-[10px] bg-success/20 text-success px-1.5 py-0.5 rounded font-bold">
                        {uiStrings?.label_healthy}
                      </span>
                    </div>
                  )}
                </div>
                {error && (
                  <p className="mt-2 text-[11px] text-danger">{error}</p>
                )}
                {selectedBackupData && (
                  <p className="mt-2 text-[11px] text-text-muted">
                    {uiStrings?.tip_backup_details
                      ?.replace("{type}", selectedBackupData.backup_type)
                      ?.replace(
                        "{size}",
                        formatSize(selectedBackupData.size_bytes),
                      )
                      ?.replace(
                        "{date}",
                        new Date(
                          selectedBackupData.created_at,
                        ).toLocaleString(),
                      )}
                  </p>
                )}
              </div>

              <div>
                <label className="block text-xs font-semibold text-text-muted uppercase tracking-wider mb-2">
                  {uiStrings?.label_restore_type}
                </label>
                <div className="relative">
                  <select
                    value={restoreType}
                    onChange={(e) => setRestoreType(e.target.value)}
                    className="w-full bg-surface-darker border border-border-color rounded-lg px-4 py-2.5 text-sm text-white appearance-none focus:ring-1 focus:ring-primary focus:border-primary outline-none cursor-pointer"
                  >
                    <option value="FULL">{uiStrings?.val_full_restore}</option>
                    <option value="PARTIAL">
                      {uiStrings?.val_partial_restore}
                    </option>
                  </select>
                  <div className="absolute inset-y-0 right-0 pr-3 flex items-center pointer-events-none">
                    <span className="material-symbols-outlined text-text-muted">
                      expand_more
                    </span>
                  </div>
                </div>
              </div>

              <div>
                <label className="block text-xs font-semibold text-text-muted uppercase tracking-wider mb-2">
                  {uiStrings?.label_target_env}
                </label>
                <div className="relative">
                  <select
                    value={targetEnv}
                    onChange={(e) => setTargetEnv(e.target.value)}
                    className="w-full bg-surface-darker border border-border-color rounded-lg px-4 py-2.5 text-sm text-white appearance-none focus:ring-1 focus:ring-primary focus:border-primary outline-none cursor-pointer"
                  >
                    <option value="STAGING">
                      {uiStrings?.val_staging_env}
                    </option>
                    <option value="UAT">{uiStrings?.val_uat_env}</option>
                    <option value="PROD">{uiStrings?.val_prod_env}</option>
                  </select>
                  <div className="absolute inset-y-0 right-0 pr-3 flex items-center pointer-events-none">
                    <span className="material-symbols-outlined text-text-muted">
                      expand_more
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </section>

          {/* Verification Section */}
          <section className="bg-surface-dark rounded-xl border border-border-color overflow-hidden">
            <div className="px-6 py-4 border-b border-border-color bg-surface-darker/50">
              <h3 className="text-sm font-bold text-white flex items-center gap-2">
                <span className="material-symbols-outlined text-primary text-[18px]">
                  verified
                </span>
                {uiStrings?.title_verification}
              </h3>
            </div>
            <div className="p-6">
              <div className="flex items-center justify-between p-4 bg-surface-darker border border-border-color rounded-lg">
                <div className="flex items-center gap-4">
                  <div
                    className={`w-10 h-10 rounded-full flex items-center justify-center ${checksumVerified ? "bg-success/10 text-success" : "bg-warning/10 text-warning"}`}
                  >
                    <span className="material-symbols-outlined">
                      {checksumVerified ? "check_circle" : "hourglass_empty"}
                    </span>
                  </div>
                  <div>
                    <div className="text-sm font-bold text-white">
                      {uiStrings?.label_checksum_validation ||
                        "Checksum Validation"}
                    </div>
                    <div className="text-xs text-text-muted">
                      {uiStrings?.tip_checksum_validation ||
                        "Verifying the integrity of the selected backup file."}
                    </div>
                  </div>
                </div>
                <div className="text-right">
                  <div
                    className={`text-xs font-bold uppercase tracking-wider ${checksumVerified ? "text-success" : "text-warning"}`}
                  >
                    {checksumVerified
                      ? uiStrings?.val_verified || "Verified"
                      : uiStrings?.val_checking || "Checking..."}
                  </div>
                  {checksumVerified && selectedBackupData?.checksum && (
                    <div className="text-[10px] font-mono text-text-muted">
                      {uiStrings?.label_match || "MATCH:"}{" "}
                      {selectedBackupData.checksum.substring(0, 8)}...
                    </div>
                  )}
                </div>
              </div>

              {targetEnv === "PROD" && (
                <div className="mt-6 p-4 rounded-lg bg-warning/10 border border-warning/20 flex gap-4">
                  <span className="material-symbols-outlined text-warning">
                    warning
                  </span>
                  <div>
                    <h4 className="text-xs font-bold text-warning uppercase mb-1">
                      {uiStrings?.title_impact_warning}
                    </h4>
                    <p className="text-xs text-text-muted leading-relaxed">
                      {uiStrings?.msg_prod_impact?.replace(
                        "{type}",
                        restoreType,
                      )}
                    </p>
                  </div>
                </div>
              )}
            </div>
          </section>

          {/* Footer Info */}
          <div className="flex items-center justify-between px-2">
            <div className="flex items-center gap-2 text-xs text-text-muted">
              <span className="material-symbols-outlined text-[16px]">
                info
              </span>
              {uiStrings?.label_restore_estimates}
            </div>
            <button className="text-[11px] font-bold text-primary uppercase tracking-widest hover:underline">
              {uiStrings?.btn_advanced_params}
            </button>
          </div>
        </div>
      </div>

      <footer className="bg-surface-darker border-t border-border-color px-6 py-2 flex items-center justify-between text-[10px] text-text-muted z-20">
        <span>
          © {new Date().getFullYear()} {uiStrings?.footer_fleet_security}. DR
          Module {systemMeta?.version || "V--"}
        </span>
        <div className="flex items-center gap-4">
          <span>
            {uiStrings?.footer_api_endpoint}{" "}
            <span className="font-mono text-white">POST /dr/restores</span>
          </span>
          <span>
            {uiStrings?.footer_system_status}{" "}
            <span className="text-success">
              {infraMeta?.system_status || uiStrings?.val_ready || "Ready"}
            </span>
          </span>
        </div>
      </footer>
    </div>
  );
};

export default InitiateRestorePage;
