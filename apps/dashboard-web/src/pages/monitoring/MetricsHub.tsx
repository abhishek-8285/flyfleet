import React, { useCallback, useEffect, useState } from "react";
import { useUIStrings } from "../../hooks/useUIStrings";
import apiClient from "../../services/apiClient";
import PageHeader from "./components/PageHeader";

import { MetricsData } from "../../services/monitoringService";

const MetricsHub: React.FC = () => {
  const [metrics, setMetrics] = useState<MetricsData | null>(null);
  const [loading, setLoading] = useState(true);
  const [systemMeta, setSystemMeta] = useState<any>(null);
  const { uiStrings } = useUIStrings("monitoring");

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      const res = await apiClient.get("/admin/metrics/performance");
      setMetrics(res.data);

      const uiRes = await apiClient.get("/ui/global_config");
      setSystemMeta(uiRes.data?.system_metadata || null);
    } catch (error) {
      console.error("Failed to fetch metrics:", error);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);
  return (
    <main className="flex-1 flex flex-col min-w-0 bg-monitor-bg-dark overflow-hidden font-public-sans text-monitor-text-muted">
      <PageHeader
        title={uiStrings?.metrics_title || "System Metrics Hub"}
        subtitle={
          <>
            <span>{uiStrings?.metrics_subtitle || "Monitoring"}</span>
            <span className="material-symbols-outlined text-[10px]">
              chevron_right
            </span>
            <span className="text-monitor-primary">
              {uiStrings?.metrics_primary || "Performance"}
            </span>
          </>
        }
        actions={
          <>
            <div className="bg-monitor-surface p-1 rounded-lg border border-monitor-border flex items-center">
              <button className="px-3 py-1 text-xs font-medium rounded text-white bg-monitor-primary shadow-sm">
                {uiStrings?.status_live || "Live"}
              </button>
              <button className="px-3 py-1 text-xs font-medium rounded text-monitor-text-muted hover:text-white transition-colors">
                1h
              </button>
            </div>
            <button
              onClick={fetchData}
              className="px-4 py-2 bg-monitor-surface hover:bg-monitor-surface-light hover:text-background-dark text-white text-xs font-semibold rounded border border-monitor-border flex items-center gap-2 transition-colors"
            >
              <span
                className={`material-symbols-outlined text-[16px] ${loading ? "animate-spin" : ""}`}
              >
                refresh
              </span>
              {uiStrings?.refresh || "Refresh"}
            </button>
          </>
        }
      />
      <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-6">
          <div className="bg-monitor-surface rounded-xl border border-monitor-border p-5 h-80 flex flex-col">
            <div className="flex justify-between items-center mb-4">
              <div>
                <h3 className="text-white text-sm font-semibold flex items-center gap-2">
                  {uiStrings?.metrics_throughput || "Throughput"}
                </h3>
                <p className="text-[10px] text-monitor-text-muted">
                  {uiStrings?.metrics_requests_sec || "Requests per Second"}
                </p>
              </div>
              <div className="text-right">
                <span className="text-xl font-bold text-white">
                  {loading ? "..." : metrics?.throughput.toLocaleString()}
                </span>
                <span className="text-xs text-monitor-success block">
                  {systemMeta?.uptime_diff || "▲ --%"}
                </span>
              </div>
            </div>
            <div className="flex-1 relative w-full h-full">
              <svg
                className="w-full h-full"
                preserveAspectRatio="none"
                viewBox="0 0 300 150"
              >
                <defs>
                  <linearGradient
                    id="chartGradient1"
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
                <g className="chart-grid">
                  <line
                    x1="0"
                    x2="300"
                    y1="37.5"
                    y2="37.5"
                    stroke="#233648"
                    strokeDasharray="4"
                  ></line>
                  <line
                    x1="0"
                    x2="300"
                    y1="75"
                    y2="75"
                    stroke="#233648"
                    strokeDasharray="4"
                  ></line>
                  <line
                    x1="0"
                    x2="300"
                    y1="112.5"
                    y2="112.5"
                    stroke="#233648"
                    strokeDasharray="4"
                  ></line>
                </g>
                <path
                  d="M0,120 L20,110 L40,115 L60,90 L80,95 L100,60 L120,70 L140,50 L160,55 L180,40 L200,45 L220,30 L240,40 L260,35 L280,20 L300,25 V150 H0 Z"
                  fill="url(#chartGradient1)"
                ></path>
                <path
                  d="M0,120 L20,110 L40,115 L60,90 L80,95 L100,60 L120,70 L140,50 L160,55 L180,40 L200,45 L220,30 L240,40 L260,35 L280,20 L300,25"
                  fill="none"
                  stroke="#ec5b13"
                  strokeWidth="2"
                  strokeLinejoin="round"
                  strokeLinecap="round"
                ></path>
              </svg>
            </div>
          </div>
          <div className="bg-monitor-surface rounded-xl border border-monitor-border p-5 h-80 flex flex-col">
            <div className="flex justify-between items-center mb-4">
              <div>
                <h3 className="text-white text-sm font-semibold flex items-center gap-2">
                  {uiStrings?.metrics_db_latency || "DB Latency"}
                </h3>
                <p className="text-[10px] text-monitor-text-muted">
                  {uiStrings?.metrics_ms || "Query Execution Time (ms)"}
                </p>
              </div>
              <div className="text-right">
                <span className="text-xl font-bold text-monitor-warning">
                  {loading ? "..." : `${metrics?.db_latency_ms}ms`}
                </span>
                <span className="text-xs text-monitor-warning block">
                  ▲ 5ms
                </span>
              </div>
            </div>
            <div className="flex-1 relative w-full h-full">
              <svg
                className="w-full h-full"
                preserveAspectRatio="none"
                viewBox="0 0 300 150"
              >
                <defs>
                  <linearGradient
                    id="chartGradient2"
                    x1="0"
                    x2="0"
                    y1="0"
                    y2="1"
                  >
                    <stop
                      offset="0%"
                      stopColor="#f59e0b"
                      stopOpacity="0.3"
                    ></stop>
                    <stop
                      offset="100%"
                      stopColor="#f59e0b"
                      stopOpacity="0"
                    ></stop>
                  </linearGradient>
                </defs>
                <g className="chart-grid">
                  <line
                    x1="0"
                    x2="300"
                    y1="37.5"
                    y2="37.5"
                    stroke="#233648"
                    strokeDasharray="4"
                  ></line>
                  <line
                    x1="0"
                    x2="300"
                    y1="75"
                    y2="75"
                    stroke="#233648"
                    strokeDasharray="4"
                  ></line>
                  <line
                    x1="0"
                    x2="300"
                    y1="112.5"
                    y2="112.5"
                    stroke="#233648"
                    strokeDasharray="4"
                  ></line>
                </g>
                <path
                  d="M0,100 L30,95 L60,105 L90,80 L120,85 L150,110 L180,90 L210,50 L240,60 L270,40 L300,55 V150 H0 Z"
                  fill="url(#chartGradient2)"
                ></path>
                <path
                  d="M0,100 L30,95 L60,105 L90,80 L120,85 L150,110 L180,90 L210,50 L240,60 L270,40 L300,55"
                  fill="none"
                  stroke="#f59e0b"
                  strokeWidth="2"
                  strokeLinejoin="round"
                  strokeLinecap="round"
                ></path>
              </svg>
            </div>
          </div>
          <div className="bg-monitor-surface rounded-xl border border-monitor-border p-5 h-80 flex flex-col">
            <div className="flex justify-between items-center mb-4">
              <div>
                <h3 className="text-white text-sm font-semibold flex items-center gap-2">
                  {uiStrings?.metrics_cache_hit || "Cache Performance"}
                </h3>
                <p className="text-[10px] text-monitor-text-muted">
                  {uiStrings?.metrics_requests_sec || "Hit vs Miss Rate"}
                </p>
              </div>
              <div className="text-right">
                <span className="text-xl font-bold text-monitor-success">
                  {loading ? "..." : `${metrics?.cache_hit_ratio}%`}
                </span>
                <span className="text-xs text-monitor-text-muted block">
                  {uiStrings?.metrics_cache_hit || "Hit Ratio"}
                </span>
              </div>
            </div>
            <div className="flex-1 relative w-full h-full">
              <svg
                className="w-full h-full"
                preserveAspectRatio="none"
                viewBox="0 0 300 150"
              >
                <defs>
                  <linearGradient
                    id="chartGradient3"
                    x1="0"
                    x2="0"
                    y1="0"
                    y2="1"
                  >
                    <stop
                      offset="0%"
                      stopColor="#10b981"
                      stopOpacity="0.3"
                    ></stop>
                    <stop
                      offset="100%"
                      stopColor="#10b981"
                      stopOpacity="0"
                    ></stop>
                  </linearGradient>
                </defs>
                <g className="chart-grid">
                  <line
                    x1="0"
                    x2="300"
                    y1="37.5"
                    y2="37.5"
                    stroke="#233648"
                    strokeDasharray="4"
                  ></line>
                  <line
                    x1="0"
                    x2="300"
                    y1="75"
                    y2="75"
                    stroke="#233648"
                    strokeDasharray="4"
                  ></line>
                  <line
                    x1="0"
                    x2="300"
                    y1="112.5"
                    y2="112.5"
                    stroke="#233648"
                    strokeDasharray="4"
                  ></line>
                </g>
                <path
                  d="M0,80 L30,75 L60,78 L90,70 L120,72 L150,65 L180,68 L210,60 L240,62 L270,55 L300,58 V150 H0 Z"
                  fill="url(#chartGradient3)"
                ></path>
                <path
                  d="M0,80 L30,75 L60,78 L90,70 L120,72 L150,65 L180,68 L210,60 L240,62 L270,55 L300,58"
                  fill="none"
                  stroke="#10b981"
                  strokeWidth="2"
                  strokeLinejoin="round"
                  strokeLinecap="round"
                ></path>
              </svg>
            </div>
          </div>
        </div>

        <div className="bg-monitor-surface rounded-xl border border-monitor-border flex flex-col overflow-hidden">
          <div className="p-4 border-b border-monitor-border flex justify-between items-center bg-monitor-surface-darker">
            <h3 className="text-white font-semibold text-sm flex items-center gap-2">
              <span className="material-symbols-outlined text-monitor-text-muted text-[18px]">
                memory
              </span>
              {uiStrings?.metrics_cluster_usage || "Cluster Resource Usage"}
            </h3>
            <div className="flex items-center gap-2">
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
                  <th className="px-6 py-3 font-semibold w-1/4">
                    {uiStrings?.th_service || "Service Cluster"}
                  </th>
                  <th className="px-6 py-3 font-semibold">
                    {uiStrings?.metrics_cpu || "CPU Usage"}
                  </th>
                  <th className="px-6 py-3 font-semibold">
                    {uiStrings?.metrics_memory || "Memory"}
                  </th>
                  <th className="px-6 py-3 font-semibold">
                    {uiStrings?.metrics_disk || "Disk I/O"}
                  </th>
                  <th className="px-6 py-3 font-semibold">
                    {uiStrings?.metrics_throughput || "Network In/Out"}
                  </th>
                  <th className="px-6 py-3 font-semibold text-right">
                    {uiStrings?.status || "Status"}
                  </th>
                </tr>
              </thead>
              <tbody className="text-sm">
                {loading ? (
                  <tr>
                    <td
                      colSpan={6}
                      className="p-4 text-center text-monitor-text-muted"
                    >
                      {uiStrings?.loading || "Loading cluster data..."}
                    </td>
                  </tr>
                ) : (
                  metrics?.cluster_usage.map((cluster, idx) => (
                    <tr
                      key={idx}
                      className="border-b border-monitor-border hover:bg-monitor-surface-darker/50 transition-colors group"
                    >
                      <td className="px-6 py-4">
                        <div className="flex items-center gap-3">
                          <div className="p-2 bg-monitor-surface-darker rounded border border-monitor-border text-monitor-primary">
                            <span className="material-symbols-outlined text-[18px]">
                              api
                            </span>
                          </div>
                          <div>
                            <div className="text-white font-medium">
                              {cluster.service}
                            </div>
                            <div className="text-[10px] text-monitor-text-muted">
                              us-east-1a
                            </div>
                          </div>
                        </div>
                      </td>
                      <td className="px-6 py-4">
                        <div className="flex items-center gap-3">
                          <div className="text-xs font-mono text-white w-8">
                            {cluster.cpu}%
                          </div>
                          <div className="w-24 bg-monitor-surface-darker h-1.5 rounded-full overflow-hidden border border-monitor-border">
                            <div
                              className="bg-monitor-primary h-full rounded-full"
                              style={{ width: `${cluster.cpu}%` }}
                            ></div>
                          </div>
                        </div>
                      </td>
                      <td className="px-6 py-4">
                        <div className="flex items-center gap-3">
                          <div className="text-xs font-mono text-white w-12">
                            {cluster.memory_gb} GB
                          </div>
                          <div className="w-24 bg-monitor-surface-darker h-1.5 rounded-full overflow-hidden border border-monitor-border">
                            <div
                              className="bg-monitor-success h-full rounded-full"
                              style={{
                                width: `${(cluster as any).memory_percent || 35}%`,
                              }}
                            ></div>
                          </div>
                        </div>
                      </td>
                      <td className="px-6 py-4 text-xs font-mono text-monitor-text-muted">
                        <span className="text-white">125</span> MB/s
                      </td>
                      <td className="px-6 py-4 text-xs font-mono text-monitor-text-muted">
                        <span className="text-monitor-success">↑ 450</span> /{" "}
                        <span className="text-monitor-primary">↓ 1.2G</span>{" "}
                        Mbps
                      </td>
                      <td className="px-6 py-4 text-right">
                        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[10px] font-bold bg-monitor-success/10 text-monitor-success border border-monitor-success/20">
                          {cluster.status}
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
    </main>
  );
};

export default MetricsHub;
