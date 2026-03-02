import {
    AdjustmentsHorizontalIcon,
    ArrowPathIcon,
    ChevronDownIcon,
    ChevronRightIcon,
    EyeIcon,
    InformationCircleIcon,
    MagnifyingGlassIcon,
    UserCircleIcon,
} from "@heroicons/react/24/outline";
import React, { useEffect, useState } from "react";
import { useAuth } from "../../context/AuthProvider";
import { api } from "../../services/api";

const AuditList: React.FC = () => {
  const { user } = useAuth();
  const [logs, setLogs] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [uiStrings, setUiStrings] = useState<any>(null);

  useEffect(() => {
    fetch("/api/v1/ui/global_config")
      .then((res) => res.json())
      .then((data) =>
        setUiStrings(
          data.ui_strings?.governance ||
            data.system_metadata?.ui_strings?.governance,
        ),
      )
      .catch((err) => console.error(err));
  }, []);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [filters, setFilters] = useState({
    severity: "",
    module: "",
    action: "",
  });
  const [selectedLog, setSelectedLog] = useState<any | null>(null);

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
    }
  };

  useEffect(() => {
    fetchLogs();
  }, [user?.org_id, page, filters]);

  const getSeverityStyles = (severity: string) => {
    switch (severity?.toLowerCase()) {
      case "critical":
        return "text-red-400 bg-red-400/10 border-red-400/20";
      case "warning":
        return "text-amber-400 bg-amber-400/10 border-amber-400/20";
      case "info":
        return "text-blue-400 bg-blue-400/10 border-blue-400/20";
      default:
        return "text-slate-400 bg-slate-400/10 border-slate-400/20";
    }
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return (
      date.toLocaleDateString("en-US", {
        month: "short",
        day: "numeric",
        year: "numeric",
      }) +
      " " +
      date.toLocaleTimeString("en-US", {
        hour: "2-digit",
        minute: "2-digit",
        hour12: false,
      })
    );
  };

  return (
    <div className="space-y-6">
      {/* Filters */}
      <div className="flex flex-wrap items-center gap-4 bg-slate-900/60 backdrop-blur-xl p-4 rounded-2xl border border-white/5">
        <div className="flex-1 min-w-[240px] relative">
          <MagnifyingGlassIcon className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-500" />
          <input
            type="text"
            placeholder={
              uiStrings?.search_placeholder || "Filter by action or resource..."
            }
            className="w-full pl-11 pr-4 py-2 bg-slate-950/50 border border-slate-800 rounded-xl text-white placeholder:text-slate-600 focus:outline-none focus:ring-2 focus:ring-blue-500/50 transition-all font-light text-sm"
          />
        </div>

        <div className="flex items-center gap-3">
          <div className="relative">
            <select
              value={filters.module}
              onChange={(e) =>
                setFilters({ ...filters, module: e.target.value })
              }
              className="pl-4 pr-10 py-2 bg-slate-950/50 border border-slate-800 rounded-xl text-white text-sm appearance-none focus:outline-none focus:ring-2 focus:ring-blue-500/50 transition-all font-light cursor-pointer min-w-[140px]"
            >
              <option value="">
                {uiStrings?.all_modules || "All Modules"}
              </option>
              <option value="AUTH">{uiStrings?.mod_auth || "Auth"}</option>
              <option value="RBAC">{uiStrings?.mod_rbac || "RBAC"}</option>
              <option value="FLEET">{uiStrings?.mod_fleet || "Fleet"}</option>
              <option value="DR">{uiStrings?.mod_dr || "DR"}</option>
              <option value="PROFILE">
                {uiStrings?.mod_profile || "Profile"}
              </option>
            </select>
            <ChevronDownIcon className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500 pointer-events-none" />
          </div>

          <div className="relative">
            <select
              value={filters.severity}
              onChange={(e) =>
                setFilters({ ...filters, severity: e.target.value })
              }
              className="pl-4 pr-10 py-2 bg-slate-950/50 border border-slate-800 rounded-xl text-white text-sm appearance-none focus:outline-none focus:ring-2 focus:ring-blue-500/50 transition-all font-light cursor-pointer min-w-[140px]"
            >
              <option value="">
                {uiStrings?.all_severities || "All Severities"}
              </option>
              <option value="info">
                {uiStrings?.opt_info_short || "Info"}
              </option>
              <option value="warning">
                {uiStrings?.opt_warning_short || "Warning"}
              </option>
              <option value="critical">
                {uiStrings?.opt_critical_short || "Critical"}
              </option>
            </select>
            <ChevronDownIcon className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500 pointer-events-none" />
          </div>

          <button
            onClick={() => fetchLogs()}
            className="p-2 bg-blue-600/10 border border-blue-500/20 text-blue-400 rounded-xl hover:bg-blue-600/20 transition-all"
          >
            <ArrowPathIcon
              className={`w-5 h-5 ${loading ? "animate-spin" : ""}`}
            />
          </button>
        </div>
      </div>

      {/* Table */}
      <div className="bg-slate-900/40 backdrop-blur-xl border border-white/5 rounded-2xl shadow-2xl overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-left">
            <thead>
              <tr className="border-b border-white/5 bg-slate-950/30">
                <th className="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">
                  {uiStrings?.th_timestamp || "Timestamp"}
                </th>
                <th className="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">
                  {uiStrings?.th_module || "Module"}
                </th>
                <th className="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">
                  {uiStrings?.th_action || "Action"}
                </th>
                <th className="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider text-right">
                  {uiStrings?.th_actions || "Actions"}
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/5">
              {loading ? (
                Array(10)
                  .fill(0)
                  .map((_, i) => (
                    <tr key={i} className="animate-pulse">
                      <td
                        colSpan={4}
                        className="px-6 py-4 h-14 bg-slate-800/10"
                      ></td>
                    </tr>
                  ))
              ) : logs.length === 0 ? (
                <tr>
                  <td
                    colSpan={4}
                    className="px-6 py-12 text-center text-slate-500"
                  >
                    {uiStrings?.no_entries || "No audit entries found."}
                  </td>
                </tr>
              ) : (
                logs.map((log) => (
                  <tr
                    key={log.id}
                    className={`group hover:bg-white/[0.02] transition-colors cursor-pointer ${selectedLog?.id === log.id ? "bg-blue-500/5" : ""}`}
                    onClick={() => setSelectedLog(log)}
                  >
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-slate-400 font-mono">
                      {formatDate(log.created_at)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className="px-2 py-0.5 rounded-lg bg-slate-800 text-slate-300 text-[10px] font-bold tracking-tight uppercase border border-slate-700">
                        {log.module || "SYSTEM"}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex flex-col">
                        <span className="text-white text-sm font-medium">
                          {log.action}
                        </span>
                        <span className="text-slate-500 text-xs font-light">
                          {log.resource}
                        </span>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right">
                      <button className="p-2 text-slate-500 hover:text-blue-400 hover:bg-blue-500/10 rounded-xl transition-all">
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
          <p className="text-xs text-slate-500 font-light">
            {uiStrings?.showing_label || "Showing"}{" "}
            <span className="text-slate-300 font-medium">{logs.length}</span>{" "}
            {uiStrings?.results_label || "results"}
          </p>
          <div className="flex items-center gap-2">
            <button
              disabled={page === 1}
              onClick={() => setPage(page - 1)}
              className="px-4 py-1.5 bg-slate-800/50 border border-slate-700/50 rounded-lg text-xs text-slate-400 hover:text-white disabled:opacity-50 disabled:cursor-not-allowed transition-all"
            >
              {uiStrings?.prev_btn || "Previous"}
            </button>
            <button
              disabled={logs.length < 50}
              onClick={() => setPage(page + 1)}
              className="px-4 py-1.5 bg-slate-800/50 border border-slate-700/50 rounded-lg text-xs text-slate-400 hover:text-white disabled:opacity-50 disabled:cursor-not-allowed transition-all"
            >
              {uiStrings?.next_btn || "Next"}
            </button>
          </div>
        </div>
      </div>

      {/* Sidebar/Drawer for Details */}
      {selectedLog && (
        <div className="fixed inset-y-0 right-0 w-full md:w-[480px] bg-slate-950/95 backdrop-blur-2xl border-l border-white/10 z-50 p-8 shadow-2xl animate-in slide-in-from-right duration-300">
          <div className="flex items-center justify-between mb-8">
            <h2 className="text-xl font-semibold text-white">
              {uiStrings?.details_title || "Event Intelligence"}
            </h2>
            <button
              onClick={() => setSelectedLog(null)}
              className="p-2 text-slate-500 hover:text-white hover:bg-white/5 rounded-full transition-all"
            >
              <ChevronRightIcon className="w-6 h-6 rotate-180" />
            </button>
          </div>

          <div className="space-y-6 overflow-y-auto max-h-[calc(100vh-160px)] pr-2">
            <div className="p-4 bg-slate-900/50 rounded-2xl border border-white/5 space-y-4">
              <div className="flex items-center justify-between">
                <span className="text-xs text-slate-500 font-bold uppercase tracking-widest">
                  {uiStrings?.meta_label || "Metadata"}
                </span>
                <span
                  className={`px-2 py-0.5 rounded-full text-[10px] font-bold uppercase border ${getSeverityStyles(selectedLog.severity)}`}
                >
                  {selectedLog.severity}
                </span>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-1">
                  <p className="text-xs text-slate-500">
                    {uiStrings?.id_label || "ID"}
                  </p>
                  <p className="text-sm text-slate-300 font-mono truncate">
                    {selectedLog.id}
                  </p>
                </div>
                <div className="space-y-1">
                  <p className="text-xs text-slate-500">
                    {uiStrings?.actor_label || "Actor"}
                  </p>
                  <div className="flex items-center gap-2">
                    <UserCircleIcon className="w-4 h-4 text-slate-400" />
                    <p className="text-sm text-slate-300 font-light truncate">
                      {selectedLog.user_id}
                    </p>
                  </div>
                </div>
                <div className="space-y-1">
                  <p className="text-xs text-slate-500">
                    {uiStrings?.ip_label || "IP Address"}
                  </p>
                  <p className="text-sm text-slate-300 font-mono">
                    {selectedLog.ip_address || "Internal"}
                  </p>
                </div>
                <div className="space-y-1">
                  <p className="text-xs text-slate-500">
                    {uiStrings?.status_label || "Status"}
                  </p>
                  <span
                    className={`text-sm font-medium ${selectedLog.status === "success" ? "text-green-400" : "text-red-400"}`}
                  >
                    {selectedLog.status}
                  </span>
                </div>
              </div>
            </div>

            {/* Details/Payload */}
            <div className="space-y-2">
              <h3 className="text-sm font-semibold text-slate-300 flex items-center gap-2">
                <AdjustmentsHorizontalIcon className="w-4 h-4" />
                {uiStrings?.action_details || "Action Details"}
              </h3>
              <div className="p-4 bg-slate-950 border border-slate-800 rounded-xl">
                <pre className="text-xs text-blue-300/80 font-mono overflow-x-auto whitespace-pre-wrap">
                  {JSON.stringify(
                    JSON.parse(selectedLog.details || "{}"),
                    null,
                    2,
                  )}
                </pre>
              </div>
            </div>

            {/* State Change (If exists) */}
            {(selectedLog.before_state || selectedLog.after_state) && (
              <div className="space-y-4 pt-4 border-t border-white/5">
                <h3 className="text-sm font-semibold text-slate-300 flex items-center gap-2">
                  <EyeIcon className="w-4 h-4" />
                  {uiStrings?.state_snapshot || "State Snapshot"}
                </h3>
                <div className="grid grid-cols-1 gap-4">
                  {selectedLog.before_state && (
                    <div className="space-y-2">
                      <p className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">
                        {uiStrings?.before_label || "Before"}
                      </p>
                      <div className="p-3 bg-red-500/5 border border-red-500/10 rounded-xl text-[10px] text-red-300/60 font-mono">
                        <pre className="whitespace-pre-wrap">
                          {JSON.stringify(selectedLog.before_state, null, 2)}
                        </pre>
                      </div>
                    </div>
                  )}
                  {selectedLog.after_state && (
                    <div className="space-y-2">
                      <p className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">
                        {uiStrings?.after_label || "After"}
                      </p>
                      <div className="p-3 bg-green-500/5 border border-green-500/10 rounded-xl text-[10px] text-green-300/60 font-mono">
                        <pre className="whitespace-pre-wrap">
                          {JSON.stringify(selectedLog.after_state, null, 2)}
                        </pre>
                      </div>
                    </div>
                  )}
                </div>
              </div>
            )}

            <div className="p-4 bg-blue-600/10 rounded-2xl border border-blue-500/20 flex items-start gap-4">
              <InformationCircleIcon className="w-6 h-6 text-blue-400 shrink-0 mt-0.5" />
              <div>
                <p className="text-sm font-medium text-blue-400">
                  {uiStrings?.context_verify || "Contextual Verification"}
                </p>
                <p className="text-xs text-blue-400/60 leading-relaxed mt-1">
                  {uiStrings?.org_context ||
                    "This event was captured under organization"}{" "}
                  {selectedLog.org_id}.
                  {uiStrings?.agent_context ||
                    "Metadata suggests the request originated from"}{" "}
                  {selectedLog.user_agent ||
                    uiStrings?.terminal_fallback ||
                    "a secure system terminal"}
                  .
                </p>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default AuditList;
