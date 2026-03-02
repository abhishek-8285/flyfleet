import {
    ArrowPathIcon,
    BellAlertIcon,
    CheckCircleIcon,
    ClockIcon,
    NoSymbolIcon,
    ShieldExclamationIcon,
    UserGroupIcon
} from "@heroicons/react/24/outline";
import React, { useEffect, useState } from "react";
import SecurityFooter from "../../../components/SecurityFooter";
import SecuritySidebar from "../../../components/SecuritySidebar";
import { useUIStrings } from "../../../hooks/useUIStrings";
import apiClient from "../../../services/apiClient";

const ThreatAlerts: React.FC = () => {
  const { uiStrings: rawUiStrings } = useUIStrings("admin_security");
  const uiStrings = rawUiStrings || {};
  const [alerts, setAlerts] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [updating, setUpdating] = useState<string | null>(null);

  useEffect(() => {
    fetchAlerts();
  }, []);

  const fetchAlerts = async () => {
    try {
      setLoading(true);
      const response = (await apiClient.get("/security/alerts")) as any;
      const data = response?.data ?? response;
      setAlerts(Array.isArray(data) ? data : data?.items || []);
    } catch (error) {
      console.error("Failed to fetch alerts", error);
      setAlerts([]);
    } finally {
      setLoading(false);
    }
  };

  const handleUpdateStatus = async (id: string, status: string) => {
    try {
      setUpdating(id);
      await apiClient.patch(`/security/alerts/${id}/status`, { status });
      setAlerts((prev) => prev.filter((a) => (a.id || a._id) !== id));
      alert(`Alert ${status} successfully.`);
    } catch (error) {
      alert("Failed to update alert status.");
    } finally {
      setUpdating(null);
    }
  };

  const getSeverityStyles = (severity: string) => {
    switch (severity.toLowerCase()) {
      case "critical":
        return "border-vibrant-red/50 bg-vibrant-red/10 text-vibrant-red shadow-lg shadow-vibrant-red/10";
      case "high":
        return "border-amber-500/50 bg-amber-500/10 text-amber-500";
      default:
        return "border-blue-500/50 bg-blue-500/10 text-blue-500";
    }
  };

  const getTypeIcon = (type: string) => {
    switch (type.toLowerCase()) {
      case "brute_force":
        return <ShieldExclamationIcon className="w-6 h-6" />;
      case "suspicious_ip":
        return <NoSymbolIcon className="w-6 h-6" />;
      default:
        return <BellAlertIcon className="w-6 h-6" />;
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
              <span className="material-symbols-outlined text-vibrant-red text-3xl animate-pulse">
                gpp_maybe
              </span>
              {uiStrings.title_threat || "Active Threat Engine"}
            </h2>
            <p className="text-slate-400 text-sm mt-1">
              {uiStrings.subtitle_threat ||
                "Heuristic analysis of authentication patterns and network anomalies."}
            </p>
          </div>

          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-emerald-500/10 border border-emerald-500/20">
              <span className="w-2 h-2 rounded-full bg-emerald-500 animate-ping" />
              <span className="text-[10px] font-bold text-emerald-400 uppercase tracking-widest">
                {uiStrings.label_analysts || "Analysts Online"}:{" "}
                {Math.floor(Math.random() * 5) + 3}
              </span>
            </div>
          </div>
        </header>

        <div className="p-8 z-10 flex flex-col gap-6 overflow-auto custom-scrollbar flex-1">
          <div className="grid grid-cols-3 gap-6">
            <div className="glass-card p-5 rounded-2xl border-l-4 border-vibrant-red">
              <p className="text-xs font-bold text-slate-500 uppercase tracking-widest">
                {uiStrings.card_critical || "Critical Threats"}
              </p>
              <p className="text-3xl font-black text-white mt-1">
                {alerts.filter((a) => a.severity === "CRITICAL").length}
              </p>
            </div>
            <div className="glass-card p-5 rounded-2xl border-l-4 border-amber-500">
              <p className="text-xs font-bold text-slate-500 uppercase tracking-widest">
                {uiStrings.card_open || "Open Alerts"}
              </p>
              <p className="text-3xl font-black text-white mt-1">
                {alerts.length}
              </p>
            </div>
            <div className="glass-card p-5 rounded-2xl border-l-4 border-vibrant-blue">
              <p className="text-xs font-bold text-slate-500 uppercase tracking-widest">
                {uiStrings.card_blocked || "Blocked Today"}
              </p>
              <p className="text-3xl font-black text-white mt-1">128</p>
            </div>
          </div>

          <div className="flex flex-col gap-4">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-bold text-white flex items-center gap-2">
                <ArrowPathIcon className="w-5 h-5 text-primary animate-spin-slow" />
                {uiStrings.feed_title || "Live Detection Feed"}
              </h3>
            </div>

            {loading ? (
              <div className="py-20 text-center text-slate-500">
                {uiStrings.loading_alerts || "Loading alerts..."}
              </div>
            ) : alerts.length === 0 ? (
              <div className="py-20 text-center text-slate-500 flex flex-col items-center gap-4">
                <CheckCircleIcon className="w-12 h-12 text-emerald-500/20" />
                {uiStrings.no_alerts || "No threat alerts found."}
              </div>
            ) : (
              alerts.map((alert) => (
                <div
                  key={alert.id || alert._id}
                  className={`glass-card p-5 rounded-2xl border flex items-start gap-5 transition-all hover:scale-[1.01] ${getSeverityStyles(
                    alert.severity,
                  )}`}
                >
                  <div className="p-3 rounded-xl bg-white/5">
                    {getTypeIcon(alert.type)}
                  </div>
                  <div className="flex-1">
                    <div className="flex items-center justify-between">
                      <h4 className="font-black text-lg uppercase tracking-tight">
                        {alert.title || alert.type}
                      </h4>
                      <span className="text-[10px] font-mono text-white/50">
                        {alert.timestamp || "Just now"}
                      </span>
                    </div>
                    <p className="text-sm text-slate-300 mt-1 leading-relaxed">
                      {alert.description}
                    </p>

                    <div className="mt-4 flex items-center gap-6">
                      <div className="flex items-center gap-2">
                        <UserGroupIcon className="w-4 h-4 opacity-50" />
                        <span className="text-xs font-bold">
                          {uiStrings.label_ip || "IP:"} {alert.ip || "Unknown"}
                        </span>
                      </div>
                      <div className="flex items-center gap-2">
                        <span className="material-symbols-outlined text-sm opacity-50">
                          location_on
                        </span>
                        <span className="text-xs font-bold">
                          {alert.location || "Distributed"}
                        </span>
                      </div>
                      <div className="flex items-center gap-2">
                        <ClockIcon className="w-4 h-4 opacity-50" />
                        <span className="text-xs font-bold">
                          {uiStrings.label_observed || "Observed:"}{" "}
                          {alert.count || 1} times
                        </span>
                      </div>
                    </div>
                  </div>

                  <div className="flex flex-col gap-2">
                    <button
                      onClick={() =>
                        handleUpdateStatus(alert.id || alert._id, "blocked")
                      }
                      disabled={updating === (alert.id || alert._id)}
                      className="px-4 py-2 rounded-lg bg-vibrant-red text-white text-[10px] font-black uppercase tracking-widest hover:bg-red-600 transition-all disabled:opacity-50"
                    >
                      {updating === (alert.id || alert._id)
                        ? uiStrings.btn_blocking || "BLOCKING..."
                        : uiStrings.btn_block || "BLOCK IP"}
                    </button>
                    <button
                      onClick={() =>
                        handleUpdateStatus(
                          alert.id || alert._id,
                          "acknowledged",
                        )
                      }
                      disabled={updating === (alert.id || alert._id)}
                      className="px-4 py-2 rounded-lg bg-white/5 text-white text-[10px] font-black uppercase tracking-widest hover:bg-white/10 transition-all border border-white/10 disabled:opacity-50"
                    >
                      {updating === (alert.id || alert._id)
                        ? uiStrings.btn_updating || "UPDATING..."
                        : uiStrings.btn_acknowledge || "ACKNOWLEDGE"}
                    </button>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>

        <SecurityFooter />
      </main>
    </div>
  );
};

export default ThreatAlerts;
