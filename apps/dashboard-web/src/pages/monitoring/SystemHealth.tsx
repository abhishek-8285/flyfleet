import React, { useEffect, useState } from "react";
import { useUIStrings } from "../../hooks/useUIStrings";
import {
    monitoringService,
    SystemHealthData,
} from "../../services/monitoringService";
import PageHeader from "./components/PageHeader";

const SystemHealth: React.FC = () => {
  const [health, setHealth] = useState<SystemHealthData | null>(null);
  const [loading, setLoading] = useState(true);
  const [systemMeta, setSystemMeta] = useState<any>(null);
  const { uiStrings } = useUIStrings("monitoring");

  const fetchData = async () => {
    setLoading(true);
    try {
      const data = await monitoringService.getSystemHealth();
      setHealth(data);
    } catch (error) {
      console.error("Failed to fetch system health", error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  return (
    <main className="flex-1 flex flex-col min-w-0 bg-monitor-bg-dark overflow-hidden font-public-sans">
      <PageHeader
        title={uiStrings?.metrics_title || "System Health Dashboard"}
        status={uiStrings?.status_operational || "All Systems Operational"}
        actions={
          <>
            <button
              onClick={fetchData}
              className="px-3 py-2 bg-monitor-surface hover:bg-monitor-surface-darker border border-monitor-border text-monitor-text-muted hover:text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors"
            >
              <span
                className={`material-symbols-outlined text-[16px] ${loading ? "animate-spin" : ""}`}
              >
                refresh
              </span>
              {uiStrings?.refresh || "Refresh Data"}
            </button>
            <button className="px-4 py-2 bg-monitor-danger hover:bg-red-600 text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors shadow-lg shadow-monitor-danger/20">
              <span className="material-symbols-outlined text-[16px]">
                report_problem
              </span>
              {uiStrings?.health_incident_btn || "Incident Report"}
            </button>
          </>
        }
      />
      <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4 mb-6">
          <div className="bg-monitor-surface rounded-xl border border-monitor-border p-4 relative overflow-hidden group">
            <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
              <span className="material-symbols-outlined text-4xl text-monitor-success">
                check_circle
              </span>
            </div>
            <div className="text-monitor-text-muted text-xs font-medium uppercase mb-1">
              {uiStrings?.health_uptime_label || "Global Uptime %"}
            </div>
            <div className="text-2xl font-bold text-white">
              {loading ? "..." : `${health?.uptime_percent}%`}
            </div>
            <div className="text-[10px] text-monitor-success flex items-center gap-1 mt-1">
              <span className="material-symbols-outlined text-[12px]">
                arrow_upward
              </span>
              Target Met
            </div>
          </div>
          <div className="bg-monitor-surface rounded-xl border border-monitor-border p-4 relative overflow-hidden group">
            <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
              <span className="material-symbols-outlined text-4xl text-monitor-primary">
                speed
              </span>
            </div>
            <div className="text-monitor-text-muted text-xs font-medium uppercase mb-1">
              {uiStrings?.health_latency_label || "Avg API Latency"}
            </div>
            <div className="text-2xl font-bold text-monitor-primary">
              {loading ? "..." : `${health?.avg_latency_ms}ms`}
            </div>
            <div className="text-[10px] text-monitor-text-muted mt-1">
              Last 5 min avg
            </div>
            <div className="w-full bg-monitor-border h-1 rounded-full mt-2 overflow-hidden">
              <div
                className="bg-monitor-primary h-full"
                style={{ width: systemMeta?.cpu_usage || "0%" }}
              ></div>
            </div>
          </div>
          <div className="bg-monitor-surface rounded-xl border border-monitor-border p-4 relative overflow-hidden group">
            <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
              <span className="material-symbols-outlined text-4xl text-monitor-warning">
                bug_report
              </span>
            </div>
            <div className="text-monitor-text-muted text-xs font-medium uppercase mb-1">
              {uiStrings?.health_error_rate_label || "Active Error Rate"}
            </div>
            <div className="text-2xl font-bold text-monitor-warning">
              {loading ? "..." : `${health?.error_rate_percent}%`}
            </div>
            <div className="text-[10px] text-monitor-text-muted mt-1">
              Within tolerance
            </div>
            <div className="w-full bg-monitor-border h-1 rounded-full mt-2 overflow-hidden">
              <div
                className="bg-monitor-warning h-full"
                style={{ width: systemMeta?.memory_usage || "0%" }}
              ></div>
            </div>
          </div>
          <div className="bg-monitor-surface rounded-xl border border-monitor-border p-4 relative overflow-hidden group">
            <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
              <span className="material-symbols-outlined text-4xl text-monitor-success">
                database
              </span>
            </div>
            <div className="text-monitor-text-muted text-xs font-medium uppercase mb-1">
              {uiStrings?.health_db_health_label || "Database Health"}
            </div>
            <div className="text-2xl font-bold text-monitor-success">
              {loading ? "..." : health?.database_status}
            </div>
            <div className="text-[10px] text-monitor-text-muted mt-1">
              Replica Lag: {health?.replica_lag_ms || "--"}ms
            </div>
            <div className="w-full bg-monitor-border h-1 rounded-full mt-2 overflow-hidden">
              <div
                className="bg-monitor-success h-full"
                style={{ width: systemMeta?.disk_usage || "0%" }}
              ></div>
            </div>
          </div>
          <div className="bg-monitor-surface rounded-xl border border-monitor-border p-4 relative overflow-hidden group">
            <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
              <span className="material-symbols-outlined text-4xl text-monitor-text-muted">
                layers
              </span>
            </div>
            <div className="text-monitor-text-muted text-xs font-medium uppercase mb-1">
              {uiStrings?.health_queue_backlog_label || "Queue Backlog"}
            </div>
            <div className="text-2xl font-bold text-white">
              {loading ? "..." : health?.queue_backlog.toLocaleString()}
            </div>
            <div className="text-[10px] text-monitor-success flex items-center gap-1 mt-1">
              <span className="material-symbols-outlined text-[12px]">
                trending_down
              </span>
              {systemMeta?.uptime_diff || "--%"}
            </div>
          </div>
          <div className="bg-monitor-surface rounded-xl border border-monitor-border p-4 relative overflow-hidden group">
            <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
              <span className="material-symbols-outlined text-4xl text-orange-500">
                integration_instructions
              </span>
            </div>
            <div className="text-monitor-text-muted text-xs font-medium uppercase mb-1">
              {uiStrings?.health_integrations_label || "Active Integrations"}
            </div>
            <div className="text-2xl font-bold text-orange-500">
              {loading
                ? "..."
                : `${health?.active_integrations.active}/${health?.active_integrations.total}`}
            </div>
            <div className="text-[10px] text-monitor-text-muted mt-1">
              {loading
                ? "..."
                : `${(health?.active_integrations.total || 0) - (health?.active_integrations.active || 0)} degraded`}
            </div>
          </div>
        </div>

        <div className="grid grid-cols-12 gap-6 mb-6">
          <div className="col-span-12 lg:col-span-8 flex flex-col gap-6">
            <div className="bg-monitor-surface rounded-xl border border-monitor-border p-5 h-80 flex flex-col">
              <div className="flex justify-between items-center mb-4">
                <h3 className="text-white text-sm font-semibold flex items-center gap-2">
                  <span className="material-symbols-outlined text-monitor-primary text-[18px]">
                    show_chart
                  </span>
                  {uiStrings?.health_latency_trend ||
                    "API Latency Trend (Last 24h)"}
                </h3>
                <div className="flex items-center gap-2">
                  <span className="flex items-center gap-1 text-[10px] text-monitor-text-muted">
                    <div className="w-2 h-2 bg-monitor-primary rounded-full"></div>{" "}
                    P95
                  </span>
                  <span className="flex items-center gap-1 text-[10px] text-monitor-text-muted">
                    <div className="w-2 h-2 bg-gray-600 rounded-full"></div> P50
                  </span>
                </div>
              </div>
              <div className="flex-1 relative w-full h-full">
                <svg
                  className="w-full h-full"
                  preserveAspectRatio="none"
                  viewBox="0 0 800 200"
                >
                  <defs>
                    <linearGradient
                      id="latencyGradient"
                      x1="0"
                      x2="0"
                      y1="0"
                      y2="1"
                    >
                      <stop
                        offset="0%"
                        stopColor="#ec5b13"
                        stopOpacity="0.3"
                      ></stop>
                      <stop
                        offset="100%"
                        stopColor="#ec5b13"
                        stopOpacity="0"
                      ></stop>
                    </linearGradient>
                  </defs>
                  <line
                    stroke="#233648"
                    strokeWidth="1"
                    x1="0"
                    x2="800"
                    y1="199"
                    y2="199"
                  ></line>
                  <line
                    stroke="#233648"
                    strokeDasharray="4"
                    strokeWidth="1"
                    x1="0"
                    x2="800"
                    y1="150"
                    y2="150"
                  ></line>
                  <line
                    stroke="#233648"
                    strokeDasharray="4"
                    strokeWidth="1"
                    x1="0"
                    x2="800"
                    y1="100"
                    y2="100"
                  ></line>
                  <line
                    stroke="#233648"
                    strokeDasharray="4"
                    strokeWidth="1"
                    x1="0"
                    x2="800"
                    y1="50"
                    y2="50"
                  ></line>
                  <path
                    d="M0,120 C100,130 200,90 300,100 S500,40 600,60 S700,80 800,50 V200 H0 Z"
                    fill="url(#latencyGradient)"
                  ></path>
                  <path
                    d="M0,120 C100,130 200,90 300,100 S500,40 600,60 S700,80 800,50"
                    fill="none"
                    stroke="#ec5b13"
                    strokeWidth="2"
                  ></path>
                  <path
                    d="M0,160 C100,165 200,150 300,155 S500,130 600,140 S700,145 800,130"
                    fill="none"
                    stroke="#4b5563"
                    strokeDasharray="4,2"
                    strokeWidth="1.5"
                  ></path>
                  <circle
                    cx="600"
                    cy="60"
                    fill="#fff"
                    r="4"
                    stroke="#ec5b13"
                    strokeWidth="2"
                  ></circle>
                </svg>
                <div className="absolute bottom-2 left-0 right-0 flex justify-between text-[10px] text-monitor-text-muted px-2">
                  <span>00:00</span>
                  <span>06:00</span>
                  <span>12:00</span>
                  <span>18:00</span>
                  <span>23:59</span>
                </div>
              </div>
            </div>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div className="bg-monitor-surface rounded-xl border border-monitor-border p-4 flex flex-col items-center justify-center gap-2 group hover:border-monitor-success/50 transition-colors">
                <div className="w-10 h-10 rounded-full bg-monitor-success/10 flex items-center justify-center text-monitor-success">
                  <span className="material-symbols-outlined">cloud_queue</span>
                </div>
                <div className="text-center">
                  <div className="text-white text-sm font-semibold">
                    Web BFF
                  </div>
                  <div className="text-[10px] text-monitor-success font-medium">
                    {systemMeta?.uptime || "--%"} Uptime
                  </div>
                </div>
              </div>
              <div className="bg-monitor-surface rounded-xl border border-monitor-border p-4 flex flex-col items-center justify-center gap-2 group hover:border-monitor-success/50 transition-colors">
                <div className="w-10 h-10 rounded-full bg-monitor-success/10 flex items-center justify-center text-monitor-success">
                  <span className="material-symbols-outlined">lock</span>
                </div>
                <div className="text-center">
                  <div className="text-white text-sm font-semibold">
                    Auth Service
                  </div>
                  <div className="text-[10px] text-monitor-success font-medium">
                    {systemMeta?.uptime || "--%"} Uptime
                  </div>
                </div>
              </div>
              <div className="bg-monitor-surface rounded-xl border border-monitor-border p-4 flex flex-col items-center justify-center gap-2 group hover:border-monitor-warning/50 transition-colors">
                <div className="w-10 h-10 rounded-full bg-monitor-warning/10 flex items-center justify-center text-monitor-warning">
                  <span className="material-symbols-outlined">payments</span>
                </div>
                <div className="text-center">
                  <div className="text-white text-sm font-semibold">
                    Billing API
                  </div>
                  <div className="text-[10px] text-monitor-warning font-medium">
                    {systemMeta?.uptime || "--%"} Uptime
                  </div>
                </div>
              </div>
              <div className="bg-monitor-surface rounded-xl border border-monitor-border p-4 flex flex-col items-center justify-center gap-2 group hover:border-monitor-success/50 transition-colors">
                <div className="w-10 h-10 rounded-full bg-monitor-success/10 flex items-center justify-center text-monitor-success">
                  <span className="material-symbols-outlined">
                    local_shipping
                  </span>
                </div>
                <div className="text-center">
                  <div className="text-white text-sm font-semibold">
                    Fleet Engine
                  </div>
                  <div className="text-[10px] text-monitor-success font-medium">
                    {systemMeta?.uptime || "--%"} Uptime
                  </div>
                </div>
              </div>
            </div>
          </div>
          <div className="col-span-12 lg:col-span-4 flex flex-col gap-6">
            <div className="bg-monitor-surface rounded-xl border border-monitor-border flex flex-col h-[280px]">
              <div className="p-4 border-b border-monitor-border flex justify-between items-center">
                <h3 className="text-white text-sm font-semibold">
                  {uiStrings?.health_error_sources || "Top Error Sources"}
                </h3>
                <button className="text-monitor-primary text-xs hover:underline">
                  {uiStrings?.logs_title || "View Logs"}
                </button>
              </div>
              <div className="flex-1 overflow-y-auto custom-scrollbar p-2">
                <div className="space-y-1">
                  {loading ? (
                    <div className="text-monitor-text-muted text-xs p-2">
                      Loading top errors...
                    </div>
                  ) : (
                    health?.top_errors.map((error, idx) => (
                      <div
                        key={idx}
                        className="flex items-center justify-between p-2 hover:bg-monitor-surface-darker rounded transition-colors group"
                      >
                        <div className="flex items-center gap-3">
                          <div className="w-8 h-8 rounded bg-monitor-surface-darker flex items-center justify-center text-xs font-mono text-monitor-text-muted border border-monitor-border">
                            {error.code}
                          </div>
                          <div className="flex flex-col">
                            <span className="text-sm text-white font-medium">
                              {error.service}
                            </span>
                            <span className="text-[10px] text-monitor-text-muted">
                              {error.type}
                            </span>
                          </div>
                        </div>
                        <span
                          className={`text-xs font-bold ${error.code >= 500 ? "text-monitor-danger bg-monitor-danger/10" : "text-monitor-warning bg-monitor-warning/10"} px-2 py-0.5 rounded`}
                        >
                          {error.count} err
                        </span>
                      </div>
                    ))
                  )}
                </div>
              </div>
            </div>
            <div className="bg-monitor-surface rounded-xl border border-monitor-border flex flex-col flex-1 min-h-[220px]">
              <div className="p-4 border-b border-monitor-border">
                <h3 className="text-white text-sm font-semibold">
                  {uiStrings?.health_queue_perf || "Queue Performance"}
                </h3>
              </div>
              <div className="p-4 flex flex-col justify-center h-full gap-4">
                <div className="space-y-1">
                  <div className="flex justify-between text-xs">
                    <span className="text-white">Email Notifications</span>
                    <span className="text-monitor-text-muted">850/1000</span>
                  </div>
                  <div className="w-full bg-monitor-border h-2 rounded-full overflow-hidden">
                    <div
                      className="bg-monitor-primary h-full"
                      style={{ width: systemMeta?.email_usage || "0%" }}
                    ></div>
                  </div>
                </div>
                <div className="space-y-1">
                  <div className="flex justify-between text-xs">
                    <span className="text-white">Data Ingestion</span>
                    <div className="flex items-center gap-1">
                      <span className="material-symbols-outlined text-[10px] text-monitor-warning">
                        warning
                      </span>
                      <span className="text-monitor-warning">920/1000</span>
                    </div>
                  </div>
                  <div className="w-full bg-monitor-border h-2 rounded-full overflow-hidden">
                    <div
                      className="bg-monitor-warning h-full"
                      style={{
                        width: systemMeta?.data_ingestion_usage || "0%",
                      }}
                    ></div>
                  </div>
                </div>
                <div className="space-y-1">
                  <div className="flex justify-between text-xs">
                    <span className="text-white">Audit Logs</span>
                    <span className="text-monitor-text-muted">240/5000</span>
                  </div>
                  <div className="w-full bg-monitor-border h-2 rounded-full overflow-hidden">
                    <div
                      className="bg-monitor-success h-full"
                      style={{ width: systemMeta?.audit_log_usage || "0%" }}
                    ></div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div className="bg-monitor-surface rounded-xl border border-monitor-border flex flex-col overflow-hidden">
          <div className="p-4 border-b border-monitor-border flex justify-between items-center bg-monitor-surface-darker">
            <h3 className="text-white font-semibold text-sm flex items-center gap-2">
              <span className="material-symbols-outlined text-monitor-danger text-[18px]">
                notifications_active
              </span>
              {uiStrings?.health_alerts_title || "System Alerts (Real-time)"}
            </h3>
            <div className="flex items-center gap-2">
              <input
                className="bg-monitor-surface border border-monitor-border text-xs text-white rounded px-3 py-1.5 w-64 focus:ring-1 focus:ring-monitor-primary focus:border-monitor-primary outline-none"
                placeholder="Filter alerts..."
                type="text"
              />
              <button className="p-1.5 hover:bg-monitor-border rounded text-monitor-text-muted hover:text-white">
                <span className="material-symbols-outlined text-[18px]">
                  filter_list
                </span>
              </button>
              <button className="p-1.5 hover:bg-monitor-border rounded text-monitor-text-muted hover:text-white">
                <span className="material-symbols-outlined text-[18px]">
                  settings
                </span>
              </button>
            </div>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse">
              <thead>
                <tr className="border-b border-monitor-border text-xs text-monitor-text-muted uppercase tracking-wider bg-monitor-surface-darker/50">
                  <th className="px-4 py-3 font-semibold">Timestamp</th>
                  <th className="px-4 py-3 font-semibold">Service</th>
                  <th className="px-4 py-3 font-semibold">Alert Message</th>
                  <th className="px-4 py-3 font-semibold">Type</th>
                  <th className="px-4 py-3 font-semibold">Instance ID</th>
                  <th className="px-4 py-3 font-semibold text-right">
                    Severity
                  </th>
                </tr>
              </thead>
              <tbody className="text-sm">
                {loading ? (
                  <tr>
                    <td
                      colSpan={6}
                      className="px-4 py-8 text-center text-monitor-text-muted"
                    >
                      Loading alerts...
                    </td>
                  </tr>
                ) : (health?.alerts || []).length > 0 ? (
                  health?.alerts?.map((alert, idx) => (
                    <tr
                      key={idx}
                      className="border-b border-monitor-border hover:bg-monitor-surface-darker/50 transition-colors"
                    >
                      <td className="px-4 py-3 text-monitor-text-muted text-xs whitespace-nowrap">
                        {new Date(alert.timestamp).toLocaleString([], {
                          month: "short",
                          day: "numeric",
                          hour: "2-digit",
                          minute: "2-digit",
                        })}
                      </td>
                      <td className="px-4 py-3 font-medium text-white">
                        {alert.service}
                      </td>
                      <td className="px-4 py-3 text-white">{alert.message}</td>
                      <td className="px-4 py-3 text-monitor-text-muted">
                        {alert.type}
                      </td>
                      <td className="px-4 py-3 text-monitor-text-muted font-mono text-xs">
                        {alert.instance_id}
                      </td>
                      <td className="px-4 py-3 text-right">
                        <span
                          className={`inline-flex items-center gap-1 px-2 py-0.5 rounded text-[10px] font-bold ${alert.severity === "CRITICAL" ? "bg-monitor-danger/10 text-monitor-danger border-monitor-danger/20" : "bg-monitor-warning/10 text-monitor-warning border-monitor-warning/20"} border`}
                        >
                          {alert.severity}
                        </span>
                      </td>
                    </tr>
                  ))
                ) : (
                  <tr>
                    <td
                      colSpan={6}
                      className="px-4 py-8 text-center text-monitor-text-muted"
                    >
                      No active alerts
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      </div>
      <footer className="bg-monitor-surface-darker border-t border-monitor-border px-6 py-2 flex items-center justify-between text-[10px] text-monitor-text-muted z-20">
        <span>
          © {new Date().getFullYear()} Enterprise Fleet Manager. Monitoring
          Module {systemMeta?.version || "V--"}
        </span>
        <div className="flex items-center gap-4">
          <span>
            Region: <span className="text-white">us-east-1</span>
          </span>
          <span>
            Environment:{" "}
            <span className="text-monitor-warning">Production</span>
          </span>
          <span>
            Last Sync:{" "}
            <span className="font-mono text-white">
              {systemMeta?.server_time || "---"}
            </span>
          </span>
        </div>
      </footer>
    </main>
  );
};

export default SystemHealth;
