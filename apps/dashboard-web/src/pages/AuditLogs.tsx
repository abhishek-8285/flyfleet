import {
    ArrowDownTrayIcon,
    ArrowPathIcon,
    CalendarIcon,
    ChevronRightIcon,
    ExclamationTriangleIcon,
    FunnelIcon,
    InformationCircleIcon, // Placeholder or use heroicons
    MagnifyingGlassIcon,
    ShieldCheckIcon,
    UserIcon,
} from "@heroicons/react/24/outline";
import React, { useEffect, useState } from "react";
import { useAuth } from "../context/AuthProvider";
import { useUIStrings } from "../hooks/useUIStrings";
import { api } from "../services/api";

const AuditLogs: React.FC = () => {
  const { user } = useAuth();
  const [logs, setLogs] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [filters, setFilters] = useState({
    severity: "",
    action: "",
    query: "",
  });
  const { uiStrings } = useUIStrings("audit");
  const [systemMeta, setSystemMeta] = useState<any>(null);

  const fetchLogs = async () => {
    if (!user?.org_id) return;
    setLoading(true);
    try {
      const data = await api.getAuditLogs({
        org_id: user.org_id,
        page,
        severity: filters.severity,
        action: filters.action,
      });
      setLogs(data.logs || []);
      setTotal(data.total);
    } catch (error) {
      console.error("Failed to fetch audit logs", error);
    } finally {
      setLoading(false);
      api
        .get("/ui/global_config")
        .then((res: any) => {
          setSystemMeta(res?.system_metadata || null);
        })
        .catch((err) => console.error("Failed to fetch system config", err));
    }
  };

  useEffect(() => {
    fetchLogs();
  }, [user?.org_id, page, filters.severity, filters.action]);

  const getSeverityColor = (severity: string) => {
    switch (severity.toLowerCase()) {
      case "critical":
        return "text-red-400 bg-red-400/10 border-red-400/20";
      case "warning":
        return "text-amber-400 bg-amber-400/10 border-amber-400/20";
      default:
        return "text-blue-400 bg-blue-400/10 border-blue-400/20";
    }
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return (
      date.toLocaleDateString("en-US", { month: "short", day: "numeric" }) +
      ", " +
      date.toLocaleTimeString("en-US", {
        hour: "2-digit",
        minute: "2-digit",
        hour12: false,
      })
    );
  };

  return (
    <div className="p-8 space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-700">
      {/* Header Section */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-4xl font-light tracking-tight text-white mb-2">
            {uiStrings?.title_prefix || "Audit "}
            <span className="font-semibold text-blue-500">
              {uiStrings?.title_highlight || "Explorer"}
            </span>
          </h1>
          <p className="text-slate-400">
            {uiStrings?.subtitle ||
              "Governance, transparency, and deep system accountability"}
          </p>
        </div>
        <div className="flex items-center gap-3">
          <button
            onClick={() => fetchLogs()}
            className="p-2 bg-slate-800/50 border border-slate-700/50 rounded-xl text-slate-400 hover:text-white transition-all"
          >
            <ArrowPathIcon
              className={`w-5 h-5 ${loading ? "animate-spin" : ""}`}
            />
          </button>
          <button className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-xl font-medium transition-all shadow-lg shadow-blue-500/20">
            <ArrowDownTrayIcon className="w-4 h-4" />
            {uiStrings?.btn_export || "Export Data"}
          </button>
        </div>
      </div>

      {/* Filter Bar */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 bg-slate-900/40 backdrop-blur-xl border border-white/5 p-4 rounded-2xl shadow-2xl">
        <div className="relative col-span-1 md:col-span-2">
          <MagnifyingGlassIcon className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-500" />
          <input
            type="text"
            placeholder={
              uiStrings?.search_placeholder ||
              "Search logs by actor, resource, or details..."
            }
            className="w-full pl-11 pr-4 py-2.5 bg-slate-950/50 border border-slate-800 rounded-xl text-white placeholder:text-slate-600 focus:outline-none focus:ring-2 focus:ring-blue-500/50 transition-all font-light"
          />
        </div>
        <div className="relative">
          <FunnelIcon className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500" />
          <select
            value={filters.severity}
            onChange={(e) =>
              setFilters({ ...filters, severity: e.target.value })
            }
            className="w-full pl-10 pr-4 py-2.5 bg-slate-950/50 border border-slate-800 rounded-xl text-white appearance-none focus:outline-none focus:ring-2 focus:ring-blue-500/50 transition-all font-light cursor-pointer"
          >
            <option value="">
              {uiStrings?.filter_all_severities || "All Severities"}
            </option>
            <option value="info">{uiStrings?.severity_info || "Info"}</option>
            <option value="warning">
              {uiStrings?.severity_warning || "Warning"}
            </option>
            <option value="critical">
              {uiStrings?.severity_critical || "Critical"}
            </option>
          </select>
        </div>
        <div className="relative">
          <CalendarIcon className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500" />
          <select className="w-full pl-10 pr-4 py-2.5 bg-slate-950/50 border border-slate-800 rounded-xl text-white appearance-none focus:outline-none focus:ring-2 focus:ring-blue-500/50 transition-all font-light cursor-pointer">
            <option value="24h">
              {uiStrings?.filter_24h || "Last 24 Hours"}
            </option>
            <option value="7d">{uiStrings?.filter_7d || "Last 7 Days"}</option>
            <option value="30d">
              {uiStrings?.filter_30d || "Last 30 Days"}
            </option>
            <option value="custom">
              {uiStrings?.filter_custom || "Custom Range"}
            </option>
          </select>
        </div>
      </div>

      {/* Logs Table */}
      <div className="bg-slate-900/40 backdrop-blur-xl border border-white/5 rounded-2xl shadow-2xl overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-left">
            <thead>
              <tr className="border-b border-white/5 bg-slate-950/30">
                <th className="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">
                  {uiStrings?.th_timestamp || "Timestamp"}
                </th>
                <th className="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">
                  {uiStrings?.th_severity || "Severity"}
                </th>
                <th className="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">
                  {uiStrings?.th_action || "Action"}
                </th>
                <th className="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">
                  {uiStrings?.th_actor || "Actor"}
                </th>
                <th className="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">
                  {uiStrings?.th_resource || "Resource"}
                </th>
                <th className="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider text-right">
                  {uiStrings?.th_details || "Details"}
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/5">
              {loading ? (
                Array(5)
                  .fill(0)
                  .map((_, i) => (
                    <tr key={i} className="animate-pulse">
                      <td
                        colSpan={6}
                        className="px-6 py-4 h-16 bg-slate-800/10"
                      ></td>
                    </tr>
                  ))
              ) : logs.length === 0 ? (
                <tr>
                  <td
                    colSpan={6}
                    className="px-6 py-12 text-center text-slate-500"
                  >
                    {uiStrings?.no_data ||
                      "No audit logs found matching your filters."}
                  </td>
                </tr>
              ) : (
                logs.map((log) => (
                  <tr
                    key={log.id}
                    className="group hover:bg-white/[0.02] transition-colors cursor-pointer"
                  >
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-slate-400 font-mono">
                      {formatDate(log.created_at)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span
                        className={`px-2 py-0.5 rounded-full text-[10px] font-bold uppercase border ${getSeverityColor(log.severity)}`}
                      >
                        {log.severity}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className="text-slate-200 font-medium">
                        {log.action}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center gap-2">
                        <div className="w-6 h-6 rounded-full bg-slate-800 border border-slate-700 flex items-center justify-center">
                          <UserIcon className="w-3 h-3 text-slate-400" />
                        </div>
                        <span className="text-sm text-slate-300">
                          {uiStrings?.default_actor || "Admin"}
                        </span>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-blue-400 font-light">
                      {log.resource}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right">
                      <button className="p-1.5 text-slate-500 hover:text-white bg-slate-800/0 hover:bg-slate-800 rounded-lg transition-all">
                        <ChevronRightIcon className="w-5 h-5" />
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination */}
        <div className="px-6 py-4 bg-slate-950/30 border-t border-white/5 flex items-center justify-between">
          <p className="text-sm text-slate-500">
            {uiStrings?.label_showing || "Showing"}{" "}
            <span className="text-white font-medium">{logs.length}</span>{" "}
            {uiStrings?.label_of || " of "}{" "}
            <span className="text-white font-medium">{total}</span>{" "}
            {uiStrings?.label_events || " events"}
          </p>
          <div className="flex items-center gap-2">
            <button
              disabled={page === 1}
              onClick={() => setPage(page - 1)}
              className="px-4 py-1.5 bg-slate-800/50 border border-slate-700/50 rounded-lg text-sm text-slate-400 hover:text-white disabled:opacity-50 disabled:cursor-not-allowed transition-all"
            >
              {uiStrings?.btn_prev || "Previous"}
            </button>
            <button
              disabled={logs.length < 50}
              onClick={() => setPage(page + 1)}
              className="px-4 py-1.5 bg-slate-800/50 border border-slate-700/50 rounded-lg text-sm text-slate-400 hover:text-white disabled:opacity-50 disabled:cursor-not-allowed transition-all"
            >
              {uiStrings?.btn_next || "Next"}
            </button>
          </div>
        </div>
      </div>

      {/* Quick Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="bg-slate-900/40 p-6 rounded-2xl border border-white/5 space-y-4">
          <div className="flex items-center justify-between">
            <ShieldCheckIcon className="w-6 h-6 text-green-400" />
            <span className="text-xs font-bold text-green-400/50 uppercase">
              {uiStrings?.stat_health || "Health"}
            </span>
          </div>
          <div>
            <p className="text-2xl font-semibold text-white">
              {systemMeta?.success_rate || "--%"}
            </p>
            <p className="text-sm text-slate-400">
              {uiStrings?.stat_health_sub || "System Integrity Score"}
            </p>
          </div>
        </div>
        <div className="bg-slate-900/40 p-6 rounded-2xl border border-white/5 space-y-4">
          <div className="flex items-center justify-between">
            <ExclamationTriangleIcon className="w-6 h-6 text-amber-400" />
            <span className="text-xs font-bold text-amber-400/50 uppercase">
              {uiStrings?.stat_anomalies || "Anomalies"}
            </span>
          </div>
          <div>
            <p className="text-2xl font-semibold text-white">02</p>
            <p className="text-sm text-slate-400">
              {uiStrings?.stat_anomalies_sub || "Flagged Events (24h)"}
            </p>
          </div>
        </div>
        <div className="bg-slate-900/40 p-6 rounded-2xl border border-white/5 space-y-4">
          <div className="flex items-center justify-between">
            <InformationCircleIcon className="w-6 h-6 text-blue-400" />
            <span className="text-xs font-bold text-blue-400/50 uppercase">
              {uiStrings?.stat_velocity || "Velocity"}
            </span>
          </div>
          <div>
            <p className="text-2xl font-semibold text-white">12.4k</p>
            <p className="text-sm text-slate-400">
              {uiStrings?.stat_velocity_sub || "Total Indexed Events"}
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default AuditLogs;
