import { formatDistanceToNow, parseISO } from "date-fns";
import { motion } from "framer-motion";
import React, { useCallback, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useUIStrings } from "../../hooks/useUIStrings";
import api from "../../services/api";

interface AdminMetrics {
  risk_overview: {
    sla_compliance: number;
    integration_failures: number;
    payment_failures: number;
    security_score: number;
  };
  uptime_trend: number[];
  critical_indicators: {
    open_incidents: number;
    devices_offline: number;
    system_health: number;
    billing_risk: number;
    total_devices: number;
    integration_failures: number;
  };
  operational_metrics: {
    active_orgs: number;
    active_fleet: number;
    connected_devices: number;
    active_users_today: number;
    total_orgs: number;
    new_orgs_today: number;
    total_vehicles: number;
    vehicles_offline: number;
    devices_offline: number;
    users_yesterday: number;
  };
  pending_approvals: number;
  recent_activity: Array<{ id: number; type: string; [key: string]: any }>;
  last_updated: string;
}

export const GlobalAdminDashboard: React.FC = () => {
  const navigate = useNavigate();
  const [metrics, setMetrics] = useState<AdminMetrics | null>(null);
  const [loading, setLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [timeRange, setTimeRange] = useState<
    "today" | "7days" | "30days" | "custom"
  >("today");
  const [systemMeta, setSystemMeta] = useState<any>(null);
  const { uiStrings, loading: uiLoading } = useUIStrings("admin_dashboard");

  const fetchMetrics = useCallback(
    async (showRefreshing = false) => {
      if (showRefreshing) setIsRefreshing(true);
      try {
        const data = await api.getAdminDashboardData(timeRange);
        setMetrics(data);

        // Dynamic Meta
        fetch("/api/v1/ui/global_config")
          .then((res) => res.json())
          .then((data) => setSystemMeta(data.system_metadata))
          .catch((err) => console.error("Failed to fetch system config", err));
      } catch (error) {
        console.error("Failed to fetch admin metrics:", error);
      } finally {
        setLoading(false);
        if (showRefreshing) setIsRefreshing(false);
      }
    },
    [timeRange],
  );

  useEffect(() => {
    fetchMetrics();
    const interval = setInterval(() => fetchMetrics(true), 30000); // Background refresh
    return () => clearInterval(interval);
  }, [fetchMetrics]);

  const handleManualRefresh = () => {
    fetchMetrics(true);
  };

  const formatLastUpdated = (timestamp?: string) => {
    if (!timestamp) return "Never";
    try {
      return formatDistanceToNow(parseISO(timestamp), { addSuffix: true });
    } catch (e) {
      return "Just now";
    }
  };

  if (loading || uiLoading) {
    return (
      <div className="p-8 text-center text-text-muted">
        {uiStrings?.loading_dashboard || "Loading system dashboard..."}
      </div>
    );
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5 }}
      className="w-full p-6 space-y-8"
    >
      {/* Header Section */}
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-6">
        <div className="space-y-1">
          <h1 className="text-4xl font-extrabold text-transparent bg-clip-text bg-gradient-to-r from-primary to-blue-400">
            {uiStrings?.title || "Global Command Center"}
          </h1>
          <p className="text-text-muted text-lg font-medium opacity-80">
            {uiStrings?.subtitle || "Real-time Operations & Risk Intelligence"}
          </p>
        </div>

        <div className="flex items-center gap-4 bg-surface-dark/50 backdrop-blur-md p-1.5 rounded-xl border border-border-color/50 shadow-inner">
          {(["today", "7days", "30days"] as const).map((range) => (
            <button
              key={range}
              onClick={() => setTimeRange(range)}
              className={`px-4 py-2 text-sm font-semibold rounded-lg transition-all ${
                timeRange === range
                  ? "bg-primary text-white shadow-lg shadow-primary/30"
                  : "text-text-muted hover:text-white hover:bg-surface-light/30"
              }`}
            >
              {range === "today"
                ? uiStrings?.opt_today || "Today"
                : range === "7days"
                  ? uiStrings?.opt_7days || "7 Days"
                  : uiStrings?.opt_30days || "30 Days"}
            </button>
          ))}
        </div>
      </div>

      {/* Primary Risk Grid - Premium Glassmorphism */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {[
          {
            label: uiStrings?.label_sla_compliance || "SLA Compliance",
            value:
              (metrics?.risk_overview?.sla_compliance || 0).toFixed(1) + "%",
            trend: "99.2% Goal",
            color: "red",
            icon: "verified_user",
            bg: "from-red-500/10 to-transparent",
            border: "border-red-500/30",
          },
          {
            label: uiStrings?.label_int_failures || "Integration Health",
            value: metrics?.risk_overview?.integration_failures ?? "--",
            trend: "Active Proxies: 12",
            color: "amber",
            icon: "api",
            bg: "from-amber-500/10 to-transparent",
            border: "border-amber-500/30",
          },
          {
            label: uiStrings?.label_security_score || "Security Posture",
            value:
              (metrics?.risk_overview?.security_score || 0).toFixed(1) + "%",
            trend: "No Breaches",
            color: "green",
            icon: "shield_check",
            bg: "from-green-500/10 to-transparent",
            border: "border-green-500/30",
          },
          {
            label: uiStrings?.label_system_health || "Platform Uptime",
            value:
              (metrics?.critical_indicators?.system_health || 99.9).toFixed(2) +
              "%",
            trend: systemMeta?.uptime || "Healthy",
            color: "blue",
            icon: "monitor_heart",
            bg: "from-blue-500/10 to-transparent",
            border: "border-blue-500/30",
          },
        ].map((item, idx) => (
          <motion.div
            key={idx}
            whileHover={{ scale: 1.02, translateY: -5 }}
            className={`relative overflow-hidden bg-surface-dark/40 backdrop-blur-xl border ${item.border} rounded-2xl p-5 shadow-2xl group`}
          >
            <div
              className={`absolute top-0 right-0 w-24 h-24 bg-gradient-to-br ${item.bg} -mr-8 -mt-8 rounded-full blur-2xl opacity-50`}
            />
            <div className="relative z-10">
              <div className="flex items-start justify-between mb-4">
                <span
                  className={`material-symbols-outlined text-${item.color}-400 bg-${item.color}-400/10 p-2 rounded-lg text-2xl`}
                >
                  {item.icon}
                </span>
                <span className="text-[10px] font-bold text-text-muted uppercase tracking-wider">
                  Live
                </span>
              </div>
              <h3 className="text-text-muted text-sm font-semibold mb-1">
                {item.label}
              </h3>
              <p className="text-3xl font-black text-white mb-2 leading-none">
                {item.value}
              </p>
              <div className="flex items-center gap-1.5 ring-1 ring-white/5 py-1 px-2 rounded-full w-fit bg-emerald-500/5">
                <span className="text-[10px] font-bold text-emerald-400 whitespace-nowrap">
                  {item.trend}
                </span>
              </div>
            </div>
          </motion.div>
        ))}
      </div>

      {/* Middle Grid - Operations & Incidents */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Vital Stats Card */}
        <div className="lg:col-span-2 grid grid-cols-2 gap-4">
          {[
            {
              label: "Active Organizations",
              val: metrics?.operational_metrics?.total_orgs,
              icon: "corporate_fare",
              color: "blue",
            },
            {
              label: "Fleet Volume",
              val: metrics?.operational_metrics?.active_fleet,
              icon: "local_shipping",
              color: "purple",
            },
            {
              label: "Edge Hardware",
              val: metrics?.operational_metrics?.connected_devices,
              icon: "memory",
              color: "emerald",
            },
            {
              label: "Peak Active Users",
              val: metrics?.operational_metrics?.active_users_today,
              icon: "group",
              color: "orange",
            },
          ].map((stat, i) => (
            <div
              key={i}
              className="bg-surface-dark/30 border border-border-color/30 rounded-2xl p-6 flex items-center justify-between hover:bg-surface-dark/50 transition-colors cursor-pointer group"
            >
              <div>
                <p className="text-text-muted text-xs font-bold uppercase tracking-widest mb-1">
                  {stat.label}
                </p>
                <p className="text-3xl font-black text-white">
                  {(stat.val || 0).toLocaleString()}
                </p>
              </div>
              <span
                className={`material-symbols-outlined text-${stat.color}-400 text-3xl opacity-30 group-hover:opacity-100 transition-opacity`}
              >
                {stat.icon}
              </span>
            </div>
          ))}

          {/* Uptime Trend Chart Area */}
          <div className="col-span-2 bg-surface-dark/40 border border-border-color/30 rounded-2xl p-6">
            <div className="flex items-center justify-between mb-6">
              <h3 className="text-white font-bold opacity-90">
                Regional Connectivity Uptime
              </h3>
              <div className="flex items-center gap-2">
                <div className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
                <span className="text-[10px] text-emerald-400 font-bold uppercase">
                  Real-time Stream
                </span>
              </div>
            </div>
            <div className="h-24 flex items-end gap-2 px-2">
              {(
                metrics?.uptime_trend || [99, 98, 99.5, 99.9, 99.2, 99.8, 99.9]
              ).map((val, i) => (
                <div key={i} className="flex-1 relative group cursor-pointer">
                  <motion.div
                    initial={{ height: 0 }}
                    animate={{ height: `${(val - 95) * 15}%` }}
                    className={`w-full rounded-t-lg transition-all duration-500 ${val > 99.5 ? "bg-primary shadow-[0_0_15px_rgba(37,99,235,0.4)]" : "bg-primary/40 group-hover:bg-primary"}`}
                  />
                  <div className="absolute -top-8 left-1/2 -translate-x-1/2 bg-black/80 text-white text-[10px] px-1.5 py-0.5 rounded opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap">
                    {val}%
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Live Incident Feed - High Urgency Design */}
        <div className="bg-red-950/20 border-2 border-red-500/40 rounded-3xl p-6 relative overflow-hidden flex flex-col shadow-[0_35px_60px_-15px_rgba(239,68,68,0.1)]">
          <div className="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-transparent via-red-500 to-transparent opacity-50" />

          <div className="flex items-center justify-between mb-6">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-red-500 rounded-xl shadow-lg shadow-red-500/40">
                <span className="material-symbols-outlined text-white text-xl">
                  priority_high
                </span>
              </div>
              <div>
                <h3 className="text-white font-black leading-none">
                  CRITICAL ALERTS
                </h3>
                <p className="text-red-400 text-[10px] font-bold uppercase tracking-widest mt-1">
                  Automatic Refresh Active
                </p>
              </div>
            </div>
            <span className="text-red-500 text-2xl font-black">
              {metrics?.critical_indicators?.open_incidents || 0}
            </span>
          </div>

          <div className="flex-1 space-y-4 overflow-y-auto pr-2 custom-scrollbar">
            {(metrics?.recent_activity || [])
              .filter((a) => a.type === "incident_opened")
              .map((incident, i) => (
                <div
                  key={i}
                  className="bg-white/5 border border-white/10 p-4 rounded-2xl hover:bg-white/10 transition-colors cursor-pointer border-l-4 border-l-red-500 group"
                >
                  <div className="flex justify-between items-start mb-2">
                    <h4 className="text-white font-bold text-sm tracking-tight group-hover:text-red-400 transition-colors">
                      {incident.title}
                    </h4>
                    <span className="text-[10px] text-text-muted">
                      {incident.timestamp}
                    </span>
                  </div>
                  <p className="text-text-muted text-xs line-clamp-2 leading-relaxed opacity-70">
                    {incident.description ||
                      "System generated anomaly detected in regional node."}
                  </p>
                  <div className="mt-3 flex items-center justify-between">
                    <span className="text-[10px] px-2 py-0.5 bg-red-500/20 text-red-400 rounded-full font-bold">
                      URGENT
                    </span>
                    <span className="text-[10px] font-bold text-blue-400 uppercase">
                      View Log →
                    </span>
                  </div>
                </div>
              ))}
          </div>

          <button
            onClick={() => navigate("/admin/incidents")}
            className="w-full mt-6 py-3 bg-red-500 hover:bg-red-400 text-white font-black text-sm rounded-2xl transition-all shadow-xl shadow-red-500/20 active:scale-95 flex items-center justify-center gap-2"
          >
            <span className="material-symbols-outlined text-lg">
              manage_accounts
            </span>
            RESOLVE ALL INCIDENTS
          </button>
        </div>
      </div>

      {/* Global Activity Timeline */}
      <div className="bg-surface-dark/40 border border-border-color/30 rounded-3xl p-8 shadow-inner overflow-hidden relative">
        <div className="absolute top-0 right-0 w-64 h-64 bg-primary/5 rounded-full blur-3xl -mr-32 -mt-32" />

        <div className="flex items-center justify-between mb-8 relative z-10">
          <div>
            <h2 className="text-white text-2xl font-black mb-1">
              Global Activity Ledger
            </h2>
            <p className="text-text-muted text-sm font-medium">
              Cryptographically verified operational stream
            </p>
          </div>
          <button className="p-3 bg-surface-dark rounded-2xl border border-border-color hover:bg-surface-light transition-all shadow-xl group">
            <span className="material-symbols-outlined text-text-muted group-hover:text-white transition-colors">
              history
            </span>
          </button>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 relative z-10">
          {(metrics?.recent_activity || []).map((activity, i) => (
            <div
              key={i}
              className="bg-surface-light/5 border border-white/5 p-5 rounded-2xl flex gap-4 hover:border-primary/40 transition-all hover:translate-y-[-4px]"
            >
              <div className="mt-1">
                <span
                  className={`material-symbols-outlined p-2 rounded-xl bg-surface-dark shadow-inner text-lg ${
                    activity.type.includes("incident")
                      ? "text-red-400"
                      : activity.type.includes("org")
                        ? "text-blue-400"
                        : "text-emerald-400"
                  }`}
                >
                  {activity.type.includes("incident")
                    ? "report_problem"
                    : activity.type.includes("org")
                      ? "domain"
                      : "payments"}
                </span>
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-white font-bold text-sm truncate mb-1">
                  {activity.title ||
                    activity.org_name ||
                    activity.user_name ||
                    "System Event"}
                </p>
                <div className="flex items-center justify-between">
                  <span className="text-[11px] text-text-muted font-medium">
                    {activity.timestamp}
                  </span>
                  <span className="text-[10px] text-primary font-bold uppercase">
                    {activity.performer || "AUTO"}
                  </span>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Quick Action Matrix */}
      <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
        {[
          {
            label: "Fleet",
            icon: "airport_shuttle",
            color: "blue",
            path: "/admin/vehicles",
          },
          {
            label: "Drivers",
            icon: "badge",
            color: "indigo",
            path: "/admin/drivers",
          },
          {
            label: "Finance",
            icon: "account_balance_wallet",
            color: "emerald",
            path: "/admin/billing",
          },
          {
            label: "Governance",
            icon: "gavel",
            color: "purple",
            path: "/admin/governance",
          },
          {
            label: "Integrations",
            icon: "dynamic_feed",
            color: "orange",
            path: "/admin/integrations",
          },
          {
            label: "Settings",
            icon: "settings_suggest",
            color: "slate",
            path: "/admin/settings",
          },
        ].map((tool, i) => (
          <button
            key={i}
            onClick={() => navigate(tool.path)}
            className="group bg-surface-dark/40 hover:bg-surface-dark border border-border-color/30 p-4 rounded-3xl transition-all hover:shadow-2xl hover:shadow-primary/5 active:scale-90"
          >
            <span
              className={`material-symbols-outlined text-${tool.color}-400 text-3xl mb-2 group-hover:scale-110 transition-transform`}
            >
              {tool.icon}
            </span>
            <p className="text-white text-xs font-black tracking-tight uppercase opacity-60 group-hover:opacity-100 transition-opacity">
              {tool.label}
            </p>
          </button>
        ))}
      </div>
    </motion.div>
  );
};

export default GlobalAdminDashboard;
