import React from "react";
import PageHeader from "./components/PageHeader";

import { useEffect, useState } from "react";
import { useUIStrings } from "../../hooks/useUIStrings";
import { monitoringService, SLAData } from "../../services/monitoringService";

const SLAReport: React.FC = () => {
  const [slaData, setSlaData] = useState<SLAData | null>(null);
  const [loading, setLoading] = useState(true);
  const [systemMeta, setSystemMeta] = useState<any>(null);
  const { uiStrings } = useUIStrings("monitoring");

  const fetchData = async () => {
    setLoading(true);
    try {
      const data = await monitoringService.getSLA();
      setSlaData(data);
    } catch (error) {
      console.error("Failed to fetch SLA data", error);
    } finally {
      setLoading(false);
      fetch("/api/v1/ui/global_config")
        .then((res) => res.json())
        .then((data) => setSystemMeta(data.system_metadata))
        .catch((err) => console.error("Failed to fetch system config", err));
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  return (
    <main className="flex-1 flex flex-col min-w-0 bg-monitor-bg-dark overflow-hidden font-public-sans text-monitor-text-muted">
      <PageHeader
        title={uiStrings?.sla_title || "SLA Compliance & Uptime Report"}
        subtitle={
          <>
            <span>{uiStrings?.metrics_subtitle || "Monitoring"}</span>
            <span className="material-symbols-outlined text-[12px]">
              chevron_right
            </span>
            <span className="text-monitor-primary">
              {uiStrings?.sla_primary || "SLA & Uptime"}
            </span>
          </>
        }
        status={uiStrings?.status_operational || "All Systems Operational"}
        actions={
          <>
            <div className="flex items-center gap-2 text-xs text-monitor-text-muted bg-monitor-surface px-3 py-1.5 rounded border border-monitor-border">
              <span className="material-symbols-outlined text-[16px]">
                calendar_month
              </span>
              <span>{uiStrings?.sla_range_90d || "Last 90 Days"}</span>
            </div>
            <button
              onClick={fetchData}
              className="px-4 py-2 bg-monitor-primary hover:bg-monitor-primary-dark text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors shadow-lg shadow-monitor-primary/20"
            >
              <span
                className={`material-symbols-outlined text-[16px] ${loading ? "animate-spin" : ""}`}
              >
                download
              </span>
              {uiStrings?.sla_btn_generate || "Generate Availability Report"}
            </button>
          </>
        }
      />
      <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
        <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-5 gap-4 mb-6">
          <div className="bg-monitor-surface rounded-xl border border-monitor-border p-4 relative overflow-hidden group">
            <div className="text-monitor-text-muted text-xs font-medium uppercase mb-1">
              {uiStrings?.sla_global_uptime || "Global Uptime (90d)"}
            </div>
            <div className="text-2xl font-bold text-monitor-success">
              {loading ? "..." : `${slaData?.global_uptime_90d}%`}
            </div>
            <div className="text-[10px] text-monitor-text-muted flex items-center gap-1 mt-1">
              {uiStrings?.sla_target || "Target:"} {systemMeta?.uptime || "--%"}
            </div>
          </div>
          <div className="bg-monitor-surface rounded-xl border border-monitor-border p-4 relative overflow-hidden group">
            <div className="text-monitor-text-muted text-xs font-medium uppercase mb-1">
              {uiStrings?.sla_mttr || "Mean Time to Resolve"}
            </div>
            <div className="text-2xl font-bold text-white">
              {loading ? "..." : `${slaData?.mttr_minutes}m`}
            </div>
            <div className="text-[10px] text-monitor-success flex items-center gap-1 mt-1">
              <span className="material-symbols-outlined text-[12px]">
                trending_down
              </span>
              {systemMeta?.mttr_diff || "--m vs last month"}
            </div>
          </div>
          <div className="bg-monitor-surface rounded-xl border border-monitor-border p-4 relative overflow-hidden group">
            <div className="text-monitor-text-muted text-xs font-medium uppercase mb-1">
              {uiStrings?.sla_error_budget || "Error Budget Remaining"}
            </div>
            <div className="text-2xl font-bold text-monitor-warning">
              {loading ? "..." : slaData?.error_budget_remaining}
            </div>
            <div className="text-[10px] text-monitor-text-muted mt-1">
              Reset in 12 days
            </div>
            <div className="w-full bg-monitor-border h-1 rounded-full mt-2 overflow-hidden">
              <div
                className="bg-monitor-warning h-full"
                style={{ width: systemMeta?.sla_compliance || "0%" }}
              ></div>
            </div>
          </div>
          <div className="bg-monitor-surface rounded-xl border border-monitor-border p-4 relative overflow-hidden group">
            <div className="text-monitor-text-muted text-xs font-medium uppercase mb-1">
              {uiStrings?.metrics_latency || "Avg Latency (Global)"}
            </div>
            <div className="text-2xl font-bold text-white">42ms</div>
            <div className="text-[10px] text-monitor-success flex items-center gap-1 mt-1">
              <span className="material-symbols-outlined text-[12px] rotate-180">
                trending_up
              </span>
              {systemMeta?.status || "--"}
            </div>
          </div>
          <div className="bg-monitor-surface rounded-xl border border-monitor-border p-4 relative overflow-hidden group">
            <div className="text-monitor-text-muted text-xs font-medium uppercase mb-1">
              {uiStrings?.health_active_incidents || "Active Incidents"}
            </div>
            <div className="text-2xl font-bold text-monitor-success">
              {loading ? "..." : slaData?.active_incidents}
            </div>
            <div className="text-[10px] text-monitor-text-muted mt-1">
              Last incident: 4d ago
            </div>
          </div>
        </div>

        <div className="grid grid-cols-12 gap-6 mb-6">
          <div className="col-span-12 lg:col-span-8 flex flex-col gap-6">
            <div className="bg-monitor-surface rounded-xl border border-monitor-border p-6 min-h-[400px]">
              <div className="flex justify-between items-center mb-6">
                <h3 className="text-white text-sm font-semibold flex items-center gap-2">
                  <span className="material-symbols-outlined text-monitor-primary text-[18px]">
                    calendar_view_week
                  </span>
                  {uiStrings?.sla_heatmap_title ||
                    "Core Services 90-Day Availability Heatmap"}
                </h3>
                <div className="flex gap-2">
                  <div className="flex items-center gap-1 text-[10px] text-monitor-text-muted">
                    <span className="w-3 h-3 rounded-sm bg-monitor-success"></span>{" "}
                    {uiStrings?.status_operational || "Operational"}
                  </div>
                  <div className="flex items-center gap-1 text-[10px] text-monitor-text-muted">
                    <span className="w-3 h-3 rounded-sm bg-monitor-warning"></span>{" "}
                    Degraded
                  </div>
                  <div className="flex items-center gap-1 text-[10px] text-monitor-text-muted">
                    <span className="w-3 h-3 rounded-sm bg-monitor-danger"></span>{" "}
                    Outage
                  </div>
                </div>
              </div>
              <div className="space-y-6">
                <div>
                  <div className="flex justify-between text-xs text-monitor-text-muted mb-2">
                    <span className="font-medium text-white w-32">
                      API Gateway
                    </span>
                    <span className="text-monitor-success font-mono">
                      {systemMeta?.success_rate || "--%"}
                    </span>
                  </div>
                  <div className="flex gap-[3px] h-6 flex-nowrap w-full">
                    {loading ? (
                      <div className="text-monitor-text-muted text-xs w-full text-center py-1">
                        Loading heatmap...
                      </div>
                    ) : (
                      slaData?.heatmap.map((status, i) => {
                        let color = "bg-monitor-success";
                        let title = "Operational";
                        if (status === 1) {
                          color = "bg-monitor-warning";
                          title = "Degraded";
                        }
                        if (status === 0) {
                          color = "bg-monitor-danger";
                          title = "Outage";
                        }
                        return (
                          <div
                            key={i}
                            className={`flex-1 ${color} rounded-sm opacity-90 hover:opacity-100 heatmap-cell h-full`}
                            title={title}
                          ></div>
                        );
                      })
                    )}
                  </div>
                </div>
              </div>
              <div className="mt-6 pt-4 border-t border-monitor-border flex justify-between items-center text-[10px] text-monitor-text-muted">
                <span>90 days ago</span>
                <span>Today</span>
              </div>
            </div>
          </div>
          <div className="col-span-12 lg:col-span-4 flex flex-col gap-6">
            <div className="bg-monitor-surface rounded-xl border border-monitor-border flex flex-col h-full">
              <div className="p-4 border-b border-monitor-border flex justify-between items-center">
                <h3 className="text-white text-sm font-semibold flex items-center gap-2">
                  <span className="material-symbols-outlined text-monitor-text-muted text-[18px]">
                    tune
                  </span>
                  {uiStrings?.sla_threshold_config ||
                    "SLA Threshold Configuration"}
                </h3>
                <button className="text-monitor-primary text-xs hover:underline">
                  {uiStrings?.sla_edit_policy || "Edit Policy"}
                </button>
              </div>
              <div className="flex-1 p-5 space-y-6">
                <div className="space-y-2">
                  <div className="flex justify-between items-center">
                    <label className="text-xs font-medium text-white">
                      Global Latency Threshold
                    </label>
                    <span className="text-[10px] bg-monitor-surface-darker px-2 py-0.5 rounded text-monitor-warning border border-monitor-warning/20">
                      Warning
                    </span>
                  </div>
                  <div className="flex items-center gap-3">
                    <input
                      className="w-full h-1 bg-monitor-border rounded-lg appearance-none cursor-pointer range-sm accent-monitor-primary"
                      type="range"
                      defaultValue="40"
                    />
                    <span className="text-xs font-mono text-white whitespace-nowrap">
                      &gt; 200ms
                    </span>
                  </div>
                  <p className="text-[10px] text-monitor-text-muted">
                    Alert triggers if p95 latency exceeds 200ms for 5m.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>
  );
};

export default SLAReport;
