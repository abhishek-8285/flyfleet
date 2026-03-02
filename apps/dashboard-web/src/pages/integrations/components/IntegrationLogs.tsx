import {
    ArrowsUpDownIcon,
    EyeIcon,
    FunnelIcon,
    MagnifyingGlassIcon,
} from "@heroicons/react/24/outline";
import React, { useState } from "react";
import { useUIStrings } from "../../../hooks/useUIStrings";

const IntegrationLogs: React.FC = () => {
  const [selectedLog, setSelectedLog] = useState<string | null>(null);

  const [logs, setLogs] = React.useState<any[]>([]);
  const [loading_logs, setLoading_logs] = React.useState(true);
  const { uiStrings } = useUIStrings("integrations");

  React.useEffect(() => {
    const fetchLogs = async () => {
      try {
        const res = await fetch(
          `{process.env.REACT_APP_API_URL || '/api/v1'}/api/v1/integrations/integrationlogs`,
        );
        const payload = await res.json();
        setLogs(payload.data || payload || []);
      } catch (error) {
        console.error("Failed to fetch logs", error);
      } finally {
        setLoading_logs(false);
      }
    };
    fetchLogs();
  }, []);

  return (
    <div className="space-y-6">
      {/* Filters */}
      <div className="bg-slate-900/40 border border-slate-800 p-4 rounded-2xl flex flex-wrap items-center gap-4 backdrop-blur-sm">
        <div className="flex-1 relative">
          <MagnifyingGlassIcon className="w-5 h-5 absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" />
          <input
            type="text"
            placeholder={
              uiStrings?.filter_logs_placeholder ||
              "Filter logs by endpoint or ID..."
            }
            className="w-full bg-slate-950/50 border border-slate-800 rounded-xl pl-10 pr-4 py-2 text-sm text-white focus:ring-2 focus:ring-primary/50 outline-none transition-all"
          />
        </div>
        <div className="flex gap-2">
          <button className="flex items-center gap-2 px-4 py-2 bg-slate-800 hover:bg-slate-700 text-white rounded-xl text-xs font-semibold border border-white/5 transition-all">
            <FunnelIcon className="w-4 h-4" />
            {uiStrings?.btn_source_integration || "Source Integration"}
          </button>
          <button className="flex items-center gap-2 px-4 py-2 bg-slate-800 hover:bg-slate-700 text-white rounded-xl text-xs font-semibold border border-white/5 transition-all">
            <ArrowsUpDownIcon className="w-4 h-4" />
            {uiStrings?.btn_newest_first || "Newest First"}
          </button>
        </div>
      </div>

      <div className="bg-slate-900/40 border border-slate-800 rounded-2xl overflow-hidden backdrop-blur-sm">
        <div className="overflow-x-auto">
          <table className="w-full text-left">
            <thead>
              <tr className="bg-slate-950/50 text-[10px] font-bold text-slate-500 uppercase tracking-widest border-b border-white/5">
                <th className="px-6 py-4">
                  {uiStrings?.th_status_method || "Status / Method"}
                </th>
                <th className="px-6 py-4">
                  {uiStrings?.th_int_endpoint || "Integration / Endpoint"}
                </th>
                <th className="px-6 py-4">
                  {uiStrings?.th_duration || "Duration"}
                </th>
                <th className="px-6 py-4">
                  {uiStrings?.th_timestamp || "Timestamp"}
                </th>
                <th className="px-6 py-4 text-right">
                  {uiStrings?.th_actions || "Actions"}
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/5">
              {logs.map((log) => (
                <tr
                  key={log.id}
                  className="hover:bg-white/[0.02] transition-colors group"
                >
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-3">
                      {log.type === "success" ? (
                        <div className="w-2 h-2 rounded-full bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.5)]" />
                      ) : (
                        <div className="w-2 h-2 rounded-full bg-rose-500 shadow-[0_0_8px_rgba(244,63,94,0.5)] animate-pulse" />
                      )}
                      <span className="text-xs font-bold text-white font-mono">
                        {log.status}
                      </span>
                      <span
                        className={`text-[10px] px-1.5 py-0.5 rounded border ${
                          log.method === "POST"
                            ? "bg-indigo-500/10 border-indigo-500/20 text-indigo-400"
                            : "bg-slate-800 border-white/5 text-slate-400"
                        } font-bold`}
                      >
                        {log.method}
                      </span>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="text-sm font-semibold text-white">
                      {log.integration}
                    </div>
                    <div className="text-[10px] text-slate-500 font-mono truncate max-w-md">
                      {log.endpoint}
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <span className="text-xs text-slate-400 font-mono">
                      {log.duration}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    <span className="text-xs text-slate-500 whitespace-nowrap">
                      {log.timestamp}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-right">
                    <button className="p-2 text-slate-500 hover:text-primary hover:bg-primary/10 rounded-xl transition-all">
                      <EyeIcon className="w-4 h-4" />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {/* Pagination Placeholder */}
        <div className="p-4 border-t border-white/5 bg-slate-950/30 flex justify-center">
          <button className="text-[11px] font-bold text-slate-500 hover:text-white transition-colors uppercase tracking-widest">
            {uiStrings?.load_more_btn || "Load More Entries"}
          </button>
        </div>
      </div>
    </div>
  );
};

export default IntegrationLogs;
