import {
    ArrowPathIcon,
    ChevronLeftIcon,
    ChevronRightIcon,
    ComputerDesktopIcon,
    DevicePhoneMobileIcon,
    FingerPrintIcon,
    FunnelIcon,
    MagnifyingGlassIcon,
    PauseIcon,
    PlayIcon,
} from "@heroicons/react/24/outline";
import React, { useEffect, useState } from "react";
import SecurityFooter from "../../../components/SecurityFooter";
import SecuritySidebar from "../../../components/SecuritySidebar";
import { useUIStrings } from "../../../hooks/useUIStrings";
import apiClient from "../../../services/apiClient";

const LoginActivityLogs: React.FC = () => {
  const { uiStrings: rawUiStrings } = useUIStrings("admin_security");
  const uiStrings = rawUiStrings || {};
  const [logs, setLogs] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [query, setQuery] = useState("");
  const [isLive, setIsLive] = useState(false);
  const [page, setPage] = useState(1);

  const orgId = localStorage.getItem("org_id") || "default";

  useEffect(() => {
    fetchLogs();

    let interval: any;
    if (isLive) {
      interval = setInterval(fetchLogs, 5000);
    }
    return () => clearInterval(interval);
  }, [isLive, page, orgId]);

  const fetchLogs = async () => {
    try {
      if (!isLive) setLoading(true);
      const response = (await apiClient.get("/security/logins", {
        params: { org_id: orgId, page, size: 20 },
      })) as any;
      const data = response?.data ?? response;
      setLogs(Array.isArray(data) ? data : data?.items || []);
    } catch (error) {
      console.error("Failed to fetch login logs", error);
      setLogs([]);
    } finally {
      setLoading(false);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status.toUpperCase()) {
      case "SUCCESS":
        return "text-emerald-400 bg-emerald-500/10 border-emerald-500/20";
      case "FAILED":
        return "text-vibrant-red bg-vibrant-red/10 border-vibrant-red/20";
      case "CHALLENGED":
        return "text-amber-400 bg-amber-500/10 border-amber-500/20";
      default:
        return "text-slate-400 bg-slate-500/10 border-slate-500/20";
    }
  };

  const getRiskLabel = (risk: string) => {
    switch (risk.toLowerCase()) {
      case "high":
        return (
          <span className="text-vibrant-red font-bold">
            {uiStrings.risk_high || "High Risk"}
          </span>
        );
      case "medium":
        return (
          <span className="text-amber-400 font-bold">
            {uiStrings.risk_medium || "Medium Risk"}
          </span>
        );
      default:
        return (
          <span className="text-emerald-400 font-bold">
            {uiStrings.risk_low || "Low Risk"}
          </span>
        );
    }
  };

  return (
    <div className="flex h-screen bg-[#0b1117] text-slate-200 overflow-hidden font-inter">
      <SecuritySidebar />

      <main className="flex-1 flex flex-col min-w-0 bg-grid-pattern relative">
        <div className="glow-effect" />

        <header className="px-8 py-6 border-b border-white/5 bg-slate-900/40 backdrop-blur-md z-10 flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-bold text-white flex items-center gap-3">
              <span className="material-symbols-outlined text-primary text-3xl">
                history
              </span>
              {uiStrings.title_logs || "Login Activity Logs"}
            </h2>
            <p className="text-slate-400 text-sm mt-1">
              {uiStrings.subtitle_logs ||
                "Real-time authentication monitoring across the enterprise fleet."}
            </p>
          </div>

          <div className="flex items-center gap-3">
            <button
              onClick={() => setIsLive(!isLive)}
              className={`flex items-center gap-2 px-4 py-2 rounded-lg text-xs font-black uppercase tracking-widest transition-all border ${isLive ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20" : "bg-slate-800 text-slate-400 border-white/10"}`}
            >
              {isLive ? (
                <PauseIcon className="w-4 h-4" />
              ) : (
                <PlayIcon className="w-4 h-4" />
              )}
              {isLive
                ? uiStrings.btn_live || "LIVE STREAMING"
                : uiStrings.btn_paused || "PAUSED"}
            </button>
            <div className="h-8 w-px bg-white/5 mx-2" />
            <div className="text-right">
              <p className="text-[10px] text-slate-500 uppercase font-black tracking-widest">
                {uiStrings.label_success_rate || "Success Rate"}
              </p>
              <p className="text-sm font-bold text-emerald-400">99.2%</p>
            </div>
          </div>
        </header>

        <div className="p-8 z-10 flex flex-col gap-6 overflow-hidden flex-1">
          <div className="glass-card rounded-2xl flex flex-col overflow-hidden">
            <div className="p-4 border-b border-white/5 bg-white/5 flex items-center justify-between">
              <div className="relative w-80">
                <MagnifyingGlassIcon className="w-5 h-5 absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" />
                <input
                  type="text"
                  placeholder="Filter by identity or IP..."
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  className="w-full bg-slate-900/50 border border-white/10 rounded-lg py-1.5 pl-10 pr-4 text-sm focus:outline-none focus:ring-1 focus:ring-primary/50"
                />
              </div>
              <div className="flex items-center gap-2">
                <button className="p-2 text-slate-400 hover:text-white transition-colors">
                  <FunnelIcon className="w-5 h-5" />
                </button>
                <button
                  onClick={fetchLogs}
                  className="p-2 text-slate-400 hover:text-white transition-colors"
                >
                  <ArrowPathIcon className="w-5 h-5" />
                </button>
              </div>
            </div>

            <div className="flex-1 overflow-auto custom-scrollbar">
              <table className="w-full text-left">
                <thead>
                  <tr className="border-b border-white/5 bg-white/[0.02]">
                    <th className="px-6 py-4 text-[10px] font-bold uppercase text-slate-500">
                      {uiStrings.th_timestamp || "Timestamp"}
                    </th>
                    <th className="px-6 py-4 text-[10px] font-bold uppercase text-slate-500">
                      {uiStrings.th_identity || "Identity"}
                    </th>
                    <th className="px-6 py-4 text-[10px] font-bold uppercase text-slate-500">
                      {uiStrings.th_device_context || "Device/Context"}
                    </th>
                    <th className="px-6 py-4 text-[10px] font-bold uppercase text-slate-500">
                      {uiStrings.th_location || "Location"}
                    </th>
                    <th className="px-6 py-4 text-[10px] font-bold uppercase text-slate-500">
                      {uiStrings.th_status || "Status"}
                    </th>
                    <th className="px-6 py-4 text-[10px] font-bold uppercase text-slate-500">
                      {uiStrings.th_risk || "Risk"}
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-white/5">
                  {loading && logs.length === 0 ? (
                    <tr>
                      <td
                        colSpan={6}
                        className="px-6 py-10 text-center text-slate-500 italic bg-slate-900/20"
                      >
                        {uiStrings.loading_logs ||
                          "Decrypting security logs..."}
                      </td>
                    </tr>
                  ) : logs.length === 0 ? (
                    <tr>
                      <td
                        colSpan={6}
                        className="px-6 py-10 text-center text-slate-500 italic"
                      >
                        {uiStrings.no_logs ||
                          "No matching authentication events found."}
                      </td>
                    </tr>
                  ) : (
                    logs.map((log) => (
                      <tr
                        key={log.id || log._id}
                        className="hover:bg-white/[0.02] transition-colors border-l-2 border-transparent hover:border-primary/30"
                      >
                        <td className="px-6 py-4 whitespace-nowrap">
                          <div className="flex flex-col">
                            <span className="text-sm font-bold text-white">
                              {log.timestamp || "Just now"}
                            </span>
                            <span className="text-[10px] text-slate-500 uppercase tracking-tighter">
                              UTC OFFSET +00:00
                            </span>
                          </div>
                        </td>
                        <td className="px-6 py-4">
                          <div className="flex items-center gap-3">
                            <div className="w-8 h-8 rounded-lg bg-slate-800 flex items-center justify-center border border-white/5">
                              <FingerPrintIcon className="w-4 h-4 text-slate-400" />
                            </div>
                            <div className="flex flex-col">
                              <span className="text-sm font-bold text-white leading-tight">
                                {log.user_name || log.email}
                              </span>
                              <span className="text-[10px] text-slate-500 font-mono tracking-widest uppercase">
                                {log.user_id?.substring(0, 8)}
                              </span>
                            </div>
                          </div>
                        </td>
                        <td className="px-6 py-4">
                          <div className="flex items-center gap-2">
                            {log.device_type === "mobile" ? (
                              <DevicePhoneMobileIcon className="w-3.5 h-3.5 opacity-50" />
                            ) : (
                              <ComputerDesktopIcon className="w-3.5 h-3.5 opacity-50" />
                            )}
                            <span className="text-xs text-slate-300">
                              {log.browser || "Browser"} / {log.os || "OS"}
                            </span>
                          </div>
                          <p className="text-[10px] text-slate-500 mt-0.5 font-mono">
                            {log.ip}
                          </p>
                        </td>
                        <td className="px-6 py-4">
                          <div className="flex items-center gap-2 text-xs text-slate-400 italic">
                            <span className="material-symbols-outlined text-sm">
                              location_on
                            </span>
                            {log.location || uiStrings.loc_unknown || "Unknown"}
                          </div>
                        </td>
                        <td className="px-6 py-4 text-center">
                          <span
                            className={`px-2 py-0.5 rounded text-[9px] font-black uppercase tracking-widest border ${getStatusColor(log.status)}`}
                          >
                            {log.status}
                          </span>
                        </td>
                        <td className="px-6 py-4">
                          <div className="flex items-center gap-2 text-[10px] uppercase tracking-widest">
                            {getRiskLabel(log.risk || "low")}
                          </div>
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>

            <div className="p-4 border-t border-white/5 bg-white/5 flex items-center justify-between">
              <p className="text-[10px] text-slate-500 font-bold uppercase tracking-widest">
                {uiStrings.label_showing_events
                  ?.replace("{start}", "1")
                  .replace("{end}", "20")
                  .replace("{total}", "1,240") ||
                  "Showing 1-20 of 1,240 events"}
              </p>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => setPage(Math.max(1, page - 1))}
                  className="p-1.5 rounded-lg border border-white/10 text-slate-400 hover:text-white hover:bg-white/5 transition-all"
                >
                  <ChevronLeftIcon className="w-4 h-4" />
                </button>
                <span className="text-xs font-bold text-white px-2">
                  {uiStrings.label_page || "PAGE"} {page}
                </span>
                <button
                  onClick={() => setPage(page + 1)}
                  className="p-1.5 rounded-lg border border-white/10 text-slate-400 hover:text-white hover:bg-white/5 transition-all"
                >
                  <ChevronRightIcon className="w-4 h-4" />
                </button>
              </div>
            </div>
          </div>
        </div>

        <SecurityFooter />
      </main>
    </div>
  );
};

export default LoginActivityLogs;
