import React from "react";
import PageHeader from "./components/PageHeader";

import { useEffect, useState } from "react";
import { useUIStrings } from "../../hooks/useUIStrings";
import { monitoringService, QueueData } from "../../services/monitoringService";

const QueueMonitor: React.FC = () => {
  const [queueData, setQueueData] = useState<QueueData | null>(null);
  const [loading, setLoading] = useState(true);
  const [systemMeta, setSystemMeta] = useState<any>(null);
  const { uiStrings } = useUIStrings("monitoring");

  const fetchData = async () => {
    setLoading(true);
    try {
      const data = await monitoringService.getQueues();
      setQueueData(data);
    } catch (error) {
      console.error("Failed to fetch queue data", error);
    } finally {
      setLoading(false);
    }
  };

  const fetchConfig = async () => {
    try {
      const res = await fetch("/api/v1/ui/global_config");
      const data = await res.json();
      setSystemMeta(data.system_metadata);
    } catch (e) {
      console.error("Failed to fetch config", e);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);
  return (
    <main className="flex-1 flex flex-col min-w-0 bg-monitor-bg-dark overflow-hidden font-public-sans text-monitor-text-muted">
      <PageHeader
        title={uiStrings?.queues_title || "Background Job & Queue Monitor"}
        subtitle={
          <>
            <span>{uiStrings?.metrics_subtitle || "Monitoring"}</span>
            <span className="material-symbols-outlined text-[10px]">
              chevron_right
            </span>
            <span className="text-monitor-primary">
              {uiStrings?.queues_primary || "Queue Status"}
            </span>
          </>
        }
        actions={
          <>
            <div className="flex items-center gap-2 px-3 py-1.5 rounded bg-monitor-surface border border-monitor-border text-xs text-monitor-text-muted">
              <span className="flex h-2 w-2 rounded-full bg-monitor-success"></span>
              {uiStrings?.status_operational || "System Healthy"}
            </div>
            <button
              onClick={fetchData}
              className="px-3 py-1.5 bg-monitor-surface hover:bg-monitor-surface-darker border border-monitor-border text-white text-xs font-semibold rounded flex items-center gap-2 transition-colors"
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
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
          <div className="bg-monitor-surface rounded-xl border border-monitor-border p-4 relative overflow-hidden group">
            <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
              <span className="material-symbols-outlined text-4xl text-monitor-primary">
                play_circle
              </span>
            </div>
            <div className="text-monitor-text-muted text-xs font-medium uppercase mb-1">
              {uiStrings?.queues_running || "Running Jobs"}
            </div>
            <div className="text-2xl font-bold text-white">
              {loading ? "..." : queueData?.running_jobs}
            </div>
            <div className="text-[10px] text-monitor-success flex items-center gap-1 mt-1">
              <span className="material-symbols-outlined text-[12px]">
                speed
              </span>
              Processing normal
            </div>
          </div>
          <div className="bg-monitor-surface rounded-xl border border-monitor-border p-4 relative overflow-hidden group">
            <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
              <span className="material-symbols-outlined text-4xl text-monitor-warning">
                pending_actions
              </span>
            </div>
            <div className="text-monitor-text-muted text-xs font-medium uppercase mb-1">
              {uiStrings?.queues_pending || "Pending Queue"}
            </div>
            <div className="text-2xl font-bold text-white">
              {loading ? "..." : queueData?.pending_queue.toLocaleString()}
            </div>
            <div className="text-[10px] text-monitor-text-muted mt-1">
              Approx 5m wait time
            </div>
            <div className="w-full bg-monitor-border h-1 rounded-full mt-2 overflow-hidden">
              <div
                className="bg-monitor-warning h-full"
                style={{ width: systemMeta?.queue_warn_level || "0%" }}
              ></div>
            </div>
          </div>
          <div className="bg-monitor-surface rounded-xl border border-monitor-border p-4 relative overflow-hidden group">
            <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
              <span className="material-symbols-outlined text-4xl text-monitor-danger">
                error
              </span>
            </div>
            <div className="text-monitor-text-muted text-xs font-medium uppercase mb-1">
              {uiStrings?.queues_failed || "Failed (24h)"}
            </div>
            <div className="text-2xl font-bold text-monitor-danger">
              {loading ? "..." : queueData?.failed_24h}
            </div>
            <div className="text-[10px] text-monitor-danger mt-1">
              <span className="font-bold">2</span> critical failures
            </div>
            <div className="w-full bg-monitor-border h-1 rounded-full mt-2 overflow-hidden">
              <div
                className="bg-monitor-danger h-full"
                style={{ width: systemMeta?.queue_danger_level || "0%" }}
              ></div>
            </div>
          </div>
          <div className="bg-monitor-surface rounded-xl border border-monitor-border p-4 relative overflow-hidden group">
            <div className="absolute right-0 top-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
              <span className="material-symbols-outlined text-4xl text-monitor-text-muted">
                delete_forever
              </span>
            </div>
            <div className="text-monitor-text-muted text-xs font-medium uppercase mb-1">
              {uiStrings?.queues_dlq || "DLQ Count"}
            </div>
            <div className="text-2xl font-bold text-monitor-text-muted">
              {loading ? "..." : queueData?.dlq_count}
            </div>
            <div className="text-[10px] text-monitor-text-muted mt-1">
              Dead Letter Queue
            </div>
          </div>
        </div>

        <div className="bg-monitor-surface rounded-xl border border-monitor-border flex flex-col overflow-hidden mb-6">
          <div className="p-4 border-b border-monitor-border flex justify-between items-center bg-monitor-surface-darker">
            <h3 className="text-white font-semibold text-sm flex items-center gap-2">
              <span className="material-symbols-outlined text-monitor-primary text-[18px]">
                bolt
              </span>
              {uiStrings?.queues_active_jobs || "Active Background Jobs"}
            </h3>
            <div className="flex items-center gap-2">
              <span className="text-[10px] text-monitor-text-muted uppercase font-bold tracking-wider">
                Auto-refresh: ON
              </span>
            </div>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse">
              <thead>
                <tr className="border-b border-monitor-border text-xs text-monitor-text-muted uppercase tracking-wider bg-monitor-surface-darker/50">
                  <th className="px-4 py-3 font-semibold w-32">
                    {uiStrings?.th_job_id || "Job ID"}
                  </th>
                  <th className="px-4 py-3 font-semibold w-48">
                    {uiStrings?.th_type || "Type"}
                  </th>
                  <th className="px-4 py-3 font-semibold">
                    {uiStrings?.th_progress || "Progress"}
                  </th>
                  <th className="px-4 py-3 font-semibold w-32">
                    {uiStrings?.th_duration || "Duration"}
                  </th>
                  <th className="px-4 py-3 font-semibold text-right w-32">
                    {uiStrings?.th_actions || "Action"}
                  </th>
                </tr>
              </thead>
              <tbody className="text-sm">
                {loading ? (
                  <tr>
                    <td
                      colSpan={5}
                      className="p-4 text-center text-monitor-text-muted"
                    >
                      {uiStrings?.loading || "Loading active jobs..."}
                    </td>
                  </tr>
                ) : (
                  queueData?.active_jobs.map((job) => (
                    <tr
                      key={job.id}
                      className="border-b border-monitor-border hover:bg-monitor-surface-darker/50 transition-colors"
                    >
                      <td className="px-4 py-3 text-white font-mono text-xs">
                        {job.id}
                      </td>
                      <td className="px-4 py-3 text-monitor-text-muted">
                        <span className="inline-flex items-center gap-1.5">
                          <span className="w-1.5 h-1.5 rounded-full bg-indigo-500"></span>
                          {job.type}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-monitor-text-muted">
                        <div className="flex items-center gap-3">
                          <div className="flex-1 bg-monitor-border h-1.5 rounded-full overflow-hidden">
                            <div
                              className="bg-monitor-primary h-full"
                              style={{ width: `${job.progress}%` }}
                            ></div>
                          </div>
                          <span className="text-xs font-medium w-8">
                            {job.progress}%
                          </span>
                        </div>
                      </td>
                      <td className="px-4 py-3 text-monitor-text-muted font-mono text-xs">
                        {job.duration}
                      </td>
                      <td className="px-4 py-3 text-right">
                        <button className="text-monitor-danger hover:text-white text-xs font-medium hover:underline">
                          {uiStrings?.btn_kill || "Kill Process"}
                        </button>
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

export default QueueMonitor;
